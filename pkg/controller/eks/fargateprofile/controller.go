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

package fargateprofile

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
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

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errNotEKSFargateProfile = "managed resource is not an EKS fargate profile custom resource"
	errKubeUpdateFailed     = "cannot update EKS fargate profile custom resource"
	errCreateFailed         = "cannot create EKS fargate profile"
	errAddTagsFailed        = "cannot add tags to EKS fargate profile"
	errDeleteFailed         = "cannot delete EKS fargate profile"
	errDescribeFailed       = "cannot describe EKS fargate profile"
)

// SetupFargateProfile adds a controller that reconciles FargateProfiles.
func SetupFargateProfile(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.FargateProfileKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newEKSClientFn: eks.NewEKSClient}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.FargateProfileGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.FargateProfile{}).
		Complete(r)
}

type connector struct {
	kube           client.Client
	newEKSClientFn func(config aws.Config) eks.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.FargateProfile)
	if !ok {
		return nil, errors.New(errNotEKSFargateProfile)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newEKSClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client eks.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.FargateProfile)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSFargateProfile)
	}

	rsp, err := e.client.DescribeFargateProfile(ctx, &awseks.DescribeFargateProfileInput{FargateProfileName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitializeFargateProfile(&cr.Spec.ForProvider, rsp.FargateProfile)

	cr.Status.AtProvider = eks.GenerateFargateProfileObservation(rsp.FargateProfile)
	// Any of the statuses we don't explicitly address should be considered as
	// the fargate profile being unavailable.
	switch cr.Status.AtProvider.Status { //nolint:exhaustive
	case v1beta1.FargateProfileStatusActive:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.FargateProfileStatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.FargateProfileStatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        eks.IsFargateProfileUpToDate(cr.Spec.ForProvider, rsp.FargateProfile),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.FargateProfile)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSFargateProfile)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.Status == v1beta1.FargateProfileStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.CreateFargateProfile(ctx, eks.GenerateCreateFargateProfileInput(meta.GetExternalName(cr), cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.FargateProfile)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSFargateProfile)
	}

	// NOTE(knappek): we have to describe the fargate profile again because
	// a fargate profile is actually immutable and can't be updated.
	rsp, err := e.client.DescribeFargateProfile(ctx, &awseks.DescribeFargateProfileInput{FargateProfileName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	if err != nil || rsp.FargateProfile == nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errDescribeFailed)
	}
	add, remove := tags.DiffTags(cr.Spec.ForProvider.Tags, rsp.FargateProfile.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResource(ctx, &awseks.UntagResourceInput{ResourceArn: rsp.FargateProfile.FargateProfileArn, TagKeys: remove}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResource(ctx, &awseks.TagResourceInput{ResourceArn: rsp.FargateProfile.FargateProfileArn, Tags: add}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.FargateProfile)
	if !ok {
		return errors.New(errNotEKSFargateProfile)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == v1beta1.FargateProfileStatusDeleting {
		return nil
	}
	_, err := e.client.DeleteFargateProfile(ctx, &awseks.DeleteFargateProfileInput{FargateProfileName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	return errorutils.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}
