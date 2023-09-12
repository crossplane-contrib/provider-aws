package capacityreservation

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
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
							CapacityReservationId: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateActive),
						},
					}},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationID: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateActive),
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
							CapacityReservationId: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateActive),
							TotalInstanceCount:    aws.Int64(2),
						},
					}},
				},
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithSpec(svcapitypes.CapacityReservationParameters{
						InstanceCount: aws.Int64(3),
					})),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationID: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateActive),
							TotalInstanceCount:    aws.Int64(2),
						},
					),
					capacityReservationTesting.WithSpec(
						svcapitypes.CapacityReservationParameters{
							InstanceCount: aws.Int64(3),
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
		{
			name: "RecreateAlreadyDeletedResource",
			args: args{
				client: &fake.MockCapacityResourceClient{
					DescribeCapacityReservationsOutput: ec2.DescribeCapacityReservationsOutput{CapacityReservations: []*ec2.CapacityReservation{
						&ec2.CapacityReservation{
							CapacityReservationId: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateCancelled),
						},
					}},
				},
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							State: aws.String(ec2.CapacityReservationStateActive),
						}),
				),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationID: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStateCancelled),
						},
					),
					capacityReservationTesting.WithExternalName(""),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
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

func TestCreate(t *testing.T) {
	type args struct {
		kube   client.Client
		client ec2iface.EC2API
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "ClientError",
			args: args{
				client: &fake.MockCapacityResourceClient{
					CreateCapacityReservationErr: errBoom,
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.Creating()),
				),
				err: awsClient.Wrap(errBoom, errCreate),
			},
		},
		{name: "ValidInput",
			args: args{
				client: &fake.MockCapacityResourceClient{
					CreateCapacityReservationOutput: ec2.CreateCapacityReservationOutput{
						CapacityReservation: &ec2.CapacityReservation{
							CapacityReservationId: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStatePending),
						},
					},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.Creating()),
					capacityReservationTesting.WithStatus(
						svcapitypes.CapacityReservationObservation{
							CapacityReservationID: aws.String("test.capacityReservation.name"),
							State:                 aws.String(ec2.CapacityReservationStatePending),
						},
					),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Create(context.Background(), tt.args.cr)
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

func TestDelete(t *testing.T) {
	type args struct {
		kube   client.Client
		client ec2iface.EC2API
		cr     resource.Managed
	}

	type want struct {
		cr  resource.Managed
		err error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "ClientError",
			args: args{
				client: &fake.MockCapacityResourceClient{
					CancelCapacityReservationErr: errBoom,
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.Deleting()),
				),
				err: awsClient.Wrap(errBoom, errDelete),
			},
		},
		{name: "ValidInput",
			args: args{
				client: &fake.MockCapacityResourceClient{
					CancelCapacityReservationOutput: ec2.CancelCapacityReservationOutput{},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.Deleting()),
				),
			},
		},
		{name: "ResourceDoesNotExist",
			args: args{
				client: &fake.MockCapacityResourceClient{
					CancelCapacityReservationErr: awserr.New("NoSuchCapacityReservation", "", nil),
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.Deleting()),
				),
				err: awsClient.Wrap(awserr.New("NoSuchCapacityReservation", "", nil), errDelete),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			err := e.Delete(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		kube   client.Client
		client ec2iface.EC2API
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "GetCapacityReservationClientError",
			args: args{
				client: &fake.MockCapacityResourceClient{
					ModifyCapacityReservationErr: errBoom,
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.ReconcileError(awsClient.Wrap(errBoom, errUpdate))),
				),
				err: awsClient.Wrap(errBoom, errUpdate),
			},
		},
		{name: "UpdateCapacityReservation",
			args: args{
				client: &fake.MockCapacityResourceClient{
					ModifyCapacityReservationOutput: ec2.ModifyCapacityReservationOutput{Return: aws.Bool(true)},
				},
				cr: capacityReservationTesting.CapacityReservation(),
			},
			want: want{
				cr: capacityReservationTesting.CapacityReservation(
					capacityReservationTesting.WithConditions(xpv1.ReconcileSuccess()),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Update(context.Background(), tt.args.cr)
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
