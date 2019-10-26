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

package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
	databasev1alpha1 "github.com/crossplaneio/crossplane/apis/database/v1alpha1"

	"github.com/crossplaneio/stack-aws/apis/database/v1beta1"
	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

const (
	claimNamespace = "claim-ns"
)

var (
	_ resource.ManagedConfigurator = resource.ManagedConfiguratorFn(ConfigurePostgreRDSInstance)
	_ resource.ManagedConfigurator = resource.ManagedConfiguratorFn(ConfigureMyRDSInstance)
)

func TestConfigurePostgreRDSInstance(t *testing.T) {
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
	engineVersion := "9.6"

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cm: &databasev1alpha1.PostgreSQLInstance{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec:       databasev1alpha1.PostgreSQLInstanceSpec{EngineVersion: engineVersion},
				},
				cs: &v1beta1.RDSInstanceClass{
					SpecTemplate: v1beta1.RDSInstanceClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: claimNamespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1beta1.RDSInstance{},
			},
			want: want{
				mg: &v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: claimNamespace,
								Name:      string(claimUID),
							},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						ForProvider: v1beta1.RDSInstanceParameters{
							Engine:        v1beta1.PostgresqlEngine,
							EngineVersion: &engineVersion,
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigurePostgreRDSInstance(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigurePostgreRDSInstance(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigurePostgreRDSInstance(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestConfigureMyRDSInstance(t *testing.T) {
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
	engineVersion := "5.6"

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cm: &databasev1alpha1.MySQLInstance{
					ObjectMeta: metav1.ObjectMeta{UID: claimUID},
					Spec:       databasev1alpha1.MySQLInstanceSpec{EngineVersion: engineVersion},
				},
				cs: &v1beta1.RDSInstanceClass{
					SpecTemplate: v1beta1.RDSInstanceClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							WriteConnectionSecretsToNamespace: claimNamespace,
							ProviderReference:                 &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:                     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1beta1.RDSInstance{},
			},
			want: want{
				mg: &v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy: runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
								Namespace: claimNamespace,
								Name:      string(claimUID),
							},
							ProviderReference: &corev1.ObjectReference{Name: providerName},
						},
						ForProvider: v1beta1.RDSInstanceParameters{
							Engine:        v1beta1.MysqlEngine,
							EngineVersion: &engineVersion,
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigureMyRDSInstance(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigureMyRDSInstance(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigureMyRDSInstance(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestValidateEngineVersion(t *testing.T) {
	type args struct {
		classValue string
		claimValue string
	}

	type want struct {
		err   error
		value string
	}

	cases := map[string]struct {
		args
		want
	}{
		"ClassValueUnset": {
			args: args{claimValue: "cool"},
			want: want{value: "cool"},
		},
		"ClaimValueUnset": {
			args: args{classValue: "cool"},
			want: want{value: "cool"},
		},
		"IdenticalValues": {
			args: args{classValue: "cool", claimValue: "cool"},
			want: want{value: "cool"},
		},
		"ClaimValueIsSubstring": {
			args: args{classValue: "cool", claimValue: "coo"},
			want: want{value: "cool"},
		},
		"ConflictingValues": {
			args: args{classValue: "lame", claimValue: "cool"},
			want: want{err: errors.New("claim value [cool] does not match class value [lame]")},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := validateEngineVersion(tc.args.classValue, tc.args.claimValue)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("validateEngineVersion(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(aws.String(tc.want.value), got); diff != "" {
				t.Errorf("validateEngineVersion(...): -want, +got:\n%s", diff)
			}

		})
	}
}
