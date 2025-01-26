package securitygrouprule

import (
	"context"
	"errors"
	"testing"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
)

var (
	sgID                        = "some sg ID"
	sgrID                       = "some sgr  ID"
	sgrIDIngress                = "some sgr ingress ID"
	sgrIDEgress                 = "some sgr egress ID"
	cidrIpv4Block               = "172.1.0.0/16"
	cidrIpv6Block               = "2001:0DB8:7654:0010:FEDC:0000:0000:3210/128"
	prefixListID                = "some prefix ID"
	fromPort              int32 = 10
	wrongFromPort         int32 = 0
	toPort                int32 = 20
	ingressTypeTest             = "ingress"
	egressTypeTest              = "egress"
	trueValue                   = true
	falseValue                  = false
	description                 = "description"
	sourceSecurityGroupID       = "some source sg ID"
)

type args struct {
	sgr  ec2.SecurityGroupRuleClient
	kube client.Client
	cr   *manualv1alpha1.SecurityGroupRule
}

type sgrModifier func(*manualv1alpha1.SecurityGroupRule)

func withSpec(p manualv1alpha1.SecurityGroupRuleParameters) sgrModifier {
	return func(r *manualv1alpha1.SecurityGroupRule) { r.Spec.ForProvider = p }
}

func withExternalName(name string) sgrModifier {
	return func(r *manualv1alpha1.SecurityGroupRule) { meta.SetExternalName(r, name) }
}

