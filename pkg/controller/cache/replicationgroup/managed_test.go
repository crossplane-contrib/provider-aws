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
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache/fake"
)

const (
	name                         = "coolGroup"
	engineVersionToTest          = "5.0.2"
	alternateEngineVersionToTest = "6.x"
)

var (
	cacheNodeType              = "n1.super.cool"
	autoFailoverEnabled        = true
	cacheParameterGroupName    = "coolParamGroup"
	engineVersion              = "5.0"
	alternateEngineVersion     = "6.x"
	port                       = 6379
	host                       = "172.16.0.1"
	maintenanceWindow          = "tomorrow"
	snapshotRetentionLimit     = 1
	snapshotWindow             = "thedayaftertomorrow"
	transitEncryptionEnabled   = true
	writeConnectionSecretToRef = xpv1.SecretReference{Name: "connectionSecretName", Namespace: "connectionSecretNamespace"}

	cacheClusterID = name + "-0001"
	cacheClusters  = []string{name + "-0001", name + "-0002", name + "-0003"}

	ctx       = context.Background()
	errorBoom = errors.New("boom")

	objectMeta = metav1.ObjectMeta{Name: name}
)

type testCase struct {
	name         string
	e            managed.ExternalClient
	r            *v1beta1.ReplicationGroup
	want         *v1beta1.ReplicationGroup
	tokenCreated bool
	returnsErr   bool
}

type testCaseIsUpToDate struct {
	name string
	e    managed.ExternalClient
	cr   *v1beta1.ReplicationGroup
	want wantIsUpToDate
}

type wantIsUpToDate struct {
	isUpToDate bool
	managed.ConnectionDetails
	err error
}

type replicationGroupModifier func(*v1beta1.ReplicationGroup)

func withAutomaticFailover(v types.AutomaticFailoverStatus) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) {
		r.Status.AtProvider.AutomaticFailover = string(v)
	}
}

func withAuthTokenSecretRef(reference *xpv1.SecretKeySelector) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.AuthTokenSecretRef = reference }
}

func withAuthTokenUpdateStrategy(s string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.AuthTokenUpdateStrategy = s }
}

func withConditions(c ...xpv1.Condition) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withProviderStatusLastAppliedAuthTokenUpdateStrategy(s string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.LastAppliedAuthTokenUpdateStrategy = s }
}

func withProviderStatus(s string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.Status = s }
}

func withProviderStatusNodeGroups(n []v1beta1.NodeGroup) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.AtProvider.NodeGroups = n }
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

func withNumNodeGroups(n int) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.NumNodeGroups = &n }
}

func withAtRestEncryptionEnabled(b bool) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.AtRestEncryptionEnabled = &b }
}

func withNumCacheClusters(n int) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.NumCacheClusters = &n }
}

func withEngineVersion(e string) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.EngineVersion = &e }
}

func withWriteConnectionSecretToReference(ref xpv1.SecretReference) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ResourceSpec.WriteConnectionSecretToReference = &ref }
}

func withLogDeliveryConfiguration(config v1beta1.LogDeliveryConfiguration) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Spec.ForProvider.LogDeliveryConfiguration = config }
}

