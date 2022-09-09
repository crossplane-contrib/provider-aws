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

package nodegroup

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errNotEKSNodeGroup  = "managed resource is not an EKS node group custom resource"
	errKubeUpdateFailed = "cannot update EKS node group custom resource"

	errCreateFailed        = "cannot create EKS node group"
	errUpdateConfigFailed  = "cannot update EKS node group configuration"
	errUpdateVersionFailed = "cannot update EKS node group version"
	errAddTagsFailed       = "cannot add tags to EKS node group"
	errDeleteFailed        = "cannot delete EKS node group"
	errDescribeFailed      = "cannot describe EKS node group"
)

// SetupNodeGroup adds a controller that reconciles NodeGroups.
func SetupNodeGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.NodeGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&manualv1alpha1.NodeGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(manualv1alpha1.NodeGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newEKSClientFn: eks.NewEKSClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube           client.Client
	newEKSClientFn func(config aws.Config) eks.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return nil, errors.New(errNotEKSNodeGroup)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSNodeGroup)
	}

	rsp, err := e.client.DescribeNodegroup(ctx, &awseks.DescribeNodegroupInput{NodegroupName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitializeNodeGroup(&cr.Spec.ForProvider, rsp.Nodegroup)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = eks.GenerateNodeGroupObservation(rsp.Nodegroup)
	// Any of the statuses we don't explicitly address should be considered as
	// the node group being unavailable.
	switch cr.Status.AtProvider.Status { // nolint:exhaustive
	case manualv1alpha1.NodeGroupStatusActive:
		cr.Status.SetConditions(xpv1.Available())
	case manualv1alpha1.NodeGroupStatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case manualv1alpha1.NodeGroupStatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: eks.IsNodeGroupUpToDate(&cr.Spec.ForProvider, rsp.Nodegroup),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSNodeGroup)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.Status == manualv1alpha1.NodeGroupStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.CreateNodegroup(ctx, eks.GenerateCreateNodeGroupInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSNodeGroup)
	}
	switch cr.Status.AtProvider.Status { // nolint:exhaustive
	case manualv1alpha1.NodeGroupStatusUpdating, manualv1alpha1.NodeGroupStatusCreating:
		return managed.ExternalUpdate{}, nil
	}

	// NOTE(hasheddan): we have to describe the node group again because
	// different fields require different update methods.
	rsp, err := e.client.DescribeNodegroup(ctx, &awseks.DescribeNodegroupInput{NodegroupName: aws.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	if err != nil || rsp.Nodegroup == nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errDescribeFailed)
	}
	add, remove := awsclient.DiffTags(cr.Spec.ForProvider.Tags, rsp.Nodegroup.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResource(ctx, &awseks.UntagResourceInput{ResourceArn: rsp.Nodegroup.NodegroupArn, TagKeys: remove}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResource(ctx, &awseks.TagResourceInput{ResourceArn: rsp.Nodegroup.NodegroupArn, Tags: add}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if !reflect.DeepEqual(rsp.Nodegroup.Version, cr.Spec.ForProvider.Version) {
		_, err := e.client.UpdateNodegroupVersion(ctx, &awseks.UpdateNodegroupVersionInput{
			ClusterName:   &cr.Spec.ForProvider.ClusterName,
			NodegroupName: awsclient.String(meta.GetExternalName(cr)),
			Version:       cr.Spec.ForProvider.Version})
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateVersionFailed)
	}
	_, err = e.client.UpdateNodegroupConfig(ctx, eks.GenerateUpdateNodeGroupConfigInput(meta.GetExternalName(cr), &cr.Spec.ForProvider, rsp.Nodegroup))
	return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateConfigFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return errors.New(errNotEKSNodeGroup)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == manualv1alpha1.NodeGroupStatusDeleting {
		return nil
	}
	_, err := e.client.DeleteNodegroup(ctx, &awseks.DeleteNodegroupInput{NodegroupName: awsclient.String(meta.GetExternalName(cr)), ClusterName: &cr.Spec.ForProvider.ClusterName})
	return awsclient.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*manualv1alpha1.NodeGroup)
	if !ok {
		return errors.New(errNotEKSNodeGroup)
	}
	if cr.Spec.ForProvider.Tags == nil {
		cr.Spec.ForProvider.Tags = map[string]string{}
	}
	for k, v := range resource.GetExternalTags(mg) {
		cr.Spec.ForProvider.Tags[k] = v
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
