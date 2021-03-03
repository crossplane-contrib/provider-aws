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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AccountQuota struct {
	AccountQuotaName *string `json:"accountQuotaName,omitempty"`
}

type AvailabilityZone struct {
	Name *string `json:"name,omitempty"`
}

type AvailableProcessorFeature struct {
	AllowedValues *string `json:"allowedValues,omitempty"`

	DefaultValue *string `json:"defaultValue,omitempty"`

	Name *string `json:"name,omitempty"`
}

type Certificate struct {
	CertificateARN *string `json:"certificateARN,omitempty"`

	CertificateIdentifier *string `json:"certificateIdentifier,omitempty"`

	CertificateType *string `json:"certificateType,omitempty"`

	CustomerOverride *bool `json:"customerOverride,omitempty"`

	CustomerOverrideValidTill *metav1.Time `json:"customerOverrideValidTill,omitempty"`

	Thumbprint *string `json:"thumbprint,omitempty"`

	ValidFrom *metav1.Time `json:"validFrom,omitempty"`

	ValidTill *metav1.Time `json:"validTill,omitempty"`
}

type CharacterSet struct {
	CharacterSetDescription *string `json:"characterSetDescription,omitempty"`

	CharacterSetName *string `json:"characterSetName,omitempty"`
}

type CloudwatchLogsExportConfiguration struct {
	DisableLogTypes []*string `json:"disableLogTypes,omitempty"`

	EnableLogTypes []*string `json:"enableLogTypes,omitempty"`
}

type ClusterPendingModifiedValues struct {
	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	MasterUserPassword *string `json:"masterUserPassword,omitempty"`
}

type ConnectionPoolConfiguration struct {
	ConnectionBorrowTimeout *int64 `json:"connectionBorrowTimeout,omitempty"`

	InitQuery *string `json:"initQuery,omitempty"`

	MaxConnectionsPercent *int64 `json:"maxConnectionsPercent,omitempty"`

	MaxIdleConnectionsPercent *int64 `json:"maxIdleConnectionsPercent,omitempty"`

	SessionPinningFilters []*string `json:"sessionPinningFilters,omitempty"`
}

type ConnectionPoolConfigurationInfo struct {
	InitQuery *string `json:"initQuery,omitempty"`

	SessionPinningFilters []*string `json:"sessionPinningFilters,omitempty"`
}

type CustomAvailabilityZone struct {
	CustomAvailabilityZoneID *string `json:"customAvailabilityZoneID,omitempty"`

	CustomAvailabilityZoneName *string `json:"customAvailabilityZoneName,omitempty"`

	CustomAvailabilityZoneStatus *string `json:"customAvailabilityZoneStatus,omitempty"`
}

type DBClusterEndpoint struct {
	CustomEndpointType *string `json:"customEndpointType,omitempty"`

	DBClusterEndpointARN *string `json:"dbClusterEndpointARN,omitempty"`

	DBClusterEndpointIdentifier *string `json:"dbClusterEndpointIdentifier,omitempty"`

	DBClusterEndpointResourceIdentifier *string `json:"dbClusterEndpointResourceIdentifier,omitempty"`

	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	Endpoint *string `json:"endpoint,omitempty"`

	EndpointType *string `json:"endpointType,omitempty"`

	ExcludedMembers []*string `json:"excludedMembers,omitempty"`

	StaticMembers []*string `json:"staticMembers,omitempty"`

	Status *string `json:"status,omitempty"`
}

type DBClusterMember struct {
	DBClusterParameterGroupStatus *string `json:"dbClusterParameterGroupStatus,omitempty"`

	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	IsClusterWriter *bool `json:"isClusterWriter,omitempty"`

	PromotionTier *int64 `json:"promotionTier,omitempty"`
}

type DBClusterOptionGroupStatus struct {
	DBClusterOptionGroupName *string `json:"dbClusterOptionGroupName,omitempty"`

	Status *string `json:"status,omitempty"`
}

