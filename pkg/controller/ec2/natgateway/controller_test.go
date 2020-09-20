package natgateway

import (
	"context"
	"net/http"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	natAllocationID       = "some allocation id"
	natNetworkInterfaceID = "some network interface id"
	natPrivateIP          = "some private ip"
	natPublicIP           = "some public ip"
	natGatewayID          = "some gateway id"
	natSubnetID           = "some subnet id"
	natVpcID              = "some vpc"
	natFailureCode        = "some failure code"
	natFailureMessage     = "some failure message"
	errBoom               = errors.New("nat boomed")
)

type natModifier func(*v1beta1.NatGateway)

func withExternalName(name string) natModifier {
	return func(r *v1beta1.NatGateway) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) natModifier {
	return func(r *v1beta1.NatGateway) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.NatGatewayParameters) natModifier {
	return func(r *v1beta1.NatGateway) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.NatGatewayObservation) natModifier {
	return func(r *v1beta1.NatGateway) { r.Status.AtProvider = s }
}

func nat(m ...natModifier) *v1beta1.NatGateway {
	cr := &v1beta1.NatGateway{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func natAddresses() []awsec2.NatGatewayAddress {
	return []awsec2.NatGatewayAddress{
		{
			AllocationId:       aws.String(natAllocationID),
			NetworkInterfaceId: aws.String(natNetworkInterfaceID),
			PrivateIp:          aws.String(natPrivateIP),
			PublicIp:           aws.String(natPublicIP),
		},
	}
}

func specAddresses() []v1beta1.NatGatewayAddress {
	return []v1beta1.NatGatewayAddress{
		{
			AllocationID:       natAllocationID,
			NetworkInterfaceID: natNetworkInterfaceID,
			PrivateIP:          natPrivateIP,
			PublicIP:           natPublicIP,
		},
	}
}

func natTags() []awsec2.Tag {
	return []awsec2.Tag{
		{
			Key:   aws.String("key1"),
			Value: aws.String("value1"),
		},
		{
			Key:   aws.String("key2"),
			Value: aws.String("value2"),
		},
	}
}

func specTags() []v1beta1.Tag {
	return []v1beta1.Tag{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

type args struct {
	nat  ec2.NatGatewayClient
	kube client.Client
	cr   *v1beta1.NatGateway
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.NatGateway
		result managed.ExternalObservation
		err    error
	}

	time := time.Now()

	cases := map[string]struct {
		args
		want
	}{
		"ExternalNameEmpty": {
			args: args{
				nat: &fake.MockNatGatewayClient{},
				cr:  nat(withExternalName("")),
			},
			want: want{
				cr: nat(withExternalName("")),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NatGatewayNotFound": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(ec2.NatGatewayNotFound, ec2.NatGatewayNotFound, errors.New(ec2.NatGatewayNotFound))},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr:     nat(withExternalName(natGatewayID)),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
		"ErrorDescribe": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr:     nat(withExternalName(natGatewayID)),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errDescribe),
			},
		},
		"ErrorMultipleNatAddresses": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{},
										{},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr:     nat(withExternalName(natGatewayID)),
				result: managed.ExternalObservation{},
				err:    errors.New(errNotSingleItem),
			},
		},
		"StatusPending": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{
											CreateTime:          &time,
											NatGatewayAddresses: natAddresses(),
											NatGatewayId:        aws.String(natGatewayID),
											State:               v1beta1.NatGatewayStatusPending,
											SubnetId:            aws.String(natSubnetID),
											Tags:                natTags(),
											VpcId:               aws.String(natVpcID),
										},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
					withStatus(v1beta1.NatGatewayObservation{
						CreateTime:          &v1.Time{Time: time},
						NatGatewayAddresses: specAddresses(),
						NatGatewayID:        natGatewayID,
						State:               v1beta1.NatGatewayStatusPending,
						SubnetID:            natSubnetID,
						Tags:                specTags(),
						VpcID:               natVpcID,
					}),
					withConditions(runtimev1alpha1.Unavailable()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"StatusFailed": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{
											CreateTime:          &time,
											DeleteTime:          &time,
											FailureCode:         aws.String(natFailureCode),
											FailureMessage:      aws.String(natFailureMessage),
											NatGatewayAddresses: natAddresses(),
											NatGatewayId:        aws.String(natGatewayID),
											State:               v1beta1.NatGatewayStatusFailed,
											SubnetId:            aws.String(natSubnetID),
											Tags:                natTags(),
											VpcId:               aws.String(natVpcID),
										},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
					withStatus(v1beta1.NatGatewayObservation{
						CreateTime:          &v1.Time{Time: time},
						DeleteTime:          &v1.Time{Time: time},
						FailureCode:         natFailureCode,
						FailureMessage:      natFailureMessage,
						NatGatewayAddresses: specAddresses(),
						NatGatewayID:        natGatewayID,
						State:               v1beta1.NatGatewayStatusFailed,
						SubnetID:            natSubnetID,
						Tags:                specTags(),
						VpcID:               natVpcID,
					}),
					withConditions(runtimev1alpha1.Unavailable().WithMessage(natFailureMessage)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"StatusAvailale": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{
											CreateTime:          &time,
											NatGatewayAddresses: natAddresses(),
											NatGatewayId:        aws.String(natGatewayID),
											State:               v1beta1.NatGatewayStatusAvailable,
											SubnetId:            aws.String(natSubnetID),
											Tags:                natTags(),
											VpcId:               aws.String(natVpcID),
										},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
					withStatus(v1beta1.NatGatewayObservation{
						CreateTime:          &v1.Time{Time: time},
						NatGatewayAddresses: specAddresses(),
						NatGatewayID:        natGatewayID,
						State:               v1beta1.NatGatewayStatusAvailable,
						SubnetID:            natSubnetID,
						Tags:                specTags(),
						VpcID:               natVpcID,
					}),
					withConditions(runtimev1alpha1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"StatusDeleting": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{
											CreateTime:          &time,
											DeleteTime:          &time,
											NatGatewayAddresses: natAddresses(),
											NatGatewayId:        aws.String(natGatewayID),
											State:               v1beta1.NatGatewayStatusDeleting,
											SubnetId:            aws.String(natSubnetID),
											Tags:                natTags(),
											VpcId:               aws.String(natVpcID),
										},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
					withStatus(v1beta1.NatGatewayObservation{
						CreateTime:          &v1.Time{Time: time},
						DeleteTime:          &v1.Time{Time: time},
						NatGatewayAddresses: specAddresses(),
						NatGatewayID:        natGatewayID,
						State:               v1beta1.NatGatewayStatusDeleting,
						SubnetID:            natSubnetID,
						Tags:                specTags(),
						VpcID:               natVpcID,
					}),
					withConditions(runtimev1alpha1.Deleting()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"StatusDeleted": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awsec2.DescribeNatGatewaysOutput{
									NatGateways: []awsec2.NatGateway{
										{
											CreateTime:          &time,
											DeleteTime:          &time,
											NatGatewayAddresses: natAddresses(),
											NatGatewayId:        aws.String(natGatewayID),
											State:               v1beta1.NatGatewayStatusDeleted,
											SubnetId:            aws.String(natSubnetID),
											Tags:                natTags(),
											VpcId:               aws.String(natVpcID),
										},
									},
								}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NatGatewayParameters{
						AllocationID: &natAllocationID,
						SubnetID:     &natSubnetID,
						Tags:         specTags(),
					}),
					withStatus(v1beta1.NatGatewayObservation{
						CreateTime:          &v1.Time{Time: time},
						DeleteTime:          &v1.Time{Time: time},
						NatGatewayAddresses: specAddresses(),
						NatGatewayID:        natGatewayID,
						State:               v1beta1.NatGatewayStatusDeleted,
						SubnetID:            natSubnetID,
						Tags:                specTags(),
						VpcID:               natVpcID,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.nat}
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
