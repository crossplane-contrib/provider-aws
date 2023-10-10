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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Redshift cluster states.
const (
	// The cluster is healthy and available
	StateAvailable = "available"
	// The cluster is being created. The cluster is inaccessible while it is being created.
	StateCreating = "creating"
	// The cluster is being deleted.
	StateDeleting = "deleting"
	// The cluster is being modified.
	StateModifying = "modifying"
	// The cluster has failed and Amazon Redshift can't recover it. Perform a point-in-time restore to the latest restorable time of the Cluster to recover the data.
	StateFailed = "failed"
)

// ClusterParameters define the parameters available for an AWS Redshift cluster
type ClusterParameters struct {
	// Region is the region you'd like the Cluster to be created in.
	Region string `json:"region"`

	// NodeType is the node type defining its size and compute capacity to be
	// provisioned for the cluster. For information about node types,
	// go to Working with Clusters (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#how-many-nodes)
	// in the Amazon Redshift Cluster Management Guide.
	NodeType string `json:"nodeType"`

	// MasterUsername is the user name associated with the master user account for the cluster that
	// is being created.
	// Constraints:
	//    * Must be 1 - 128 alphanumeric characters. The user name can't be PUBLIC.
	//    * First character must be a letter.
	//    * Cannot be a reserved word. A list of reserved words can be found in
	//    Reserved Words (https://docs.aws.amazon.com/redshift/latest/dg/r_pg_keywords.html)
	//    in the Amazon Redshift Database Developer Guide.
	// +immutable
	MasterUsername string `json:"masterUsername"`

	// AllowVersionUpgrade indicates that major engine upgrades are applied automatically to the
	// cluster during the maintenance window.
	// default=true
	// +optional
	AllowVersionUpgrade *bool `json:"allowVersionUpgrade,omitempty"`

	// AutomatedSnapshotRetentionPeriod is the number of days for which
	// automated backups are retained. Setting this parameter to a positive
	// number enables backups. Setting this parameter to  0 disables automated backups.
	// default=1
	// +kubebuilder:validation:Maximum=35
	// +kubebuilder:validation:Minimum=0
	// +optional
	AutomatedSnapshotRetentionPeriod *int32 `json:"automatedSnapshotRetentionPeriod,omitempty"`

	// AvailabilityZone is the EC2 Availability Zone in which you want
	// Amazon Redshift to provision the cluster.
	// Default: A random, system-chosen Availability Zone in the region that is
	// specified by the endpoint.
	// Example: us-east-2d
	// Constraint: The specified Availability Zone must be in the same region as
	// the current endpoint. The Availability Zone parameter can't be specified
	// if the MultiAZ parameter is set to true.
	// The specified Availability Zone must be in the same AWS Region as the current endpoint.
	// +optional
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// ClusterParameterGroupName is the name of the cluster parameter group to use for the cluster.
	// Default: The default Amazon Redshift cluster parameter group. For information
	// about the default parameter group, go to Working with Amazon Redshift Parameter
	// Groups (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-parameter-groups.html)
	// +optional
	ClusterParameterGroupName *string `json:"clusterParameterGroupName,omitempty"`

	// SecurityGroups is a list of security groups to associate with this cluster.
	// Default: The default cluster security group for Amazon Redshift.
	// +optional
	ClusterSecurityGroups []string `json:"clusterSecurityGroups,omitempty"`

	// ClusterSecurityGroupRefs are references to ClusterSecurityGroups used to set
	// the ClusterSecurityGroups.
	// +immutable
	// +optional
	ClusterSecurityGroupRefs []xpv1.Reference `json:"clusterSecurityGroupRefs,omitempty"`

	// ClusterSecurityGroupSelector selects references to ClusterSecurityGroups used
	// to set the ClusterSecurityGroups.
	// +immutable
	// +optional
	ClusterSecurityGroupSelector *xpv1.Selector `json:"clusterSecurityGroupSelector,omitempty"`

	// ClusterSubnetGroupName is the name of a cluster subnet group to be associated with this cluster.
	// If this parameter is not provided the resulting cluster will be deployed
	// outside virtual private cloud (VPC).
	// +optional
	ClusterSubnetGroupName *string `json:"clusterSubnetGroupName,omitempty"`

	// ClusterType is the type of the cluster you want.
	// When cluster type is specified as
	//    * single-node, the NumberOfNodes parameter is not required.
	//    * multi-node, the NumberOfNodes parameter is required.
	// default=multi-node
	// +kubebuilder:validation:Enum=multi-node;single-node
	// +optional
	ClusterType *string `json:"clusterType,omitempty"`

	// ClusterVersion is the version of the Amazon Redshift engine software
	// that you want to deploy on the cluster. The version selected runs on all the nodes in the cluster.
	// Constraints: Only version 1.0 is currently available.
	// +optional
	ClusterVersion *string `json:"clusterVersion,omitempty"`

	// DBName is the name of the first database to be created when the cluster is created.
	// To create additional databases after the cluster is created, connect to the
	// cluster with a SQL client and use SQL commands to create a database. For
	// more information, go to Create a Database (https://docs.aws.amazon.com/redshift/latest/dg/t_creating_database.html)
	// in the Amazon Redshift Database Developer Guide.
	// Constraints:
	//    * Must contain 1 to 64 alphanumeric characters.
	//    * Must contain only lowercase letters.
	//    * Cannot be a word that is reserved by the service. A list of reserved
	//    words can be found in Reserved Words (https://docs.aws.amazon.com/redshift/latest/dg/r_pg_keywords.html)
	//    in the Amazon Redshift Database Developer Guide.
	// default=dev
	// +optional
	DBName *string `json:"dbName,omitempty"`

	// The Elastic IP (EIP) address for the cluster.
	// Constraints: The cluster must be provisioned in EC2-VPC and publicly-accessible
	// through an Internet gateway. For more information about provisioning clusters
	// in EC2-VPC, go to Supported Platforms to Launch Your Cluster (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#cluster-platforms)
	// in the Amazon Redshift Cluster Management Guide.
	// +optional
	ElasticIP *string `json:"elasticIP,omitempty"`

	// Encrypted defines whether your data in the cluster will be encrypted at rest or not.
	// default=false
	// +optional
	Encrypted *bool `json:"encrypted,omitempty"`

	// EnhancedVPCRouting specifies whether to create the cluster with enhanced VPC
	// routing enabled. To create a cluster that uses enhanced VPC routing, the
	// cluster must be in a VPC. For more information, see Enhanced VPC Routing
	// (https://docs.aws.amazon.com/redshift/latest/mgmt/enhanced-vpc-routing.html)
	// in the Amazon Redshift Cluster Management Guide.
	// If this option is true, enhanced VPC routing is enabled.
	// default=false
	// +optional
	EnhancedVPCRouting *bool `json:"enhancedVPCRouting,omitempty"`

	// FinalClusterSnapshotIdentifier is the identifier of the final snapshot
	// that is to be created immediately before deleting the cluster.
	// If this parameter is provided, SkipFinalClusterSnapshot must be false.
	// Constraints:
	//    * Must be 1 to 255 alphanumeric characters.
	//    * First character must be a letter.
	//    * Cannot end with a hyphen or contain two consecutive hyphens.
	// +optional
	FinalClusterSnapshotIdentifier *string `json:"finalClusterSnapshotIdentifier,omitempty"`

	// FinalClusterSnapshotRetentionPeriod is the number of days that
	// a manual snapshot is retained.
	// If the value is -1, the manual snapshot is retained indefinitely.
	// The value must be either -1 or an integer between 1 and 3,653.
	// default -1
	// +optional
	FinalClusterSnapshotRetentionPeriod *int32 `json:"finalClusterSnapshotRetentionPeriod,omitempty"`

	// HSMClientCertificateIdentifier specifies the name of the HSM client certificate
	// the Amazon Redshift cluster uses to retrieve the data encryption keys stored in an HSM.
	// +optional
	HSMClientCertificateIdentifier *string `json:"hsmClientCertificateIdentifier,omitempty"`

	// HSMConfigurationIdentifier specifies the name of the HSM configuration that
	// contains the information the Amazon Redshift cluster can use to retrieve
	// and store keys in an HSM.
	// +optional
	HSMConfigurationIdentifier *string `json:"hsmConfigurationIdentifier,omitempty"`

	// IAMRoles is a list of AWS Identity and Access Management (IAM) roles that can be used
	// by the cluster to access other AWS services. You must supply the IAM roles
	// in their Amazon Resource Name (ARN) format. You can supply up to 10 IAM roles
	// in a single request.
	// A cluster can have up to 10 IAM roles associated with it at any time.
	// kubebuilder:validation:MaxItems=10
	// +optional
	IAMRoles []string `json:"iamRoles,omitempty"`

	// IAMRoleRefs are references to IAMRoles used to set
	// the IAMRoles.
	// +immutable
	// +optional
	IAMRoleRefs []xpv1.Reference `json:"iamRoleRefs,omitempty"`

	// IAMRoleSelector selects references to IAMRoles used
	// to set the IAMRoles.
	// +immutable
	// +optional
	IAMRoleSelector *xpv1.Selector `json:"iamRoleSelector,omitempty"`

	// KMSKeyID is the Amazon Resource Name (ARN) for the KMS encryption
	// key. If you are creating a cluster with the same AWS account that owns
	// the KMS encryption key used to encrypt the new cluster, then you can
	// use the KMS key alias instead of the ARN for the KM encryption key.
	// +optional
	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	// MaintenanceTrackName an optional parameter for the name of the maintenance track for the cluster.
	// +optional
	MaintenanceTrackName *string `json:"maintenanceTrackName,omitempty"`

	// ManualSnapshotRetentionPeriod is the default number of days to retain a manual snapshot.
	// If the value is -1, the snapshot is retained indefinitely.
	// This setting doesn't change the retention period of existing snapshots.
	// default=1
	// +kubebuilder:validation:Maximum=3653
	// +optional
	ManualSnapshotRetentionPeriod *int32 `json:"manualSnapshotRetentionPeriod,omitempty"`

	// NewMasterUserPassword is the new password to be associated with the master user account
	// for the cluster that has being created.
	// Set this value if you want to change the existing password of the cluster.
	// Constraints:
	//    * Must be between 8 and 64 characters in length.
	//    * Must contain at least one uppercase letter.
	//    * Must contain at least one lowercase letter.
	//    * Must contain one number.
	//    * Can be any printable ASCII character (ASCII code 33 to 126) except '
	//    (single quote), " (double quote), \, /, @, or space.
	// +optional
	NewMasterUserPassword *string `json:"newMasterUserPassword,omitempty"`

	// NewClusterIdentifier is the new identifier you want to use for the cluster.
	// +optional
	NewClusterIdentifier *string `json:"newClusterIdentifier,omitempty"`

	// NumberOfNodes defines the number of compute nodes in the cluster.
	// This parameter is required when the ClusterType parameter is specified as multi-node.
	// For information about determining how many nodes you need, go to Working
	// with Clusters (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#how-many-nodes)
	// in the Amazon Redshift Cluster Management Guide.
	// If you don't specify this parameter, you get a single-node cluster. When
	// requesting a multi-node cluster, you must specify the number of nodes that
	// you want in the cluster.
	// default=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=1
	// +optional
	NumberOfNodes *int32 `json:"numberOfNodes,omitempty"`

	// Port specifies the port number on which the cluster accepts incoming connections.
	// The cluster is accessible only via the JDBC and ODBC connection strings.
	// Part of the connection string requires the port on which the cluster will
	// listen for incoming connections.
	// default=5439
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1150
	// +optional
	Port *int32 `json:"port,omitempty"`

	// PreferredMaintenanceWindow is the weekly time range (in UTC) during which
	// automated cluster maintenance can occur.
	// Default: A 30-minute window selected at random from an 8-hour block of time
	// per region, occurring on a random day of the week. For more information about
	// the time blocks for each region, see Maintenance Windows (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-maintenance-windows)
	// in Amazon Redshift Cluster Management Guide.
	// Constraints: Minimum 30-minute window.
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	// PubliclyAccessible is to specify if the cluster can be accessed from a public network.
	// +optional
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// SkipFinalClusterSnapshot determines whether a final snapshot of the cluster
	// is created before Amazon Redshift deletes the cluster.
	// If true, a final cluster snapshot is not created.
	// If false, a final cluster snapshot is created before the cluster is deleted.
	// The FinalClusterSnapshotIdentifier parameter must be specified if SkipFinalClusterSnapshot
	// is false.
	// Default: false
	// +optional
	SkipFinalClusterSnapshot *bool `json:"skipFinalClusterSnapshot,omitempty"`

	// SnapshotScheduleIdentifier is a unique identifier for the snapshot schedule.
	// +optional
	SnapshotScheduleIdentifier *string `json:"snapshotScheduleIdentifier,omitempty"`

	// Tags indicates a list of tags for the clusters.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// VPCSecurityGroupIDs a list of Virtual Private Cloud (VPC) security groups to be associated with
	// the cluster.
	// +optional
	VPCSecurityGroupIDs []string `json:"vpcSecurityGroupIds,omitempty"`

	// VPCSecurityGroupIDRefs are references to VPCSecurityGroups used to set
	// the VPCSecurityGroupIDs.
	// +immutable
	// +optional
	VPCSecurityGroupIDRefs []xpv1.Reference `json:"vpcSecurityGroupIDRefs,omitempty"`

	// VPCSecurityGroupIDSelector selects references to VPCSecurityGroups used
	// to set the VPCSecurityGroupIDs.
	// +immutable
	// +optional
	VPCSecurityGroupIDSelector *xpv1.Selector `json:"vpcSecurityGroupIDSelector,omitempty"`
}

