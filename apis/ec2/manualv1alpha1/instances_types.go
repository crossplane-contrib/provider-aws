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
	BlockDeviceMappings []BlockDeviceMapping `json:"blockDeviceMappings,omitempty"`

	// Information about the Capacity Reservation targeting option. If you do not
	// specify this parameter, the instance's Capacity Reservation preference defaults
	// to open, which enables it to run in any open Capacity Reservation that has
	// matching attributes (instance type, platform, Availability Zone).
	// +optional
	CapacityReservationSpecification *CapacityReservationSpecification `json:"capacityReservationSpecification,omitempty"`

	// Unique, case-sensitive identifier you provide to ensure the idempotency of
	// the request. If you do not specify a client token, a randomly generated token
	// is used for the request to ensure idempotency.
	//
	// For more information, see Ensuring Idempotency (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Run_Instance_Idempotency.html).
	//
	// Constraints: Maximum 64 ASCII characters
	// +optional
	ClientToken *string `json:"clientToken,omitempty"`

	// The CPU options for the instance. For more information, see Optimizing CPU
	// Options (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	CPUOptions *CPUOptionsRequest `json:"cpuOptions,omitempty"`

	// The credit option for CPU usage of the burstable performance instance. Valid
	// values are standard and unlimited. To change this attribute after launch,
	// use ModifyInstanceCreditSpecification (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyInstanceCreditSpecification.html).
	// For more information, see Burstable Performance Instances (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/burstable-performance-instances.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	//
	// Default: standard (T2 instances) or unlimited (T3/T3a instances)
	// +optional
	CreditSpecification *CreditSpecificationRequest `json:"creditSpecification,omitempty"`

	// If you set this parameter to true, you can't terminate the instance using
	// the Amazon EC2 console, CLI, or API; otherwise, you can. To change this attribute
	// after launch, use ModifyInstanceAttribute (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyInstanceAttribute.html).
	// Alternatively, if you set InstanceInitiatedShutdownBehavior to terminate,
	// you can terminate the instance by running the shutdown command from the instance.
	//
	// Default: false
	// +optional
	DisableAPITermination *bool `json:"disableAPITermination,omitempty"`

	// Indicates whether the instance is optimized for Amazon EBS I/O. This optimization
	// provides dedicated throughput to Amazon EBS and an optimized configuration
	// stack to provide optimal Amazon EBS I/O performance. This optimization isn't
	// available with all instance types. Additional usage charges apply when using
	// an EBS-optimized instance.
	//
	// Default: false
	// +optional
	EBSOptimized *bool `json:"ebsOptimized,omitempty"`

	// An elastic GPU to associate with the instance. An Elastic GPU is a GPU resource
	// that you can attach to your Windows instance to accelerate the graphics performance
	// of your applications. For more information, see Amazon EC2 Elastic GPUs (https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/elastic-graphics.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	ElasticGPUSpecification []ElasticGPUSpecification `json:"ElasticGpuSpecification,omitempty"`

	// An elastic inference accelerator to associate with the instance. Elastic
	// inference accelerators are a resource you can attach to your Amazon EC2 instances
	// to accelerate your Deep Learning (DL) inference workloads.
	//
	// You cannot specify accelerators from different generations in the same request.
	// +optional
	ElasticInferenceAccelerators []ElasticInferenceAccelerator `json:"elasticInferenceAccelerators,omitempty"`

	// Indicates whether an instance is enabled for hibernation. For more information,
	// see Hibernate Your Instance (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Hibernate.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	HibernationOptions *HibernationOptionsRequest `json:"hibernationOptions,omitempty"`

	// The IAM instance profile.
	// +optional
	IAMInstanceProfile *IAMInstanceProfileSpecification `json:"iamInstanceProfile,omitempty"`

	// The ID of the AMI. An AMI ID is required to launch an instance and must be
	// specified here or in a launch template.
	ImageID *string `json:"imageId"`

	// Indicates whether an instance stops or terminates when you initiate shutdown
	// from the instance (using the operating system command for system shutdown).
	//
	// Default: stop
	// +optional
	InstanceInitiatedShutdownBehavior string `json:"instanceInitiatedShutdownBehavior,omitempty"`

	// The market (purchasing) option for the instances.
	//
	// For RunInstances, persistent Spot Instance requests are only supported when
	// InstanceInterruptionBehavior is set to either hibernate or stop.
	// +optional
	InstanceMarketOptions *InstanceMarketOptionsRequest `json:"instanceMarketOptions,omitempty"`

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
	IPv6AddressCount *int32 `json:"ipv6AddressCount,omitempty"`

	// [EC2-VPC] The IPv6 addresses from the range of the subnet to associate with
	// the primary network interface. You cannot specify this option and the option
	// to assign a number of IPv6 addresses in the same request. You cannot specify
	// this option if you've specified a minimum number of instances to launch.
	//
	// You cannot specify this option and the network interfaces option in the same
	// request.
	// +optional
	IPv6Addresses []InstanceIPv6Address `json:"ipv6Addresses,omitempty"`

	// The ID of the kernel.
	//
	// AWS recommends that you use PV-GRUB instead of kernels and RAM disks. For more
	// information, see PV-GRUB (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UserProvidedkernels.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	KernelID *string `json:"kernelId,omitempty"`

	// The name of the key pair. You can create a key pair using CreateKeyPair (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateKeyPair.html)
	// or ImportKeyPair (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ImportKeyPair.html).
	//
	// If you do not specify a key pair, you can't connect to the instance unless
	// you choose an AMI that is configured to allow users another way to log in.
	// +optional
	KeyName *string `json:"keyName,omitempty"`

	// The launch template to use to launch the instances. Any parameters that you
	// specify in RunInstances override the same parameters in the launch template.
	// You can specify either the name or ID of a launch template, but not both.
	// +optional
	LaunchTemplate *LaunchTemplateSpecification `json:"launchTemplate,omitempty"`

	// The Amazon Resource Name (ARN) of the license configuration
	// +optional
	LicenseSpecifications []LicenseConfigurationRequest `json:"licenseSpecifications,omitempty"`

	// The metadata options for the instance. For more information, see Instance
	// Metadata and User Data (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).
	// +optional
	MetadataOptions *InstanceMetadataOptionsRequest `json:"metadataOptions"`

	// Specifies whether detailed monitoring is enabled for the instance.
	// +optional
	Monitoring *RunInstancesMonitoringEnabled `json:"monitoring,omitempty"`

	// The network interfaces to associate with the instance. If you specify a network
	// interface, you must specify any security groups and subnets as part of the
	// network interface.
	// +optional
	NetworkInterfaces []InstanceNetworkInterfaceSpecification `json:"networkInterfaces,omitempty"`

	// The placement for the instance.
	// +optional
	Placement *Placement `json:"placement,omitempty"`

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
	PrivateIPAddress *string `json:"privateIpAddress,omitempty"`

	// The ID of the RAM disk to select. Some kernels require additional drivers
	// at launch. Check the kernel requirements for information about whether you
	// need to specify a RAM disk. To find kernel requirements, go to the AWS Resource
	// Center and search for the kernel ID.
	//
	// AWS recommends that you use PV-GRUB instead of kernels and RAM disks. For more
	// information, see PV-GRUB (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UserProvidedkernels.html)
	// in the Amazon Elastic Compute Cloud User Guide.
	// +optional
	RAMDiskID *string `json:"ramDiskId,omitempty"`

	// Region is the region you'd like your Instance to be created in.
	Region *string `json:"region"`

	// The IDs of the security groups. You can create a security group using CreateSecurityGroup
	// (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateSecurityGroup.html).
	//
	// If you specify a network interface, you must specify any security groups
	// as part of the network interface.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupSelector
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupsRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupRefs []xpv1.Reference `json:"securityGroupRefs,omitempty"`

	// SecurityGroupsSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupSelector *xpv1.Selector `json:"securityGroupSelector,omitempty"`

	// [EC2-VPC] The ID of the subnet to launch the instance into.
	//
	// If you specify a network interface, you must specify any subnets as part
	// of the network interface.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRef is a reference to a Subnet used to set the SubnetID.
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector selects a reference to a Subnet used to set the SubnetID.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// Tags are used as identification helpers between AWS resources.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

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

