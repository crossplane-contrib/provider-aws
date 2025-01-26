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
package cluster

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsredshift "github.com/aws/aws-sdk-go-v2/service/redshift"
	awsredshifttypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/redshift/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/redshift"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/redshift/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	masterUsername    = "root"
	replaceMe         = "replace-me!"
	errBoom           = errors.New("boom")
	nodeType          = "dc1.large"
	singleNode        = "single-node"
	name              = "redshift-test"
	vpcSecurityGroups = []awsredshifttypes.VpcSecurityGroupMembership{{VpcSecurityGroupId: aws.String("id-sg")}}
)

type args struct {
	redshift redshift.Client
	kube     client.Client
	cr       *v1alpha1.Cluster
}

type redshiftModifier func(*v1alpha1.Cluster)

func withMasterUsername(s string) redshiftModifier {
	return func(r *v1alpha1.Cluster) { r.Spec.ForProvider.MasterUsername = s }
}

func withConditions(c ...xpv1.Condition) redshiftModifier {
	return func(r *v1alpha1.Cluster) { r.Status.ConditionedStatus.Conditions = c }
}

func withClusterStatus(s string) redshiftModifier {
	return func(r *v1alpha1.Cluster) { r.Status.AtProvider.ClusterStatus = s }
}

func withNewClusterIdentifier(s string) redshiftModifier {
	return func(r *v1alpha1.Cluster) { r.Spec.ForProvider.NewClusterIdentifier = aws.String(s) }
}

func withNewExternalName(s string) redshiftModifier {
	return func(r *v1alpha1.Cluster) { meta.SetExternalName(r, s) }
}