// ClusterObservation is the representation of the current state that is observed.
type ClusterObservation struct {
	// ClusterAvailabilityStatus is the availability status of the cluster.
	ClusterAvailabilityStatus string `json:"clusterAvailabilityStatus,omitempty"`

	// ClusterCreateTime is the date and time that the cluster was created.
	ClusterCreateTime *metav1.Time `json:"clusterCreateTime,omitempty"`

	// The nodes in the cluster.
	ClusterNodes []ClusterNode `json:"clusterNodes,omitempty"`

	// The list of cluster parameter groups that are associated with this cluster.
	// Each parameter group in the list is returned with its status.
	ClusterParameterGroups []ClusterParameterGroupStatus `json:"clusterParameterGroups,omitempty"`

	// The public key for the cluster.
	ClusterPublicKey string `json:"clusterPublicKey,omitempty"`

	// The specific revision number of the database in the cluster.
	ClusterRevisionNumber string `json:"clusterRevisionNumber,omitempty"`

	// A value that returns the destination region and retention period that are
	// configured for cross-region snapshot copy.
	ClusterSnapshotCopyStatus ClusterSnapshotCopyStatus `json:"clusterSnapshotCopyStatus,omitempty"`

	// ClusterStatus is the current state of the cluster.
	ClusterStatus string `json:"clusterStatus,omitempty"`

	// Describes the status of a cluster while it is in the process of resizing
	// with an incremental resize.
	DataTransferProgress DataTransferProgress `json:"dataTransferProgress,omitempty"`

	// Describes a group of DeferredMaintenanceWindow objects.
	DeferredMaintenanceWindows []DeferredMaintenanceWindow `json:"deferredMaintenanceWindows,omitempty"`

	// The status of the elastic IP (EIP) address.
	ElasticIPStatus ElasticIPStatus `json:"elasticIPStatus,omitempty"`

	// The number of nodes that you can resize the cluster to with the elastic resize
	// method.
	ElasticResizeNumberOfNodeOptions string `json:"elasticResizeNumberOfNodeOptions,omitempty"`

	// Endpoint specifies the connection endpoint.
	Endpoint Endpoint `json:"endpoint,omitempty"`

	// The date and time when the next snapshot is expected to be taken for clusters
	// with a valid snapshot schedule and backups enabled.
	ExpectedNextSnapshotScheduleTime *metav1.Time `json:"expectedNextSnapshotScheduleTime,omitempty"`

	// The status of next expected snapshot for clusters having a valid snapshot
	// schedule and backups enabled. Possible values are the following:
	//
	//    * OnTrack - The next snapshot is expected to be taken on time.
	//
	//    * Pending - The next snapshot is pending to be taken.
	ExpectedNextSnapshotScheduleTimeStatus string `json:"expectedNextSnapshotScheduleTimeStatus,omitempty"`

	// A value that reports whether the Amazon Redshift cluster has finished applying
	// any hardware security module (HSM) settings changes specified in a modify
	// cluster command.
	//
	// Values: active, applying
	HSMStatus HSMStatus `json:"hsmStatus,omitempty"`

	// The status of a modify operation, if any, initiated for the cluster.
	ModifyStatus string `json:"modifyStatus,omitempty"`

	// The date and time in UTC when system maintenance can begin.
	NextMaintenanceWindowStartTime *metav1.Time `json:"nextMaintenanceWindowStartTime,omitempty"`

	// Cluster operations that are waiting to be started.
	PendingActions []string `json:"pendingActions,omitempty"`

	// The current state of the cluster snapshot schedule.
	SnapshotScheduleState string `json:"snapshotScheduleState,omitempty"`

	// The identifier of the VPC the cluster is in, if the cluster is in a VPC.
	VPCID string `json:"vpcId,omitempty"`
}

