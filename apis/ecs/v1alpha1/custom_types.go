package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomClusterParameters provides custom parameters for the Cluster type
type CustomClusterParameters struct{}

// CustomClusterObservation includes the custom status fields of Cluster.
type CustomClusterObservation struct{}

// CustomAWSVPCConfiguration provides custom parameters for the
// AWSVPCConfiguration type
type CustomAWSVPCConfiguration struct {
	AssignPublicIP *string `json:"assignPublicIP,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupSelector
	SecurityGroups        []*string        `json:"securityGroups,omitempty"`
	SecurityGroupRefs     []xpv1.Reference `json:"securityGroupRefs,omitempty"`
	SecurityGroupSelector *xpv1.Selector   `json:"securityGroupSelector,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetSelector
	Subnets        []*string        `json:"subnets,omitempty"`
	SubnetRefs     []xpv1.Reference `json:"subnetRefs,omitempty"`
	SubnetSelector *xpv1.Selector   `json:"subnetSelector,omitempty"`
}

// CustomLoadBalancer provides custom parameters for the LoadBalancer type
type CustomLoadBalancer struct {
	// The name of the container (as it appears in a container definition) to associate
	// with the load balancer.
	ContainerName *string `json:"containerName,omitempty"`

	// The port on the container to associate with the load balancer. This port
	// must correspond to a containerPort in the task definition the tasks in the
	// service are using. For tasks that use the EC2 launch type, the container
	// instance they're launched on must allow ingress traffic on the hostPort of
	// the port mapping.
	ContainerPort *int64 `json:"containerPort,omitempty"`

	// The name of the load balancer to associate with the Amazon ECS service or
	// task set.
	//
	// A load balancer name is only specified when using a Classic Load Balancer.
	// If you are using an Application Load Balancer or a Network Load Balancer
	// the load balancer name parameter should be omitted.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1.LoadBalancer
	// +crossplane:generate:reference:extractor=LoadBalancerName()
	LoadBalancerName         *string         `json:"loadBalancerName,omitempty"`
	LoadBalancerNameRef      *xpv1.Reference `json:"loadBalancerNameRef,omitempty"`
	LoadBalancerNameSelector *xpv1.Selector  `json:"loadBalancerNameSelector,omitempty"`

	// The full Amazon Resource Name (ARN) of the Elastic Load Balancing target
	// group or groups associated with a service or task set.
	//
	// A target group ARN is only specified when using an Application Load Balancer
	// or Network Load Balancer. If you're using a Classic Load Balancer, omit the
	// target group ARN.
	//
	// For services using the ECS deployment controller, you can specify one or
	// multiple target groups. For more information, see Registering multiple target
	// groups with a service (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/register-multiple-targetgroups.html)
	// in the Amazon Elastic Container Service Developer Guide.
	//
	// For services using the CODE_DEPLOY deployment controller, you're required
	// to define two target groups for the load balancer. For more information,
	// see Blue/green deployment with CodeDeploy (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-type-bluegreen.html)
	// in the Amazon Elastic Container Service Developer Guide.
	//
	// If your service's task definition uses the awsvpc network mode, you must
	// choose ip as the target type, not instance. Do this when creating your target
	// groups because tasks that use the awsvpc network mode are associated with
	// an elastic network interface, not an Amazon EC2 instance. This network mode
	// is required for the Fargate launch type.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1.TargetGroup
	TargetGroupARN         *string         `json:"targetGroupARN,omitempty"`
	TargetGroupARNRef      *xpv1.Reference `json:"targetGroupARNRef,omitempty"`
	TargetGroupARNSelector *xpv1.Selector  `json:"targetGroupARNSelector,omitempty"`
}

// CustomNetworkConfiguration provides custom parameters for the
// NetworkConfiguration type
type CustomNetworkConfiguration struct {
	// An object representing the networking details for a task or service.
	AWSvpcConfiguration *CustomAWSVPCConfiguration `json:"awsvpcConfiguration,omitempty"`
}

