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

package eks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
)

var (
	ipName           = "my-oidc-provider"
	ipClientID       = "my-oidc-clientId"
	ipIssuerURL      = "https://example.com"
	ipGroupsClaim    = "group"
	ipGroupsPrefix   = "g:"
	ipUsernameClaim  = "user"
	ipUsernamePrefix = "u:"
	ipRequiredClaims = map[string]string{"claim1": "value1", "claim2": "value2"}
)

func TestGenerateAssociateIdentityProviderConfigInput(t *testing.T) {
	type args struct {
		name string
		p    manualv1alpha1.IdentityProviderConfigParameters
	}

	cases := map[string]struct {
		args args
		want *eks.AssociateIdentityProviderConfigInput
	}{
		"AllFields": {
			args: args{
				name: ipName,
				p: manualv1alpha1.IdentityProviderConfigParameters{
					ClusterName: clusterName,
					Oidc: &manualv1alpha1.OIDCIdentityProvider{
						ClientID:       ipClientID,
						IssuerURL:      ipIssuerURL,
						GroupsClaim:    ipGroupsClaim,
						GroupsPrefix:   ipGroupsPrefix,
						RequiredClaims: ipRequiredClaims,
						UsernameClaim:  ipUsernameClaim,
						UsernamePrefix: ipUsernamePrefix,
					},
					Tags: map[string]string{"cool": "tag"},
				},
			},
			want: &eks.AssociateIdentityProviderConfigInput{
				ClusterName: &clusterName,
				Oidc: &types.OidcIdentityProviderConfigRequest{
					ClientId:                   &ipClientID,
					IdentityProviderConfigName: &ipName,
					IssuerUrl:                  &ipIssuerURL,
					GroupsClaim:                &ipGroupsClaim,
					GroupsPrefix:               &ipGroupsPrefix,
					RequiredClaims:             ipRequiredClaims,
					UsernameClaim:              &ipUsernameClaim,
					UsernamePrefix:             &ipUsernamePrefix,
				},
				Tags: map[string]string{"cool": "tag"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAssociateIdentityProviderConfigInput(tc.args.name, &tc.args.p)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDisassociateIdentityProviderConfigInput(t *testing.T) {
	type args struct {
		name    string
		cluster string
	}

	cases := map[string]struct {
		args args
		want *eks.DisassociateIdentityProviderConfigInput
	}{
		"AllFields": {
			args: args{
				name:    ipName,
				cluster: clusterName,
			},
			want: &eks.DisassociateIdentityProviderConfigInput{
				ClusterName: &clusterName,
				IdentityProviderConfig: &types.IdentityProviderConfig{
					Name: &ipName,
					Type: aws.String(string(manualv1alpha1.OidcIdentityProviderConfigType)),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateDisassociateIdentityProviderConfigInput(tc.args.name, tc.args.cluster)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIdentityProviderConfigObservation(t *testing.T) {
	ipArn := "cool:arn"

	type args struct {
		n *types.IdentityProviderConfigResponse
	}

	cases := map[string]struct {
		args args
		want manualv1alpha1.IdentityProviderConfigObservation
	}{
		"Full": {
			args: args{
				n: &types.IdentityProviderConfigResponse{
					Oidc: &types.OidcIdentityProviderConfig{
						IdentityProviderConfigArn: &ipArn,
						Status:                    types.ConfigStatusActive,
					},
				},
			},
			want: manualv1alpha1.IdentityProviderConfigObservation{
				Status:                    manualv1alpha1.IdentityProviderConfigStatusActive,
				IdentityProviderConfigArn: ipArn,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIdentityProviderConfigObservation(tc.args.n)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}

func TestTestLateInitializeIdentityProviderConfigProfile(t *testing.T) {
	type args struct {
		p *manualv1alpha1.IdentityProviderConfigParameters
		n *types.IdentityProviderConfigResponse
	}

	cases := map[string]struct {
		args args
		want *manualv1alpha1.IdentityProviderConfigParameters
	}{
		"AllFieldsEmpty": {
			args: args{
				p: &manualv1alpha1.IdentityProviderConfigParameters{},
				n: &types.IdentityProviderConfigResponse{
					Oidc: &types.OidcIdentityProviderConfig{
						Tags: map[string]string{"cool": "tag"},
					},
				},
			},
			want: &manualv1alpha1.IdentityProviderConfigParameters{
				Tags: map[string]string{"cool": "tag"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIdentityProviderConfig(tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, tc.args.p); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIdentityProviderConfigUpToDate(t *testing.T) {
	type args struct {
		p manualv1alpha1.IdentityProviderConfigParameters
		n *types.IdentityProviderConfigResponse
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				p: manualv1alpha1.IdentityProviderConfigParameters{
					Tags: map[string]string{"cool": "tag"},
				},
				n: &types.IdentityProviderConfigResponse{
					Oidc: &types.OidcIdentityProviderConfig{
						Tags: map[string]string{"cool": "tag"},
					},
				},
			},
			want: true,
		},
		"UpdateTags": {
			args: args{
				p: manualv1alpha1.IdentityProviderConfigParameters{
					Tags: map[string]string{"cool": "tag", "another": "tag"},
				},
				n: &types.IdentityProviderConfigResponse{
					Oidc: &types.OidcIdentityProviderConfig{
						Tags: map[string]string{"cool": "tag"},
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			upToDate := IsIdentityProviderConfigUpToDate(&tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, upToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