// ClusterParameterGroupStatus is the status of the Cluster parameter group.
type ClusterParameterGroupStatus struct {

	// The list of parameter statuses.
	//
	// For more information about parameters and parameter groups, go to Amazon
	// Redshift Parameter Groups (https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-parameter-groups.html)
	// in the Amazon Redshift Cluster Management Guide.
	ClusterParameterStatusList []ClusterParameterStatus `json:"clusterParameterStatusList,omitempty"`

	// The status of parameter updates.
	ParameterApplyStatus string `json:"parameterApplyStatus,omitempty"`

	// The name of the cluster parameter group.
	ParameterGroupName string `json:"parameterGroupName,omitempty"`
}

// ClusterParameterStatus describes the status of a Cluster parameter.
type ClusterParameterStatus struct {

	// The error that prevented the parameter from being applied to the database.
	ParameterApplyErrorDescription string `json:"parameterApplyErrorDescription,omitempty"`

	// The status of the parameter that indicates whether the parameter is in sync
	// with the database, waiting for a cluster reboot, or encountered an error
	// when being applied.
	//
	// The following are possible statuses and descriptions.
	//
	//    * in-sync: The parameter value is in sync with the database.
	//
	//    * pending-reboot: The parameter value will be applied after the cluster
	//    reboots.
	//
	//    * applying: The parameter value is being applied to the database.
	//
	//    * invalid-parameter: Cannot apply the parameter value because it has an
	//    invalid value or syntax.
	//
	//    * apply-deferred: The parameter contains static property changes. The
	//    changes are deferred until the cluster reboots.
	//
	//    * apply-error: Cannot connect to the cluster. The parameter change will
	//    be applied after the cluster reboots.
	//
	//    * unknown-error: Cannot apply the parameter change right now. The change
	//    will be applied after the cluster reboots.
	ParameterApplyStatus string `json:"parameterApplyStatus,omitempty"`

	// The name of the parameter.
	ParameterName string `json:"parameterName,omitempty"`
}

