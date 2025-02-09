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

package securitygroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	sgID              = "some sgID"
	port80      int32 = 80
	port100     int32 = 100
	cidr              = "192.168.0.0/32"
	tcpProtocol       = "tcp"

	errBoom = errors.New("boom")
)

type args struct {
	sg   ec2.SecurityGroupClient
	kube client.Client
	cr   *v1beta1.SecurityGroup
}

type sgModifier func(*v1beta1.SecurityGroup)

func specPermissions() []v1beta1.IPPermission {
	return []v1beta1.IPPermission{
		{
			FromPort: &port80,
			ToPort:   &port80,
			IPRanges: []v1beta1.IPRange{
				{CIDRIP: cidr},
			},
			IPProtocol: tcpProtocol,
		},
	}
}

func sgPermissions(port int32, cidrs ...string) []awsec2types.IpPermission {
	ranges := make([]awsec2types.IpRange, 0, len(cidrs))
	for _, cidr := range cidrs {
		ranges = append(ranges, awsec2types.IpRange{
			CidrIp: aws.String(cidr),
		})
	}
	return []awsec2types.IpPermission{
		{
			FromPort:   aws.Int32(port),
			ToPort:     aws.Int32(port),
			IpProtocol: aws.String(tcpProtocol),
			IpRanges:   ranges,
		},
	}
}

func withExternalName(name string) sgModifier {
	return func(r *v1beta1.SecurityGroup) { meta.SetExternalName(r, name) }
}

func withSpec(p v1beta1.SecurityGroupParameters) sgModifier {
	return func(r *v1beta1.SecurityGroup) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.SecurityGroupObservation) sgModifier {
	return func(r *v1beta1.SecurityGroup) { r.Status.AtProvider = s }
}

