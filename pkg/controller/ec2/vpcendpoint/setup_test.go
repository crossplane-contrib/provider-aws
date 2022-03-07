package vpcendpoint

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	aws "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	testVPCID           = "vpc-id"
	testVPCEndpointID   = "vpcendpoint-id"
	testSecurityGroupID = "sg-id"
	testSubnetID1       = "subnet-id-1"
	testSubnetID2       = "subnet-id-2"

	testErrCreateVPCEndpointFailed   = "CreateVPCEndpoint failed"
	testErrDeleteVPCEndpointFailed   = "DeleteVPCEndpoint failed"
	testErrDescribeVPCEndpointFailed = "DescribeVPCEndpoint failed"
	testErrUpdateVPCEndpointFailed   = "UpdateVPCEndpoint failed"
)

type args struct {
	vpcendpoint *fake.MockVPCEndpointClient
	kube        client.Client
	cr          *v1alpha1.VPCEndpoint
}

type vpcEndpointModifier func(*v1alpha1.VPCEndpoint)

func withExternalName(name string) vpcEndpointModifier {
	return func(r *v1alpha1.VPCEndpoint) { meta.SetExternalName(r, name) }
}

func withSpec(p v1alpha1.VPCEndpointParameters) vpcEndpointModifier {
	return func(o *v1alpha1.VPCEndpoint) { o.Spec.ForProvider = p }
}

func withStatusAtProvider(statusAtProvider v1alpha1.VPCEndpointObservation) vpcEndpointModifier {
	return func(o *v1alpha1.VPCEndpoint) { o.Status.AtProvider = statusAtProvider }
}

func withConditions(value ...xpv1.Condition) vpcEndpointModifier {
	return func(o *v1alpha1.VPCEndpoint) { o.Status.SetConditions(value...) }
}

