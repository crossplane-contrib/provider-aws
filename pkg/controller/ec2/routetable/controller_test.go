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

package routetable

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha4"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	rtID     = "some rt"
	vpcID    = "some vpc"
	igID     = "some ig"
	subnetID = "some subnet"

	errBoom = errors.New("boom")
)

type args struct {
	rt   ec2.RouteTableClient
	kube client.Client
	cr   *v1alpha4.RouteTable
}

type rtModifier func(*v1alpha4.RouteTable)

func withExternalName(name string) rtModifier {
	return func(r *v1alpha4.RouteTable) { meta.SetExternalName(r, name) }
}

func withSpec(p v1alpha4.RouteTableParameters) rtModifier {
	return func(r *v1alpha4.RouteTable) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha4.RouteTableObservation) rtModifier {
	return func(r *v1alpha4.RouteTable) { r.Status.AtProvider = s }
}

func withConditions(c ...runtimev1alpha1.Condition) rtModifier {
	return func(r *v1alpha4.RouteTable) { r.Status.ConditionedStatus.Conditions = c }
}

func rt(m ...rtModifier) *v1alpha4.RouteTable {
	cr := &v1alpha4.RouteTable{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha4.RouteTable
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(input *awsec2.DescribeRouteTablesInput) awsec2.DescribeRouteTablesRequest {
						return awsec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeRouteTablesOutput{
								RouteTables: []awsec2.RouteTable{{
									VpcId: aws.String(vpcID),
								}},
							}},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID), withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MulitpleTables": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(input *awsec2.DescribeRouteTablesInput) awsec2.DescribeRouteTablesRequest {
						return awsec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeRouteTablesOutput{
								RouteTables: []awsec2.RouteTable{{}, {}},
							}},
						}
					},
				},
				cr: rt(withExternalName(rtID)),
			},
			want: want{
				cr:  rt(withExternalName(rtID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(input *awsec2.DescribeRouteTablesInput) awsec2.DescribeRouteTablesRequest {
						return awsec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rt(withExternalName(rtID)),
			},
			want: want{
				cr:  rt(withExternalName(rtID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rt}
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
		cr     *v1alpha4.RouteTable
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
					MockUpdate:       test.NewMockClient().MockUpdate,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				rt: &fake.MockRouteTableClient{
					MockCreate: func(input *awsec2.CreateRouteTableInput) awsec2.CreateRouteTableRequest {
						return awsec2.CreateRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateRouteTableOutput{
								RouteTable: &awsec2.RouteTable{RouteTableId: aws.String(rtID)},
							}},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
			},
		},
		"EmptyResult": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().MockUpdate,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				rt: &fake.MockRouteTableClient{
					MockCreate: func(input *awsec2.CreateRouteTableInput) awsec2.CreateRouteTableRequest {
						return awsec2.CreateRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateRouteTableOutput{}},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
				err: errors.New(errCreate),
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().MockUpdate,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				rt: &fake.MockRouteTableClient{
					MockCreate: func(input *awsec2.CreateRouteTableInput) awsec2.CreateRouteTableRequest {
						return awsec2.CreateRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rt}
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
		cr     *v1alpha4.RouteTable
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(input *awsec2.DescribeRouteTablesInput) awsec2.DescribeRouteTablesRequest {
						return awsec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeRouteTablesOutput{
								RouteTables: []awsec2.RouteTable{{}},
							}},
						}
					},
					MockAssociate: func(input *awsec2.AssociateRouteTableInput) awsec2.AssociateRouteTableRequest {
						return awsec2.AssociateRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AssociateRouteTableOutput{}},
						}
					},
					MockCreateRoute: func(input *awsec2.CreateRouteInput) awsec2.CreateRouteRequest {
						return awsec2.CreateRouteRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateRouteOutput{}},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					Routes: []v1alpha4.Route{{
						GatewayID: aws.String(igID),
					}},
					Associations: []v1alpha4.Association{{
						SubnetID: aws.String(subnetID),
					}},
				}),
					withStatus(v1alpha4.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					Routes: []v1alpha4.Route{{
						GatewayID: aws.String(igID),
					}},
					Associations: []v1alpha4.Association{{
						SubnetID: aws.String(subnetID),
					}},
				}),
					withStatus(v1alpha4.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"CreateRouteFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(input *awsec2.DescribeRouteTablesInput) awsec2.DescribeRouteTablesRequest {
						return awsec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeRouteTablesOutput{
								RouteTables: []awsec2.RouteTable{{}},
							}},
						}
					},
					MockCreateRoute: func(input *awsec2.CreateRouteInput) awsec2.CreateRouteRequest {
						return awsec2.CreateRouteRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					Routes: []v1alpha4.Route{{
						GatewayID: aws.String(igID),
					}},
				}),
					withStatus(v1alpha4.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1alpha4.RouteTableParameters{
					Routes: []v1alpha4.Route{{
						GatewayID: aws.String(igID),
					}},
				}),
					withStatus(v1alpha4.RouteTableObservation{
						RouteTableID: rtID,
					})),
				err: errors.Wrap(errBoom, errCreateRoute),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rt}
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
		cr  *v1alpha4.RouteTable
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDelete: func(input *awsec2.DeleteRouteTableInput) awsec2.DeleteRouteTableRequest {
						return awsec2.DeleteRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteRouteTableOutput{}},
						}
					},
				},
				cr: rt(withStatus(v1alpha4.RouteTableObservation{
					RouteTableID: rtID,
				})),
			},
			want: want{
				cr: rt(withStatus(v1alpha4.RouteTableObservation{
					RouteTableID: rtID,
				}), withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"DeleteFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDelete: func(input *awsec2.DeleteRouteTableInput) awsec2.DeleteRouteTableRequest {
						return awsec2.DeleteRouteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rt(withStatus(v1alpha4.RouteTableObservation{
					RouteTableID: rtID,
				})),
			},
			want: want{
				cr: rt(withStatus(v1alpha4.RouteTableObservation{
					RouteTableID: rtID,
				}), withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rt}
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
