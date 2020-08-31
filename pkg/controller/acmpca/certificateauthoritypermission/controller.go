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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/acmpca"
)

const (
	errUnexpectedObject = "The managed resource is not an ACMPCA resource"
	errGet              = "failed to get ACMPCA with name"
	errCreate           = "failed to create the ACMPCA resource"
	errDelete           = "failed to delete the ACMPCA resource"

	principal = "acm.amazonaws.com"
)

// SetupCertificateAuthorityPermission adds a controller that reconciles ACMPCA.
func SetupCertificateAuthorityPermission(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CertificateAuthorityPermissionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CertificateAuthorityPermission{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateAuthorityPermissionGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewCAPermissionClient}),
			managed.WithConnectionPublishers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) acmpca.CAPermissionClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awscommon.GetConfig(ctx, c.client, mg, "")
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

	response, err := e.client.ListPermissionsRequest(&awsacmpca.ListPermissionsInput{
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errGet)
	}

	if len(response.Permissions) == 0 {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.SetConditions(runtimev1alpha1.Available())

	return managed.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha1.CertificateAuthorityPermission)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.client.CreatePermissionRequest(&awsacmpca.CreatePermissionInput{

		Actions:                 []awsacmpca.ActionType{awsacmpca.ActionTypeIssueCertificate, awsacmpca.ActionTypeGetCertificate, awsacmpca.ActionTypeListPermissions},
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(principal),
	}).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.CertificateAuthorityPermission)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeletePermissionRequest(&awsacmpca.DeletePermissionInput{
		CertificateAuthorityArn: cr.Spec.ForProvider.CertificateAuthorityARN,
		Principal:               aws.String(principal),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
}
