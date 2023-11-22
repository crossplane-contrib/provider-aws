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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacm "github.com/aws/aws-sdk-go-v2/service/acm"
	awsacmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
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

	"github.com/crossplane-contrib/provider-aws/apis/acm/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acm"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an ACM resource"
	errGet              = "failed to get Certificate with name"
	errCreate           = "failed to create the Certificate resource"
	errDelete           = "failed to delete the Certificate resource"
	errUpdate           = "failed to update the Certificate resource"
	errSDK              = "empty Certificate received from ACM API"

	errAddTagsFailed    = "cannot add tags to Certificate"
	errListTagsFailed   = "failed to list tags for Certificate"
	errRemoveTagsFailed = "failed to remove tags for Certificate"
)

// SetupCertificate adds a controller that reconciles Certificates.
func SetupCertificate(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.CertificateGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: acm.NewClient}),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.CertificateGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Certificate{}).
		Complete(r)
}

type connector struct {
	client      client.Client
	newClientFn func(aws.Config) acm.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Certificate)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.client, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mgd.(*v1beta1.Certificate)
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
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errGet)
	}

	if response.Certificate == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	certificate := *response.Certificate
	current := cr.Spec.ForProvider.DeepCopy()
	acm.LateInitializeCertificate(&cr.Spec.ForProvider, &certificate)
	if certificate.Status == awsacmtypes.CertificateStatusIssued {
		cr.SetConditions(xpv1.Available())
	}

	cr.Status.AtProvider = acm.GenerateCertificateStatus(certificate)

	tags, err := e.client.ListTagsForCertificate(ctx, &awsacm.ListTagsForCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errListTagsFailed)
	}

	// TODO(muvaf): We can possibly call `GetCertificate` and publish the actual
	// certificate in connection details.

	return managed.ExternalObservation{
		ResourceUpToDate:        acm.IsCertificateUpToDate(cr.Spec.ForProvider, certificate, tags.Tags),
		ResourceExists:          true,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Certificate)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	response, err := e.client.RequestCertificate(ctx, acm.GenerateCreateCertificateInput(cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.ToString(response.CertificateArn))
	return managed.ExternalCreation{}, nil

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Certificate)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	currentTags, err := e.client.ListTagsForCertificate(ctx, &awsacm.ListTagsForCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errListTagsFailed)
	}

	add, remove := acm.DiffTags(cr.Spec.ForProvider.Tags, currentTags.Tags)
	if len(remove) != 0 {
		if _, err := e.client.RemoveTagsFromCertificate(ctx, &awsacm.RemoveTagsFromCertificateInput{
			CertificateArn: aws.String(meta.GetExternalName(cr)),
			Tags:           remove,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errRemoveTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err = e.client.AddTagsToCertificate(ctx, &awsacm.AddTagsToCertificateInput{
			CertificateArn: aws.String(meta.GetExternalName(cr)),
			Tags:           add,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddTagsFailed)
		}
	}

	// the UpdateCertificateOptions command is not permitted for private certificates.
	if cr.Status.AtProvider.Type != string(awsacmtypes.CertificateTypePrivate) &&
		cr.Spec.ForProvider.Options != nil {
		if _, err := e.client.UpdateCertificateOptions(ctx, &awsacm.UpdateCertificateOptionsInput{
			CertificateArn: aws.String(meta.GetExternalName(cr)),
			Options: &awsacmtypes.CertificateOptions{
				CertificateTransparencyLoggingPreference: awsacmtypes.CertificateTransparencyLoggingPreference(cr.Spec.ForProvider.Options.CertificateTransparencyLoggingPreference),
			},
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Certificate)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	_, err := e.client.DeleteCertificate(ctx, &awsacm.DeleteCertificateInput{
		CertificateArn: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(acm.IsErrorNotFound, err), errDelete)
}