// ClusterSecurityGroupMembership is used as a response element for queries on Cluster security.
type ClusterSecurityGroupMembership struct {
	// The name of the cluster security group.
	ClusterSecurityGroupName string `json:"clusterSecurityGroupName,omitempty"`

	// The status of the cluster security group.
	Status string `json:"status,omitempty"`
}

// ClusterNode is the identifier of a node in a cluster.
type ClusterNode struct {

	// Whether the node is a leader node or a compute node.
	NodeRole string `json:"nodeRole,omitempty"`

	// The private IP address of a node within a cluster.
	PrivateIPAddress string `json:"privateIPAddress,omitempty"`

	// The public IP address of a node within a cluster.
	PublicIPAddress string `json:"publicIPAddress,omitempty"`
}

// ClusterSnapshotCopyStatus returns the destination region and retention period
// that are configured for cross-region snapshot copy.
type ClusterSnapshotCopyStatus struct {

	// The destination region that snapshots are automatically copied to when cross-region
	// snapshot copy is enabled.
	DestinationRegion string `json:"destinationRegion,omitempty"`

	// The number of days that automated snapshots are retained in the destination
	// region after they are copied from a source region. If the value is -1, the
	// manual snapshot is retained indefinitely.
	//
	// The value must be either -1 or an integer between 1 and 3,653.
	ManualSnapshotRetentionPeriod int32 `json:"manualSnapshotRetentionPeriod,omitempty"`

	// The number of days that automated snapshots are retained in the destination
	// region after they are copied from a source region.
	RetentionPeriod int64 `json:"retentionPeriod,omitempty"`

	// The name of the snapshot copy grant.
	SnapshotCopyGrantName string `json:"snapshotCopyGrantName,omitempty"`
}

