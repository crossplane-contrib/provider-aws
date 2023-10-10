package dbinstance

import (
	"context"
	"fmt"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	kubemock "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/kube"
)

const (
	secretNamespace = "crossplane-system"
)

var (
	errBoom = errors.New("boom")
)

type mockKubeClientModifier func(m *kubemock.MockClient)

func withMockKubeClient(t *testing.T, mod mockKubeClientModifier) *kubemock.MockClient {
	ctrl := gomock.NewController(t)
	mock := kubemock.NewMockClient(ctrl)
	if mod != nil {
		mod(mock)
	}
	return mock
}

func Test_getSecret(t *testing.T) {
	type args struct {
		kube client.Client
		ref  xpv1.SecretReference
	}
	type want struct {
		secret *corev1.Secret
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				ref: xpv1.SecretReference{
					Name:      "sName",
					Namespace: "sNamespace",
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Name:      "sName",
							Namespace: "sNamespace",
						}, new(corev1.Secret)).
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.ResourceVersion = "123"
							s.Data = map[string][]byte{
								"secretKey": []byte("oldPassword"),
							}
						}).Return(nil)
				}),
			},
			want: want{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "123",
					},
					Data: map[string][]byte{
						"secretKey": []byte("oldPassword"),
					},
				},
				err: nil,
			},
		},
		"Err": {
			args: args{
				ref: xpv1.SecretReference{
					Name:      "sName",
					Namespace: "sNamespace",
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Name:      "sName",
							Namespace: "sNamespace",
						}, new(corev1.Secret)).
						Return(errBoom)
				}),
			},
			want: want{
				secret: new(corev1.Secret),
				err:    errBoom,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			secret, err := getSecret(context.Background(), tc.args.kube, tc.args.ref)

			if diff := cmp.Diff(tc.want.secret, secret); diff != "" {
				t.Errorf("\n%s\nGetSecret(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetSecret(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_getCachingSecretRef(t *testing.T) {
	type args struct {
		cr v1alpha1.RDSClusterOrInstance
	}

	type want struct {
		ref xpv1.SecretReference
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Instance": {
			args: args{
				cr: &v1alpha1.DBInstance{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbInstance",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
			},
			want: want{
				ref: xpv1.SecretReference{
					Name:      "dbinstance.uid",
					Namespace: secretNamespace,
				},
			},
		},
		"Cluster": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
			},
			want: want{
				ref: xpv1.SecretReference{
					Name:      "dbcluster.uid",
					Namespace: secretNamespace,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := getCachingSecretRef(tc.args.cr)

			if diff := cmp.Diff(tc.want.ref, got); diff != "" {
				t.Errorf("\n%s\nGetSecret(...): -want, +got:\n", diff)
			}
		})
	}
}

func Test_updateOrCreateSecret(t *testing.T) {
	type args struct {
		kube client.Client
		ref  xpv1.SecretReference
		kv   map[string]string
	}

	type want struct {
		secret *corev1.Secret
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any())
					m.EXPECT().
						Patch(context.Background(), gomock.Any(), gomock.Any())
				}),
				ref: xpv1.SecretReference{
					Name:      "name",
					Namespace: "namespace",
				},
				kv: map[string]string{
					"foo": "bar",
					"foz": "baz",
				},
			},
			want: want{
				secret: &corev1.Secret{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
					Data: map[string][]byte{
						"foo": []byte("bar"),
						"foz": []byte("baz"),
					},
				},
				err: nil,
			},
		},
		"Err": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()).
						Return(errBoom)
				}),
				ref: xpv1.SecretReference{
					Name:      "name",
					Namespace: "namespace",
				},
				kv: map[string]string{
					"foo": "bar",
					"foz": "baz",
				},
			},
			want: want{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
					Data: map[string][]byte{
						"foo": []byte("bar"),
						"foz": []byte("baz"),
					},
				},
				err: fmt.Errorf("cannot get object: %w", errBoom),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := updateOrCreateSecret(context.Background(), tc.args.kube, tc.args.ref, tc.args.kv)

			if diff := cmp.Diff(tc.want.secret, got); diff != "" {
				t.Errorf("\n%s\nupdateOrCreateSecret(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nupdateOrCreateSecret(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_GetSecretValue(t *testing.T) {
	type args struct {
		kube client.Client
		ref  *xpv1.SecretKeySelector
	}
	type want struct {
		value string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				ref: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "name",
						Namespace: "namespace",
					},
					Key: "key",
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Name:      "name",
							Namespace: "namespace",
						}, new(corev1.Secret)).
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"key": []byte("foo")}
						})
				}),
			},
			want: want{
				value: "foo",
				err:   nil,
			},
		},
		"NotFound": {
			args: args{
				ref: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "name",
						Namespace: "namespace",
					},
					Key: "key",
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Name:      "name",
							Namespace: "namespace",
						}, new(corev1.Secret)).
						Return(apierr.NewNotFound(corev1.Resource("secret"), "cache"))
				}),
			},
			want: want{
				value: "",
				err:   nil,
			},
		},
		"Err": {
			args: args{
				ref: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      "name",
						Namespace: "namespace",
					},
					Key: "key",
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Name:      "name",
							Namespace: "namespace",
						}, new(corev1.Secret)).
						Return(errBoom)
				}),
			},
			want: want{
				value: "",
				err:   errors.Wrap(errBoom, errGetSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GetSecretValue(context.Background(), tc.args.kube, tc.args.ref)

			if diff := cmp.Diff(tc.want.value, got); diff != "" {
				t.Errorf("\n%s\nGetSecret(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetSecret(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_GetDesiredPassword(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
	}
	type want struct {
		value string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"CachedInstance": {
			args: args{
				cr: &v1alpha1.DBInstance{},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cached
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						}).
						Return(nil)
				}),
			},
			want: want{
				value: "cachedPassword",
				err:   nil,
			},
		},
		"CachedInstanceNotFound": {
			args: args{
				cr: &v1alpha1.DBInstance{},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cached
						Return(apierr.NewNotFound(corev1.Resource("secret"), "cache"))
				}),
			},
			want: want{
				value: "",
				err:   nil,
			},
		},
		"CachedInstanceErr": {
			args: args{
				cr: &v1alpha1.DBInstance{},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cached
						Return(errBoom)
				}),
			},
			want: want{
				value: "",
				err: errors.Wrap(errors.Wrap(
					errBoom,
					errGetSecret),
					errGetCachedPassword),
			},
		},
		"MasterPassSecretSuccess": {
			args: args{
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{},
									Key:             "key",
								},
							},
						},
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT(). // cached
							Get(context.Background(), gomock.Any(), gomock.Any()).
							Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"key": []byte("cached")}
						})
					m.EXPECT(). // masterpass
							Get(context.Background(), gomock.Any(), gomock.Any()).
							Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"key": []byte("masterPassword")}
						}).
						Return(nil)
				}),
			},
			want: want{
				value: "masterPassword",
				err:   nil,
			},
		},
		"MasterPassSecretNotFound": {
			args: args{
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{},
									Key:             "key",
								},
							},
						},
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT(). // cached
							Get(context.Background(), gomock.Any(), gomock.Any()).
							Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"key": []byte("cached")}
						})
					m.EXPECT(). // masterpass
							Get(context.Background(), gomock.Any(), gomock.Any()).
							Return(apierr.NewNotFound(corev1.Resource("secret"), "masterPass"))
				}),
			},
			want: want{
				value: "",
				err:   nil,
			},
		},
		"MasterPassSecretErr": {
			args: args{
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{},
									Key:             "key",
								},
							},
						},
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT(). // cached
							Get(context.Background(), gomock.Any(), gomock.Any())
					m.EXPECT(). // masterpass
							Get(context.Background(), gomock.Any(), gomock.Any()).
							Return(errBoom)
				}),
			},
			want: want{
				value: "",
				err: errors.Wrap(errors.Wrap(
					errBoom,
					errGetSecret),
					errGetMasterPassword),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := GetDesiredPassword(context.Background(), tc.args.kube, tc.args.cr)

			if diff := cmp.Diff(tc.want.value, got); diff != "" {
				t.Errorf("\n%s\nGetDesiredPassword(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGetDesiredPassword(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_getCachedPassword(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
	}
	type want struct {
		value string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"CachedCluster": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Namespace: secretNamespace,
							Name:      "dbcluster.uid",
						}, gomock.Any()).
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						}).
						Return(nil)
				}),
			},
			want: want{
				value: "cachedPassword",
				err:   nil,
			},
		},
		"CachedClusterNotFound": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()).
						Return(apierr.NewNotFound(corev1.Resource("secret"), "cache"))
				}),
			},
			want: want{
				value: "",
				err:   nil,
			},
		},
		"CachedClusterErr": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()).
						Return(errBoom)
				}),
			},
			want: want{
				value: "",
				err:   errors.Wrap(errBoom, errGetSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := getCachedPassword(context.Background(), tc.args.kube, tc.args.cr)

			if diff := cmp.Diff(tc.want.value, got); diff != "" {
				t.Errorf("\n%s\ngetCachedPassword(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ngetCachedPassword(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_getCachedRestoreInfo(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
	}
	type want struct {
		value restoreSate
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"InfoClusterNormal": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Namespace: secretNamespace,
							Name:      "dbcluster.uid",
						}, gomock.Any()).
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateNormal)}
						}).
						Return(nil)
				}),
			},
			want: want{
				value: RestoreStateNormal,
				err:   nil,
			},
		},
		"InfoClusterUnset": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Namespace: secretNamespace,
							Name:      "dbcluster.uid",
						}, gomock.Any()).
						Return(nil)
				}),
			},
			want: want{
				value: RestoreStateNormal,
				err:   nil,
			},
		},
		"InfoClusterRestored": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), types.NamespacedName{
							Namespace: secretNamespace,
							Name:      "dbcluster.uid",
						}, gomock.Any()).
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateRestored)}
						}).
						Return(nil)
				}),
			},
			want: want{
				value: RestoreStateRestored,
				err:   nil,
			},
		},
		"InfoClusterNotFound": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()).
						Return(apierr.NewNotFound(corev1.Resource("secret"), "cached"))
				}),
			},
			want: want{
				value: RestoreStateNormal,
				err:   nil,
			},
		},
		"InfoClusterErr": {
			args: args{
				cr: &v1alpha1.DBCluster{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbCluster",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()).
						Return(errBoom)
				}),
			},
			want: want{
				value: RestoreStateNormal,
				err:   errors.Wrap(errBoom, errGetSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := getCachedRestoreInfo(context.Background(), tc.args.kube, tc.args.cr)

			if diff := cmp.Diff(tc.want.value, got); diff != "" {
				t.Errorf("\n%s\ngetCachedRestoreInfo(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ngetCachedRestoreInfo(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_Cache(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
		kv   map[string]string
	}

	type want struct {
		secret *corev1.Secret
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any())
					m.EXPECT().
						Patch(context.Background(), gomock.Any(), gomock.Any())
				}),
				cr: &v1alpha1.DBInstance{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbInstance",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
				kv: map[string]string{
					"foo": "bar",
					"foz": "baz",
				},
			},
			want: want{
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dbinstance.uid",
						Namespace: secretNamespace,
					},
					Data: map[string][]byte{
						"foo": []byte("bar"),
						"foz": []byte("baz"),
					},
				},
				err: nil,
			},
		},
		// other cases are pretty identical to updateOrCreateSecret() see impl. of Cache
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Cache(context.Background(), tc.args.kube, tc.args.cr, tc.args.kv)

			if diff := cmp.Diff(tc.want.secret, got); diff != "" {
				t.Errorf("\n%s\nCache(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nCache(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_DeleteCache(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Delete(context.Background(), &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "dbinstance.uid",
								Namespace: secretNamespace,
							},
						})
				}),
				cr: &v1alpha1.DBInstance{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbInstance",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
			},
			want: want{
				err: nil,
			},
		},
		"NotFound": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Delete(context.Background(), &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "dbinstance.uid",
								Namespace: secretNamespace,
							},
						}).
						Return(apierr.NewNotFound(corev1.Resource("secret"), "dbinstance.uid"))
				}),
				cr: &v1alpha1.DBInstance{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbInstance",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
			},
			want: want{
				err: nil,
			},
		},
		"Err": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Delete(context.Background(), &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "dbinstance.uid",
								Namespace: secretNamespace,
							},
						}).
						Return(errBoom)
				}),
				cr: &v1alpha1.DBInstance{
					TypeMeta: metav1.TypeMeta{
						Kind: "dbInstance",
					},
					ObjectMeta: metav1.ObjectMeta{
						UID: "uid",
					},
				},
			},
			want: want{
				err: errors.Wrapf(errBoom, errDeleteSecretF, "dbinstance.uid", secretNamespace),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := DeleteCache(context.Background(), tc.args.kube, tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDeleteCache(...): -want error, +got error:\n", diff)
			}
		})
	}
}

