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
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
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
func SetupFargateProfile(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.FargateProfileKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.FargateProfile{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.FargateProfileGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newEKSClientFn: eks.NewEKSClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube           client.Client
	newEKSClientFn func(config aws.Config) eks.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return nil, errors.New(errNotEKSFargateProfile)
	}
	cfg, err := awsclients.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSFargateProfile)
	}

	rsp, err := e.client.DescribeFargateProfileRequest(&awseks.DescribeFargateProfileInput{FargateProfileName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitializeFargateProfile(&cr.Spec.ForProvider, rsp.FargateProfile)

	cr.Status.AtProvider = eks.GenerateFargateProfileObservation(rsp.FargateProfile)
	// Any of the statuses we don't explicitly address should be considered as
	// the fargate profile being unavailable.
	switch cr.Status.AtProvider.Status { // nolint:exhaustive
	case v1alpha1.FargateProfileStatusActive:
		cr.Status.SetConditions(runtimev1alpha1.Available())
	case v1alpha1.FargateProfileStatusCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.FargateProfileStatusDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        eks.IsFargateProfileUpToDate(cr.Spec.ForProvider, rsp.FargateProfile),
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSFargateProfile)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.Status == v1alpha1.FargateProfileStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.CreateFargateProfileRequest(eks.GenerateCreateFargateProfileInput(meta.GetExternalName(cr), cr.Spec.ForProvider)).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSFargateProfile)
	}

	// NOTE(knappek): we have to describe the fargate profile again because
	// a fargate profile is actually immutable and can't be updated.
	rsp, err := e.client.DescribeFargateProfileRequest(&awseks.DescribeFargateProfileInput{FargateProfileName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName}).Send(ctx)
	if err != nil || rsp.FargateProfile == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribeFailed)
	}
	add, remove := awsclients.DiffTags(cr.Spec.ForProvider.Tags, rsp.FargateProfile.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResourceRequest(&awseks.UntagResourceInput{ResourceArn: rsp.FargateProfile.FargateProfileArn, TagKeys: remove}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResourceRequest(&awseks.TagResourceInput{ResourceArn: rsp.FargateProfile.FargateProfileArn, Tags: add}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return errors.New(errNotEKSFargateProfile)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.Status == v1alpha1.FargateProfileStatusDeleting {
		return nil
	}
	_, err := e.client.DeleteFargateProfileRequest(&awseks.DeleteFargateProfileInput{FargateProfileName: awsclients.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName}).Send(ctx)
	return errors.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.FargateProfile)
	if !ok {
		return errors.New(errNotEKSFargateProfile)
	}
	if cr.Spec.ForProvider.Tags == nil {
		cr.Spec.ForProvider.Tags = map[string]string{}
	}
	for k, v := range resource.GetExternalTags(mg) {
		cr.Spec.ForProvider.Tags[k] = v
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
