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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/cache/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache/fake"
)

const (
	name = "coolGroup"
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

func withConditions(c ...xpv1.Condition) replicationGroupModifier {
	return func(r *v1beta1.ReplicationGroup) { r.Status.ConditionedStatus.Conditions = c }
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
var _ managed.ExternalConnecter = &connector{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockClient{
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
			e: &external{client: &fake.MockClient{
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
	cases := []testCase{
		{
			name: "SuccessfulObserveWhileGroupCreating",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
					return &elasticache.DescribeReplicationGroupsOutput{
						ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusCreating)}},
					}, nil
				},
			}},
			r: replicationGroup(withReplicationGroupID(name)),
			want: replicationGroup(
				withProviderStatus(v1beta1.StatusCreating),
				withReplicationGroupID(name),
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "SuccessfulObserveWhileGroupDeleting",
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
					return &elasticache.DescribeReplicationGroupsOutput{
						ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusDeleting)}},
					}, nil
				},
			}},
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
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
					return &elasticache.DescribeReplicationGroupsOutput{
						ReplicationGroups: []types.ReplicationGroup{{Status: aws.String(v1beta1.StatusModifying)}},
					}, nil
				},
			}},
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
			e: &external{client: &fake.MockClient{
				MockDescribeReplicationGroups: func(ctx context.Context, _ *elasticache.DescribeReplicationGroupsInput, opts []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
					return &elasticache.DescribeReplicationGroupsOutput{
						ReplicationGroups: []types.ReplicationGroup{{
							ClusterEnabled:        aws.Bool(true),
							Status:                aws.String(v1beta1.StatusAvailable),
							ConfigurationEndpoint: &types.Endpoint{Address: aws.String(host), Port: int32(port)},
						}},
					}, nil
				},
			}},
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
			name: "SuccessfulObserveLateInitialized",
			e: &external{
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
			e: &external{client: &fake.MockClient{
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
							ConfigurationEndpoint:  &types.Endpoint{Address: aws.String(host), Port: int32(port)},
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
			),
			want: replicationGroup(
				withReplicationGroupID(name),
				withProviderStatus(v1beta1.StatusAvailable),
				withConditions(xpv1.Available()),
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
			e: &external{client: &fake.MockClient{
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
			e: &external{client: &fake.MockClient{
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
			e: &external{client: &fake.MockClient{
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
				err: awsclient.Wrap(errorBoom, errUpdateReplicationGroupCR),
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
