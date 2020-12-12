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
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
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
func SetupCacheCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CacheClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CacheCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CacheClusterGroupVersionKind),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elasticache.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CacheCluster)
	if !ok {
		return nil, errors.New(errNotCacheCluster)
	}
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mg.(*v1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCacheCluster)
	}

	resp, err := e.client.DescribeCacheClustersRequest(elasticache.NewDescribeCacheClustersInput(meta.GetExternalName(cr))).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(elasticache.IsClusterNotFound, err), errDescribeCacheCluster)
	}

	cluster := resp.CacheClusters[0]
	current := cr.Spec.ForProvider.DeepCopy()
	elasticache.LateInitializeCluster(&cr.Spec.ForProvider, cluster)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCacheClusterCR)
		}
	}

	cr.Status.AtProvider = elasticache.GenerateClusterObservation(cluster)

	switch cr.Status.AtProvider.CacheClusterStatus {
	case v1alpha1.StatusAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case v1alpha1.StatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1alpha1.StatusDeleting:
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
	cr, ok := mg.(*v1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCacheCluster)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateCacheClusterRequest(elasticache.GenerateCreateCacheClusterInput(cr.Spec.ForProvider, meta.GetExternalName(cr))).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateCacheCluster)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CacheCluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCacheCluster)
	}

	// AWS API rejects modification requests if the state is not `available`
	if cr.Status.AtProvider.CacheClusterStatus != v1alpha1.StatusAvailable {
		return managed.ExternalUpdate{}, nil
	}

	_, err := e.client.ModifyCacheClusterRequest(elasticache.GenerateModifyCacheClusterInput(cr.Spec.ForProvider, meta.GetExternalName(cr))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, errModifyCacheCluster)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CacheCluster)
	if !ok {
		return errors.New(errNotCacheCluster)
	}

	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.CacheClusterStatus == v1alpha1.StatusDeleted ||
		cr.Status.AtProvider.CacheClusterStatus == v1alpha1.StatusDeleting {
		return nil
	}

	_, err := e.client.DeleteCacheClusterRequest(&elasticacheservice.DeleteCacheClusterInput{
		CacheClusterId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(elasticache.IsClusterNotFound, err), errDeleteCacheCluster)
}
