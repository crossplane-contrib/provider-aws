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

package acm

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacm "github.com/aws/aws-sdk-go-v2/service/acm"
	awsacmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/google/go-cmp/cmp"
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

	"github.com/crossplane/provider-aws/apis/acm/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/acm"
)

const (
	errUnexpectedObject = "The managed resource is not an ACM resource"
	errGet              = "failed to get Certificate with name"
	errCreate           = "failed to create the Certificate resource"
	errDelete           = "failed to delete the Certificate resource"
	errUpdate           = "failed to update the Certificate resource"
	errSDK              = "empty Certificate received from ACM API"

	errKubeUpdateFailed = "cannot late initialize Certificate"

	errAddTagsFailed        = "cannot add tags to Certificate"
	errListTagsFailed       = "failed to list tags for Certificate"
	errRemoveTagsFailed     = "failed to remove tags for Certificate"
	errRenewalFailed        = "failed to renew Certificate"
	errIneligibleForRenewal = "ineligible to renew Certificate"
)

// SetupCertificate adds a controller that reconciles Certificates.
func SetupCertificate(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.CertificateGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.Certificate{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CertificateGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acm.NewClient}),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(aws.Config) acm.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Certificate)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.client, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(*cfg), c.client}, nil
}

type external struct {
	client acm.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeCertificate(ctx, &awsacm.DescribeCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errGet)
	}

	if response.Certificate == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	certificate := *response.Certificate
	current := cr.Spec.ForProvider.DeepCopy()
	acm.LateInitializeCertificate(&cr.Spec.ForProvider, &certificate)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	if certificate.Status == awsacmtypes.CertificateStatusIssued {
		cr.SetConditions(xpv1.Available())
	}

	cr.Status.AtProvider = acm.GenerateCertificateStatus(certificate)

	tags, err := e.client.ListTagsForCertificate(ctx, &awsacm.ListTagsForCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errListTagsFailed)
	}

	return managed.ExternalObservation{
		ResourceUpToDate: acm.IsCertificateUpToDate(cr.Spec.ForProvider, certificate, tags.Tags),
		ResourceExists:   true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.RequestCertificate(ctx, acm.GenerateCreateCertificateInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.ToString(response.CertificateArn))
	return managed.ExternalCreation{}, nil

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo

	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Update Certificate tags
	if len(cr.Spec.ForProvider.Tags) > 0 {

		desiredTags := make([]awsacmtypes.Tag, len(cr.Spec.ForProvider.Tags))
		for i, t := range cr.Spec.ForProvider.Tags {
			desiredTags[i] = awsacmtypes.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}

		currentTags, err := e.client.ListTagsForCertificate(ctx, &awsacm.ListTagsForCertificateInput{
			CertificateArn: aws.String(meta.GetExternalName(cr)),
		})

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errListTagsFailed)
		}

		if len(desiredTags) != len(currentTags.Tags) {
			_, err := e.client.RemoveTagsFromCertificate(ctx, &awsacm.RemoveTagsFromCertificateInput{
				CertificateArn: aws.String(meta.GetExternalName(cr)),
				Tags:           currentTags.Tags,
			})
			if err != nil {
				return managed.ExternalUpdate{}, awsclient.Wrap(err, errRemoveTagsFailed)
			}
		}
		_, err = e.client.AddTagsToCertificate(ctx, &awsacm.AddTagsToCertificateInput{
			CertificateArn: aws.String(meta.GetExternalName(cr)),
			Tags:           desiredTags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errAddTagsFailed)
		}
	}

	// the UpdateCertificateOptions command is not permitted for private certificates.
	if cr.Status.AtProvider.Type != awsacmtypes.CertificateTypePrivate {
		// Update the Certificate Option
		if cr.Spec.ForProvider.CertificateTransparencyLoggingPreference != nil {
			_, err := e.client.UpdateCertificateOptions(ctx, &awsacm.UpdateCertificateOptionsInput{
				CertificateArn: aws.String(meta.GetExternalName(cr)),
				Options:        &awsacmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: *cr.Spec.ForProvider.CertificateTransparencyLoggingPreference},
			})

			if err != nil {
				return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
			}
		}
	}

	// Renew the certificate if request for RenewCertificate and Certificate is eligible
	if aws.ToBool(cr.Spec.ForProvider.RenewCertificate) {
		if cr.Status.AtProvider.RenewalEligibility == awsacmtypes.RenewalEligibilityEligible {
			_, err := e.client.RenewCertificate(ctx, &awsacm.RenewCertificateInput{
				CertificateArn: aws.String(meta.GetExternalName(cr)),
			})

			if err != nil {
				return managed.ExternalUpdate{}, awsclient.Wrap(err, errRenewalFailed)
			}
		}
		cr.Spec.ForProvider.RenewCertificate = aws.Bool(false)
		return managed.ExternalUpdate{}, errors.New(errIneligibleForRenewal)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.Certificate)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteCertificate(ctx, &awsacm.DeleteCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errDelete)
}
