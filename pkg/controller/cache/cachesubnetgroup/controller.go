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

package cachesubnetgroup

import (
	"context"

	commonaws "github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
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

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
)

// Error strings.
const (
	errNotSubnetGroup      = "managed resource is not a Subnet Group"
	errDescribeSubnetGroup = "cannot describe Subnet Group"
	errCreateSubnetGroup   = "cannot create Subnet Group"
	errModifySubnetGroup   = "cannot modify Subnet Group"
	errDeleteSubnetGroup   = "cannot delete Subnet Group"

	errNewClient         = "cannot create new ElastiCache client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"
)

// SetupCacheSubnetGroup adds a controller that reconciles SubnetGroups.
func SetupCacheSubnetGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CacheSubnetGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CacheSubnetGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CacheSubnetGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: elasticache.NewClient}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type connector struct {
	client      client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (elasticache.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	g, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return nil, errors.New(errNotSubnetGroup)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(g.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if commonaws.BoolValue(p.Spec.UseServiceAccount) {
		awsClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: awsClient, kube: c.client}, errors.Wrap(err, errNewClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	awsClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: awsClient, kube: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	client elasticache.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubnetGroup)
	}

	resp, err := e.client.DescribeCacheSubnetGroupsRequest(&awscache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil || resp.CacheSubnetGroups == nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(elasticache.IsSubnetGroupNotFound, err), errDescribeSubnetGroup)
	}

	sg := resp.CacheSubnetGroups[0]

	cr.Status.AtProvider = v1alpha1.CacheSubnetGroupExternalStatus{
		VpcID: aws.StringValue(sg.VpcId),
	}

	isUpToDate := elasticache.IsSubnetGroupUpToDate(cr.Spec.ForProvider, sg)

	cr.SetConditions(runtimev1alpha1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isUpToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubnetGroup)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.client.CreateCacheSubnetGroupRequest(&awscache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIds,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateSubnetGroup)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubnetGroup)
	}

	// If the list of spec.subnetIds is different from the resource subnetIds,
	// this call will return non-nil error
	_, err := e.client.ModifyCacheSubnetGroupRequest(&awscache.ModifyCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIds,
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errModifySubnetGroup)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return errors.New(errNotSubnetGroup)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteCacheSubnetGroupRequest(&awscache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteSubnetGroup)
}
