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

package replicationgroup

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	awselasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
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
	errUpdateReplicationGroupCR            = "cannot update ReplicationGroup Custom Resource"
	errGetCacheClusterList                 = "cannot get cache cluster list"
	errNotReplicationGroup                 = "managed resource is not an ElastiCache replication group"
	errDescribeReplicationGroup            = "cannot describe ElastiCache replication group"
	errGenerateAuthToken                   = "cannot generate ElastiCache auth token"
	errGetPasswordSecret                   = "cannot get ElastiCache auth password from secret"
	errCreateReplicationGroup              = "cannot create ElastiCache replication group"
	errModifyReplicationGroup              = "cannot modify ElastiCache replication group"
	errDeleteReplicationGroup              = "cannot delete ElastiCache replication group"
	errModifyReplicationGroupSC            = "cannot modify ElastiCache replication group shard configuration"
	errModifyReplicationGroupCC            = "cannot modify ElastiCache replication group num cache clusters"
	errListReplicationGroupTags            = "cannot list ElastiCache replication group tags"
	errUpdateReplicationGroupTags          = "cannot update ElastiCache replication group tags"
	errReplicationGroupCacheClusterMinimum = "at least 1 replica is required"
	errReplicationGroupCacheClusterMaximum = "maximum of 5 replicas are allowed"
	errVersionInput                        = "unable to parse version number"
)

// SetupReplicationGroup adds a controller that reconciles ReplicationGroups.
func SetupReplicationGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.ReplicationGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elasticache.NewClient}),
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
		resource.ManagedKind(v1beta1.ReplicationGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.ReplicationGroup{}).
		Complete(r)
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
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{&cache{}, c.newClientFn(*cfg), c.kube}, nil
}

type external struct {
	cache  *cache
	client elasticache.Client
	kube   client.Client
}

