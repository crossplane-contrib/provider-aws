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

package subnet

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	subnetID = "some Id"

	errBoom = errors.New("boom")
)

type args struct {
	subnet ec2.SubnetClient
	kube   client.Client
	cr     *v1beta1.Subnet
}

type subnetModifier func(*v1beta1.Subnet)

func withExternalName(name string) subnetModifier {
	return func(r *v1beta1.Subnet) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.SubnetParameters) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.SubnetObservation) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.AtProvider = s }
}

func subnet(m ...subnetModifier) *v1beta1.Subnet {
	cr := &v1beta1.Subnet{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Subnet
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return &awsec2.DescribeSubnetsOutput{
							Subnets: []awsec2types.Subnet{
								{
									State: awsec2types.SubnetStateAvailable,
								},
							},
						}, nil
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetState: "available",
				}), withExternalName(subnetID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleSubnets": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return &awsec2.DescribeSubnetsOutput{
							Subnets: []awsec2types.Subnet{{}, {}},
						}, nil
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errors.New(errMultipleItems),
			},
		},
		"NotUpToDate": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return &awsec2.DescribeSubnetsOutput{
							Subnets: []awsec2types.Subnet{
								{
									State:                       awsec2types.SubnetStateAvailable,
									MapPublicIpOnLaunch:         aws.Bool(false),
									AssignIpv6AddressOnCreation: aws.Bool(false),
								},
							},
						}, nil
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}),
					withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch:         aws.Bool(true),
					AssignIPv6AddressOnCreation: aws.Bool(false),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetState: string(awsec2types.SubnetStateAvailable),
				}), withExternalName(subnetID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
		"DescribeFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return nil, errBoom
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr     *v1beta1.Subnet
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateSubnetInput, opts []func(*awsec2.Options)) (*awsec2.CreateSubnetOutput, error) {
						return &awsec2.CreateSubnetOutput{
							Subnet: &awsec2types.Subnet{
								SubnetId: aws.String(subnetID),
							},
						}, nil
					},
				},
				cr: subnet(),
			},
			want: want{
				cr:     subnet(withExternalName(subnetID)),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				subnet: &fake.MockSubnetClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateSubnetInput, opts []func(*awsec2.Options)) (*awsec2.CreateSubnetOutput, error) {
						return nil, errBoom
					},
				},
				cr: subnet(),
			},
			want: want{
				cr:  subnet(),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
			o, err := e.Create(context.Background(), tc.args.cr)

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

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1beta1.Subnet
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(ctx context.Context, input *awsec2.ModifySubnetAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifySubnetAttributeOutput, error) {
						return &awsec2.ModifySubnetAttributeOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return &awsec2.DescribeSubnetsOutput{
							Subnets: []awsec2types.Subnet{{
								SubnetId:            aws.String(subnetID),
								MapPublicIpOnLaunch: aws.Bool(false),
							}},
						}, nil
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(ctx context.Context, input *awsec2.ModifySubnetAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifySubnetAttributeOutput, error) {
						return &awsec2.ModifySubnetAttributeOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSubnetsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error) {
						return &awsec2.DescribeSubnetsOutput{
							Subnets: []awsec2types.Subnet{{
								SubnetId:            aws.String(subnetID),
								MapPublicIpOnLaunch: aws.Bool(false),
							}},
						}, nil
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr  *v1beta1.Subnet
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteSubnetInput, opts []func(*awsec2.Options)) (*awsec2.DeleteSubnetOutput, error) {
						return &awsec2.DeleteSubnetOutput{}, nil
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteSubnetInput, opts []func(*awsec2.Options)) (*awsec2.DeleteSubnetOutput, error) {
						return nil, errBoom
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