// DataTransferProgress describes the status of a cluster while it is in the process of resizing
// with an incremental resize.
type DataTransferProgress struct {

	// Describes the data transfer rate in MB's per second.
	CurrentRateInMegaBytesPerSecond int `json:"currentRateInMegaBytesPerSecond,omitempty"`

	// Describes the total amount of data that has been transferred in MB's.
	DataTransferredInMegaBytes int64 `json:"dataTransferredInMegaBytes,omitempty"`

	// Describes the number of seconds that have elapsed during the data transfer.
	ElapsedTimeInSeconds int64 `json:"elapsedTimeInSeconds,omitempty"`

	// Describes the estimated number of seconds remaining to complete the transfer.
	EstimatedTimeToCompletionInSeconds int64 `json:"estimatedTimeToCompletionInSeconds,omitempty"`

	// Describes the status of the cluster. While the transfer is in progress the
	// status is transferringdata.
	Status string `json:"status,omitempty"`

	// Describes the total amount of data to be transferred in megabytes.
	TotalDataInMegaBytes int64 `json:"totalDataInMegaBytes,omitempty"`
}

// DeferredMaintenanceWindow describes a deferred maintenance window
type DeferredMaintenanceWindow struct {

	// A timestamp for the end of the time period when we defer maintenance.
	DeferMaintenanceEndTime *metav1.Time `json:"deferMaintenanceEndTime,omitempty"`

	// A unique identifier for the maintenance window.
	DeferMaintenanceIdentifier string `json:"deferMaintenanceIdentifier,omitempty"`

	// A timestamp for the beginning of the time period when we defer maintenance.
	DeferMaintenanceStartTime *metav1.Time `json:"deferMaintenanceStartTime,omitempty"`
}

