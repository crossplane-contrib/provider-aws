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
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/cache/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	elasticacheclient "github.com/crossplane/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache/fake"
)

const (
	namespace = "coolNamespace"
	name      = "coolGroup"

	providerName       = "cool-aws"
	providerSecretName = "cool-aws-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyini"

	connectionSecretName = "cool-connection-secret"
)

var (
	cacheNodeType            = "n1.super.cool"
	autoFailoverEnabled      = true
	cacheParameterGroupName  = "coolParamGroup"
	engineVersion            = "5.0.0"
	port                     = 6379
	host                     = "172.16.0.1"
	maintenanceWindow        = "tomorrow"
	snapshotRetentionLimit   = 1
	snapshotWindow           = "thedayaftertomorrow"
	transitEncryptionEnabled = true

	cacheClusterID = name + "-0001"

	ctx       = context.Background()
	errorBoom = errors.New("boom")

	objectMeta = metav1.ObjectMeta{Name: name}

	provider = awsv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: awsv1alpha3.ProviderSpec{
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Namespace: namespace,
						Name:      providerSecretName,
					},
					Key: providerSecretKey,
				},
			},
		},
	}

	providerSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}
)

type testCase struct {
	name         string
	e            managed.ExternalClient
	r            *v1beta1.ReplicationGroup
	want         *v1beta1.ReplicationGroup
	tokenCreated bool
	returnsErr   bool
}

type replicationGroupModifier func(*v1beta1.ReplicationGroup)

func withConditions(c ...runtimev1alpha1.Condition) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.SetBindingPhase(p) }
}

func withProviderStatus(s string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.Status = s }
}

func withReplicationGroupID(n string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { meta.SetExternalName(r, n) }
}

func withEndpoint(e string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.ConfigurationEndpoint.Address = e }
}

func withPort(p int) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.ConfigurationEndpoint.Port = p }
}

func withAuthEnabled(v bool) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.AuthEnabled = &v }
}

func withMemberClusters(members []string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.MemberClusters = members }
}

func withClusterEnabled(e bool) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.ClusterEnabled = e }
}

func withAutomaticFailoverStatus(s string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.AutomaticFailover = s }
}

func withTags(tagMaps ...map[string]string) replicationGroupModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.Tags = tagList }
}

