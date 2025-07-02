/*
Copyright 2021 The Crossplane Authors.

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

package vpcendpointserviceallowedprincipal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	serviceID    = "vpce-svc-123456789abcdef01"
	principalARN = "arn:aws:iam::123456789012:root"
	externalName = "vpce-svc-123456789abcdef01:arn:aws:iam::123456789012:root"
	region       = "us-east-1"

	errBoom = errors.New("boom")
)

type args struct {
	client VPCEndpointServiceAllowedPrincipalClient
	kube   client.Client
	cr     *manualv1alpha1.VPCEndpointServiceAllowedPrincipal
}

type vpcEndpointServiceAllowedPrincipalModifier func(*manualv1alpha1.VPCEndpointServiceAllowedPrincipal)

func withExternalName(name string) vpcEndpointServiceAllowedPrincipalModifier {
	return func(r *manualv1alpha1.VPCEndpointServiceAllowedPrincipal) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) vpcEndpointServiceAllowedPrincipalModifier {
	return func(r *manualv1alpha1.VPCEndpointServiceAllowedPrincipal) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters) vpcEndpointServiceAllowedPrincipalModifier {
	return func(r *manualv1alpha1.VPCEndpointServiceAllowedPrincipal) { r.Spec.ForProvider = p }
}

func withStatus(s manualv1alpha1.VPCEndpointServiceAllowedPrincipalObservation) vpcEndpointServiceAllowedPrincipalModifier {
	return func(r *manualv1alpha1.VPCEndpointServiceAllowedPrincipal) { r.Status.AtProvider = s }
}

func vpcEndpointServiceAllowedPrincipal(m ...vpcEndpointServiceAllowedPrincipalModifier) *manualv1alpha1.VPCEndpointServiceAllowedPrincipal {
	cr := &manualv1alpha1.VPCEndpointServiceAllowedPrincipal{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

// Mock client for testing
type MockVPCEndpointServiceAllowedPrincipalClient struct {
	MockDescribeVpcEndpointServicePermissions func(ctx context.Context, input *awsec2.DescribeVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointServicePermissionsOutput, error)
	MockModifyVpcEndpointServicePermissions   func(ctx context.Context, input *awsec2.ModifyVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.ModifyVpcEndpointServicePermissionsOutput, error)
}

func (m *MockVPCEndpointServiceAllowedPrincipalClient) DescribeVpcEndpointServicePermissions(ctx context.Context, input *awsec2.DescribeVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointServicePermissionsOutput, error) {
	return m.MockDescribeVpcEndpointServicePermissions(ctx, input, opts...)
}

func (m *MockVPCEndpointServiceAllowedPrincipalClient) ModifyVpcEndpointServicePermissions(ctx context.Context, input *awsec2.ModifyVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.ModifyVpcEndpointServicePermissionsOutput, error) {
	return m.MockModifyVpcEndpointServicePermissions(ctx, input, opts...)
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.VPCEndpointServiceAllowedPrincipal
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				client: &MockVPCEndpointServiceAllowedPrincipalClient{
					MockDescribeVpcEndpointServicePermissions: func(ctx context.Context, input *awsec2.DescribeVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointServicePermissionsOutput, error) {
						return &awsec2.DescribeVpcEndpointServicePermissionsOutput{
							AllowedPrincipals: []types.AllowedPrincipal{
								{
									Principal:           aws.String(principalARN),
									PrincipalType:       types.PrincipalTypeAccount,
									ServicePermissionId: aws.String("perm-123"),
									ServiceId:           aws.String(serviceID),
								},
							},
						}, nil
					},
				},
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				}), withExternalName(externalName)),
			},
			want: want{
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				}), withStatus(manualv1alpha1.VPCEndpointServiceAllowedPrincipalObservation{
					Principal:           aws.String(principalARN),
					PrincipalType:       aws.String("Account"),
					ServicePermissionID: aws.String("perm-123"),
					ServiceID:           aws.String(serviceID),
				}), withExternalName(externalName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NotFound": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				client: &MockVPCEndpointServiceAllowedPrincipalClient{
					MockDescribeVpcEndpointServicePermissions: func(ctx context.Context, input *awsec2.DescribeVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointServicePermissionsOutput, error) {
						return &awsec2.DescribeVpcEndpointServicePermissionsOutput{
							AllowedPrincipals: []types.AllowedPrincipal{},
						}, nil
					},
				},
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				}), withExternalName(externalName)),
			},
			want: want{
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				}), withExternalName(externalName)),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr     *manualv1alpha1.VPCEndpointServiceAllowedPrincipal
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &MockVPCEndpointServiceAllowedPrincipalClient{
					MockModifyVpcEndpointServicePermissions: func(ctx context.Context, input *awsec2.ModifyVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.ModifyVpcEndpointServicePermissionsOutput, error) {
						return &awsec2.ModifyVpcEndpointServicePermissionsOutput{
							ReturnValue: aws.Bool(true),
						}, nil
					},
				},
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				})),
			},
			want: want{
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				}), withExternalName(externalName)),
				result: managed.ExternalCreation{},
			},
		},
		"Failed": {
			args: args{
				client: &MockVPCEndpointServiceAllowedPrincipalClient{
					MockModifyVpcEndpointServicePermissions: func(ctx context.Context, input *awsec2.ModifyVpcEndpointServicePermissionsInput, opts ...func(*awsec2.Options)) (*awsec2.ModifyVpcEndpointServicePermissionsOutput, error) {
						return &awsec2.ModifyVpcEndpointServicePermissionsOutput{}, errBoom
					},
				},
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				})),
			},
			want: want{
				cr: vpcEndpointServiceAllowedPrincipal(withSpec(manualv1alpha1.VPCEndpointServiceAllowedPrincipalParameters{
					Region:               region,
					VPCEndpointServiceID: serviceID,
					PrincipalARN:         principalARN,
				})),
				err: errorutils.Wrap(errBoom, errModifyPermissions),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
