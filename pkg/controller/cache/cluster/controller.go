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

package cluster

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	elasticacheservice "github.com/aws/aws-sdk-go-v2/service/elasticache"
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
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// Error strings.
const (
	errUpdateCacheClusterCR = "cannot update Cache Cluster Custom Resource"
	errNotCacheCluster      = "managed resource is not a Cache Cluster"
	errDescribeCacheCluster = "cannot describe Cache Cluster"
	errCreateCacheCluster   = "cannot create Cache Cluster"
	errModifyCacheCluster   = "cannot modify Cache Cluster"
	errDeleteCacheCluster   = "cannot delete Cache Cluster"
)

// SetupCacheCluster adds a controller that reconciles CacheCluster.
func SetupCacheCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(cachev1alpha1.CacheClusterGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(cachev1alpha1.CacheClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&cachev1alpha1.CacheCluster{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elasticache.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*cachev1alpha1.CacheCluster)
	if !ok {
		return nil, errors.New(errNotCacheCluster)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(*cfg), c.kube}, nil
}

type external struct {
	client elasticache.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*cachev1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCacheCluster)
	}

	resp, err := e.client.DescribeCacheClusters(ctx, elasticache.NewDescribeCacheClustersInput(meta.GetExternalName(cr)))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(resource.Ignore(elasticache.IsClusterNotFound, err), errDescribeCacheCluster)
	}

	cluster := resp.CacheClusters[0]
	current := cr.Spec.ForProvider.DeepCopy()
	elasticache.LateInitializeCluster(&cr.Spec.ForProvider, cluster)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(err, errUpdateCacheClusterCR)
		}
	}

	cr.Status.AtProvider = elasticache.GenerateClusterObservation(cluster)

	switch cr.Status.AtProvider.CacheClusterStatus {
	case cachev1alpha1.StatusAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case cachev1alpha1.StatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case cachev1alpha1.StatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	upToDate, err := elasticache.IsClusterUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, &cluster)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*cachev1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCacheCluster)
	}

	cr.Status.SetConditions(xpv1.Creating())

	input, err := elasticache.GenerateCreateCacheClusterInput(cr.Spec.ForProvider, meta.GetExternalName(cr))
	if err == nil {
		_, err = e.client.CreateCacheCluster(ctx, input)
	}

	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreateCacheCluster)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*cachev1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCacheCluster)
	}

	// AWS API rejects modification requests if the state is not `available`
	if cr.Status.AtProvider.CacheClusterStatus != cachev1alpha1.StatusAvailable {
		return managed.ExternalUpdate{}, nil
	}

	clusterInput, err := elasticache.GenerateModifyCacheClusterInput(cr.Spec.ForProvider, meta.GetExternalName(cr))
	if err == nil {
		_, err = e.client.ModifyCacheCluster(ctx, clusterInput)
	}

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyCacheCluster)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*cachev1alpha1.CacheCluster)
	if !ok {
		return errors.New(errNotCacheCluster)
	}

	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.CacheClusterStatus == cachev1alpha1.StatusDeleted ||
		cr.Status.AtProvider.CacheClusterStatus == cachev1alpha1.StatusDeleting {
		return nil
	}

	_, err := e.client.DeleteCacheCluster(ctx, &elasticacheservice.DeleteCacheClusterInput{
		CacheClusterId: aws.String(meta.GetExternalName(cr)),
	})
	return errorutils.Wrap(resource.Ignore(elasticache.IsClusterNotFound, err), errDeleteCacheCluster)
}
