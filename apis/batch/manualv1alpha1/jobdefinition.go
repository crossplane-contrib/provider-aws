/*
Copyright 2022 The Crossplane Authors.

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

// FargatePlatformConfiguration defines the platform configuration for jobs that are running on Fargate resources.
// Jobs that run on EC2 resources must not specify this parameter.
type FargatePlatformConfiguration struct {
	// The Fargate platform version where the jobs are running. A platform version
	// is specified only for jobs that are running on Fargate resources. If one
	// isn't specified, the LATEST platform version is used by default. This uses
	// a recent, approved version of the Fargate platform for compute resources.
	// For more information, see Fargate platform versions (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/platform_versions.html)
	// in the Amazon Elastic Container Service Developer Guide.
	PlatformVersion *string `json:"platformVersion,omitempty"`
}

// Device defines a container instance host device.
//
// This object isn't applicable to jobs that are running on Fargate resources
// and shouldn't be provided.
type Device struct {
	// The path inside the container that's used to expose the host device. By default,
	// the hostPath value is used.
	ContainerPath *string `json:"containerPath,omitempty"`

	// The path for the device on the host container instance.
	//
	// HostPath is a required field
	// +kubebuilder:validation:Required
	HostPath string `json:"hostPath"`

	// The explicit permissions to provide to the container for the device. By default,
	// the container has permissions for read, write, and mknod for the device.
	Permissions []*string `json:"permissions,omitempty"`
}

// Tmpfs defines the container path, mount options, and size of the tmpfs mount.
//
// This object isn't applicable to jobs that are running on Fargate resources.
type Tmpfs struct {
	// The absolute file path in the container where the tmpfs volume is mounted.
	//
	// ContainerPath is a required field
	// +kubebuilder:validation:Required
	ContainerPath string `json:"containerPath"`

	// The list of tmpfs volume mount options.
	//
	// Valid values: "defaults" | "ro" | "rw" | "suid" | "nosuid" | "dev" | "nodev"
	// | "exec" | "noexec" | "sync" | "async" | "dirsync" | "remount" | "mand" |
	// "nomand" | "atime" | "noatime" | "diratime" | "nodiratime" | "bind" | "rbind"
	// | "unbindable" | "runbindable" | "private" | "rprivate" | "shared" | "rshared"
	// | "slave" | "rslave" | "relatime" | "norelatime" | "strictatime" | "nostrictatime"
	// | "mode" | "uid" | "gid" | "nr_inodes" | "nr_blocks" | "mpol"
	MountOptions []*string `json:"mountOptions,omitempty"`

	// The size (in MiB) of the tmpfs volume.
	//
	// Size is a required field
	// +kubebuilder:validation:Required
	Size int64 `json:"size"`
}

// LinuxParameters define linux-specific modifications that are applied to the container, such as details
// for device mappings.
type LinuxParameters struct {
	// Any host devices to expose to the container. This parameter maps to Devices
	// in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --device option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	Devices []*Device `json:"devices,omitempty"`

	// If true, run an init process inside the container that forwards signals and
	// reaps processes. This parameter maps to the --init option to docker run (https://docs.docker.com/engine/reference/run/).
	// This parameter requires version 1.25 of the Docker Remote API or greater
	// on your container instance. To check the Docker Remote API version on your
	// container instance, log into your container instance and run the following
	// command: sudo docker version | grep "Server API version"
	InitProcessEnabled *bool `json:"initProcessEnabled,omitempty"`

	// The total amount of swap memory (in MiB) a container can use. This parameter
	// is translated to the --memory-swap option to docker run (https://docs.docker.com/engine/reference/run/)
	// where the value is the sum of the container memory plus the maxSwap value.
	// For more information, see --memory-swap details (https://docs.docker.com/config/containers/resource_constraints/#--memory-swap-details)
	// in the Docker documentation.
	//
	// If a maxSwap value of 0 is specified, the container doesn't use swap. Accepted
	// values are 0 or any positive integer. If the maxSwap parameter is omitted,
	// the container doesn't use the swap configuration for the container instance
	// it is running on. A maxSwap value must be set for the swappiness parameter
	// to be used.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	MaxSwap *int64 `json:"maxSwap,omitempty"`

	// The value for the size (in MiB) of the /dev/shm volume. This parameter maps
	// to the --shm-size option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	SharedMemorySize *int64 `json:"sharedMemorySize,omitempty"`

	// This allows you to tune a container's memory swappiness behavior. A swappiness
	// value of 0 causes swapping not to happen unless absolutely necessary. A swappiness
	// value of 100 causes pages to be swapped very aggressively. Accepted values
	// are whole numbers between 0 and 100. If the swappiness parameter isn't specified,
	// a default value of 60 is used. If a value isn't specified for maxSwap, then
	// this parameter is ignored. If maxSwap is set to 0, the container doesn't
	// use swap. This parameter maps to the --memory-swappiness option to docker
	// run (https://docs.docker.com/engine/reference/run/).
	//
	// Consider the following when you use a per-container swap configuration.
	//
	//    * Swap space must be enabled and allocated on the container instance for
	//    the containers to use. The Amazon ECS optimized AMIs don't have swap enabled
	//    by default. You must enable swap on the instance to use this feature.
	//    For more information, see Instance Store Swap Volumes (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-store-swap-volumes.html)
	//    in the Amazon EC2 User Guide for Linux Instances or How do I allocate
	//    memory to work as swap space in an Amazon EC2 instance by using a swap
	//    file? (http://aws.amazon.com/premiumsupport/knowledge-center/ec2-memory-swap-file/)
	//
	//    * The swap space parameters are only supported for job definitions using
	//    EC2 resources.
	//
	//    * If the maxSwap and swappiness parameters are omitted from a job definition,
	//    each container will have a default swappiness value of 60, and the total
	//    swap usage will be limited to two times the memory reservation of the
	//    container.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	Swappiness *int64 `json:"swappiness,omitempty"`

	// The container path, mount options, and size (in MiB) of the tmpfs mount.
	// This parameter maps to the --tmpfs option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	Tmpfs []*Tmpfs `json:"tmpfs,omitempty"`
}

// Secret defines the secret to expose to your container. Secrets can
// be exposed to a container in the following ways:
//
//   - To inject sensitive data into your containers as environment variables,
//     use the secrets container definition parameter.
//
//   - To reference sensitive information in the log configuration of a container,
//     use the secretOptions container definition parameter.
//
// For more information, see Specifying sensitive data (https://docs.aws.amazon.com/batch/latest/userguide/specifying-sensitive-data.html)
// in the Batch User Guide.
type Secret struct {
	// The name of the secret.
	//
	// Name is a required field
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The secret to expose to the container. The supported values are either the
	// full ARN of the Secrets Manager secret or the full ARN of the parameter in
	// the Amazon Web Services Systems Manager Parameter Store.
	//
	// If the Amazon Web Services Systems Manager Parameter Store parameter exists
	// in the same Region as the job you're launching, then you can use either the
	// full ARN or name of the parameter. If the parameter exists in a different
	// Region, then the full ARN must be specified.
	//
	// ValueFrom is a required field
	// +kubebuilder:validation:Required
	ValueFrom string `json:"valueFrom"`
}

// LogConfiguration defines the log configuration options to send to a custom log driver for the container.
type LogConfiguration struct {
	// The log driver to use for the container. The valid values listed for this
	// parameter are log drivers that the Amazon ECS container agent can communicate
	// with by default.
	//
	// The supported log drivers are awslogs, fluentd, gelf, json-file, journald,
	// logentries, syslog, and splunk.
	//
	// Jobs that are running on Fargate resources are restricted to the awslogs
	// and splunk log drivers.
	//
	// awslogs
	//
	// Specifies the Amazon CloudWatch Logs logging driver. For more information,
	// see Using the awslogs Log Driver (https://docs.aws.amazon.com/batch/latest/userguide/using_awslogs.html)
	// in the Batch User Guide and Amazon CloudWatch Logs logging driver (https://docs.docker.com/config/containers/logging/awslogs/)
	// in the Docker documentation.
	//
	// fluentd
	//
	// Specifies the Fluentd logging driver. For more information, including usage
	// and options, see Fluentd logging driver (https://docs.docker.com/config/containers/logging/fluentd/)
	// in the Docker documentation.
	//
	// gelf
	//
	// Specifies the Graylog Extended Format (GELF) logging driver. For more information,
	// including usage and options, see Graylog Extended Format logging driver (https://docs.docker.com/config/containers/logging/gelf/)
	// in the Docker documentation.
	//
	// journald
	//
	// Specifies the journald logging driver. For more information, including usage
	// and options, see Journald logging driver (https://docs.docker.com/config/containers/logging/journald/)
	// in the Docker documentation.
	//
	// json-file
	//
	// Specifies the JSON file logging driver. For more information, including usage
	// and options, see JSON File logging driver (https://docs.docker.com/config/containers/logging/json-file/)
	// in the Docker documentation.
	//
	// splunk
	//
	// Specifies the Splunk logging driver. For more information, including usage
	// and options, see Splunk logging driver (https://docs.docker.com/config/containers/logging/splunk/)
	// in the Docker documentation.
	//
	// syslog
	//
	// Specifies the syslog logging driver. For more information, including usage
	// and options, see Syslog logging driver (https://docs.docker.com/config/containers/logging/syslog/)
	// in the Docker documentation.
	//
	// If you have a custom driver that's not listed earlier that you want to work
	// with the Amazon ECS container agent, you can fork the Amazon ECS container
	// agent project that's available on GitHub (https://github.com/aws/amazon-ecs-agent)
	// and customize it to work with that driver. We encourage you to submit pull
	// requests for changes that you want to have included. However, Amazon Web
	// Services doesn't currently support running modified copies of this software.
	//
	// This parameter requires version 1.18 of the Docker Remote API or greater
	// on your container instance. To check the Docker Remote API version on your
	// container instance, log into your container instance and run the following
	// command: sudo docker version | grep "Server API version"
	//
	// LogDriver is a required field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=json-file;syslog;journald;gelf;fluentd;awslogs;splunk
	LogDriver string `json:"logDriver"`

	// The configuration options to send to the log driver. This parameter requires
	// version 1.19 of the Docker Remote API or greater on your container instance.
	// To check the Docker Remote API version on your container instance, log into
	// your container instance and run the following command: sudo docker version
	// | grep "Server API version"
	Options map[string]*string `json:"options,omitempty"`

	// The secrets to pass to the log configuration. For more information, see Specifying
	// Sensitive Data (https://docs.aws.amazon.com/batch/latest/userguide/specifying-sensitive-data.html)
	// in the Batch User Guide.
	SecretOptions []*Secret `json:"secretOptions,omitempty"`
}

// MountPoint defines the details on a Docker volume mount point that's used in a job's container properties.
// This parameter maps to Volumes in the Create a container (https://docs.docker.com/engine/reference/api/docker_remote_api_v1.19/#create-a-container)
// section of the Docker Remote API and the --volume option to docker run.
type MountPoint struct {
	// The path on the container where the host volume is mounted.
	ContainerPath *string `json:"containerPath,omitempty"`

	// If this value is true, the container has read-only access to the volume.
	// Otherwise, the container can write to the volume. The default value is false.
	ReadOnly *bool `json:"readOnly,omitempty"`

	// The name of the volume to mount.
	SourceVolume *string `json:"sourceVolume,omitempty"`
}

// NetworkConfiguration defines the network configuration for jobs that are running on Fargate resources.
// Jobs that are running on EC2 resources must not specify this parameter.
type NetworkConfiguration struct {
	// Indicates whether the job should have a public IP address. For a job that
	// is running on Fargate resources in a private subnet to send outbound traffic
	// to the internet (for example, to pull container images), the private subnet
	// requires a NAT gateway be attached to route requests to the internet. For
	// more information, see Amazon ECS task networking (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html).
	// The default value is "DISABLED".
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	AssignPublicIP *string `json:"assignPublicIp,omitempty"`
}

// Ulimit defines the ulimit settings to pass to the container.
//
// This object isn't applicable to jobs that are running on Fargate resources.
type Ulimit struct {
	// The hard limit for the ulimit type.
	//
	// HardLimit is a required field
	// +kubebuilder:validation:Required
	HardLimit int64 `json:"hardLimit"`

	// The type of the ulimit.
	//
	// Name is a required field
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The soft limit for the ulimit type.
	//
	// SoftLimit is a required field
	// +kubebuilder:validation:Required
	SoftLimit int64 `json:"softLimit"`
}

// EFSAuthorizationConfig defines the authorization configuration details for the Amazon EFS file system.
type EFSAuthorizationConfig struct {
	// The Amazon EFS access point ID to use. If an access point is specified, the
	// root directory value specified in the EFSVolumeConfiguration must either
	// be omitted or set to / which will enforce the path set on the EFS access
	// point. If an access point is used, transit encryption must be enabled in
	// the EFSVolumeConfiguration. For more information, see Working with Amazon
	// EFS Access Points (https://docs.aws.amazon.com/efs/latest/ug/efs-access-points.html)
	// in the Amazon Elastic File System User Guide.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1.AccessPoint
	// +crossplane:generate:reference:refFieldName=AccessPointIDRef
	// +crossplane:generate:reference:selectorFieldName=AccessPointIDSelector
	AccessPointID *string `json:"accessPointId,omitempty"`

	// AccessPointIDRef are references to AccessPoint used to set
	// the AccessPointID.
	// +optional
	AccessPointIDRef *xpv1.Reference `json:"accessPointIdRef,omitempty"`

	// AccessPointIDSelector selects references to AccessPoint used
	// to set the AccessPointID.
	// +optional
	AccessPointIDSelector *xpv1.Selector `json:"accessPointIdSelector,omitempty"`

	// Whether or not to use the Batch job IAM role defined in a job definition
	// when mounting the Amazon EFS file system. If enabled, transit encryption
	// must be enabled in the EFSVolumeConfiguration. If this parameter is omitted,
	// the default value of DISABLED is used. For more information, see Using Amazon
	// EFS Access Points (https://docs.aws.amazon.com/batch/latest/userguide/efs-volumes.html#efs-volume-accesspoints)
	// in the Batch User Guide. EFS IAM authorization requires that TransitEncryption
	// be ENABLED and that a JobRoleArn is specified.
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	IAM *string `json:"iam,omitempty"`
}

// EFSVolumeConfiguration is used when you're using an Amazon Elastic File System file system
// for job storage. For more information, see Amazon EFS Volumes (https://docs.aws.amazon.com/batch/latest/userguide/efs-volumes.html)
// in the Batch User Guide.
type EFSVolumeConfiguration struct {
	// The authorization configuration details for the Amazon EFS file system.
	AuthorizationConfig *EFSAuthorizationConfig `json:"authorizationConfig,omitempty"`

	// The Amazon EFS file system ID to use.
	//
	// FileSystemID is a required field
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1.FileSystem
	// +crossplane:generate:reference:refFieldName=FileSystemIDRef
	// +crossplane:generate:reference:selectorFieldName=FileSystemIDSelector
	FileSystemID string `json:"fileSystemId,omitempty"`

	// FileSystemIDRef are references to Filesystem used to set
	// the FileSystemID.
	// +optional
	FileSystemIDRef *xpv1.Reference `json:"fileSystemIdRef,omitempty"`

	// FileSystemIDSelector selects references to Filesystem used
	// to set the FileSystemID.
	// +optional
	FileSystemIDSelector *xpv1.Selector `json:"fileSystemIdSelector,omitempty"`

	// The directory within the Amazon EFS file system to mount as the root directory
	// inside the host. If this parameter is omitted, the root of the Amazon EFS
	// volume is used instead. Specifying / has the same effect as omitting this
	// parameter. The maximum length is 4,096 characters.
	//
	// If an EFS access point is specified in the authorizationConfig, the root
	// directory parameter must either be omitted or set to /, which enforces the
	// path set on the Amazon EFS access point.
	RootDirectory *string `json:"rootDirectory,omitempty"`

	// Determines whether to enable encryption for Amazon EFS data in transit between
	// the Amazon ECS host and the Amazon EFS server. Transit encryption must be
	// enabled if Amazon EFS IAM authorization is used. If this parameter is omitted,
	// the default value of DISABLED is used. For more information, see Encrypting
	// data in transit (https://docs.aws.amazon.com/efs/latest/ug/encryption-in-transit.html)
	// in the Amazon Elastic File System User Guide.
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	TransitEncryption *string `json:"transitEncryption,omitempty"`

	// The port to use when sending encrypted data between the Amazon ECS host and
	// the Amazon EFS server. If you don't specify a transit encryption port, it
	// uses the port selection strategy that the Amazon EFS mount helper uses. The
	// value must be between 0 and 65,535. For more information, see EFS Mount Helper
	// (https://docs.aws.amazon.com/efs/latest/ug/efs-mount-helper.html) in the
	// Amazon Elastic File System User Guide.
	TransitEncryptionPort *int64 `json:"transitEncryptionPort,omitempty"`
}

// Host determines whether your data volume persists on the host container instance
// and where it is stored. If this parameter is empty, then the Docker daemon
// assigns a host path for your data volume, but the data isn't guaranteed to
// persist after the containers associated with it stop running.
type Host struct {
	// The path on the host container instance that's presented to the container.
	// If this parameter is empty, then the Docker daemon has assigned a host path
	// for you. If this parameter contains a file location, then the data volume
	// persists at the specified location on the host container instance until you
	// delete it manually. If the source path location doesn't exist on the host
	// container instance, the Docker daemon creates it. If the location does exist,
	// the contents of the source path folder are exported.
	//
	// This parameter isn't applicable to jobs that run on Fargate resources and
	// shouldn't be provided.
	SourcePath *string `json:"sourcePath,omitempty"`
}

// Volume defines a data volume used in a job's container properties.
type Volume struct {
	// This parameter is specified when you are using an Amazon Elastic File System
	// file system for job storage. Jobs that are running on Fargate resources must
	// specify a platformVersion of at least 1.4.0.
	EfsVolumeConfiguration *EFSVolumeConfiguration `json:"efsVolumeConfiguration,omitempty"`

	// The contents of the host parameter determine whether your data volume persists
	// on the host container instance and where it is stored. If the host parameter
	// is empty, then the Docker daemon assigns a host path for your data volume.
	// However, the data isn't guaranteed to persist after the containers associated
	// with it stop running.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	Host *Host `json:"host,omitempty"`

	// The name of the volume. Up to 255 letters (uppercase and lowercase), numbers,
	// hyphens, and underscores are allowed. This name is referenced in the sourceVolume
	// parameter of container definition mountPoints.
	Name *string `json:"name,omitempty"`
}

// ContainerProperties defines the container that's launched as part of a job.
type ContainerProperties struct {
	// The command that's passed to the container. This parameter maps to Cmd in
	// the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the COMMAND parameter to docker run (https://docs.docker.com/engine/reference/run/).
	// For more information, see https://docs.docker.com/engine/reference/builder/#cmd
	// (https://docs.docker.com/engine/reference/builder/#cmd).
	Command []*string `json:"command,omitempty"`

	// The environment variables to pass to a container. This parameter maps to
	// Env in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --env option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// We don't recommend using plaintext environment variables for sensitive information,
	// such as credential data.
	//
	// Environment variables must not start with AWS_BATCH; this naming convention
	// is reserved for variables that are set by the Batch service.
	Environment []*KeyValuePair `json:"environment,omitempty"`

	// The Amazon Resource Name (ARN) of the execution role that Batch can assume.
	// For jobs that run on Fargate resources, you must provide an execution role.
	// For more information, see Batch execution IAM role (https://docs.aws.amazon.com/batch/latest/userguide/execution-IAM-role.html)
	// in the Batch User Guide.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=ExecutionRoleARNRef
	// +crossplane:generate:reference:selectorFieldName=ExecutionRoleARNSelector
	ExecutionRoleArn *string `json:"executionRoleArn,omitempty"`

	// ExecutionRoleARNRef is a reference to an ARN of the IAM role used to set
	// the ExecutionRoleARN.
	// +optional
	ExecutionRoleARNRef *xpv1.Reference `json:"executionRoleARNRef,omitempty"`

	// ExecutionRoleARNSelector selects references to an ARN of the IAM role used
	// to set the ExecutionRoleARN.
	// +optional
	ExecutionRoleARNSelector *xpv1.Selector `json:"executionRoleARNSelector,omitempty"`

	// The platform configuration for jobs that are running on Fargate resources.
	// Jobs that are running on EC2 resources must not specify this parameter.
	FargatePlatformConfiguration *FargatePlatformConfiguration `json:"fargatePlatformConfiguration,omitempty"`

	// The image used to start a container. This string is passed directly to the
	// Docker daemon. Images in the Docker Hub registry are available by default.
	// Other repositories are specified with repository-url/image:tag . Up to 255
	// letters (uppercase and lowercase), numbers, hyphens, underscores, colons,
	// periods, forward slashes, and number signs are allowed. This parameter maps
	// to Image in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the IMAGE parameter of docker run (https://docs.docker.com/engine/reference/run/).
	//
	// Docker image architecture must match the processor architecture of the compute
	// resources that they're scheduled on. For example, ARM-based Docker images
	// can only run on ARM-based compute resources.
	//
	//    * Images in Amazon ECR repositories use the full registry and repository
	//    URI (for example, 012345678910.dkr.ecr.<region-name>.amazonaws.com/<repository-name>).
	//
	//    * Images in official repositories on Docker Hub use a single name (for
	//    example, ubuntu or mongo).
	//
	//    * Images in other repositories on Docker Hub are qualified with an organization
	//    name (for example, amazon/amazon-ecs-agent).
	//
	//    * Images in other online repositories are qualified further by a domain
	//    name (for example, quay.io/assemblyline/ubuntu).
	Image *string `json:"image,omitempty"`

	// The instance type to use for a multi-node parallel job. All node groups in
	// a multi-node parallel job must use the same instance type.
	//
	// This parameter isn't applicable to single-node container jobs or jobs that
	// run on Fargate resources, and shouldn't be provided.
	InstanceType *string `json:"instanceType,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role that the container can assume
	// for Amazon Web Services permissions. For more information, see IAM Roles
	// for Tasks (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
	// in the Amazon Elastic Container Service Developer Guide.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=JobRoleARNRef
	// +crossplane:generate:reference:selectorFieldName=JobRoleARNSelector
	JobRoleArn *string `json:"jobRoleArn,omitempty"`

	// JobRoleARNRef is a reference to an ARN of the IAM role used to set
	// the JobRoleARN.
	// +optional
	JobRoleARNRef *xpv1.Reference `json:"jobRoleARNRef,omitempty"`

	// JobRoleARNSelector selects references to an ARN of the IAM role used
	// to set the JobRoleARN.
	// +optional
	JobRoleARNSelector *xpv1.Selector `json:"jobRoleARNSelector,omitempty"`

	// Linux-specific modifications that are applied to the container, such as details
	// for device mappings.
	LinuxParameters *LinuxParameters `json:"linuxParameters,omitempty"`

	// The log configuration specification for the container.
	//
	// This parameter maps to LogConfig in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --log-driver option to docker run (https://docs.docker.com/engine/reference/run/).
	// By default, containers use the same logging driver that the Docker daemon
	// uses. However the container might use a different logging driver than the
	// Docker daemon by specifying a log driver with this parameter in the container
	// definition. To use a different logging driver for a container, the log system
	// must be configured properly on the container instance (or on a different
	// log server for remote logging options). For more information on the options
	// for different supported log drivers, see Configure logging drivers (https://docs.docker.com/engine/admin/logging/overview/)
	// in the Docker documentation.
	//
	// Batch currently supports a subset of the logging drivers available to the
	// Docker daemon (shown in the LogConfiguration data type).
	//
	// This parameter requires version 1.18 of the Docker Remote API or greater
	// on your container instance. To check the Docker Remote API version on your
	// container instance, log into your container instance and run the following
	// command: sudo docker version | grep "Server API version"
	//
	// The Amazon ECS container agent running on a container instance must register
	// the logging drivers available on that instance with the ECS_AVAILABLE_LOGGING_DRIVERS
	// environment variable before containers placed on that instance can use these
	// log configuration options. For more information, see Amazon ECS Container
	// Agent Configuration (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-agent-config.html)
	// in the Amazon Elastic Container Service Developer Guide.
	LogConfiguration *LogConfiguration `json:"logConfiguration,omitempty"`

	// The mount points for data volumes in your container. This parameter maps
	// to Volumes in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --volume option to docker run (https://docs.docker.com/engine/reference/run/).
	MountPoints []*MountPoint `json:"mountPoints,omitempty"`

	// The network configuration for jobs that are running on Fargate resources.
	// Jobs that are running on EC2 resources must not specify this parameter.
	NetworkConfiguration *NetworkConfiguration `json:"networkConfiguration,omitempty"`

	// When this parameter is true, the container is given elevated permissions
	// on the host container instance (similar to the root user). This parameter
	// maps to Privileged in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --privileged option to docker run (https://docs.docker.com/engine/reference/run/).
	// The default value is false.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided, or specified as false.
	Privileged *bool `json:"privileged,omitempty"`

	// When this parameter is true, the container is given read-only access to its
	// root file system. This parameter maps to ReadonlyRootfs in the Create a container
	// (https://docs.docker.com/engine/api/v1.23/#create-a-container) section of
	// the Docker Remote API (https://docs.docker.com/engine/api/v1.23/) and the
	// --read-only option to docker run.
	ReadonlyRootFilesystem *bool `json:"readonlyRootFilesystem,omitempty"`

	// The type and amount of resources to assign to a container. The supported
	// resources include GPU, MEMORY, and VCPU.
	ResourceRequirements []*ResourceRequirement `json:"resourceRequirements,omitempty"`

	// The secrets for the container. For more information, see Specifying sensitive
	// data (https://docs.aws.amazon.com/batch/latest/userguide/specifying-sensitive-data.html)
	// in the Batch User Guide.
	Secrets []*Secret `json:"secrets,omitempty"`

	// A list of ulimits to set in the container. This parameter maps to Ulimits
	// in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --ulimit option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources
	// and shouldn't be provided.
	Ulimits []*Ulimit `json:"ulimits,omitempty"`

	// The user name to use inside the container. This parameter maps to User in
	// the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --user option to docker run (https://docs.docker.com/engine/reference/run/).
	User *string `json:"user,omitempty"`

	// A list of data volumes used in a job.
	Volumes []*Volume `json:"volumes,omitempty"`
}

// NodeRangeProperty defines the properties of the node range for a multi-node parallel job.
type NodeRangeProperty struct {
	// The container details for the node range.
	Container *ContainerProperties `json:"container,omitempty"`

	// The range of nodes, using node index values. A range of 0:3 indicates nodes
	// with index values of 0 through 3. If the starting range value is omitted
	// (:n), then 0 is used to start the range. If the ending range value is omitted
	// (n:), then the highest possible node index is used to end the range. Your
	// accumulative node ranges must account for all nodes (0:n). You can nest node
	// ranges, for example 0:10 and 4:5, in which case the 4:5 range properties
	// override the 0:10 properties.
	//
	// TargetNodes is a required field
	// +kubebuilder:validation:Required
	TargetNodes string `json:"targetNodes"`
}

// NodeProperties define the node properties of a multi-node parallel job.
type NodeProperties struct {
	// Specifies the node index for the main node of a multi-node parallel job.
	// This node index value must be fewer than the number of nodes.
	//
	// MainNode is a required field
	// +kubebuilder:validation:Required
	MainNode int64 `json:"mainNode"`

	// A list of node ranges and their properties associated with a multi-node parallel
	// job.
	//
	// NodeRangeProperties is a required field
	// +kubebuilder:validation:Required
	NodeRangeProperties []NodeRangeProperty `json:"nodeRangeProperties"`

	// The number of nodes associated with a multi-node parallel job.
	//
	// NumNodes is a required field
	// +kubebuilder:validation:Required
	NumNodes int64 `json:"numNodes"`
}

// JobDefinitionParameters define the desired state of a Batch JobDefinition
type JobDefinitionParameters struct {
	// Region is which region the Function will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// An object with various properties specific to single-node container-based
	// jobs. If the job definition's type parameter is container, then you must
	// specify either containerProperties or nodeProperties.
	//
	// If the job runs on Fargate resources, then you must not specify nodeProperties;
	// use only containerProperties.
	ContainerProperties *ContainerProperties `json:"containerProperties,omitempty"`

	// An object with various properties specific to multi-node parallel jobs.
	//
	// If the job runs on Fargate resources, then you must not specify nodeProperties;
	// use containerProperties instead.
	NodeProperties *NodeProperties `json:"nodeProperties,omitempty"`

	// Default parameter substitution placeholders to set in the job definition.
	// Parameters are specified as a key-value pair mapping. Parameters in a SubmitJob
	// request override any corresponding parameter defaults from the job definition.
	Parameters map[string]*string `json:"parameters,omitempty"`

	// The platform capabilities required by the job definition. If no value is
	// specified, it defaults to EC2. To run the job on Fargate resources, specify
	// FARGATE.
	PlatformCapabilities []*string `json:"platformCapabilities,omitempty"`

	// Specifies whether to propagate the tags from the job or job definition to
	// the corresponding Amazon ECS task. If no value is specified, the tags are
	// not propagated. Tags can only be propagated to the tasks during task creation.
	// For tags with the same name, job tags are given priority over job definitions
	// tags. If the total number of combined tags from the job and job definition
	// is over 50, the job is moved to the FAILED state.
	PropagateTags *bool `json:"propagateTags,omitempty"`

	// The retry strategy to use for failed jobs that are submitted with this job
	// definition. Any retry strategy that's specified during a SubmitJob operation
	// overrides the retry strategy defined here. If a job is terminated due to
	// a timeout, it isn't retried.
	RetryStrategy *RetryStrategy `json:"retryStrategy,omitempty"`

	// The tags that you apply to the job definition to help you categorize and
	// organize your resources. Each tag consists of a key and an optional value.
	// For more information, see Tagging Amazon Web Services Resources (https://docs.aws.amazon.com/batch/latest/userguide/using-tags.html)
	// in Batch User Guide.
	Tags map[string]*string `json:"tags,omitempty"`

	// The timeout configuration for jobs that are submitted with this job definition,
	// after which Batch terminates your jobs if they have not finished. If a job
	// is terminated due to a timeout, it isn't retried. The minimum value for the
	// timeout is 60 seconds. Any timeout configuration that's specified during
	// a SubmitJob operation overrides the timeout configuration defined here. For
	// more information, see Job Timeouts (https://docs.aws.amazon.com/batch/latest/userguide/job_timeouts.html)
	// in the Batch User Guide.
	Timeout *JobTimeout `json:"timeout,omitempty"`

	// The type of job definition. For more information about multi-node parallel
	// jobs, see Creating a multi-node parallel job definition (https://docs.aws.amazon.com/batch/latest/userguide/multi-node-job-def.html)
	// in the Batch User Guide.
	//
	// If the job is run on Fargate resources, then multinode isn't supported.
	//
	// Type is a required field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=container;multinode
	JobDefinitionType string `json:"jobDefinitionType"` // renamed from Type bc json:"type_"
}

// A JobDefinitionSpec defines the desired state of a JobDefinition.
type JobDefinitionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       JobDefinitionParameters `json:"forProvider"`
}

// JobDefinitionObservation keeps the state for the external resource
type JobDefinitionObservation struct {
	// The Amazon Resource Name (ARN) for the job definition.
	JobDefinitionArn *string `json:"jobDefinitionArn,omitempty"`

	// The revision of the job definition.
	Revision *int64 `json:"revision,omitempty"`

	// The status of the job definition.
	Status *string `json:"status,omitempty"`
}

// A JobDefinitionStatus represents the observed state of a JobDefinition.
type JobDefinitionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          JobDefinitionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A JobDefinition is a managed resource that represents an AWS Batch JobDefinition.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type JobDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobDefinitionSpec   `json:"spec"`
	Status JobDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JobDefinitionList contains a list of JobDefinitions
type JobDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobDefinition `json:"items"`
}
