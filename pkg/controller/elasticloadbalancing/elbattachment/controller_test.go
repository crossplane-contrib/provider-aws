/*
Copyright 2020 The Crossplane Authors.

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

package elbattachment

import (
	"context"
	"testing"

	awselb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awselbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/elasticloadbalancing/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticloadbalancing/elb"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticloadbalancing/elb/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	elbName    = "some-elb"
	instanceID = "someID"

	errBoom = errors.New("boom")

	loadBalancer = awselbtypes.LoadBalancerDescription{
		Instances: []awselbtypes.Instance{
			{
				InstanceId: &instanceID,
			},
		},
	}
)

type args struct {
	elb elb.Client
	cr  resource.Managed
}

type elbAttachmentModifier func(*v1alpha1.ELBAttachment)

func withConditions(c ...xpv1.Condition) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.ELBAttachmentParameters) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { r.Spec.ForProvider = p }
}

func withExternalName(name string) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { meta.SetExternalName(r, name) }
}

func elbAttachmentResource(m ...elbAttachmentModifier) *v1alpha1.ELBAttachment {
	cr := &v1alpha1.ELBAttachment{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				}),
					withExternalName(elbName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NoAttachment": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
				},
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName: elbName,
				})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName: elbName,
				})),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
		"DescribeError": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				})),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				elb: &fake.MockClient{
					MockRegisterInstancesWithLoadBalancer: func(ctx context.Context, input *awselb.RegisterInstancesWithLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.RegisterInstancesWithLoadBalancerOutput, error) {
						return &awselb.RegisterInstancesWithLoadBalancerOutput{}, nil
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					}),
					withConditions(xpv1.Creating())),
			},
		},
		"CreateError": {
			args: args{
				elb: &fake.MockClient{
					MockRegisterInstancesWithLoadBalancer: func(ctx context.Context, input *awselb.RegisterInstancesWithLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.RegisterInstancesWithLoadBalancerOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					}),
					withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				elb: &fake.MockClient{
					MockDeregisterInstancesFromLoadBalancer: func(ctx context.Context, input *awselb.DeregisterInstancesFromLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.DeregisterInstancesFromLoadBalancerOutput, error) {
						return &awselb.DeregisterInstancesFromLoadBalancerOutput{}, nil
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withConditions(xpv1.Deleting())),
			},
		},
		"DeleteError": {
			args: args{
				elb: &fake.MockClient{
					MockDeregisterInstancesFromLoadBalancer: func(ctx context.Context, input *awselb.DeregisterInstancesFromLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.DeregisterInstancesFromLoadBalancerOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
