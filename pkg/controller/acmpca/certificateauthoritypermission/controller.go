/*
Copyright 2020 The Crossplane Authors.

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

package certificateauthoritypermission

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	awsacmpcatypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/acmpca/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acmpca"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an ACMPCA resource"
	errGet              = "failed to get ACMPCA with name"
	errCreate           = "failed to create the ACMPCA resource"
	errDelete           = "failed to delete the ACMPCA resource"
)

// SetupCertificateAuthorityPermission adds a controller that reconciles ACMPCA.
func SetupCertificateAuthorityPermission(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.CertificateAuthorityPermissionGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewCAPermissionClient}),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CertificateAuthorityPermissionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.CertificateAuthorityPermission{}).
		Complete(r)
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) acmpca.CAPermissionClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.CertificateAuthorityPermission)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.client, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(cfg), c.client}, nil
}

type external struct {
	client acmpca.CAPermissionClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.CertificateAuthorityPermission)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	// ARN can have its own slashes, we'll use the first part and assume the rest
	// is ARN.
	nn := strings.SplitN(meta.GetExternalName(cr), "/", 2)
	if len(nn) != 2 {
		return managed.ExternalObservation{}, errors.New("external name has to be in the following format <principal>/<ca-arn>")
	}
	principal, caARN := nn[0], nn[1]

	response, err := e.client.ListPermissions(ctx, &awsacmpca.ListPermissionsInput{
		CertificateAuthorityArn: &caARN,
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errGet)
	}

	var attachedPermission *awsacmpcatypes.Permission
	for i := range response.Permissions {
		if pointer.StringValue(response.Permissions[i].Principal) == principal {
			attachedPermission = &response.Permissions[i]
			break
		}
	}

	if attachedPermission == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.CertificateAuthorityPermission)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	in := &awsacmpca.CreatePermissionInput{
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(cr.Spec.ForProvider.Principal),
	}
	in.Actions = make([]awsacmpcatypes.ActionType, len(cr.Spec.ForProvider.Actions))
	for i := range cr.Spec.ForProvider.Actions {
		in.Actions[i] = awsacmpcatypes.ActionType(cr.Spec.ForProvider.Actions[i])
	}

	_, err := e.client.CreatePermission(ctx, in)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	// This resource is interesting in that it's a binding without its own
	// external identity. We therefore derive an external name from the
	// identity of the CA it applies to, and the principal it applies.
	meta.SetExternalName(cr, cr.Spec.ForProvider.Principal+"/"+pointer.StringValue(cr.Spec.ForProvider.CertificateAuthorityARN))

	return managed.ExternalCreation{}, nil

}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.CertificateAuthorityPermission)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	_, err := e.client.DeletePermission(ctx, &awsacmpca.DeletePermissionInput{
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(cr.Spec.ForProvider.Principal),
	})

	return errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
}
