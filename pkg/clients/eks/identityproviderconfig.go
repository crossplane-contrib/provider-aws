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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
)

// GenerateAssociateIdentityProviderConfigInput from IdentityProviderConfigParameters
func GenerateAssociateIdentityProviderConfigInput(name string, p *manualv1alpha1.IdentityProviderConfigParameters) *eks.AssociateIdentityProviderConfigInput {
	ip := &eks.AssociateIdentityProviderConfigInput{
		ClusterName: &p.ClusterName,
		Tags:        p.Tags,
	}
	if p.Oidc != nil {
		ip.Oidc = &types.OidcIdentityProviderConfigRequest{
			ClientId:                   &p.Oidc.ClientID,
			IdentityProviderConfigName: aws.String(name),
			IssuerUrl:                  &p.Oidc.IssuerURL,
			GroupsClaim:                &p.Oidc.GroupsClaim,
			GroupsPrefix:               &p.Oidc.GroupsPrefix,
			RequiredClaims:             p.Oidc.RequiredClaims,
			UsernameClaim:              &p.Oidc.UsernameClaim,
			UsernamePrefix:             &p.Oidc.UsernamePrefix,
		}
	}
	return ip
}

// GenerateDisassociateIdentityProviderConfigInput from IdentityProviderConfigParameters
func GenerateDisassociateIdentityProviderConfigInput(name string, clusterName string) *eks.DisassociateIdentityProviderConfigInput {
	ip := &eks.DisassociateIdentityProviderConfigInput{
		ClusterName: &clusterName,
		IdentityProviderConfig: &types.IdentityProviderConfig{
			Name: aws.String(name),
			Type: aws.String(string(manualv1alpha1.OidcIdentityProviderConfigType)),
		},
	}
	return ip
}

// GenerateIdentityProviderConfigObservation is used to produce manualv1alpha1.IdentityProviderConfigObservation
// from eks.IdentityProviderConfigResponse.
func GenerateIdentityProviderConfigObservation(ip *types.IdentityProviderConfigResponse) manualv1alpha1.IdentityProviderConfigObservation {
	if ip == nil {
		return manualv1alpha1.IdentityProviderConfigObservation{}
	}
	o := manualv1alpha1.IdentityProviderConfigObservation{
		Status: manualv1alpha1.IdentityProviderConfigStatusType(ip.Oidc.Status),
	}
	if ip.Oidc != nil && ip.Oidc.IdentityProviderConfigArn != nil {
		o.IdentityProviderConfigArn = *ip.Oidc.IdentityProviderConfigArn
	}
	return o
}

// IsIdentityProviderConfigUpToDate checks whether there is a change in the tags.
// Any other field is immutable and can't be updated.
func IsIdentityProviderConfigUpToDate(p *manualv1alpha1.IdentityProviderConfigParameters, ip *types.IdentityProviderConfigResponse) bool {
	return cmp.Equal(p.Tags, ip.Oidc.Tags, cmpopts.EquateEmpty())
}

// LateInitializeIdentityProviderConfig fills the empty fields in *manualv1alpha1.IdentityProviderConfigParameters with the
// values seen in eks.IdentityProviderConfigResponse.
func LateInitializeIdentityProviderConfig(in *manualv1alpha1.IdentityProviderConfigParameters, ip *types.IdentityProviderConfigResponse) {
	if ip == nil {
		return
	}

	if len(in.Tags) == 0 {
		in.Tags = ip.Oidc.Tags
	}
}
