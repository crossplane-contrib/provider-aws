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
package secretsmanager

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssecretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/secretsmanager/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/secretsmanager"
	"github.com/crossplane/provider-aws/pkg/clients/secretsmanager/fake"
)

const (
	secretKey = "credentials"
	credData  = "confidential"
)

var (
	secretName      = "some-name"
	secretNamespace = "some-namespace"
	kmsKeyIDRef     = "kms-key-id"

	randomDate = time.Now()
	tags       = []v1alpha1.Tag{
		{
			Key:   "some-key",
			Value: "some-value",
		},
		{
			Key:   "some-other-key",
			Value: "some-other-value",
		},
	}
	awsTags = []awssecretsmanager.Tag{
		{
			Key:   aws.String("some-key"),
			Value: aws.String("some-value"),
		},
		{
			Key:   aws.String("some-other-key"),
			Value: aws.String("some-other-value"),
		},
	}

	errBoom = errors.New("boom")
)

type args struct {
	kube           client.Client
	secretsmanager secretsmanager.Client
	cr             *v1alpha1.Secret
}

func secret(m ...secretModifier) *v1alpha1.Secret {
	cr := &v1alpha1.Secret{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type secretModifier func(*v1alpha1.Secret)

func withExternalName(s string) secretModifier {
	return func(r *v1alpha1.Secret) { meta.SetExternalName(r, s) }
}

func withDeletedDate(s time.Time) secretModifier {
	return func(r *v1alpha1.Secret) { r.Status.AtProvider.DeletedDate = &metav1.Time{Time: s} }
}

func withDeletionDate(s time.Time) secretModifier {
	return func(r *v1alpha1.Secret) { r.Status.AtProvider.DeletionDate = &metav1.Time{Time: s} }
}

func withForceDeleteWithoutRecovery(b bool) secretModifier {
	return func(r *v1alpha1.Secret) { r.Spec.ForProvider.ForceDeleteWithoutRecovery = &b }
}

func withRecoveryWindow(i int64) secretModifier {
	return func(r *v1alpha1.Secret) { r.Spec.ForProvider.RecoveryWindowInDays = &i }
}

func withConditions(c ...runtimev1alpha1.Condition) secretModifier {
	return func(r *v1alpha1.Secret) { r.Status.ConditionedStatus.Conditions = c }
}

func withSecretRef(n, ns, key string) secretModifier {
	return func(r *v1alpha1.Secret) {
		r.Spec.ForProvider.SecretRef = &v1alpha1.SecretSelector{
			SecretReference: &runtimev1alpha1.SecretReference{
				Name:      n,
				Namespace: ns,
			},
			Key: key,
		}
	}
}

func withKmsKeyIDRef(kmsKeyIDRef string) secretModifier {
	return func(r *v1alpha1.Secret) {
		r.Spec.ForProvider.KmsKeyRef = &runtimev1alpha1.Reference{
			Name: kmsKeyIDRef,
		}
	}
}

func withTagList(tagMaps ...map[string]string) secretModifier {
	var tagList []v1alpha1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1alpha1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1alpha1.Secret) { r.Spec.ForProvider.Tags = tagList }
}

func withTags(p []v1alpha1.Tag) secretModifier {
	return func(r *v1alpha1.Secret) { r.Spec.ForProvider.Tags = p }
}

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Secret
		kube client.Client
	}
	type want struct {
		cr  *v1alpha1.Secret
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   secret(withTagList(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: secret(withTagList(resource.GetExternalTags(secret()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   secret(),
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

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Secret
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulObservation": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{},
								Data: &awssecretsmanager.GetSecretValueOutput{
									SecretString: awsclients.String(credData),
								},
							},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
					withSecretRef(secretName, secretNamespace, secretKey),
					withKmsKeyIDRef(kmsKeyIDRef),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
					withConditions(runtimev1alpha1.Available()),
					withSecretRef(secretName, secretNamespace, secretKey),
					withKmsKeyIDRef(kmsKeyIDRef),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"SuccessfulObservationWithTags": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								Tags: awsTags,
							}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{},
								Data: &awssecretsmanager.GetSecretValueOutput{
									SecretString: awsclients.String(credData),
								},
							},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
					withConditions(runtimev1alpha1.Available()),
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"SuccessfulObservationWithoutExternalName": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								Tags: awsTags,
							}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{},
								Data: &awssecretsmanager.GetSecretValueOutput{
									SecretString: awsclients.String(credData),
								},
							},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"SecretNotUpToDate": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{},
								Data: &awssecretsmanager.GetSecretValueOutput{
									SecretString: awsclients.String("some-outdated-secret-value"),
								},
							},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
					withConditions(runtimev1alpha1.Available()),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
			},
		},
		"FailedDescribeSecretRequest": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
				),
				err: errors.Wrap(errBoom, errDescribeSecretFailed),
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								Tags: awsTags,
							}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.GetSecretValueOutput{}},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
					withTags(tags)),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
		"FailedGetSecretValueRequest": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
				),
				err: errors.Wrap(errBoom, errGetSecretValueFailed),
			},
		},
		"NotFound": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awssecretsmanager.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: secret(),
			},
			want: want{
				cr:     secret(),
				result: managed.ExternalObservation{},
			},
		},
		"DeletedDateNotNil": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{}
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								DeletedDate: &randomDate,
								Tags:        awsTags,
							}},
						}
					},
					MockGetSecretValueRequest: func(input *awssecretsmanager.GetSecretValueInput) awssecretsmanager.GetSecretValueRequest {
						return awssecretsmanager.GetSecretValueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.GetSecretValueOutput{}},
						}
					},
				},
				cr: secret(
					withExternalName(secretName),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withExternalName(secretName),
					withConditions(runtimev1alpha1.Deleting().WithMessage(secretMarkedForDeletion)),
					withSecretRef(secretName, secretNamespace, secretKey),
					withDeletedDate(randomDate),
					withTags(tags),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.secretsmanager}
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
		cr     *v1alpha1.Secret
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreation": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockCreateSecretRequest: func(input *awssecretsmanager.CreateSecretInput) awssecretsmanager.CreateSecretRequest {
						return awssecretsmanager.CreateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.CreateSecretOutput{}},
						}
					},
				},
				cr: secret(withSecretRef(secretName, secretNamespace, secretKey)),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Creating()),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedGetSecret": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: secret(withSecretRef(secretName, secretNamespace, secretKey)),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Creating()),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errK8sSecretNotFound),
			},
		},
		"FailedCreateSecretRequest": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockCreateSecretRequest: func(input *awssecretsmanager.CreateSecretInput) awssecretsmanager.CreateSecretRequest {
						return awssecretsmanager.CreateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Creating()),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.secretsmanager}
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
		cr     *v1alpha1.Secret
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockUpdateSecretRequest: func(input *awssecretsmanager.UpdateSecretInput) awssecretsmanager.UpdateSecretRequest {
						return awssecretsmanager.UpdateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UpdateSecretOutput{}},
						}
					},
					MockTagResourceRequest: func(input *awssecretsmanager.TagResourceInput) awssecretsmanager.TagResourceRequest {
						return awssecretsmanager.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.TagResourceOutput{}},
						}
					},
					MockUntagResourceRequest: func(input *awssecretsmanager.UntagResourceInput) awssecretsmanager.UntagResourceRequest {
						return awssecretsmanager.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UntagResourceOutput{}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
			},
		},
		"SuccessfulUpdateRemoveTags": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								Tags: awsTags,
							}},
						}
					},
					MockUpdateSecretRequest: func(input *awssecretsmanager.UpdateSecretInput) awssecretsmanager.UpdateSecretRequest {
						return awssecretsmanager.UpdateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UpdateSecretOutput{}},
						}
					},
					MockTagResourceRequest: func(input *awssecretsmanager.TagResourceInput) awssecretsmanager.TagResourceRequest {
						return awssecretsmanager.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.TagResourceOutput{}},
						}
					},
					MockUntagResourceRequest: func(input *awssecretsmanager.UntagResourceInput) awssecretsmanager.UntagResourceRequest {
						return awssecretsmanager.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UntagResourceOutput{}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
		},
		"FailedDescribeSecretRequest": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				err: errors.Wrap(errBoom, errDescribeSecretFailed),
			},
		},
		"FailedGetSecret": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				err: errors.Wrap(errBoom, errK8sSecretNotFound),
			},
		},
		"FailedUpdateSecretRequest": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockUpdateSecretRequest: func(input *awssecretsmanager.UpdateSecretInput) awssecretsmanager.UpdateSecretRequest {
						return awssecretsmanager.UpdateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
					MockTagResourceRequest: func(input *awssecretsmanager.TagResourceInput) awssecretsmanager.TagResourceRequest {
						return awssecretsmanager.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.TagResourceOutput{}},
						}
					},
					MockUntagResourceRequest: func(input *awssecretsmanager.UntagResourceInput) awssecretsmanager.UntagResourceRequest {
						return awssecretsmanager.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UntagResourceOutput{}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"FailedTagResourceRequest": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{}},
						}
					},
					MockUpdateSecretRequest: func(input *awssecretsmanager.UpdateSecretInput) awssecretsmanager.UpdateSecretRequest {
						return awssecretsmanager.UpdateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UpdateSecretOutput{}},
						}
					},
					MockTagResourceRequest: func(input *awssecretsmanager.TagResourceInput) awssecretsmanager.TagResourceRequest {
						return awssecretsmanager.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
					MockUntagResourceRequest: func(input *awssecretsmanager.UntagResourceInput) awssecretsmanager.UntagResourceRequest {
						return awssecretsmanager.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UntagResourceOutput{}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
					withTags(tags),
				),
				err: errors.Wrap(errBoom, errCreateTags),
			},
		},
		"FailedUntagResourceRequest": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				secretsmanager: &fake.MockClient{
					MockDescribeSecretRequest: func(input *awssecretsmanager.DescribeSecretInput) awssecretsmanager.DescribeSecretRequest {
						return awssecretsmanager.DescribeSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DescribeSecretOutput{
								Tags: awsTags,
							}},
						}
					},
					MockUpdateSecretRequest: func(input *awssecretsmanager.UpdateSecretInput) awssecretsmanager.UpdateSecretRequest {
						return awssecretsmanager.UpdateSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.UpdateSecretOutput{}},
						}
					},
					MockTagResourceRequest: func(input *awssecretsmanager.TagResourceInput) awssecretsmanager.TagResourceRequest {
						return awssecretsmanager.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.TagResourceOutput{}},
						}
					},
					MockUntagResourceRequest: func(input *awssecretsmanager.UntagResourceInput) awssecretsmanager.UntagResourceRequest {
						return awssecretsmanager.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				err: errors.Wrap(errBoom, errRemoveTags),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.secretsmanager}
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

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.Secret
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDeletion": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDeleteSecretRequest: func(input *awssecretsmanager.DeleteSecretInput) awssecretsmanager.DeleteSecretRequest {
						return awssecretsmanager.DeleteSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DeleteSecretOutput{
								DeletionDate: &randomDate,
							}},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Deleting()),
					withDeletionDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
		},
		"SuccessfulDeletionWithForceDeleteWithoutRecoveryTrue": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDeleteSecretRequest: func(input *awssecretsmanager.DeleteSecretInput) awssecretsmanager.DeleteSecretRequest {
						return awssecretsmanager.DeleteSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DeleteSecretOutput{
								DeletionDate: &randomDate,
							}},
						}
					},
				},
				cr: secret(
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(true),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Deleting()),
					withDeletionDate(randomDate),
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(true),
				),
			},
		},
		"SuccessfulDeletionWithForceDeleteWithoutRecoveryTrueAndRecoveryWindowIsNotNil": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDeleteSecretRequest: func(input *awssecretsmanager.DeleteSecretInput) awssecretsmanager.DeleteSecretRequest {
						return awssecretsmanager.DeleteSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awssecretsmanager.ErrCodeInvalidParameterException)},
						}
					},
				},
				cr: secret(
					withDeletionDate(randomDate),
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(true),
					withRecoveryWindow(int64(7)),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Deleting()),
					withDeletionDate(randomDate),
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(true),
					withRecoveryWindow(int64(7)),
				),
				err: errors.Wrap(errors.New(awssecretsmanager.ErrCodeInvalidParameterException), errDeleteFailed),
			},
		},
		"ForceDeleteWithoutRecoveryIsFalseAndRecoveryWindowIsNil": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDeleteSecretRequest: func(input *awssecretsmanager.DeleteSecretInput) awssecretsmanager.DeleteSecretRequest {
						return awssecretsmanager.DeleteSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssecretsmanager.DeleteSecretOutput{
								DeletionDate: secretsmanager.TimeToPtr(randomDate.AddDate(0, 0, 30)),
							}},
						}
					},
				},
				cr: secret(
					withDeletionDate(time.Now().Add(5*time.Minute)),
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(false),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Deleting()),
					withDeletionDate(randomDate.AddDate(0, 0, 30)),
					withDeletedDate(randomDate),
					withSecretRef(secretName, secretNamespace, secretKey),
					withForceDeleteWithoutRecovery(false),
				),
			},
		},
		"FailedDeleteSecretRequest": {
			args: args{
				secretsmanager: &fake.MockClient{
					MockDeleteSecretRequest: func(input *awssecretsmanager.DeleteSecretInput) awssecretsmanager.DeleteSecretRequest {
						return awssecretsmanager.DeleteSecretRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: secret(
					withSecretRef(secretName, secretNamespace, secretKey),
				),
			},
			want: want{
				cr: secret(
					withConditions(runtimev1alpha1.Deleting()),
					withSecretRef(secretName, secretNamespace, secretKey),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.secretsmanager}
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
