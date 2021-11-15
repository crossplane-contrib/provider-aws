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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
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

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
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
func SetupCacheSubnetGroup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.CacheSubnetGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.CacheSubnetGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CacheSubnetGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elasticache.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return nil, errors.New(errNotSubnetGroup)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg)}, nil
}

type external struct {
	client elasticache.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubnetGroup)
	}

	resp, err := e.client.DescribeCacheSubnetGroups(ctx, &awscache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: awsclient.String(meta.GetExternalName(cr)),
	})
	if err != nil || resp.CacheSubnetGroups == nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(elasticache.IsSubnetGroupNotFound, err), errDescribeSubnetGroup)
	}

	sg := resp.CacheSubnetGroups[0]

	cr.Status.AtProvider = v1alpha1.CacheSubnetGroupExternalStatus{
		VPCID: awsclient.StringValue(sg.VpcId),
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: elasticache.IsSubnetGroupUpToDate(cr.Spec.ForProvider, sg),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubnetGroup)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateCacheSubnetGroup(ctx, &awscache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: awsclient.String(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        awsclient.String(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIDs,
	})
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateSubnetGroup)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubnetGroup)
	}

	_, err := e.client.ModifyCacheSubnetGroup(ctx, &awscache.ModifyCacheSubnetGroupInput{
		CacheSubnetGroupDescription: awsclient.String(cr.Spec.ForProvider.Description),
		CacheSubnetGroupName:        awsclient.String(meta.GetExternalName(cr)),
		SubnetIds:                   cr.Spec.ForProvider.SubnetIDs,
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifySubnetGroup)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CacheSubnetGroup)
	if !ok {
		return errors.New(errNotSubnetGroup)
	}

	cr.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteCacheSubnetGroup(ctx, &awscache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: awsclient.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteSubnetGroup)
}
