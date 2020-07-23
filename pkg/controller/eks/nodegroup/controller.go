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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
)

const (
	errNotEKSNodeGroup  = "managed resource is not an EKS node group custom resource"
	errKubeUpdateFailed = "cannot update EKS node group custom resource"

	errCreateEKSClient   = "cannot create EKS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed        = "cannot create EKS node group"
	errUpdateConfigFailed  = "cannot update EKS node group configuration"
	errUpdateVersionFailed = "cannot update EKS node group version"
	errAddTagsFailed       = "cannot add tags to EKS node group"
	errDeleteFailed        = "cannot delete EKS node group"
	errDescribeFailed      = "cannot describe EKS node group"
)

// SetupNodeGroup adds a controller that reconciles NodeGroups.
func SetupNodeGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.NodeGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NodeGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.NodeGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: eks.NewClient}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (eks.Client, eks.STSClient, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.NodeGroup)
	if !ok {
		return nil, errors.New(errNotEKSNodeGroup)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.Spec.ProviderReference.Name}, p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		eksClient, stsClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: eksClient, sts: stsClient, kube: c.kube}, errors.Wrap(err, errCreateEKSClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	eksClient, stsClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: eksClient, sts: stsClient, kube: c.kube}, errors.Wrap(err, errCreateEKSClient)
}

type external struct {
	client eks.Client
	sts    eks.STSClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSNodeGroup)
	}

	rsp, err := e.client.DescribeNodegroupRequest(&awseks.DescribeNodegroupInput{NodegroupName: aws.String(meta.GetExternalName(cr)), ClusterName: cr.Spec.ForProvider.ClusterName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitializeNodeGroup(&cr.Spec.ForProvider, rsp.Nodegroup)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = eks.GenerateNodeGroupObservation(rsp.Nodegroup)
	switch cr.Status.AtProvider.Status {
	case v1alpha1.NodeGroupStatusActive:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1alpha1.NodeGroupStatusCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.NodeGroupStatusDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: eks.IsNodeGroupUpToDate(&cr.Spec.ForProvider, rsp.Nodegroup),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSNodeGroup)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.Status == v1alpha1.NodeGroupStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.CreateNodegroupRequest(eks.GenerateCreateNodeGroupInput(meta.GetExternalName(cr), &cr.Spec.ForProvider)).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.NodeGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSNodeGroup)
	}
	switch cr.Status.AtProvider.Status {
	case v1alpha1.NodeGroupStatusUpdating, v1alpha1.NodeGroupStatusCreating:
		return managed.ExternalUpdate{}, nil
	}

	// NOTE(hasheddan): we have to describe the node group again because
	// different fields require different update methods.
	rsp, err := e.client.DescribeNodegroupRequest(&awseks.DescribeNodegroupInput{NodegroupName: aws.String(meta.GetExternalName(cr)), ClusterName: cr.Spec.ForProvider.ClusterName}).Send(ctx)
	if err != nil || rsp.Nodegroup == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribeFailed)
	}
	add, remove := awsclients.DiffTags(cr.Spec.ForProvider.Tags, rsp.Nodegroup.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResourceRequest(&awseks.UntagResourceInput{ResourceArn: rsp.Nodegroup.NodegroupArn, TagKeys: remove}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResourceRequest(&awseks.TagResourceInput{ResourceArn: rsp.Nodegroup.NodegroupArn, Tags: add}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if !reflect.DeepEqual(rsp.Nodegroup.Version, cr.Spec.ForProvider.Version) {
		_, err := e.client.UpdateNodegroupVersionRequest(&awseks.UpdateNodegroupVersionInput{
			ClusterName:   cr.Spec.ForProvider.ClusterName,
			NodegroupName: awsclients.String(meta.GetExternalName(cr)),
			Version:       cr.Spec.ForProvider.Version}).Send(ctx)
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateVersionFailed)
	}
	_, err = e.client.UpdateNodegroupConfigRequest(eks.GenerateUpdateNodeGroupConfigInput(meta.GetExternalName(cr), &cr.Spec.ForProvider, rsp.Nodegroup)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateConfigFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.NodeGroup)
	if !ok {
		return errors.New(errNotEKSNodeGroup)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.Status == v1alpha1.NodeGroupStatusDeleting {
		return nil
	}
	_, err := e.client.DeleteNodegroupRequest(&awseks.DeleteNodegroupInput{NodegroupName: awsclients.String(meta.GetExternalName(cr)), ClusterName: cr.Spec.ForProvider.ClusterName}).Send(ctx)
	return errors.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.NodeGroup)
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
