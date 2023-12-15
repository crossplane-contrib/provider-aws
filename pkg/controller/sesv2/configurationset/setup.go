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

package configurationset

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/aws/aws-sdk-go/service/sesv2/sesv2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/sesv2/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/sesv2/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupConfigurationSet adds a controller that reconciles SES ConfigurationSet.
func SetupConfigurationSet(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ConfigurationSetGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client, kube: e.kube}
			e.isUpToDate = isUpToDate
			e.preObserve = preObserve
			e.postObserve = h.postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.update = h.update
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ConfigurationSet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ConfigurationSetGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type hooks struct {
	client                      sesv2iface.SESV2API
	kube                        client.Client
	ConfigurationSetObservation *svcsdk.GetConfigurationSetOutput
}

func isUpToDate(_ context.Context, cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) (bool, string, error) {
	if !isUpToDateDeliveryOptions(cr, resp) {
		return false, "", nil
	}

	if !isUpToDateReputationOptions(cr, resp) {
		return false, "", nil
	}

	if !isUpToDateSendingOptions(cr, resp) {
		return false, "", nil
	}

	if !isUpToDateSuppressionOptions(cr, resp) {
		return false, "", nil
	}

	if !isUpToDateTrackingOptions(cr, resp) {
		return false, "", nil
	}

	areTagsUpToDate, err := svcutils.AreTagsUpToDate(cr.Spec.ForProvider.Tags, resp.Tags)
	return areTagsUpToDate, "", err
}

// isUpToDateDeliveryOptions checks if DeliveryOptions Object are up to date
func isUpToDateDeliveryOptions(cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) bool {
	if cr.Spec.ForProvider.DeliveryOptions != nil && resp.DeliveryOptions != nil {
		if pointer.StringValue(cr.Spec.ForProvider.DeliveryOptions.SendingPoolName) != pointer.StringValue(resp.DeliveryOptions.SendingPoolName) {
			return false
		}
		if pointer.StringValue(cr.Spec.ForProvider.DeliveryOptions.TLSPolicy) != pointer.StringValue(resp.DeliveryOptions.TlsPolicy) {
			return false
		}
	}
	return true
}

// isUpToDateReputationOptions checks if ReputationOptions Object are up to date
func isUpToDateReputationOptions(cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) bool {
	if cr.Spec.ForProvider.ReputationOptions != nil && resp.ReputationOptions != nil {
		if pointer.BoolValue(cr.Spec.ForProvider.ReputationOptions.ReputationMetricsEnabled) != pointer.BoolValue(resp.ReputationOptions.ReputationMetricsEnabled) {
			return false
		}
	}
	return true
}

// isUpToDateTrackingOptions checks if TrackingOptions Object are up to date
func isUpToDateTrackingOptions(cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) bool {
	// Once disabled, output response will not populate this option anymore
	if cr.Spec.ForProvider.TrackingOptions != nil && resp.TrackingOptions == nil {
		return false
	}

	if cr.Spec.ForProvider.TrackingOptions != nil && resp.TrackingOptions != nil {
		if pointer.StringValue(cr.Spec.ForProvider.TrackingOptions.CustomRedirectDomain) != pointer.StringValue(resp.TrackingOptions.CustomRedirectDomain) {
			return false
		}
	}
	return true
}

// isUpToDateSuppressionOptions checks if SuppressionOptions Object are up to date
func isUpToDateSuppressionOptions(cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) bool {
	var crSuppressedReasons []*string
	var awsSuppressedReasons []*string

	if cr.Spec.ForProvider.SuppressionOptions != nil && cr.Spec.ForProvider.SuppressionOptions.SuppressedReasons != nil {
		crSuppressedReasons = cr.Spec.ForProvider.SuppressionOptions.SuppressedReasons
	}

	// SuppressedReasons Response return empty slice if not being configured (e.g. "SuppressedReasons": [])
	if resp.SuppressionOptions != nil && resp.SuppressionOptions.SuppressedReasons != nil {
		awsSuppressedReasons = resp.SuppressionOptions.SuppressedReasons
	}

	if len(crSuppressedReasons) != len(awsSuppressedReasons) {
		return false
	}

	sortCmp := cmpopts.SortSlices(func(i, j *string) bool {
		return aws.StringValue(i) < aws.StringValue(j)
	})

	return cmp.Equal(crSuppressedReasons, awsSuppressedReasons, sortCmp, cmpopts.EquateEmpty())

}

// isUpToDateSendingOptions checks if SendingOptions Object are up to date
func isUpToDateSendingOptions(cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput) bool {
	if cr.Spec.ForProvider.SendingOptions != nil && resp.SendingOptions != nil {
		if pointer.BoolValue(cr.Spec.ForProvider.SendingOptions.SendingEnabled) != pointer.BoolValue(resp.SendingOptions.SendingEnabled) {
			return false
		}
	}
	return true
}