func securityGroupRule(m ...sgrModifier) *manualv1alpha1.SecurityGroupRule {
	cr := &manualv1alpha1.SecurityGroupRule{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withStatus(s manualv1alpha1.SecurityGroupRuleObservation) sgrModifier {
	return func(r *manualv1alpha1.SecurityGroupRule) { r.Status.AtProvider = s }
}

func withConditions(c ...xpv1.Condition) sgrModifier {
	return func(r *manualv1alpha1.SecurityGroupRule) { r.Status.ConditionedStatus.Conditions = c }
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.SecurityGroupRule
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{

		"HasNoExternalName": {
			// The sgr is newly created and has no external name
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &fromPort,
									ToPort:              &toPort,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				})),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"HasExternalNameBusDoesNotExist": {
			// No sgr exist at aws
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return nil, errors.New("Does not exist")
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrID)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrID)),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NeedsUpdate": {
			// The actual and the wanted sgr differ
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &wrongFromPort,
									ToPort:              &toPort,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrID)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withStatus(manualv1alpha1.SecurityGroupRuleObservation{
					SecurityGroupRuleID: &sgrID,
				}), withExternalName(sgrID), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"NeedsUpdateDifferentType": {
			// The existing and the wanted sgr differ in type (egress vs ingress)
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &fromPort,
									ToPort:              &toPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrID)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withStatus(manualv1alpha1.SecurityGroupRuleObservation{
					SecurityGroupRuleID: &sgrID,
				}), withExternalName(sgrID), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"DoesExistInDesiredState": {
			// The wanted and the existing sgr are in the same state, no update needed
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &fromPort,
									ToPort:              &toPort,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrID)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withStatus(manualv1alpha1.SecurityGroupRuleObservation{
					SecurityGroupRuleID: &sgrID,
				}), withExternalName(sgrID), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.sgr}
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
		cr     *manualv1alpha1.SecurityGroupRule
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{

		"CreateIngressipv4": {
			// Create a ingress sgr with an ipv4 cird-block
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrIDIngress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateIngressipv6": {
			// Create a ingress sgr with an ipv6 cird-block
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					Ipv6CidrBlock:   &cidrIpv6Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
					Description:     &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					Ipv6CidrBlock:   &cidrIpv6Block,
					Description:     &description,
				}), withExternalName(sgrIDIngress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateIngressPrefixListId": {
			// Create an ingress sgr with a prefix list id
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var prefixListID *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].PrefixListIds) > 0 {
							prefixListID = input.IpPermissions[0].PrefixListIds[0].PrefixListId
							description = input.IpPermissions[0].PrefixListIds[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									PrefixListId:        prefixListID,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var prefixListID *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].PrefixListIds) > 0 {
							prefixListID = input.IpPermissions[0].PrefixListIds[0].PrefixListId
							description = input.IpPermissions[0].PrefixListIds[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									PrefixListId:        prefixListID,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					PrefixListID:    &prefixListID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
					Description:     &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					PrefixListID:    &prefixListID,
					Description:     &description,
				}), withExternalName(sgrIDIngress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateIngressRefSG": {
			// Create a ingress sgr with a reference security group
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SourceSecurityGroupID: &sourceSecurityGroupID,
					FromPort:              &fromPort,
					ToPort:                &toPort,
					Type:                  &ingressTypeTest,
					SecurityGroupID:       &sgID,
					Description:           &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID:       &sgID,
					FromPort:              &fromPort,
					ToPort:                &toPort,
					Type:                  &ingressTypeTest,
					SourceSecurityGroupID: &sourceSecurityGroupID,
					Description:           &description,
				}), withExternalName(sgrIDIngress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateEgressipv4": {
			// Create a egress sgr with an ipv4 cird-block
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					SecurityGroupID: &sgID,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
				}), withExternalName(sgrIDEgress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateEgressipv6": {
			// Create a egress sgr with an ipv6 cird-block
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					Ipv6CidrBlock:   &cidrIpv6Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					SecurityGroupID: &sgID,
					Description:     &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					Ipv6CidrBlock:   &cidrIpv6Block,
					Description:     &description,
				}), withExternalName(sgrIDEgress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateEgressPrefixListId": {
			// Create an egress sgr with a prefix list id
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var prefixListID *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].PrefixListIds) > 0 {
							prefixListID = input.IpPermissions[0].PrefixListIds[0].PrefixListId
							description = input.IpPermissions[0].PrefixListIds[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									PrefixListId:        prefixListID,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var prefixListID *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].PrefixListIds) > 0 {
							prefixListID = input.IpPermissions[0].PrefixListIds[0].PrefixListId
							description = input.IpPermissions[0].PrefixListIds[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									PrefixListId:        prefixListID,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					PrefixListID:    &prefixListID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					SecurityGroupID: &sgID,
					Description:     &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					PrefixListID:    &prefixListID,
					Description:     &description,
				}), withExternalName(sgrIDEgress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateEgressRefSG": {
			// Create a egress sgr with a reference security group
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockAuthorizeIngress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupIngressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDIngress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &falseValue,
								},
							},
						}, nil
					},
					MockAuthorizeEgress: func(ctx context.Context, input *awsec2.AuthorizeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.AuthorizeSecurityGroupEgressOutput, error) {
						var cidrIpv6 *string = nil
						var cidrIpv4 *string = nil
						var description *string = nil
						var refSg *types.ReferencedSecurityGroup = nil
						if len(input.IpPermissions[0].Ipv6Ranges) > 0 {
							cidrIpv6 = input.IpPermissions[0].Ipv6Ranges[0].CidrIpv6
							description = input.IpPermissions[0].Ipv6Ranges[0].Description
						}
						if len(input.IpPermissions[0].IpRanges) > 0 {
							cidrIpv4 = input.IpPermissions[0].IpRanges[0].CidrIp
							description = input.IpPermissions[0].IpRanges[0].Description
						}
						if len(input.IpPermissions[0].UserIdGroupPairs) > 0 {
							refSg = &types.ReferencedSecurityGroup{
								GroupId: input.IpPermissions[0].UserIdGroupPairs[0].GroupId,
							}
							description = input.IpPermissions[0].UserIdGroupPairs[0].Description
						}
						return &awsec2.AuthorizeSecurityGroupEgressOutput{
							Return: &trueValue,
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            cidrIpv4,
									CidrIpv6:            cidrIpv6,
									Description:         description,
									GroupId:             input.GroupId,
									SecurityGroupRuleId: &sgrIDEgress,
									ReferencedGroupInfo: refSg,
									FromPort:            input.IpPermissions[0].FromPort,
									ToPort:              input.IpPermissions[0].ToPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SourceSecurityGroupID: &sourceSecurityGroupID,
					FromPort:              &fromPort,
					ToPort:                &toPort,
					Type:                  &egressTypeTest,
					SecurityGroupID:       &sgID,
					Description:           &description,
				})),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					SecurityGroupID:       &sgID,
					FromPort:              &fromPort,
					ToPort:                &toPort,
					Type:                  &egressTypeTest,
					SourceSecurityGroupID: &sourceSecurityGroupID,
					Description:           &description,
				}), withExternalName(sgrIDEgress)),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.sgr}
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
		cr  *manualv1alpha1.SecurityGroupRule
		err error
	}

	cases := map[string]struct {
		args
		want
	}{

		"DeleteIngress": {
			// Delete a sgr of type ingress
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockRevokeIngress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupIngressOutput, error) {
						return &awsec2.RevokeSecurityGroupIngressOutput{
							Return: &trueValue,
						}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return &awsec2.RevokeSecurityGroupEgressOutput{
							Return: &trueValue,
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
				}), withExternalName(sgrIDIngress)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrIDIngress), withConditions(xpv1.Deleting())),
				err: nil,
			},
		},
		"DeleteEgress": {
			// Delete a sgr of type egress
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockRevokeIngress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupIngressOutput, error) {
						return &awsec2.RevokeSecurityGroupIngressOutput{
							Return: &trueValue,
						}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return &awsec2.RevokeSecurityGroupEgressOutput{
							Return: &trueValue,
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
					SecurityGroupID: &sgID,
				}), withExternalName(sgrIDEgress)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &egressTypeTest,
				}), withExternalName(sgrIDEgress), withConditions(xpv1.Deleting())),
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.sgr}
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

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.SecurityGroupRule
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{

		"UpdateIngressInplace": {
			// Update a property (fromPort) that can be done in place
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockModify: func(ctx context.Context, input *awsec2.ModifySecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.ModifySecurityGroupRulesOutput, error) {
						return &awsec2.ModifySecurityGroupRulesOutput{
							Return: &trueValue,
						}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &wrongFromPort,
									ToPort:              &toPort,
								},
							},
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
				}), withExternalName(sgrIDIngress)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrIDIngress)),
				err: nil,
			},
		},
		"UpdateIngressReplace": {
			// Update a property (type) that foreces us to recreate the sgr
			args: args{
				sgr: &fake.MockSecurityGroupRuleClient{
					MockModify: func(ctx context.Context, input *awsec2.ModifySecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.ModifySecurityGroupRulesOutput, error) {
						return &awsec2.ModifySecurityGroupRulesOutput{
							Return: &trueValue,
						}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeSecurityGroupRulesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error) {
						return &awsec2.DescribeSecurityGroupRulesOutput{
							SecurityGroupRules: []types.SecurityGroupRule{
								{
									CidrIpv4:            &cidrIpv4Block,
									GroupId:             &sgID,
									SecurityGroupRuleId: &sgrID,
									FromPort:            &fromPort,
									ToPort:              &toPort,
									IsEgress:            &trueValue,
								},
							},
						}, nil
					},
					MockRevokeIngress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupIngressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupIngressOutput, error) {
						return &awsec2.RevokeSecurityGroupIngressOutput{
							Return: &trueValue,
						}, nil
					},
					MockRevokeEgress: func(ctx context.Context, input *awsec2.RevokeSecurityGroupEgressInput, opts []func(*awsec2.Options)) (*awsec2.RevokeSecurityGroupEgressOutput, error) {
						return &awsec2.RevokeSecurityGroupEgressOutput{
							Return: &trueValue,
						}, nil
					},
				},
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
					SecurityGroupID: &sgID,
				}), withExternalName(sgrIDIngress)),
			},
			want: want{
				cr: securityGroupRule(withSpec(manualv1alpha1.SecurityGroupRuleParameters{
					CidrBlock:       &cidrIpv4Block,
					SecurityGroupID: &sgID,
					FromPort:        &fromPort,
					ToPort:          &toPort,
					Type:            &ingressTypeTest,
				}), withExternalName(sgrIDIngress)),
				err: errors.New("Update needs recreation"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.sgr}
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
