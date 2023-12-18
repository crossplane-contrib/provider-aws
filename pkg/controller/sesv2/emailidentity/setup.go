/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package emailidentity

import (
	"context"
	"encoding/base64"

	svcsdk "github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/aws/aws-sdk-go/service/sesv2/sesv2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/sesv2/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/sesv2/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotEmailIdentity          = "managed resource is not a SES EmailIdentity custom resource"
	errKubeUpdateFailed          = "cannot update SES EmailIdentity custom resource"
	errGetPrivateKeySecretFailed = "cannot get DKIM Signing private key"
)

// SetupEmailIdentity adds a controller that reconciles SES EmailIdentity.
func SetupEmailIdentity(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.EmailIdentityGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client, kube: e.kube}
			e.isUpToDate = h.isUpToDate
			e.postObserve = postObserve
			e.preCreate = h.preCreate
			e.update = h.update
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.EmailIdentity{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.EmailIdentityGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type hooks struct {
	client sesv2iface.SESV2API
	kube   client.Client
}

func (e *hooks) isUpToDate(_ context.Context, cr *svcapitypes.EmailIdentity, resp *svcsdk.GetEmailIdentityOutput) (bool, string, error) { //nolint:gocyclo
	if pointer.StringValue(cr.Spec.ForProvider.ConfigurationSetName) != pointer.StringValue(resp.ConfigurationSetName) {
		return false, "", nil
	}

	// Checks if MailFromAttributes Object are up to date
	if cr.Spec.ForProvider.MailFromAttributes != nil && resp.MailFromAttributes != nil {
		// BehaviorOnMxFailure Response by default return "USE_DEFAULT_VALUE"
		if pointer.StringValue(cr.Spec.ForProvider.MailFromAttributes.BehaviorOnMxFailure) != pointer.StringValue(resp.MailFromAttributes.BehaviorOnMxFailure) {
			return false, "", nil
		}
		if pointer.StringValue(cr.Spec.ForProvider.MailFromAttributes.MailFromDomain) != pointer.StringValue(resp.MailFromAttributes.MailFromDomain) {
			return false, "", nil
		}
	}

	// Checks if DkimSigningAttributes Object are up to date
	// TODO(kelvinwijaya): This is temporary solution because GetEmailIdentityOutput.DkimSigningAttributes is not supported in aws-sdk-go v1.44.174
	// To trigger update need toggle based on NextSigningKeyLength OR DomainSigningPrivateKey/DomainSigningPrivateKeySecretRef
	if cr.Spec.ForProvider.DkimSigningAttributes != nil && resp.DkimAttributes.NextSigningKeyLength != nil {
		// EasyDKIM mode - Update NextSigningKeyLength or Toggle mode if cr value is empty
		if pointer.StringValue(cr.Spec.ForProvider.DkimSigningAttributes.NextSigningKeyLength) != pointer.StringValue(resp.DkimAttributes.NextSigningKeyLength) {
			return false, "", nil
		}
	}

	if cr.Spec.ForProvider.DkimSigningAttributes != nil && len(resp.DkimAttributes.Tokens) == 1 {
		// BYODKIM mode - Update DkimSigningAttributes domain or Toggle mode if cr value is empty
		if pointer.StringValue(cr.Spec.ForProvider.DkimSigningAttributes.DomainSigningSelector) != pointer.StringValue(resp.DkimAttributes.Tokens[0]) {
			return false, "", nil
		}
	}

	areTagsUpToDate, err := svcutils.AreTagsUpToDate(cr.Spec.ForProvider.Tags, resp.Tags)
	return areTagsUpToDate, "", err
}

func postObserve(_ context.Context, cr *svcapitypes.EmailIdentity, resp *svcsdk.GetEmailIdentityOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.BoolValue(resp.VerifiedForSendingStatus) {
	case true:
		cr.Status.SetConditions(xpv1.Available())
	case false:
		cr.Status.SetConditions(xpv1.Unavailable())
	default:
		cr.Status.SetConditions(xpv1.Creating())
	}

	return obs, nil
}

func (e *hooks) preCreate(ctx context.Context, cr *svcapitypes.EmailIdentity, obj *svcsdk.CreateEmailIdentityInput) error {
	if cr.Spec.ForProvider.DkimSigningAttributes != nil && cr.Spec.ForProvider.DkimSigningAttributes.DomainSigningSelector != nil {
		// Retrieve from Secret Object if DomainSigningPrivateKeySecretRef is available
		if cr.Spec.ForProvider.DomainSigningPrivateKeySecretRef != nil {
			pk, err := e.getPrivateKeyFromRef(ctx, cr.Spec.ForProvider.DomainSigningPrivateKeySecretRef)
			if err != nil {
				return errors.Wrap(err, "private key retrival failed")
			}
			obj.DkimSigningAttributes.DomainSigningPrivateKey = pointer.ToOrNilIfZeroValue(base64.StdEncoding.EncodeToString([]byte(pk)))
		}
	}
	return nil
}

func (e *hooks) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	// NOTE: Update operation NOT generated by ACK code-generator
	cr, ok := mg.(*svcapitypes.EmailIdentity)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// update DKIMAttributes (PutEmailIdentityDKIMSigningAttributes) - Enabled as part of SES for Domain Validation using DKIM Records
	// NOTE: GetEmailIdentityOutput.DkimSigningAttributes is not supported in aws-sdk-go v1.44.174 - update has to be triggered together with changes in other field e.g. MailFromAttributes
	// This is done in postCreate stage instead

	// update ConfigurationSetAttributes (PutEmailIdentityConfigurationSetAttributes)
	configurationSetAttributesInput := &svcsdk.PutEmailIdentityConfigurationSetAttributesInput{
		ConfigurationSetName: cr.Spec.ForProvider.ConfigurationSetName,
		EmailIdentity:        cr.Spec.ForProvider.EmailIdentity,
	}
	if _, err := e.client.PutEmailIdentityConfigurationSetAttributesWithContext(ctx, configurationSetAttributesInput); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for EmailIdentityConfigurationSetAttributes")
	}

	// Update MailFromAttributes (PutEmailIdentityMailFromAttributes)
	// NOTE: Currently MailFromAttributes is not supported in aws-sdk-go v1.44.174 and is declared in Custom-Type
	if cr.Spec.ForProvider.MailFromAttributes != nil {
		mailFromAttributesInput := &svcsdk.PutEmailIdentityMailFromAttributesInput{
			BehaviorOnMxFailure: cr.Spec.ForProvider.MailFromAttributes.BehaviorOnMxFailure,
			EmailIdentity:       cr.Spec.ForProvider.EmailIdentity,
			MailFromDomain:      cr.Spec.ForProvider.MailFromAttributes.MailFromDomain,
		}

		if _, err := e.client.PutEmailIdentityMailFromAttributesWithContext(ctx, mailFromAttributesInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for EmailIdentityMailFromAttributes")
		}
	}

	// Update DKIMSigningAttributes (PutEmailIdentityDkimSigningAttributes)
	// NOTE: Should be triggered by isUpToDateDkimSigningAttributes(cr, resp)
	// TODO(kelvinwijaya): This is temporary solution until aws-sdk-go version upgrade
	dkimSigningAttributeExternal := "EXTERNAL" // Ref: https://docs.aws.amazon.com/ses/latest/DeveloperGuide/send-email-authentication-dkim-bring-your-own.html
	dkimSigningAttributeEasyDKIM := "AWS_SES"  // Ref: https://docs.aws.amazon.com/ses/latest/dg/send-email-authentication-dkim-easy.html
	if cr.Spec.ForProvider.DkimSigningAttributes != nil && cr.Spec.ForProvider.DkimSigningAttributes.DomainSigningSelector != nil && cr.Spec.ForProvider.DomainSigningPrivateKeySecretRef != nil {
		// Retrieve from Secret Object if DomainSigningPrivateKeySecretRef is available
		pk, err := e.getPrivateKeyFromRef(ctx, cr.Spec.ForProvider.DomainSigningPrivateKeySecretRef)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "private key retrival failed")
		}
		if len(pk) == 0 {
			return managed.ExternalUpdate{}, errors.Wrap(err, "private key is not supplied")
		}

		dkimSigningAttributesInput := &svcsdk.PutEmailIdentityDkimSigningAttributesInput{
			EmailIdentity: cr.Spec.ForProvider.EmailIdentity,
			SigningAttributes: &svcsdk.DkimSigningAttributes{
				DomainSigningPrivateKey: pointer.ToOrNilIfZeroValue(base64.StdEncoding.EncodeToString([]byte(pk))),
				DomainSigningSelector:   cr.Spec.ForProvider.DkimSigningAttributes.DomainSigningSelector,
			},
			SigningAttributesOrigin: pointer.ToOrNilIfZeroValue(dkimSigningAttributeExternal),
		}
		if _, err := e.client.PutEmailIdentityDkimSigningAttributesWithContext(ctx, dkimSigningAttributesInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for EmailIdentityDkimSigningAttributes")
		}
	}
	// Else default to EasyDKIM if no DKIMSignatureAttributes with SigningAttributesOrigin: AWS_SES
	if cr.Spec.ForProvider.DkimSigningAttributes != nil && cr.Spec.ForProvider.DkimSigningAttributes.NextSigningKeyLength != nil {
		dkimSigningAttributesInput := &svcsdk.PutEmailIdentityDkimSigningAttributesInput{
			EmailIdentity:           cr.Spec.ForProvider.EmailIdentity,
			SigningAttributesOrigin: pointer.ToOrNilIfZeroValue(dkimSigningAttributeEasyDKIM),
			SigningAttributes: &svcsdk.DkimSigningAttributes{
				NextSigningKeyLength: cr.Spec.ForProvider.DkimSigningAttributes.NextSigningKeyLength,
			},
		}
		if _, err := e.client.PutEmailIdentityDkimSigningAttributesWithContext(ctx, dkimSigningAttributesInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for EmailIdentityDkimSigningAttributes")
		}
	}

	// All notifications customization should be added through Configuration sets setup
	return managed.ExternalUpdate{}, nil
}

func (e *hooks) getPrivateKeyFromRef(ctx context.Context, in *xpv1.SecretKeySelector) (newPrivateKey string, err error) {
	if in == nil {
		return "", nil
	}

	nn := types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
	s := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, s); err != nil {
		return "", errors.Wrap(err, errGetPrivateKeySecretFailed)
	}
	return string(s.Data[xpv1.ResourceCredentialsSecretClientKeyKey]), nil
}