func preObserve(_ context.Context, cr *svcapitypes.ConfigurationSet, obj *svcsdk.GetConfigurationSetInput) error {
	obj.ConfigurationSetName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (e *hooks) postObserve(_ context.Context, cr *svcapitypes.ConfigurationSet, resp *svcsdk.GetConfigurationSetOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.BoolValue(resp.SendingOptions.SendingEnabled) {
	case true:
		cr.Status.SetConditions(xpv1.Available())
	case false:
		cr.Status.SetConditions(xpv1.Unavailable())
	default:
		cr.Status.SetConditions(xpv1.Creating())
	}

	// Passing ConfigurationSet object from Observation into hooks for Update function to access
	e.ConfigurationSetObservation = resp

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.ConfigurationSet, obj *svcsdk.CreateConfigurationSetInput) error {
	obj.ConfigurationSetName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.ConfigurationSet, obj *svcsdk.DeleteConfigurationSetInput) (bool, error) {
	obj.ConfigurationSetName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func (e *hooks) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*svcapitypes.ConfigurationSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	// Update Resource is not provided other than individual PUT operation
	// NOTE: Update operation NOT generated by ACK code-generator

	// Populate ConfigurationSetName from meta.AnnotationKeyExternalName
	configurationSetName := pointer.ToOrNilIfZeroValue(mg.GetAnnotations()[meta.AnnotationKeyExternalName])

	// update DeliveryOptions (PutConfigurationSetDeliveryOptions)
	if !isUpToDateDeliveryOptions(cr, e.ConfigurationSetObservation) {
		deliveryOptionsInput := &svcsdk.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: configurationSetName,
			SendingPoolName:      cr.Spec.ForProvider.DeliveryOptions.SendingPoolName,
			TlsPolicy:            cr.Spec.ForProvider.DeliveryOptions.TLSPolicy,
		}
		if _, err := e.client.PutConfigurationSetDeliveryOptionsWithContext(ctx, deliveryOptionsInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for ConfigurationSetDeliveryOptions")
		}
	}

	// update ReputationOptions (PutConfigurationSetReputationOptions)
	if !isUpToDateReputationOptions(cr, e.ConfigurationSetObservation) {
		reputationOptionsInput := &svcsdk.PutConfigurationSetReputationOptionsInput{
			ConfigurationSetName:     configurationSetName,
			ReputationMetricsEnabled: cr.Spec.ForProvider.ReputationOptions.ReputationMetricsEnabled,
		}
		if _, err := e.client.PutConfigurationSetReputationOptionsWithContext(ctx, reputationOptionsInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for ConfigurationSetReputationOptions")
		}
	}

	// update SuppressionOptions (PutConfigurationSetSuppressionOptions)
	var suppresssedReasons []*string
	if !isUpToDateSuppressionOptions(cr, e.ConfigurationSetObservation) {
		suppresssedReasons = cr.Spec.ForProvider.SuppressionOptions.SuppressedReasons
		supressOptionsInput := &svcsdk.PutConfigurationSetSuppressionOptionsInput{
			ConfigurationSetName: configurationSetName,
			SuppressedReasons:    suppresssedReasons,
		}
		if _, err := e.client.PutConfigurationSetSuppressionOptionsWithContext(ctx, supressOptionsInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for ConfigurationSetSuppressionOptions")
		}
	}

	// update TrackingOptions (PutConfigurationSetTrackingOptions)
	if !isUpToDateTrackingOptions(cr, e.ConfigurationSetObservation) {
		trackingOptionInput := &svcsdk.PutConfigurationSetTrackingOptionsInput{
			ConfigurationSetName: configurationSetName,
			CustomRedirectDomain: cr.Spec.ForProvider.TrackingOptions.CustomRedirectDomain,
		}
		if _, err := e.client.PutConfigurationSetTrackingOptionsWithContext(ctx, trackingOptionInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for ConfigurationSetTrackingOptions")
		}
	}

	// update SendingOptions (PutConfigurationSetSendingOptions)
	if !isUpToDateSendingOptions(cr, e.ConfigurationSetObservation) {
		sendingOptionInput := &svcsdk.PutConfigurationSetSendingOptionsInput{
			ConfigurationSetName: configurationSetName,
			SendingEnabled:       cr.Spec.ForProvider.SendingOptions.SendingEnabled,
		}
		if _, err := e.client.PutConfigurationSetSendingOptionsWithContext(ctx, sendingOptionInput); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "update failed for ConfigurationSetSendingOptions")
		}
	}

	return managed.ExternalUpdate{}, nil
}
