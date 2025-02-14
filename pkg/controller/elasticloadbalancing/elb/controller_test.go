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

package elb

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/elasticloadbalancing/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticloadbalancing/elb"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticloadbalancing/elb/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	elbName                 = "some-elb"
	protocol                = "HTTP"
	port80            int32 = 80
	availabilityZones       = []string{"us-east-2a"}
	securityGroups          = []string{"sg-someid"}
	subnets                 = []string{"subnet1"}
	listener                = awselbtypes.Listener{
		InstancePort:     &port80,
		InstanceProtocol: &protocol,
		LoadBalancerPort: port80,
		Protocol:         &protocol,
	}

	errBoom = errors.New("boom")

	loadBalancer = awselbtypes.LoadBalancerDescription{
		AvailabilityZones: availabilityZones,
	}
)

type args struct {
	kube client.Client
	elb  elb.Client
	cr   resource.Managed
}

type elbModifier func(*v1alpha1.ELB)

func withConditions(c ...xpv1.Condition) elbModifier {
	return func(r *v1alpha1.ELB) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.ELBParameters) elbModifier {
	return func(r *v1alpha1.ELB) { r.Spec.ForProvider = p }
}

func withExternalName(name string) elbModifier {
	return func(r *v1alpha1.ELB) { meta.SetExternalName(r, name) }
}

func elbResource(m ...elbModifier) *v1alpha1.ELB {
	cr := &v1alpha1.ELB{}
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
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withSpec(v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
				}),
					withExternalName(elbName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleELB": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer, loadBalancer},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr:  elbResource(withExternalName(elbName)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeELBError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr:  elbResource(withExternalName(elbName)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
		"KubeClientError": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
				err: errorutils.Wrap(errBoom, errSpecUpdate),
			},
		},
		"NotUptoDate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroupIDs: securityGroups,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
						SecurityGroupIDs:  securityGroups,
					}),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb, kube: tc.kube}
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
					MockCreateLoadBalancer: func(ctx context.Context, input *awselb.CreateLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.CreateLoadBalancerOutput, error) {
						return &awselb.CreateLoadBalancerOutput{}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					}),
					withConditions(xpv1.Creating())),
			},
		},
		"CreateError": {
			args: args{
				elb: &fake.MockClient{
					MockCreateLoadBalancer: func(ctx context.Context, input *awselb.CreateLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.CreateLoadBalancerOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
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

func TestUpdate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpdateAZ": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{loadBalancer},
						}, nil
					},
					MockEnableAvailabilityZonesForLoadBalancer: func(ctx context.Context, input *awselb.EnableAvailabilityZonesForLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.EnableAvailabilityZonesForLoadBalancerOutput, error) {
						return &awselb.EnableAvailabilityZonesForLoadBalancerOutput{}, nil
					},
					MockDisableAvailabilityZonesForLoadBalancer: func(ctx context.Context, input *awselb.DisableAvailabilityZonesForLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.DisableAvailabilityZonesForLoadBalancerOutput, error) {
						return &awselb.DisableAvailabilityZonesForLoadBalancerOutput{}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: []string{"us-east-2b"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: []string{"us-east-2b"},
					})),
			},
		},
		"UpdateSubnet": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{
								{
									Subnets: subnets,
								},
							},
						}, nil
					},
					MockAttachLoadBalancerToSubnets: func(ctx context.Context, input *awselb.AttachLoadBalancerToSubnetsInput, opts []func(*awselb.Options)) (*awselb.AttachLoadBalancerToSubnetsOutput, error) {
						return &awselb.AttachLoadBalancerToSubnetsOutput{}, nil
					},
					MockDetachLoadBalancerFromSubnets: func(ctx context.Context, input *awselb.DetachLoadBalancerFromSubnetsInput, opts []func(*awselb.Options)) (*awselb.DetachLoadBalancerFromSubnetsOutput, error) {
						return &awselb.DetachLoadBalancerFromSubnetsOutput{}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SubnetIDs: []string{"subnet2"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SubnetIDs: []string{"subnet2"},
					})),
			},
		},
		"UpdateSG": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{
								{
									SecurityGroups: securityGroups,
								},
							},
						}, nil
					},
					MockApplySecurityGroupsToLoadBalancer: func(ctx context.Context, input *awselb.ApplySecurityGroupsToLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.ApplySecurityGroupsToLoadBalancerOutput, error) {
						return &awselb.ApplySecurityGroupsToLoadBalancerOutput{}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroupIDs: []string{"sg-other"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroupIDs: []string{"sg-other"},
					})),
			},
		},
		"UpdateListener": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancers: func(ctx context.Context, input *awselb.DescribeLoadBalancersInput, opts []func(*awselb.Options)) (*awselb.DescribeLoadBalancersOutput, error) {
						return &awselb.DescribeLoadBalancersOutput{
							LoadBalancerDescriptions: []awselbtypes.LoadBalancerDescription{
								{
									ListenerDescriptions: []awselbtypes.ListenerDescription{
										{
											Listener: &listener,
										},
									},
								},
							},
						}, nil
					},
					MockCreateLoadBalancerListeners: func(ctx context.Context, input *awselb.CreateLoadBalancerListenersInput, opts []func(*awselb.Options)) (*awselb.CreateLoadBalancerListenersOutput, error) {
						return &awselb.CreateLoadBalancerListenersOutput{}, nil
					},
					MockDeleteLoadBalancerListeners: func(ctx context.Context, input *awselb.DeleteLoadBalancerListenersInput, opts []func(*awselb.Options)) (*awselb.DeleteLoadBalancerListenersOutput, error) {
						return &awselb.DeleteLoadBalancerListenersOutput{}, nil
					},
					MockDescribeTags: func(ctx context.Context, input *awselb.DescribeTagsInput, opts []func(*awselb.Options)) (*awselb.DescribeTagsOutput, error) {
						return &awselb.DescribeTagsOutput{
							TagDescriptions: []awselbtypes.TagDescription{
								{LoadBalancerName: &elbName},
							},
						}, nil
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Listeners: []v1alpha1.Listener{
							{
								InstancePort:     8180,
								InstanceProtocol: &protocol,
								LoadBalancerPort: 8180,
								Protocol:         protocol,
							},
						},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Listeners: []v1alpha1.Listener{
							{
								InstancePort:     8180,
								InstanceProtocol: &protocol,
								LoadBalancerPort: 8180,
								Protocol:         protocol,
							},
						},
					})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
					MockDeleteLoadBalancer: func(ctx context.Context, input *awselb.DeleteLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.DeleteLoadBalancerOutput, error) {
						return &awselb.DeleteLoadBalancerOutput{}, nil
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withConditions(xpv1.Deleting())),
			},
		},
		"DeleteError": {
			args: args{
				elb: &fake.MockClient{
					MockDeleteLoadBalancer: func(ctx context.Context, input *awselb.DeleteLoadBalancerInput, opts []func(*awselb.Options)) (*awselb.DeleteLoadBalancerOutput, error) {
						return nil, errBoom
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
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