type DBClusterParameterGroup struct {
	DBClusterParameterGroupARN *string `json:"dbClusterParameterGroupARN,omitempty"`

	DBClusterParameterGroupName *string `json:"dbClusterParameterGroupName,omitempty"`

	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`

	Description *string `json:"description,omitempty"`
}

type DBClusterRole struct {
	FeatureName *string `json:"featureName,omitempty"`

	RoleARN *string `json:"roleARN,omitempty"`

	Status *string `json:"status,omitempty"`
}

type DBClusterSnapshot struct {
	AvailabilityZones []*string `json:"availabilityZones,omitempty"`

	ClusterCreateTime *metav1.Time `json:"clusterCreateTime,omitempty"`

	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	DBClusterSnapshotARN *string `json:"dbClusterSnapshotARN,omitempty"`

	DBClusterSnapshotIdentifier *string `json:"dbClusterSnapshotIdentifier,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MasterUsername *string `json:"masterUsername,omitempty"`

	SnapshotCreateTime *metav1.Time `json:"snapshotCreateTime,omitempty"`

	SnapshotType *string `json:"snapshotType,omitempty"`

	SourceDBClusterSnapshotARN *string `json:"sourceDBClusterSnapshotARN,omitempty"`

	Status *string `json:"status,omitempty"`

	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`
	// A list of tags. For more information, see Tagging Amazon RDS Resources (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	TagList []*Tag `json:"tagList,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type DBClusterSnapshotAttribute struct {
	AttributeName *string `json:"attributeName,omitempty"`
}

type DBClusterSnapshotAttributesResult struct {
	DBClusterSnapshotIdentifier *string `json:"dbClusterSnapshotIdentifier,omitempty"`
}

type DBCluster_SDK struct {
	ActivityStreamKinesisStreamName *string `json:"activityStreamKinesisStreamName,omitempty"`

	ActivityStreamKMSKeyID *string `json:"activityStreamKMSKeyID,omitempty"`

	ActivityStreamMode *string `json:"activityStreamMode,omitempty"`

	ActivityStreamStatus *string `json:"activityStreamStatus,omitempty"`

	AllocatedStorage *int64 `json:"allocatedStorage,omitempty"`

	AssociatedRoles []*DBClusterRole `json:"associatedRoles,omitempty"`

	AvailabilityZones []*string `json:"availabilityZones,omitempty"`

	BacktrackConsumedChangeRecords *int64 `json:"backtrackConsumedChangeRecords,omitempty"`

	BacktrackWindow *int64 `json:"backtrackWindow,omitempty"`

	BackupRetentionPeriod *int64 `json:"backupRetentionPeriod,omitempty"`

	Capacity *int64 `json:"capacity,omitempty"`

	CharacterSetName *string `json:"characterSetName,omitempty"`

	CloneGroupID *string `json:"cloneGroupID,omitempty"`

	ClusterCreateTime *metav1.Time `json:"clusterCreateTime,omitempty"`

	CopyTagsToSnapshot *bool `json:"copyTagsToSnapshot,omitempty"`

	CrossAccountClone *bool `json:"crossAccountClone,omitempty"`

	CustomEndpoints []*string `json:"customEndpoints,omitempty"`

	DBClusterARN *string `json:"dbClusterARN,omitempty"`

	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	DBClusterMembers []*DBClusterMember `json:"dbClusterMembers,omitempty"`

	DBClusterOptionGroupMemberships []*DBClusterOptionGroupStatus `json:"dbClusterOptionGroupMemberships,omitempty"`

	DBClusterParameterGroup *string `json:"dbClusterParameterGroup,omitempty"`

	DBSubnetGroup *string `json:"dbSubnetGroup,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	DBClusterResourceID *string `json:"dbClusterResourceID,omitempty"`

	DeletionProtection *bool `json:"deletionProtection,omitempty"`
	// List of Active Directory Domain membership records associated with a DB instance
	// or cluster.
	DomainMemberships []*DomainMembership `json:"domainMemberships,omitempty"`

	EarliestBacktrackTime *metav1.Time `json:"earliestBacktrackTime,omitempty"`

	EarliestRestorableTime *metav1.Time `json:"earliestRestorableTime,omitempty"`

	EnabledCloudwatchLogsExports []*string `json:"enabledCloudwatchLogsExports,omitempty"`

	Endpoint *string `json:"endpoint,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineMode *string `json:"engineMode,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	GlobalWriteForwardingRequested *bool `json:"globalWriteForwardingRequested,omitempty"`

	GlobalWriteForwardingStatus *string `json:"globalWriteForwardingStatus,omitempty"`

	HostedZoneID *string `json:"hostedZoneID,omitempty"`

	HTTPEndpointEnabled *bool `json:"httpEndpointEnabled,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	LatestRestorableTime *metav1.Time `json:"latestRestorableTime,omitempty"`

	MasterUsername *string `json:"masterUsername,omitempty"`

	MultiAZ *bool `json:"multiAZ,omitempty"`

	PercentProgress *string `json:"percentProgress,omitempty"`

	Port *int64 `json:"port,omitempty"`

	PreferredBackupWindow *string `json:"preferredBackupWindow,omitempty"`

	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	ReadReplicaIdentifiers []*string `json:"readReplicaIdentifiers,omitempty"`

	ReaderEndpoint *string `json:"readerEndpoint,omitempty"`

	ReplicationSourceIdentifier *string `json:"replicationSourceIdentifier,omitempty"`
	// Shows the scaling configuration for an Aurora DB cluster in serverless DB
	// engine mode.
	//
	// For more information, see Using Amazon Aurora Serverless (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless.html)
	// in the Amazon Aurora User Guide.
	ScalingConfigurationInfo *ScalingConfigurationInfo `json:"scalingConfigurationInfo,omitempty"`

	Status *string `json:"status,omitempty"`

	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`
	// A list of tags. For more information, see Tagging Amazon RDS Resources (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	TagList []*Tag `json:"tagList,omitempty"`

	VPCSecurityGroups []*VPCSecurityGroupMembership `json:"vpcSecurityGroups,omitempty"`
}

