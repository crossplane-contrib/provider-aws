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

package repository

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	awsecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	ecr "github.com/crossplane/provider-aws/pkg/clients/ecr"
	"github.com/crossplane/provider-aws/pkg/clients/ecr/fake"
)

var (
	repoName            = "repoName"
	testARN             = "testARN"
	tagKey              = "test"
	tagValue            = "value"
	testECRTag          = awsecrtypes.Tag{Key: &tagKey, Value: &tagValue}
	testTag             = v1alpha1.Tag{Key: "test", Value: "value"}
	errBoom             = errors.New("boom")
	imageScanConfigTrue = v1alpha1.ImageScanningConfiguration{
		ScanOnPush: true,
	}
	imageScanConfigFalse = v1alpha1.ImageScanningConfiguration{
		ScanOnPush: false,
	}
	awsImageScanConfigFalse = awsecrtypes.ImageScanningConfiguration{
		ScanOnPush: imageScanConfigFalse.ScanOnPush,
	}
	lifecyclePolicyString           = `{"rules":[{"rulePriority":1,"description":"Expire images older than 14 days","selection":{"tagStatus":"untagged","countType":"sinceImagePushed","countUnit":"days","countNumber":14},"action":{"type":"expire"}}]}`
	multipleLifecyclePoliciesString = `{"rules":[{"rulePriority":1,"description":"Rule 1","selection":{"tagStatus":"tagged","tagPrefixList":["prod"],"countType":"imageCountMoreThan","countNumber":1},"action":{"type":"expire"}},{"rulePriority":2,"description":"Rule 2","selection":{"tagStatus":"tagged","tagPrefixList":["beta"],"countType":"imageCountMoreThan","countNumber":1},"action":{"type":"expire"}}]}`
	/* 	lifecyclePolicyStringMarshalled = &v1alpha1.LifecyclePolicy{
		Rules: []v1alpha1.LifecyclePolicyRule{
			{RulePriority: 1, Description: "Expire images older than 14 days", Selection: v1alpha1.LifecyclePolicySelection{TagStatus: "untagged", CountType: "sinceImagePushed", CountUnit: "days", CountNumber: 14}, Action: v1alpha1.LifecyclePolicyAction{Type: "expire"}},
		},
	} */
	multipleLifecyclePolicyStringMarshalled = &v1alpha1.LifecyclePolicy{
		Rules: []v1alpha1.LifecyclePolicyRule{
			{RulePriority: 2, Description: "Rule 2", Selection: v1alpha1.LifecyclePolicySelection{TagStatus: "tagged", TagPrefixList: []string{"beta"}, CountType: "imageCountMoreThan", CountNumber: 1}, Action: v1alpha1.LifecyclePolicyAction{Type: "expire"}},
			{RulePriority: 1, Description: "Rule 1", Selection: v1alpha1.LifecyclePolicySelection{TagStatus: "tagged", TagPrefixList: []string{"prod"}, CountType: "imageCountMoreThan", CountNumber: 1}, Action: v1alpha1.LifecyclePolicyAction{Type: "expire"}},
		},
	}
)

type args struct {
	repository ecr.RepositoryClient
	kube       client.Client
	cr         *v1alpha1.Repository
}

type repositoryModifier func(*v1alpha1.Repository)

func withTags(tagMaps ...map[string]string) repositoryModifier {
	var tagList []v1alpha1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1alpha1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1alpha1.Repository) { r.Spec.ForProvider.Tags = tagList }
}

func withExternalName(name string) repositoryModifier {
	return func(r *v1alpha1.Repository) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) repositoryModifier {
	return func(r *v1alpha1.Repository) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.RepositoryParameters) repositoryModifier {
	return func(r *v1alpha1.Repository) { r.Spec.ForProvider = p }
}

func withForceDelete(forceDelete bool) repositoryModifier {
	return func(r *v1alpha1.Repository) { r.Spec.ForProvider.ForceDelete = &forceDelete }
}

func withStatus(s v1alpha1.RepositoryObservation) repositoryModifier {
	return func(r *v1alpha1.Repository) { r.Status.AtProvider = s }
}

