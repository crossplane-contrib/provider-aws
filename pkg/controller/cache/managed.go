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

package cache

import (
	"context"
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	awselasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// Error strings.
const (
	errUpdateReplicationGroupCR            = "cannot update ReplicationGroup Custom Resource"
	errGetCacheClusterList                 = "cannot get cache cluster list"
	errNotReplicationGroup                 = "managed resource is not an ElastiCache replication group"
	errDescribeReplicationGroup            = "cannot describe ElastiCache replication group"
	errGenerateAuthToken                   = "cannot generate ElastiCache auth token"
	errCreateReplicationGroup              = "cannot create ElastiCache replication group"
	errModifyReplicationGroup              = "cannot modify ElastiCache replication group"
	errDeleteReplicationGroup              = "cannot delete ElastiCache replication group"
	errModifyReplicationGroupSC            = "cannot modify ElastiCache replication group shard configuration"
	errModifyReplicationGroupCC            = "cannot modify ElastiCache replication group num cache clusters"
	errListReplicationGroupTags            = "cannot list ElastiCache replication group tags"
	errUpdateReplicationGroupTags          = "cannot update ElastiCache replication group tags"
	errReplicationGroupCacheClusterMinimum = "at least 1 replica is required"
	errReplicationGroupCacheClusterMaximum = "maximum of 5 replicas are allowed"
)

// SetupReplicationGroup adds a controller that reconciles ReplicationGroups.
func SetupReplicationGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.ReplicationGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.ReplicationGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.ReplicationGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elasticache.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return nil, errors.New(errNotReplicationGroup)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(*cfg), c.kube}, nil
}

type external struct {
	client elasticache.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotReplicationGroup)
	}

	rsp, err := e.client.DescribeReplicationGroups(ctx, elasticache.NewDescribeReplicationGroupsInput(meta.GetExternalName(cr)))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, awsclient.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDescribeReplicationGroup)
	}
	// DescribeReplicationGroups can return one or many replication groups. We
	// ask for one group by name, so we should get either a single element list
	// or an error.
	rg := rsp.ReplicationGroups[0]
	ccList, err := getCacheClusterList(ctx, e.client, rg.MemberClusters)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errGetCacheClusterList)
	}
	var oneCC awselasticachetypes.CacheCluster
	if len(ccList) > 0 {
		oneCC = ccList[0]
	}

	current := cr.Spec.ForProvider.DeepCopy()
	elasticache.LateInitialize(&cr.Spec.ForProvider, rg, oneCC)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateReplicationGroupCR)
		}
	}
	cr.Status.AtProvider = elasticache.GenerateObservation(rg)

	switch cr.Status.AtProvider.Status {
	case v1beta1.StatusAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.StatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.StatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	var tagsNeedUpdate bool
	if cr.Status.AtProvider.Status == v1beta1.StatusAvailable {
		tags, err := e.client.ListTagsForResource(ctx, elasticache.NewListTagsForResourceInput(rg.ARN))
		if err != nil {
			return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(elasticache.IsNotFound, err), errListReplicationGroupTags)
		}
		tagsNeedUpdate = elasticache.ReplicationGroupTagsNeedsUpdate(cr.Spec.ForProvider.Tags, tags.TagList)
	}

	rgDiff := elasticache.ReplicationGroupNeedsUpdate(cr.Spec.ForProvider, rg, ccList)
	return managed.ExternalObservation{
		ResourceExists: true,
		ResourceUpToDate: rgDiff == "" &&
			!elasticache.ReplicationGroupShardConfigurationNeedsUpdate(cr.Spec.ForProvider, rg) &&
			!tagsNeedUpdate,
		ConnectionDetails: elasticache.ConnectionEndpoint(rg),
		Diff:              rgDiff,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotReplicationGroup)
	}

	cr.Status.SetConditions(xpv1.Creating())
	// Our create request will fail if auth is enabled but transit encryption is
	// not. We don't check for the latter here because it's less surprising to
	// submit the request as the operator intended and let the reconcile fail
	// with an explanatory message from AWS explaining that transit encryption
	// is required.
	var token *string
	if aws.ToBool(cr.Spec.ForProvider.AuthEnabled) {
		t, err := password.Generate()
		if err != nil {
			return managed.ExternalCreation{}, awsclient.Wrap(err, errGenerateAuthToken)
		}
		token = &t
	}
	_, err := e.client.CreateReplicationGroup(ctx, elasticache.NewCreateReplicationGroupInput(cr.Spec.ForProvider, meta.GetExternalName(cr), token))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateReplicationGroup)
	}
	if token != nil {
		return managed.ExternalCreation{
			ConnectionDetails: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretPasswordKey: []byte(*token),
			},
		}, nil
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotReplicationGroup)
	}
	// NOTE(muvaf): AWS API rejects modification requests if the state is not
	// `available`
	if cr.Status.AtProvider.Status != v1beta1.StatusAvailable {
		return managed.ExternalUpdate{}, nil
	}

	rsp, err := e.client.DescribeReplicationGroups(ctx, elasticache.NewDescribeReplicationGroupsInput(meta.GetExternalName(cr)))
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errDescribeReplicationGroup)
	}
	rg := rsp.ReplicationGroups[0]

	if elasticache.ReplicationGroupShardConfigurationNeedsUpdate(cr.Spec.ForProvider, rg) {
		_, err = e.client.ModifyReplicationGroupShardConfiguration(ctx, elasticache.NewModifyReplicationGroupShardConfigurationInput(cr.Spec.ForProvider, meta.GetExternalName(cr), rg))
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyReplicationGroupSC)
		}
		// we can only do one change at a time, so we'll have to return early here
		return managed.ExternalUpdate{}, nil
	}

	ccList, err := getCacheClusterList(ctx, e.client, rg.MemberClusters)
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errGetCacheClusterList)
	}

	if elasticache.ReplicationGroupNumCacheClustersNeedsUpdate(cr.Spec.ForProvider, ccList) {
		err := e.updateReplicationGroupNumCacheClusters(ctx, meta.GetExternalName(cr), len(ccList), aws.ToInt(cr.Spec.ForProvider.NumCacheClusters))
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyReplicationGroupCC)
		}
		return managed.ExternalUpdate{}, nil
	}

	if diff := elasticache.ReplicationGroupNeedsUpdate(cr.Spec.ForProvider, rg, ccList); diff != "" {
		_, err = e.client.ModifyReplicationGroup(ctx, elasticache.NewModifyReplicationGroupInput(cr.Spec.ForProvider, meta.GetExternalName(cr)))
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyReplicationGroup)
		}
		return managed.ExternalUpdate{}, nil

	}
	err = e.updateTags(ctx, cr.Spec.ForProvider.Tags, rg.ARN)
	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateReplicationGroupTags)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return errors.New(errNotReplicationGroup)
	}
	mg.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == v1beta1.StatusDeleting {
		return nil
	}
	_, err := e.client.DeleteReplicationGroup(ctx, elasticache.NewDeleteReplicationGroupInput(meta.GetExternalName(cr)))
	return awsclient.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteReplicationGroup)
}

