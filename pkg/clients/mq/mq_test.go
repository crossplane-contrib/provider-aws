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

package mq

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	secretNamespace      = "crossplane-system"
	connectionSecretName = "my-little-secret"
	connectionSecretKey  = "credentials"
	connectionCredData   = "confidential!"
	outputSecretName     = "my-saved-secret"

	errBoom = errors.New("boom")
)

func TestGetPassword(t *testing.T) {
	type args struct {
		in   *xpv1.SecretKeySelector
		out  *xpv1.SecretReference
		kube client.Client
	}
	type want struct {
		Pwd     string
		Changed bool
		Err     error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SamePassword": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: false,
				Err:     nil,
			},
		},
		"DifferentPassword": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("not" + connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: true,
				Err:     nil,
			},
		},
		"ErrorOnInput": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						return errBoom
					},
				},
			},
			want: want{
				Pwd:     "",
				Changed: false,
				Err:     errors.Wrap(errBoom, errGetPasswordSecretFailed),
			},
		},
		"OutputDoesNotExistYet": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							return kerrors.NewNotFound(schema.GroupResource{
								Resource: "Secret",
							}, outputSecretName)
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: true,
				Err:     nil,
			},
		},

		"NoInputPassword": {
			args: args{
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("not" + connectionCredData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Pwd:     "",
				Changed: false,
				Err:     nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			pwd, changed, err := GetPassword(ctx, tc.args.kube, tc.args.in, tc.args.out)
			if diff := cmp.Diff(tc.want, want{
				Pwd:     pwd,
				Changed: changed,
				Err:     err,
			}, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
