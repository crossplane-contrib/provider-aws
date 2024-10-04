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

// ClusterStatusType is the status of an EKS cluster.
type ClusterStatusType string

// Cluster statuses.
const (
	ClusterStatusCreating ClusterStatusType = "CREATING"
	ClusterStatusActive   ClusterStatusType = "ACTIVE"
	ClusterStatusDeleting ClusterStatusType = "DELETING"
	ClusterStatusFailed   ClusterStatusType = "FAILED"
	ClusterStatusUpdating ClusterStatusType = "UPDATING"
)

// LogType is a type of logging.
type LogType string

// Log types.
const (
	LogTypeAPI               LogType = "api"
	LogTypeAudit             LogType = "audit"
	LogTypeAuthenticator     LogType = "authenticator"
	LogTypeControllerManager LogType = "controllerManager"
	LogTypeScheduler         LogType = "scheduler"
)

// AuthenticationMode specifies the authentication mode of the cluster
type AuthenticationMode string

const (
	AuthenticationModeApi             AuthenticationMode = "API"
	AuthenticationModeApiAndConfigMap AuthenticationMode = "API_AND_CONFIG_MAP"
	AuthenticationModeConfigMap       AuthenticationMode = "CONFIG_MAP"
)

type AccessConfig struct {
	// The desired authentication mode for the cluster.
	// +kubebuilder:validation:Enum=API;API_AND_CONFIG_MAP;CONFIG_MAP
	// +optional
	AuthenticationMode *AuthenticationMode `json:"authenticationMode,omitempty"`
}