func (e *external) updateTags(ctx context.Context, tags []v1beta1.Tag, arn *string) error {
	resp, err := e.client.ListTagsForResource(ctx, elasticache.NewListTagsForResourceInput(arn))
	if err != nil {
		return awsclient.Wrap(err, errListReplicationGroupTags)
	}
	add, remove := elasticache.DiffTags(tags, resp.TagList)
	if len(remove) != 0 {
		if _, err := e.client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{ResourceName: arn, TagKeys: remove}); err != nil {
			return awsclient.Wrap(err, errUpdateReplicationGroupTags)
		}
	}
	if len(add) != 0 {
		addTags := []awselasticachetypes.Tag{}
		for k, v := range add {
			addTags = append(addTags, awselasticachetypes.Tag{Key: aws.String(k), Value: aws.String(v)})
		}
		if _, err := e.client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{ResourceName: arn, Tags: addTags}); err != nil {
			return awsclient.Wrap(err, errUpdateReplicationGroupTags)
		}
	}
	return nil
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return errors.New(errNotReplicationGroup)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mg) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]v1beta1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = v1beta1.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errUpdateReplicationGroupCR)
}

func getCacheClusterList(ctx context.Context, client awselasticache.DescribeCacheClustersAPIClient, idList []string) ([]awselasticachetypes.CacheCluster, error) {
	if len(idList) < 1 {
		return nil, nil
	}
	ccList := make([]awselasticachetypes.CacheCluster, len(idList))
	for i, cc := range idList {
		rsp, err := client.DescribeCacheClusters(ctx, elasticache.NewDescribeCacheClustersInput(cc))
		if err != nil {
			return nil, err
		}
		ccList[i] = rsp.CacheClusters[0]
	}
	return ccList, nil
}

// updateReplicationGroupNumCacheClusters updates the number of Cache Clusters in a replica group
func (e *external) updateReplicationGroupNumCacheClusters(ctx context.Context, replicaGroup string, existingClusterSize, desiredClusterSize int) error {
	// Cache clusters consist of 1 primary and 1-5 replicas.
	// The AWS API modifies the number of replicas
	newReplicaCount := desiredClusterSize - 1
	switch {
	case newReplicaCount < 1:
		return errors.New(errReplicationGroupCacheClusterMinimum)
	case newReplicaCount > 5:
		return errors.New(errReplicationGroupCacheClusterMaximum)
	case desiredClusterSize > existingClusterSize:
		input := elasticache.NewIncreaseReplicaCountInput(replicaGroup, awsclient.Int32(newReplicaCount))
		_, err := e.client.IncreaseReplicaCount(ctx, input)
		return err
	case desiredClusterSize < existingClusterSize:
		input := elasticache.NewDecreaseReplicaCountInput(replicaGroup, awsclient.Int32(newReplicaCount))
		_, err := e.client.DecreaseReplicaCount(ctx, input)
		return err
	default:
		return nil
	}
}
