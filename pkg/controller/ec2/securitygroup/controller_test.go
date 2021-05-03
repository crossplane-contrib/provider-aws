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
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	sgID              = "some sgID"
	port80      int64 = 80
	port100     int64 = 100
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
			FromPort: aws.Int64(port80),
			ToPort:   aws.Int64(80),
			IPRanges: []v1beta1.IPRange{
				{CIDRIP: cidr},
			},
			IPProtocol: tcpProtocol,
		},
	}
}

func sgPersmissions() []awsec2.IpPermission {
	return []awsec2.IpPermission{
		{
			FromPort:   aws.Int64(port100),
			ToPort:     aws.Int64(port100),
			IpProtocol: aws.String(tcpProtocol),
			IpRanges: []awsec2.IpRange{{
				CidrIp: aws.String(cidr),
			}},
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{}},
							}},
						}
					},
				},
				cr: sg(withStatus(v1beta1.SecurityGroupObservation{
					SecurityGroupID: sgID,
				}),
					withExternalName(sgID)),
			},
			want: want{
				cr: sg(withExternalName(sgID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleSGs": {
			args: args{
				sg: &fake.MockSecurityGroupClient{
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{}, {}},
							}},
						}
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
				err: awsclient.Wrap(errBoom, errDescribe),
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
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(input *awsec2.CreateSecurityGroupInput) awsec2.CreateSecurityGroupRequest {
						return awsec2.CreateSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateSecurityGroupOutput{
								GroupId: aws.String(sgID),
							}},
						}
					},
					MockRevokeEgress: func(input *awsec2.RevokeSecurityGroupEgressInput) awsec2.RevokeSecurityGroupEgressRequest {
						return awsec2.RevokeSecurityGroupEgressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.RevokeSecurityGroupEgressOutput{}},
						}
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
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(input *awsec2.CreateSecurityGroupInput) awsec2.CreateSecurityGroupRequest {
						return awsec2.CreateSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: sg(),
			},
			want: want{
				cr:  sg(withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
		"RevokeFail": {
			args: args{
				kube: &test.MockClient{
					MockGet:          test.NewMockGetFn(nil),
					MockUpdate:       test.NewMockUpdateFn(nil),
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
				},
				sg: &fake.MockSecurityGroupClient{
					MockCreate: func(input *awsec2.CreateSecurityGroupInput) awsec2.CreateSecurityGroupRequest {
						return awsec2.CreateSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateSecurityGroupOutput{
								GroupId: aws.String(sgID),
							}},
						}
					},
					MockRevokeEgress: func(input *awsec2.RevokeSecurityGroupEgressInput) awsec2.RevokeSecurityGroupEgressRequest {
						return awsec2.RevokeSecurityGroupEgressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: sg(),
			},
			want: want{
				err: awsclient.Wrap(errBoom, errRevokeEgress),
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{
									IpPermissions:       sgPersmissions(),
									IpPermissionsEgress: sgPersmissions(),
								}},
							}},
						}
					},
					MockAuthorizeIgress: func(input *awsec2.AuthorizeSecurityGroupIngressInput) awsec2.AuthorizeSecurityGroupIngressRequest {
						return awsec2.AuthorizeSecurityGroupIngressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AuthorizeSecurityGroupIngressOutput{}},
						}
					},
					MockAuthorizeEgress: func(input *awsec2.AuthorizeSecurityGroupEgressInput) awsec2.AuthorizeSecurityGroupEgressRequest {
						return awsec2.AuthorizeSecurityGroupEgressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AuthorizeSecurityGroupEgressOutput{}},
						}
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{
									IpPermissions:       sgPersmissions(),
									IpPermissionsEgress: sgPersmissions(),
								}},
							}},
						}
					},
					MockAuthorizeIgress: func(input *awsec2.AuthorizeSecurityGroupIngressInput) awsec2.AuthorizeSecurityGroupIngressRequest {
						return awsec2.AuthorizeSecurityGroupIngressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
				err: awsclient.Wrap(errBoom, errAuthorizeIngress),
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

func compareTags(a awsec2.Tag, b awsec2.Tag) bool {
	return aws.StringValue(a.Key) < aws.StringValue(b.Key)
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{
									Tags: []awsec2.Tag{
										{
											Key:   aws.String("k1"),
											Value: aws.String("v1"),
										}, {
											Key:   aws.String("k2"),
											Value: aws.String("v2"),
										},
									},
								}},
							}},
						}
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{
									Tags: []awsec2.Tag{
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
							}},
						}
					},

					MockCreateTags: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						if diff := cmp.Diff(input.Tags, []awsec2.Tag{
							{
								Key:   aws.String("k2"),
								Value: aws.String("v2"),
							}, {
								Key:   aws.String("k3"),
								Value: aws.String("v3"),
							},
						}, cmpopts.SortSlices(compareTags)); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockDeleteTags: func(input *awsec2.DeleteTagsInput) awsec2.DeleteTagsRequest {
						if diff := cmp.Diff(input.Tags, []awsec2.Tag{{Key: aws.String("k4")}}); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						return awsec2.DeleteTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteTagsOutput{}},
						}
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
					MockDescribe: func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
						return awsec2.DescribeSecurityGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSecurityGroupsOutput{
								SecurityGroups: []awsec2.SecurityGroup{{
									Tags: []awsec2.Tag{},
								}},
							}},
						}
					},
					MockCreateTags: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
				err: awsclient.Wrap(errBoom, errCreateTags),
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
					MockDelete: func(input *awsec2.DeleteSecurityGroupInput) awsec2.DeleteSecurityGroupRequest {
						return awsec2.DeleteSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteSecurityGroupOutput{}},
						}
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
					MockDelete: func(input *awsec2.DeleteSecurityGroupInput) awsec2.DeleteSecurityGroupRequest {
						return awsec2.DeleteSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteSecurityGroupOutput{}},
						}
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
					MockDelete: func(input *awsec2.DeleteSecurityGroupInput) awsec2.DeleteSecurityGroupRequest {
						return awsec2.DeleteSecurityGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, sg: tc.sg}
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