type DBEngineVersion struct {
	DBEngineDescription *string `json:"dbEngineDescription,omitempty"`

	DBEngineVersionDescription *string `json:"dbEngineVersionDescription,omitempty"`

	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	ExportableLogTypes []*string `json:"exportableLogTypes,omitempty"`

	Status *string `json:"status,omitempty"`

	SupportsGlobalDatabases *bool `json:"supportsGlobalDatabases,omitempty"`

	SupportsLogExportsToCloudwatchLogs *bool `json:"supportsLogExportsToCloudwatchLogs,omitempty"`

	SupportsParallelQuery *bool `json:"supportsParallelQuery,omitempty"`

	SupportsReadReplica *bool `json:"supportsReadReplica,omitempty"`
}

type DBInstance struct {
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	CACertificateIdentifier *string `json:"caCertificateIdentifier,omitempty"`

	CharacterSetName *string `json:"characterSetName,omitempty"`

	CopyTagsToSnapshot *bool `json:"copyTagsToSnapshot,omitempty"`

	CustomerOwnedIPEnabled *bool `json:"customerOwnedIPEnabled,omitempty"`

	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	DBInstanceARN *string `json:"dbInstanceARN,omitempty"`

	DBInstanceClass *string `json:"dbInstanceClass,omitempty"`

	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	DBInstanceStatus *string `json:"dbInstanceStatus,omitempty"`

	DBName *string `json:"dbName,omitempty"`

	DBIResourceID *string `json:"dbiResourceID,omitempty"`

	DeletionProtection *bool `json:"deletionProtection,omitempty"`
	// List of Active Directory Domain membership records associated with a DB instance
	// or cluster.
	DomainMemberships []*DomainMembership `json:"domainMemberships,omitempty"`

	EnabledCloudwatchLogsExports []*string `json:"enabledCloudwatchLogsExports,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	EnhancedMonitoringResourceARN *string `json:"enhancedMonitoringResourceARN,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	InstanceCreateTime *metav1.Time `json:"instanceCreateTime,omitempty"`

	IOPS *int64 `json:"iops,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	LatestRestorableTime *metav1.Time `json:"latestRestorableTime,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MasterUsername *string `json:"masterUsername,omitempty"`

	MaxAllocatedStorage *int64 `json:"maxAllocatedStorage,omitempty"`

	MonitoringInterval *int64 `json:"monitoringInterval,omitempty"`

	MonitoringRoleARN *string `json:"monitoringRoleARN,omitempty"`

	MultiAZ *bool `json:"multiAZ,omitempty"`

	NcharCharacterSetName *string `json:"ncharCharacterSetName,omitempty"`

	PerformanceInsightsEnabled *bool `json:"performanceInsightsEnabled,omitempty"`

	PerformanceInsightsKMSKeyID *string `json:"performanceInsightsKMSKeyID,omitempty"`

	PerformanceInsightsRetentionPeriod *int64 `json:"performanceInsightsRetentionPeriod,omitempty"`

	PreferredBackupWindow *string `json:"preferredBackupWindow,omitempty"`

	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	PromotionTier *int64 `json:"promotionTier,omitempty"`

	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	ReadReplicaSourceDBInstanceIdentifier *string `json:"readReplicaSourceDBInstanceIdentifier,omitempty"`

	SecondaryAvailabilityZone *string `json:"secondaryAvailabilityZone,omitempty"`

	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	StorageType *string `json:"storageType,omitempty"`
	// A list of tags. For more information, see Tagging Amazon RDS Resources (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	TagList []*Tag `json:"tagList,omitempty"`

	TDECredentialARN *string `json:"tdeCredentialARN,omitempty"`

	Timezone *string `json:"timezone,omitempty"`

	VPCSecurityGroups []*VPCSecurityGroupMembership `json:"vpcSecurityGroups,omitempty"`
}

