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

package s3

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	storage "github.com/crossplane/provider-aws/apis/storage/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

const (
	namespace   = "default"
	bucketName  = "test-bucket"
	policy      = "some-policy"
	otherPolicy = "some otherPolicy"

	providerName    = "aws-creds"
	secretNamespace = "crossplane-system"
	testRegion      = "us-east-1"

	connectionSecretName = "my-little-secret"
	secretKey            = "credentials"
	credData             = "confidential!"
)

var (
	errBoom = errors.New("boom")
)

type args struct {
	s3   s3.Client
	kube client.Client
	cr   *storage.S3Bucket
}

func bucket(m ...func(b *storage.S3Bucket)) *storage.S3Bucket {
	cr := &storage.S3Bucket{
		Spec: storage.S3BucketSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestConnect(t *testing.T) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectionSecretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			secretKey: []byte(credData),
		},
	}

	providerSA := func(saVal bool) awsv1alpha3.Provider {
		return awsv1alpha3.Provider{
			Spec: awsv1alpha3.ProviderSpec{
				Region:            testRegion,
				UseServiceAccount: &saVal,
				ProviderSpec: runtimev1alpha1.ProviderSpec{
					CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
						SecretReference: runtimev1alpha1.SecretReference{
							Namespace: secretNamespace,
							Name:      connectionSecretName,
						},
						Key: secretKey,
					},
				},
			},
		}
	}
	type args struct {
		kube        client.Client
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (s3.Client, error)
		cr          *storage.S3Bucket
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i s3.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: bucket(),
			},
		},
		"SuccessfulUseServiceAccount": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						if key == (client.ObjectKey{Name: providerName}) {
							p := providerSA(true)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i s3.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: bucket(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: bucket(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProvider),
			},
		},
		"SecretGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: bucket(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProviderSecret),
			},
		},
		"SecretGetFailedNil": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.SetCredentialsSecretReference(nil)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: bucket(),
			},
			want: want{
				err: errors.New(errGetProviderSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newClientFn: tc.newClientFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *storage.S3Bucket
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
					MockGet: test.NewMockGetFn(errBoom),
				},
				s3: &fake.MockS3Client{
					MockHeadBucket: func(input *awsS3.HeadBucketInput) awsS3.HeadBucketRequest {
						return awsS3.HeadBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.HeadBucketOutput{}},
						}
					},
					MockGetBucketPolicy: func(input *awsS3.GetBucketPolicyInput) awsS3.GetBucketPolicyRequest {
						return awsS3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.GetBucketPolicyOutput{}},
						}
					},
				},
				cr: bucket(),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					b.SetConditions(runtimev1alpha1.Available())
				}),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockS3Client{
					MockHeadBucket: func(input *awsS3.HeadBucketInput) awsS3.HeadBucketRequest {
						return awsS3.HeadBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: bucket(),
			},
			want: want{
				cr: bucket(),
				//),
				err: errors.Wrap(errBoom, errGetBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.s3}
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
		cr     *storage.S3Bucket
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				s3: &fake.MockS3Client{
					MockCreateBucket: func(input *awsS3.CreateBucketInput) awsS3.CreateBucketRequest {
						return awsS3.CreateBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.CreateBucketOutput{}},
						}
					},
					MockPutBucketPolicy: func(input *awsS3.PutBucketPolicyInput) awsS3.PutBucketPolicyRequest {
						return awsS3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.PutBucketPolicyOutput{}},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.SetConditions(runtimev1alpha1.Creating())
				}),
			},
		},
		"CreateError": {
			args: args{
				s3: &fake.MockS3Client{
					MockCreateBucket: func(input *awsS3.CreateBucketInput) awsS3.CreateBucketRequest {
						return awsS3.CreateBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.SetConditions(runtimev1alpha1.Creating())
				}),
				err: errors.Wrap(errBoom, errCreateBucket),
			},
		},
		"PolicyAttachError": {
			args: args{
				s3: &fake.MockS3Client{
					MockCreateBucket: func(input *awsS3.CreateBucketInput) awsS3.CreateBucketRequest {
						return awsS3.CreateBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.CreateBucketOutput{}},
						}
					},
					MockPutBucketPolicy: func(input *awsS3.PutBucketPolicyInput) awsS3.PutBucketPolicyRequest {
						return awsS3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(policy)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(policy)
					b.SetConditions(runtimev1alpha1.Creating())
				}),
				err: errors.Wrap(errBoom, errPolicyAttach),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.s3}
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
		cr     *storage.S3Bucket
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				s3: &fake.MockS3Client{
					MockGetBucketPolicy: func(input *awsS3.GetBucketPolicyInput) awsS3.GetBucketPolicyRequest {
						return awsS3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.GetBucketPolicyOutput{
								Policy: aws.String(policy),
							}},
						}
					},
					MockPutBucketPolicy: func(input *awsS3.PutBucketPolicyInput) awsS3.PutBucketPolicyRequest {
						return awsS3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.PutBucketPolicyOutput{}},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(otherPolicy)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(otherPolicy)
				}),
			},
		},
		"PutPolicyFail": {
			args: args{
				s3: &fake.MockS3Client{
					MockGetBucketPolicy: func(input *awsS3.GetBucketPolicyInput) awsS3.GetBucketPolicyRequest {
						return awsS3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.GetBucketPolicyOutput{
								Policy: aws.String(policy),
							}},
						}
					},
					MockPutBucketPolicy: func(input *awsS3.PutBucketPolicyInput) awsS3.PutBucketPolicyRequest {
						return awsS3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(otherPolicy)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.Spec.ForProvider.Policy = aws.String(otherPolicy)
				}),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.s3}
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
		cr  *storage.S3Bucket
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				s3: &fake.MockS3Client{
					MockDeleteBucket: func(input *awsS3.DeleteBucketInput) awsS3.DeleteBucketRequest {
						return awsS3.DeleteBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsS3.DeleteBucketOutput{}},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.SetConditions(runtimev1alpha1.Deleting())
				}),
			},
		},
		"DeleteFail": {
			args: args{
				s3: &fake.MockS3Client{
					MockDeleteBucket: func(input *awsS3.DeleteBucketInput) awsS3.DeleteBucketRequest {
						return awsS3.DeleteBucketRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
				}),
			},
			want: want{
				cr: bucket(func(b *storage.S3Bucket) {
					meta.SetExternalName(b, bucketName)
					b.SetConditions(runtimev1alpha1.Deleting())
				}),
				err: errors.Wrap(errBoom, errDeleteBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.s3}
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