// InstanceObservation keeps the state for the external resource. The below fields
// follow the Instance response object as closely as possible.
type InstanceObservation struct {
	// +optional
	AmiLaunchIndex *int32 `json:"amiLaunchIndex,omitempty"`
	Architecture   string `json:"architecture"`
	// +optional
	BlockDeviceMapping []InstanceBlockDeviceMapping `json:"blockDeviceMapping,omitempty"`
	// +optional
	CapacityReservationID *string `json:"capacityReservationId,omitempty"`
	// +optional
	CapacityReservationSpecification *CapacityReservationSpecificationResponse `json:"capacityReservationSpecification,omitempty"`
	// +optional
	ClientToken *string `json:"clientToken,omitempty"`
	// +optional
	CPUOptons *CPUOptionsRequest `json:"cpuOptions,omitempty"`
	// +optional
	DisableAPITermination *bool `json:"disableAPITermination,omitempty"`
	// +optional
	EBSOptimized *bool `json:"ebsOptimized,omitempty"`
	// +optional
	ElasticGPUAssociations []ElasticGPUAssociation `json:"elasticGpuAssociation,omitempty"`
	// +optional
	ElasticInferenceAcceleratorAssociations []ElasticInferenceAcceleratorAssociation `json:"elasticInferenceAcceleratorAssociations,omitempty"`
	// +optional
	EnaSupport *bool `json:"enaSupport,omitempty"`
	// +optional
	HibernationOptions *HibernationOptionsRequest `json:"hibernationOptions,omitempty"`
	// +optional
	Hypervisor string `json:"hypervisor"`
	// +optional
	IAMInstanceProfile *IAMInstanceProfile `json:"iamInstanceProfile,omitempty"`
	// +optional
	ImageID *string `json:"imageId,omitempty"`
	// +optional
	InstanceID *string `json:"instanceId,omitempty"`
	// +optional
	InstanceInitiatedShutdownBehavior *string `json:"instanceInitiatedShutdownBehavior,omitempty"`
	// +optional
	InstanceLifecycle string `json:"instanceLifecyle"`
	// Supported instance family when set instanceInterruptionBehavior to hibernate
	// C3, C4, C5, M4, M5, R3, R4
	InstanceType string `json:"instanceType"`
	// +optional
	KernelID *string `json:"kernelId,omitempty"`
	// +optional
	LaunchTime *metav1.Time `json:"launchTime,omitempty"`
	// +optional
	Licenses []LicenseConfigurationRequest `json:"licenseSet,omitempty"`
	// +optional
	MetadataOptions *InstanceMetadataOptionsRequest `json:"metadataOptions,omitempty"`
	// +optional
	Monitoring *Monitoring `json:"monitoring,omitempty"`
	// +optional
	NetworkInterfaces []InstanceNetworkInterface `json:"networkInterfaces,omitempty"`
	// +optional
	OutpostARN *string `json:"outpostArn,omitempty"`
	// +optional
	Placement *Placement `json:"placement,omitempty"`
	Platform  string     `json:"platform"`
	// +optional
	PrivateDNSName *string `json:"privateDnsName,omitempty"`
	// +optional
	PrivateIPAddress *string `json:"privateIpAddress,omitempty"`
	// +optional
	ProductCodes []ProductCode `json:"productCodes,omitempty"`
	// +optional
	PublicDNSName *string `json:"publicDnsName,omitempty"`
	// +optional
	PublicIPAddress *string `json:"publicIpAddress,omitempty"`
	// +optional
	RAMDiskID *string `json:"ramDiskId,omitempty"`
	// +optional
	RootDeviceName *string `json:"ebs,omitempty"`
	RootDeviceType string  `json:"rootDeviceType"`
	// +optional
	SecurityGroups []GroupIdentifier `json:"securityGroups,omitempty"`
	// +optional
	SourceDestCheck *bool `json:"sourceDestCheck,omitempty"`
	// +optional
	SpotInstanceRequestID *string `json:"spotInstanceId,omitempty"`
	// +optional
	SriovNetSupport *string `json:"sriovNetSupport,omitempty"`
	State           string  `json:"state"`
	// +optional
	StateReason *StateReason `json:"stateReason,omitempty"`
	// +optional
	StateTransitionReason *string `json:"stateTransitionReason,omitempty"`
	// +optional
	SubnetID *string `json:"subnetId,omitempty"`
	// +optional
	Tags []Tag `json:"tags,omitempty"`
	// +optional
	UserData           *string `json:"userData,omitempty"`
	VirtualizationType string  `json:"virualizationType"`
	// +optional
	VPCID *string `json:"vpcId,omitempty"`
}

// +kubebuilder:object:root=true

// Instance is a managed resource that represents a specified number of AWS EC2 Instance
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
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