type DBInstanceAutomatedBackup struct {
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	BackupRetentionPeriod *int64 `json:"backupRetentionPeriod,omitempty"`

	DBInstanceARN *string `json:"dbInstanceARN,omitempty"`

	DBInstanceAutomatedBackupsARN *string `json:"dbInstanceAutomatedBackupsARN,omitempty"`

	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	DBIResourceID *string `json:"dbiResourceID,omitempty"`

	Encrypted *bool `json:"encrypted,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	InstanceCreateTime *metav1.Time `json:"instanceCreateTime,omitempty"`

	IOPS *int64 `json:"iops,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MasterUsername *string `json:"masterUsername,omitempty"`

	OptionGroupName *string `json:"optionGroupName,omitempty"`

	Region *string `json:"region,omitempty"`

	Status *string `json:"status,omitempty"`

	StorageType *string `json:"storageType,omitempty"`

	TDECredentialARN *string `json:"tdeCredentialARN,omitempty"`

	Timezone *string `json:"timezone,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type DBInstanceAutomatedBackupsReplication struct {
	DBInstanceAutomatedBackupsARN *string `json:"dbInstanceAutomatedBackupsARN,omitempty"`
}

type DBInstanceRole struct {
	FeatureName *string `json:"featureName,omitempty"`

	RoleARN *string `json:"roleARN,omitempty"`

	Status *string `json:"status,omitempty"`
}

type DBInstanceStatusInfo struct {
	Message *string `json:"message,omitempty"`

	Normal *bool `json:"normal,omitempty"`

	Status *string `json:"status,omitempty"`

	StatusType *string `json:"statusType,omitempty"`
}

type DBParameterGroup struct {
	DBParameterGroupARN *string `json:"dbParameterGroupARN,omitempty"`

	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`

	DBParameterGroupName *string `json:"dbParameterGroupName,omitempty"`

	Description *string `json:"description,omitempty"`
}

type DBParameterGroupStatus struct {
	DBParameterGroupName *string `json:"dbParameterGroupName,omitempty"`

	ParameterApplyStatus *string `json:"parameterApplyStatus,omitempty"`
}

type DBProxy struct {
	CreatedDate *metav1.Time `json:"createdDate,omitempty"`

	DBProxyARN *string `json:"dbProxyARN,omitempty"`

	DBProxyName *string `json:"dbProxyName,omitempty"`

	DebugLogging *bool `json:"debugLogging,omitempty"`

	Endpoint *string `json:"endpoint,omitempty"`

	EngineFamily *string `json:"engineFamily,omitempty"`

	RequireTLS *bool `json:"requireTLS,omitempty"`

	RoleARN *string `json:"roleARN,omitempty"`

	UpdatedDate *metav1.Time `json:"updatedDate,omitempty"`

	VPCSecurityGroupIDs []*string `json:"vpcSecurityGroupIDs,omitempty"`

	VPCSubnetIDs []*string `json:"vpcSubnetIDs,omitempty"`
}

type DBProxyTarget struct {
	Endpoint *string `json:"endpoint,omitempty"`

	RdsResourceID *string `json:"rdsResourceID,omitempty"`

	TargetARN *string `json:"targetARN,omitempty"`

	TrackedClusterID *string `json:"trackedClusterID,omitempty"`
}

type DBProxyTargetGroup struct {
	CreatedDate *metav1.Time `json:"createdDate,omitempty"`

	DBProxyName *string `json:"dbProxyName,omitempty"`

	IsDefault *bool `json:"isDefault,omitempty"`

	Status *string `json:"status,omitempty"`

	TargetGroupARN *string `json:"targetGroupARN,omitempty"`

	TargetGroupName *string `json:"targetGroupName,omitempty"`

	UpdatedDate *metav1.Time `json:"updatedDate,omitempty"`
}

type DBSecurityGroup struct {
	DBSecurityGroupARN *string `json:"dbSecurityGroupARN,omitempty"`

	DBSecurityGroupDescription *string `json:"dbSecurityGroupDescription,omitempty"`

	DBSecurityGroupName *string `json:"dbSecurityGroupName,omitempty"`

	OwnerID *string `json:"ownerID,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type DBSecurityGroupMembership struct {
	DBSecurityGroupName *string `json:"dbSecurityGroupName,omitempty"`

	Status *string `json:"status,omitempty"`
}

type DBSnapshot struct {
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	DBSnapshotARN *string `json:"dbSnapshotARN,omitempty"`

	DBSnapshotIdentifier *string `json:"dbSnapshotIdentifier,omitempty"`

	DBIResourceID *string `json:"dbiResourceID,omitempty"`

	Encrypted *bool `json:"encrypted,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	InstanceCreateTime *metav1.Time `json:"instanceCreateTime,omitempty"`

	IOPS *int64 `json:"iops,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MasterUsername *string `json:"masterUsername,omitempty"`

	OptionGroupName *string `json:"optionGroupName,omitempty"`

	SnapshotCreateTime *metav1.Time `json:"snapshotCreateTime,omitempty"`

	SnapshotType *string `json:"snapshotType,omitempty"`

	SourceDBSnapshotIdentifier *string `json:"sourceDBSnapshotIdentifier,omitempty"`

	SourceRegion *string `json:"sourceRegion,omitempty"`

	Status *string `json:"status,omitempty"`

	StorageType *string `json:"storageType,omitempty"`
	// A list of tags. For more information, see Tagging Amazon RDS Resources (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	TagList []*Tag `json:"tagList,omitempty"`

	TDECredentialARN *string `json:"tdeCredentialARN,omitempty"`

	Timezone *string `json:"timezone,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type DBSnapshotAttribute struct {
	AttributeName *string `json:"attributeName,omitempty"`
}

