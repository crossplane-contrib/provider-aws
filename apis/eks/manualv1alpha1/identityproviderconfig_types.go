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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IdentityProviderConfigType is a type of IdentityProviderConfig
type IdentityProviderConfigType string

const (
	// OidcIdentityProviderConfigType represent an OpenID Connect identity provider.
	OidcIdentityProviderConfigType IdentityProviderConfigType = "oidc"
)

// IdentityProviderConfigStatusType is a type of IdentityProviderConfig status.
type IdentityProviderConfigStatusType string

// Types of IdentityProviderConfig status.
const (
	IdentityProviderConfigStatusCreating     IdentityProviderConfigStatusType = "CREATING"
	IdentityProviderConfigStatusActive       IdentityProviderConfigStatusType = "ACTIVE"
	IdentityProviderConfigStatusDeleting     IdentityProviderConfigStatusType = "DELETING"
	IdentityProviderConfigStatusCreateFailed IdentityProviderConfigStatusType = "CREATE_FAILED"
	IdentityProviderConfigStatusDeleteFailed IdentityProviderConfigStatusType = "DELETE_FAILED"
)

// OIDCIdentityProvider describes an OpenID identity provider configuration
type OIDCIdentityProvider struct {
	// This is also known as audience. The ID for the client application that makes
	// authentication requests to the OpenID identity provider.
	// +immutable
	ClientID string `json:"clientId"`

	// The URL of the OpenID identity provider that allows the API server to discover
	// public signing keys for verifying tokens. The URL must begin with https:// and
	// should correspond to the iss claim in the provider's OIDC ID tokens. Per the
	// OIDC standard, path components are allowed but query parameters are not.
	// Typically the URL consists of only a hostname, like https://server.example.org
	// or https://example.com. This URL should point to the level below
	// .well-known/openid-configuration and must be publicly accessible over the
	// internet.
	// +immutable
	IssuerURL string `json:"issuerUrl"`

	// The JWT claim that the provider uses to return your groups.
	// +immutable
	// +optional
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// The prefix that is prepended to group claims to prevent clashes with existing
	// names (such as system: groups). For example, the value oidc: will create group
	// names like oidc:engineering and oidc:infra.
	// +immutable
	// +optional
	GroupsPrefix string `json:"groupsPrefix,omitempty"`

	// The key value pairs that describe required claims in the identity token. If set,
	// each claim is verified to be present in the token with a matching value. For the
	// maximum number of claims that you can require, see Amazon EKS service quotas
	// (https://docs.aws.amazon.com/eks/latest/userguide/service-quotas.html) in the
	// Amazon EKS User Guide.
	// +immutable
	// +optional
	RequiredClaims map[string]string `json:"requiredClaims,omitempty"`

	// The JSON Web Token (JWT) claim to use as the username. The default is sub, which
	// is expected to be a unique identifier of the end user. You can choose other
	// claims, such as email or name, depending on the OpenID identity provider. Claims
	// other than email are prefixed with the issuer URL to prevent naming clashes with
	// other plug-ins.
	// +immutable
	// +optional
	UsernameClaim string `json:"usernameClaim,omitempty"`

	// The prefix that is prepended to username claims to prevent clashes with existing
	// names. If you do not provide this field, and username is a value other than
	// email, the prefix defaults to issuerurl#. You can use the value - to disable all
	// prefixing.
	// +optional
	UsernamePrefix string `json:"usernamePrefix,omitempty"`
}

// IdentityProviderConfigParameters define the desired state of an AWS Elastic Kubernetes
// Service Identity Provider.
type IdentityProviderConfigParameters struct {
	// Region is the region you'd like the identity provider to be created in.
	// +immutable
	Region string `json:"region"`

	// The name of the cluster to associate the identity provider with.
	// +immutable
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterNameRef is a reference to a Cluster used to set
	// the ClusterName.
	// +immutable
	// +optional
	ClusterNameRef *xpv1.Reference `json:"clusterNameRef,omitempty"`

	// ClusterNameSelector selects references to a Cluster used
	// to set the ClusterName.
	// +optional
	ClusterNameSelector *xpv1.Selector `json:"clusterNameSelector,omitempty"`

	// An object that represents an OpenID Connect (OIDC) identity provider
	// configuration.
	// +immutable
	Oidc *OIDCIdentityProvider `json:"oidc"`

	// The metadata to apply to the configuration to assist with categorization and
	// organization. Each tag consists of a key and an optional value, both of which
	// you define.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// IdentityProviderConfigObservation is the observed state of an identity provider.
type IdentityProviderConfigObservation struct {
	// The current status of the managed identity provider config.
	Status IdentityProviderConfigStatusType `json:"status,omitempty"`

	IdentityProviderConfigArn string `json:"identityProviderConfigArn,omitempty"`
}

// A IdentityProviderConfigSpec defines the desired state of an EKS identity provider.
type IdentityProviderConfigSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       IdentityProviderConfigParameters `json:"forProvider"`
}

// An IdentityProviderConfigStatus represents the observed state of an EKS associated identity provider.
type IdentityProviderConfigStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          IdentityProviderConfigObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An IdentityProviderConfig is a managed resource that represents an AWS Elastic Kubernetes
// Service IdentityProviderConfig.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="CLUSTER",type="string",JSONPath=".spec.forProvider.clusterName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IdentityProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityProviderConfigSpec   `json:"spec"`
	Status IdentityProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IdentityProviderConfigList contains a list of IdentityProviderConfig items
type IdentityProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityProviderConfig `json:"items"`
}
