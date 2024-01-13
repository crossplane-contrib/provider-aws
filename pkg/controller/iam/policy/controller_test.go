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

package policy

import (
	"context"
	"net/url"
	"testing"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/middleware"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	unexpectedItem resource.Managed
	policyArn      = "arn:aws:iam::123456789012:policy/policy-name"
	name           = "policy-name"
	document       = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			  "Sid": "VisualEditor0",
			  "Effect": "Allow",
			  "Action": "elastic-inference:Connect",
			  "Resource": "*"
		  }
		]
	  }`
	documentURLEscaped = url.QueryEscape(document)
	boolFalse          = false

	errBoom = errors.New("boom")

	getCallerIdentityOutput = &sts.GetCallerIdentityOutput{
		Account:        pointer.ToOrNilIfZeroValue("123456789012"),
		Arn:            pointer.ToOrNilIfZeroValue("arn:aws:iam::123456789012:user/DevAdmin"),
		UserId:         pointer.ToOrNilIfZeroValue("AIDASAMPLEUSERID"),
		ResultMetadata: middleware.Metadata{},
	}

	tagComparer = cmp.Comparer(func(expected, actual awsiamtypes.Tag) bool {
		return cmp.Equal(expected.Key, actual.Key) &&
			cmp.Equal(expected.Value, actual.Value)
	})

	createInputComparer = cmp.Comparer(func(expected, actual *awsiam.CreatePolicyInput) bool {
		return cmp.Equal(expected.PolicyName, actual.PolicyName) &&
			cmp.Equal(expected.Path, actual.Path) &&
			cmp.Equal(expected.Description, actual.Description) &&
			cmp.Equal(expected.PolicyDocument, actual.PolicyDocument) &&
			cmp.Equal(expected.Tags, actual.Tags, tagComparer, sortIAMTags)
	})

	tagInputComparer = cmp.Comparer(func(expected, actual *awsiam.TagPolicyInput) bool {
		return cmp.Equal(expected.PolicyArn, actual.PolicyArn) &&
			cmp.Equal(expected.Tags, actual.Tags, tagComparer, sortIAMTags)
	})

	untagInputComparer = cmp.Comparer(func(expected, actual *awsiam.UntagPolicyInput) bool {
		return cmp.Equal(expected.PolicyArn, actual.PolicyArn) &&
			cmp.Equal(expected.TagKeys, actual.TagKeys, sortStrings)
	})

	sortIAMTags = cmpopts.SortSlices(func(a, b awsiamtypes.Tag) bool {
		return *a.Key > *b.Key
	})

	sortTags = cmpopts.SortSlices(func(a, b v1beta1.Tag) bool {
		return a.Key > b.Key
	})

	sortStrings = cmpopts.SortSlices(func(x, y string) bool {
		return x < y
	})
)

type args struct {
	kube client.Client
	iam  *fake.MockPolicyClient
	sts  iam.STSClient
	cr   resource.Managed
}

type policyModifier func(*v1beta1.Policy)

func withExternalName(s string) policyModifier {
	return func(r *v1beta1.Policy) { meta.SetExternalName(r, s) }
}

func withConditions(c ...xpv1.Condition) policyModifier {
	return func(r *v1beta1.Policy) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(spec v1beta1.PolicyParameters) policyModifier {
	return func(r *v1beta1.Policy) {
		r.Spec.ForProvider = spec
	}
}

func withPath(path string) policyModifier {
	return func(r *v1beta1.Policy) {
		r.Spec.ForProvider.Path = pointer.ToOrNilIfZeroValue(path)
	}
}

func withTags(tagMaps ...map[string]string) policyModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.Policy) {
		r.Spec.ForProvider.Tags = tagList
	}
}

func policy(m ...policyModifier) *v1beta1.Policy {
	cr := &v1beta1.Policy{}
	cr.Spec.ForProvider.Name = name
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
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
				}), withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
				}), withExternalName(policyArn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessfulURLEscapedPolicy": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &documentURLEscaped,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
				}), withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
				}), withExternalName(policyArn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"GetUPolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr:  policy(withExternalName(policyArn)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"EmptySpecPolicy": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withExternalName(policyArn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"EmptyExternalNameAndEntityDoesNotExist": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil, func(obj client.Object) error { return nil })},
				sts: &fake.MockSTSClient{MockGetCallerIdentity: func(ctx context.Context, input *sts.GetCallerIdentityInput, opts []func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
					return getCallerIdentityOutput, nil
				},
				},
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, &awsiamtypes.NoSuchEntityException{}
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withExternalName("")),
			},
			want: want{
				cr: policy(withExternalName(policyArn)),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"EmptyExternalNameAndEntityDoesNotExistForPolicyWithPath": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil, func(obj client.Object) error { return nil })},
				sts: &fake.MockSTSClient{MockGetCallerIdentity: func(ctx context.Context, input *sts.GetCallerIdentityInput, opts []func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
					return getCallerIdentityOutput, nil
				},
				},
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, &awsiamtypes.NoSuchEntityException{}
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withExternalName(""), withPath("/org-unit/")),
			},
			want: want{
				cr: policy(withExternalName("arn:aws:iam::123456789012:policy/org-unit/policy-name"), withPath("/org-unit/")),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"EmptyExternalNameAndEntityAlreadyExists": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil, func(obj client.Object) error { return nil })},
				sts: &fake.MockSTSClient{MockGetCallerIdentity: func(ctx context.Context, input *sts.GetCallerIdentityInput, opts []func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
					return getCallerIdentityOutput, nil
				},
				},
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withExternalName("")),
			},
			want: want{
				cr: policy(withExternalName(policyArn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"DifferentTags": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
								},
							},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
					Tags: []v1beta1.Tag{
						{Key: "key2", Value: "value2"},
					},
				}), withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Document: document,
					Name:     name,
					Tags: []v1beta1.Tag{
						{Key: "key2", Value: "value2"},
					},
				}), withExternalName(policyArn),
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
			e := &external{client: tc.iam, sts: tc.sts, kube: tc.kube}
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
		input  *awsiam.CreatePolicyInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				iam: &fake.MockPolicyClient{
					MockCreatePolicy: func(ctx context.Context, input *awsiam.CreatePolicyInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyOutput, error) {
						return &awsiam.CreatePolicyOutput{
							Policy: &awsiamtypes.Policy{
								Arn: &policyArn,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1beta1.PolicyParameters{
					Name:        name,
					Document:    document,
					Description: aws.String("description"),
					Path:        aws.String("path"),
					Tags: []v1beta1.Tag{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				})),
			},
			want: want{
				cr: policy(
					withSpec(v1beta1.PolicyParameters{
						Name:        name,
						Document:    document,
						Description: aws.String("description"),
						Path:        aws.String("path"),
						Tags: []v1beta1.Tag{
							{Key: "key2", Value: "value2"},
							{Key: "key1", Value: "value1"},
						},
					}),
					withExternalName(policyArn)),
				result: managed.ExternalCreation{},
				input: &awsiam.CreatePolicyInput{
					PolicyName:     aws.String(name),
					Description:    aws.String("description"),
					Path:           aws.String("path"),
					PolicyDocument: aws.String(document),
					Tags: []awsiamtypes.Tag{
						{Key: aws.String("key1"), Value: aws.String("value1")},
						{Key: aws.String("key2"), Value: aws.String("value2")},
					},
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockCreatePolicy: func(ctx context.Context, input *awsiam.CreatePolicyInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(),
			},
			want: want{
				cr:  policy(),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: test.NewMockClient(), client: tc.iam}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), sortTags); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.want.input != nil {
				actual := tc.args.iam.MockPolicyInput.CreatePolicyInput
				if diff := cmp.Diff(tc.want.input, actual, createInputComparer); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
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
		"Successful": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockCreatePolicyVersion: func(ctx context.Context, input *awsiam.CreatePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyVersionOutput, error) {
						return &awsiam.CreatePolicyVersionOutput{}, nil
					},
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withExternalName(policyArn)),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ListVersionsError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr:  policy(withExternalName(policyArn)),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
		"CreateVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockCreatePolicyVersion: func(ctx context.Context, input *awsiam.CreatePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyVersionOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr:  policy(withExternalName(policyArn)),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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

func TestUpdate_Tags(t *testing.T) {

	type want struct {
		cr         resource.Managed
		result     managed.ExternalUpdate
		err        error
		tagInput   *awsiam.TagPolicyInput
		untagInput *awsiam.UntagPolicyInput
	}

	listPolicyVersions := func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
		return &awsiam.ListPolicyVersionsOutput{}, nil
	}

	deletePolicyVersions := func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
		return &awsiam.DeletePolicyVersionOutput{}, nil
	}

	createPolicyVersion := func(ctx context.Context, input *awsiam.CreatePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyVersionOutput, error) {
		return &awsiam.CreatePolicyVersionOutput{}, nil
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddTagsError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockTagPolicy: func(ctx context.Context, input *awsiam.TagPolicyInput, opts []func(*awsiam.Options)) (*awsiam.TagPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key": "value",
					})),
			},
			want: want{
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key": "value",
					})),
				err: errorutils.Wrap(errBoom, errTag),
			},
		},
		"AddTagsSuccess": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockTagPolicy: func(ctx context.Context, input *awsiam.TagPolicyInput, opts []func(*awsiam.Options)) (*awsiam.TagPolicyOutput, error) {
						return nil, nil
					},
				},
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key": "value",
					})),
			},
			want: want{
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key": "value",
					})),
				tagInput: &awsiam.TagPolicyInput{
					PolicyArn: aws.String(policyArn),
					Tags: []awsiamtypes.Tag{
						{Key: aws.String("key"), Value: aws.String("value")},
					},
				},
			},
		},
		"UpdateTagsSuccess": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockTagPolicy: func(ctx context.Context, input *awsiam.TagPolicyInput, opts []func(*awsiam.Options)) (*awsiam.TagPolicyOutput, error) {
						return nil, nil
					},
				},
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key1": "value1",
						"key2": "value1",
					})),
			},
			want: want{
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key2": "value1",
						"key1": "value1",
					})),
				tagInput: &awsiam.TagPolicyInput{
					PolicyArn: aws.String(policyArn),
					Tags: []awsiamtypes.Tag{
						{Key: aws.String("key2"), Value: aws.String("value1")},
					},
				},
			},
		},
		"RemoveTagsError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockUntagPolicy: func(ctx context.Context, input *awsiam.UntagPolicyInput, opts []func(*awsiam.Options)) (*awsiam.UntagPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key2": "value2",
					})),
			},
			want: want{
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key2": "value2",
					})),
				err: errorutils.Wrap(errBoom, errUntag),
			},
		},
		"RemoveTagsSuccess": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockUntagPolicy: func(ctx context.Context, input *awsiam.UntagPolicyInput, opts []func(*awsiam.Options)) (*awsiam.UntagPolicyOutput, error) {
						return nil, nil
					},
				},
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key2": "value2",
					})),
			},
			want: want{
				cr: policy(
					withExternalName(policyArn),
					withTags(map[string]string{
						"key2": "value2",
					})),
				untagInput: &awsiam.UntagPolicyInput{
					PolicyArn: aws.String(policyArn),
					TagKeys:   []string{"key1"},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.iam.MockListPolicyVersions = listPolicyVersions
			tc.iam.MockDeletePolicyVersion = deletePolicyVersions
			tc.iam.MockCreatePolicyVersion = createPolicyVersion

			e := &external{client: tc.iam}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), sortTags); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.want.tagInput != nil {
				if diff := cmp.Diff(tc.want.tagInput, tc.iam.MockPolicyInput.TagPolicyInput, tagInputComparer); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
			}
			if tc.want.untagInput != nil {
				if diff := cmp.Diff(tc.want.untagInput, tc.iam.MockPolicyInput.UntagPolicyInput, untagInputComparer); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
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
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockDeletePolicy: func(ctx context.Context, input *awsiam.DeletePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyOutput, error) {
						return &awsiam.DeletePolicyOutput{}, nil
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withExternalName(policyArn),
					withConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"DeleteVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{
							Versions: []awsiamtypes.PolicyVersion{
								{
									IsDefaultVersion: boolFalse,
								},
							},
						}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr:  policy(withExternalName(policyArn)),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"DeletePolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockDeletePolicy: func(ctx context.Context, input *awsiam.DeletePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExternalName(policyArn)),
			},
			want: want{
				cr: policy(withExternalName(policyArn),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
