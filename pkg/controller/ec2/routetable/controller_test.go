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
	rtID            = "some rt"
	vpcID           = "some vpc"
	igID            = "some ig"
	destinationCIDR = "0.0.0.0/0"
	instanceID      = "natID"
	associationID   = "someAssociation"
	subnetID        = "some subnet"
	removeSubnetID  = "removeMe"
	testKey         = "testKey"
	testValue       = "testValue"
	CIDR            = "10.0.0.0/8"
	instanceCIDR    = "10.1.1.1/32"
	errBoom         = errors.New("boom")
)

type args struct {
	rt   ec2.RouteTableClient
	kube client.Client
	cr   *v1beta1.RouteTable
}

type rtModifier func(*v1beta1.RouteTable)

func withExternalName(name string) rtModifier {
	return func(r *v1beta1.RouteTable) { meta.SetExternalName(r, name) }
}

func withSpec(p v1beta1.RouteTableParameters) rtModifier {
	return func(r *v1beta1.RouteTable) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.RouteTableObservation) rtModifier {
	return func(r *v1beta1.RouteTable) { r.Status.AtProvider = s }
}

func withConditions(c ...xpv1.Condition) rtModifier {
	return func(r *v1beta1.RouteTable) { r.Status.ConditionedStatus.Conditions = c }
}

