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
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
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
	testECRTag          = awsecr.Tag{Key: &tagKey, Value: &tagValue}
	testTag             = v1alpha1.Tag{Key: "test", Value: "value"}
	errBoom             = errors.New("boom")
	imageScanConfigTrue = v1alpha1.ImageScanningConfiguration{
		ScanOnPush: true,
	}
	imageScanConfigFalse = v1alpha1.ImageScanningConfiguration{
		ScanOnPush: false,
	}
	awsImageScanConfigFalse = awsecr.ImageScanningConfiguration{
		ScanOnPush: &imageScanConfigFalse.ScanOnPush,
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
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:      &testARN,
									RepositoryName:     &repoName,
									ImageTagMutability: awsecr.ImageTagMutabilityMutable,
								}},
							}},
						}
					},
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{
								Tags: []awsecr.Tag{testECRTag},
							}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				}), withExternalName(repoName)),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecr.ImageTagMutabilityMutable)),
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
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:  &testARN,
									RepositoryName: &repoName,
								},
									{
										RepositoryArn:  &testARN,
										RepositoryName: &repoName,
									}},
							}},
						}
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
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:  &testARN,
									RepositoryName: &repoName,
								}},
							}},
						}
					},
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
			},
			want: want{
				cr:  repository(withSpec(v1alpha1.RepositoryParameters{}), withExternalName(repoName)),
				err: awsclient.Wrap(errBoom, errListTags),
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
					MockCreate: func(input *awsecr.CreateRepositoryInput) awsecr.CreateRepositoryRequest {
						return awsecr.CreateRepositoryRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.CreateRepositoryOutput{
								Repository: &awsecr.Repository{
									RepositoryName: &repoName,
									RepositoryArn:  &testARN,
								},
							}},
						}
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
					MockCreate: func(input *awsecr.CreateRepositoryInput) awsecr.CreateRepositoryRequest {
						return awsecr.CreateRepositoryRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
					MockTag: func(input *awsecr.TagResourceInput) awsecr.TagResourceRequest {
						return awsecr.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.TagResourceOutput{}},
						}
					},
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:  &testARN,
									RepositoryName: &repoName,
								}},
							}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				})),
			},
		},
		"SuccessfulRemoveTag": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockUntag: func(input *awsecr.UntagResourceInput) awsecr.UntagResourceRequest {
						return awsecr.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.UntagResourceOutput{}},
						}
					},
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{
								Tags: []awsecr.Tag{{Key: aws.String("something"), Value: aws.String("extra")}},
							}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:  &testARN,
									RepositoryName: &repoName,
								}},
							}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{})),
			},
		},
		"ModifyTagFailed": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockTag: func(input *awsecr.TagResourceInput) awsecr.TagResourceRequest {
						return awsecr.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{testTag},
				})),
				err: awsclient.Wrap(errBoom, errCreateTags),
			},
		},
		"SuccessfulImageMutate": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:      &testARN,
									RepositoryName:     &repoName,
									ImageTagMutability: awsecr.ImageTagMutabilityImmutable,
								}},
							}},
						}
					},
					MockPutImageTagMutability: func(input *awsecr.PutImageTagMutabilityInput) awsecr.PutImageTagMutabilityRequest {
						return awsecr.PutImageTagMutabilityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.PutImageTagMutabilityOutput{}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecr.ImageTagMutabilityMutable)),
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecr.ImageTagMutabilityMutable)),
				})),
			},
		},
		"FailedImageMutate": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:      &testARN,
									RepositoryName:     &repoName,
									ImageTagMutability: awsecr.ImageTagMutabilityImmutable,
								}},
							}},
						}
					},
					MockPutImageTagMutability: func(input *awsecr.PutImageTagMutabilityInput) awsecr.PutImageTagMutabilityRequest {
						return awsecr.PutImageTagMutabilityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecr.ImageTagMutabilityMutable)),
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageTagMutability: aws.String(string(awsecr.ImageTagMutabilityMutable)),
				})),
				err: awsclient.Wrap(errBoom, errUpdateMutability),
			},
		},
		"SuccessfulScanConfig": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:              &testARN,
									RepositoryName:             &repoName,
									ImageScanningConfiguration: &awsImageScanConfigFalse,
								}},
							}},
						}
					},
					MockPutImageScan: func(input *awsecr.PutImageScanningConfigurationInput) awsecr.PutImageScanningConfigurationRequest {
						return awsecr.PutImageScanningConfigurationRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.PutImageScanningConfigurationOutput{}},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
			},
		},
		"FailedScanConfig": {
			args: args{
				repository: &fake.MockRepositoryClient{
					MockListTags: func(input *awsecr.ListTagsForResourceInput) awsecr.ListTagsForResourceRequest {
						return awsecr.ListTagsForResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.ListTagsForResourceOutput{}},
						}
					},
					MockDescribe: func(input *awsecr.DescribeRepositoriesInput) awsecr.DescribeRepositoriesRequest {
						return awsecr.DescribeRepositoriesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DescribeRepositoriesOutput{
								Repositories: []awsecr.Repository{{
									RepositoryArn:              &testARN,
									RepositoryName:             &repoName,
									ImageScanningConfiguration: &awsImageScanConfigFalse,
								}},
							}},
						}
					},
					MockPutImageScan: func(input *awsecr.PutImageScanningConfigurationInput) awsecr.PutImageScanningConfigurationRequest {
						return awsecr.PutImageScanningConfigurationRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
			},
			want: want{
				cr: repository(withSpec(v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfigTrue,
				})),
				err: awsclient.Wrap(errBoom, errUpdateScan),
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
					MockDelete: func(input *awsecr.DeleteRepositoryInput) awsecr.DeleteRepositoryRequest {
						var err error
						if !aws.BoolValue(input.Force) {
							err = errors.New("force must be set when forceDelete=true")
						}
						return awsecr.DeleteRepositoryRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DeleteRepositoryOutput{}, Error: err},
						}
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
					MockDelete: func(input *awsecr.DeleteRepositoryInput) awsecr.DeleteRepositoryRequest {
						var err error
						if aws.BoolValue(input.Force) {
							err = errors.New("force must not be true when forceDelete is not set")
						}
						return awsecr.DeleteRepositoryRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsecr.DeleteRepositoryOutput{}, Error: err},
						}
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
					MockDelete: func(input *awsecr.DeleteRepositoryInput) awsecr.DeleteRepositoryRequest {
						return awsecr.DeleteRepositoryRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
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
