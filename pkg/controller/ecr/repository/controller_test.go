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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	ecr "github.com/crossplane-contrib/provider-aws/pkg/clients/ecr"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ecr/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	repoName            = "repoName"
	testARN             = "testARN"
	tagKey              = "test"
	tagValue            = "value"
	testECRTag          = awsecrtypes.Tag{Key: &tagKey, Value: &tagValue}
	testTag             = v1beta1.Tag{Key: "test", Value: "value"}
	errBoom             = errors.New("boom")
	imageScanConfigTrue = v1beta1.ImageScanningConfiguration{
		ScanOnPush: true,
	}
	imageScanConfigFalse = v1beta1.ImageScanningConfiguration{
		ScanOnPush: false,
	}
	awsImageScanConfigFalse = awsecrtypes.ImageScanningConfiguration{
		ScanOnPush: imageScanConfigFalse.ScanOnPush,
	}
)

type args struct {
	repository ecr.RepositoryClient
	kube       client.Client
	cr         *v1beta1.Repository
}

type repositoryModifier func(*v1beta1.Repository)

func withExternalName(name string) repositoryModifier {
	return func(r *v1beta1.Repository) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) repositoryModifier {
	return func(r *v1beta1.Repository) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.RepositoryParameters) repositoryModifier {
	return func(r *v1beta1.Repository) { r.Spec.ForProvider = p }
}

func withForceDelete(forceDelete bool) repositoryModifier {
	return func(r *v1beta1.Repository) { r.Spec.ForProvider.ForceDelete = &forceDelete }
}

func withStatus(s v1beta1.RepositoryObservation) repositoryModifier {
	return func(r *v1beta1.Repository) { r.Status.AtProvider = s }
}

func repository(m ...repositoryModifier) *v1beta1.Repository {
	cr := &v1beta1.Repository{
		TypeMeta: metav1.TypeMeta{
			Kind: "Repository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "name",
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Repository
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{testTag},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
					Tags:               []v1beta1.Tag{testTag},
				}), withStatus(v1beta1.RepositoryObservation{
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
				cr: repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
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
				cr: repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
				err: errorutils.Wrap(errBoom, errDescribe),
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
				cr: repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1beta1.RepositoryParameters{}), withExternalName(repoName)),
				err: errorutils.Wrap(errBoom, errListTags),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository}
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
		cr     *v1beta1.Repository
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
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository}
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
		cr     *v1beta1.Repository
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{testTag},
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{testTag},
				})),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{})),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{testTag},
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{testTag},
				})),
				err: errorutils.Wrap(errBoom, errCreateTags),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				})),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecrtypes.ImageTagMutabilityMutable)),
				})),
				err: errorutils.Wrap(errBoom, errUpdateMutability),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
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
				},
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
			},
			want: want{
				cr: repository(withSpec(v1beta1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
				err: errorutils.Wrap(errBoom, errUpdateScan),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository}
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
		cr  *v1beta1.Repository
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
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.repository}
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
