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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeGroupStatusType is a type of NodeGroup status.
type NodeGroupStatusType string

// Types of NodeGroup status.
const (
	NodeGroupStatusCreating     NodeGroupStatusType = "CREATING"
	NodeGroupStatusActive       NodeGroupStatusType = "ACTIVE"
	NodeGroupStatusUpdating     NodeGroupStatusType = "UPDATING"
	NodeGroupStatusDeleting     NodeGroupStatusType = "DELETING"
	NodeGroupStatusCreateFailed NodeGroupStatusType = "CREATE_FAILED"
	NodeGroupStatusDeleteFailed NodeGroupStatusType = "DELETE_FAILED"
	NodeGroupStatusDegraded     NodeGroupStatusType = "DEGRADED"
)

// NodeGroupParameters define the desired state of an AWS Elastic Kubernetes
// Service NodeGroup.
type NodeGroupParameters struct {
	// Region is the region you'd like  the NodeGroup to be created in.
	Region string `json:"region"`

	// The AMI type for your node group.
	// GPU instance can use
	// AL2_x86_64_GPU AMI type,
	// which uses the Amazon EKS-optimized Linux AMI with GPU support.
	// Non-GPU instances can use
	// AL2_x86_64 (default) AMI type,
	// which uses the Amazon EKS-optimized Linux AMI or,
	// BOTTLEROCKET_ARM_64 AMI type,
	// which uses the Amazon Bottlerocket AMI for ARM instances, or
	// BOTTLEROCKET_x86_64 AMI type,
	// which uses the Amazon Bottlerocket AMI fir x86_64 instances.
	//
	// +immutable
	// +optional
	AMIType *string `json:"amiType,omitempty"`

	// The name of the cluster to create the node group in.
	//
	// ClusterName is a required field
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1.Cluster
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

	// CapacityType for your node group.
	// +kubebuilder:validation:Enum=ON_DEMAND;SPOT
	CapacityType *string `json:"capacityType,omitempty"`

	// The root device disk size (in GiB) for your node group instances. The default
	// disk size is 20 GiB.
	// +immutable
	// +optional
	DiskSize *int32 `json:"diskSize,omitempty"`

	// The instance type to use for your node group. Currently, you can specify
	// a single instance type for a node group. The default value for this parameter
	// is t3.medium. If you choose a GPU instance type, be sure to specify the AL2_x86_64_GPU
	// with the amiType parameter.
	// +immutable
	// +optional
	InstanceTypes []string `json:"instanceTypes,omitempty"`

	// The Kubernetes labels to be applied to the nodes in the node group when they
	// are created.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// An object representing a node group's launch template specification. If
	// specified, then do not specify instanceTypes, diskSize, or remoteAccess and make
	// sure that the launch template meets the requirements in
	// launchTemplateSpecification.
	LaunchTemplate *LaunchTemplateSpecification `json:"launchTemplate,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role to associate with your node
	// group. The Amazon EKS worker node kubelet daemon makes calls to AWS APIs
	// on your behalf. Worker nodes receive permissions for these API calls through
	// an IAM instance profile and associated policies. Before you can launch worker
	// nodes and register them into a cluster, you must create an IAM role for those
	// worker nodes to use when they are launched. For more information, see Amazon
	// EKS Worker Node IAM Role (https://docs.aws.amazon.com/eks/latest/userguide/worker_node_IAM_role.html)
	// in the Amazon EKS User Guide .
	//
	// NodeRole is a required field
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	NodeRole string `json:"nodeRole,omitempty"`

	// NodeRoleRef is a reference to a Cluster used to set the NodeRole.
	// +immutable
	// +optional
	NodeRoleRef *xpv1.Reference `json:"nodeRoleRef,omitempty"`

	// NodeRoleSelector selects references to a Cluster used
	// to set the NodeRole.
	// +optional
	NodeRoleSelector *xpv1.Selector `json:"nodeRoleSelector,omitempty"`

	// The AMI version of the Amazon EKS-optimized AMI to use with your node group.
	// By default, the latest available AMI version for the node group's current
	// Kubernetes version is used. For more information, see Amazon EKS-Optimized
	// Linux AMI Versions (https://docs.aws.amazon.com/eks/latest/userguide/eks-linux-ami-versions.html)
	// in the Amazon EKS User Guide.
	// +immutable
	// +optional
	ReleaseVersion *string `json:"releaseVersion,omitempty"`

	// The remote access (SSH) configuration to use with your node group.
	// +immutable
	// +optional
	RemoteAccess *RemoteAccessConfig `json:"remoteAccess,omitempty"`

	// The scaling configuration details for the Auto Scaling group that is created
	// for your node group.
	// +optional
	ScalingConfig *NodeGroupScalingConfig `json:"scalingConfig,omitempty"`

	// The subnets to use for the Auto Scaling group that is created for your node
	// group. These subnets must have the tag key kubernetes.io/cluster/CLUSTER_NAME
	// with a value of shared, where CLUSTER_NAME is replaced with the name of your
	// cluster.
	//
	// Subnets is a required field
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetSelector
	Subnets []string `json:"subnets,omitempty"`

	// SubnetRefs are references to Subnets used to set the Subnets.
	// +immutable
	// +optional
	SubnetRefs []xpv1.Reference `json:"subnetRefs,omitempty"`

	// SubnetSelector selects references to Subnets used to set the Subnets.
	// +optional
	SubnetSelector *xpv1.Selector `json:"subnetSelector,omitempty"`

	// The metadata to apply to the node group to assist with categorization and
	// organization. Each tag consists of a key and an optional value, both of which
	// you define. Node group tags do not propagate to any other resources associated
	// with the node group, such as the Amazon EC2 instances or subnets.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The Kubernetes taints to be applied to the nodes in the node group.
	Taints []Taint `json:"taints,omitempty"`

	// Specifies details on how the Nodes in this NodeGroup should be updated.
	// +optional
	UpdateConfig *NodeGroupUpdateConfig `json:"updateConfig,omitempty"`

	// The Kubernetes version to use for your managed nodes. By default, the Kubernetes
	// version of the cluster is used, and this is the only accepted specified value.
	// +optional
	Version *string `json:"version,omitempty"`
}

// Taint is a property that allows a node to repel a set of pods.
type Taint struct {
	// The effect of the taint.
	// +kubebuilder:validation:Enum=NO_SCHEDULE;NO_EXECUTE;PREFER_NO_SCHEDULE
	Effect string `json:"effect"`

	// The key of the taint.
	Key *string `json:"key,omitempty"`

	// The value of the taint.
	Value *string `json:"value,omitempty"`
}

// LaunchTemplateSpecification is an object representing a node group launch
// template specification. The launch
// template cannot include SubnetId
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html),
// IamInstanceProfile
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IamInstanceProfile.html),
// RequestSpotInstances
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotInstances.html),
// HibernationOptions
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_HibernationOptionsRequest.html),
// or TerminateInstances
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html),
// or the node group deployment or update will fail. For more information about
// launch templates, see CreateLaunchTemplate
// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateLaunchTemplate.html)
// in the Amazon EC2 API Reference. For more information about using launch
// templates with Amazon EKS, see Launch template support
// (https://docs.aws.amazon.com/eks/latest/userguide/launch-templates.html) in the
// Amazon EKS User Guide. Specify either name or id, but not both.
type LaunchTemplateSpecification struct {

	// The ID of the launch template.
	ID *string `json:"id,omitempty"`

	// The name of the launch template.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1.LaunchTemplate
	Name *string `json:"name,omitempty"`

	// NameRef is a reference to a LaunchTemplate used to set
	// the Name.
	// +immutable
	// +optional
	NameRef *xpv1.Reference `json:"nameRef,omitempty"`

	// NameSelector selects references to a LaunchTemplate used
	// to set the Name.
	// +optional
	NameSelector *xpv1.Selector `json:"nameSelector,omitempty"`

	// The version of the launch template to use. If no version is specified, then the
	// template's default version is used.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1.LaunchTemplateVersion
	Version *string `json:"version,omitempty"`

	// VersionRef is a reference to a LaunchTemplateVersion used to set
	// the Version.
	// +immutable
	// +optional
	VersionRef *xpv1.Reference `json:"versionRef,omitempty"`

	// VersionSelector selects references to a LaunchTemplateVersion used
	// to set the Version.
	// +optional
	VersionSelector *xpv1.Selector `json:"versionSelector,omitempty"`
}

// RemoteAccessConfig is the configuration for remotely accessing a node.
type RemoteAccessConfig struct {
	// The Amazon EC2 SSH key that provides access for SSH communication with the
	// worker nodes in the managed node group. For more information, see Amazon
	// EC2 Key Pairs (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)
	// in the Amazon Elastic Compute Cloud User Guide for Linux Instances.
	EC2SSHKey *string `json:"ec2SSHKey,omitempty"`

	// The security groups that are allowed SSH access (port 22) to the worker nodes.
	// If you specify an Amazon EC2 SSH key but do not specify a source security
	// group when you create a managed node group, then port 22 on the worker nodes
	// is opened to the internet (0.0.0.0/0). For more information, see Security
	// Groups for Your VPC (https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html)
	// in the Amazon Virtual Private Cloud User Guide.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SourceSecurityGroupRefs
	// +crossplane:generate:reference:selectorFieldName=SourceSecurityGroupSelector
	SourceSecurityGroups []string `json:"sourceSecurityGroups,omitempty"`

	// SourceSecurityGroupRefs are references to SecurityGroups used to set
	// the SourceSecurityGroups.
	// +optional
	SourceSecurityGroupRefs []xpv1.Reference `json:"sourceSecurityGroupRefs,omitempty"`

	// SourceSecurityGroupSelector selects references to SecurityGroups used
	// to set the SourceSecurityGroups.
	// +optional
	SourceSecurityGroupSelector *xpv1.Selector `json:"sourceSecurityGroupSelector,omitempty"`
}

// NodeGroupScalingConfig is the configuration for scaling a node group.
type NodeGroupScalingConfig struct {
	// The current number of worker nodes that the managed node group should maintain.
	// This value should be left unset if another controller, such as cluster-autoscaler,
	// is expected to manage the desired size of the node group. If not set, the initial
	// desired size will be the configured minimum size of the node group.
	// +optional
	DesiredSize *int32 `json:"desiredSize,omitempty"`

	// The maximum number of worker nodes that the managed node group can scale
	// out to. Managed node groups can support up to 100 nodes by default.
	// +optional
	MaxSize *int32 `json:"maxSize,omitempty"`

	// The minimum number of worker nodes that the managed node group can scale
	// in to. This number must be greater than zero.
	// +optional
	MinSize *int32 `json:"minSize,omitempty"`
}

// NodeGroupScalingConfigStatus is the observed scaling configuration for a node group.
type NodeGroupScalingConfigStatus struct {
	// The current number of worker nodes for the managed node group.
	DesiredSize *int32 `json:"desiredSize,omitempty"`
}

// NodeGroupUpdateConfig specifies how an Update to the NodeGroup should be
// performed.
type NodeGroupUpdateConfig struct {
	// The maximum number of nodes unavailable at once during a version update.
	// Nodes will be updated in parallel. The maximum number is 100.
	// This value or maxUnavailablePercentage is required to have a value, but
	// not both.
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`

	// The maximum percentage of nodes unavailable during a version update. This
	// percentage of nodes will be updated in parallel, up to 100 nodes at once.
	// This value or maxUnavailable is required to have a value, but not both.
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=100
	// +optional
	MaxUnavailablePercentage *int32 `json:"maxUnavailablePercentage,omitempty"`

	// Force the update if the existing node group's pods are unable to be
	// drained due to a pod disruption budget issue. If an update fails because
	// pods could not be drained, you can force the update after it fails to
	// terminate the old node whether any pods are running on the node.
	// +optional
	Force *bool `json:"force,omitempty"`
}

