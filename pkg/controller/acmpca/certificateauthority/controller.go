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
	awsacmpcatypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/acmpca/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acmpca"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an ACMPCA resource"
	errGet              = "failed to get ACMPCA with name"
	errCreate           = "failed to create the ACMPCA resource"
	errDelete           = "failed to delete the ACMPCA resource"
	errEmpty            = "empty ACMPCA received from ACMPCA API"

	errKubeUpdateFailed = "cannot late initialize ACMPCA"

	errAddTagsFailed        = "cannot add tags to ACMPCA"
	errListTagsFailed       = "failed to list tags for ACMPCA"
	errRemoveTagsFailed     = "failed to remove tags for ACMPCA"
	errCertificateAuthority = "failed to update the ACMPCA resource"
)

// SetupCertificateAuthority adds a controller that reconciles ACMPCA.
func SetupCertificateAuthority(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.CertificateAuthorityGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acmpca.NewClient}),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CertificateAuthorityGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.CertificateAuthority{}).
		Complete(r)
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) acmpca.Client
}

func (conn *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.CertificateAuthority)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, conn.client, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{conn.newClientFn(cfg), conn.client}, nil
}

type external struct {
	client acmpca.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.CertificateAuthority)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeCertificateAuthority(ctx, &awsacmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errGet)
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
	if certificateAuthority.Status == awsacmpcatypes.CertificateAuthorityStatusActive {
		cr.SetConditions(xpv1.Available())
	}

	cr.Status.AtProvider = acmpca.GenerateCertificateAuthorityExternalStatus(certificateAuthority)

	tags, err := e.client.ListTags(ctx, &awsacmpca.ListTagsInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errListTagsFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: acmpca.IsCertificateAuthorityUpToDate(cr, certificateAuthority, tags.Tags),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1beta1.CertificateAuthority)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.CreateCertificateAuthority(ctx, acmpca.GenerateCreateCertificateAuthorityInput(&cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.ToString(response.CertificateAuthorityArn))
	return managed.ExternalCreation{}, nil

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {

	cr, ok := mgd.(*v1beta1.CertificateAuthority)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Update the Certificate Authority tags
	if len(cr.Spec.ForProvider.Tags) > 0 {
		tags := make([]awsacmpcatypes.Tag, len(cr.Spec.ForProvider.Tags))
		for i, t := range cr.Spec.ForProvider.Tags {
			tag := t
			tags[i] = awsacmpcatypes.Tag{Key: &tag.Key, Value: &tag.Value}
		}
		currentTags, err := e.client.ListTags(ctx, &awsacmpca.ListTagsInput{
			CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errListTagsFailed)
		}
		if len(tags) != len(currentTags.Tags) {
			_, err := e.client.UntagCertificateAuthority(ctx, &awsacmpca.UntagCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Tags:                    currentTags.Tags,
			})
			if err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errRemoveTagsFailed)
			}
		}
		_, err = e.client.TagCertificateAuthority(ctx, &awsacmpca.TagCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
			Tags:                    tags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddTagsFailed)
		}
	}

	// Update Certificate Authority configuration
	_, err := e.client.UpdateCertificateAuthority(ctx, &awsacmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
		RevocationConfiguration: acmpca.GenerateRevocationConfiguration(cr.Spec.ForProvider.RevocationConfiguration),
		Status:                  awsacmpcatypes.CertificateAuthorityStatus(aws.ToString(cr.Spec.ForProvider.Status)),
	})

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errCertificateAuthority)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.CertificateAuthority)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := e.client.DescribeCertificateAuthority(ctx, &awsacmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
	}

	if response != nil {
		if response.CertificateAuthority.Status == awsacmpcatypes.CertificateAuthorityStatusActive {
			_, err = e.client.UpdateCertificateAuthority(ctx, &awsacmpca.UpdateCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(meta.GetExternalName(cr)),
				Status:                  awsacmpcatypes.CertificateAuthorityStatusDisabled,
			})

			if err != nil {
				return errorutils.Wrap(err, errDelete)
			}
		}
	}

	_, err = e.client.DeleteCertificateAuthority(ctx, &awsacmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn:     aws.String(meta.GetExternalName(cr)),
		PermanentDeletionTimeInDays: cr.Spec.ForProvider.PermanentDeletionTimeInDays,
	})

	return errorutils.Wrap(resource.Ignore(acmpca.IsErrorNotFound, err), errDelete)
}
