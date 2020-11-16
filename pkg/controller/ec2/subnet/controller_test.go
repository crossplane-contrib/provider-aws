/*
Copyright 2019 The Crossplane Authors.

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

package subnet

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	subnetID = "some Id"

	errBoom = errors.New("boom")
)

type args struct {
	subnet ec2.SubnetClient
	kube   client.Client
	cr     *v1beta1.Subnet
}

type subnetModifier func(*v1beta1.Subnet)

func withExternalName(name string) subnetModifier {
	return func(r *v1beta1.Subnet) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.SubnetParameters) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.SubnetObservation) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.AtProvider = s }
}

func subnet(m ...subnetModifier) *v1beta1.Subnet {
	cr := &v1beta1.Subnet{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Subnet
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{
									{
										State: awsec2.SubnetStateAvailable,
									},
								},
							}},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetState: "available",
				}), withExternalName(subnetID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleSubnets": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{}, {}},
							}},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errors.New(errMultipleItems),
			},
		},
		"NotUpToDate": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{
									{
										State:               awsec2.SubnetStateAvailable,
										MapPublicIpOnLaunch: aws.Bool(false),
									},
								},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}),
					withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetState: string(awsec2.SubnetStateAvailable),
				}), withExternalName(subnetID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"DescribeFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr     *v1beta1.Subnet
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockCreate: func(input *awsec2.CreateSubnetInput) awsec2.CreateSubnetRequest {
						return awsec2.CreateSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateSubnetOutput{
								Subnet: &awsec2.Subnet{
									SubnetId: aws.String(subnetID),
								},
							}},
						}
					},
				},
				cr: subnet(),
			},
			want: want{
				cr:     subnet(withExternalName(subnetID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				subnet: &fake.MockSubnetClient{
					MockCreate: func(input *awsec2.CreateSubnetInput) awsec2.CreateSubnetRequest {
						return awsec2.CreateSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(),
			},
			want: want{
				cr:  subnet(),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr     *v1beta1.Subnet
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(input *awsec2.ModifySubnetAttributeInput) awsec2.ModifySubnetAttributeRequest {
						return awsec2.ModifySubnetAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifySubnetAttributeOutput{}},
						}
					},
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{
									SubnetId:            aws.String(subnetID),
									MapPublicIpOnLaunch: aws.Bool(false),
								}},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(input *awsec2.ModifySubnetAttributeInput) awsec2.ModifySubnetAttributeRequest {
						return awsec2.ModifySubnetAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifySubnetAttributeOutput{}},
						}
					},
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{
									SubnetId:            aws.String(subnetID),
									MapPublicIpOnLaunch: aws.Bool(false),
								}},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1beta1.Subnet
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(input *awsec2.DeleteSubnetInput) awsec2.DeleteSubnetRequest {
						return awsec2.DeleteSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteSubnetOutput{}},
						}
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(input *awsec2.DeleteSubnetInput) awsec2.DeleteSubnetRequest {
						return awsec2.DeleteSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
