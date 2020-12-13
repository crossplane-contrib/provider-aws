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
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	vpcID          = "some Id"
	cidr           = "192.168.0.0/32"
	tenancyDefault = "default"

	errBoom = errors.New("boom")
)

type args struct {
	vpc  ec2.VPCClient
	kube client.Client
	cr   *v1beta1.VPC
}

type vpcModifier func(*v1beta1.VPC)

func withTags(tagMaps ...map[string]string) vpcModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.VPC) { r.Spec.ForProvider.Tags = tagList }
}

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
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{
									InstanceTenancy: awsec2.TenancyDefault,
									State:           awsec2.VpcStateAvailable,
								}},
							}},
						}
					},
					MockDescribeVpcAttributeRequest: func(input *awsec2.DescribeVpcAttributeInput) awsec2.DescribeVpcAttributeRequest {
						return awsec2.DescribeVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcAttributeOutput{
								EnableDnsHostnames: &awsec2.AttributeBooleanValue{},
								EnableDnsSupport:   &awsec2.AttributeBooleanValue{},
							}},
						}
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
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{}, {}},
							}},
						}
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
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
				err: errors.Wrap(errBoom, errDescribe),
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
					MockCreate: func(input *awsec2.CreateVpcInput) awsec2.CreateVpcRequest {
						return awsec2.CreateVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateVpcOutput{
								Vpc: &awsec2.Vpc{
									VpcId:     aws.String(vpcID),
									CidrBlock: aws.String(cidr),
								},
							}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:     vpc(withExternalName(vpcID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"CreateFail": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockCreate: func(input *awsec2.CreateVpcInput) awsec2.CreateVpcRequest {
						return awsec2.CreateVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(),
				err: errors.Wrap(errBoom, errCreate),
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
					MockModifyTenancy: func(input *awsec2.ModifyVpcTenancyInput) awsec2.ModifyVpcTenancyRequest {
						return awsec2.ModifyVpcTenancyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcTenancyOutput{}},
						}
					},
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
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
					MockModifyTenancy: func(input *awsec2.ModifyVpcTenancyInput) awsec2.ModifyVpcTenancyRequest {
						return awsec2.ModifyVpcTenancyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
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
				err: errors.Wrap(errBoom, errUpdate),
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
					MockDelete: func(input *awsec2.DeleteVpcInput) awsec2.DeleteVpcRequest {
						return awsec2.DeleteVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteVpcOutput{}},
						}
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
					MockDelete: func(input *awsec2.DeleteVpcInput) awsec2.DeleteVpcRequest {
						return awsec2.DeleteVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
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

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1beta1.VPC
		kube client.Client
	}
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
				cr:   vpc(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: vpc(withTags(resource.GetExternalTags(vpc()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   vpc(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