func replicationGroup(rm ...replicationGroupModifier) *v1beta1.ReplicationGroup {
	r := &v1beta1.ReplicationGroup{
		ObjectMeta: objectMeta,
		Spec: v1beta1.ReplicationGroupSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1beta1.ReplicationGroupParameters{
				AutomaticFailoverEnabled:   &autoFailoverEnabled,
				CacheNodeType:              cacheNodeType,
				CacheParameterGroupName:    &cacheParameterGroupName,
				EngineVersion:              &engineVersion,
				PreferredMaintenanceWindow: &maintenanceWindow,
				SnapshotRetentionLimit:     &snapshotRetentionLimit,
				SnapshotWindow:             &snapshotWindow,
				TransitEncryptionEnabled:   &transitEncryptionEnabled,
			},
		},
	}
	meta.SetExternalName(r, r.Name)
	for _, m := range rm {
		m(r)
	}

	return r
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockClient{
				MockCreateReplicationGroupRequest: func(_ *elasticache.CreateReplicationGroupInput) elasticache.CreateReplicationGroupRequest {
					return elasticache.CreateReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &elasticache.CreateReplicationGroupOutput{}},
					}
				},
			}},
			r: replicationGroup(withAuthEnabled(true)),
			want: replicationGroup(
				withAuthEnabled(true),
				withConditions(runtimev1alpha1.Creating()),
				withReplicationGroupID(name),
			),
			tokenCreated: true,
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockClient{
				MockCreateReplicationGroupRequest: func(_ *elasticache.CreateReplicationGroupInput) elasticache.CreateReplicationGroupRequest {
					return elasticache.CreateReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errorBoom},
					}
				},
			}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(runtimev1alpha1.Creating()),
				withReplicationGroupID(name),
			),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			creation, err := tc.e.Create(ctx, tc.r)
			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.e.Create(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if tc.tokenCreated != (len(creation.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretPasswordKey]) != 0) {
				t.Errorf("tc.e.Create(...) token creation: want: %t got: %t", tc.tokenCreated, len(creation.ConnectionDetails) != 0)
			}
			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulObserveWhileGroupCreating",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{Status: aws.String(v1beta1.StatusCreating)}},
							},
						},
					}
				},
			}},
			r: replicationGroup(withReplicationGroupID(name)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusCreating),
				withReplicationGroupID(name),
				withConditions(runtimev1alpha1.Creating()),
			),
		},
		{
			name: "SuccessfulObserveWhileGroupDeleting",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{Status: aws.String(v1beta1.StatusDeleting)}},
							},
						},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusDeleting),
				withConditions(runtimev1alpha1.Deleting()),
			),
		},
		{
			name: "SuccessfulObserveWhileGroupModifying",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{Status: aws.String(v1beta1.StatusModifying)}},
							},
						},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
			),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusModifying),
				withReplicationGroupID(name),
				withConditions(runtimev1alpha1.Unavailable()),
			),
		},
		{
			name: "SuccessfulObserveAfterCreationCompleted",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{
									ClusterEnabled:        aws.Bool(true),
									Status:                aws.String(v1beta1.StatusAvailable),
									ConfigurationEndpoint: &elasticache.Endpoint{Address: aws.String(host), Port: aws.Int64(int64(port))},
								}},
							},
						},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withConditions(runtimev1alpha1.Creating()),
				withClusterEnabled(true),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(runtimev1alpha1.Available()),
				withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
				withEndpoint(host),
				withPort(port),
				withClusterEnabled(true),
			),
			tokenCreated: true,
		},
		{
			name: "SuccessfulObserveLateInitialized",
			e: &external{
				client: &fake.MockClient{
					MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
						return elasticache.DescribeReplicationGroupsRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &elasticache.DescribeReplicationGroupsOutput{
									ReplicationGroups: []elasticache.ReplicationGroup{
										{
											AuthTokenEnabled: aws.Bool(true),
											Status:           aws.String(v1beta1.StatusCreating),
										},
									},
								},
							},
						}
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			r: replicationGroup(withReplicationGroupID(name)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusCreating),
				withReplicationGroupID(name),
				withAuthEnabled(true),
				withConditions(runtimev1alpha1.Creating()),
			),
		},
		{
			name: "FailedObserveLateInitializeError",
			e: &external{
				client: &fake.MockClient{
					MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
						return elasticache.DescribeReplicationGroupsRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &elasticache.DescribeReplicationGroupsOutput{
									ReplicationGroups: []elasticache.ReplicationGroup{
										{
											AuthTokenEnabled: aws.Bool(true),
											Status:           aws.String(v1beta1.StatusCreating),
										},
									},
								},
							},
						}
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errorBoom),
				},
			},
			r: replicationGroup(withReplicationGroupID(name)),
			want: replicationGroup(
				withReplicationGroupID(name),
				withAuthEnabled(true)),
			returnsErr: true,
		},
		{
			name: "FailedDescribeReplicationGroups",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errorBoom},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withConditions(runtimev1alpha1.Available()),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withConditions(runtimev1alpha1.Available()),
			),
			returnsErr: true,
		},
		{
			name: "FailedDescribeCacheClusters",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{
									Status:                 aws.String(v1beta1.StatusAvailable),
									AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int64(int64(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
								}},
							},
						},
					}
				},
				MockDescribeCacheClustersRequest: func(_ *elasticache.DescribeCacheClustersInput) elasticache.DescribeCacheClustersRequest {
					return elasticache.DescribeCacheClustersRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errorBoom},
					}
				},
				MockModifyReplicationGroupRequest: func(_ *elasticache.ModifyReplicationGroupInput) elasticache.ModifyReplicationGroupRequest {
					return elasticache.ModifyReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &elasticache.ModifyReplicationGroupOutput{}},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withClusterEnabled(true),
				withMemberClusters([]string{cacheClusterID}),
				withProviderStatus(v1beta1.StatusAvailable),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withClusterEnabled(true),
				withMemberClusters([]string{cacheClusterID}),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(runtimev1alpha1.Available()),
				withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
				withAutomaticFailoverStatus(string(elasticache.AutomaticFailoverStatusEnabled)),
			),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			observation, err := tc.e.Observe(ctx, tc.r)
			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.e.Observe(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if tc.tokenCreated != (len(observation.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretEndpointKey]) != 0) {
				t.Errorf("tc.e.Observe(...) token creation: want: %t got: %t", tc.tokenCreated, len(observation.ConnectionDetails) != 0)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:       "NotInAvailableState",
			e:          &external{},
			r:          replicationGroup(withProviderStatus(v1beta1.StatusCreating)),
			want:       replicationGroup(withProviderStatus(v1beta1.StatusCreating)),
			returnsErr: false,
		},
		{
			name: "FailedModifyReplicationGroup",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroupsRequest: func(_ *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
					return elasticache.DescribeReplicationGroupsRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []elasticache.ReplicationGroup{{
									Status:                 aws.String(v1beta1.StatusAvailable),
									MemberClusters:         []string{cacheClusterID},
									AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int64(int64(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									ConfigurationEndpoint:  &elasticache.Endpoint{Address: aws.String(host), Port: aws.Int64(int64(port))},
								}},
							},
						},
					}
				},
				MockDescribeCacheClustersRequest: func(_ *elasticache.DescribeCacheClustersInput) elasticache.DescribeCacheClustersRequest {
					return elasticache.DescribeCacheClustersRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Retryer:     aws.NoOpRetryer{},
							Data: &elasticache.DescribeCacheClustersOutput{
								CacheClusters: []elasticache.CacheCluster{{
									EngineVersion:              aws.String(engineVersion),
									PreferredMaintenanceWindow: aws.String("never!"), // This field needs to be updated.
								}},
							},
						},
					}
				},
				MockModifyReplicationGroupRequest: func(_ *elasticache.ModifyReplicationGroupInput) elasticache.ModifyReplicationGroupRequest {
					return elasticache.ModifyReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errorBoom},
					}
				},
			}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(runtimev1alpha1.Available()),
				withMemberClusters([]string{cacheClusterID}),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(runtimev1alpha1.Available()),
				withMemberClusters([]string{cacheClusterID}),
			),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			update, err := tc.e.Update(ctx, tc.r)
			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.e.Update(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if tc.tokenCreated != (len(update.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretPasswordKey]) != 0) {
				t.Errorf("tc.e.Update(...) token creation: want: %t got: %t", tc.tokenCreated, len(update.ConnectionDetails) != 0)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name: "Successful",
			e: &external{client: &fake.MockClient{
				MockDeleteReplicationGroupRequest: func(_ *elasticache.DeleteReplicationGroupInput) elasticache.DeleteReplicationGroupRequest {
					return elasticache.DeleteReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &elasticache.DeleteReplicationGroupOutput{}},
					}
				},
			}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(runtimev1alpha1.Deleting()),
			),
			returnsErr: false,
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockClient{
				MockDeleteReplicationGroupRequest: func(_ *elasticache.DeleteReplicationGroupInput) elasticache.DeleteReplicationGroupRequest {
					return elasticache.DeleteReplicationGroupRequest{
						Request: &aws.Request{
							HTTPRequest: &http.Request{},
							Error: awserr.New(
								elasticache.ErrCodeReplicationGroupNotFoundFault,
								"NotFound",
								fmt.Errorf("NotFound"))},
					}
				},
			}},
			r:          replicationGroup(),
			want:       replicationGroup(withConditions(runtimev1alpha1.Deleting())),
			returnsErr: false,
		},
		{
			name: "AlreadyDeletingState",
			e:    &external{},
			r:    replicationGroup(withProviderStatus(v1beta1.StatusDeleting)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusDeleting),
				withConditions(runtimev1alpha1.Deleting())),
			returnsErr: false,
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockClient{
				MockDeleteReplicationGroupRequest: func(_ *elasticache.DeleteReplicationGroupInput) elasticache.DeleteReplicationGroupRequest {
					return elasticache.DeleteReplicationGroupRequest{
						Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errorBoom},
					}
				},
			}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(runtimev1alpha1.Deleting()),
			),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.Delete(ctx, tc.r)

			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.csd.Delete(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	providerSA := func(saVal bool) awsv1alpha3.Provider {
		return awsv1alpha3.Provider{
			Spec: awsv1alpha3.ProviderSpec{
				UseServiceAccount: &saVal,
				ProviderSpec:      runtimev1alpha1.ProviderSpec{},
			},
		}
	}

	cases := []struct {
		name    string
		conn    *connecter
		i       *v1beta1.ReplicationGroup
		wantErr error
	}{
		{
			name: "SuccessfulConnect",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*awsv1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							*obj.(*corev1.Secret) = providerSecret
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return &fake.MockClient{}, nil
				},
			},
			i: replicationGroup(),
		},
		{
			name: "SuccessfulConnectWithServiceAccount",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*awsv1alpha3.Provider) = providerSA(true)
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							*obj.(*corev1.Secret) = providerSecret
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return &fake.MockClient{}, nil
				},
			},
			i: replicationGroup(),
		},
		{
			name: "FailedToGetProvider",
			conn: &connecter{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return kerrors.NewNotFound(schema.GroupResource{}, providerName)
				}},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return &fake.MockClient{}, nil
				},
			},
			i:       replicationGroup(),
			wantErr: errors.WithStack(errors.Errorf("cannot get provider:  \"%s\" not found", providerName)),
		},
		{
			name: "FailedToGetProviderSecret",
			conn: &connecter{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*awsv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return kerrors.NewNotFound(schema.GroupResource{}, providerSecretName)
					}
					return nil
				}},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return &fake.MockClient{}, nil
				},
			},
			i:       replicationGroup(),
			wantErr: errors.WithStack(errors.Errorf("cannot get provider secret:  \"%s\" not found", providerSecretName)),
		},
		{
			name: "FailedToGetProviderSecretNil",
			conn: &connecter{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*awsv1alpha3.Provider) = providerSA(false)
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return kerrors.NewNotFound(schema.GroupResource{}, providerSecretName)
					}
					return nil
				}},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return &fake.MockClient{}, nil
				},
			},
			i:       replicationGroup(),
			wantErr: errors.New("cannot get provider secret"),
		},
		{
			name: "FailedToCreateElastiCacheClient",
			conn: &connecter{
				client: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*awsv1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = providerSecret
					}
					return nil
				}},
				newClientFn: func(_ context.Context, _ []byte, _ string, _ awsclients.AuthMethod) (elasticacheclient.Client, error) {
					return nil, errorBoom
				},
			},
			i:       replicationGroup(),
			wantErr: errors.Wrap(errorBoom, errNewClient),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, gotErr := tc.conn.Connect(ctx, tc.i)
			if diff := cmp.Diff(tc.wantErr, gotErr, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1beta1.ReplicationGroup
		kube client.Client
	}
	type want struct {
		cr  *v1beta1.ReplicationGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   replicationGroup(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: replicationGroup(withTags(resource.GetExternalTags(replicationGroup()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   replicationGroup(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errorBoom)},
			},
			want: want{
				err: errors.Wrap(errorBoom, errUpdateReplicationGroupCR),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tagger{kube: tc.kube}
			err := e.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.SortSlices(func(a, b v1beta1.Tag) bool { return a.Key > b.Key })); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
