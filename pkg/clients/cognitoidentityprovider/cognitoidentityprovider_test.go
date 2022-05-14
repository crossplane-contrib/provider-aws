package cognitoidentityprovider

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mockkube "github.com/crossplane/provider-aws/pkg/clients/mock/kube"
)

type mockModifier func(*mockkube.MockClient)

func withMockKubeClient(t *testing.T, mod mockModifier) *mockkube.MockClient {
	ctrl := gomock.NewController(t)
	mock := mockkube.NewMockClient(ctrl)
	mod(mock)
	return mock
}

var (
	testName      = "secret"
	testNamespace = "namespace"
	testID        = "id"
	testSecret    = "secret"
	errBoom       = errors.New("boom")
)

func TestGetProviderDetails(t *testing.T) {
	type args struct {
		client client.Client
		ref    *xpv1.SecretReference
	}
	type want struct {
		result map[string]*string
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"GetDetails": {
			args: args{
				client: withMockKubeClient(t, func(mcs *mockkube.MockClient) {
					mcs.EXPECT().Get(
						gomock.Any(),
						types.NamespacedName{
							Name:      testName,
							Namespace: testNamespace,
						},
						&corev1.Secret{}).DoAndReturn(func(_ context.Context, _ client.ObjectKey, s *corev1.Secret) error {
						s.Data = map[string][]byte{
							"client_id":     []byte(testID),
							"client_secret": []byte(testSecret),
						}
						return nil
					})
				}),
				ref: &xpv1.SecretReference{
					Name:      testName,
					Namespace: testNamespace,
				},
			},
			want: want{
				result: map[string]*string{
					"client_id":     &testID,
					"client_secret": &testSecret,
				},
				err: nil,
			},
		},
		"NoReference": {
			args: args{
				client: withMockKubeClient(t, func(mcs *mockkube.MockClient) {
					mcs.EXPECT().Get(
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).Times(0)
				}),
				ref: &xpv1.SecretReference{},
			},
			want: want{
				result: make(map[string]*string),
				err:    errors.New(errMissingProviderDetailsSecretRef),
			},
		},
		"NoSecret": {
			args: args{
				client: withMockKubeClient(t, func(mcs *mockkube.MockClient) {
					mcs.EXPECT().Get(
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).Return(errBoom)
				}),
				ref: &xpv1.SecretReference{
					Name:      testName,
					Namespace: testNamespace,
				},
			},
			want: want{
				result: make(map[string]*string),
				err:    errors.Wrap(errBoom, errGetProviderDetailsSecretFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, err := NewResolver().GetProviderDetails(context.Background(), tc.args.client, tc.args.ref)

			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