func Test_PasswordUpToDate(t *testing.T) {
	type args struct {
		kube client.Client
		cr   v1alpha1.RDSClusterOrInstance
	}

	type want struct {
		upToDate bool
		err      error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"RestoredWithAutogeneratedPassword.!UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateRestored)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"RestoredWithPasswordFromSecret.!UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateRestored)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // masterPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"passwordKey": []byte("cachedPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									Key: "passwordKey",
								},
							},
						},
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"RestoredWithChangedPasswordFromSecret.!UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateRestored)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // masterPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"passwordKey": []byte("newPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									Key: "passwordKey",
								},
							},
						},
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"NormalWithAutogeneratedPassword.UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateNormal)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"NormalWithPasswordFromSecret.UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateNormal)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("masterPassword")}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // masterPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"passwordKey": []byte("masterPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									Key: "passwordKey",
								},
							},
						},
					},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"NormalWithChangedPasswordFromSecret.!UpToDate": {
			args: args{
				kube: withMockKubeClient(t, func(m *kubemock.MockClient) {
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // restore
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{RestoreFlagCacheKay: []byte(RestoreStateNormal)}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // cachedPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{PasswordCacheKey: []byte("cachedPassword")}
						})
					m.EXPECT().
						Get(context.Background(), gomock.Any(), gomock.Any()). // masterPass
						Do(func(_ context.Context, _ types.NamespacedName, s *corev1.Secret) {
							s.Data = map[string][]byte{"passwordKey": []byte("newPassword")}
						})
				}),
				cr: &v1alpha1.DBInstance{
					Spec: v1alpha1.DBInstanceSpec{
						ForProvider: v1alpha1.DBInstanceParameters{
							CustomDBInstanceParameters: v1alpha1.CustomDBInstanceParameters{
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									Key: "passwordKey",
								},
							},
						},
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := PasswordUpToDate(context.Background(), tc.args.kube, tc.args.cr)

			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nPasswordUpToDate(...): -want, +got:\n", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nPasswordUpToDate(...): -want error, +got error:\n", diff)
			}
		})
	}
}