// ClusterParameters define the desired state of an AWS Elastic Kubernetes
// Service cluster.
type ClusterParameters struct {
	// The access configuration for the cluster.
	// +optional
	AccessConfig *AccessConfig `json:"accessConfig,omitempty"`

	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your Cluster to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// The encryption configuration for the cluster.
	// +immutable
	// +optional
	EncryptionConfig []EncryptionConfig `json:"encryptionConfig,omitempty"`

	// The Kubernetes network configuration for the cluster.
	// +immutable
	// +optional
	KubernetesNetworkConfig *KubernetesNetworkConfigRequest `json:"kubernetesNetworkConfig,omitempty"`

	// Enable or disable exporting the Kubernetes control plane logs for your cluster
	// to CloudWatch Logs. By default, cluster control plane logs aren't exported
	// to CloudWatch Logs. For more information, see Amazon EKS Cluster Control
	// Plane Logs (https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html)
	// in the Amazon EKS User Guide .
	//
	// CloudWatch Logs ingestion, archive storage, and data scanning rates apply
	// to exported control plane logs. For more information, see Amazon CloudWatch
	// Pricing (http://aws.amazon.com/cloudwatch/pricing/).
	// +optional
	Logging *Logging `json:"logging,omitempty"`

	// An object representing the configuration of your local Amazon EKS cluster on an
	// Amazon Web Services Outpost. Before creating a local cluster on an Outpost,
	// review Creating an Amazon EKS cluster on an Amazon Web Services Outpost
	// (https://docs.aws.amazon.com/eks/latest/userguide/create-cluster-outpost.html)
	// in the Amazon EKS User Guide. This object isn't available for creating Amazon
	// EKS clusters on the Amazon Web Services cloud.
	// +optional
	OutpostConfig *OutpostConfigRequest `json:"outpostConfig,omitempty"`

	// The VPC configuration used by the cluster control plane. Amazon EKS VPC resources
	// have specific requirements to work properly with Kubernetes. For more information,
	// see Cluster VPC Considerations (https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html)
	// and Cluster Security Group Considerations (https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html)
	// in the Amazon EKS User Guide. You must specify at least two subnets. You
	// can specify up to five security groups, but we recommend that you use a dedicated
	// security group for your cluster control plane.
	//
	// ResourcesVpcConfig is a required field
	ResourcesVpcConfig VpcConfigRequest `json:"resourcesVpcConfig"`

	// The Amazon Resource Name (ARN) of the IAM role that provides permissions
	// for Amazon EKS to make calls to other AWS API operations on your behalf.
	// For more information, see Amazon EKS Service IAM Role (https://docs.aws.amazon.com/eks/latest/userguide/service_IAM_role.html)
	// in the Amazon EKS User Guide .
	//
	// RoleArn is a required field
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	RoleArn string `json:"roleArn,omitempty"`

	// RoleArnRef is a reference to an IAMRole used to set
	// the RoleArn.
	// +immutable
	// +optional
	RoleArnRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleArnSelector selects references to IAMRole used
	// to set the RoleArn.
	// +optional
	RoleArnSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`

	// The metadata to apply to the cluster to assist with categorization and organization.
	// Each tag consists of a key and an optional value, both of which you define.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The desired Kubernetes version for your cluster. If you don't specify a value
	// here, the latest version available in Amazon EKS is used.
	// Example: 1.15
	// +optional
	Version *string `json:"version,omitempty"`
}

// EncryptionConfig is the encryption configuration for a cluster.
type EncryptionConfig struct {

	// AWS Key Management Service (AWS KMS) customer master key (CMK). Either the
	// ARN or the alias can be used.
	Provider Provider `json:"provider"`

	// Specifies the resources to be encrypted. The only supported value is "secrets".
	Resources []string `json:"resources"`
}

// Provider is an encryption provider.
type Provider struct {

	// Amazon Resource Name (ARN) or alias of the customer master key (CMK). The
	// CMK must be symmetric, created in the same region as the cluster, and if
	// the CMK was created in a different account, the user must have access to
	// the CMK. For more information, see Allowing Users in Other Accounts to Use
	// a CMK (https://docs.aws.amazon.com/kms/latest/developerguide/key-policy-modifying-external-accounts.html)
	// in the AWS Key Management Service Developer Guide.
	KeyArn string `json:"keyArn"`
}

// IPFamily specifies the ip family
type IPFamily string

const (
	// IPFamilyIpv4 means ipv4
	IPFamilyIpv4 IPFamily = "ipv4"
	// IPFamilyIpv6 means ipv6
	IPFamilyIpv6 IPFamily = "ipv6"
)

// KubernetesNetworkConfigRequest specifies the Kubernetes network configuration for the cluster.
type KubernetesNetworkConfigRequest struct {
	// Specify which IP family is used to assign Kubernetes pod and service IP
	// addresses. If you don't specify a value, ipv4 is used by default. You can only
	// specify an IP family when you create a cluster and can't change this value once
	// the cluster is created. If you specify ipv6, the VPC and subnets that you
	// specify for cluster creation must have both IPv4 and IPv6 CIDR blocks assigned
	// to them. You can't specify ipv6 for clusters in China Regions. You can only
	// specify ipv6 for 1.21 and later clusters that use version 1.10.1 or later of the
	// Amazon VPC CNI add-on. If you specify ipv6, then ensure that your VPC meets the
	// requirements listed in the considerations listed in Assigning IPv6 addresses to
	// pods and services
	// (https://docs.aws.amazon.com/eks/latest/userguide/cni-ipv6.html) in the Amazon
	// EKS User Guide. Kubernetes assigns services IPv6 addresses from the unique local
	// address range (fc00::/7). You can't specify a custom IPv6 CIDR block. Pod
	// addresses are assigned from the subnet's IPv6 CIDR.
	IPFamily IPFamily `json:"ipFamily"`

	// Don't specify a value if you select ipv6 for ipFamily. The CIDR block to assign
	// Kubernetes service IP addresses from. If you don't specify a block, Kubernetes
	// assigns addresses from either the 10.100.0.0/16 or 172.20.0.0/16 CIDR blocks. We
	// recommend that you specify a block that does not overlap with resources in other
	// networks that are peered or connected to your VPC. The block must meet the
	// following requirements:
	//
	// * Within one of the following private IP address
	// blocks: 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16.
	//
	// * Doesn't overlap with
	// any CIDR block assigned to the VPC that you selected for VPC.
	//
	// * Between /24 and
	// /12.
	//
	// You can only specify a custom CIDR block when you create a cluster and
	// can't change this value once the cluster is created.
	ServiceIpv4Cidr string `json:"serviceIpv4Cidr,omitempty"`
}

// Logging in the logging configuration for a cluster.
type Logging struct {
	// The cluster control plane logging configuration for your cluster.
	ClusterLogging []LogSetup `json:"clusterLogging"`
}

// LogSetup specifies the logging types that are enabled.
type LogSetup struct {
	// If a log type is enabled, that log type exports its control plane logs to
	// CloudWatch Logs. If a log type isn't enabled, that log type doesn't export
	// its control plane logs. Each individual log type can be enabled or disabled
	// independently.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// The available cluster control plane log types.
	Types []LogType `json:"types,omitempty"`
}

// VpcConfigRequest specifies the VPC configuration for a cluster.
type VpcConfigRequest struct {
	// Set this value to true to enable private access for your cluster's Kubernetes
	// API server endpoint. If you enable private access, Kubernetes API requests
	// from within your cluster's VPC use the private VPC endpoint. The default
	// value for this parameter is false, which disables private access for your
	// Kubernetes API server. For more information, see Amazon EKS Cluster Endpoint
	// Access Control (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	EndpointPrivateAccess *bool `json:"endpointPrivateAccess,omitempty"`

	// Set this value to false to disable public access for your cluster's Kubernetes
	// API server endpoint. If you disable public access, your cluster's Kubernetes
	// API server can receive only requests from within the cluster VPC. The default
	// value for this parameter is true, which enables public access for your Kubernetes
	// API server. For more information, see Amazon EKS Cluster Endpoint Access
	// Control (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	EndpointPublicAccess *bool `json:"endpointPublicAccess,omitempty"`

	// The CIDR blocks that are allowed access to your cluster's public Kubernetes
	// API server endpoint. Communication to the endpoint from addresses outside
	// of the CIDR blocks that you specify is denied. The default value is 0.0.0.0/0.
	// If you've disabled private endpoint access and you have worker nodes or AWS
	// Fargate pods in the cluster, then ensure that you specify the necessary CIDR
	// blocks. For more information, see Amazon EKS Cluster Endpoint Access Control
	// (https://docs.aws.amazon.com/eks/latest/userguide/cluster-endpoint.html)
	// in the Amazon EKS User Guide.
	// +optional
	PublicAccessCidrs []string `json:"publicAccessCidrs,omitempty"`

	// Specify one or more security groups for the cross-account elastic network
	// interfaces that Amazon EKS creates to use to allow communication between
	// your worker nodes and the Kubernetes control plane. If you don't specify
	// a security group, the default security group for your VPC is used.
	// +immutable
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// Specify subnets for your Amazon EKS worker nodes. Amazon EKS creates cross-account
	// elastic network interfaces in these subnets to allow communication between
	// your worker nodes and the Kubernetes control plane.
	// +immutable
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs are references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}

// ClusterObservation is the observed state of a cluster.
type ClusterObservation struct {
	// The Amazon Resource Name (ARN) of the cluster.
	Arn string `json:"arn,omitempty"`

	// The Unix epoch timestamp in seconds for when the cluster was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The endpoint for your Kubernetes API server.
	Endpoint string `json:"endpoint,omitempty"`

	// The Base64-encoded certificate data required to communicate with your cluster.
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`

	// The identity provider information for the cluster.
	Identity Identity `json:"identity,omitempty"`

	// The kubernetes version of your Amazon EKS cluster. For more information, see
	// Kubernetes Versions (https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html)
	// in the Amazon EKS User Guide .
	Version string `json:"version,omitempty"`

	// The platform version of your Amazon EKS cluster. For more information, see
	// Platform Versions (https://docs.aws.amazon.com/eks/latest/userguide/platform-versions.html)
	// in the Amazon EKS User Guide .
	PlatformVersion string `json:"platformVersion,omitempty"`

	// An object representing the configuration of your local Amazon EKS cluster on an
	// Amazon Web Services Outpost. This object isn't available for clusters on the
	// Amazon Web Services cloud.
	OutpostConfig OutpostConfigResponse `json:"outpostConfig,omitempty"`

	// The Kubernetes network configuration for the cluster.
	KubernetesNetworkConfig KubernetesNetworkConfigResponse `json:"kubernetesNetworkConfig,omitempty"`

	// The VPC configuration used by the cluster control plane. Amazon EKS VPC resources
	// have specific requirements to work properly with Kubernetes. For more information,
	// see Cluster VPC Considerations (https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html)
	// and Cluster Security Group Considerations (https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html)
	// in the Amazon EKS User Guide.
	ResourcesVpcConfig VpcConfigResponse `json:"resourcesVpcConfig,omitempty"`

	// The access configuration for the cluster.
	AccessConfig AccessConfigResponse `json:"accessConfig,omitempty"`

	// The current status of the cluster.
	Status ClusterStatusType `json:"status,omitempty"`
}

