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

	"github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cachev1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// Error strings.
const (
	errNotSubnetGroup      = "managed resource is not a Subnet Group"
	errDescribeSubnetGroup = "cannot describe Subnet Group"
	errCreateSubnetGroup   = "cannot create Subnet Group"
	errModifySubnetGroup   = "cannot modify Subnet Group"
	errDeleteSubnetGroup   = "cannot delete Subnet Group"
)

// SetupCacheSubnetGroup adds a controller that reconciles SubnetGroups.
func SetupCacheSubnetGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(cachev1alpha1.CacheSubnetGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
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
		resource.ManagedKind(cachev1alpha1.CacheSubnetGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&cachev1alpha1.CacheSubnetGroup{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elasticache.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*cachev1alpha1.CacheSubnetGroup)
	if !ok {
		return nil, errors.New(errNotSubnetGroup)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg)}, nil
}

type external struct {
	client elasticache.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*cachev1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubnetGroup)
	}

	resp, err := e.client.DescribeCacheSubnetGroups(ctx, &awscache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil || resp.CacheSubnetGroups == nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(elasticache.IsSubnetGroupNotFound, err), errDescribeSubnetGroup)
	}

	sg := resp.CacheSubnetGroups[0]

	cr.Status.AtProvider = cachev1alpha1.CacheSubnetGroupExternalStatus{
		VPCID: pointer.StringValue(sg.VpcId),
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: elasticache.IsSubnetGroupUpToDate(cr.Spec.ForProvider, sg),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*cachev1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubnetGroup)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateCacheSubnetGroup(ctx, &awscache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIDs,
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateSubnetGroup)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*cachev1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubnetGroup)
	}

	_, err := e.client.ModifyCacheSubnetGroup(ctx, &awscache.ModifyCacheSubnetGroupInput{
		CacheSubnetGroupDescription: pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIDs,
	})

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifySubnetGroup)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*cachev1alpha1.CacheSubnetGroup)
	if !ok {
		return errors.New(errNotSubnetGroup)
	}

	cr.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteCacheSubnetGroup(ctx, &awscache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteSubnetGroup)
}
