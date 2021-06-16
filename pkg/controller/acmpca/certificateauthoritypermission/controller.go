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
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/acmpca"
)

const (
	errUnexpectedObject = "The managed resource is not an ACMPCA resource"
	errGet              = "failed to get ACMPCA with name"
	errCreate           = "failed to create the ACMPCA resource"
	errDelete           = "failed to delete the ACMPCA resource"
)

// SetupCertificateAuthorityPermission adds a controller that reconciles ACMPCA.
func SetupCertificateAuthorityPermission(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.CertificateAuthorityPermissionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.CertificateAuthorityPermission{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateAuthorityPermissionGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewCAPermissionClient}),
			managed.WithConnectionPublishers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) acmpca.CAPermissionClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CertificateAuthorityPermission)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.client, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mgd.(*v1alpha1.CertificateAuthorityPermission)
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

	response, err := e.client.ListPermissionsRequest(&awsacmpca.ListPermissionsInput{
		CertificateAuthorityArn: &caARN,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errGet)
	}

	var attachedPermission *awsacmpca.Permission
	for i := range response.Permissions {
		if awsclient.StringValue(response.Permissions[i].Principal) == principal {
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
	cr, ok := mgd.(*v1alpha1.CertificateAuthorityPermission)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.CreatePermissionRequest(&awsacmpca.CreatePermissionInput{
		Actions:                 []awsacmpca.ActionType{awsacmpca.ActionTypeIssueCertificate, awsacmpca.ActionTypeGetCertificate, awsacmpca.ActionTypeListPermissions},
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(cr.Spec.ForProvider.Principal),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	// This resource is interesting in that it's a binding without its own
	// external identity. We therefore derive an external name from the
	// identity of the CA it applies to, and the principal it applies.
	meta.SetExternalName(cr, cr.Spec.ForProvider.Principal+"/"+awsclient.StringValue(cr.Spec.ForProvider.CertificateAuthorityARN))

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.CertificateAuthorityPermission)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	_, err := e.client.DeletePermissionRequest(&awsacmpca.DeletePermissionInput{
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(cr.Spec.ForProvider.Principal),
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
}