type DBSnapshotAttributesResult struct {
	DBSnapshotIdentifier *string `json:"dbSnapshotIdentifier,omitempty"`
}

type DBSubnetGroup struct {
	DBSubnetGroupARN *string `json:"dbSubnetGroupARN,omitempty"`

	DBSubnetGroupDescription *string `json:"dbSubnetGroupDescription,omitempty"`

	DBSubnetGroupName *string `json:"dbSubnetGroupName,omitempty"`

	SubnetGroupStatus *string `json:"subnetGroupStatus,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type DescribeDBLogFilesDetails struct {
	LogFileName *string `json:"logFileName,omitempty"`
}

type DomainMembership struct {
	Domain *string `json:"domain,omitempty"`

	FQDN *string `json:"fQDN,omitempty"`

	IAMRoleName *string `json:"iamRoleName,omitempty"`

	Status *string `json:"status,omitempty"`
}

type EC2SecurityGroup struct {
	EC2SecurityGroupID *string `json:"ec2SecurityGroupID,omitempty"`

	EC2SecurityGroupName *string `json:"ec2SecurityGroupName,omitempty"`

	EC2SecurityGroupOwnerID *string `json:"ec2SecurityGroupOwnerID,omitempty"`

	Status *string `json:"status,omitempty"`
}

type Endpoint struct {
	Address *string `json:"address,omitempty"`

	HostedZoneID *string `json:"hostedZoneID,omitempty"`
}

type EngineDefaults struct {
	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`

	Marker *string `json:"marker,omitempty"`
}

type Event struct {
	Date *metav1.Time `json:"date,omitempty"`

	Message *string `json:"message,omitempty"`

	SourceARN *string `json:"sourceARN,omitempty"`

	SourceIdentifier *string `json:"sourceIdentifier,omitempty"`
}

type EventCategoriesMap struct {
	SourceType *string `json:"sourceType,omitempty"`
}

type EventSubscription struct {
	CustSubscriptionID *string `json:"custSubscriptionID,omitempty"`

	CustomerAWSID *string `json:"customerAWSID,omitempty"`

	Enabled *bool `json:"enabled,omitempty"`

	EventSubscriptionARN *string `json:"eventSubscriptionARN,omitempty"`

	SnsTopicARN *string `json:"snsTopicARN,omitempty"`

	SourceType *string `json:"sourceType,omitempty"`

	Status *string `json:"status,omitempty"`

	SubscriptionCreationTime *string `json:"subscriptionCreationTime,omitempty"`
}

type ExportTask struct {
	ExportOnly []*string `json:"exportOnly,omitempty"`

	ExportTaskIdentifier *string `json:"exportTaskIdentifier,omitempty"`

	FailureCause *string `json:"failureCause,omitempty"`

	IAMRoleARN *string `json:"iamRoleARN,omitempty"`

	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	S3Bucket *string `json:"s3Bucket,omitempty"`

	S3Prefix *string `json:"s3Prefix,omitempty"`

	SnapshotTime *metav1.Time `json:"snapshotTime,omitempty"`

	SourceARN *string `json:"sourceARN,omitempty"`

	Status *string `json:"status,omitempty"`

	TaskEndTime *metav1.Time `json:"taskEndTime,omitempty"`

	TaskStartTime *metav1.Time `json:"taskStartTime,omitempty"`

	WarningMessage *string `json:"warningMessage,omitempty"`
}

type Filter struct {
	Name *string `json:"name,omitempty"`

	Values []*string `json:"values,omitempty"`
}