// ElasticIPStatus describes the status of the elastic IP (EIP) address.
type ElasticIPStatus struct {

	// The elastic IP (EIP) address for the cluster.
	ElasticIP string `json:"elasticIP,omitempty"`

	// The status of the elastic IP (EIP) address.
	Status string `json:"status,omitempty"`
}

// Endpoint is used as a response element in the following actions:
//   - CreateCluster
//   - DescribeClusters
//   - DeleteCluster
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/Endpoint
type Endpoint struct {
	// Address specifies the DNS address of the cluster.
	Address string `json:"address,omitempty"`

	// Port specifies the port that the database engine is listening on.
	Port int32 `json:"port,omitempty"`
}

// HSMStatus describes the status of changes to HSM settings.
type HSMStatus struct {

	// Specifies the name of the HSM client certificate the Amazon Redshift cluster
	// uses to retrieve the data encryption keys stored in an HSM.
	HSMClientCertificateIdentifier string `json:"hsmClientCertificateIdentifier,omitempty"`

	// Specifies the name of the HSM configuration that contains the information
	// the Amazon Redshift cluster can use to retrieve and store keys in an HSM.
	HSMConfigurationIdentifier string `json:"hsmConfigurationIdentifier,omitempty"`

	// Reports whether the Amazon Redshift cluster has finished applying any HSM
	// settings changes specified in a modify cluster command.
	//
	// Values: active, applying
	Status string `json:"status,omitempty"`
}