func rt(m ...rtModifier) *v1beta1.RouteTable {
	cr := &v1beta1.RouteTable{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.RouteTable
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								VpcId: aws.String(vpcID),
							}},
						}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleTables": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}, {}},
						}, nil
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withExternalName(rtID)),
			},
			want: want{
				cr:  rt(withExternalName(rtID)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.rt}
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
		cr     *v1beta1.RouteTable
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
					MockCreate: func(ctx context.Context, input *awsec2.CreateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.CreateRouteTableOutput, error) {
						return &awsec2.CreateRouteTableOutput{
							RouteTable: &awsec2types.RouteTable{RouteTableId: aws.String(rtID)},
						}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				}), withExternalName(rtID)),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().MockUpdate,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				rt: &fake.MockRouteTableClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.CreateRouteTableOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					VPCID: aws.String(vpcID),
				})),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.rt}
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
		cr     *v1beta1.RouteTable
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockAssociate: func(ctx context.Context, input *awsec2.AssociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.AssociateRouteTableOutput, error) {
						return &awsec2.AssociateRouteTableOutput{}, nil
					},
					MockDeleteRoute: func(ctx context.Context, input *awsec2.DeleteRouteInput, opts []func(*awsec2.Options)) (*awsec2.DeleteRouteOutput, error) {
						return &awsec2.DeleteRouteOutput{}, nil
					},
					MockCreateRoute: func(ctx context.Context, input *awsec2.CreateRouteInput, opts []func(*awsec2.Options)) (*awsec2.CreateRouteOutput, error) {
						return &awsec2.CreateRouteOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						GatewayID:            aws.String(igID),
						DestinationCIDRBlock: aws.String(destinationCIDR),
					}},
					Associations: []v1beta1.Association{{
						SubnetID: aws.String(subnetID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						GatewayID:            aws.String(igID),
						DestinationCIDRBlock: aws.String(destinationCIDR),
					}},
					Associations: []v1beta1.Association{{
						SubnetID: aws.String(subnetID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"SuccessfulAddTags": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Tags: []v1beta1.Tag{{
						Key:   testKey,
						Value: testValue,
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Tags: []v1beta1.Tag{{
						Key:   testKey,
						Value: testValue,
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"SuccessfulRemoveTags": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Tags: []awsec2types.Tag{{Key: aws.String(testKey), Value: aws.String(testValue)}},
							}},
						}, nil
					},
					MockDeleteTags: func(ctx context.Context, input *awsec2.DeleteTagsInput, opts []func(*awsec2.Options)) (*awsec2.DeleteTagsOutput, error) {
						return &awsec2.DeleteTagsOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"SuccessfulAddAssociation": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockAssociate: func(ctx context.Context, input *awsec2.AssociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.AssociateRouteTableOutput, error) {
						return &awsec2.AssociateRouteTableOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{{SubnetID: aws.String(subnetID)}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{{SubnetID: aws.String(subnetID)}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"FailedAddAssociation": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockAssociate: func(ctx context.Context, input *awsec2.AssociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.AssociateRouteTableOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{{SubnetID: aws.String(subnetID)}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{{SubnetID: aws.String(subnetID)}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
				err: errors.Wrap(errBoom, errAssociateSubnet),
			},
		},
		"SuccessfulDisassociation": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Associations: []awsec2types.RouteTableAssociation{
									{
										RouteTableAssociationId: aws.String(associationID),
										SubnetId:                aws.String(subnetID),
									},
									{
										RouteTableAssociationId: aws.String(associationID),
										SubnetId:                aws.String(removeSubnetID),
									},
								},
							}},
						}, nil
					},
					MockDisassociate: func(ctx context.Context, input *awsec2.DisassociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.DisassociateRouteTableOutput, error) {
						return &awsec2.DisassociateRouteTableOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{
						{
							SubnetID: aws.String(subnetID),
						},
					},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Associations: []v1beta1.AssociationState{
							{
								AssociationID: associationID,
								SubnetID:      subnetID,
							},
							{
								AssociationID: associationID,
								SubnetID:      removeSubnetID,
							},
						},
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{
						{
							SubnetID: aws.String(subnetID),
						},
					},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Associations: []v1beta1.AssociationState{
							{
								AssociationID: associationID,
								SubnetID:      subnetID,
							},
							{
								AssociationID: associationID,
								SubnetID:      removeSubnetID,
							},
						},
					})),
			},
		},
		"FailedDisassociation": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Associations: []awsec2types.RouteTableAssociation{
									{
										RouteTableAssociationId: aws.String(associationID),
										SubnetId:                aws.String(subnetID),
									},
									{
										RouteTableAssociationId: aws.String(associationID),
										SubnetId:                aws.String(removeSubnetID),
									},
								},
							}},
						}, nil
					},
					MockDisassociate: func(ctx context.Context, input *awsec2.DisassociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.DisassociateRouteTableOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{
						{
							SubnetID: aws.String(subnetID),
						},
					},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Associations: []v1beta1.AssociationState{
							{
								AssociationID: associationID,
								SubnetID:      subnetID,
							},
							{
								AssociationID: associationID,
								SubnetID:      removeSubnetID,
							},
						},
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Associations: []v1beta1.Association{
						{
							SubnetID: aws.String(subnetID),
						},
					},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Associations: []v1beta1.AssociationState{
							{
								AssociationID: associationID,
								SubnetID:      subnetID,
							},
							{
								AssociationID: associationID,
								SubnetID:      removeSubnetID,
							},
						},
					})),
				err: errors.Wrap(errBoom, errDisassociateSubnet),
			},
		},
		"SuccessfulAddRoute": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Routes: []awsec2types.Route{
									{
										DestinationCidrBlock: aws.String(CIDR),
										GatewayId:            aws.String(igID)},
								},
							}},
						}, nil
					},
					MockCreateRoute: func(ctx context.Context, input *awsec2.CreateRouteInput, opts []func(*awsec2.Options)) (*awsec2.CreateRouteOutput, error) {
						return &awsec2.CreateRouteOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					},
						{
							DestinationCIDRBlock: aws.String(instanceCIDR),
							InstanceID:           aws.String(instanceCIDR),
						}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					},
						{
							DestinationCIDRBlock: aws.String(instanceCIDR),
							InstanceID:           aws.String(instanceCIDR),
						}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
		},
		"SuccessfulDeleteRoute": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Routes: []awsec2types.Route{
									{
										DestinationCidrBlock: aws.String(CIDR),
										GatewayId:            aws.String(igID),
									},
									{
										DestinationCidrBlock: aws.String(instanceCIDR),
										GatewayId:            aws.String(instanceID),
									},
								},
							}},
						}, nil
					},
					MockDeleteRoute: func(ctx context.Context, input *awsec2.DeleteRouteInput, opts []func(*awsec2.Options)) (*awsec2.DeleteRouteOutput, error) {
						return &awsec2.DeleteRouteOutput{}, nil
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: testValue,
						Routes: []v1beta1.RouteState{
							{
								DestinationCIDRBlock: CIDR,
								GatewayID:            igID,
							},
							{
								DestinationCIDRBlock: instanceCIDR,
								InstanceID:           instanceID,
							},
						},
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: testValue,
						Routes: []v1beta1.RouteState{
							{
								DestinationCIDRBlock: CIDR,
								GatewayID:            igID,
							},
							{
								DestinationCIDRBlock: instanceCIDR,
								InstanceID:           instanceID,
							},
						},
					})),
			},
		},
		"CreateRouteFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockCreateRoute: func(ctx context.Context, input *awsec2.CreateRouteInput, opts []func(*awsec2.Options)) (*awsec2.CreateRouteOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						GatewayID:            aws.String(igID),
						DestinationCIDRBlock: aws.String(destinationCIDR),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						GatewayID:            aws.String(igID),
						DestinationCIDRBlock: aws.String(destinationCIDR),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
					})),
				err: errorutils.Wrap(errBoom, errCreateRoute),
			},
		},
		"DeleteRouteFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{
								Routes: []awsec2types.Route{
									{
										DestinationCidrBlock: aws.String(CIDR),
										GatewayId:            aws.String(igID),
									},
									{
										DestinationCidrBlock: aws.String(instanceCIDR),
										GatewayId:            aws.String(instanceID),
									},
								},
							}},
						}, nil
					},
					MockDeleteRoute: func(ctx context.Context, input *awsec2.DeleteRouteInput, opts []func(*awsec2.Options)) (*awsec2.DeleteRouteOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Routes: []v1beta1.RouteState{
							{
								DestinationCIDRBlock: CIDR,
								GatewayID:            igID,
							},
							{
								DestinationCIDRBlock: instanceCIDR,
								InstanceID:           instanceID,
							},
						},
					})),
			},
			want: want{
				cr: rt(withSpec(v1beta1.RouteTableParameters{
					Routes: []v1beta1.RouteBeta{{
						DestinationCIDRBlock: aws.String(CIDR),
						GatewayID:            aws.String(igID),
					}},
				}),
					withStatus(v1beta1.RouteTableObservation{
						RouteTableID: rtID,
						Routes: []v1beta1.RouteState{
							{
								DestinationCIDRBlock: CIDR,
								GatewayID:            igID,
							},
							{
								DestinationCIDRBlock: instanceCIDR,
								InstanceID:           instanceID,
							},
						},
					})),
				err: errors.Wrap(errBoom, errDeleteRoute),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.rt}
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
		cr  *v1beta1.RouteTable
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockDelete: func(ctx context.Context, input *awsec2.DeleteRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.DeleteRouteTableOutput, error) {
						return &awsec2.DeleteRouteTableOutput{}, nil
					},
				},
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
				}), withExternalName(rtID), withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailAssociation": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockDisassociate: func(ctx context.Context, input *awsec2.DisassociateRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.DisassociateRouteTableOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
					Associations: []v1beta1.AssociationState{{AssociationID: associationID}},
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
					Associations: []v1beta1.AssociationState{{AssociationID: associationID}},
				}), withExternalName(rtID), withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDisassociateSubnet),
			},
		},
		"DeleteFail": {
			args: args{
				rt: &fake.MockRouteTableClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeRouteTablesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeRouteTablesOutput, error) {
						return &awsec2.DescribeRouteTablesOutput{
							RouteTables: []awsec2types.RouteTable{{}},
						}, nil
					},
					MockDelete: func(ctx context.Context, input *awsec2.DeleteRouteTableInput, opts []func(*awsec2.Options)) (*awsec2.DeleteRouteTableOutput, error) {
						return nil, errBoom
					},
				},
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
				}), withExternalName(rtID)),
			},
			want: want{
				cr: rt(withStatus(v1beta1.RouteTableObservation{
					RouteTableID: rtID,
				}), withExternalName(rtID), withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.rt}
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
