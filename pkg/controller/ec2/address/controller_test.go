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

package address

import (
	"context"
	"testing"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	allocationID   = "some Id"
	domainVpc      = "vpc"
	domainStandard = "standard"
	publicIP       = "1.1.1.1"
	errBoom        = errors.New("boom")
)

type args struct {
	address ec2.AddressClient
	kube    client.Client
	cr      *v1beta1.Address
}

type addressModifier func(*v1beta1.Address)

func withExternalName(name string) addressModifier {
	return func(r *v1beta1.Address) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) addressModifier {
	return func(r *v1beta1.Address) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.AddressParameters) addressModifier {
	return func(r *v1beta1.Address) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.AddressObservation) addressModifier {
	return func(r *v1beta1.Address) { r.Status.AtProvider = s }
}

func address(m ...addressModifier) *v1beta1.Address {
	cr := &v1beta1.Address{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Address
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
				address: &fake.MockAddressClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeAddressesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeAddressesOutput, error) {
						return &awsec2.DescribeAddressesOutput{
							Addresses: []awsec2types.Address{{
								AllocationId: &allocationID,
							}},
						}, nil
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withStatus(v1beta1.AddressObservation{
					AllocationID: allocationID,
				}), withExternalName(allocationID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleAddresses": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				address: &fake.MockAddressClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeAddressesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeAddressesOutput, error) {
						return &awsec2.DescribeAddressesOutput{
							Addresses: []awsec2types.Address{{}, {}},
						}, nil
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withExternalName(allocationID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				address: &fake.MockAddressClient{
					MockDescribe: func(ctx context.Context, input *awsec2.DescribeAddressesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeAddressesOutput, error) {
						return nil, errBoom

					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				}), withExternalName(allocationID)),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.address}
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
		cr     *v1beta1.Address
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulVPC": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				address: &fake.MockAddressClient{
					MockAllocate: func(ctx context.Context, input *awsec2.AllocateAddressInput, opts []func(*awsec2.Options)) (*awsec2.AllocateAddressOutput, error) {
						return &awsec2.AllocateAddressOutput{
							AllocationId: &allocationID,
						}, nil
					},
				},
				cr: address(),
			},
			want: want{
				cr: address(withExternalName(allocationID),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulStandard": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				address: &fake.MockAddressClient{
					MockAllocate: func(ctx context.Context, input *awsec2.AllocateAddressInput, opts []func(*awsec2.Options)) (*awsec2.AllocateAddressOutput, error) {
						return &awsec2.AllocateAddressOutput{
							PublicIp: &publicIP,
						}, nil
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainStandard,
				})),
			},
			want: want{
				cr: address(withExternalName(publicIP),
					withConditions(xpv1.Creating()),
					withSpec(v1beta1.AddressParameters{
						Domain: &domainStandard,
					})),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				address: &fake.MockAddressClient{
					MockAllocate: func(ctx context.Context, input *awsec2.AllocateAddressInput, opts []func(*awsec2.Options)) (*awsec2.AllocateAddressOutput, error) {
						return nil, errBoom
					},
				},
				cr: address(),
			},
			want: want{
				cr:  address(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.address}
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
		cr     *v1beta1.Address
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				address: &fake.MockAddressClient{
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				})),
			},
			want: want{
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				address: &fake.MockAddressClient{
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return nil, errBoom
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				})),
			},
			want: want{
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainVpc,
				})),
				err: errorutils.Wrap(errBoom, errCreateTags),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.address}
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

func TestRelease(t *testing.T) {
	type want struct {
		cr  *v1beta1.Address
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulVPC": {
			args: args{
				address: &fake.MockAddressClient{
					MockRelease: func(ctx context.Context, input *awsec2.ReleaseAddressInput, opts []func(*awsec2.Options)) (*awsec2.ReleaseAddressOutput, error) {
						return &awsec2.ReleaseAddressOutput{}, nil
					},
				},
				cr: address(),
			},
			want: want{
				cr: address(withConditions(xpv1.Deleting())),
			},
		},
		"SuccessfulStandard": {
			args: args{
				address: &fake.MockAddressClient{
					MockRelease: func(ctx context.Context, input *awsec2.ReleaseAddressInput, opts []func(*awsec2.Options)) (*awsec2.ReleaseAddressOutput, error) {
						return &awsec2.ReleaseAddressOutput{}, nil
					},
				},
				cr: address(withSpec(v1beta1.AddressParameters{
					Domain: &domainStandard,
				})),
			},
			want: want{
				cr: address(withConditions(xpv1.Deleting()),
					withSpec(v1beta1.AddressParameters{
						Domain: &domainStandard,
					}),
				),
			},
		},
		"DeleteFailed": {
			args: args{
				address: &fake.MockAddressClient{
					MockRelease: func(ctx context.Context, input *awsec2.ReleaseAddressInput, opts []func(*awsec2.Options)) (*awsec2.ReleaseAddressOutput, error) {
						return nil, errBoom
					},
				},
				cr: address(),
			},
			want: want{
				cr:  address(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.address}
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
