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

package resourcetag

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	errBoom = errors.New("boom")

	testResourceID      = "test-resource"
	testOtherResourceID = "test-other-resource"
	testTagKey          = "key-1"
	testTagValue        = "value-1"
	testOtherTagKey     = "key-2"
	testOtherTagValue   = "value-2"
)

type args struct {
	ResourceTag ec2.ResourceTagClient
	kube        client.Client
	cr          *manualv1alpha1.ResourceTag
}

type ResourceTagModifier func(*manualv1alpha1.ResourceTag)

func withConditions(c ...xpv1.Condition) ResourceTagModifier {
	return func(r *manualv1alpha1.ResourceTag) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p manualv1alpha1.ResourceTagParameters) ResourceTagModifier {
	return func(r *manualv1alpha1.ResourceTag) { r.Spec.ForProvider = p }
}

func ResourceTag(m ...ResourceTagModifier) *manualv1alpha1.ResourceTag {
	cr := &manualv1alpha1.ResourceTag{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.ResourceTag
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDescribeTagsRequest: func(dti *awsec2.DescribeTagsInput) awsec2.DescribeTagsRequest {
						return awsec2.DescribeTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeTagsOutput{
								Tags: []awsec2.TagDescription{
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testOtherResourceID)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue), ResourceId: awsclient.String(testOtherResourceID)},
								},
							}},
						}
					},
				},
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
							{Key: testOtherTagKey, Value: testOtherTagValue},
						},
					}),
				),
			},
			want: want{
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
							{Key: testOtherTagKey, Value: testOtherTagValue},
						},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MissingTags": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDescribeTagsRequest: func(dti *awsec2.DescribeTagsInput) awsec2.DescribeTagsRequest {
						return awsec2.DescribeTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeTagsOutput{
								Tags: []awsec2.TagDescription{
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testOtherResourceID)},
								},
							}},
						}
					},
				},
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
							{Key: testOtherTagKey, Value: testOtherTagValue},
						},
					}),
				),
			},
			want: want{
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
							{Key: testOtherTagKey, Value: testOtherTagValue},
						},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"TooManyTags": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDescribeTagsRequest: func(dti *awsec2.DescribeTagsInput) awsec2.DescribeTagsRequest {
						return awsec2.DescribeTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeTagsOutput{
								Tags: []awsec2.TagDescription{
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue), ResourceId: awsclient.String(testResourceID)},
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue), ResourceId: awsclient.String(testOtherResourceID)},
								},
							}},
						}
					},
				},
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
						},
					}),
				),
			},
			want: want{
				cr: ResourceTag(
					withSpec(manualv1alpha1.ResourceTagParameters{
						ResourceIDs: []string{
							testResourceID,
							testOtherResourceID,
						},
						Tags: []manualv1alpha1.ResourceTagTag{
							{Key: testTagKey, Value: testTagValue},
						},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DescribeFailed": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDescribeTagsRequest: func(dti *awsec2.DescribeTagsInput) awsec2.DescribeTagsRequest {
						return awsec2.DescribeTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr:  ResourceTag(),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ResourceTag}
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
		cr     *manualv1alpha1.ResourceTag
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockCreateTagsRequest: func(cti *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr:     ResourceTag(),
				result: managed.ExternalCreation{},
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ResourceTag: &fake.MockResourceTagClient{
					MockCreateTagsRequest: func(cti *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr:  ResourceTag(),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ResourceTag}
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
		cr     *manualv1alpha1.ResourceTag
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockCreateTagsRequest: func(cti *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr: ResourceTag(),
			},
		},
		"ModifyFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ResourceTag: &fake.MockResourceTagClient{
					MockCreateTagsRequest: func(cti *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr:  ResourceTag(),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ResourceTag}
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
		cr  *manualv1alpha1.ResourceTag
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDeleteTagsRequest: func(dti *awsec2.DeleteTagsInput) awsec2.DeleteTagsRequest {
						return awsec2.DeleteTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteTagsOutput{}},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr: ResourceTag(),
			},
		},
		"DeleteFailed": {
			args: args{
				ResourceTag: &fake.MockResourceTagClient{
					MockDeleteTagsRequest: func(dti *awsec2.DeleteTagsInput) awsec2.DeleteTagsRequest {
						return awsec2.DeleteTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ResourceTag(),
			},
			want: want{
				cr:  ResourceTag(),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ResourceTag}
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