func replicationGroup(rm ...replicationGroupModifier) *v1beta1.ReplicationGroup {
	r := &v1beta1.ReplicationGroup{
		ObjectMeta: objectMeta,
		Spec: v1beta1.ReplicationGroupSpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &writeConnectionSecretToRef,
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

func mockGettingSecretDataWithDiff(obj client.Object) error {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return errors.New("the mock function only supports secret objects")
	}
	switch obj.GetName() {
	case "authTokenSecret":
		secret.Data = map[string][]byte{
			xpv1.ResourceCredentialsSecretPasswordKey: []byte("desiredAuthToken"),
		}
	case "writeConnectionSecretToRefName":
		secret.Data = map[string][]byte{
			xpv1.ResourceCredentialsSecretPasswordKey: []byte("currentAuthToken"),
		}
	default:
		return errors.New("unexpected secret name: " + obj.GetName())

	}
	return nil
}

func mockGettingSecretData(obj client.Object) error {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return errors.New("the mock function only supports secret objects")
	}
	secret.Data = map[string][]byte{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte("token1234567890!"),
	}
	return nil
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulCreate",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockCreateReplicationGroup: func(ctx context.Context, _ *elasticache.CreateReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.CreateReplicationGroupOutput, error) {
						return &elasticache.CreateReplicationGroupOutput{}, nil
					},
				}},
			r: replicationGroup(withAuthEnabled(true)),
			want: replicationGroup(
				withAuthEnabled(true),
				withConditions(xpv1.Creating()),
				withReplicationGroupID(name),
			),
			tokenCreated: true,
		},
		{
			name: "FailedCreate",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockCreateReplicationGroup: func(ctx context.Context, _ *elasticache.CreateReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.CreateReplicationGroupOutput, error) {
						return nil, errorBoom
					},
				}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(xpv1.Creating()),
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

			if tc.tokenCreated != (len(creation.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]) != 0) {
				t.Errorf("tc.e.Create(...) token creation: want: %t got: %t", tc.tokenCreated, len(creation.ConnectionDetails) != 0)
			}
			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	var makeStringPtr = func(id string) *string {
		var p = new(string)
		*p = id
		return p
	}
	var makeBoolPtr = func(v bool) *bool {
		var p = new(bool)
		*p = v
		return p
	}
	var makeTimePtr = func(t time.Time) *time.Time {
		var p = new(time.Time)
		*p = t
		return p
	}
	var makeArn = func(id string) *string {
		return makeStringPtr("arn:aws:elasticache:eu-central-1:1001001ESOES:replicationgroup:" + id)
	}
	var makeDescription = func(id string) *string {
		return makeStringPtr("Redis Group:" + id)
	}

	var successfulObserveAfterCreationFailed = "SuccessfulObserveAfterCreationFailed"

	cases := []testCase{
		{
			name: "SuccessfulObserveWhileGroupCreating",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusCreating)}},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil),
				},
			},
			r: replicationGroup(
				withReplicationGroupID(name)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusCreating),
				withReplicationGroupID(name),
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "SuccessfulObserveWhileGroupDeleting",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusDeleting)}},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil),
				},
			},
			r: replicationGroup(
				withReplicationGroupID(name),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusDeleting),
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulObserveWhileGroupModifying",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusModifying)}},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil),
				},
			},
			r: replicationGroup(
				withReplicationGroupID(name),
			),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusModifying),
				withReplicationGroupID(name),
				withConditions(xpv1.Unavailable()),
			),
		},
		{
			name: "SuccessfulObserveAfterCreationCompleted",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								ClusterEnabled:        aws.Bool(true),
								Status:                aws.String(v1beta1.StatusAvailable),
								ConfigurationEndpoint: &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockListTagsForResource: func(ctx context.Context, _ *elasticache.ListTagsForResourceInput, opts []func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error) {
						return &elasticache.ListTagsForResourceOutput{
							TagList: []types.Tag{
								{Key: aws.String("key1"), Value: aws.String("val1")},
								{Key: aws.String("key2"), Value: aws.String("val2")},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil),
				},
			},
			r: replicationGroup(
				withReplicationGroupID(name),
				withConditions(xpv1.Creating()),
				withClusterEnabled(true),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withEndpoint(host),
				withPort(port),
				withClusterEnabled(true),
			),
			tokenCreated: true,
		},
		{
			name: successfulObserveAfterCreationFailed, // Replicates issue #1838
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								ARN:                       makeArn(successfulObserveAfterCreationFailed),
								AtRestEncryptionEnabled:   makeBoolPtr(true),
								AuthTokenEnabled:          makeBoolPtr(true),
								AutomaticFailover:         types.AutomaticFailoverStatusEnabled,
								AuthTokenLastModifiedDate: makeTimePtr(time.Date(2023, 6, 15, 12, 21, 05, 0, time.UTC)),
								DataTiering:               types.DataTieringStatusDisabled,
								Description:               makeDescription(successfulObserveAfterCreationFailed),
								MultiAZ:                   types.MultiAZStatusDisabled,
								NodeGroups: []types.NodeGroup{
									{NodeGroupId: aws.String("0001"), Status: makeStringPtr(v1beta1.StatusCreateFailed)},
									{NodeGroupId: aws.String("0002"), Status: makeStringPtr(v1beta1.StatusCreateFailed)},
								},
								ReplicationGroupCreateTime: makeTimePtr(time.Date(2023, 6, 15, 12, 21, 05, 0, time.UTC)),
								Status:                     aws.String(v1beta1.StatusCreateFailed),
								TransitEncryptionEnabled:   makeBoolPtr(true),
							}},
						}, nil
					},
					MockListTagsForResource: func(ctx context.Context, _ *elasticache.ListTagsForResourceInput, opts []func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error) {
						return &elasticache.ListTagsForResourceOutput{
							TagList: []types.Tag{
								{Key: aws.String("key1"), Value: aws.String("val1")},
								{Key: aws.String("key2"), Value: aws.String("val2")},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil),
				},
			},
			r: replicationGroup(
				withReplicationGroupID(name),
				withConditions(xpv1.Creating()),
				withClusterEnabled(true),
				withNumNodeGroups(2),
				withAtRestEncryptionEnabled(true),
				withAuthEnabled(true),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusCreateFailed),
				withConditions(xpv1.Unavailable()),
				withNumNodeGroups(2),
				withAutomaticFailover(types.AutomaticFailoverStatusEnabled),
				withProviderStatusNodeGroups([]v1beta1.NodeGroup{
					{NodeGroupID: "0001", Status: v1beta1.StatusCreateFailed},
					{NodeGroupID: "0002", Status: v1beta1.StatusCreateFailed},
				}),
				withAtRestEncryptionEnabled(true),
				withAuthEnabled(true),
			),
			tokenCreated: false,
		},
		{
			name: "SuccessfulObserveLateInitialized",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									AuthTokenEnabled: aws.Bool(true),
									Status:           aws.String(v1beta1.StatusCreating),
								},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet:    test.NewMockGetFn(nil),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			r: replicationGroup(withReplicationGroupID(name)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusCreating),
				withReplicationGroupID(name),
				withAuthEnabled(true),
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedObserveLateInitializeError",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									AuthTokenEnabled: aws.Bool(true),
									Status:           aws.String(v1beta1.StatusCreating),
								},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet:    test.NewMockGetFn(nil),
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
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return nil, errorBoom
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withConditions(xpv1.Available()),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withConditions(xpv1.Available()),
			),
			returnsErr: true,
		},
		{
			name: "FailedDescribeCacheClusters",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								MemberClusters:         []string{cacheClusterID},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return nil, errorBoom
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

			if tc.tokenCreated != (len(observation.ConnectionDetails[xpv1.ResourceCredentialsSecretEndpointKey]) != 0) {
				t.Errorf("tc.e.Observe(...) token creation: want: %t got: %t", tc.tokenCreated, len(observation.ConnectionDetails) != 0)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	cases := []testCaseIsUpToDate{
		{
			name: "PasswordChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns different auth tokens from secrets specified in spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretDataWithDiff),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret", // This secret has "desiredAuthToken"(set in mockGettingSecretDataWithDiff)
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName", // This secret has "currentAuthToken"(set in mockGettingSecretDataWithDiff)
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: false,
				err:        nil,
			},
		},
		{
			name: "PasswordIsNoTChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns the same auth token for both secrets(spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef)
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretData),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret",
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName",
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: true,
				err:        nil,
			},
		},
		{
			name: "AuthTokenAppliedStrategyChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns the same auth token for both secrets(spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef)
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretData),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret",
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withAuthTokenUpdateStrategy("SET"),
				withProviderStatusLastAppliedAuthTokenUpdateStrategy("ROTATE"),

				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName",
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: false,
				err:        nil,
			},
		},
		{
			name: "AuthTokenAppliedStrategyIsNotChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns the same auth token for both secrets(spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef)
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretData),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret",
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withAuthTokenUpdateStrategy("ROTATE"),
				withProviderStatusLastAppliedAuthTokenUpdateStrategy("ROTATE"),

				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName",
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: true,
				err:        nil,
			},
		},
		{
			name: "LogDeliveryConfigurationChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
									LogDeliveryConfigurations: []types.LogDeliveryConfiguration{
										{
											LogFormat:       "json",
											LogType:         types.LogTypeSlowLog,
											DestinationType: types.DestinationTypeCloudWatchLogs,
											DestinationDetails: &types.DestinationDetails{
												CloudWatchLogsDetails: &types.CloudWatchLogsDestinationDetails{
													LogGroup: aws.String("old-log-group"),
												},
											},
											Status: "active",
										},
										{
											LogFormat:       "text",
											LogType:         types.LogTypeEngineLog,
											DestinationType: types.DestinationTypeKinesisFirehose,
											DestinationDetails: &types.DestinationDetails{
												KinesisFirehoseDetails: &types.KinesisFirehoseDestinationDetails{
													DeliveryStream: aws.String("firehose-stream"),
												},
											},
											Status: "modifying",
										},
									},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns the same auth token for both secrets(spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef)
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretData),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret",
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withAuthTokenUpdateStrategy("ROTATE"),
				withProviderStatusLastAppliedAuthTokenUpdateStrategy("ROTATE"),

				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName",
				}),
				withLogDeliveryConfiguration(v1beta1.LogDeliveryConfiguration{
					SlowLogs: &v1beta1.LogsConf{
						DestinationDetails: &v1beta1.DestinationDetails{
							CloudWatchLogsDetails: &v1beta1.CloudWatchLogsDestinationDetails{
								LogGroup: aws.String("new-log-group"),
							},
						},
						LogFormat: "JSON",
					},
					EngineLogs: &v1beta1.LogsConf{
						DestinationDetails: &v1beta1.DestinationDetails{
							KinesisFirehoseDetails: &v1beta1.KinesisFirehoseDestinationDetails{
								DeliveryStream: aws.String("firehose-stream"),
							},
						},
						LogFormat: "TEXT",
						Enabled:   aws.Bool(true),
					},
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: false,
				err:        nil,
			},
		},
		{
			name: "LogDeliveryConfigurationIsNotChanged",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{
								{
									Status:                 aws.String(v1beta1.StatusModifying),
									AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
									CacheNodeType:          aws.String(cacheNodeType),
									SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
									SnapshotWindow:         aws.String(snapshotWindow),
									ClusterEnabled:         aws.Bool(true),
									MemberClusters:         []string{cacheClusterID},
									LogDeliveryConfigurations: []types.LogDeliveryConfiguration{
										{
											LogFormat:       "json",
											LogType:         types.LogTypeSlowLog,
											DestinationType: types.DestinationTypeCloudWatchLogs,
											DestinationDetails: &types.DestinationDetails{
												CloudWatchLogsDetails: &types.CloudWatchLogsDestinationDetails{
													LogGroup: aws.String("old-log-group"),
												},
											},
											Status: "active",
										},
										{
											LogFormat:       "text",
											LogType:         types.LogTypeEngineLog,
											DestinationType: types.DestinationTypeKinesisFirehose,
											DestinationDetails: &types.DestinationDetails{
												KinesisFirehoseDetails: &types.KinesisFirehoseDestinationDetails{
													DeliveryStream: aws.String("firehose-stream"),
												},
											},
											Status: "modifying",
										},
									},
								},
							},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion), PreferredMaintenanceWindow: aws.String(maintenanceWindow)},
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					// This mock returns the same auth token for both secrets(spec.forProvider.authTokenSecretRef and spec.writeConnectionSecretToRef)
					MockGet:    test.NewMockGetFn(nil, mockGettingSecretData),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			cr: replicationGroup(
				withReplicationGroupID(name),
				withAuthTokenSecretRef(&xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "authTokenSecret",
						Namespace: "default"},
					Key: xpv1.ResourceCredentialsSecretPasswordKey,
				}),
				withAuthTokenUpdateStrategy("ROTATE"),
				withProviderStatusLastAppliedAuthTokenUpdateStrategy("ROTATE"),

				withWriteConnectionSecretToReference(xpv1.SecretReference{
					Name: "writeConnectionSecretToRefName",
				}),
				withLogDeliveryConfiguration(v1beta1.LogDeliveryConfiguration{
					SlowLogs: &v1beta1.LogsConf{
						DestinationDetails: &v1beta1.DestinationDetails{
							CloudWatchLogsDetails: &v1beta1.CloudWatchLogsDestinationDetails{
								LogGroup: aws.String("old-log-group"),
							},
						},
						LogFormat: "JSON",
						// If Enabled is not set, it is considered true by default
						//Enabled: aws.Bool(true),
					},
					EngineLogs: &v1beta1.LogsConf{
						DestinationDetails: &v1beta1.DestinationDetails{
							KinesisFirehoseDetails: &v1beta1.KinesisFirehoseDestinationDetails{
								DeliveryStream: aws.String("firehose-stream"),
							},
						},
						LogFormat: "TEXT",
						Enabled:   aws.Bool(true),
					},
				}),
			),
			want: wantIsUpToDate{
				isUpToDate: true,
				err:        nil,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			observation, err := tc.e.Observe(ctx, tc.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.isUpToDate, observation.ResourceUpToDate); diff != "" {
				t.Error("isUpTODate status: -want, +got\n", diff, "\ndiff:", observation.Diff)
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
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         []string{cacheClusterID},
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{{
								EngineVersion:              aws.String(engineVersion),
								PreferredMaintenanceWindow: aws.String("never!"), // This field needs to be updated.
							}},
						}, nil
					},
					MockModifyReplicationGroup: func(ctx context.Context, _ *elasticache.ModifyReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupOutput, error) {
						return nil, errorBoom
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters([]string{cacheClusterID}),
				withNumCacheClusters(1),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters([]string{cacheClusterID}),
				withNumCacheClusters(1),
			),
			returnsErr: true,
		},
		{
			name: "CallsModifyReplicationGroupShardConfiguration",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         []string{cacheClusterID},
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								NodeGroups:             []types.NodeGroup{{NodeGroupId: aws.String("ng-01")}, {NodeGroupId: aws.String("ng-02")}},
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{{
								EngineVersion:              aws.String(engineVersion),
								PreferredMaintenanceWindow: aws.String("never!"), // This field needs to be updated.
							}},
						}, nil

					},
					MockModifyReplicationGroupShardConfiguration: func(ctx context.Context, _ *elasticache.ModifyReplicationGroupShardConfigurationInput, opts []func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupShardConfigurationOutput, error) {
						return nil, errorBoom
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters([]string{cacheClusterID}),
				withNumNodeGroups(3),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters([]string{cacheClusterID}),
				withNumNodeGroups(3),
			),
			returnsErr: true,
		},
		{
			name: "FailedDecreaseReplicationGroupNumCacheClusters",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         cacheClusters,
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
							},
						}, nil
					},
					MockDecreaseReplicaCount: func(ctx context.Context, _ *elasticache.DecreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error) {
						return &elasticache.DecreaseReplicaCountOutput{}, errors.New("error decreasing number of cache clusters")
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(1),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(1),
			),
			returnsErr: true,
		},
		{
			name: "DecreaseReplicationGroupNumCacheClusters",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         cacheClusters,
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
							},
						}, nil
					},
					MockDecreaseReplicaCount: func(ctx context.Context, _ *elasticache.DecreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error) {
						return &elasticache.DecreaseReplicaCountOutput{}, nil
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(2),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(2),
			),
			returnsErr: false,
		},
		{
			name: "IncreaseReplicationGroupNumCacheClusters",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         cacheClusters,
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
							},
						}, nil
					},
					MockIncreaseReplicaCount: func(ctx context.Context, _ *elasticache.IncreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
						return &elasticache.IncreaseReplicaCountOutput{}, nil
					},
				}},
			r: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			returnsErr: false,
		},
		{
			name: "IncreaseReplicationsAndCheckBehaviourVersion",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         cacheClusters,
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
							},
						}, nil
					},
					MockIncreaseReplicaCount: func(ctx context.Context, _ *elasticache.IncreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
						return &elasticache.IncreaseReplicaCountOutput{}, nil
					},
				}},
			r: replicationGroup(
				withEngineVersion(engineVersionToTest),
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			want: replicationGroup(
				withEngineVersion(engineVersion),
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			returnsErr: false,
		},
		{
			name: "IncreaseReplicationsAndCheckBehaviourVersionx",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
						return &elasticache.DescribeReplicationGroupsOutput{
							ReplicationGroups: []types.ReplicationGroup{{
								Status:                 aws.String(v1beta1.StatusAvailable),
								MemberClusters:         cacheClusters,
								AutomaticFailover:      types.AutomaticFailoverStatusEnabled,
								CacheNodeType:          aws.String(cacheNodeType),
								SnapshotRetentionLimit: aws.Int32(int32(snapshotRetentionLimit)),
								SnapshotWindow:         aws.String(snapshotWindow),
								ClusterEnabled:         aws.Bool(true),
								ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: ptr.To(int32(port))},
							}},
						}, nil
					},
					MockDescribeCacheClusters: func(ctx context.Context, _ *elasticache.DescribeCacheClustersInput, opts []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
						return &elasticache.DescribeCacheClustersOutput{
							CacheClusters: []types.CacheCluster{
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
								{EngineVersion: aws.String(engineVersion)},
							},
						}, nil
					},
					MockIncreaseReplicaCount: func(ctx context.Context, _ *elasticache.IncreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
						return &elasticache.IncreaseReplicaCountOutput{}, nil
					},
				}},
			r: replicationGroup(
				withEngineVersion(alternateEngineVersionToTest),
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			want: replicationGroup(
				withEngineVersion(alternateEngineVersion),
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
				withMemberClusters(cacheClusters),
				withNumCacheClusters(4),
			),
			returnsErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			update, err := tc.e.Update(ctx, tc.r)
			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.e.Update(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if tc.tokenCreated != (len(update.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]) != 0) {
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
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDeleteReplicationGroup: func(ctx context.Context, _ *elasticache.DeleteReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error) {
						return &elasticache.DeleteReplicationGroupOutput{}, nil
					},
				}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(xpv1.Deleting()),
			),
			returnsErr: false,
		},
		{
			name: "SuccessfulNotFound",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDeleteReplicationGroup: func(ctx context.Context, _ *elasticache.DeleteReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error) {
						return nil, &types.ReplicationGroupNotFoundFault{}
					},
				},
			},
			r:          replicationGroup(),
			want:       replicationGroup(withConditions(xpv1.Deleting())),
			returnsErr: false,
		},
		{
			name: "AlreadyDeletingState",
			e:    &external{},
			r:    replicationGroup(withProviderStatus(v1beta1.StatusDeleting)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusDeleting),
				withConditions(xpv1.Deleting())),
			returnsErr: false,
		},
		{
			name: "Failed",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDeleteReplicationGroup: func(ctx context.Context, _ *elasticache.DeleteReplicationGroupInput, opts []func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error) {
						return nil, errorBoom
					},
				}},
			r: replicationGroup(),
			want: replicationGroup(
				withConditions(xpv1.Deleting()),
			),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Delete(ctx, tc.r)

			if tc.returnsErr != (err != nil) {
				t.Errorf("tc.csd.Delete(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateReplicationGroupNumCacheClusters(t *testing.T) {

	cases := []struct {
		name                string
		e                   *external
		rg                  string
		existingClusterSize int
		desiredclusterSize  int
		want                error
	}{
		{
			name:                "ErrDesiredTooSmall",
			e:                   &external{cache: &cache{}, client: &fake.MockClient{}},
			desiredclusterSize:  0,
			existingClusterSize: 1,
			want:                errors.New("at least 1 replica is required"),
		},
		{
			name:                "ErrDesiredTooLarge",
			e:                   &external{cache: &cache{}, client: &fake.MockClient{}},
			desiredclusterSize:  7,
			existingClusterSize: 1,
			want:                errors.New("maximum of 5 replicas are allowed"),
		},
		{
			name: "ErrIncreaseReplicaCount",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockIncreaseReplicaCount: func(ctx context.Context, _ *elasticache.IncreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
						return &elasticache.IncreaseReplicaCountOutput{}, errors.New("error increasing number of cache clusters")
					},
				}},
			desiredclusterSize:  4,
			existingClusterSize: 1,
			want:                errors.New("error increasing number of cache clusters"),
		},
		{
			name: "IncreaseReplicaCount",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockIncreaseReplicaCount: func(ctx context.Context, _ *elasticache.IncreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
						return &elasticache.IncreaseReplicaCountOutput{}, nil
					},
				}},
			desiredclusterSize:  4,
			existingClusterSize: 1,
			want:                nil,
		},
		{
			name: "ErrDecreaseReplicaCount",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDecreaseReplicaCount: func(ctx context.Context, _ *elasticache.DecreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error) {
						return &elasticache.DecreaseReplicaCountOutput{}, errors.New("error decreasing number of cache clusters")
					},
				}},
			desiredclusterSize:  3,
			existingClusterSize: 5,
			want:                errors.New("error decreasing number of cache clusters"),
		},
		{
			name: "DecreaseReplicaCount",
			e: &external{
				cache: &cache{},
				client: &fake.MockClient{
					MockDecreaseReplicaCount: func(ctx context.Context, _ *elasticache.DecreaseReplicaCountInput, opts []func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error) {
						return &elasticache.DecreaseReplicaCountOutput{}, nil
					},
				}},
			desiredclusterSize:  3,
			existingClusterSize: 5,
			want:                nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.updateReplicationGroupNumCacheClusters(ctx, tc.rg, tc.existingClusterSize, tc.desiredclusterSize)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
