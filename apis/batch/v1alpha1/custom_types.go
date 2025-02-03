package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomComputeEnvironmentParameters includes custom additional fields for ComputeEnvironment
type CustomComputeEnvironmentParameters struct {
	// The full Amazon Resource Name (ARN) of the IAM role that allows Batch to
	// make calls to other Amazon Web Services services on your behalf. For more
	// information, see Batch service IAM role (https://docs.aws.amazon.com/batch/latest/userguide/service_IAM_role.html)
	// If the compute environment has a service-linked role, it can't be changed to use a regular IAM role.
	// Likewise, if the compute environment has a regular IAM role, it can't be changed to use a service-linked role.
	// If your specified role has a path other than /, then you must either specify the full role ARN (this is recommended)
	// or prefix the role name with the path.
	// Depending on how you created your Batch service role, its ARN might contain the service-role path prefix.
	// When you only specify the name of the service role, Batch assumes that your ARN doesn't use the service-role path prefix.
	// Because of this, we recommend that you specify the full ARN of your service role when you create compute environments
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=ServiceRoleARNRef
	// +crossplane:generate:reference:selectorFieldName=ServiceRoleARNSelector
	ServiceRoleARN *string `json:"serviceRoleARN,omitempty"`

	// ServiceRoleARNRef is a reference to an ARN of the IAM role used to set
	// the ServiceRoleARN.
	// +optional
	ServiceRoleARNRef *xpv1.Reference `json:"serviceRoleARNRef,omitempty"`

	// ServiceRoleARNSelector selects references to an ARN of the IAM role used
	// to set the ServiceRoleARN.
	// +optional
	ServiceRoleARNSelector *xpv1.Selector `json:"serviceRoleARNSelector,omitempty"`

	// Custom parameter to control the state of the compute environment. The valid values are ENABLED or DISABLED.
	//
	// If the state is ENABLED, then the Batch scheduler can attempt to place jobs
	// from an associated job queue on the compute resources within the environment.
	// If the compute environment is managed, then it can scale its instances out
	// or in automatically, based on the job queue demand.
	//
	// If the state is DISABLED, then the Batch scheduler doesn't attempt to place
	// jobs within the environment. Jobs in a STARTING or RUNNING state continue
	// to progress normally. Managed compute environments in the DISABLED state
	// don't scale out. However, they scale in to minvCpus value after instances
	// become idle.
	// +optional
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	DesiredState *string `json:"desiredState,omitempty"`

	// The VPC subnets where the compute resources are launched. These subnets must
	// be within the same VPC. Fargate compute resources can contain up to 16 subnets.
	// For more information, see VPCs and Subnets (https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Subnets.html)
	// in the Amazon VPC User Guide.
	// (Subnets is originally a field of ComputeResources)
	// Subnets is a required field for CE type MANAGED.
	// For a MANGED CE of type EC2 or SPOT to be able to update this field
	// Allocation Strategy BEST_FIT_PROGRESSIVE or SPOT_CAPACITY_OPTIMIZED is required.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []*string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to SubnetIDs used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects references to SubnetIDs used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// The Amazon EC2 security groups associated with instances launched in the
	// compute environment. One or more security groups must be specified, either
	// in securityGroupIds or using a launch template referenced in launchTemplate.
	// This parameter is required for jobs that are running on Fargate resources
	// and must contain at least one security group. Fargate doesn't support launch
	// templates. If security groups are specified using both securityGroupIds and
	// launchTemplate, the values in securityGroupIds are used.
	// For a MANGED CE of type EC2 or SPOT to be able to update this field
	// Allocation Strategy BEST_FIT_PROGRESSIVE or SPOT_CAPACITY_OPTIMIZED is required.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []*string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// The Amazon ECS instance profile applied to Amazon EC2 instances in a compute
	// environment. You can specify the short name or full Amazon Resource Name
	// (ARN) of an instance profile. For example, ecsInstanceRole or arn:aws:iam::<aws_account_id>:instance-profile/ecsInstanceRole
	// . For more information, see Amazon ECS Instance Role (https://docs.aws.amazon.com/batch/latest/userguide/instance_IAM_role.html)
	// in the Batch User Guide.
	// Only applicable to MANGED CE of type EC2 or SPOT.
	// This field can be updated for CE only
	// with Allocation Strategy BEST_FIT_PROGRESSIVE or SPOT_CAPACITY_OPTIMIZED.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources,
	// and shouldn't be specified.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1alpha1.InstanceProfile
	// +crossplane:generate:reference:refFieldName=InstanceRoleRef
	// +crossplane:generate:reference:selectorFieldName=InstanceRoleSelector
	InstanceRole *string `json:"instanceRole,omitempty"`

	// InstanceRoleRef is a reference to the IAM InstanceProfile used to set
	// the InstanceRole.
	// +optional
	InstanceRoleRef *xpv1.Reference `json:"instanceRoleRef,omitempty"`

	// InstanceRoleSelector selects references to the IAM InstanceProfile used
	// to set the InstanceRole.
	// +optional
	InstanceRoleSelector *xpv1.Selector `json:"instanceRoleSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon EC2 Spot Fleet IAM role applied
	// to a SPOT compute environment. This role is required if the allocation strategy
	// set to BEST_FIT or if the allocation strategy isn't specified. For more information,
	// see Amazon EC2 Spot Fleet Role (https://docs.aws.amazon.com/batch/latest/userguide/spot_fleet_IAM_role.html)
	// in the Batch User Guide.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources,
	// and shouldn't be specified.
	//
	// To tag your Spot Instances on creation, the Spot Fleet IAM role specified
	// here must use the newer AmazonEC2SpotFleetTaggingRole managed policy. The
	// previously recommended AmazonEC2SpotFleetRole managed policy doesn't have
	// the required permissions to tag Spot Instances. For more information, see
	// Spot Instances not tagged on creation (https://docs.aws.amazon.com/batch/latest/userguide/troubleshooting.html#spot-instance-no-tag)
	// in the Batch User Guide.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=SpotIAMFleetRoleRef
	// +crossplane:generate:reference:selectorFieldName=SpotIAMFleetRoleSelector
	SpotIAMFleetRole *string `json:"spotIamFleetRole,omitempty"`

	// SpotIAMFleetRoleRef is a reference to an ARN of the IAM role used to set
	// the SpotIAMFleetRole.
	// +optional
	SpotIAMFleetRoleRef *xpv1.Reference `json:"spotIAMFleetRoleRef,omitempty"`

	// SpotIAMFleetRoleSelector selects references to an ARN of the IAM role used
	// to set the SpotIAMFleetRole.
	// +optional
	SpotIAMFleetRoleSelector *xpv1.Selector `json:"spotIamFleetRoleSelector,omitempty"`

	// Specifies the infrastructure update policy for the compute environment. For
	// more information about infrastructure updates, see Updating compute environments
	// (https://docs.aws.amazon.com/batch/latest/userguide/updating-compute-environments.html)
	// in the Batch User Guide.
	// Only applicable to MANGED CE of type EC2 or SPOT.
	// This field requires an update request to be set and it can be updated for CE only
	// with Allocation Strategy BEST_FIT_PROGRESSIVE or SPOT_CAPACITY_OPTIMIZED.
	//
	// JobExecutionTimeoutMinutes specifies the job timeout (in minutes) when the compute environment
	// infrastructure is updated. The default value is 30.
	//
	// TerminateJobsOnUpdate specifies whether jobs are automatically terminated when the computer
	// environment infrastructure is updated. The default value is false.
	UpdatePolicy *UpdatePolicy `json:"updatePolicy,omitempty"`

	// Specifies whether the AMI ID is updated to the latest one that's supported
	// by Batch when the compute environment has an infrastructure update.
	// The default value is false.
	// Only applicable to MANGED CE of type EC2 or SPOT.
	// This field requires an update request to be set and it can be updated for CE only
	// with Allocation Strategy BEST_FIT_PROGRESSIVE or SPOT_CAPACITY_OPTIMIZED.
	// Also to get this field changed, you need to include another change to trigger an update.
	//
	// If an AMI ID is specified in the imageIdOverride parameters or
	// by the launch template specified in the launchTemplate parameter, this parameter
	// is ignored. For more information on updating AMI IDs during an infrastructure
	// update, see Updating the AMI ID (https://docs.aws.amazon.com/batch/latest/userguide/updating-compute-environments.html#updating-compute-environments-ami)
	// in the Batch User Guide.
	//
	// When updating a compute environment, changing this setting requires an infrastructure
	// update of the compute environment. For more information, see Updating compute
	// environments (https://docs.aws.amazon.com/batch/latest/userguide/updating-compute-environments.html)
	// in the Batch User Guide.
	UpdateToLatestImageVersion *bool `json:"updateToLatestImageVersion,omitempty"`
}

// CustomComputeEnvironmentObservation includes custom additional status fields for ComputeEnvironment
type CustomComputeEnvironmentObservation struct{}

// CustomJobQueueParameters includes custom additional fields for JobQueue
type CustomJobQueueParameters struct {
	// Custom parameter to control the state of the job queue. The valid values are ENABLED or DISABLED.
	//
	// The state of the job queue. If the job queue state is ENABLED, it is able to accept jobs.
	// If the job queue state is DISABLED, new jobs can't be added to the queue, but jobs already in the queue can finish.
	// +optional
	// +kubebuilder:validation:Enum=ENABLED;DISABLED
	DesiredState *string `json:"desiredState,omitempty"`

	// The set of compute environments mapped to a job queue and their order relative
	// to each other. The job scheduler uses this parameter to determine which compute
	// environment should run a specific job. Compute environments must be in the
	// VALID state before you can associate them with a job queue. You can associate
	// up to three compute environments with a job queue. All of the compute environments
	// must be either EC2 (EC2 or SPOT) or Fargate (FARGATE or FARGATE_SPOT); EC2
	// and Fargate compute environments can't be mixed.
	//
	// All compute environments that are associated with a job queue must share
	// the same architecture. Batch doesn't support mixing compute environment architecture
	// types in a single job queue.
	//
	// ComputeEnvironmentOrder is a required field
	// +kubebuilder:validation:Required
	ComputeEnvironmentOrder []CustomComputeEnvironmentOrder `json:"computeEnvironmentOrder"`
}

// CustomJobQueueObservation includes custom additional status fields for JobQueue
type CustomJobQueueObservation struct{}

// CustomComputeEnvironmentOrder includes custom additional fields for ComputeEnvironmentOrder
type CustomComputeEnvironmentOrder struct {
	// The Amazon Resource Name (ARN) of the compute environment.
	//
	// ComputeEnvironment is a required field
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/batch/v1alpha1.ComputeEnvironment
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/batch/v1alpha1.ComputeEnvironmentARN()
	// +crossplane:generate:reference:refFieldName=ComputeEnvironmentRef
	// +crossplane:generate:reference:selectorFieldName=ComputeEnvironmentSelector
	ComputeEnvironment string `json:"computeEnvironment,omitempty"`

	// ComputeEnvironmentRef is a reference to ComputeEnvironment used to set
	// the ComputeEnvironment.
	// +optional
	ComputeEnvironmentRef *xpv1.Reference `json:"computeEnvironmentRef,omitempty"`

	// ComputeEnvironmentsSelector selects a reference to ComputeEnvironment used
	// to set the ComputeEnvironment.
	// +optional
	ComputeEnvironmentSelector *xpv1.Selector `json:"computeEnvironmentSelector,omitempty"`

	// The order of the compute environment. Compute environments are tried in ascending
	// order. For example, if two compute environments are associated with a job
	// queue, the compute environment with a lower order integer value is tried
	// for job placement first.
	//
	// Order is a required field
	// +kubebuilder:validation:Required
	Order int64 `json:"order"`
}
