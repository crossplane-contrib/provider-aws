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

package certificateauthority

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
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
	errEmpty            = "empty ACMPCA received from ACMPCA API"

	errKubeUpdateFailed    = "cannot late initialize ACMPCA"
	errPersistExternalName = "failed to persist Certificate ARN"

	errAddTagsFailed        = "cannot add tags to ACMPCA"
	errListTagsFailed       = "failed to list tags for ACMPCA"
	errRemoveTagsFailed     = "failed to remove tags for ACMPCA"
	errCertificateAuthority = "failed to update the ACMPCA resource"
)

// SetupCertificateAuthority adds a controller that reconciles ACMPCA.
func SetupCertificateAuthority(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CertificateAuthorityGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CertificateAuthority{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateAuthorityGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewClient}),
			managed.WithConnectionPublishers(),

			// TODO: implement tag initializer

			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) acmpca.Client
}

func (conn *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CertificateAuthority)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awscommon.GetConfig(ctx, conn.client, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{conn.newClientFn(cfg), conn.client}, nil
}

type external struct {
	client acmpca.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
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
		return managed.ExternalObservation{}, errors.New(errEmpty)
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

	cr.Status.AtProvider = acmpca.GenerateCertificateAuthorityExternalStatus(certificateAuthority)

	tags, err := e.client.ListTagsRequest(&awsacmpca.ListTagsInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errListTagsFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: acmpca.IsCertificateAuthorityUpToDate(cr, certificateAuthority, tags.Tags),
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
	}

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo

	cr, ok := mgd.(*v1alpha1.CertificateAuthority)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Update the Certificate Authority tags
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
		if len(tags) != len(currentTags.Tags) {
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

	// Update Certificate Authority configuration
	_, err := e.client.UpdateCertificateAuthorityRequest(&awsacmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
		RevocationConfiguration: acmpca.GenerateRevocationConfiguration(cr.Spec.ForProvider.RevocationConfiguration),
		Status:                  awsacmpca.CertificateAuthorityStatus(cr.Spec.ForProvider.Status),
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errCertificateAuthority)
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