// RestoreStatus describes the status of a cluster restore action. Returns null if the cluster
// was not created by restoring a snapshot.
type RestoreStatus struct {

	// The number of megabytes per second being transferred from the backup storage.
	// Returns the average rate for a completed backup. This field is only updated
	// when you restore to DC2 and DS2 node types.
	CurrentRestoreRateInMegaBytesPerSecond float64 `json:"currentRestoreRateInMegaBytesPerSecond,omitempty"`

	// The amount of time an in-progress restore has been running, or the amount
	// of time it took a completed restore to finish. This field is only updated
	// when you restore to DC2 and DS2 node types.
	ElapsedTimeInSeconds int64 `json:"elapsedTimeInSeconds,omitempty"`

	// The estimate of the time remaining before the restore will complete. Returns
	// 0 for a completed restore. This field is only updated when you restore to
	// DC2 and DS2 node types.
	EstimatedTimeToCompletionInSeconds int64 `json:"estimatedTimeToCompletionInSeconds,omitempty"`

	// The number of megabytes that have been transferred from snapshot storage.
	// This field is only updated when you restore to DC2 and DS2 node types.
	ProgressInMegaBytes int64 `json:"progressInMegaBytes,omitempty"`

	// The size of the set of snapshot data used to restore the cluster. This field
	// is only updated when you restore to DC2 and DS2 node types.
	SnapshotSizeInMegaBytes int64 `json:"snapshotSizeInMegaBytes,omitempty"`

	// The status of the restore action. Returns starting, restoring, completed,
	// or failed.
	Status string `json:"status,omitempty"`
}

// VPCSecurityGroupMembership is used as a response element for queries on VPC security
// Can be removed after moving to v1beta1
type VPCSecurityGroupMembership struct {

	// The status of the VPC security group.
	Status string `json:"status"`

	// The identifier of the VPC security group.
	VPCSecurityGroupID string `json:"vpcSecurityGroupID"`
}

// Tag represetnt a key-pair metadata assigned to a Redshift Cluster
type Tag struct {

	// The key of the tag.
	Key string `json:"tag"`

	// The value of the tag.
	Value string `json:"value"`
}

// ClusterSpec defines the desired state of an AWS Redshift Cluster.
type ClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ClusterParameters `json:"forProvider"`
}

// ClusterStatus represents the observed state of an AWS Redshift Cluster.
type ClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Cluster is a managed resource that represents an AWS Redshift cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.clusterStatus"
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

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}
