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

package certificateauthority

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
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
	acmpca "github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificateauthority"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errPendingStatus    = "The managed resource in pending status, please open the ACM Private CA console https://console.aws.amazon.com/acm-pca/home install CA certificate "
	errUnexpectedObject = "The managed resource is not an ACMPCA resource"
	errClient           = "cannot create a new ACMPCA client"
	errGet              = "failed to get ACMPCA with name"
	errCreate           = "failed to create the ACMPCA resource"
	errDelete           = "failed to delete the ACMPCA resource"
	errSDK              = "empty ACMPCA received from ACMPCA API"

	errKubeUpdateFailed    = "cannot late initialize ACMPCA"
	errUpToDateFailed      = "cannot check whether object is up-to-date"
	errPersistExternalName = "failed to persist Certificate ARN"

	errAddTagsFailed        = "cannot add tags to ACMPCA"
	errListTagsFailed       = "failed to list tags for ACMPCA"
	errRemoveTagsFailed     = "failed to remove tags for ACMPCA"
	errCertificateAuthority = "failed to update the ACMPCA resource"
	errPermissionFailed     = "failed to update ACMPCA permission"

	principal = "acm.amazonaws.com"
)

// SetupCertificateAuthority adds a controller that reconciles ACMPCA.
func SetupCertificateAuthority(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CertificateAuthorityGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CertificateAuthority{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateAuthorityGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (acmpca.Client, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
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
	client acmpca.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeCertificateAuthorityRequest(&awsacmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errGet)
	}

	if response.CertificateAuthority == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	// Check the PCA status and return error if PCA is in Pending State.
	if response.CertificateAuthority.Status == awsacmpca.CertificateAuthorityStatusPendingCertificate {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, errors.New(errPendingStatus)
	}

	certificateAuthority := *response.CertificateAuthority
	current := cr.Spec.ForProvider.DeepCopy()
	acmpca.LateInitializeCertificateAuthority(&cr.Spec.ForProvider, &certificateAuthority)

	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = acmpca.GenerateCertificateAuthorityExternalStatus(certificateAuthority, cr)

	tags, err := e.client.ListTagsRequest(&awsacmpca.ListTagsInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errListTagsFailed)
	}

	upToDate := acmpca.IsCertificateAuthorityUpToDate(cr, certificateAuthority, tags.Tags)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	response, err := e.client.CreateCertificateAuthorityRequest(acmpca.GenerateCreateCertificateAuthorityInput(&cr.Spec.ForProvider)).Send(ctx)

	if response != nil {

		meta.SetExternalName(cr, aws.StringValue(response.CreateCertificateAuthorityOutput.CertificateAuthorityArn))
		if err = e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
		}

		if cr.Spec.ForProvider.CertificateRenewalPermissionAllow {

			_, err = e.client.CreatePermissionRequest(&awsacmpca.CreatePermissionInput{

				Actions:                 []awsacmpca.ActionType{awsacmpca.ActionTypeIssueCertificate, awsacmpca.ActionTypeGetCertificate, awsacmpca.ActionTypeListPermissions},
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Principal:               aws.String(principal),
			}).Send(ctx)

		}
	}

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo

	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	if cr.Spec.ForProvider.CertificateRenewalPermissionAllow != cr.Status.AtProvider.RenewalPermission {

		cr.Status.AtProvider.RenewalPermission = cr.Spec.ForProvider.CertificateRenewalPermissionAllow

		if cr.Spec.ForProvider.CertificateRenewalPermissionAllow {

			_, err := e.client.CreatePermissionRequest(&awsacmpca.CreatePermissionInput{
				Actions:                 []awsacmpca.ActionType{awsacmpca.ActionTypeIssueCertificate, awsacmpca.ActionTypeGetCertificate, awsacmpca.ActionTypeListPermissions},
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Principal:               aws.String(principal),
			}).Send(ctx)

			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errPermissionFailed)
			}

		} else {
			_, err := e.client.DeletePermissionRequest(&awsacmpca.DeletePermissionInput{
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Principal:               aws.String(principal),
			}).Send(ctx)

			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errPermissionFailed)
			}
		}

	}

	if len(cr.Spec.ForProvider.Tags) > 0 {

		tags := make([]awsacmpca.Tag, len(cr.Spec.ForProvider.Tags))
		for i, t := range cr.Spec.ForProvider.Tags {
			tags[i] = awsacmpca.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}

		currentTags, err := e.client.ListTagsRequest(&awsacmpca.ListTagsInput{
			CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
		}).Send(ctx)

		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errListTagsFailed)
		}

		if len(tags) < len(currentTags.Tags) {
			_, err := e.client.UntagCertificateAuthorityRequest(&awsacmpca.UntagCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Tags:                    currentTags.Tags,
			}).Send(ctx)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errRemoveTagsFailed)
			}
		}
		_, err = e.client.TagCertificateAuthorityRequest(&awsacmpca.TagCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
			Tags:                    tags,
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAddTagsFailed)
		}
	}

	_, err := e.client.UpdateCertificateAuthorityRequest(acmpca.GenerateUpdateCertificateAuthorityInput(cr)).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCertificateAuthority)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	response, err := e.client.DescribeCertificateAuthorityRequest(&awsacmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
	}

	if response != nil {
		if response.CertificateAuthority.Status == awsacmpca.CertificateAuthorityStatusActive {
			_, err = e.client.UpdateCertificateAuthorityRequest(&awsacmpca.UpdateCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Status:                  awsacmpca.CertificateAuthorityStatusDisabled,
			}).Send(ctx)

			if err != nil {
				return errors.Wrap(err, errDelete)
			}
		}
	}

	_, err = e.client.DeleteCertificateAuthorityRequest(&awsacmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn:     aws.String(meta.GetExternalName(cr)),
		PermanentDeletionTimeInDays: cr.Spec.ForProvider.PermanentDeletionTimeInDays,
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
}