type cache struct {
	currentAuthToken string
	desiredAuthToken string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotReplicationGroup)
	}

	rsp, err := e.client.DescribeReplicationGroups(ctx, elasticache.NewDescribeReplicationGroupsInput(meta.GetExternalName(cr)))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDescribeReplicationGroup)
	}
	// DescribeReplicationGroups can return one or many replication groups. We
	// ask for one group by name, so we should get either a single element list
	// or an error.
	rg := rsp.ReplicationGroups[0]
	ccList, err := getCacheClusterList(ctx, e.client, rg.MemberClusters)
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errGetCacheClusterList)
	}
	var oneCC awselasticachetypes.CacheCluster
	if len(ccList) > 0 {
		oneCC = ccList[0]
	}

	// keep track of last applied auth token strategy from the previous reconciliation loop
	lastAppliedAuthTokenUpdateStrategy := cr.Status.AtProvider.DeepCopy().LastAppliedAuthTokenUpdateStrategy

	current := cr.Spec.ForProvider.DeepCopy()
	elasticache.LateInitialize(&cr.Spec.ForProvider, rg, oneCC)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateReplicationGroupCR)
		}
	}
	cr.Status.AtProvider = elasticache.GenerateObservation(rg)
	// AuthTokenUpdateStrategy is not a declarative parameter from the AWS API's perspective,
	// so there's no API call to determine whether it's up to date.
	// We store the last applied strategy in status to compare it to the desired value in spec.
	cr.Status.AtProvider.LastAppliedAuthTokenUpdateStrategy = lastAppliedAuthTokenUpdateStrategy

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
			return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(elasticache.IsNotFound, err), errListReplicationGroupTags)
		}
		tagsNeedUpdate = elasticache.ReplicationGroupTagsNeedsUpdate(cr.Spec.ForProvider.Tags, tags.TagList)
	}

	rgDiff := elasticache.ReplicationGroupNeedsUpdate(cr.Spec.ForProvider, rg, ccList)
	// retrieve current pass from the connection secret
	crPass, err := getCurrentPassword(ctx, e.kube, cr)
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errGetPasswordSecret)
	}
	e.cache.currentAuthToken = crPass
	// retrieve desired pass from the spec secret ref(if exists)
	if cr.GetAuthTokenSecretRef() != nil {
		dsPass, err := getSecretPassword(ctx, e.kube, cr.GetAuthTokenSecretRef())
		e.cache.desiredAuthToken = dsPass
		if err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(err, errGetPasswordSecret)
		}
	}
	// If authTokenSecretRef is not set and auth is enabled, we want to keep using the
	// current password.
	if e.cache.desiredAuthToken == "" && aws.ToBool(cr.Spec.ForProvider.AuthEnabled) {
		e.cache.desiredAuthToken = e.cache.currentAuthToken
	}

	pwChanged := e.cache.currentAuthToken != e.cache.desiredAuthToken || cr.Status.AtProvider.LastAppliedAuthTokenUpdateStrategy != cr.Spec.ForProvider.AuthTokenUpdateStrategy
	diff := fmt.Sprintf("ReplicationGroup diff: '%s', Tags need update: %t, AuthToken changed: %t", rgDiff, tagsNeedUpdate, pwChanged)

	return managed.ExternalObservation{
		ResourceExists: true,
		ResourceUpToDate: rgDiff == "" &&
			!elasticache.ReplicationGroupShardConfigurationNeedsUpdate(cr.Spec.ForProvider, rg) &&
			!tagsNeedUpdate &&
			!pwChanged,
		ConnectionDetails: elasticache.ConnectionEndpoint(rg),
		Diff:              diff,
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
	var token string
	if cr.Spec.ForProvider.AuthTokenSecretRef != nil {
		t, err := getSecretPassword(ctx, e.kube, cr.GetAuthTokenSecretRef())
		if err != nil {
			return managed.ExternalCreation{}, errorutils.Wrap(err, errGetPasswordSecret)
		}
		token = t
	} else if aws.ToBool(cr.Spec.ForProvider.AuthEnabled) {
		t, err := password.Generate()
		if err != nil {
			return managed.ExternalCreation{}, errorutils.Wrap(err, errGenerateAuthToken)
		}
		token = t
	}
	_, err := e.client.CreateReplicationGroup(ctx, elasticache.NewCreateReplicationGroupInput(cr.Spec.ForProvider, meta.GetExternalName(cr), &token))
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateReplicationGroup)
	}
	if token != "" {
		return managed.ExternalCreation{

			ConnectionDetails: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretPasswordKey: []byte(token),
			},
		}, nil
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotReplicationGroup)
	}
	// NOTE(muvaf): AWS API rejects modification requests if the state is not
	// `available`
	if cr.Status.AtProvider.Status != v1beta1.StatusAvailable {
		return managed.ExternalUpdate{}, nil
	}

	// updates the engine version to the required format
	var version *string
	version, err := getVersion(cr.Spec.ForProvider.EngineVersion)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errVersionInput)
	}
	cr.Spec.ForProvider.EngineVersion = version

	rsp, err := e.client.DescribeReplicationGroups(ctx, elasticache.NewDescribeReplicationGroupsInput(meta.GetExternalName(cr)))
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errDescribeReplicationGroup)
	}
	rg := rsp.ReplicationGroups[0]

	if elasticache.ReplicationGroupShardConfigurationNeedsUpdate(cr.Spec.ForProvider, rg) {
		_, err = e.client.ModifyReplicationGroupShardConfiguration(ctx, elasticache.NewModifyReplicationGroupShardConfigurationInput(cr.Spec.ForProvider, meta.GetExternalName(cr), rg))
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyReplicationGroupSC)
		}
		// we can only do one change at a time, so we'll have to return early here
		return managed.ExternalUpdate{}, nil
	}

	ccList, err := getCacheClusterList(ctx, e.client, rg.MemberClusters)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errGetCacheClusterList)
	}

	if elasticache.ReplicationGroupNumCacheClustersNeedsUpdate(cr.Spec.ForProvider, ccList) {
		err := e.updateReplicationGroupNumCacheClusters(ctx, meta.GetExternalName(cr), len(ccList), aws.ToInt(cr.Spec.ForProvider.NumCacheClusters))
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyReplicationGroupCC)
		}
		return managed.ExternalUpdate{}, nil
	}

	if diff := elasticache.ReplicationGroupNeedsUpdate(cr.Spec.ForProvider, rg, ccList); diff != "" ||
		e.cache.currentAuthToken != e.cache.desiredAuthToken ||
		cr.Status.AtProvider.LastAppliedAuthTokenUpdateStrategy != cr.Spec.ForProvider.AuthTokenUpdateStrategy {

		_, err = e.client.ModifyReplicationGroup(ctx,
			elasticache.NewModifyReplicationGroupInput(alignLogDeliveryConfigurations(cr, rg), meta.GetExternalName(cr), &e.cache.desiredAuthToken),
		)
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyReplicationGroup)
		}
		cr.Status.AtProvider.LastAppliedAuthTokenUpdateStrategy = cr.Spec.ForProvider.AuthTokenUpdateStrategy
	}
	return managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(e.cache.desiredAuthToken),
	}}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotReplicationGroup)
	}
	mg.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == v1beta1.StatusDeleting {
		return managed.ExternalDelete{}, nil
	}
	_, err := e.client.DeleteReplicationGroup(ctx, elasticache.NewDeleteReplicationGroupInput(meta.GetExternalName(cr)))
	return managed.ExternalDelete{}, errorutils.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteReplicationGroup)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