func vpcEndpoint(m ...vpcEndpointModifier) *v1alpha1.VPCEndpoint {
	cr := &v1alpha1.VPCEndpoint{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.VPCEndpoint
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockCreateVpcEndpointWithContext: func(ctx context.Context, input *ec2.CreateVpcEndpointInput, req ...request.Option) (*ec2.CreateVpcEndpointOutput, error) {
						return &ec2.CreateVpcEndpointOutput{
							VpcEndpoint: &ec2.VpcEndpoint{
								VpcEndpointId: aws.String(testVPCEndpointID),
							},
						}, nil
					},
				},
				cr: vpcEndpoint(
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					})),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Creating()),
					withStatusAtProvider(v1alpha1.VPCEndpointObservation{
						VPCEndpointID: aws.String(testVPCEndpointID),
					}),
				),
			},
		},
		"ErrCreate": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockCreateVpcEndpointWithContext: func(ctx context.Context, input *ec2.CreateVpcEndpointInput, req ...request.Option) (*ec2.CreateVpcEndpointOutput, error) {
						return &ec2.CreateVpcEndpointOutput{}, errors.New(testErrCreateVPCEndpointFailed)
					},
				},
				cr: vpcEndpoint(
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					})),
			},
			want: want{
				cr: vpcEndpoint(
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{
					ExternalNameAssigned: false,
				},
				err: errors.Wrap(errors.New(testErrCreateVPCEndpointFailed), errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.vpcendpoint, opts)
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
		cr  *v1alpha1.VPCEndpoint
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDeleteVpcEndpoints: func(*ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error) {
						return &ec2.DeleteVpcEndpointsOutput{}, nil
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"ErrDelete": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDeleteVpcEndpoints: func(*ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error) {
						return &ec2.DeleteVpcEndpointsOutput{}, errors.New(testErrDeleteVPCEndpointFailed)
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Deleting()),
				),
				err: errors.New(testErrDeleteVPCEndpointFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.vpcendpoint, opts)
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

func TestModify(t *testing.T) {
	type want struct {
		cr     *v1alpha1.VPCEndpoint
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulModify": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockModifyVpcEndpointWithContext: func(ctx context.Context, input *ec2.ModifyVpcEndpointInput, req ...request.Option) (*ec2.ModifyVpcEndpointOutput, error) {
						return &ec2.ModifyVpcEndpointOutput{}, nil
					},
					MockDescribeVpcEndpoints: func(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{
							VpcEndpoints: []*ec2.VpcEndpoint{
								{
									VpcEndpointId: aws.String(testVPCEndpointID),
									SubnetIds: []*string{
										aws.String(testSubnetID1),
										aws.String(testSubnetID2),
									},
									Groups: []*ec2.SecurityGroupIdentifier{
										{
											GroupId: aws.String(testSecurityGroupID),
										},
									},
								},
							},
						}, nil
					},
				},
				cr: vpcEndpoint(),
			},
			want: want{
				cr: vpcEndpoint(
					withConditions(xpv1.Creating()),
				),
			},
		},
		"ErrModify_DescribeEndpointOutput": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockModifyVpcEndpointWithContext: func(ctx context.Context, input *ec2.ModifyVpcEndpointInput, req ...request.Option) (*ec2.ModifyVpcEndpointOutput, error) {
						return &ec2.ModifyVpcEndpointOutput{}, nil
					},
					MockDescribeVpcEndpoints: func(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{}, errors.New(testErrDescribeVPCEndpointFailed)
					},
				},
				cr: vpcEndpoint(),
			},
			want: want{
				cr:  vpcEndpoint(),
				err: errors.Wrap(errors.New(testErrDescribeVPCEndpointFailed), "pre-update failed"),
			},
		},
		"ErrModify_ModifyVpcEndpoint": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockModifyVpcEndpointWithContext: func(ctx context.Context, input *ec2.ModifyVpcEndpointInput, req ...request.Option) (*ec2.ModifyVpcEndpointOutput, error) {
						return &ec2.ModifyVpcEndpointOutput{}, errors.New(testErrUpdateVPCEndpointFailed)
					},
					MockDescribeVpcEndpoints: func(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{
							VpcEndpoints: []*ec2.VpcEndpoint{
								{
									VpcEndpointId: aws.String(testVPCEndpointID),
									SubnetIds: []*string{
										aws.String(testSubnetID1),
										aws.String(testSubnetID2),
									},
									Groups: []*ec2.SecurityGroupIdentifier{
										{
											GroupId: aws.String(testSecurityGroupID),
										},
									},
								},
							},
						}, nil
					},
				},
				cr: vpcEndpoint(),
			},
			want: want{
				cr:  vpcEndpoint(),
				err: errors.Wrap(errors.New(testErrUpdateVPCEndpointFailed), errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.vpcendpoint, opts)
			res, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, res, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.VPCEndpoint
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableState_and_UpToDate": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDescribeVpcEndpointsWithContext: func(context.Context, *ec2.DescribeVpcEndpointsInput, ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{
							VpcEndpoints: []*ec2.VpcEndpoint{
								{
									VpcEndpointId:     aws.String(testVPCEndpointID),
									CreationTimestamp: &time.Time{},
								},
							},
						}, nil
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
					withStatusAtProvider(v1alpha1.VPCEndpointObservation{
						CreationTimestamp: &v1.Time{},
						DNSEntries:        []*v1alpha1.DNSEntry{},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"AvailableState_but_Deleting": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDescribeVpcEndpointsWithContext: func(context.Context, *ec2.DescribeVpcEndpointsInput, ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{
							VpcEndpoints: []*ec2.VpcEndpoint{
								{
									VpcEndpointId:     aws.String(testVPCEndpointID),
									CreationTimestamp: &time.Time{},
									State:             aws.String("deleting"),
								},
							},
						}, nil
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Deleting()),
					withStatusAtProvider(v1alpha1.VPCEndpointObservation{
						CreationTimestamp: &v1.Time{},
						DNSEntries:        []*v1alpha1.DNSEntry{},
						State:             aws.String("deleting"),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"CreatingState_but_NowAvailable": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDescribeVpcEndpointsWithContext: func(context.Context, *ec2.DescribeVpcEndpointsInput, ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{
							VpcEndpoints: []*ec2.VpcEndpoint{
								{
									VpcEndpointId:     aws.String(testVPCEndpointID),
									CreationTimestamp: &time.Time{},
									State:             aws.String("available"),
								},
							},
						}, nil
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Creating()),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
					withStatusAtProvider(v1alpha1.VPCEndpointObservation{
						CreationTimestamp: &v1.Time{},
						DNSEntries:        []*v1alpha1.DNSEntry{},
						State:             aws.String("available"),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ErrInDescribing": {
			args: args{
				vpcendpoint: &fake.MockVPCEndpointClient{
					MockDescribeVpcEndpointsWithContext: func(context.Context, *ec2.DescribeVpcEndpointsInput, ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error) {
						return &ec2.DescribeVpcEndpointsOutput{}, errors.New(testErrDescribeVPCEndpointFailed)
					},
				},
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
				),
			},
			want: want{
				cr: vpcEndpoint(
					withExternalName(testVPCEndpointID),
					withSpec(v1alpha1.VPCEndpointParameters{
						CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
							VPCID: aws.String(testVPCID),
						},
					}),
					withConditions(xpv1.Available()),
				),
				err: errors.Wrap(errors.New(testErrDescribeVPCEndpointFailed), errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.vpcendpoint, opts)
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