// CustomServiceParameters provides custom parameters for the Service type
type CustomServiceParameters struct {
	// The short name or full Amazon Resource Name (ARN) of the cluster on which
	// to run your service. If you do not specify a cluster, the default cluster
	// is assumed.
	// +immutable
	// +crossplane:generate:reference:type=Cluster
	Cluster         *string         `json:"cluster,omitempty"`
	ClusterRef      *xpv1.Reference `json:"clusterRef,omitempty"`
	ClusterSelector *xpv1.Selector  `json:"clusterSelector,omitempty"`

	// Force Service to be deleted, even with task Running or Pending
	// +optional
	ForceDeletion bool `json:"forceDeletion,omitempty"`

	// A load balancer object representing the load balancers to use with your service.
	// For more information, see Service Load Balancing (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-load-balancing.html)
	// in the Amazon Elastic Container Service Developer Guide.
	//
	// If the service is using the rolling update (ECS) deployment controller and
	// using either an Application Load Balancer or Network Load Balancer, you must
	// specify one or more target group ARNs to attach to the service. The service-linked
	// role is required for services that make use of multiple target groups. For
	// more information, see Using service-linked roles for Amazon ECS (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/using-service-linked-roles.html)
	// in the Amazon Elastic Container Service Developer Guide.
	//
	// If the service is using the CODE_DEPLOY deployment controller, the service
	// is required to use either an Application Load Balancer or Network Load Balancer.
	// When creating an CodeDeploy deployment group, you specify two target groups
	// (referred to as a targetGroupPair). During a deployment, CodeDeploy determines
	// which task set in your service has the status PRIMARY and associates one
	// target group with it, and then associates the other target group with the
	// replacement task set. The load balancer can also have up to two listeners:
	// a required listener for production traffic and an optional listener that
	// allows you perform validation tests with Lambda functions before routing
	// production traffic to it.
	//
	// After you create a service using the ECS deployment controller, the load
	// balancer name or target group ARN, container name, and container port specified
	// in the service definition are immutable. If you are using the CODE_DEPLOY
	// deployment controller, these values can be changed when updating the service.
	//
	// For Application Load Balancers and Network Load Balancers, this object must
	// contain the load balancer target group ARN, the container name (as it appears
	// in a container definition), and the container port to access from the load
	// balancer. The load balancer name parameter must be omitted. When a task from
	// this service is placed on a container instance, the container instance and
	// port combination is registered as a target in the target group specified
	// here.
	//
	// For Classic Load Balancers, this object must contain the load balancer name,
	// the container name (as it appears in a container definition), and the container
	// port to access from the load balancer. The target group ARN parameter must
	// be omitted. When a task from this service is placed on a container instance,
	// the container instance is registered with the load balancer specified here.
	//
	// Services with tasks that use the awsvpc network mode (for example, those
	// with the Fargate launch type) only support Application Load Balancers and
	// Network Load Balancers. Classic Load Balancers are not supported. Also, when
	// you create any target groups for these services, you must choose ip as the
	// target type, not instance, because tasks that use the awsvpc network mode
	// are associated with an elastic network interface, not an Amazon EC2 instance.
	LoadBalancers []*CustomLoadBalancer `json:"loadBalancers,omitempty"`

	// The network configuration for the service. This parameter is required for
	// task definitions that use the awsvpc network mode to receive their own elastic
	// network interface, and it is not supported for other network modes. For more
	// information, see Task networking (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html)
	// in the Amazon Elastic Container Service Developer Guide.
	NetworkConfiguration *CustomNetworkConfiguration `json:"networkConfiguration,omitempty"`

	// The family and revision (family:revision) or full ARN of the task definition
	// to run in your service. If a revision is not specified, the latest ACTIVE
	// revision is used.
	//
	// A task definition must be specified if the service is using either the ECS
	// or CODE_DEPLOY deployment controllers.
	// +optional
	// +crossplane:generate:reference:type=TaskDefinition
	TaskDefinition         *string         `json:"taskDefinition,omitempty"`
	TaskDefinitionRef      *xpv1.Reference `json:"taskDefinitionRef,omitempty"`
	TaskDefinitionSelector *xpv1.Selector  `json:"taskDefinitionSelector,omitempty"`
}

// CustomServiceObservation includes the custom status fields of Service.
type CustomServiceObservation struct{}

// CustomEFSAuthorizationConfig provides custom parameters for the
// EFSAuthorizationConfig type
type CustomEFSAuthorizationConfig struct {
	// The Amazon EFS access point ID to use. If an access point is specified, the
	// root directory value specified in the EFSVolumeConfiguration must either
	// be omitted or set to / which will enforce the path set on the EFS access
	// point. If an access point is used, transit encryption must be enabled in
	// the EFSVolumeConfiguration. For more information, see Working with Amazon
	// EFS Access Points (https://docs.aws.amazon.com/efs/latest/ug/efs-access-points.html)
	// in the Amazon Elastic File System User Guide.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1.AccessPoint
	AccessPointID         *string         `json:"accessPointID,omitempty"`
	AccessPointIDRef      *xpv1.Reference `json:"accessPointIDRef,omitempty"`
	AccessPointIDSelector *xpv1.Selector  `json:"accessPointIDSelector,omitempty"`

	// Determines whether to use the Amazon ECS task IAM role defined in a task
	// definition when mounting the Amazon EFS file system. If enabled, transit
	// encryption must be enabled in the EFSVolumeConfiguration. If this parameter
	// is omitted, the default value of DISABLED is used. For more information,
	// see Using Amazon EFS Access Points (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/efs-volumes.html#efs-volume-accesspoints)
	// in the Amazon Elastic Container Service Developer Guide.
	IAM *string `json:"iam,omitempty"`
}

