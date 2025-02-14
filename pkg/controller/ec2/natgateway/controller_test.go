package natgateway

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	natAllocationID         = "some allocation id"
	natNetworkInterfaceID   = "some network interface id"
	natPrivateIP            = "some private ip"
	natPublicIP             = "some public ip"
	natGatewayID            = "some gateway id"
	natSubnetID             = "some subnet id"
	natVpcID                = "some vpc"
	natFailureCode          = "some failure code"
	natFailureMessage       = "some failure message"
	connectivityTypePrivate = "private"
	connectivityTypePublic  = "public"
	errBoom                 = errors.New("nat boomed")
)

type natModifier func(*v1beta1.NATGateway)

func withExternalName(name string) natModifier {
	return func(r *v1beta1.NATGateway) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) natModifier {
	return func(r *v1beta1.NATGateway) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.NATGatewayParameters) natModifier {
	return func(r *v1beta1.NATGateway) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.NATGatewayObservation) natModifier {
	return func(r *v1beta1.NATGateway) { r.Status.AtProvider = s }
}

func nat(m ...natModifier) *v1beta1.NATGateway {
	cr := &v1beta1.NATGateway{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func natAddresses() []awsec2types.NatGatewayAddress {
	return []awsec2types.NatGatewayAddress{
		{
			AllocationId:       aws.String(natAllocationID),
			NetworkInterfaceId: aws.String(natNetworkInterfaceID),
			PrivateIp:          aws.String(natPrivateIP),
			PublicIp:           aws.String(natPublicIP),
		},
	}
}

func specAddresses() []v1beta1.NATGatewayAddress {
	return []v1beta1.NATGatewayAddress{
		{
			AllocationID:       natAllocationID,
			NetworkInterfaceID: natNetworkInterfaceID,
			PrivateIP:          natPrivateIP,
			PublicIP:           natPublicIP,
		},
	}
}

func specNatStatus(state string, time time.Time, failureCode *string, failureMessage *string, delete bool) v1beta1.NATGatewayObservation {
	observation := v1beta1.NATGatewayObservation{
		CreateTime:          &metav1.Time{Time: time},
		NatGatewayAddresses: specAddresses(),
		NatGatewayID:        natGatewayID,
		State:               state,
		VpcID:               natVpcID,
	}
	if state == v1beta1.NatGatewayStatusFailed {
		observation.FailureCode = aws.ToString(failureCode)
		observation.FailureMessage = aws.ToString(failureMessage)
	}
	if delete {
		observation.DeleteTime = &metav1.Time{Time: time}
	}
	return observation
}

func specNatSpec() v1beta1.NATGatewayParameters {
	return v1beta1.NATGatewayParameters{
		AllocationID: &natAllocationID,
		SubnetID:     &natSubnetID,
		Tags:         specTags(),
	}
}

func specNatPrivateSpec() v1beta1.NATGatewayParameters {
	return v1beta1.NATGatewayParameters{
		ConnectivityType: connectivityTypePrivate,
		SubnetID:         &natSubnetID,
		Tags:             specTags(),
	}
}

func specNatPublicSpec() v1beta1.NATGatewayParameters {
	return v1beta1.NATGatewayParameters{
		ConnectivityType: connectivityTypePublic,
		AllocationID:     &natAllocationID,
		SubnetID:         &natSubnetID,
		Tags:             specTags(),
	}
}

func natTags() []awsec2types.Tag {
	return []awsec2types.Tag{
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

func natGatewayDescription(state awsec2types.NatGatewayState, time time.Time, failureCode *string, failureMessage *string, delete bool) *awsec2.DescribeNatGatewaysOutput {
	natGatewayDescription := []awsec2types.NatGateway{
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
	if state == awsec2types.NatGatewayStateFailed {
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
	cr   *v1beta1.NATGateway
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.NATGateway
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return nil, &smithy.GenericAPIError{Code: ec2.NatGatewayNotFound}
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return nil, errBoom
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr:     nat(withExternalName(natGatewayID)),
				result: managed.ExternalObservation{},
				err:    errorutils.Wrap(errBoom, errDescribe),
			},
		},
		"ErrorMultipleNatAddresses": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return &awsec2.DescribeNatGatewaysOutput{
							NatGateways: []awsec2types.NatGateway{
								{},
								{},
							},
						}, nil
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStatePending, time, nil, nil, false), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusPending, time, nil, nil, false)),
					withConditions(xpv1.Unavailable()),
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateFailed, time, aws.String(natFailureCode), aws.String(natFailureMessage), true), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusFailed, time, &natFailureCode, &natFailureMessage, true)),
					withConditions(xpv1.Unavailable().WithMessage(natFailureMessage)),
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateAvailable, time, nil, nil, false), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusAvailable, time, nil, nil, false)),
					withConditions(xpv1.Available()),
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateDeleting, time, nil, nil, true), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleting, time, nil, nil, true)),
					withConditions(xpv1.Deleting()),
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateDeleted, time, nil, nil, true), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleted, time, nil, nil, true)),
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
		cr     *v1beta1.NATGateway
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
					MockCreate: func(ctx context.Context, input *awsec2.CreateNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.CreateNatGatewayOutput, error) {
						return &awsec2.CreateNatGatewayOutput{
							NatGateway: &awsec2types.NatGateway{
								CreateTime:          &time,
								NatGatewayAddresses: natAddresses(),
								NatGatewayId:        aws.String(natGatewayID),
								State:               awsec2types.NatGatewayStatePending,
								SubnetId:            aws.String(natSubnetID),
								Tags:                natTags(),
								VpcId:               aws.String(natVpcID),
							},
						}, nil
					},
				},
				cr: nat(withSpec(v1beta1.NATGatewayParameters{
					AllocationID: &natAllocationID,
					SubnetID:     &natSubnetID,
					Tags:         specTags(),
				})),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulPublic": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.CreateNatGatewayOutput, error) {
						return &awsec2.CreateNatGatewayOutput{
							NatGateway: &awsec2types.NatGateway{
								CreateTime:          &time,
								ConnectivityType:    awsec2types.ConnectivityTypePublic,
								NatGatewayAddresses: natAddresses(),
								NatGatewayId:        aws.String(natGatewayID),
								State:               awsec2types.NatGatewayStatePending,
								SubnetId:            aws.String(natSubnetID),
								Tags:                natTags(),
								VpcId:               aws.String(natVpcID),
							},
						}, nil
					},
				},
				cr: nat(withSpec(v1beta1.NATGatewayParameters{
					AllocationID: &natAllocationID,
					SubnetID:     &natSubnetID,
					Tags:         specTags(),
				})),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulPublicSet": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.CreateNatGatewayOutput, error) {
						return &awsec2.CreateNatGatewayOutput{
							NatGateway: &awsec2types.NatGateway{
								CreateTime:          &time,
								ConnectivityType:    awsec2types.ConnectivityTypePublic,
								NatGatewayAddresses: natAddresses(),
								NatGatewayId:        aws.String(natGatewayID),
								State:               awsec2types.NatGatewayStatePending,
								SubnetId:            aws.String(natSubnetID),
								Tags:                natTags(),
								VpcId:               aws.String(natVpcID),
							},
						}, nil
					},
				},
				cr: nat(withSpec(v1beta1.NATGatewayParameters{
					ConnectivityType: string(awsec2types.ConnectivityTypePublic),
					AllocationID:     &natAllocationID,
					SubnetID:         &natSubnetID,
					Tags:             specTags(),
				})),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatPublicSpec()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulPrivate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.CreateNatGatewayOutput, error) {
						return &awsec2.CreateNatGatewayOutput{
							NatGateway: &awsec2types.NatGateway{
								CreateTime:          &time,
								ConnectivityType:    awsec2types.ConnectivityTypePrivate,
								NatGatewayAddresses: natAddresses(),
								NatGatewayId:        aws.String(natGatewayID),
								State:               awsec2types.NatGatewayStatePending,
								SubnetId:            aws.String(natSubnetID),
								Tags:                natTags(),
								VpcId:               aws.String(natVpcID),
							},
						}, nil
					},
				},
				cr: nat(withSpec(v1beta1.NATGatewayParameters{
					ConnectivityType: string(awsec2types.ConnectivityTypePrivate),
					SubnetID:         &natSubnetID,
					Tags:             specTags(),
				})),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatPrivateSpec()),
				),
				result: managed.ExternalCreation{},
			},
		},

		"FailedRequest": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				nat: &fake.MockNatGatewayClient{
					MockCreate: func(ctx context.Context, input *awsec2.CreateNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.CreateNatGatewayOutput, error) {
						return nil, errBoom
					},
				},
				cr: nat(),
			},
			want: want{
				cr:  nat(),
				err: errorutils.Wrap(errBoom, errCreate),
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
		cr     *v1beta1.NATGateway
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
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateAvailable, time, nil, nil, false), nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusAvailable, time, nil, nil, false)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusAvailable, time, nil, nil, false))),
				result: managed.ExternalUpdate{},
			},
		},
		"TagsNotInSync": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeNatGatewaysInput, opts []func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error) {
						return natGatewayDescription(awsec2types.NatGatewayStateAvailable, time, nil, nil, false), nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
					MockDeleteTags: func(ctx context.Context, input *awsec2.DeleteTagsInput, opts []func(*awsec2.Options)) (*awsec2.DeleteTagsOutput, error) {
						return &awsec2.DeleteTagsOutput{}, nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NATGatewayParameters{
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
					withStatus(specNatStatus(v1beta1.NatGatewayStatusAvailable, time, nil, nil, false)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withSpec(v1beta1.NATGatewayParameters{
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
					withStatus(specNatStatus(v1beta1.NatGatewayStatusAvailable, time, nil, nil, false))),
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
		cr  *v1beta1.NATGateway
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
					MockDelete: func(ctx context.Context, input *awsec2.DeleteNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.DeleteNatGatewayOutput, error) {
						return &awsec2.DeleteNatGatewayOutput{}, nil
					},
				},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec())),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(xpv1.Deleting()),
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
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleting, time, nil, nil, true)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(xpv1.Deleting()),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleting, time, nil, nil, true)),
				),
				err: nil,
			},
		},
		"SkipDeleteForStateDeleted": {
			args: args{
				nat: &fake.MockNatGatewayClient{},
				cr: nat(withExternalName(natGatewayID),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleted, time, nil, nil, true)),
				),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(xpv1.Deleting()),
					withSpec(specNatSpec()),
					withStatus(specNatStatus(v1beta1.NatGatewayStatusDeleted, time, nil, nil, true)),
				),
				err: nil,
			},
		},
		"DeleteFail": {
			args: args{
				nat: &fake.MockNatGatewayClient{
					MockDelete: func(ctx context.Context, input *awsec2.DeleteNatGatewayInput, opts []func(*awsec2.Options)) (*awsec2.DeleteNatGatewayOutput, error) {
						return nil, errBoom
					},
				},
				cr: nat(withExternalName(natGatewayID)),
			},
			want: want{
				cr: nat(withExternalName(natGatewayID),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.nat}
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