// Identity is the identity information for a cluster.
type Identity struct {

	// The OpenID Connect (https://openid.net/connect/) identity provider information
	// for the cluster.
	OIDC OIDC `json:"oidc,omitempty"`
}

// OIDC is the OpenID Connect issuer URL.
type OIDC struct {
	// The issuer URL for the OpenID Connect identity provider.
	Issuer string `json:"issuer,omitempty"`
}

// OutpostConfigRequest describes the Outposts configuration for eks
type OutpostConfigRequest struct {
	// The Amazon EC2 instance type that you want to use for your local Amazon EKS
	// cluster on Outposts. The instance type that you specify is used for all
	// Kubernetes control plane instances. The instance type can't be changed after
	// cluster creation. Choose an instance type based on the number of nodes that your
	// cluster will have. If your cluster will have:
	//
	// * 1–20 nodes, then we recommend
	// specifying a large instance type.
	//
	// * 21–100 nodes, then we recommend specifying
	// an xlarge instance type.
	//
	// * 101–250 nodes, then we recommend specifying a
	// 2xlarge instance type.
	//
	// For a list of the available Amazon EC2 instance types,
	// see Compute and storage in Outposts rack features
	// (http://aws.amazon.com/outposts/rack/features/). The control plane is not
	// automatically scaled by Amazon EKS.
	//
	// This member is required.
	ControlPlaneInstanceType string `json:"controlPlaneInstanceType"`

	// The ARN of the Outpost that you want to use for your local Amazon EKS cluster on
	// Outposts. Only a single Outpost ARN is supported.
	//
	// This member is required.
	OutpostArns []string `json:"outpostArns"`
	// contains filtered or unexported fields
}

