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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane/provider-aws/apis/storage/v1alpha3"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"
)

var _ claimbinding.ManagedConfigurator = claimbinding.ManagedConfiguratorFn(ConfigureS3Bucket)

func TestConfigureBucket(t *testing.T) {
	type args struct {
		ctx context.Context
		cm  resource.Claim
		cs  resource.Class
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	claimUID := types.UID("definitely-a-uuid")
	providerName := "coolprovider"

	bucketPrivate := storagev1alpha1.ACLPrivate
	s3BucketPrivate := s3.BucketCannedACLPrivate

	ro := storagev1alpha1.ReadOnlyPermission

	cases := map[string]struct {
		args args
		want want
	}{
		"ClaimACLAndPermissions": {
			args: args{
				cm: &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec: storagev1alpha1.BucketSpec{
						PredefinedACL:   &bucketPrivate,
						LocalPermission: &ro,
					},
				},
				cs: &v1alpha3.S3BucketClass{
					SpecTemplate: v1alpha3.S3BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: namespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1alpha3.S3Bucket{},
			},
			want: want{
				mg: &v1alpha3.S3Bucket{
					Spec: v1alpha3.S3BucketSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: namespace,
								Name:      string(claimUID),
							},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						S3BucketParameters: v1alpha3.S3BucketParameters{
							CannedACL:       &s3BucketPrivate,
							LocalPermission: &ro,
						},
					},
				},
				err: nil,
			},
		},
		"ClassACLAndPermissions": {
			args: args{
				cm: &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec:       storagev1alpha1.BucketSpec{},
				},
				cs: &v1alpha3.S3BucketClass{
					SpecTemplate: v1alpha3.S3BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: namespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
						S3BucketParameters: v1alpha3.S3BucketParameters{
							CannedACL:       &s3BucketPrivate,
							LocalPermission: &ro,
						},
					},
				},
				mg: &v1alpha3.S3Bucket{},
			},
			want: want{
				mg: &v1alpha3.S3Bucket{
					Spec: v1alpha3.S3BucketSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: namespace,
								Name:      string(claimUID)},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						S3BucketParameters: v1alpha3.S3BucketParameters{
							CannedACL:       &s3BucketPrivate,
							LocalPermission: &ro,
						},
					},
				},
				err: nil,
			},
		},
		"ClassAndClaimACLAndPermissions": {
			args: args{
				cm: &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec: storagev1alpha1.BucketSpec{
						PredefinedACL:   &bucketPrivate,
						LocalPermission: &ro,
					},
				},
				cs: &v1alpha3.S3BucketClass{
					SpecTemplate: v1alpha3.S3BucketClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: namespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
						S3BucketParameters: v1alpha3.S3BucketParameters{
							CannedACL:       &s3BucketPrivate,
							LocalPermission: &ro,
						},
					},
				},
				mg: &v1alpha3.S3Bucket{},
			},
			want: want{
				mg: &v1alpha3.S3Bucket{
					Spec: v1alpha3.S3BucketSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: namespace,
								Name:      string(claimUID),
							},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						S3BucketParameters: v1alpha3.S3BucketParameters{
							CannedACL:       &s3BucketPrivate,
							LocalPermission: &ro,
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigureS3Bucket(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigureS3Bucket(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigureS3Bucket(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTranslateACL(t *testing.T) {
	auth := storagev1alpha1.ACLAuthenticatedRead
	s3Auth := s3.BucketCannedACLAuthenticatedRead
	weird := storagev1alpha1.PredefinedACL("weird")

	cases := map[string]struct {
		acl  *storagev1alpha1.PredefinedACL
		want *s3.BucketCannedACL
	}{
		"ACLIsNil": {
			acl:  nil,
			want: nil,
		},
		"ACLExists": {
			acl:  &auth,
			want: &s3Auth,
		},
		"ACLDoesNotExist": {
			acl:  &weird,
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := translateACL(tc.acl)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("translateACL(...): -want, +got:\n%s", diff)
			}
		})
	}
}
