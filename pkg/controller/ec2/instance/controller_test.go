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

package instance

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
)

var (
	instanceID = "some Id"

	errBoom = errors.New("boom")
)

type args struct {
	instance ec2.InstanceClient
	kube     client.Client
	cr       *manualv1alpha1.Instance
}

type instanceModifier func(*manualv1alpha1.Instance)

func withExternalName(name string) instanceModifier {
	return func(r *manualv1alpha1.Instance) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) instanceModifier {
	return func(r *manualv1alpha1.Instance) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p manualv1alpha1.InstanceParameters) instanceModifier {
	return func(r *manualv1alpha1.Instance) { r.Spec.ForProvider = p }
}

func withStatus(s manualv1alpha1.InstanceObservation) instanceModifier {
	return func(r *manualv1alpha1.Instance) { r.Status.AtProvider = s }
}

func withTags(tagMaps ...map[string]string) instanceModifier {
	var tagList []manualv1alpha1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, manualv1alpha1.Tag{Key: k, Value: v})
		}
	}
	return func(r *manualv1alpha1.Instance) { r.Spec.ForProvider.Tags = tagList }
}

func instance(m ...instanceModifier) *manualv1alpha1.Instance {
	cr := &manualv1alpha1.Instance{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.Instance
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
				instance: &fake.MockInstanceClient{
					MockDescribeInstances: func(ctx context.Context, input *awsec2.DescribeInstancesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
						return &awsec2.DescribeInstancesOutput{
							Reservations: []types.Reservation{{
								Instances: []types.Instance{
									{
										InstanceId:   &instanceID,
										InstanceType: types.InstanceTypeM1Small,
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							}},
						}, nil
					},
					MockDescribeInstanceAttribute: func(ctx context.Context, input *awsec2.DescribeInstanceAttributeInput, opts []func(*awsec2.Options)) (*awsec2.DescribeInstanceAttributeOutput, error) {
						return &awsec2.DescribeInstanceAttributeOutput{
							InstanceId: &instanceID,
							InstanceType: &types.AttributeValue{
								Value: aws.String(string(types.InstanceTypeM1Small)),
							},
						}, nil
					},
				},
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withExternalName(instanceID)),
			},
			want: want{
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withStatus(manualv1alpha1.InstanceObservation{
					InstanceID:   &instanceID,
					InstanceType: string(types.InstanceTypeM1Small),
					State:        "running",
				}), withExternalName(instanceID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleInstances": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				instance: &fake.MockInstanceClient{
					MockDescribeInstances: func(ctx context.Context, input *awsec2.DescribeInstancesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
						return &awsec2.DescribeInstancesOutput{
							Reservations: []types.Reservation{{
								Instances: []types.Instance{
									{
										InstanceId:   &instanceID,
										InstanceType: types.InstanceTypeM1Small,
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
									{
										InstanceId:   &instanceID,
										InstanceType: types.InstanceTypeM1Small,
										State: &types.InstanceState{
											Name: types.InstanceStateNameRunning,
										},
									},
								},
							}},
						}, nil
					},
				},
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withExternalName(instanceID)),
			},
			want: want{
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withExternalName(instanceID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				instance: &fake.MockInstanceClient{
					MockDescribeInstances: func(ctx context.Context, input *awsec2.DescribeInstancesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
						return &awsec2.DescribeInstancesOutput{}, errBoom
					},
				},
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withExternalName(instanceID)),
			},
			want: want{
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{
					InstanceType: string(types.InstanceTypeM1Small),
				}), withExternalName(instanceID)),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.instance}
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
		cr     *manualv1alpha1.Instance
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				instance: &fake.MockInstanceClient{
					MockRunInstances: func(ctx context.Context, input *awsec2.RunInstancesInput, opts []func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
						return &awsec2.RunInstancesOutput{
							Instances: []types.Instance{
								{
									InstanceId: &instanceID,
								},
							},
						}, nil
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr:     instance(withExternalName(instanceID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"CreateFail": {
			args: args{
				instance: &fake.MockInstanceClient{
					MockRunInstances: func(ctx context.Context, input *awsec2.RunInstancesInput, opts []func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
						return &awsec2.RunInstancesOutput{}, errBoom
					},
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.instance}
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
		cr  *manualv1alpha1.Instance
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				instance: &fake.MockInstanceClient{
					MockTerminateInstances: func(ctx context.Context, input *awsec2.TerminateInstancesInput, opts []func(*awsec2.Options)) (*awsec2.TerminateInstancesOutput, error) {
						return &awsec2.TerminateInstancesOutput{}, nil
					},
					MockDescribeInstances: func(ctx context.Context, input *awsec2.DescribeInstancesInput, opts []func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
						return &awsec2.DescribeInstancesOutput{
							Reservations: []types.Reservation{
								{
									Instances: []types.Instance{
										{
											InstanceId: aws.String(instanceID),
										},
									},
								},
							},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.args.kube, client: tc.instance}
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

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.Instance
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				instance: &fake.MockInstanceClient{
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, nil
					},
					MockModifyInstanceAttribute: func(ctx context.Context, input *awsec2.ModifyInstanceAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifyInstanceAttributeOutput, error) {
						return &awsec2.ModifyInstanceAttributeOutput{}, nil
					},
				},
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{})),
			},
			want: want{
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{})),
			},
		},
		"ModifyFailed": {
			args: args{
				instance: &fake.MockInstanceClient{
					MockCreateTags: func(ctx context.Context, input *awsec2.CreateTagsInput, opts []func(*awsec2.Options)) (*awsec2.CreateTagsOutput, error) {
						return &awsec2.CreateTagsOutput{}, errBoom
					},
					MockModifyInstanceAttribute: func(ctx context.Context, input *awsec2.ModifyInstanceAttributeInput, opts []func(*awsec2.Options)) (*awsec2.ModifyInstanceAttributeOutput, error) {
						return &awsec2.ModifyInstanceAttributeOutput{}, nil
					},
				},
				cr: instance(withSpec(manualv1alpha1.InstanceParameters{})),
			},
			want: want{
				cr:  instance(withSpec(manualv1alpha1.InstanceParameters{})),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.instance}
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

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *manualv1alpha1.Instance
		kube client.Client
	}
	type want struct {
		cr  *manualv1alpha1.Instance
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   instance(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: instance(withTags(resource.GetExternalTags(instance()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   instance(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tagger{kube: tc.kube}
			err := e.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.SortSlices(func(a, b manualv1alpha1.Tag) bool { return a.Key < b.Key })); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