// NodeGroupUpdateConfigStatus is the observed update configuration for a node group.
type NodeGroupUpdateConfigStatus struct {
	// The current maximum number of nodes unavailable at once during a version update.
	// +optional
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`

	// The current maximum percentage of nodes unavailable during a version
	// update. This percentage of nodes will be updated in parallel.
	// +optional
	MaxUnavailablePercentage *int32 `json:"maxUnavailablePercentage,omitempty"`
}

// NodeGroupObservation is the observed state of a NodeGroup.
type NodeGroupObservation struct {
	// The Unix epoch timestamp in seconds for when the managed node group was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The health status of the node group. If there are issues with your node group's
	// health, they are listed here.
	Health NodeGroupHealth `json:"nodeGroupHealth,omitempty"`

	// The Unix epoch timestamp in seconds for when the managed node group was last
	// modified.
	ModifiedAt *metav1.Time `json:"modifiedAt,omitempty"`

	// The Amazon Resource Name (ARN) associated with the managed node group.
	NodeGroupArn string `json:"nodeGroupArn,omitempty"`

	// The Kubernetes version to use for your managed nodes. By default, the Kubernetes
	// version of the cluster is used, and this is the only accepted specified value.
	Version string `json:"version,omitempty"`

	// The AMI version of the Amazon EKS-optimized AMI to use with your node group. By
	// default, the latest available AMI version for the node group's current Kubernetes
	// version is used. For more information, see Amazon EKS-Optimized Linux AMI Versions
	// (https://docs.aws.amazon.com/eks/latest/userguide/eks-linux-ami-versions.html) in
	// the Amazon EKS User Guide.
	ReleaseVersion string `json:"releaseVersion,omitempty"`

	// The resources associated with the node group, such as Auto Scaling groups
	// and security groups for remote access.
	Resources NodeGroupResources `json:"resources,omitempty"`

	// The scaling configuration details for the Auto Scaling group that is created
	// for your node group.
	ScalingConfig NodeGroupScalingConfigStatus `json:"scalingConfig,omitempty"`

	// The current update configuration of the node group
	UpdateConfig NodeGroupUpdateConfigStatus `json:"updateConfig,omitempty"`

	// The current status of the managed node group.
	Status NodeGroupStatusType `json:"status,omitempty"`
}

// NodeGroupHealth describes the health of a node group.
type NodeGroupHealth struct {
	// Any issues that are associated with the node group.
	Issues []Issue `json:"issues,omitempty"`
}

// Issue is an issue with a NodeGroup.
type Issue struct {

	// A brief description of the error.
	//
	//    * AutoScalingGroupNotFound: We couldn't find the Auto Scaling group associated
	//    with the managed node group. You may be able to recreate an Auto Scaling
	//    group with the same settings to recover.
	//
	//    * Ec2SecurityGroupNotFound: We couldn't find the cluster security group
	//    for the cluster. You must recreate your cluster.
	//
	//    * Ec2SecurityGroupDeletionFailure: We could not delete the remote access
	//    security group for your managed node group. Remove any dependencies from
	//    the security group.
	//
	//    * Ec2LaunchTemplateNotFound: We couldn't find the Amazon EC2 launch template
	//    for your managed node group. You may be able to recreate a launch template
	//    with the same settings to recover.
	//
	//    * Ec2LaunchTemplateVersionMismatch: The Amazon EC2 launch template version
	//    for your managed node group does not match the version that Amazon EKS
	//    created. You may be able to revert to the version that Amazon EKS created
	//    to recover.
	//
	//    * IamInstanceProfileNotFound: We couldn't find the IAM instance profile
	//    for your managed node group. You may be able to recreate an instance profile
	//    with the same settings to recover.
	//
	//    * IamNodeRoleNotFound: We couldn't find the IAM role for your managed
	//    node group. You may be able to recreate an IAM role with the same settings
	//    to recover.
	//
	//    * AsgInstanceLaunchFailures: Your Auto Scaling group is experiencing failures
	//    while attempting to launch instances.
	//
	//    * NodeCreationFailure: Your launched instances are unable to register
	//    with your Amazon EKS cluster. Common causes of this failure are insufficient
	//    worker node IAM role (https://docs.aws.amazon.com/eks/latest/userguide/worker_node_IAM_role.html)
	//    permissions or lack of outbound internet access for the nodes.
	//
	//    * InstanceLimitExceeded: Your AWS account is unable to launch any more
	//    instances of the specified instance type. You may be able to request an
	//    Amazon EC2 instance limit increase to recover.
	//
	//    * InsufficientFreeAddresses: One or more of the subnets associated with
	//    your managed node group does not have enough available IP addresses for
	//    new nodes.
	//
	//    * AccessDenied: Amazon EKS or one or more of your managed nodes is unable
	//    to communicate with your cluster API server.
	//
	//    * InternalFailure: These errors are usually caused by an Amazon EKS server-side
	//    issue.
	Code string `json:"code,omitempty"`

	// The error message associated with the issue.
	Message string `json:"message,omitempty"`

	// The AWS resources that are afflicted by this issue.
	ResourceIDs []string `json:"resourceIds,omitempty"`
}

// NodeGroupResources describe resources in a NodeGroup.
type NodeGroupResources struct {
	// The Auto Scaling groups associated with the node group.
	AutoScalingGroups []AutoScalingGroup `json:"autoScalingGroup,omitempty"`

	// The remote access security group associated with the node group. This security
	// group controls SSH access to the worker nodes.
	RemoteAccessSecurityGroup string `json:"remoteAccessSecurityGroup,omitempty"`
}

// AutoScalingGroup is an autoscaling group associated with a NodeGroup.
type AutoScalingGroup struct {
	// The name of the Auto Scaling group associated with an Amazon EKS managed
	// node group.
	Name string `json:"name,omitempty"`
}

// A NodeGroupSpec defines the desired state of an EKS NodeGroup.
type NodeGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       NodeGroupParameters `json:"forProvider"`
}

// A NodeGroupStatus represents the observed state of an EKS NodeGroup.
type NodeGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          NodeGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A NodeGroup is a managed resource that represents an AWS Elastic Kubernetes
// Service NodeGroup.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="CLUSTER",type="string",JSONPath=".spec.forProvider.clusterName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeGroupSpec   `json:"spec"`
	Status NodeGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeGroupList contains a list of NodeGroup items
type NodeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeGroup `json:"items"`
}
