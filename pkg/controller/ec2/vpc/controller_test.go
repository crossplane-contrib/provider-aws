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

package vpc

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
	vpcID          = "some Id"
	cidr           = "192.168.0.0/32"
	tenancyDefault = "default"
	enableDNS      = true

	errBoom = errors.New("boom")
)

type args struct {
	vpc  ec2.VPCClient
	kube client.Client
	cr   *v1beta1.VPC
}

type vpcModifier func(*v1beta1.VPC)

func withExternalName(name string) vpcModifier {
	return func(r *v1beta1.VPC) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) vpcModifier {
	return func(r *v1beta1.VPC) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.VPCParameters) vpcModifier {
	return func(r *v1beta1.VPC) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.VPCObservation) vpcModifier {
	return func(r *v1beta1.VPC) { r.Status.AtProvider = s }
}

func vpc(m ...vpcModifier) *v1beta1.VPC {
	cr := &v1beta1.VPC{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.VPC
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeVpcsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error) {
						return &awsec2.DescribeVpcsOutput{
							Vpcs: []awsec2types.Vpc{{
								InstanceTenancy: awsec2types.TenancyDefault,
								State:           awsec2types.VpcStateAvailable,
							}},
						}, nil
					},
					MockDescribeVpcAttribute: func(ctx context.Context, input *awsec2.DescribeVpcAttributeInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcAttributeOutput, error) {
						return &awsec2.DescribeVpcAttributeOutput{
							EnableDnsHostnames: &awsec2types.AttributeBooleanValue{},
							EnableDnsSupport:   &awsec2types.AttributeBooleanValue{},
						}, nil
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy:    aws.String(tenancyDefault),
					CIDRBlock:          cidr,
					EnableDNSHostNames: aws.Bool(false),
					EnableDNSSupport:   aws.Bool(false),
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy:    aws.String(tenancyDefault),
					CIDRBlock:          cidr,
					EnableDNSHostNames: aws.Bool(false),
					EnableDNSSupport:   aws.Bool(false),
				}), withStatus(v1beta1.VPCObservation{
					VPCState: "available",
				}), withExternalName(vpcID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleVpcs": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeVpcsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error) {
						return &awsec2.DescribeVpcsOutput{
							Vpcs: []awsec2types.Vpc{{}, {}},
						}, nil
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeVpcsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error) {
						return nil, errBoom
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr     *v1beta1.VPC
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateVpcInput, opts []func(*awsec2.Options)) (*awsec2.CreateVpcOutput, error) {
						return &awsec2.CreateVpcOutput{
							Vpc: &awsec2types.Vpc{
								VpcId:     aws.String(vpcID),
								CidrBlock: aws.String(cidr),
							},
						}, nil
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:     vpc(withExternalName(vpcID)),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulWithAttributes": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateVpcInput, opts []func(*awsec2.Options)) (*awsec2.CreateVpcOutput, error) {
						return &awsec2.CreateVpcOutput{
							Vpc: &awsec2types.Vpc{
								VpcId:     aws.String(vpcID),
								CidrBlock: aws.String(cidr),
							},
						}, nil
					},
					MockModifyAttribute: func(ctx context.Context, input *awsec2.ModifyVpcAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifyVpcAttributeOutput, error) {
						return &awsec2.ModifyVpcAttributeOutput{}, nil
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					EnableDNSSupport: &enableDNS,
				})),
			},
			want: want{
				cr: vpc(withExternalName(vpcID),
					withSpec(v1beta1.VPCParameters{
						EnableDNSSupport: &enableDNS,
					})),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFail": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateVpcInput, opts []func(*awsec2.Options)) (*awsec2.CreateVpcOutput, error) {
						return nil, errBoom
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr     *v1beta1.VPC
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeVpcsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error) {
						return &awsec2.DescribeVpcsOutput{
							Vpcs: []awsec2types.Vpc{{
								VpcId: aws.String(vpcID),
							}},
						}, nil
					},
					MockModifyTenancy: func(ctx context.Context, input *awsec2.ModifyVpcTenancyInput, opts []func(*awsec2.Options)) (*awsec2.ModifyVpcTenancyOutput, error) {
						return &awsec2.ModifyVpcTenancyOutput{}, nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
					MockModifyAttribute: func(ctx context.Context, input *awsec2.ModifyVpcAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifyVpcAttributeOutput, error) {
						return &awsec2.ModifyVpcAttributeOutput{}, nil
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeVpcsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error) {
						return &awsec2.DescribeVpcsOutput{
							Vpcs: []awsec2types.Vpc{{
								VpcId: aws.String(vpcID),
							}},
						}, nil
					},
					MockModifyTenancy: func(ctx context.Context, input *awsec2.ModifyVpcTenancyInput, opts []func(*awsec2.Options)) (*awsec2.ModifyVpcTenancyOutput, error) {
						return nil, errBoom
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
					MockModifyAttribute: func(ctx context.Context, input *awsec2.ModifyVpcAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifyVpcAttributeOutput, error) {
						return &awsec2.ModifyVpcAttributeOutput{}, nil
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr  *v1beta1.VPC
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteVpcInput, opts []func(*awsec2.Options)) (*awsec2.DeleteVpcOutput, error) {
						return &awsec2.DeleteVpcOutput{}, nil
					},
				},
				cr: vpc(),
			},
			want: want{
				cr: vpc(withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteVpcInput, opts []func(*awsec2.Options)) (*awsec2.DeleteVpcOutput, error) {
						return nil, errBoom
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
