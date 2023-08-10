package capacityreservation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	awsClient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
	capacityReservationTesting "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/capacityreservation/testing"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

func TestObserve(t *testing.T) {
	type args struct {
		kube   client.Client
		client ec2iface.EC2API
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{
			name: "ClientError",
			args: args{
				client: &fake.MockCapacityResourceClient{
					DescribeCapacityReservationsErr: errBoom,
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr:  capacityReservationTesting.CapacityReservation(),
				err: awsClient.Wrap(errBoom, errDescribe),
			},
		},
		{
			name: "ResourceDoesNotExist",
			args: args{
				client: &fake.MockCapacityResourceClient{
					DescribeCapacityReservationsOutput: ec2.DescribeCapacityReservationsOutput{},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr:     capacityReservationTesting.CapacityReservation(),
				result: managed.ExternalObservation{},
			},
		},
		{
			name: "UpToDate",
			args: args{
				client: &fake.MockCapacityResourceClient{
					DescribeCapacityReservationsOutput: ec2.DescribeCapacityReservationsOutput{CapacityReservations: []*ec2.CapacityReservation{
						&ec2.CapacityReservation{
							CapacityReservationArn: aws.String("test.capacityReservation.name"),
							State:                  aws.String(ec2.CapacityReservationStateActive),
						},
					}},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationARN: aws.String("test.capacityReservation.name"),
							State:                  aws.String(ec2.CapacityReservationStateActive),
						},
					),
					capacityReservationTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		{
			name: "NotUpToDate",
			args: args{
				client: &fake.MockCapacityResourceClient{
					DescribeCapacityReservationsOutput: ec2.DescribeCapacityReservationsOutput{CapacityReservations: []*ec2.CapacityReservation{
						&ec2.CapacityReservation{
							CapacityReservationArn: aws.String("test.capacityReservation.name"),
							State:                  aws.String(ec2.CapacityReservationStateActive),
							TotalInstanceCount:     aws.Int64(2),
						},
					}},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationARN: aws.String("test.capacityReservation.name"),
							State:                  aws.String(ec2.CapacityReservationStateActive),
						},
					),
					capacityReservationTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		//{name: "ValidInputWithPolicy",
		//	args: args{
		//		client: &fake.MockS3ControlClient{
		//			GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
		//			GetAccessPointPolicyErr:         s3controlTesting.NoSuchAccessPointPolicy(),
		//		},
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV1),
		//		),
		//	},
		//	want: want{
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV1),
		//			s3controlTesting.WithConditions(xpv1.Available()),
		//		),
		//		result: managed.ExternalObservation{
		//			ResourceExists:   true,
		//			ResourceUpToDate: false,
		//		},
		//	},
		//},
		//{name: "AccessPointPolicyUpToDate",
		//	args: args{
		//		client: &fake.MockS3ControlClient{
		//			GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
		//			GetAccessPointPolicyOutput:      testOutputPolicyV1,
		//		},
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV1),
		//		),
		//	},
		//	want: want{
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV1),
		//			s3controlTesting.WithConditions(xpv1.Available()),
		//		),
		//		result: managed.ExternalObservation{
		//			ResourceExists:   true,
		//			ResourceUpToDate: true,
		//		},
		//	},
		//},
		//{name: "AccessPointPolicyNotUpToDate",
		//	args: args{
		//		client: &fake.MockS3ControlClient{
		//			GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
		//			GetAccessPointPolicyOutput:      testOutputPolicyV1,
		//		},
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV2),
		//		),
		//	},
		//	want: want{
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV2),
		//			s3controlTesting.WithConditions(xpv1.Available()),
		//		),
		//		result: managed.ExternalObservation{
		//			ResourceExists:   true,
		//			ResourceUpToDate: false,
		//		},
		//	},
		//},
		//{name: "AccessPointPolicyNotUpToDateNeedsDeletion",
		//	args: args{
		//		client: &fake.MockS3ControlClient{
		//			GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
		//			GetAccessPointPolicyOutput:      testOutputPolicyV1,
		//		},
		//		cr: s3controlTesting.AccessPoint(),
		//	},
		//	want: want{
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithConditions(xpv1.Available()),
		//		),
		//		result: managed.ExternalObservation{
		//			ResourceExists:   true,
		//			ResourceUpToDate: false,
		//		},
		//	},
		//},
		//{name: "FailedToDescribeAccessPointPolicy",
		//	args: args{
		//		client: &fake.MockS3ControlClient{
		//			GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
		//			GetAccessPointPolicyErr:         errBoom,
		//		},
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV2),
		//		),
		//	},
		//	want: want{
		//		cr: s3controlTesting.AccessPoint(
		//			s3controlTesting.WithPolicy(testPolicyV2),
		//		),
		//		err: errors.Wrap(awsClient.Wrap(errBoom, errDescribePolicy), "isUpToDate check failed"),
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Observe(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
