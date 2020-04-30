/*
Copyright 2019 The Crossplane Authors.

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

package certificatemanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacm "github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha1 "github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	acm "github.com/crossplane/provider-aws/pkg/clients/certificatemanager"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an ACM resource"
	errClient           = "cannot create a new ACM client"
	errGet              = "failed to get Certificate with name"
	errCreate           = "failed to create the Certificate resource"
	errDelete           = "failed to delete the Certificate resource"
	errUpdate           = "failed to update the Certificate resource"
	errSDK              = "empty Certificate received from ACM API"

	errKubeUpdateFailed = "cannot late initialize Certificate"
	// errUpToDateFailed   = "cannot check whether object is up-to-date"
)

// SetupCertificate adds a controller that reconciles Certificates.
func SetupCertificate(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CertificateGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Certificate{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acm.NewClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (acm.Client, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}
	return &external{c, conn.client}, nil
}

type external struct {
	client acm.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {

	fmt.Println("Observ | Entry")

	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if cr.Status.AtProvider.CertificateArn == "" {
		fmt.Println("CertificateArn is empty")
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeCertificateRequest(&awsacm.DescribeCertificateInput{
		CertificateArn: aws.String(cr.Status.AtProvider.CertificateArn),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errGet)
	}

	if response.Certificate == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	certificate := *response.Certificate
	current := cr.Spec.ForProvider.DeepCopy()
	fmt.Println("Calling LateInitialize")
	acm.LateInitializeCertificate(&cr.Spec.ForProvider, &certificate)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = acm.GenerateCertificateStatus(certificate)

	// upToDate, err := acm.IsCertificateUpToDate(cr.Spec.ForProvider, certificate)
	// if err != nil {
	// 	return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	// }

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	fmt.Println("Create | Entry")

	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	response, err := e.client.RequestCertificateRequest(acm.GenerateCreateCertificateInput(meta.GetExternalName(cr), &cr.Spec.ForProvider)).Send(ctx)

	if response != nil {
		cr.Status.AtProvider.CertificateArn = aws.StringValue(response.RequestCertificateOutput.CertificateArn)
	}

	fmt.Println("Create | Exit")
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	fmt.Println("Update")
	return managed.ExternalUpdate{}, errors.New(errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	fmt.Println("Delete | Entry")
	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteCertificateRequest(&awsacm.DeleteCertificateInput{
		CertificateArn: aws.String(cr.Status.AtProvider.CertificateArn),
	}).Send(ctx)
	fmt.Println("Delete | Exit")
	return errors.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errDelete)
}
