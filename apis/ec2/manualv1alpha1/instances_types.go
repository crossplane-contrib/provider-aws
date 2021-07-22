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

// InstanceParameters define the desired state of the Instances
type InstanceParameters struct {
	// The block device mapping entries.
	// +optional
	// BlockDeviceMappings []BlockDeviceMapping `json:"blockDeviceMappings,omitempty"`

	// Information about the Capacity Reservation targeting option. If you do not
	// specify this parameter, the instance's Capacity Reservation preference defaults
	// to open, which enables it to run in any open Capacity Reservation that has
	// matching attributes (instance type, platform, Availability Zone).
	// +optional
	// CapacityReservationSpecification *CapacityReservationSpecification `json:"capacityReservationSpecification,omitempty"`

	// Unique, case-sensitive identifier you provide to ensure the idempotency of
	// the request. If you do not specify a client token, a randomly generated token
	// is used for the request to ensure idempotency.
	//
	// For more information, see Ensuring Idempotency (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Run_Instance_Idempotency.html).
	//
	// Constraints: Maximum 64 ASCII characters
	// +optional
	// ClientToken *string `json:"clientToken,omitempty"`

	// The CPU options for the instance. For more information, see Optimizing CPU
	// Options (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	// CPUOptions *CPUOptionsRequest `json:"cpuOptions,omitempty"`

	// The credit option for CPU usage of the burstable performance instance. Valid
	// values are standard and unlimited. To change this attribute after launch,
	// use ModifyInstanceCreditSpecification (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyInstanceCreditSpecification.html).
	// For more information, see Burstable Performance Instances (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/burstable-performance-instances.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	//
	// Default: standard (T2 instances) or unlimited (T3/T3a instances)
	// +optional
	// CreditSpecification *CreditSpecificationRequest `json:"creditSpecification,omitempty"`

	// If you set this parameter to true, you can't terminate the instance using
	// the Amazon EC2 console, CLI, or API; otherwise, you can. To change this attribute
	// after launch, use ModifyInstanceAttribute (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyInstanceAttribute.html).
	// Alternatively, if you set InstanceInitiatedShutdownBehavior to terminate,
	// you can terminate the instance by running the shutdown command from the instance.
	//
	// Default: false
	// +optional
	// DisableAPITermination *bool `json:"disableAPITermination,omitempty"`

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	// +optional
	DryRun *bool `json:"dryRun,omitempty"`

	// Indicates whether the instance is optimized for Amazon EBS I/O. This optimization
	// provides dedicated throughput to Amazon EBS and an optimized configuration
	// stack to provide optimal Amazon EBS I/O performance. This optimization isn't
	// available with all instance types. Additional usage charges apply when using
	// an EBS-optimized instance.
	//
	// Default: false
	// +optional
	// EBSOptimized *bool `json:"ebsOptimized,omitempty"`

	// An elastic GPU to associate with the instance. An Elastic GPU is a GPU resource
	// that you can attach to your Windows instance to accelerate the graphics performance
	// of your applications. For more information, see Amazon EC2 Elastic GPUs (https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/elastic-graphics.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	// ElasticGpuSpecification []ElasticGpuSpecification `json:"ElasticGpuSpecification,omitempty"`

	// An elastic inference accelerator to associate with the instance. Elastic
	// inference accelerators are a resource you can attach to your Amazon EC2 instances
	// to accelerate your Deep Learning (DL) inference workloads.
	//
	// You cannot specify accelerators from different generations in the same request.
	// +optional
	// ElasticInferenceAccelerators []ElasticInferenceAccelerator `json:"elasticInferenceAccelerators,omitempty"`

	// Indicates whether an instance is enabled for hibernation. For more information,
	// see Hibernate Your Instance (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Hibernate.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	// HibernationOptions *HibernationOptionsRequest `json:"hibernationOptions,omitempty"`

	// The IAM instance profile.
	// +optional
	// IamInstanceProfile *IamInstanceProfileSpecification `json:"iamInstanceProfile,omitempty"`

	// The ID of the AMI. An AMI ID is required to launch an instance and must be
	// specified here or in a launch template.
	ImageID *string `json:"imageId"`

	// Indicates whether an instance stops or terminates when you initiate shutdown
	// from the instance (using the operating system command for system shutdown).
	//
	// Default: stop
	// +optional
	// InstanceInitiatedShutdownBehavior *string `json:"instanceInitiatedShutdownBehavior,omitempty"`

	// The market (purchasing) option for the instances.
	//
	// For RunInstances, persistent Spot Instance requests are only supported when
	// InstanceInterruptionBehavior is set to either hibernate or stop.
	// +optional
	// InstanceMarketOptions *InstanceMarketOptionsRequest `type:"structure"`

	// The instance type. For more information, see Instance Types (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	//
	// Default: m1.small
	// +optional
	InstanceType string `json:"instanceType,omitempty"`

	// [EC2-VPC] The number of IPv6 addresses to associate with the primary network
	// interface. Amazon EC2 chooses the IPv6 addresses from the range of your subnet.
	// You cannot specify this option and the option to assign specific IPv6 addresses
	// in the same request. You can specify this option if you've specified a minimum
	// number of instances to launch.
	//
	// You cannot specify this option and the network interfaces option in the same
	// request.
	// +optional
	// Ipv6AddressCount *int64 `json:"ipv6AddressCount,omitempty"`

	// [EC2-VPC] The IPv6 addresses from the range of the subnet to associate with
	// the primary network interface. You cannot specify this option and the option
	// to assign a number of IPv6 addresses in the same request. You cannot specify
	// this option if you've specified a minimum number of instances to launch.
	//
	// You cannot specify this option and the network interfaces option in the same
	// request.
	// +optional
	// IPV6Addresses []InstanceIPV6Address `json:"ipv6Addresses,omitempty"`

	// The ID of the kernel.
	//
	// We recommend that you use PV-GRUB instead of kernels and RAM disks. For more
	// information, see PV-GRUB (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UserProvidedkernels.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	// KernelID *string `json:"kernelId,omitempty"`

	// The name of the key pair. You can create a key pair using CreateKeyPair (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateKeyPair.html)
	// or ImportKeyPair (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ImportKeyPair.html).
	//
	// If you do not specify a key pair, you can't connect to the instance unless
	// you choose an AMI that is configured to allow users another way to log in.
	// +optional
	// KeyName *string `json:"keyName,omitempty"`

	// The launch template to use to launch the instances. Any parameters that you
	// specify in RunInstances override the same parameters in the launch template.
	// You can specify either the name or ID of a launch template, but not both.
	// +optional
	// LaunchTemplate *LaunchTemplateSpecification `json:"launchTemplate,omitempty"`

	// The Amazon Resource Name (ARN) of the license configuration
	// +optional
	// LicenseSpecifications []LicenseConfigurationRequest `json:"licenseSpecifications,omitempty"`

	// The maximum number of instances to launch. If you specify more instances
	// than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches
	// the largest possible number of instances above MinCount.
	//
	// Constraints: Between 1 and the maximum number you're allowed for the specified
	// instance type. For more information about the default limits, and how to
	// request an increase, see How many instances can I run in Amazon EC2 (http://aws.amazon.com/ec2/faqs/#How_many_instances_can_I_run_in_Amazon_EC2)
	// in the Amazon EC2 FAQ.
	//
	// MaxCount is a required field
	MaxCount *int64 `json:"maxCount"` // TODO this is probably updatable

	// The metadata options for the instance. For more information, see Instance
	// Metadata and User Data (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).
	// MetadataOptions *InstanceMetadataOptionsRequest `type:"structure"` TODO

	// The minimum number of instances to launch. If you specify a minimum that
	// is more instances than Amazon EC2 can launch in the target Availability Zone,
	// Amazon EC2 launches no instances.
	//
	// Constraints: Between 1 and the maximum number you're allowed for the specified
	// instance type. For more information about the default limits, and how to
	// request an increase, see How many instances can I run in Amazon EC2 (http://aws.amazon.com/ec2/faqs/#How_many_instances_can_I_run_in_Amazon_EC2)
	// in the Amazon EC2 General FAQ.
	//
	// MinCount is a required field
	MinCount *int64 `json:"minCount"` // TODO this is probably updatable

	// Specifies whether detailed monitoring is enabled for the instance.
	// +optional
	// Monitoring *bool `json:"monitoring,omitempty"`

	// The network interfaces to associate with the instance. If you specify a network
	// interface, you must specify any security groups and subnets as part of the
	// network interface.
	// +optional
	// NetworkInterfaces []InstanceNetworkInterfaceSpecification `json:"networkInterfaces,omitempty"`

	// The placement for the instance.
	// +optional
	// Placement *Placement `json:"placement,omitempty"`

	// [EC2-VPC] The primary IPv4 address. You must specify a value from the IPv4
	// address range of the subnet.
	//
	// Only one private IP address can be designated as primary. You can't specify
	// this option if you've specified the option to designate a private IP address
	// as the primary IP address in a network interface specification. You cannot
	// specify this option if you're launching more than one instance in the request.
	//
	// You cannot specify this option and the network interfaces option in the same
	// request.
	// +optional
	// PrivateIPAddress *string `json:"privateIpAddress,omitempty"`

	// The ID of the RAM disk to select. Some kernels require additional drivers
	// at launch. Check the kernel requirements for information about whether you
	// need to specify a RAM disk. To find kernel requirements, go to the AWS Resource
	// Center and search for the kernel ID.
	//
	// We recommend that you use PV-GRUB instead of kernels and RAM disks. For more
	// information, see PV-GRUB (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UserProvidedkernels.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	// RAMDiskID *string `json:"ramDiskId,omitempty"`

	// Region is the region you'd like your Instance to be created in.
	Region *string `json:"region"`

	// The IDs of the security groups. You can create a security group using CreateSecurityGroup
	// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateSecurityGroup.html).
	//
	// If you specify a network interface, you must specify any security groups
	// as part of the network interface.
	// +optional
	// SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	// SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	// SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// [EC2-Classic, default VPC] The names of the security groups. For a nondefault
	// VPC, you must use security group IDs instead.
	//
	// If you specify a network interface, you must specify any security groups
	// as part of the network interface.
	//
	// Default: Amazon EC2 uses the default security group.
	// +optional
	// SecurityGroups []string `json:"securityGroups,omitempty"`

	// SecurityGroupsRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	// SecurityGroupsRefs []xpv1.Reference `json:"securityGroupsRefs,omitempty"`

	// SecurityGroupsSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +optional
	// SecurityGroupsSelector *xpv1.Selector `json:"securityGroupsSelector,omitempty"`

	// [EC2-VPC] The ID of the subnet to launch the instance into.
	//
	// If you specify a network interface, you must specify any subnets as part
	// of the network interface.
	// +optional
	// SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRef is a reference to a Subnet used to set the SubnetID.
	// +optional
	// SubnetIDRef []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects a reference to a Subnet used to set the SubnetID.
	// +optional
	// SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// The tags to apply to the resources during launch. You can only tag instances
	// and volumes on launch. The specified tags are applied to all instances or
	// volumes that are created during launch. To tag a resource after it has been
	// created, see CreateTags (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateTags.html).
	// +immutable
	// +optional
	TagSpecifications []TagSpecification `json:"tagSpecifications,omitempty"`

	// The user data to make available to the instance. For more information, see
	// Running Commands on Your Linux Instance at Launch (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html)
	// (Linux) and Adding User Data (https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ec2-instance-metadata.html#instancedata-add-user-data)
	// (Windows). If you are using a command line tool, base64-encoding is performed
	// for you, and you can load the text from a file. Otherwise, you must provide
	// base64-encoded text. User data is limited to 16 KB.
	// +optional
	// +kubebuilder:validation:Pattern=`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`
	UserData *string `json:"userData,omitempty"`
}

// An InstanceSpec defines the desired state of Instances.
type InstanceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       InstanceParameters `json:"forProvider"`
}

// An InstanceStatus represents the observed state of Instances.
type InstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          InstanceObservation `json:"atProvider,omitempty"`
}

// InstanceObservation keeps the state for the external resource
type InstanceObservation struct{}

// +kubebuilder:object:root=true

// Instance is a managed resource that represents a specified number of AWS EC2 Instance
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InstanceList contains a list of Instances
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}
