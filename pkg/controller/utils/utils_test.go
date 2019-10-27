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

package utils

import (
	"context"
	"errors"
	"os"
	"testing"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	coption "sigs.k8s.io/controller-runtime/pkg/client"

	awsv1alpha3 "github.com/crossplaneio/stack-aws/apis/v1alpha3"
)

type mockClient struct {
	mockGet func(context.Context, types.NamespacedName, runtime.Object) error
}

func (m *mockClient) Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) error {
	return m.mockGet(ctx, key, obj)
}

func (m *mockClient) List(ctx context.Context, list runtime.Object, opts ...coption.ListOption) error {
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test_RetrieveAwsConfigFromProvider(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockSecret := &corev1.Secret{
		Data: map[string][]byte{
			"mockawskey": []byte(`[default]
aws_access_key_id = mock_aws_access_key_id
aws_secret_access_key = mock_aws_secret_access_key`),
		},
	}

	mockProvider := &awsv1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: awsv1alpha3.ProviderSpec{
			Region: "mock-region",
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{},
				Key:             "mockawskey",
			},
		},
	}

	m := mockClient{
		mockGet: func(ctx context.Context, n types.NamespacedName, o runtime.Object) error {
			switch n.Name {
			case "mockprovidername":
				obj := o.(*awsv1alpha3.Provider)
				obj.ObjectMeta = *(mockProvider.ObjectMeta.DeepCopy())
				obj.Spec = *(mockProvider.Spec.DeepCopy())
				return nil
			case "mocksecretname":
				obj := o.(*corev1.Secret)
				obj.ObjectMeta = *(mockSecret.ObjectMeta.DeepCopy())
				obj.Data = mockSecret.Data
				return nil
			}
			return errors.New("not found")
		},
	}

	for _, tc := range []struct {
		description     string
		providerName    string
		secretNamespace string
		secretName      string
		expectConfigNil bool
		expectErrNil    bool
	}{
		{
			"valid input should return expected",
			"mockprovidername",
			"mocksecretnamespace",
			"mocksecretname",
			false,
			true,
		},
		{
			"invalid provider reference should return error",
			"nonexisting",
			"mocksecretnamespace",
			"mocksecretname",
			true,
			false,
		},
		{
			"invalid secret should return error",
			"mockprovidername",
			"mocksecretnamespace",
			"nonexisting",
			true,
			false,
		},
	} {

		mockProvider.ObjectMeta.Name = tc.providerName
		mockProvider.Spec.Secret.SecretReference.Namespace = tc.secretNamespace
		mockProvider.Spec.Secret.SecretReference.Name = tc.secretName

		config, err := RetrieveAwsConfigFromProvider(context.Background(), &m, &corev1.ObjectReference{Name: mockProvider.Name, Namespace: mockProvider.Namespace})
		g.Expect(config == nil).To(gomega.Equal(tc.expectConfigNil), tc.description)
		g.Expect(err == nil).To(gomega.Equal(tc.expectErrNil), tc.description)
	}
}
