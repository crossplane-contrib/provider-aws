/*
Copyright 2020 The Crossplane Authors.

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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A ProviderConfigSpec defines the desired state of a ProviderConfig.
type ProviderConfigSpec struct {
	// Credentials required to authenticate to this provider.
	Credentials ProviderCredentials `json:"credentials"`

	// AssumeRole defines the options for assuming an IAM role
	AssumeRole *AssumeRoleOptions `json:"assumeRole,omitempty"`

	// AssumeRoleWithWebIdentity defines the options for assuming an IAM role with a Web Identity
	AssumeRoleWithWebIdentity *AssumeRoleWithWebIdentityOptions `json:"assumeRoleWithWebIdentity,omitempty"`

	// AssumeRoleARN to assume with provider credentials
	// This setting will be deprecated. Use the roleARN field under assumeRole instead.
	// +optional
	AssumeRoleARN *string `json:"assumeRoleARN,omitempty"`

	// ExternalID is the external ID used when assuming role.
	// This setting will be deprecated. Use the externalID field under assumeRole instead.
	// +optional
	ExternalID *string `json:"externalID,omitempty"`

	// Endpoint is where you can override the default endpoint configuration
	// of AWS calls made by the provider.
	// +optional
	Endpoint *EndpointConfig `json:"endpoint,omitempty"`
}

// ProviderCredentials required to authenticate.
type ProviderCredentials struct {
	// Source of the provider credentials.
	// +kubebuilder:validation:Enum=None;Secret;InjectedIdentity;Environment;Filesystem
	Source xpv1.CredentialsSource `json:"source"`

	xpv1.CommonCredentialSelectors `json:",inline"`
}

// Tag is session tag that can be used to assume an IAM Role
type Tag struct {
	// Name of the tag.
	// Key is a required field
	Key *string `json:"key"`

	// Value of the tag.
	// Value is a required field
	Value *string `json:"value"`
}

// AssumeRoleOptions define the options for assuming an IAM Role
// Fields are similar to the STS AssumeRoleOptions in the AWS SDK
type AssumeRoleOptions struct {
	// AssumeRoleARN to assume with provider credentials
	RoleARN *string `json:"roleARN,omitempty"`

	// ExternalID is the external ID used when assuming role.
	// +optional
	ExternalID *string `json:"externalID,omitempty"`

	// Tags is list of session tags that you want to pass. Each session tag consists of a key
	// name and an associated value. For more information about session tags, see
	// Tagging STS Sessions
	// (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_session-tags.html).
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// TransitiveTagKeys is a list of keys for session tags that you want to set as transitive. If you set a
	// tag key as transitive, the corresponding key and value passes to subsequent
	// sessions in a role chain. For more information, see Chaining Roles with Session Tags
	// (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_session-tags.html#id_session-tags_role-chaining).
	// +optional
	TransitiveTagKeys []string `json:"transitiveTagKeys,omitempty"`
}

// AssumeRoleWithWebIdentityOptions define the options for assuming an IAM Role
// Fields are similar to the STS WebIdentityRoleOptions in the AWS SDK
type AssumeRoleWithWebIdentityOptions struct {
	// AssumeRoleARN to assume with provider credentials
	RoleARN *string `json:"roleARN,omitempty"`

	// RoleSessionName is the session name, if you wish to uniquely identify this session.
	// +optional
	RoleSessionName string `json:"roleSessionName,omitempty"`
}

// EndpointConfig is used to configure the AWS client for a custom endpoint.
type EndpointConfig struct {
	// URL lets you configure the endpoint URL to be used in SDK calls.
	URL URLConfig `json:"url"`

	// Specifies if the endpoint's hostname can be modified by the SDK's API
	// client.
	//
	// If the hostname is mutable the SDK API clients may modify any part of
	// the hostname based on the requirements of the API, (e.g. adding, or
	// removing content in the hostname). Such as, Amazon S3 API client
	// prefixing "bucketname" to the hostname, or changing the
	// hostname service name component from "s3." to "s3-accesspoint.dualstack."
	// for the dualstack endpoint of an S3 Accesspoint resource.
	//
	// Care should be taken when providing a custom endpoint for an API. If the
	// endpoint hostname is mutable, and the client cannot modify the endpoint
	// correctly, the operation call will most likely fail, or have undefined
	// behavior.
	//
	// If hostname is immutable, the SDK API clients will not modify the
	// hostname of the URL. This may cause the API client not to function
	// correctly if the API requires the operation specific hostname values
	// to be used by the client.
	//
	// This flag does not modify the API client's behavior if this endpoint
	// will be used instead of Endpoint Discovery, or if the endpoint will be
	// used to perform Endpoint Discovery. That behavior is configured via the
	// API Client's Options.
	// Note that this is effective only for resources that use AWS SDK v2.
	// +optional
	HostnameImmutable *bool `json:"hostnameImmutable,omitempty"`

	// The AWS partition the endpoint belongs to.
	// +optional
	PartitionID *string `json:"partitionId,omitempty"`

	// The service name that should be used for signing the requests to the
	// endpoint.
	// +optional
	SigningName *string `json:"signingName,omitempty"`

	// The region that should be used for signing the request to the endpoint.
	// For IAM, which doesn't have any region, us-east-1 is used to sign the
	// requests, which is the only signing region of IAM.
	// +optional
	SigningRegion *string `json:"signingRegion,omitempty"`

	// The signing method that should be used for signing the requests to the
	// endpoint.
	// +optional
	SigningMethod *string `json:"signingMethod,omitempty"`

	// The source of the Endpoint. By default, this will be ServiceMetadata.
	// When providing a custom endpoint, you should set the source as Custom.
	// If source is not provided when providing a custom endpoint, the SDK may not
	// perform required host mutations correctly. Source should be used along with
	// HostnameImmutable property as per the usage requirement.
	// Note that this is effective only for resources that use AWS SDK v2.
	// +optional
	// +kubebuilder:validation:Enum=ServiceMetadata;Custom
	Source *string `json:"source,omitempty"`
}

// URLConfig lets users configure the URL of the AWS SDK calls.
type URLConfig struct {
	// You can provide a static URL that will be used regardless of the service
	// and region by choosing Static type. Alternatively, you can provide
	// configuration for dynamically resolving the URL with the config you provide
	// once you set the type as Dynamic.
	// +kubebuilder:validation:Enum=Static;Dynamic
	Type string `json:"type"`

	// Static is the full URL you'd like the AWS SDK to use.
	// Recommended for using tools like localstack where a single host is exposed
	// for all services and regions.
	// +optional
	Static *string `json:"static,omitempty"`

	// Dynamic lets you configure the behavior of endpoint URL resolver.
	// +optional
	Dynamic *DynamicURLConfig `json:"dynamic,omitempty"`
}

// DynamicURLConfig lets users configure endpoint resolving functionality.
type DynamicURLConfig struct {
	// Protocol is the HTTP protocol that will be used in the URL. Currently,
	// only http and https are supported.
	// +kubebuilder:validation:Enum=http;https
	Protocol string `json:"protocol"`

	// Host is the address of the main host that the resolver will use to
	// prepend protocol, service and region configurations.
	// For example, the final URL for EC2 in us-east-1 looks like https://ec2.us-east-1.amazonaws.com
	// You would need to use "amazonaws.com" as Host and "https" as protocol
	// to have the resolver construct it.
	Host string `json:"host"`
}

// A ProviderConfigStatus represents the status of a ProviderConfig.
type ProviderConfigStatus struct {
	xpv1.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProviderConfig configures how AWS controllers will connect to AWS API.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,aws}
// +kubebuilder:subresource:status
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,aws}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	xpv1.ProviderConfigUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// ProviderConfigUsageList contains a list of ProviderConfigUsage
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}