func repository(m ...repositoryModifier) *v1alpha1.Repository {
	cr := &v1alpha1.Repository{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Repository
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
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityMutable,
							}},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{
							Tags: []awsecrtypes.Tag{testECRTag},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
					Tags:               []v1alpha1.Tag{testTag},
				}), withStatus(v1alpha1.RepositoryObservation{
					RepositoryName: repoName,
					RepositoryArn:  testARN,
				}), withExternalName(repoName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleRepository": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							},
								{
									RepositoryArn:  &testARN,
									RepositoryName: &repoName,
								}},
						}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return nil, errBoom
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
		"ListTagsFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errListTags),
			},
		},
		"LifecyclePolicyIsUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityMutable,
							}},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{
							Tags: []awsecrtypes.Tag{testECRTag},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return &awsecr.GetLifecyclePolicyOutput{
							RepositoryName:      &repoName,
							LifecyclePolicyText: &multipleLifecyclePoliciesString,
						}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags:            []v1alpha1.Tag{testTag},
					LifecyclePolicy: multipleLifecyclePolicyStringMarshalled,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
					Tags:               []v1alpha1.Tag{testTag},
					LifecyclePolicy:    multipleLifecyclePolicyStringMarshalled,
				}), withStatus(v1alpha1.RepositoryObservation{
					RepositoryName: repoName,
					RepositoryArn:  testARN,
				}), withExternalName(repoName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"LifecyclePolicyIsMissing": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityMutable,
							}},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{
							Tags: []awsecrtypes.Tag{testECRTag},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return &awsecr.GetLifecyclePolicyOutput{
							RepositoryName:      &repoName,
							LifecyclePolicyText: &lifecyclePolicyString,
						}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags:            []v1alpha1.Tag{testTag},
					LifecyclePolicy: nil,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
					Tags:               []v1alpha1.Tag{testTag},
					LifecyclePolicy:    nil,
				}), withStatus(v1alpha1.RepositoryObservation{
					RepositoryName: repoName,
					RepositoryArn:  testARN,
				}), withExternalName(repoName),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"LifecyclePolicyIsNotUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				repository: &fake.MockRepositoryClient{
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityMutable,
							}},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{
							Tags: []awsecrtypes.Tag{testECRTag},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return &awsecr.GetLifecyclePolicyOutput{
							RepositoryName:      &repoName,
							LifecyclePolicyText: &lifecyclePolicyString,
						}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
					LifecyclePolicy: &v1alpha1.LifecyclePolicy{
						Rules: []v1alpha1.LifecyclePolicyRule{
							{RulePriority: 1, Description: "NOT MATCHING", Selection: v1alpha1.LifecyclePolicySelection{TagStatus: "untagged", CountType: "sinceImagePushed", CountUnit: "days", CountNumber: 14}, Action: v1alpha1.LifecyclePolicyAction{Type: "expire"}},
						},
					},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
					Tags:               []v1alpha1.Tag{testTag},
					LifecyclePolicy: &v1alpha1.LifecyclePolicy{
						Rules: []v1alpha1.LifecyclePolicyRule{
							{RulePriority: 1, Description: "NOT MATCHING", Selection: v1alpha1.LifecyclePolicySelection{TagStatus: "untagged", CountType: "sinceImagePushed", CountUnit: "days", CountNumber: 14}, Action: v1alpha1.LifecyclePolicyAction{Type: "expire"}},
						},
					},
				}), withStatus(v1alpha1.RepositoryObservation{
					RepositoryName: repoName,
					RepositoryArn:  testARN,
				}), withExternalName(repoName),
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
			e := &external{kube: tc.kube, client: tc.repository, subresourceClients: NewSubresourceClients(tc.repository)}
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
		cr     *v1alpha1.Repository
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				repository: &fake.MockRepositoryClient{
					MockCreate: func(ctx context.Context, input *awsecr.CreateRepositoryInput, opts []func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error) {
						return &awsecr.CreateRepositoryOutput{
							Repository: &awsecrtypes.Repository{
								RepositoryName: &repoName,
								RepositoryArn:  &testARN,
							},
						}, nil
					},
				},
				cr: repository(),
			},
			want: want{
				cr: repository(
					withConditions(xpv1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				repository: &fake.MockRepositoryClient{
					MockCreate: func(ctx context.Context, input *awsecr.CreateRepositoryInput, opts []func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error) {
						return nil, errBoom
					},
				},
				cr: repository(),
			},
			want: want{
				cr:  repository(withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository, subresourceClients: NewSubresourceClients(tc.repository)}
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
		cr     *v1alpha1.Repository
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAddTag": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockTag: func(ctx context.Context, input *awsecr.TagResourceInput, opts []func(*awsecr.Options)) (*awsecr.TagResourceOutput, error) {
						return &awsecr.TagResourceOutput{}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
			},
		},
		"SuccessfulRemoveTag": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockUntag: func(ctx context.Context, input *awsecr.UntagResourceInput, opts []func(*awsecr.Options)) (*awsecr.UntagResourceOutput, error) {
						return &awsecr.UntagResourceOutput{}, nil
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{
							Tags: []awsecrtypes.Tag{{Key: aws.String("something"), Value: aws.String("extra")}},
						}, nil
					},

					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
		},
		"ModifyTagFailed": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockTag: func(ctx context.Context, input *awsecr.TagResourceInput, opts []func(*awsecr.Options)) (*awsecr.TagResourceOutput, error) {
						return nil, errBoom
					},
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errCreateTags),
			},
		},
		"SuccessfulImageMutate": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityImmutable,
							}},
						}, nil
					},
					MockPutImageTagMutability: func(ctx context.Context, input *awsecr.PutImageTagMutabilityInput, opts []func(*awsecr.Options)) (*awsecr.PutImageTagMutabilityOutput, error) {
						return &awsecr.PutImageTagMutabilityOutput{}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				}), withExternalName(repoName)),
			},
		},
		"FailedImageMutate": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:      &testARN,
								RepositoryName:     &repoName,
								ImageTagMutability: awsecrtypes.ImageTagMutabilityImmutable,
							}},
						}, nil
					},
					MockPutImageTagMutability: func(ctx context.Context, input *awsecr.PutImageTagMutabilityInput, opts []func(*awsecr.Options)) (*awsecr.PutImageTagMutabilityOutput, error) {
						return nil, errBoom
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errUpdateMutability),
			},
		},
		"SuccessfulScanConfig": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:              &testARN,
								RepositoryName:             &repoName,
								ImageScanningConfiguration: &awsImageScanConfigFalse,
							}},
						}, nil
					},
					MockPutImageScan: func(ctx context.Context, input *awsecr.PutImageScanningConfigurationInput, opts []func(*awsecr.Options)) (*awsecr.PutImageScanningConfigurationOutput, error) {
						return &awsecr.PutImageScanningConfigurationOutput{}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				}), withExternalName(repoName)),
			},
		},
		"FailedScanConfig": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:              &testARN,
								RepositoryName:             &repoName,
								ImageScanningConfiguration: &awsImageScanConfigFalse,
							}},
						}, nil
					},
					MockPutImageScan: func(ctx context.Context, input *awsecr.PutImageScanningConfigurationInput, opts []func(*awsecr.Options)) (*awsecr.PutImageScanningConfigurationOutput, error) {
						return nil, errBoom
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errUpdateScan),
			},
		},
		"SuccessfulAddLifecyclePolicy": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
					MockPutLifecyclePolicy: func(ctx context.Context, input *awsecr.PutLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.PutLifecyclePolicyOutput, error) {
						return &awsecr.PutLifecyclePolicyOutput{LifecyclePolicyText: &multipleLifecyclePoliciesString, RepositoryName: &repoName}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					LifecyclePolicy: multipleLifecyclePolicyStringMarshalled,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					LifecyclePolicy: multipleLifecyclePolicyStringMarshalled,
				}), withExternalName(repoName)),
			},
		},
		"SuccessfulMutateLifecyclePolicy": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return &awsecr.GetLifecyclePolicyOutput{LifecyclePolicyText: &lifecyclePolicyString, RepositoryName: &repoName}, nil
					},
					MockPutLifecyclePolicy: func(ctx context.Context, input *awsecr.PutLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.PutLifecyclePolicyOutput, error) {
						return &awsecr.PutLifecyclePolicyOutput{LifecyclePolicyText: &multipleLifecyclePoliciesString, RepositoryName: &repoName}, nil
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					LifecyclePolicy: multipleLifecyclePolicyStringMarshalled,
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					LifecyclePolicy: multipleLifecyclePolicyStringMarshalled,
				}), withExternalName(repoName)),
			},
		},
		"FailedLifecyclePolicy": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(ctx context.Context, input *awsecr.ListTagsForResourceInput, opts []func(*awsecr.Options)) (*awsecr.ListTagsForResourceOutput, error) {
						return &awsecr.ListTagsForResourceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsecr.DescribeRepositoriesInput, opts []func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
						return &awsecr.DescribeRepositoriesOutput{
							Repositories: []awsecrtypes.Repository{{
								RepositoryArn:  &testARN,
								RepositoryName: &repoName,
							}},
						}, nil
					},
					MockGetLifecyclePolicy: func(ctx context.Context, input *awsecr.GetLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetLifecyclePolicyOutput, error) {
						return nil, &awsecrtypes.LifecyclePolicyNotFoundException{}
					},
					MockPutLifecyclePolicy: func(ctx context.Context, input *awsecr.PutLifecyclePolicyInput, opts []func(*awsecr.Options)) (*awsecr.PutLifecyclePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{LifecyclePolicy: multipleLifecyclePolicyStringMarshalled}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1alpha1.RepositoryParameters{LifecyclePolicy: multipleLifecyclePolicyStringMarshalled}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errCreateLifecyclePolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository, subresourceClients: NewSubresourceClients(tc.repository)}
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
		cr  *v1alpha1.Repository
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulForce": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryOutput, error) {
						var err error
						if !input.Force {
							err = errors.New("force must be set when forceDelete=true")
						}
						return &awsecr.DeleteRepositoryOutput{}, err
					},
				},
				cr: repository(withForceDelete(true)),
			},
			want: want{
				cr: repository(withForceDelete(true), withConditions(xpv1.Deleting())),
			},
		},
		"SuccessfulNoForce": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryOutput, error) {
						var err error
						if input.Force {
							err = errors.New("force must not be true when forceDelete is not set")
						}
						return &awsecr.DeleteRepositoryOutput{}, err
					},
				},
				cr: repository(withForceDelete(false)),
			},
			want: want{
				cr: repository(withForceDelete(false), withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryOutput, error) {
						return nil, errBoom
					},
				},
				cr: repository(),
			},
			want: want{
				cr:  repository(withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository, subresourceClients: NewSubresourceClients(tc.repository)}
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

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Repository
		kube client.Client
	}
	type want struct {
		cr  *v1alpha1.Repository
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   repository(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: repository(withTags(resource.GetExternalTags(repository()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   repository(),
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
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.SortSlices(func(a, b v1alpha1.Tag) bool { return a.Key > b.Key })); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