// CustomEFSVolumeConfiguration provides custom parameters for the
// EFSVolumeConfiguration type
type CustomEFSVolumeConfiguration struct {

	// The authorization configuration details for the Amazon EFS file system.
	AuthorizationConfig *CustomEFSAuthorizationConfig `json:"authorizationConfig,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1.FileSystem
	FileSystemID         *string         `json:"fileSystemID,omitempty"`
	FileSystemIDRef      *xpv1.Reference `json:"fileSystemIDRef,omitempty"`
	FileSystemIDSelector *xpv1.Selector  `json:"fileSystemIDSelector,omitempty"`

	RootDirectory         *string `json:"rootDirectory,omitempty"`
	TransitEncryption     *string `json:"transitEncryption,omitempty"`
	TransitEncryptionPort *int64  `json:"transitEncryptionPort,omitempty"`
}

// CustomVolume provides custom parameters for the Volume type
type CustomVolume struct {
	// This parameter is specified when you are using Docker volumes. Docker volumes
	// are only supported when you are using the EC2 launch type. Windows containers
	// only support the use of the local driver. To use bind mounts, specify a host
	// instead.
	DockerVolumeConfiguration *DockerVolumeConfiguration `json:"dockerVolumeConfiguration,omitempty"`
	// This parameter is specified when you are using an Amazon Elastic File System
	// file system for task storage. For more information, see Amazon EFS Volumes
	// (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/efs-volumes.html)
	// in the Amazon Elastic Container Service Developer Guide.
	EFSVolumeConfiguration *CustomEFSVolumeConfiguration `json:"efsVolumeConfiguration,omitempty"`
	// This parameter is specified when you are using Amazon FSx for Windows File
	// Server (https://docs.aws.amazon.com/fsx/latest/WindowsGuide/what-is.html)
	// file system for task storage.
	//
	// For more information and the input format, see Amazon FSx for Windows File
	// Server Volumes (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/wfsx-volumes.html)
	// in the Amazon Elastic Container Service Developer Guide.
	FsxWindowsFileServerVolumeConfiguration *FSxWindowsFileServerVolumeConfiguration `json:"fsxWindowsFileServerVolumeConfiguration,omitempty"`
	// Details on a container instance bind mount host volume.
	Host *HostVolumeProperties `json:"host,omitempty"`

	// +kubebuilder:validation:Required
	Name *string `json:"name"`
}

// CustomTaskDefinitionParameters provides custom parameters for the
// TaskDefinition type
type CustomTaskDefinitionParameters struct {
	// The Amazon Resource Name (ARN) of the task execution role that grants the
	// Amazon ECS container agent permission to make Amazon Web Services API calls
	// on your behalf. The task execution IAM role is required depending on the
	// requirements of your task. For more information, see Amazon ECS task execution
	// IAM role (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html)
	// in the Amazon Elastic Container Service Developer Guide.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	ExecutionRoleARN         *string         `json:"executionRoleARN,omitempty"`
	ExecutionRoleARNRef      *xpv1.Reference `json:"executionRoleARNRef,omitempty"`
	ExecutionRoleARNSelector *xpv1.Selector  `json:"executionRoleARNSelector,omitempty"`

	// The short name or full Amazon Resource Name (ARN) of the IAM role that containers
	// in this task can assume. All containers in this task are granted the permissions
	// that are specified in this role. For more information, see IAM Roles for
	// Tasks (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
	// in the Amazon Elastic Container Service Developer Guide.
	// A list of volume definitions in JSON format that containers in your task
	// may use.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	TaskRoleARN         *string         `json:"taskRoleARN,omitempty"`
	TaskRoleARNRef      *xpv1.Reference `json:"taskRoleARNRef,omitempty"`
	TaskRoleARNSelector *xpv1.Selector  `json:"taskRoleARNSelector,omitempty"`

	Volumes []*CustomVolume `json:"volumes,omitempty"`
}

// CustomTaskDefinitionObservation includes the custom status fields of TaskDefinition.
type CustomTaskDefinitionObservation struct{}
