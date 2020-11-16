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

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
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

type natModifier func(*v1alpha1.NATGateway)

func withExternalName(name string) natModifier {
	return func(r *v1alpha1.NATGateway) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) natModifier {
	return func(r *v1alpha1.NATGateway) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.NATGatewayParameters) natModifier {
	return func(r *v1alpha1.NATGateway) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha1.NATGatewayObservation) natModifier {
	return func(r *v1alpha1.NATGateway) { r.Status.AtProvider = s }
}

func nat(m ...natModifier) *v1alpha1.NATGateway {
	cr := &v1alpha1.NATGateway{}
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

func specAddresses() []v1alpha1.NATGatewayAddress {
	return []v1alpha1.NATGatewayAddress{
		{
			AllocationID:       natAllocationID,
			NetworkInterfaceID: natNetworkInterfaceID,
			PrivateIP:          natPrivateIP,
			PublicIP:           natPublicIP,
		},
	}
}

func specNatStatus(state string, time time.Time, failureCode *string, failureMessage *string, delete bool) v1alpha1.NATGatewayObservation {
	observation := v1alpha1.NATGatewayObservation{
		CreateTime:          &v1.Time{Time: time},
		NatGatewayAddresses: specAddresses(),
		NatGatewayID:        natGatewayID,
		State:               state,
		VpcID:               natVpcID,
	}
	if state == v1alpha1.NatGatewayStatusFailed {
		observation.FailureCode = aws.StringValue(failureCode)
		observation.FailureMessage = aws.StringValue(failureMessage)
	}
	if delete {
		observation.DeleteTime = &v1.Time{Time: time}
	}
	return observation
}

func specNatSpec() v1alpha1.NATGatewayParameters {
	return v1alpha1.NATGatewayParameters{
		AllocationID: &natAllocationID,
		SubnetID:     &natSubnetID,
		Tags:         specTags(),
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

func natGatewayDescription(state awsec2.NatGatewayState, time time.Time, failureCode *string, failureMessage *string, delete bool) *awsec2.DescribeNatGatewaysOutput {
	natGatewayDescription := []awsec2.NatGateway{
		{
			CreateTime:          &time,
			NatGatewayAddresses: natAddresses(),
			NatGatewayId:        aws.String(natGatewayID),
			State:               state,
			SubnetId:            aws.String(natSubnetID),
			Tags:                natTags(),
			VpcId:               aws.String(natVpcID),
		},
	}
	if state == awsec2.NatGatewayStateFailed {
		natGatewayDescription[0].FailureCode = failureCode
		natGatewayDescription[0].FailureMessage = failureMessage
	}
	if delete {
		natGatewayDescription[0].DeleteTime = &time
	}
	return &awsec2.DescribeNatGatewaysOutput{
		NatGateways: natGatewayDescription,
	}
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

type args struct {
	nat  ec2.NatGatewayClient
	kube client.Client
	cr   *v1alpha1.NATGateway
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.NATGateway
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
								Data:        natGatewayDescription(awsec2.NatGatewayStatePending, time, nil, nil, false),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusPending, time, nil, nil, false)),
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
								Data:        natGatewayDescription(awsec2.NatGatewayStateFailed, time, aws.String(natFailureCode), aws.String(natFailureMessage), true),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusFailed, time, &natFailureCode, &natFailureMessage, true)),
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
								Data:        natGatewayDescription(awsec2.NatGatewayStateAvailable, time, nil, nil, false),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusAvailable, time, nil, nil, false)),
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
								Data:        natGatewayDescription(awsec2.NatGatewayStateDeleting, time, nil, nil, true),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleting, time, nil, nil, true)),
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
								Data:        natGatewayDescription(awsec2.NatGatewayStateDeleted, time, nil, nil, true),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleted, time, nil, nil, true)),
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

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.NATGateway
		result managed.ExternalCreation
		err    error
	}

	time := time.Now()

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(e *awsec2.CreateNatGatewayInput) awsec2.CreateNatGatewayRequest {
						return awsec2.CreateNatGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateNatGatewayOutput{
								NatGateway: &awsec2.NatGateway{
									CreateTime:          &time,
									NatGatewayAddresses: natAddresses(),
									NatGatewayId:        aws.String(natGatewayID),
									State:               awsec2.NatGatewayStatePending,
									SubnetId:            aws.String(natSubnetID),
									Tags:                natTags(),
									VpcId:               aws.String(natVpcID),
								},
							}},
						}
					},
				},
				cr: nat(withSpec(v1alpha1.NATGatewayParameters{
					AllocationID: &natAllocationID,
					SubnetID:     &natSubnetID,
					Tags:         specTags(),
				})),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedRequest": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(e *awsec2.CreateNatGatewayInput) awsec2.CreateNatGatewayRequest {
						return awsec2.CreateNatGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: nat(),
			},
			want: want{
				cr:  nat(),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.nat}
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
		cr     *v1alpha1.NATGateway
		result managed.ExternalUpdate
		err    error
	}

	time := time.Now()

	cases := map[string]struct {
		args
		want
	}{
		"TagsInSync": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        natGatewayDescription(awsec2.NatGatewayStateAvailable, time, nil, nil, false),
							},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusAvailable, time, nil, nil, false)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusAvailable, time, nil, nil, false))),
				result: managed.ExternalUpdate{},
			},
		},
		"TagsNotInSync": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(e *awsec2.DescribeNatGatewaysInput) awsec2.DescribeNatGatewaysRequest {
						return awsec2.DescribeNatGatewaysRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        natGatewayDescription(awsec2.NatGatewayStateAvailable, time, nil, nil, false),
							},
						}
					},
					MockCreateTags: func(e *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockDeleteTags: func(e *awsec2.DeleteTagsInput) awsec2.DeleteTagsRequest {
						return awsec2.DeleteTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteTagsOutput{}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1alpha1.NATGatewayParameters{
						AllocationID: aws.String(natAllocationID),
						SubnetID:     aws.String(natSubnetID),
						Tags: []v1beta1.Tag{
							{
								Key:   "somekey",
								Value: "somevalue",
							},
							{
								Key:   "somekey1",
								Value: "somevalue1",
							},
						},
					}),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusAvailable, time, nil, nil, false)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1alpha1.NATGatewayParameters{
						AllocationID: aws.String(natAllocationID),
						SubnetID:     aws.String(natSubnetID),
						Tags: []v1beta1.Tag{
							{
								Key:   "somekey",
								Value: "somevalue",
							},
							{
								Key:   "somekey1",
								Value: "somevalue1",
							},
						},
					}),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusAvailable, time, nil, nil, false))),
				result: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.nat}
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
		cr  *v1alpha1.NATGateway
		err error
	}

	time := time.Now()

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDelete: func(e *awsec2.DeleteNatGatewayInput) awsec2.DeleteNatGatewayRequest {
						return awsec2.DeleteNatGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteNatGatewayOutput{}},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec())),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(runtimev1alpha1.Deleting()),
					withSpec(specNatSpec()),
				),
				err: nil,
			},
		},
		"SkipDeleteForStateDeleting": {
			args: args{
				nat: &fake.MockNatGatewayClient{},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleting, time, nil, nil, true)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(runtimev1alpha1.Deleting()),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleting, time, nil, nil, true)),
				),
				err: nil,
			},
		},
		"SkipDeleteForStateDeleted": {
			args: args{
				nat: &fake.MockNatGatewayClient{},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleted, time, nil, nil, true)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(runtimev1alpha1.Deleting()),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1alpha1.NatGatewayStatusDeleted, time, nil, nil, true)),
				),
				err: nil,
			},
		},
		"DeleteFail": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDelete: func(e *awsec2.DeleteNatGatewayInput) awsec2.DeleteNatGatewayRequest {
						return awsec2.DeleteNatGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.nat}
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