func cluster(m ...redshiftModifier) *v1alpha1.Cluster {
	cr := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ForProvider: v1alpha1.ClusterParameters{
				MasterUsername:      masterUsername,
				NodeType:            nodeType,
				ClusterType:         &singleNode,
				NumberOfNodes:       aws.Int32(1),
				VPCSecurityGroupIDs: []string{"id-sg"},
				// Late Initialize
				AllowVersionUpgrade:              aws.Bool(false),
				AutomatedSnapshotRetentionPeriod: aws.Int32(0),
				Encrypted:                        aws.Bool(false),
				EnhancedVPCRouting:               aws.Bool(false),
				ManualSnapshotRetentionPeriod:    aws.Int32(0),
				PubliclyAccessible:               aws.Bool(false),
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{
								{
									ClusterStatus:                    aws.String(string(v1alpha1.StateAvailable)),
									NumberOfNodes:                    ptr.To[int32](1),
									ClusterIdentifier:                &name,
									MasterUsername:                   &masterUsername,
									NodeType:                         &nodeType,
									VpcSecurityGroups:                vpcSecurityGroups,
									AllowVersionUpgrade:              aws.Bool(false),
									AutomatedSnapshotRetentionPeriod: aws.Int32(0),
									Encrypted:                        aws.Bool(false),
									EnhancedVpcRouting:               aws.Bool(false),
									ManualSnapshotRetentionPeriod:    aws.Int32(0),
									PubliclyAccessible:               aws.Bool(false),
								},
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Available()),
					withClusterStatus(string(v1alpha1.StateAvailable))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: redshift.GetConnectionDetails(v1alpha1.Cluster{}),
				},
			},
		},
		"DeletingState": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{
								{
									ClusterStatus:                    aws.String(string(v1alpha1.StateDeleting)),
									NumberOfNodes:                    ptr.To[int32](1),
									ClusterIdentifier:                &name,
									MasterUsername:                   &masterUsername,
									NodeType:                         &nodeType,
									VpcSecurityGroups:                vpcSecurityGroups,
									AllowVersionUpgrade:              aws.Bool(false),
									AutomatedSnapshotRetentionPeriod: aws.Int32(0),
									Encrypted:                        aws.Bool(false),
									EnhancedVpcRouting:               aws.Bool(false),
									ManualSnapshotRetentionPeriod:    aws.Int32(0),
									PubliclyAccessible:               aws.Bool(false),
								},
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Deleting()),
					withClusterStatus(string(v1alpha1.StateDeleting))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: redshift.GetConnectionDetails(v1alpha1.Cluster{}),
				},
			},
		},
		"FailedState": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{
								{
									ClusterStatus:                    aws.String(string(v1alpha1.StateFailed)),
									NumberOfNodes:                    ptr.To[int32](1),
									ClusterIdentifier:                &name,
									MasterUsername:                   &masterUsername,
									NodeType:                         &nodeType,
									VpcSecurityGroups:                vpcSecurityGroups,
									AllowVersionUpgrade:              aws.Bool(false),
									AutomatedSnapshotRetentionPeriod: aws.Int32(0),
									Encrypted:                        aws.Bool(false),
									EnhancedVpcRouting:               aws.Bool(false),
									ManualSnapshotRetentionPeriod:    aws.Int32(0),
									PubliclyAccessible:               aws.Bool(false),
								},
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Unavailable()),
					withClusterStatus(string(v1alpha1.StateFailed))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: redshift.GetConnectionDetails(v1alpha1.Cluster{}),
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return nil, &awsredshifttypes.ClusterNotFoundFault{}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{
								{
									ClusterStatus:                    aws.String(v1alpha1.StateCreating),
									NumberOfNodes:                    ptr.To[int32](1),
									ClusterIdentifier:                &name,
									MasterUsername:                   &masterUsername,
									NodeType:                         &nodeType,
									VpcSecurityGroups:                vpcSecurityGroups,
									AllowVersionUpgrade:              aws.Bool(false),
									AutomatedSnapshotRetentionPeriod: aws.Int32(0),
									Encrypted:                        aws.Bool(false),
									EnhancedVpcRouting:               aws.Bool(false),
									ManualSnapshotRetentionPeriod:    aws.Int32(0),
									PubliclyAccessible:               aws.Bool(false),
								},
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withClusterStatus(string(v1alpha1.StateCreating)),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: redshift.GetConnectionDetails(v1alpha1.Cluster{}),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.redshift}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockCreate: func(ctx context.Context, input *awsredshift.CreateClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.CreateClusterOutput, error) {
						return &awsredshift.CreateClusterOutput{}, nil
					},
				},
				cr: cluster(withMasterUsername(masterUsername)),
			},
			want: want{
				cr: cluster(
					withMasterUsername(masterUsername),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: cluster(withClusterStatus(v1alpha1.StateCreating)),
			},
			want: want{
				cr: cluster(
					withClusterStatus(v1alpha1.StateCreating),
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockCreate: func(ctx context.Context, input *awsredshift.CreateClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.CreateClusterOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.redshift}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if string(tc.want.result.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]) == replaceMe {
				tc.want.result.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey] =
					o.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				redshift: &fake.MockRedshiftClient{
					MockModify: func(ctx context.Context, input *awsredshift.ModifyClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.ModifyClusterOutput, error) {
						return &awsredshift.ModifyClusterOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{{}},
						}, nil
					},
				},
				cr: cluster(withNewClusterIdentifier("update")),
			},
			want: want{
				cr: cluster(withNewClusterIdentifier("update"), withNewExternalName("update")),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: cluster(withClusterStatus(v1alpha1.StateModifying)),
			},
			want: want{
				cr: cluster(withClusterStatus(v1alpha1.StateModifying)),
			},
		},
		"FailedDescribe": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"FailedModify": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockModify: func(ctx context.Context, input *awsredshift.ModifyClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.ModifyClusterOutput, error) {
						return nil, errBoom
					},
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{{}},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errModifyFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.redshift}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.Cluster
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				redshift: &fake.MockRedshiftClient{
					MockDelete: func(ctx context.Context, input *awsredshift.DeleteClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.DeleteClusterOutput, error) {
						return &awsredshift.DeleteClusterOutput{}, nil
					},
					MockModify: func(ctx context.Context, input *awsredshift.ModifyClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.ModifyClusterOutput, error) {
						return &awsredshift.ModifyClusterOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{{}},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: cluster(withClusterStatus(v1alpha1.StateDeleting)),
			},
			want: want{
				cr: cluster(withClusterStatus(v1alpha1.StateDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},

				redshift: &fake.MockRedshiftClient{
					MockDelete: func(ctx context.Context, input *awsredshift.DeleteClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.DeleteClusterOutput, error) {
						return nil, &awsredshifttypes.ClusterNotFoundFault{}
					},
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return nil, &awsredshifttypes.ClusterNotFoundFault{}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				redshift: &fake.MockRedshiftClient{
					MockDelete: func(ctx context.Context, input *awsredshift.DeleteClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.DeleteClusterOutput, error) {
						return nil, errBoom
					},
					MockModify: func(ctx context.Context, input *awsredshift.ModifyClusterInput, opts []func(*awsredshift.Options)) (*awsredshift.ModifyClusterOutput, error) {
						return &awsredshift.ModifyClusterOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsredshift.DescribeClustersInput, opts []func(*awsredshift.Options)) (*awsredshift.DescribeClustersOutput, error) {
						return &awsredshift.DescribeClustersOutput{
							Clusters: []awsredshifttypes.Cluster{{}},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.redshift}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