func (e *external) updateTags(ctx context.Context, tags []v1beta1.Tag, arn *string) error {
	resp, err := e.client.ListTagsForResource(ctx, elasticache.NewListTagsForResourceInput(arn))
	if err != nil {
		return errorutils.Wrap(err, errListReplicationGroupTags)
	}
	add, remove := elasticache.DiffTags(tags, resp.TagList)
	if len(remove) != 0 {
		if _, err := e.client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{ResourceName: arn, TagKeys: remove}); err != nil {
			return errorutils.Wrap(err, errUpdateReplicationGroupTags)
		}
	}
	if len(add) != 0 {
		addTags := []awselasticachetypes.Tag{}
		for k, v := range add {
			addTags = append(addTags, awselasticachetypes.Tag{Key: aws.String(k), Value: aws.String(v)})
		}
		if _, err := e.client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{ResourceName: arn, Tags: addTags}); err != nil {
			return errorutils.Wrap(err, errUpdateReplicationGroupTags)
		}
	}
	return nil
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
		input := elasticache.NewIncreaseReplicaCountInput(replicaGroup, pointer.ToIntAsInt32(newReplicaCount))
		_, err := e.client.IncreaseReplicaCount(ctx, input)
		return err
	case desiredClusterSize < existingClusterSize:
		input := elasticache.NewDecreaseReplicaCountInput(replicaGroup, pointer.ToIntAsInt32(newReplicaCount))
		_, err := e.client.DecreaseReplicaCount(ctx, input)
		return err
	default:
		return nil
	}
}

func getVersion(version *string) (*string, error) {
	versionSplit := strings.Split(aws.ToString(version), ".")
	version1, err := strconv.Atoi(versionSplit[0])
	if err != nil {
		return nil, errors.Wrap(err, errVersionInput)
	}
	versionOut := strconv.Itoa(version1)
	if len(versionSplit) > 1 {
		if versionSplit[1] != "x" {
			version2, err := strconv.Atoi(versionSplit[1])
			if err != nil {
				return nil, errors.Wrap(err, errVersionInput)
			}
			versionOut += "." + strconv.Itoa(version2)
		} else {
			versionOut += ".x"
		}
	}
	return &versionOut, nil
}

// GetSecretValue retrieves the value of a secret key from a Kubernetes Secret.
func getSecretPassword(ctx context.Context, kube client.Client, sks *xpv1.SecretKeySelector) (pw string, err error) {
	secret := new(corev1.Secret)
	secret.SetName(sks.Name) // TODO(teeverr): it is needed only for mock testing functions, will be overwritten by Get. didn't find a better way.
	ref := sks.SecretReference
	err = kube.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, secret)
	if err != nil {
		return "", err
	}
	pwdRaw := secret.Data[sks.Key]
	return string(pwdRaw), nil
}

// GetCurrentPassword retrieves the current password from the connection secret of the managed resource.
func getCurrentPassword(ctx context.Context, kube client.Client, mg resource.Managed) (pw string, err error) {
	secretKeyRef := &xpv1.SecretKeySelector{
		SecretReference: *mg.GetWriteConnectionSecretToReference(),
		Key:             xpv1.ResourceCredentialsSecretPasswordKey,
	}
	return getSecretPassword(ctx, kube, secretKeyRef)
}

// alignLogDeliveryConfigurations aligns desired log delivery configuration based on external configuration before converting to aws sdk input
func alignLogDeliveryConfigurations(desired *v1beta1.ReplicationGroup, resp awselasticachetypes.ReplicationGroup) v1beta1.ReplicationGroupParameters { //nolint: gocyclo
	ds := desired.Spec.ForProvider.DeepCopy()
	crLDC := elasticache.GenerateLogDeliveryConfigurations(resp.LogDeliveryConfigurations)

	// If in the desired state log delivery configuration is nil but in the current state it is enabled,
	// we need to explicitly set it to disabled to avoid aws sdk ignoring the change
	if ds.LogDeliveryConfiguration.EngineLogs == nil && crLDC.EngineLogs != nil && crLDC.EngineLogs.Enabled != nil && *crLDC.EngineLogs.Enabled {
		ds.LogDeliveryConfiguration.EngineLogs = &v1beta1.LogsConf{Enabled: aws.Bool(false)}
	}
	if ds.LogDeliveryConfiguration.SlowLogs == nil && crLDC.SlowLogs != nil && crLDC.SlowLogs.Enabled != nil && *crLDC.SlowLogs.Enabled {
		ds.LogDeliveryConfiguration.SlowLogs = &v1beta1.LogsConf{Enabled: aws.Bool(false)}
	}

	// If log delivery is explicitly disabled in the desired config but the current state is nil,
	// unset the desired configuration. Otherwise, the AWS API fails because this update isn't idempotent
	// when the log delivery config is not set on the AWS side.
	if (ds.LogDeliveryConfiguration.EngineLogs != nil && ds.LogDeliveryConfiguration.EngineLogs.Enabled != nil && !*ds.LogDeliveryConfiguration.EngineLogs.Enabled) && crLDC.EngineLogs == nil {
		ds.LogDeliveryConfiguration.EngineLogs = &v1beta1.LogsConf{}
	}
	if (ds.LogDeliveryConfiguration.SlowLogs != nil && ds.LogDeliveryConfiguration.SlowLogs.Enabled != nil && !*ds.LogDeliveryConfiguration.SlowLogs.Enabled) && crLDC.SlowLogs == nil {
		ds.LogDeliveryConfiguration.SlowLogs = &v1beta1.LogsConf{}
	}
	return *ds
}
