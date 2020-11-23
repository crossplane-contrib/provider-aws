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

package secretsmanager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/secretsmanager/v1alpha1"
)

var (
	secretName  = "cool-secret"
	namespace   = "my-namespace"
	secretKey   = "my-secret"
	secretValue = "top-secret"
	secretMap   = fmt.Sprintf("{\"%s\":\"%s\"}", secretKey, string([]byte(secretValue)))
	description = "my awesome secret"
	kmsKeyID    = "kms-key-id"
	tagsKey     = "key"
	tagsValue   = "val"

	errBoom = errors.New("boom")
)

func TestIsErrorNotFound(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"IsErrorNotFound": {
			err:  errors.New(secretsmanager.ErrCodeResourceNotFoundException),
			want: true,
		},
		"Nil": {
			err:  nil,
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsErrorNotFound(tc.err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetSecretValue(t *testing.T) {
	type args struct {
		s    v1alpha1.Secret
		kube client.Client
	}
	type want struct {
		Secret string
		Err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SuccessfullyGetSecretValue": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
								Key: secretKey,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(secretValue)
						secret.Data["some-key"] = []byte("some-value")
						secret.Data["some-other-key"] = []byte("some-other-value")
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Secret: secretValue,
				Err:    nil,
			},
		},
		"ErrorGetSecret": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
								Key: secretKey,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						return errBoom
					},
				},
			},
			want: want{
				Secret: "",
				Err:    errors.Wrap(errBoom, errGetSecretFailed),
			},
		},
		"EmptySecretData": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
								Key: secretKey,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Secret: "",
				Err:    nil,
			},
		},
		"NoKeyGivenInSecretKeySelector": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(secretValue)
						secret.Data["some-key"] = []byte("some-value")
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Secret: fmt.Sprintf("{\"%s\":\"%s\",\"%s\":\"%s\"}",
					secretKey, secretValue,
					"some-key", "some-value"),
				Err: nil,
			},
		},
		"SecretKeyNotInSecretData": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
								Key: secretKey,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data["some-key"] = []byte("some-value")
						secret.Data["some-other-key"] = []byte("some-other-value")
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Secret: "",
				Err:    errors.New(errKeyNotFoundInSecretData),
			},
		},
		"NoKeyGivenInSecretKeySelectorAndSecretDataLengthOfOne": {
			args: args{
				s: v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Region: "aws-region",
							SecretRef: &v1alpha1.SecretSelector{
								SecretReference: &runtimev1.SecretReference{
									Name:      secretName,
									Namespace: namespace,
								},
								Key: secretKey,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(secretValue)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Secret: secretValue,
				Err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			secret, err := GetSecretValue(ctx, tc.args.kube, &tc.args.s)
			if diff := cmp.Diff(tc.want, want{
				Secret: secret,
				Err:    err,
			}, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateSecretsmanagerInput(t *testing.T) {
	type args struct {
		name   string
		p      *v1alpha1.SecretParameters
		secret string
	}

	cases := map[string]struct {
		args args
		want *secretsmanager.CreateSecretInput
	}{
		"AllFields": {
			args: args{
				name: secretName,
				p: &v1alpha1.SecretParameters{
					Region:      "aws-region",
					Description: &description,
					KmsKeyID:    &kmsKeyID,
					SecretRef: &v1alpha1.SecretSelector{
						SecretReference: &runtimev1.SecretReference{
							Name:      secretName,
							Namespace: namespace,
						},
						Key: secretKey,
					},
					Tags: []v1alpha1.Tag{
						{
							Key:   tagsKey,
							Value: tagsValue,
						},
					},
				},
				secret: secretMap,
			},
			want: &secretsmanager.CreateSecretInput{
				Description:  &description,
				KmsKeyId:     &kmsKeyID,
				Name:         &secretName,
				SecretString: &secretMap,
				Tags: []secretsmanager.Tag{
					{
						Key:   &tagsKey,
						Value: &tagsValue,
					},
				},
			},
		},
		"SomeFields": {
			args: args{
				name: secretName,
				p: &v1alpha1.SecretParameters{
					Region: "aws-region",
					SecretRef: &v1alpha1.SecretSelector{
						SecretReference: &runtimev1.SecretReference{
							Name:      secretName,
							Namespace: namespace,
						},
						Key: secretKey,
					},
				},
				secret: secretMap,
			},
			want: &secretsmanager.CreateSecretInput{
				Name:         &secretName,
				SecretString: &secretMap,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateSecretsmanagerInput(tc.args.name, tc.args.p, tc.args.secret)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateSecretInput(t *testing.T) {
	type args struct {
		name   string
		p      v1alpha1.SecretParameters
		secret string
	}

	cases := map[string]struct {
		args args
		want *secretsmanager.UpdateSecretInput
	}{
		"AllFields": {
			args: args{
				name: secretName,
				p: v1alpha1.SecretParameters{
					Region:      "aws-region",
					Description: &description,
					KmsKeyID:    &kmsKeyID,
					SecretRef: &v1alpha1.SecretSelector{
						SecretReference: &runtimev1.SecretReference{
							Name:      secretName,
							Namespace: namespace,
						},
						Key: secretKey,
					},
				},
				secret: secretMap,
			},
			want: &secretsmanager.UpdateSecretInput{
				Description:  &description,
				KmsKeyId:     &kmsKeyID,
				SecretId:     &secretName,
				SecretString: &secretMap,
			},
		},
		"OnlySecret": {
			args: args{
				name:   secretName,
				secret: secretMap,
				p: v1alpha1.SecretParameters{
					Region: "aws-region",
					SecretRef: &v1alpha1.SecretSelector{
						SecretReference: &runtimev1.SecretReference{
							Name:      secretName,
							Namespace: namespace,
						},
						Key: secretKey,
					},
				},
			},
			want: &secretsmanager.UpdateSecretInput{
				SecretId:     &secretName,
				SecretString: &secretMap,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateSecretInput(tc.args.name, tc.args.p, tc.args.secret)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDeleteSecretInput(t *testing.T) {
	type args struct {
		name string
		p    v1alpha1.SecretParameters
	}

	cases := map[string]struct {
		args args
		want *secretsmanager.DeleteSecretInput
	}{
		"WithRecoveryWindow": {
			args: args{
				name: secretName,
				p: v1alpha1.SecretParameters{
					Region:               "aws-region",
					RecoveryWindowInDays: IntToPtr(int64(8)),
				},
			},
			want: &secretsmanager.DeleteSecretInput{
				SecretId:             &secretName,
				RecoveryWindowInDays: IntToPtr(int64(8)),
			},
		},
		"ForceDeletion": {
			args: args{
				name: secretName,
				p: v1alpha1.SecretParameters{
					Region:                     "aws-region",
					ForceDeleteWithoutRecovery: boolToPtr(true),
				},
			},
			want: &secretsmanager.DeleteSecretInput{
				SecretId:                   &secretName,
				ForceDeleteWithoutRecovery: boolToPtr(true),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateDeleteSecretInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateObservation(t *testing.T) {
	var (
		createdDate    = time.Now()
		newCreatedDate = time.Now().Add(time.Minute)
		deletedDate    = time.Now().Add(2 * time.Minute)
	)
	type args struct {
		o   *v1alpha1.SecretObservation
		svr *secretsmanager.GetSecretValueResponse
		do  *secretsmanager.DescribeSecretOutput
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.SecretObservation
	}{
		"AllFields": {
			args: args{
				o: &v1alpha1.SecretObservation{
					CreatedDate: &metav1.Time{Time: createdDate},
				},
				svr: &secretsmanager.GetSecretValueResponse{
					GetSecretValueOutput: &secretsmanager.GetSecretValueOutput{
						CreatedDate: &newCreatedDate,
					},
				},
				do: &secretsmanager.DescribeSecretOutput{
					DeletedDate: &deletedDate,
				},
			},
			want: &v1alpha1.SecretObservation{
				CreatedDate: &metav1.Time{Time: newCreatedDate},
				DeletedDate: &metav1.Time{Time: deletedDate},
			},
		},
		"OnlyDeletedDate": {
			args: args{
				o: &v1alpha1.SecretObservation{
					CreatedDate: &metav1.Time{Time: createdDate},
				},
				svr: nil,
				do: &secretsmanager.DescribeSecretOutput{
					DeletedDate: &deletedDate,
				},
			},
			want: &v1alpha1.SecretObservation{
				CreatedDate: &metav1.Time{Time: createdDate},
				DeletedDate: &metav1.Time{Time: deletedDate},
			},
		},
		"Nothing": {
			args: args{
				o: &v1alpha1.SecretObservation{
					CreatedDate: &metav1.Time{Time: createdDate},
				},
				svr: nil,
				do:  nil,
			},
			want: &v1alpha1.SecretObservation{
				CreatedDate: &metav1.Time{Time: createdDate},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			UpdateObservation(tc.args.o, tc.args.svr, tc.args.do)
			if diff := cmp.Diff(tc.want, tc.args.o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		p  *v1alpha1.SecretParameters
		so *secretsmanager.DescribeSecretOutput
	}

	cases := map[string]struct {
		args args
		want *v1alpha1.SecretParameters
	}{
		"AllFieldsEmpty": {
			args: args{
				p: &v1alpha1.SecretParameters{},
				so: &secretsmanager.DescribeSecretOutput{
					Description: &description,
					KmsKeyId:    &kmsKeyID,
					Tags: []secretsmanager.Tag{
						{
							Key:   &tagsKey,
							Value: &tagsValue,
						},
					},
				},
			},
			want: &v1alpha1.SecretParameters{
				Description: &description,
				KmsKeyID:    &kmsKeyID,
				Tags: []v1alpha1.Tag{
					{
						Key:   tagsKey,
						Value: tagsValue,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.args.p, tc.args.so)
			if diff := cmp.Diff(tc.want, tc.args.p); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr     *v1alpha1.Secret
		req    *secretsmanager.DescribeSecretResponse
		secret string
		svo    *secretsmanager.GetSecretValueResponse
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"IsUpToDate": {
			args: args{
				cr: &v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Description: &description,
							KmsKeyID:    &kmsKeyID,
							Tags: []v1alpha1.Tag{
								{
									Key:   tagsKey,
									Value: tagsValue,
								},
							},
						},
					},
				},
				req: &secretsmanager.DescribeSecretResponse{
					DescribeSecretOutput: &secretsmanager.DescribeSecretOutput{
						Description: &description,
						KmsKeyId:    &kmsKeyID,
						Tags: []secretsmanager.Tag{
							{
								Key:   &tagsKey,
								Value: &tagsValue,
							},
						},
					},
				},
				secret: secretMap,
				svo: &secretsmanager.GetSecretValueResponse{
					GetSecretValueOutput: &secretsmanager.GetSecretValueOutput{
						SecretString: &secretMap,
					},
				},
			},
			want: true,
		},
		"IsNotUpToDate": {
			args: args{
				cr: &v1alpha1.Secret{
					Spec: v1alpha1.SecretSpec{
						ForProvider: v1alpha1.SecretParameters{
							Description: &description,
							KmsKeyID:    &kmsKeyID,
							Tags: []v1alpha1.Tag{
								{
									Key:   "some-key",
									Value: "some-val",
								},
							},
						},
					},
				},
				req: &secretsmanager.DescribeSecretResponse{
					DescribeSecretOutput: &secretsmanager.DescribeSecretOutput{
						Description: &description,
						KmsKeyId:    &kmsKeyID,
						Tags: []secretsmanager.Tag{
							{
								Key:   &tagsKey,
								Value: &tagsValue,
							},
						},
					},
				},
				secret: secretMap,
				svo: &secretsmanager.GetSecretValueResponse{
					GetSecretValueOutput: &secretsmanager.GetSecretValueOutput{
						SecretString: &secretMap,
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(tc.args.cr, tc.args.req, tc.args.secret, tc.args.svo)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
