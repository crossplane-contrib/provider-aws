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

package cache

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplaneio/stack-aws/apis/cache/v1alpha2"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
	cachev1alpha1 "github.com/crossplaneio/crossplane/apis/cache/v1alpha1"
)

const (
	claimVersion32 = "3.2"
	claimVersion40 = "4.0"
	claimversion99 = "9.9"
)

var (
	_                 resource.ManagedConfigurator = resource.ManagedConfiguratorFn(ConfigureReplicationGroup)
	awsClassVersion32                              = string(v1alpha2.LatestSupportedPatchVersion[claimVersion32])
)

func TestConfigureReplicationGroup(t *testing.T) {
	testEngineVersion := "3.2.10"
	type args struct {
		ctx context.Context
		cm  resource.Claim
		cs  resource.NonPortableClass
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	claimUID := types.UID("definitely-a-uuid")
	providerName := "coolprovider"

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cm: &cachev1alpha1.RedisCluster{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec:       cachev1alpha1.RedisClusterSpec{EngineVersion: "3.2"},
				},
				cs: &v1alpha2.ReplicationGroupClass{
					SpecTemplate: v1alpha2.ReplicationGroupClassSpecTemplate{
						NonPortableClassSpecTemplate: runtimev1alpha1.NonPortableClassSpecTemplate{
							ProviderReference: &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1alpha2.ReplicationGroup{},
			},
			want: want{
				mg: &v1alpha2.ReplicationGroup{
					Spec: v1alpha2.ReplicationGroupSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy:                    runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: corev1.LocalObjectReference{Name: string(claimUID)},
							ProviderReference:                &corev1.ObjectReference{Name: providerName},
						},
						ForProvider: v1alpha2.ReplicationGroupParameters{
							Engine: v1alpha2.CacheEngineRedis,
							EngineVersion: &testEngineVersion,
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigureReplicationGroup(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigureReplicationGroup(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigureReplicationGroup(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestResolveAWSClassValues(t *testing.T) {
	cases := []struct {
		name    string
		class   *v1alpha2.ReplicationGroupParameters
		claim   *cachev1alpha1.RedisCluster
		want    *v1alpha2.ReplicationGroupParameters
		wantErr error
	}{
		{
			name:    "ClassUnsetClaimUnset",
			class:   &v1alpha2.ReplicationGroupParameters{},
			claim:   &cachev1alpha1.RedisCluster{},
			want:    &v1alpha2.ReplicationGroupParameters{},
			wantErr: nil,
		},
		{
			name: "ClassSetClaimUnset",
			class: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			claim: &cachev1alpha1.RedisCluster{},
			want: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			wantErr: nil,
		},
		{
			name:  "ClassUnsetClaimSet",
			class: &v1alpha2.ReplicationGroupParameters{},
			claim: &cachev1alpha1.RedisCluster{Spec: cachev1alpha1.RedisClusterSpec{EngineVersion: claimVersion32}},
			want: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			wantErr: nil,
		},
		{
			name:    "ClassUnsetClaimSetUnsupported",
			class:   &v1alpha2.ReplicationGroupParameters{},
			claim:   &cachev1alpha1.RedisCluster{Spec: cachev1alpha1.RedisClusterSpec{EngineVersion: claimversion99}},
			want:    &v1alpha2.ReplicationGroupParameters{},
			wantErr: errors.WithStack(errors.Errorf("cannot resolve class claim values: minor version %s is not currently supported", claimversion99)),
		},
		{
			name: "ClassSetClaimSetMatching",
			class: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			claim: &cachev1alpha1.RedisCluster{Spec: cachev1alpha1.RedisClusterSpec{EngineVersion: claimVersion32}},
			want: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			wantErr: nil,
		},
		{
			name: "ClassSetClaimSetConflict",
			class: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			claim: &cachev1alpha1.RedisCluster{Spec: cachev1alpha1.RedisClusterSpec{EngineVersion: claimVersion40}},
			want: &v1alpha2.ReplicationGroupParameters{
				EngineVersion: &awsClassVersion32,
			},
			wantErr: errors.WithStack(errors.Errorf("cannot resolve class claim values: class version %s is not a patch of claim version %s", awsClassVersion32, claimVersion40)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := resolveAWSClassInstanceValues(tc.class, tc.claim)
			if diff := cmp.Diff(tc.wantErr, gotErr, test.EquateErrors()); diff != "" {
				t.Errorf("-want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.class); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
