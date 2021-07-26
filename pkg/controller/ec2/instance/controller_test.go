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
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	instanceID = "some Id"
	// tenancyDefault = "default"

	// errBoom = errors.New("boom")
)

type args struct {
	instance ec2.InstanceClient
	kube     client.Client
	cr       *manualv1alpha1.Instance
}

type instanceModifier func(*manualv1alpha1.Instance)

// func withTags(tagMaps ...map[string]string) vpcModifier {
// 	var tagList []v1beta1.Tag
// 	for _, tagMap := range tagMaps {
// 		for k, v := range tagMap {
// 			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
// 		}
// 	}
// 	return func(r *v1beta1.VPC) { r.Spec.ForProvider.Tags = tagList }
// }

func withExternalName(name string) instanceModifier {
	return func(r *manualv1alpha1.Instance) { meta.SetExternalName(r, name) }
}

func withObjectName(name string) instanceModifier {
	return func(r *manualv1alpha1.Instance) { r.Name = name }
}

func withConditions(c ...xpv1.Condition) instanceModifier {
	return func(r *manualv1alpha1.Instance) { r.Status.ConditionedStatus.Conditions = c }
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
					MockRunInstancesRequest: func(input *awsec2.RunInstancesInput) awsec2.RunInstancesRequest {
						return awsec2.RunInstancesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.RunInstancesOutput{
								Instances: []awsec2.Instance{
									{
										InstanceId: &instanceID,
									},
								},
							}},
						}
					},
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: instance(withObjectName(instanceID)),
			},
			want: want{
				cr:     instance(withExternalName(instanceID), withObjectName(instanceID)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
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
					MockTerminateInstancesRequest: func(input *awsec2.TerminateInstancesInput) awsec2.TerminateInstancesRequest {
						return awsec2.TerminateInstancesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.TerminateInstancesOutput{}},
						}
					},
					MockDescribeInstancesRequest: func(input *awsec2.DescribeInstancesInput) awsec2.DescribeInstancesRequest {
						return awsec2.DescribeInstancesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInstancesOutput{
								Reservations: []awsec2.Reservation{
									{
										Instances: []awsec2.Instance{
											{
												InstanceId: aws.String(instanceID),
											},
										},
									},
								},
							}},
						}
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