type GlobalCluster struct {
	DatabaseName *string `json:"databaseName,omitempty"`

	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	GlobalClusterARN *string `json:"globalClusterARN,omitempty"`

	GlobalClusterIdentifier *string `json:"globalClusterIdentifier,omitempty"`

	GlobalClusterResourceID *string `json:"globalClusterResourceID,omitempty"`

	Status *string `json:"status,omitempty"`

	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`
}

type GlobalClusterMember struct {
	DBClusterARN *string `json:"dbClusterARN,omitempty"`

	GlobalWriteForwardingStatus *string `json:"globalWriteForwardingStatus,omitempty"`

	IsWriter *bool `json:"isWriter,omitempty"`
}

type IPRange struct {
	CIDRIP *string `json:"cidrIP,omitempty"`

	Status *string `json:"status,omitempty"`
}

type InstallationMedia struct {
	CustomAvailabilityZoneID *string `json:"customAvailabilityZoneID,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineInstallationMediaPath *string `json:"engineInstallationMediaPath,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	InstallationMediaID *string `json:"installationMediaID,omitempty"`

	OSInstallationMediaPath *string `json:"oSInstallationMediaPath,omitempty"`

	Status *string `json:"status,omitempty"`
}

type InstallationMediaFailureCause struct {
	Message *string `json:"message,omitempty"`
}

type MinimumEngineVersionPerAllowedValue struct {
	AllowedValue *string `json:"allowedValue,omitempty"`

	MinimumEngineVersion *string `json:"minimumEngineVersion,omitempty"`
}

type Option struct {
	OptionDescription *string `json:"optionDescription,omitempty"`

	OptionName *string `json:"optionName,omitempty"`

	OptionVersion *string `json:"optionVersion,omitempty"`

	Permanent *bool `json:"permanent,omitempty"`

	Persistent *bool `json:"persistent,omitempty"`

	Port *int64 `json:"port,omitempty"`

	VPCSecurityGroupMemberships []*VPCSecurityGroupMembership `json:"vpcSecurityGroupMemberships,omitempty"`
}

type OptionConfiguration struct {
	OptionName *string `json:"optionName,omitempty"`

	OptionVersion *string `json:"optionVersion,omitempty"`

	Port *int64 `json:"port,omitempty"`
}

type OptionGroup struct {
	AllowsVPCAndNonVPCInstanceMemberships *bool `json:"allowsVPCAndNonVPCInstanceMemberships,omitempty"`

	EngineName *string `json:"engineName,omitempty"`

	MajorEngineVersion *string `json:"majorEngineVersion,omitempty"`

	OptionGroupARN *string `json:"optionGroupARN,omitempty"`

	OptionGroupDescription *string `json:"optionGroupDescription,omitempty"`

	OptionGroupName *string `json:"optionGroupName,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type OptionGroupMembership struct {
	OptionGroupName *string `json:"optionGroupName,omitempty"`

	Status *string `json:"status,omitempty"`
}

type OptionGroupOption struct {
	DefaultPort *int64 `json:"defaultPort,omitempty"`

	Description *string `json:"description,omitempty"`

	EngineName *string `json:"engineName,omitempty"`

	MajorEngineVersion *string `json:"majorEngineVersion,omitempty"`

	MinimumRequiredMinorEngineVersion *string `json:"minimumRequiredMinorEngineVersion,omitempty"`

	Name *string `json:"name,omitempty"`

	Permanent *bool `json:"permanent,omitempty"`

	Persistent *bool `json:"persistent,omitempty"`

	PortRequired *bool `json:"portRequired,omitempty"`

	RequiresAutoMinorEngineVersionUpgrade *bool `json:"requiresAutoMinorEngineVersionUpgrade,omitempty"`

	SupportsOptionVersionDowngrade *bool `json:"supportsOptionVersionDowngrade,omitempty"`

	VPCOnly *bool `json:"vpcOnly,omitempty"`
}

type OptionGroupOptionSetting struct {
	AllowedValues *string `json:"allowedValues,omitempty"`

	ApplyType *string `json:"applyType,omitempty"`

	DefaultValue *string `json:"defaultValue,omitempty"`

	IsModifiable *bool `json:"isModifiable,omitempty"`

	IsRequired *bool `json:"isRequired,omitempty"`

	SettingDescription *string `json:"settingDescription,omitempty"`

	SettingName *string `json:"settingName,omitempty"`
}

type OptionSetting struct {
	AllowedValues *string `json:"allowedValues,omitempty"`

	ApplyType *string `json:"applyType,omitempty"`

	DataType *string `json:"dataType,omitempty"`

	DefaultValue *string `json:"defaultValue,omitempty"`

	Description *string `json:"description,omitempty"`

	IsCollection *bool `json:"isCollection,omitempty"`

	IsModifiable *bool `json:"isModifiable,omitempty"`

	Name *string `json:"name,omitempty"`

	Value *string `json:"value,omitempty"`
}

type OptionVersion struct {
	IsDefault *bool `json:"isDefault,omitempty"`

	Version *string `json:"version,omitempty"`
}

type OrderableDBInstanceOption struct {
	AvailabilityZoneGroup *string `json:"availabilityZoneGroup,omitempty"`

	DBInstanceClass *string `json:"dbInstanceClass,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MaxIOPSPerDBInstance *int64 `json:"maxIOPSPerDBInstance,omitempty"`

	MaxStorageSize *int64 `json:"maxStorageSize,omitempty"`

	MinIOPSPerDBInstance *int64 `json:"minIOPSPerDBInstance,omitempty"`

	MinStorageSize *int64 `json:"minStorageSize,omitempty"`

	MultiAZCapable *bool `json:"multiAZCapable,omitempty"`

	OutpostCapable *bool `json:"outpostCapable,omitempty"`

	ReadReplicaCapable *bool `json:"readReplicaCapable,omitempty"`

	StorageType *string `json:"storageType,omitempty"`

	SupportsEnhancedMonitoring *bool `json:"supportsEnhancedMonitoring,omitempty"`

	SupportsGlobalDatabases *bool `json:"supportsGlobalDatabases,omitempty"`

	SupportsIAMDatabaseAuthentication *bool `json:"supportsIAMDatabaseAuthentication,omitempty"`

	SupportsIOPS *bool `json:"supportsIOPS,omitempty"`

	SupportsKerberosAuthentication *bool `json:"supportsKerberosAuthentication,omitempty"`

	SupportsPerformanceInsights *bool `json:"supportsPerformanceInsights,omitempty"`

	SupportsStorageAutoscaling *bool `json:"supportsStorageAutoscaling,omitempty"`

	SupportsStorageEncryption *bool `json:"supportsStorageEncryption,omitempty"`

	VPC *bool `json:"vpc,omitempty"`
}