// OutpostConfigResponse describse the observed Outposts configuration for a cluster
type OutpostConfigResponse struct {
	// The Amazon EC2 instance type used for the control plane. The instance type is
	// the same for all control plane instances.
	//
	// This member is required.
	ControlPlaneInstanceType string `json:"controlPlaneInstanceType,omitempty"`

	// The ARN of the Outpost that you specified for use with your local Amazon EKS
	// cluster on Outposts.
	//
	// This member is required.
	OutpostArns []string `json:"outpostArns,omitempty"`
	// contains filtered or unexported fields
}

// KubernetesNetworkConfigResponse specifies the Kubernetes network configuration for the cluster.
// The response contains a value for serviceIpv6Cidr or serviceIpv4Cidr, but not both.
type KubernetesNetworkConfigResponse struct {
	// The IP family used to assign Kubernetes pod and service IP addresses. The IP
	// family is always ipv4, unless you have a 1.21 or later cluster running version
	// 1.10.1 or later of the Amazon VPC CNI add-on and specified ipv6 when you created
	// the cluster.
	IPFamily IPFamily `json:"ipFamily,omitempty"`

	// The CIDR block that Kubernetes pod and service IP addresses are assigned from.
	// Kubernetes assigns addresses from an IPv4 CIDR block assigned to a subnet that
	// the node is in. If you didn't specify a CIDR block when you created the cluster,
	// then Kubernetes assigns addresses from either the 10.100.0.0/16 or 172.20.0.0/16
	// CIDR blocks. If this was specified, then it was specified when the cluster was
	// created and it can't be changed.
	ServiceIpv4Cidr string `json:"serviceIpv4Cidr,omitempty"`

	// The CIDR block that Kubernetes pod and service IP addresses are assigned from if
	// you created a 1.21 or later cluster with version 1.10.1 or later of the Amazon
	// VPC CNI add-on and specified ipv6 for ipFamily when you created the cluster.
	// Kubernetes assigns service addresses from the unique local address range
	// (fc00::/7) because you can't specify a custom IPv6 CIDR block when you create
	// the cluster.
	ServiceIpv6Cidr string `json:"serviceIpv6Cidr,omitempty"`
}

// VpcConfigResponse is the observed VPC configuration for a cluster.
type VpcConfigResponse struct {
	// The cluster security group that was created by Amazon EKS for the cluster.
	// Managed node groups use this security group for control-plane-to-data-plane
	// communication.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId,omitempty"`

	// The VPC associated with your cluster.
	VpcID string `json:"vpcId,omitempty"`
}

// AccessConfigResponse is the observed access configuration for a cluster.
type AccessConfigResponse struct {
	// The authentication mode used for the cluster.
	AuthenticationMode AuthenticationMode `json:"authenticationMode,omitempty"`
}

// A ClusterSpec defines the desired state of an EKS Cluster.
type ClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ClusterParameters `json:"forProvider"`
}

// A ClusterStatus represents the observed state of an EKS Cluster.
type ClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Cluster is a managed resource that represents an AWS Elastic Kubernetes
// Service cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster items
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
