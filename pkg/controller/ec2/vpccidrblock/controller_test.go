/*
Copyright 2021 The Crossplane Authors.

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

package vpccidrblock

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	vpcID              = "some Id"
	cidr               = "192.168.0.0/32"
	matchAssociationID = "test"
	ipv6CIDR           = "2002::1234:abcd:ffff:c0a8:101/64"
	testStatus         = "status"
	testState          = string(awsec2.VpcCidrBlockStateCodeAssociated)

	errBoom = errors.New("boom")
)

type args struct {
	vpc  ec2.VPCCIDRBlockClient
	kube client.Client
	cr   *v1alpha1.VPCCIDRBlock
}

type vpcCIDRBlockModifier func(*v1alpha1.VPCCIDRBlock)

func withExternalName(name string) vpcCIDRBlockModifier {
	return func(r *v1alpha1.VPCCIDRBlock) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) vpcCIDRBlockModifier {
	return func(r *v1alpha1.VPCCIDRBlock) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.VPCCIDRBlockParameters) vpcCIDRBlockModifier {
	return func(r *v1alpha1.VPCCIDRBlock) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha1.VPCCIDRBlockObservation) vpcCIDRBlockModifier {
	return func(r *v1alpha1.VPCCIDRBlock) { r.Status.AtProvider = s }
}

func vpcCIDRBlock(m ...vpcCIDRBlockModifier) *v1alpha1.VPCCIDRBlock {
	cr := &v1alpha1.VPCCIDRBlock{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.VPCCIDRBlock
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailableIPv4": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCCIDRBlockClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{
									CidrBlockAssociationSet: []awsec2.VpcCidrBlockAssociation{
										{
											AssociationId: &matchAssociationID,
											CidrBlock:     &cidr,
											CidrBlockState: &awsec2.VpcCidrBlockState{
												State:         awsec2.VpcCidrBlockStateCodeAssociated,
												StatusMessage: &testStatus,
											},
										}},
								}},
							}},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				}), withExternalName(matchAssociationID)),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					VPCID:     &vpcID,
					CIDRBlock: &cidr,
				}), withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					CIDRBlock:     &cidr,
					CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				}), withExternalName(matchAssociationID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessfulAvailableIPv6": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCCIDRBlockClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{
									Ipv6CidrBlockAssociationSet: []awsec2.VpcIpv6CidrBlockAssociation{
										{
											AssociationId: &matchAssociationID,
											Ipv6CidrBlock: &ipv6CIDR,
											Ipv6CidrBlockState: &awsec2.VpcCidrBlockState{
												State:         awsec2.VpcCidrBlockStateCodeAssociated,
												StatusMessage: &testStatus,
											},
										}},
								}},
							}},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					IPv6CIDRBlock: &ipv6CIDR,
					VPCID:         &vpcID,
				}), withExternalName(matchAssociationID)),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					VPCID:         &vpcID,
					IPv6CIDRBlock: &ipv6CIDR,
				}), withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					IPv6CIDRBlock: &ipv6CIDR,
					IPv6CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				}), withExternalName(matchAssociationID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCCIDRBlockClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				}), withExternalName(matchAssociationID)),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					VPCID:     &vpcID,
					CIDRBlock: &cidr,
				}), withStatus(v1alpha1.VPCCIDRBlockObservation{}), withExternalName(matchAssociationID)),
				err: awsclient.Wrap(errBoom, errDescribe),
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
		cr     *v1alpha1.VPCCIDRBlock
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulIPv4": {
			args: args{
				vpc: &fake.MockVPCCIDRBlockClient{
					MockAssociate: func(input *awsec2.AssociateVpcCidrBlockInput) awsec2.AssociateVpcCidrBlockRequest {
						return awsec2.AssociateVpcCidrBlockRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AssociateVpcCidrBlockOutput{
								VpcId: aws.String(vpcID),
								CidrBlockAssociation: &awsec2.VpcCidrBlockAssociation{
									AssociationId:  aws.String(matchAssociationID),
									CidrBlock:      aws.String(cidr),
									CidrBlockState: &awsec2.VpcCidrBlockState{},
								},
							}},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				})),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				}), withExternalName(matchAssociationID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"SuccessfulIPv6": {
			args: args{
				vpc: &fake.MockVPCCIDRBlockClient{
					MockAssociate: func(input *awsec2.AssociateVpcCidrBlockInput) awsec2.AssociateVpcCidrBlockRequest {
						return awsec2.AssociateVpcCidrBlockRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AssociateVpcCidrBlockOutput{
								VpcId: aws.String(vpcID),
								Ipv6CidrBlockAssociation: &awsec2.VpcIpv6CidrBlockAssociation{
									AssociationId:      aws.String(matchAssociationID),
									Ipv6CidrBlock:      aws.String(ipv6CIDR),
									Ipv6CidrBlockState: &awsec2.VpcCidrBlockState{},
								},
							}},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					IPv6CIDRBlock: &ipv6CIDR,
					VPCID:         &vpcID,
				})),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					IPv6CIDRBlock: &ipv6CIDR,
					VPCID:         &vpcID,
				}), withExternalName(matchAssociationID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"CreateFail": {
			args: args{
				vpc: &fake.MockVPCCIDRBlockClient{
					MockAssociate: func(input *awsec2.AssociateVpcCidrBlockInput) awsec2.AssociateVpcCidrBlockRequest {
						return awsec2.AssociateVpcCidrBlockRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				})),
			},
			want: want{
				cr: vpcCIDRBlock(withSpec(v1alpha1.VPCCIDRBlockParameters{
					CIDRBlock: &cidr,
					VPCID:     &vpcID,
				})),
				err: awsclient.Wrap(errBoom, errAssociate),
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

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.VPCCIDRBlock
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCCIDRBlockClient{
					MockDisassociate: func(input *awsec2.DisassociateVpcCidrBlockInput) awsec2.DisassociateVpcCidrBlockRequest {
						return awsec2.DisassociateVpcCidrBlockRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DisassociateVpcCidrBlockOutput{}},
						}
					},
				},
				cr: vpcCIDRBlock(withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					CIDRBlock:     &cidr,
					CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				})),
			},
			want: want{
				cr: vpcCIDRBlock(withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					CIDRBlock:     &cidr,
					CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				})),
			},
		},
		"DeleteFailed": {
			args: args{
				vpc: &fake.MockVPCCIDRBlockClient{
					MockDisassociate: func(input *awsec2.DisassociateVpcCidrBlockInput) awsec2.DisassociateVpcCidrBlockRequest {
						return awsec2.DisassociateVpcCidrBlockRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpcCIDRBlock(withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					CIDRBlock:     &cidr,
					CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				})),
			},
			want: want{
				cr: vpcCIDRBlock(withStatus(v1alpha1.VPCCIDRBlockObservation{
					AssociationID: &matchAssociationID,
					CIDRBlock:     &cidr,
					CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
						State:         &testState,
						StatusMessage: &testStatus,
					},
				})),
				err: awsclient.Wrap(errBoom, errDisassociate),
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
