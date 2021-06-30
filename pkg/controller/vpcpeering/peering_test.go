package vpcpeering

import (
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"net/http"

	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/pkg/clients/peering"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	svcapitypes "github.com/crossplane/provider-aws/apis/vpcpeering/v1alpha1"

	"github.com/crossplane/provider-aws/pkg/clients/peering/fake"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type args struct {
	kube       client.Client
	client     peering.EC2Client
	route53Cli peering.Route53Client
	cr         *svcapitypes.VPCPeeringConnection
}

func TestObserve(t *testing.T) {
	type want struct {
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Create": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
					MockStatusUpdate: func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return nil
					},
				},
				cr: buildVPCPeerConnection("test"),
				client: &fake.MockEC2Client{
					DescribeVpcPeeringConnectionsRequestFun: func(input *ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest {
						return ec2.DescribeVpcPeeringConnectionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.DescribeVpcPeeringConnectionsOutput{
								//Attributes: attributes,
								VpcPeeringConnections: []ec2.VpcPeeringConnection{},
							}},
						}
					},
				},
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"UpToDate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
					MockStatusUpdate: func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return nil
					},
				},
				cr: func() *svcapitypes.VPCPeeringConnection {
					cr := buildVPCPeerConnection("test")
					cr.Status.SetConditions(Approved())

					return cr
				}(),
				client: &fake.MockEC2Client{
					DescribeVpcPeeringConnectionsRequestFun: func(input *ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest {
						return ec2.DescribeVpcPeeringConnectionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.DescribeVpcPeeringConnectionsOutput{
								//Attributes: attributes,
								VpcPeeringConnections: []ec2.VpcPeeringConnection{
									{
										Status: &ec2.VpcPeeringConnectionStateReason{
											Code: ec2.VpcPeeringConnectionStateReasonCodeActive,
										},

										Tags: []ec2.Tag{
											{
												Key:   aws.String("Name"),
												Value: aws.String("test"),
											},
										},
										VpcPeeringConnectionId: aws.String("pcx-xxx"),
									},
								},
							}},
						}
					},
				},
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"Update": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
					MockStatusUpdate: func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return nil
					},
				},
				cr: buildVPCPeerConnection("test"),
				client: &fake.MockEC2Client{
					DescribeVpcPeeringConnectionsRequestFun: func(input *ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest {
						return ec2.DescribeVpcPeeringConnectionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.DescribeVpcPeeringConnectionsOutput{
								//Attributes: attributes,
								VpcPeeringConnections: []ec2.VpcPeeringConnection{
									{
										Status: &ec2.VpcPeeringConnectionStateReason{
											Code: ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance,
										},

										Tags: []ec2.Tag{
											{
												Key:   aws.String("Name"),
												Value: aws.String("test"),
											},
										},
										VpcPeeringConnectionId: aws.String("pcx-xxx"),
									},
								},
							}},
						}
					},
				},
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client:        tc.client,
				kube:          tc.kube,
				route53Client: tc.route53Cli,
			}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	g := NewGomegaWithT(t)

	type want struct {
		result managed.ExternalCreation
		err    error
		vpcID  string
	}

	cases := map[string]struct {
		args
		want
	}{
		"Create": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
					MockStatusUpdate: func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
						return nil
					},
				},
				route53Cli: &fake.MockRoute53Client{},
				cr:         buildVPCPeerConnection("test"),
				client: &fake.MockEC2Client{
					CreateVpcPeeringConnectionRequestFun: func(input *ec2.CreateVpcPeeringConnectionInput) ec2.CreateVpcPeeringConnectionRequest {
						g.Expect(*input.PeerRegion).Should(Equal("peerRegion"))
						g.Expect(*input.PeerOwnerId).Should(Equal("peerOwner"))
						g.Expect(*input.PeerVpcId).Should(Equal("peerVpc"))
						g.Expect(*input.VpcId).Should(Equal("ownerVpc"))

						return ec2.CreateVpcPeeringConnectionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.CreateVpcPeeringConnectionOutput{
								//Attributes: attributes,
								VpcPeeringConnection: &ec2.VpcPeeringConnection{
									VpcPeeringConnectionId: aws.String("pcx-xxx"),
								},
							}},
						}
					},

					CreateTagsRequestFun: func(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
						g.Expect(len(input.Tags)).Should(Equal(1))
						g.Expect(*input.Tags[0].Key).Should(Equal("Name"))
						g.Expect(*input.Tags[0].Value).Should(Equal("test"))
						return ec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.CreateTagsOutput{}},
						}
					},

					DescribeRouteTablesRequestFun: func(input *ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest {
						return ec2.DescribeRouteTablesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &ec2.DescribeRouteTablesOutput{
								RouteTables: make([]ec2.RouteTable, 0),
							}},
						}
					},
				},
			},
			want: want{
				result: managed.ExternalCreation{},
				vpcID:  "pcx-xxx",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client:        tc.client,
				kube:          tc.kube,
				route53Client: tc.route53Cli,
			}
			_, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(meta.GetExternalName(tc.args.cr), tc.want.vpcID); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func buildVPCPeerConnection(name string) *svcapitypes.VPCPeeringConnection {
	cr := &svcapitypes.VPCPeeringConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},

		Spec: svcapitypes.VPCPeeringConnectionSpec{
			ForProvider: svcapitypes.VPCPeeringConnectionParameters{
				VPCID:       aws.String("ownerVpc"),
				Region:      "ownerRegion",
				HostZoneID:  aws.String("owner"),
				PeerOwnerID: aws.String("peerOwner"),
				PeerVPCID:   aws.String("peerVpc"),
				PeerRegion:  aws.String("peerRegion"),
				PeerCIDR:    aws.String("10.0.0.1/32"),
			},
		},
	}

	meta.SetExternalName(cr, name)

	return cr
}
