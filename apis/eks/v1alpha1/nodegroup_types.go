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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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

	// The AMI type for your node group. GPU instance types should use the AL2_x86_64_GPU
	// AMI type, which uses the Amazon EKS-optimized Linux AMI with GPU support.
	// Non-GPU instances should use the AL2_x86_64 AMI type, which uses the Amazon
	// EKS-optimized Linux AMI.
	// +immutable
	// +optional
	AMIType *string `json:"amiType,omitempty"`

	// The name of the cluster to create the node group in.
	//
	// ClusterName is a required field
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

	// The Kubernetes version to use for your managed nodes. By default, the Kubernetes
	// version of the cluster is used, and this is the only accepted specified value.
	// +optional
	Version *string `json:"version,omitempty"`
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

	// The resources associated with the node group, such as Auto Scaling groups
	// and security groups for remote access.
	Resources NodeGroupResources `json:"resources,omitempty"`

	// The scaling configuration details for the Auto Scaling group that is created
	// for your node group.
	ScalingConfig NodeGroupScalingConfigStatus `json:"scalingConfig,omitempty"`

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
