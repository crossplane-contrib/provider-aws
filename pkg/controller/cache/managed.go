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
	"fmt"
	"reflect"
	"strings"

	commonaws "github.com/aws/aws-sdk-go-v2/aws"
	elasticacheservice "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/stack-aws/apis/cache/v1beta1"
	awsv1alpha3 "github.com/crossplaneio/stack-aws/apis/v1alpha3"
	"github.com/crossplaneio/stack-aws/pkg/clients/elasticache"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"
)

// Error strings.
const (
	errUpdateReplicationGroupCR = "cannot update ReplicationGroup Custom Resource"
	errGetCacheClusterList      = "cannot get cache cluster list"

	errNewClient                = "cannot create new ElastiCache client"
	errNotReplicationGroup      = "managed resource is not an ElastiCache replication group"
	errDescribeReplicationGroup = "cannot describe ElastiCache replication group"
	errGenerateAuthToken        = "cannot generate ElastiCache auth token"
	errCreateReplicationGroup   = "cannot create ElastiCache replication group"
	errModifyReplicationGroup   = "cannot modify ElastiCache replication group"
	errDeleteReplicationGroup   = "cannot delete ElastiCache replication group"
)

// Note this is the length of the generated random byte slice before base64
// encoding, which adds ~33% overhead.
const maxAuthTokenData = 32

// ReplicationGroupController is responsible for adding the ReplicationGroup
// controller and its corresponding reconciler to the manager with any runtime configuration.
type ReplicationGroupController struct{}

// SetupWithManager creates a new ReplicationGroup Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *ReplicationGroupController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.ReplicationGroupGroupVersionKind),
		resource.WithExternalConnecter(&connecter{
			client:      mgr.GetClient(),
			newClientFn: elasticache.NewClient,
		}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.ReplicationGroupKind, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.ReplicationGroup{}).
		Complete(r)
}

type connecter struct {
	client      client.Client
	newClientFn func(credentials []byte, region string) (elasticache.Client, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	g, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return nil, errors.New(errNotReplicationGroup)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(g.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, "cannot get provider")
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, "cannot get provider secret")
	}
	awsClient, err := c.newClientFn(s.Data[p.Spec.Secret.Key], p.Spec.Region)
	return &external{client: awsClient, kube: c.client}, errors.Wrap(err, errNewClient)
}

type external struct {
	client elasticache.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotReplicationGroup)
	}

	dr := e.client.DescribeReplicationGroupsRequest(elasticache.NewDescribeReplicationGroupsInput(meta.GetExternalName(cr)))
	dr.SetContext(ctx)
	rsp, err := dr.Send()
	if err != nil {
		return resource.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDescribeReplicationGroup)
	}
	// DescribeReplicationGroups can return one or many replication groups. We
	// ask for one group by name, so we should get either a single element list
	// or an error.
	rg := rsp.ReplicationGroups[0]
	current := cr.Spec.ForProvider.DeepCopy()
	elasticache.LateInitialize(&cr.Spec.ForProvider, rg)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errUpdateReplicationGroupCR)
		}
	}
	cr.Status.AtProvider = elasticache.GenerateObservation(rg)

	switch cr.Status.AtProvider.Status {
	case v1beta1.StatusAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.StatusCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1beta1.StatusDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	ccList, err := getCacheClusterList(ctx, e.client, cr.Status.AtProvider.MemberClusters)
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetCacheClusterList)
	}
	return resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  !elasticache.ReplicationGroupNeedsUpdate(cr.Spec.ForProvider, rg, ccList),
		ConnectionDetails: elasticache.ConnectionEndpoint(rg),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotReplicationGroup)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	// Our create request will fail if auth is enabled but transit encryption is
	// not. We don't check for the latter here because it's less surprising to
	// submit the request as the operator intended and let the reconcile fail
	// with an explanatory message from AWS explaining that transit encryption
	// is required.
	var token *string
	if commonaws.BoolValue(cr.Spec.ForProvider.AuthEnabled) {
		t, err := util.GeneratePassword(maxAuthTokenData)
		if err != nil {
			return resource.ExternalCreation{}, errors.Wrap(err, errGenerateAuthToken)
		}
		token = &t
	}
	r := e.client.CreateReplicationGroupRequest(elasticache.NewCreateReplicationGroupInput(cr.Spec.ForProvider, meta.GetExternalName(cr), token))
	r.SetContext(ctx)
	if _, err := r.Send(); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(resource.Ignore(elasticache.IsAlreadyExists, err), errCreateReplicationGroup)
	}
	if token != nil {
		return resource.ExternalCreation{
			ConnectionDetails: resource.ConnectionDetails{
				runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(*token),
			},
		}, nil
	}
	return resource.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotReplicationGroup)
	}
	// NOTE(muvaf): AWS API rejects modification requests if the state is not
	// `available`
	if cr.Status.AtProvider.Status != v1beta1.StatusAvailable {
		return resource.ExternalUpdate{}, nil
	}
	mr := e.client.ModifyReplicationGroupRequest(elasticache.NewModifyReplicationGroupInput(cr.Spec.ForProvider, meta.GetExternalName(cr)))
	mr.SetContext(ctx)
	_, err := mr.Send()
	return resource.ExternalUpdate{}, errors.Wrap(err, errModifyReplicationGroup)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.ReplicationGroup)
	if !ok {
		return errors.New(errNotReplicationGroup)
	}
	mg.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.Status == v1beta1.StatusDeleting {
		return nil
	}
	req := e.client.DeleteReplicationGroupRequest(elasticache.NewDeleteReplicationGroupInput(meta.GetExternalName(cr)))
	req.SetContext(ctx)
	_, err := req.Send()
	return errors.Wrap(resource.Ignore(elasticache.IsNotFound, err), errDeleteReplicationGroup)
}

func getCacheClusterList(ctx context.Context, client elasticache.Client, idList []string) ([]elasticacheservice.CacheCluster, error) {
	ccList := make([]elasticacheservice.CacheCluster, len(idList))
	for i, cc := range idList {
		dcc := client.DescribeCacheClustersRequest(elasticache.NewDescribeCacheClustersInput(cc))
		dcc.SetContext(ctx)
		rsp, err := dcc.Send()
		if err != nil {
			return nil, err
		}
		ccList[i] = rsp.CacheClusters[0]
	}
	return ccList, nil
}