func withConditions(c ...xpv1.Condition) sgModifier {
	return func(r *v1beta1.SecurityGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func sg(m ...sgModifier) *v1beta1.SecurityGroup {
	cr := &v1beta1.SecurityGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.SecurityGroup
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
			},
			want: want{
				cr: sg(
					withExternalName(sgID),
					withStatus(v1beta1.SecurityGroupObservation{
						IngressRules: []v1beta1.SecurityGroupRuleObservation{},
						EgressRules:  []v1beta1.SecurityGroupRuleObservation{},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SameTags": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{
									{
										Key:   ptr.To("key1"),
										Value: ptr.To("value1"),
									},
									{
										Key:   ptr.To("key2"),
										Value: ptr.To("value2"),
									},
									{
										Key:   ptr.To("key3"),
										Value: ptr.To("value3"),
									},
								},
							}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key1",
								Value: "value1",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key3",
								Value: "value3",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					}),
				),
			},
			want: want{
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key1",
								Value: "value1",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key3",
								Value: "value3",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						IngressRules: []v1beta1.SecurityGroupRuleObservation{},
						EgressRules:  []v1beta1.SecurityGroupRuleObservation{},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SameTagsDifferentOrder": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{
									{
										Key:   ptr.To("key1"),
										Value: ptr.To("value1"),
									},
									{
										Key:   ptr.To("key3"),
										Value: ptr.To("value3"),
									},
									{
										Key:   ptr.To("key2"),
										Value: ptr.To("value2"),
									},
								},
							}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key3",
								Value: "value3",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key1",
								Value: "value1",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					}),
				),
			},
			want: want{
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key3",
								Value: "value3",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key1",
								Value: "value1",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						IngressRules: []v1beta1.SecurityGroupRuleObservation{},
						EgressRules:  []v1beta1.SecurityGroupRuleObservation{},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DifferentTagValues": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{
									{
										Key:   ptr.To("key1"),
										Value: ptr.To("othervalue"),
									},
									{
										Key:   ptr.To("key2"),
										Value: ptr.To("value2"),
									},
									{
										Key:   ptr.To("key3"),
										Value: ptr.To("value3"),
									},
								},
							}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key1",
								Value: "value1",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key3",
								Value: "value3",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					}),
				),
			},
			want: want{
				cr: sg(
					withExternalName(sgID),
					withSpec(v1beta1.SecurityGroupParameters{
						Tags: []v1beta1.Tag{
							{
								Key:   "key1",
								Value: "value1",
							},
							{
								Key:   "key2",
								Value: "value2",
							},
							{
								Key:   "key3",
								Value: "value3",
							},
						},
					}),
					withStatus(v1beta1.SecurityGroupObservation{
						IngressRules: []v1beta1.SecurityGroupRuleObservation{},
						EgressRules:  []v1beta1.SecurityGroupRuleObservation{},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"EmptyExternalNameExistingSG": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{GroupId: aws.String(sgID)}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(withExternalName("")),
			},
			want: want{
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
					IngressRules:    []v1beta1.SecurityGroupRuleObservation{},
					EgressRules:     []v1beta1.SecurityGroupRuleObservation{},
				}),
					withExternalName(sgID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"EmptyExternalNameNonExistingSG": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(),
			},
			want: want{
				cr: sg(),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"EmptyExternalNameClientError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return nil, errBoom
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(),
			},
			want: want{
				cr: sg(),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: errorutils.Wrap(errBoom, errGetSecurityGroup),
			},
		},
		"MultipleSGs": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{}, {}},
						}, nil
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
			},
			want: want{
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFailure": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return nil, errBoom
					},
					MockDescribeRules: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []awsec2types.SecurityGroupRule{},
						}, nil
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
			},
			want: want{
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
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
		cr     *v1beta1.SecurityGroup
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockGet:          test.NewMockGetFn(nil),
					MockUpdate:       test.NewMockUpdateFn(nil),
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.CreateSecurityGroupOutput, error) {
						return &awsec2.CreateSecurityGroupOutput{
							GroupId: aws.String(sgID),
						}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return &awsec2.RevokeSecurityGroupEgressOutput{}, nil
					},
				},
				cr: sg(),
			},
			want: want{
				cr: sg(withExternalName(sgID),
					withConditions(xpv1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockUpdateFn(nil),
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.CreateSecurityGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: sg(),
			},
			want: want{
				cr:  sg(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
		"RevokeFail": {
			args: args{
				kube: &test.MockClient{
					MockGet:          test.NewMockGetFn(nil),
					MockUpdate:       test.NewMockUpdateFn(nil),
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.CreateSecurityGroupOutput, error) {
						return &awsec2.CreateSecurityGroupOutput{
							GroupId: aws.String(sgID),
						}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return nil, errBoom
					},
				},
				cr: sg(),
			},
			want: want{
				err: errorutils.Wrap(errBoom, errRevokeEgress),
				cr: sg(withExternalName(sgID),
					withConditions(xpv1.Creating())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
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
		cr     *v1beta1.SecurityGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								IpPermissions:       sgPermissions(port100, cidr),
								IpPermissionsEgress: sgPermissions(port100, cidr),
							}},
						}, nil
					},
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						return &awsec2.AuthorizeSecurityGroupIngressOutput{}, nil
					},
					MockRevokeIngress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupIngressOutput, error) {
						return &awsec2.RevokeSecurityGroupIngressOutput{}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						return &awsec2.AuthorizeSecurityGroupEgressOutput{}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return &awsec2.RevokeSecurityGroupEgressOutput{}, nil
					},
				},
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Ingress: specPermissions(),
					Egress:  specPermissions(),
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
			want: want{
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Ingress: specPermissions(),
					Egress:  specPermissions(),
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
		},
		"IngressFail": {
			args: args{
				sg: &fake.MockSecurityGroupClient{

					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								IpPermissions:       sgPermissions(port100, cidr),
								IpPermissionsEgress: sgPermissions(port100, cidr),
							}},
						}, nil
					},
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						return nil, errBoom
					},
					MockRevokeIngress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupIngressOutput, error) {
						return nil, nil
					},
				},
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Ingress: specPermissions(),
					Egress:  specPermissions(),
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
			want: want{
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Ingress: specPermissions(),
					Egress:  specPermissions(),
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
				err: errorutils.Wrap(errBoom, errAuthorizeIngress),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
			o, err := e.Update(context.Background(), tc.args.cr)

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

func compareTags(a awsec2types.Tag, b awsec2types.Tag) bool {
	return pointer.StringValue(a.Key) < pointer.StringValue(b.Key)
}

func TestUpdateTags(t *testing.T) {
	type want struct {
		cr     *v1beta1.SecurityGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Same": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{
									{
										Key:   aws.String("k1"),
										Value: aws.String("v1"),
									}, {
										Key:   aws.String("k2"),
										Value: aws.String("v2"),
									},
								},

								IpPermissions:       sgPermissions(port100),
								IpPermissionsEgress: sgPermissions(port100),
							}},
						}, nil
					},
				},
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						}, {
							Key:   "k2",
							Value: "v2",
						},
					},
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
			want: want{
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						}, {
							Key:   "k2",
							Value: "v2",
						},
					}}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
		},
		"Change": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{
									{
										Key:   aws.String("k1"),
										Value: aws.String("v1"),
									},
									{
										Key:   aws.String("k2"),
										Value: aws.String("vx"),
									},
									{
										Key:   aws.String("k4"),
										Value: aws.String("v4"),
									},
								},
							}},
						}, nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						if diff := cmp.Diff(input.Tags, []awsec2types.Tag{
							{
								Key:   aws.String("k2"),
								Value: aws.String("v2"),
							}, {
								Key:   aws.String("k3"),
								Value: aws.String("v3"),
							},
						}, cmpopts.SortSlices(compareTags), cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						return &awsec2.CreateTagsOutput{}, nil
					},
					MockDeleteTags: func(ctx context.Context, input *awsec2.DeleteTagsInput, opts []func(*awsec2.Options)) (*awsec2.DeleteTagsOutput, error) {
						if diff := cmp.Diff(input.Tags, []awsec2types.Tag{{Key: aws.String("k4")}}, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						return &awsec2.DeleteTagsOutput{}, nil
					},
				},
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						}, {
							Key:   "k2",
							Value: "v2",
						},
						{
							Key:   "k3",
							Value: "v3",
						},
					},
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
			want: want{
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						}, {
							Key:   "k2",
							Value: "v2",
						},
						{
							Key:   "k3",
							Value: "v3",
						},
					}}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
		},
		"TagsFail": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupsInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error) {
						return &awsec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []awsec2types.SecurityGroup{{
								Tags: []awsec2types.Tag{},
							}},
						}, nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, errBoom
					},
				},
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						},
					},
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
			},
			want: want{
				cr: sg(withSpec(v1beta1.SecurityGroupParameters{
					Tags: []v1beta1.Tag{
						{
							Key:   "k1",
							Value: "v1",
						},
					},
				}),
					withStatus(v1beta1.SecurityGroupObservation{
						SecurityGroupID: sgID,
					})),
				err: errorutils.Wrap(errBoom, errCreateTags),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
			o, err := e.Update(context.Background(), tc.args.cr)

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
		cr  *v1beta1.SecurityGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.DeleteSecurityGroupOutput, error) {
						return &awsec2.DeleteSecurityGroupOutput{}, nil
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				})),
			},
			want: want{
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}), withConditions(xpv1.Deleting())),
			},
		},
		"InvalidSgId": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.DeleteSecurityGroupOutput, error) {
						return &awsec2.DeleteSecurityGroupOutput{}, nil
					},
				},
				cr: sg(),
			},
			want: want{
				cr: sg(withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailure": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteSecurityGroupInput, opts []func(*awsec2.Options)) (*awsec2.DeleteSecurityGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				})),
			},
			want: want{
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}), withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
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