type Outpost struct {
	ARN *string `json:"arn,omitempty"`
}

type Parameter struct {
	AllowedValues *string `json:"allowedValues,omitempty"`

	ApplyType *string `json:"applyType,omitempty"`

	DataType *string `json:"dataType,omitempty"`

	Description *string `json:"description,omitempty"`

	IsModifiable *bool `json:"isModifiable,omitempty"`

	MinimumEngineVersion *string `json:"minimumEngineVersion,omitempty"`

	ParameterName *string `json:"parameterName,omitempty"`

	ParameterValue *string `json:"parameterValue,omitempty"`

	Source *string `json:"source,omitempty"`
}

type PendingCloudwatchLogsExports struct {
	LogTypesToDisable []*string `json:"logTypesToDisable,omitempty"`

	LogTypesToEnable []*string `json:"logTypesToEnable,omitempty"`
}

type PendingMaintenanceAction struct {
	Action *string `json:"action,omitempty"`

	AutoAppliedAfterDate *metav1.Time `json:"autoAppliedAfterDate,omitempty"`

	CurrentApplyDate *metav1.Time `json:"currentApplyDate,omitempty"`

	Description *string `json:"description,omitempty"`

	ForcedApplyDate *metav1.Time `json:"forcedApplyDate,omitempty"`

	OptInStatus *string `json:"optInStatus,omitempty"`
}

type PendingModifiedValues struct {
	AllocatedStorage *int64 `json:"allocatedStorage,omitempty"`

	BackupRetentionPeriod *int64 `json:"backupRetentionPeriod,omitempty"`

	CACertificateIdentifier *string `json:"caCertificateIdentifier,omitempty"`

	DBInstanceClass *string `json:"dbInstanceClass,omitempty"`

	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	DBSubnetGroupName *string `json:"dbSubnetGroupName,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IAMDatabaseAuthenticationEnabled *bool `json:"iamDatabaseAuthenticationEnabled,omitempty"`

	IOPS *int64 `json:"iops,omitempty"`

	LicenseModel *string `json:"licenseModel,omitempty"`

	MasterUserPassword *string `json:"masterUserPassword,omitempty"`

	MultiAZ *bool `json:"multiAZ,omitempty"`

	Port *int64 `json:"port,omitempty"`

	StorageType *string `json:"storageType,omitempty"`
}

type ProcessorFeature struct {
	Name *string `json:"name,omitempty"`

	Value *string `json:"value,omitempty"`
}

type Range struct {
	Step *int64 `json:"step,omitempty"`
}

type RecurringCharge struct {
	RecurringChargeFrequency *string `json:"recurringChargeFrequency,omitempty"`
}

type ReservedDBInstance struct {
	CurrencyCode *string `json:"currencyCode,omitempty"`

	DBInstanceClass *string `json:"dbInstanceClass,omitempty"`

	LeaseID *string `json:"leaseID,omitempty"`

	MultiAZ *bool `json:"multiAZ,omitempty"`

	OfferingType *string `json:"offeringType,omitempty"`

	ProductDescription *string `json:"productDescription,omitempty"`

	ReservedDBInstanceARN *string `json:"reservedDBInstanceARN,omitempty"`

	ReservedDBInstanceID *string `json:"reservedDBInstanceID,omitempty"`

	ReservedDBInstancesOfferingID *string `json:"reservedDBInstancesOfferingID,omitempty"`

	StartTime *metav1.Time `json:"startTime,omitempty"`

	State *string `json:"state,omitempty"`
}

type ReservedDBInstancesOffering struct {
	CurrencyCode *string `json:"currencyCode,omitempty"`

	DBInstanceClass *string `json:"dbInstanceClass,omitempty"`

	MultiAZ *bool `json:"multiAZ,omitempty"`

	OfferingType *string `json:"offeringType,omitempty"`

	ProductDescription *string `json:"productDescription,omitempty"`

	ReservedDBInstancesOfferingID *string `json:"reservedDBInstancesOfferingID,omitempty"`
}

type ResourcePendingMaintenanceActions struct {
	ResourceIdentifier *string `json:"resourceIdentifier,omitempty"`
}

type RestoreWindow struct {
	EarliestTime *metav1.Time `json:"earliestTime,omitempty"`

	LatestTime *metav1.Time `json:"latestTime,omitempty"`
}

type ScalingConfiguration struct {
	AutoPause *bool `json:"autoPause,omitempty"`

	MaxCapacity *int64 `json:"maxCapacity,omitempty"`

	MinCapacity *int64 `json:"minCapacity,omitempty"`

	SecondsUntilAutoPause *int64 `json:"secondsUntilAutoPause,omitempty"`

	TimeoutAction *string `json:"timeoutAction,omitempty"`
}

type ScalingConfigurationInfo struct {
	AutoPause *bool `json:"autoPause,omitempty"`

	MaxCapacity *int64 `json:"maxCapacity,omitempty"`

	MinCapacity *int64 `json:"minCapacity,omitempty"`

	SecondsUntilAutoPause *int64 `json:"secondsUntilAutoPause,omitempty"`

	TimeoutAction *string `json:"timeoutAction,omitempty"`
}

type SourceRegion struct {
	Endpoint *string `json:"endpoint,omitempty"`

	RegionName *string `json:"regionName,omitempty"`

	Status *string `json:"status,omitempty"`

	SupportsDBInstanceAutomatedBackupsReplication *bool `json:"supportsDBInstanceAutomatedBackupsReplication,omitempty"`
}

type Subnet struct {
	SubnetIdentifier *string `json:"subnetIdentifier,omitempty"`

	SubnetStatus *string `json:"subnetStatus,omitempty"`
}

type Tag struct {
	Key *string `json:"key,omitempty"`

	Value *string `json:"value,omitempty"`
}

type TargetHealth struct {
	Description *string `json:"description,omitempty"`
}

type Timezone struct {
	TimezoneName *string `json:"timezoneName,omitempty"`
}

type UpgradeTarget struct {
	AutoUpgrade *bool `json:"autoUpgrade,omitempty"`

	Description *string `json:"description,omitempty"`

	Engine *string `json:"engine,omitempty"`

	EngineVersion *string `json:"engineVersion,omitempty"`

	IsMajorVersionUpgrade *bool `json:"isMajorVersionUpgrade,omitempty"`
}

type UserAuthConfig struct {
	Description *string `json:"description,omitempty"`

	SecretARN *string `json:"secretARN,omitempty"`

	UserName *string `json:"userName,omitempty"`
}

type UserAuthConfigInfo struct {
	Description *string `json:"description,omitempty"`

	SecretARN *string `json:"secretARN,omitempty"`

	UserName *string `json:"userName,omitempty"`
}

type VPCSecurityGroupMembership struct {
	Status *string `json:"status,omitempty"`

	VPCSecurityGroupID *string `json:"vpcSecurityGroupID,omitempty"`
}

type VPNDetails struct {
	VPNGatewayIP *string `json:"vpnGatewayIP,omitempty"`

	VPNID *string `json:"vpnID,omitempty"`

	VPNName *string `json:"vpnName,omitempty"`

	VPNState *string `json:"vpnState,omitempty"`

	VPNTunnelOriginatorIP *string `json:"vpnTunnelOriginatorIP,omitempty"`
}

type ValidStorageOptions struct {
	StorageType *string `json:"storageType,omitempty"`

	SupportsStorageAutoscaling *bool `json:"supportsStorageAutoscaling,omitempty"`
}
