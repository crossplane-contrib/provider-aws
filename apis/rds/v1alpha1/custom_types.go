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

// CustomDBParameterGroupParameters are custom parameters for DBParameterGroup
type CustomDBParameterGroupParameters struct {
	// A list of parameters to associate with this DB parameter group.
	// The fields ApplyMethod, ParameterName and ParameterValue are required
	// for every parameter.
	// Note: AWS actually only modifies the ApplyMethod of a parameter,
	// if the ParameterValue changes too.
	// +optional
	Parameters []CustomParameter `json:"parameters,omitempty"`

	// The DB parameter group family name. A DB parameter group can be associated
	// with one and only one DB parameter group family, and can be applied only
	// to a DB instance running a database engine and engine version compatible
	// with that DB parameter group family.
	//
	// To list all of the available parameter group families for a DB engine, use
	// the following command:
	//
	// aws rds describe-db-engine-versions --query "DBEngineVersions[].DBParameterGroupFamily"
	// --engine <engine>
	//
	// For example, to list all of the available parameter group families for the
	// MySQL DB engine, use the following command:
	//
	// aws rds describe-db-engine-versions --query "DBEngineVersions[].DBParameterGroupFamily"
	// --engine mysql
	//
	// The output contains duplicates.
	//
	// The following are the valid DB engine values:
	//
	//    * aurora (for MySQL 5.6-compatible Aurora)
	//
	//    * aurora-mysql (for MySQL 5.7-compatible Aurora)
	//
	//    * aurora-postgresql
	//
	//    * mariadb
	//
	//    * mysql
	//
	//    * oracle-ee
	//
	//    * oracle-ee-cdb
	//
	//    * oracle-se2
	//
	//    * oracle-se2-cdb
	//
	//    * postgres
	//
	//    * sqlserver-ee
	//
	//    * sqlserver-se
	//
	//    * sqlserver-ex
	//
	//    * sqlserver-web
	//
	// One of DBParameterGroupFamily or DBParameterGroupFamilySelector is required.
	//
	// +optional
	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`

	// DBParameterGroupFamilySelector determines DBParameterGroupFamily from
	// the engine and engine version.
	//
	// One of DBParameterGroupFamily or DBParameterGroupFamilySelector is required.
	//
	// Will not be used if DBParameterGroupFamily is already set.
	// +optional
	DBParameterGroupFamilySelector *DBParameterGroupFamilyNameSelector `json:"dbParameterGroupFamilySelector,omitempty"`
}

// CustomDBParameterGroupObservation includes the custom status fields of DBParameterGroup.
type CustomDBParameterGroupObservation struct{}

// CustomDBClusterParameterGroupParameters are custom parameters for DBClusterParameterGroup
type CustomDBClusterParameterGroupParameters struct {
	// A list of parameters to associate with this DB cluster parameter group.
	// The fields ApplyMethod, ParameterName and ParameterValue are required
	// for every parameter.
	// Note: AWS actually only modifies the ApplyMethod of a parameter,
	// if the ParameterValue changes too.
	// +optional
	Parameters []CustomParameter `json:"parameters,omitempty"`

	// The DB cluster parameter group family name. A DB cluster parameter group
	// can be associated with one and only one DB cluster parameter group family,
	// and can be applied only to a DB cluster running a database engine and engine
	// version compatible with that DB cluster parameter group family.
	//
	// Aurora MySQL
	//
	// Example: aurora5.6, aurora-mysql5.7
	//
	// Aurora PostgreSQL
	//
	// Example: aurora-postgresql9.6
	//
	// To list all of the available parameter group families for a DB engine, use
	// the following command:
	//
	// aws rds describe-db-engine-versions --query "DBEngineVersions[].DBParameterGroupFamily"
	// --engine <engine>
	//
	// For example, to list all of the available parameter group families for the
	// Aurora PostgreSQL DB engine, use the following command:
	//
	// aws rds describe-db-engine-versions --query "DBEngineVersions[].DBParameterGroupFamily"
	// --engine aurora-postgresql
	//
	// The output contains duplicates.
	//
	// The following are the valid DB engine values:
	//
	//    * aurora (for MySQL 5.6-compatible Aurora)
	//
	//    * aurora-mysql (for MySQL 5.7-compatible Aurora)
	//
	//    * aurora-postgresql
	//
	// One of DBParameterGroupFamily or DBParameterGroupFamilySelector is required.
	//
	// +optional
	DBParameterGroupFamily *string `json:"dbParameterGroupFamily"`

	// DBParameterGroupFamilySelector determines DBParameterGroupFamily from
	// the engine and engine version.
	//
	// One of DBParameterGroupFamily or DBParameterGroupFamilySelector is required.
	//
	// Will not be used if DBParameterGroupFamily is already set.
	// +optional
	DBParameterGroupFamilySelector *DBParameterGroupFamilyNameSelector `json:"dbParameterGroupFamilySelector,omitempty"`
}

// CustomDBClusterParameterGroupObservation includes the custom status fields of DBClusterParameterGroup.
type CustomDBClusterParameterGroupObservation struct{}

// DBParameterGroupFamilyNameSelector allows determining the family name from the
// database engine and engine version.
type DBParameterGroupFamilyNameSelector struct {
	// Engine is the name of the database engine.
	// +kubebuilder:validation:Required
	Engine string `json:"engine"`

	// EngineVersion is the version of the database engine.
	// If it is nil, the default engine version given by AWS will be used.
	// +optional
	EngineVersion *string `json:"engineVersion,omitempty"`
}

// CustomDBClusterParameters are custom parameters for DBCluster
type CustomDBClusterParameters struct {

	// AutogeneratePassword indicates whether the controller should generate
	// a random password for the master user if one is not provided via
	// MasterUserPasswordSecretRef.
	//
	// If a password is generated, it will
	// be stored as a secret at the location specified by MasterUserPasswordSecretRef.
	// +optional
	AutogeneratePassword bool `json:"autogeneratePassword,omitempty"`

	// The version number of the database engine to use.
	//
	// To list all of the available engine versions for MySQL 5.6-compatible Aurora,
	// use the following command:
	//
	// aws rds describe-db-engine-versions --engine aurora --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for MySQL 5.7-compatible and
	// MySQL 8.0-compatible Aurora, use the following command:
	//
	// aws rds describe-db-engine-versions --engine aurora-mysql --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for Aurora PostgreSQL, use the
	// following command:
	//
	// aws rds describe-db-engine-versions --engine aurora-postgresql --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for RDS for MySQL, use the following
	// command:
	//
	// aws rds describe-db-engine-versions --engine mysql --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for RDS for PostgreSQL, use
	// the following command:
	//
	// aws rds describe-db-engine-versions --engine postgres --query "DBEngineVersions[].EngineVersion"
	//
	// Aurora MySQL
	//
	// For information, see MySQL on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Updates.html)
	// in the Amazon Aurora User Guide.
	//
	// Aurora PostgreSQL
	//
	// For information, see Amazon Aurora PostgreSQL releases and engine versions
	// (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraPostgreSQL.Updates.20180305.html)
	// in the Amazon Aurora User Guide.
	//
	// MySQL
	//
	// For information, see MySQL on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MySQL.html#MySQL.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	//
	// PostgreSQL
	//
	// For information, see Amazon RDS for PostgreSQL versions and extensions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts)
	// in the Amazon RDS User Guide.
	//
	// Note: Downgrades are not allowed by AWS and attempts to set a lower version
	// will be ignored.
	//
	// Valid for: Aurora DB clusters and Multi-AZ DB clusters
	EngineVersion *string `json:"engineVersion,omitempty"`

	// DomainIAMRoleNameRef is a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	DomainIAMRoleNameRef *xpv1.Reference `json:"domainIAMRoleNameRef,omitempty"`

	// DomainIAMRoleNameSelector selects a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	DomainIAMRoleNameSelector *xpv1.Selector `json:"domainIAMRoleNameSelector,omitempty"`

	// KMSKeyIDRef is a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIDRef,omitempty"`

	// KMSKeyIDSelector selects a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIDSelector,omitempty"`

	// The password for the master database user. This password can contain any
	// printable ASCII character except "/", """, or "@".
	//
	// This parameter will be required in the following scenarios:
	// - The first cluster for a global Aurora cluster
	// - Any cluster as long as it doesn't belong to a global Aurora cluster
	//
	// This parameter is required for creation of a primary cluster. However, it is not required when attaching a secondary regional cluster to an existing global cluster.
	//
	// Constraints: Must contain from 8 to 41 characters.
	MasterUserPasswordSecretRef *xpv1.SecretKeySelector `json:"masterUserPasswordSecretRef,omitempty"`

	// A list of VPC security groups that the DB cluster will belong to.
	//
	// Valid for: Aurora DB clusters and Multi-AZ DB clusters
	VPCSecurityGroupIDs []string `json:"vpcSecurityGroupIDs,omitempty"`

	// VPCSecurityGroupIDRefs are references to VPCSecurityGroups used to set
	// the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDRefs []xpv1.Reference `json:"vpcSecurityGroupIDRefs,omitempty"`

	// VPCSecurityGroupIDSelector selects references to VPCSecurityGroups used
	// to set the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDSelector *xpv1.Selector `json:"vpcSecurityGroupIDSelector,omitempty"`

	// DBSubnetGroupNameRef is a reference to a DBSubnetGroup used to set
	// DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameRef *xpv1.Reference `json:"dbSubnetGroupNameRef,omitempty"`

	// DBSubnetGroupNameSelector selects a reference to a DBSubnetGroup used to
	// set DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameSelector *xpv1.Selector `json:"dbSubnetGroupNameSelector,omitempty"`

	// The DB cluster snapshot identifier of the new DB cluster snapshot created
	// when SkipFinalSnapshot is disabled.
	//
	// Specifying this parameter and also skipping the creation of a final DB cluster
	// snapshot with the SkipFinalShapshot parameter results in an error.
	//
	// Constraints:
	//
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//
	//    * First character must be a letter
	//
	//    * Can't end with a hyphen or contain two consecutive hyphens
	// +immutable
	// +optional
	FinalDBSnapshotIdentifier string `json:"finalDBSnapshotIdentifier,omitempty"`

	// A value that indicates whether to skip the creation of a final DB cluster
	// snapshot before the DB cluster is deleted. If skip is specified, no DB cluster
	// snapshot is created. If skip isn't specified, a DB cluster snapshot is created
	// before the DB cluster is deleted. By default, skip isn't specified, and the
	// DB cluster snapshot is created. By default, this parameter is disabled.
	//
	// You must specify a FinalDBSnapshotIdentifier parameter if SkipFinalSnapshot
	// is disabled.
	// +immutable
	// +optional
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,omitempty"`

	// DBClusterParameterGroupNameRef is a reference to a DBClusterParameterGroup used to set
	// DBClusterParameterGroupName.
	// +optional
	DBClusterParameterGroupNameRef *xpv1.Reference `json:"dbClusterParameterGroupNameRef,omitempty"`

	// DBClusterParameterGroupNameSelector selects a reference to a DBClusterParameterGroup used to
	// set DBClusterParameterGroupName.
	// +optional
	DBClusterParameterGroupNameSelector *xpv1.Selector `json:"dbClusterParameterGroupNameSelector,omitempty"`

	// A value that indicates whether the modifications in this request and any
	// pending modifications are asynchronously applied as soon as possible, regardless
	// of the PreferredMaintenanceWindow setting for the DB cluster. If this parameter
	// is disabled, changes to the DB cluster are applied during the next maintenance
	// window.
	//
	// The ApplyImmediately parameter only affects the EnableIAMDatabaseAuthentication,
	// MasterUserPassword values. If the ApplyImmediately
	// parameter is disabled, then changes to the EnableIAMDatabaseAuthentication,
	// MasterUserPassword values are applied during
	// the next maintenance window. All other changes are applied immediately, regardless
	// of the value of the ApplyImmediately parameter.
	//
	// By default, this parameter is disabled.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// RestoreFrom specifies the details of the backup to restore when creating a new DBCluster.
	// +optional
	RestoreFrom *RestoreDBClusterBackupConfiguration `json:"restoreFrom,omitempty"`
}

// CustomDBClusterObservation includes the custom status fields of DBCluster.
type CustomDBClusterObservation struct{}

// S3RestoreBackupConfiguration defines the details of the S3 backup to restore from.
type S3RestoreBackupConfiguration struct {
	// BucketName is the name of the S3 bucket containing the backup to restore.
	BucketName *string `json:"bucketName"`

	// IngestionRoleARN is the IAM role RDS can assume that will allow it to access the contents of the S3 bucket.
	IngestionRoleARN *string `json:"ingestionRoleARN"`

	// Prefix is the path prefix of the S3 bucket within which the backup to restore is located.
	// +optional
	Prefix *string `json:"prefix,omitempty"`

	// SourceEngine is the engine used to create the backup.
	// Must be "mysql".
	SourceEngine *string `json:"sourceEngine"`

	// SourceEngineVersion is the version of the engine used to create the backup.
	// Example: "5.7.30"
	SourceEngineVersion *string `json:"sourceEngineVersion"`
}

// SnapshotRestoreBackupConfiguration defines the details of the snapshot to restore from.
type SnapshotRestoreBackupConfiguration struct {
	// SnapshotIdentifier is the identifier of the snapshot to restore.
	SnapshotIdentifier *string `json:"snapshotIdentifier"`
}

// PointInTimeRestoreBackupConfiguration defines the details of the time to restore from
type PointInTimeRestoreBackupConfiguration struct {
	// RestoreTime is the date and time (UTC) to restore from.
	// Must be before the latest restorable time for the DB instance.
	// Can't be specified if the useLatestRestorableTime parameter is enabled.
	// Example: 2011-09-07T23:45:00Z
	// +optional
	RestoreTime *metav1.Time `json:"restoreTime"`

	// UseLatestRestorableTime indicates that the DB instance is restored from the latest backup
	// Can't be specified if the restoreTime parameter is provided.
	// +optional
	UseLatestRestorableTime bool `json:"useLatestRestorableTime"`

	// SourceDBInstanceAutomatedBackupsArn specifies the Amazon Resource Name (ARN) of the replicated automated backups
	// from which to restore. Example: arn:aws:rds:useast-1:123456789012:auto-backup:ab-L2IJCEXJP7XQ7HOJ4SIEXAMPLE
	// +optional
	SourceDBInstanceAutomatedBackupsArn *string `json:"sourceDBInstanceAutomatedBackupsArn"`

	// SourceDBInstanceIdentifier specifies the identifier of the source DB instance from which to restore. Constraints:
	// Must match the identifier of an existing DB instance.
	// +optional
	SourceDBInstanceIdentifier *string `json:"sourceDBInstanceIdentifier"`

	// SourceDbiResourceID specifies the resource ID of the source DB instance from which to restore.
	// +optional
	SourceDbiResourceID *string `json:"sourceDbiResourceId"`
}

// PointInTimeRestoreDBClusterBackupConfiguration defines the details of the time to restore from
type PointInTimeRestoreDBClusterBackupConfiguration struct {
	// RestoreTime is the date and time (UTC) to restore from.
	// Must be before the latest restorable time for the DB instance.
	// Can't be specified if the useLatestRestorableTime parameter is enabled.
	// Example: 2011-09-07T23:45:00Z
	// +optional
	RestoreTime *metav1.Time `json:"restoreTime"`

	// UseLatestRestorableTime indicates that the DB instance is restored from the latest backup
	// Can't be specified if the restoreTime parameter is provided.
	// +optional
	UseLatestRestorableTime bool `json:"useLatestRestorableTime"`

	// SourceDBInstanceAutomatedBackupsArn specifies the Amazon Resource Name (ARN) of the replicated automated backups
	// from which to restore. Example: arn:aws:rds:useast-1:123456789012:auto-backup:ab-L2IJCEXJP7XQ7HOJ4SIEXAMPLE
	// +optional
	SourceDBInstanceAutomatedBackupsArn *string `json:"sourceDBInstanceAutomatedBackupsArn"`

	// SourceDBClusterIdentifier specifies the identifier of the source DB cluster from which to restore. Constraints:
	// Must match the identifier of an existing DB instance.
	// +optional
	SourceDBClusterIdentifier *string `json:"sourceDBClusterIdentifier"`

	// SourceDbiResourceID specifies the resource ID of the source DB instance from which to restore.
	// +optional
	SourceDbiResourceID *string `json:"sourceDbiResourceId"`

	// The type of restore to be performed. You can specify one of the following
	// values:
	//
	//    * full-copy - The new DB cluster is restored as a full copy of the source
	//    DB cluster.
	//
	//    * copy-on-write - The new DB cluster is restored as a clone of the source
	//    DB cluster.
	//
	// Constraints: You can't specify copy-on-write if the engine version of the
	// source DB cluster is earlier than 1.11.
	//
	// If you don't specify a RestoreType value, then the new DB cluster is restored
	// as a full copy of the source DB cluster.
	//
	// Valid for: Aurora DB clusters and Multi-AZ DB clusters
	// +optional
	// +kubebuilder:validation:Enum=full-copy;copy-on-write
	RestoreType *string `json:"restoreType"`
}

// RestoreDBInstanceBackupConfiguration defines the backup to restore a new DBCluster from.
type RestoreDBInstanceBackupConfiguration struct {
	// S3 specifies the details of the S3 backup to restore from.
	// +optional
	S3 *S3RestoreBackupConfiguration `json:"s3,omitempty"`

	// Snapshot specifies the details of the snapshot to restore from.
	// +optional
	Snapshot *SnapshotRestoreBackupConfiguration `json:"snapshot,omitempty"`

	// PointInTime specifies the details of the point in time restore.
	// +optional
	PointInTime *PointInTimeRestoreBackupConfiguration `json:"pointInTime,omitempty"`

	// Source is the type of the backup to restore when creating a new  DBCluster or DBInstance.
	// S3, Snapshot and PointInTime are supported.
	// +kubebuilder:validation:Enum=S3;Snapshot;PointInTime
	Source *string `json:"source"`
}

// RestoreDBClusterBackupConfiguration defines the backup to restore a new DBCluster from.
type RestoreDBClusterBackupConfiguration struct {
	// S3 specifies the details of the S3 backup to restore from.
	// +optional
	S3 *S3RestoreBackupConfiguration `json:"s3,omitempty"`

	// Snapshot specifies the details of the snapshot to restore from.
	// +optional
	Snapshot *SnapshotRestoreBackupConfiguration `json:"snapshot,omitempty"`

	// PointInTime specifies the details of the point in time restore.
	// +optional
	PointInTime *PointInTimeRestoreDBClusterBackupConfiguration `json:"pointInTime,omitempty"`

	// Source is the type of the backup to restore when creating a new  DBCluster or DBInstance.
	// S3, Snapshot and PointInTime are supported.
	// +kubebuilder:validation:Enum=S3;Snapshot;PointInTime
	Source *string `json:"source"`
}

// CustomGlobalClusterParameters are custom parameters for a GlobalCluster
type CustomGlobalClusterParameters struct {
	// SourceDBClusterIdentifierRef is a reference to a DBCluster used to set
	// SourceDBClusterIdentifier.
	// +immutable
	// +optional
	SourceDBClusterIdentifierRef *xpv1.Reference `json:"sourceDBClusterIdentifierRef,omitempty"`

	// SourceDBClusterIdentifierSelector selects a reference to a DBCluster used to
	// set SourceDBClusterIdentifier.
	// +immutable
	// +optional
	SourceDBClusterIdentifierSelector *xpv1.Selector `json:"sourceDBClusterIdentifierSelector,omitempty"`
}

// CustomGlobalClusterObservation includes the custom status fields of GlobalCluster.
type CustomGlobalClusterObservation struct{}

// CustomDBInstanceParameters are custom parameters for the DBInstance
type CustomDBInstanceParameters struct {
	// AutogeneratePassword indicates whether the controller should generate
	// a random password for the master user if one is not provided via
	// MasterUserPasswordSecretRef.
	//
	// If a password is generated, it will
	// be stored as a secret at the location specified by MasterUserPasswordSecretRef.
	// +optional
	AutogeneratePassword bool `json:"autogeneratePassword,omitempty"`

	// DBClusterIdentifierRef is a reference to a DBCluster used to set
	// DBClusterIdentifier.
	// +immutable
	// +optional
	DBClusterIdentifierRef *xpv1.Reference `json:"dbClusterIdentifierRef,omitempty"`

	// DBClusterIdentifierSelector selects a reference to a DBCluster used to
	// set DBClusterIdentifier.
	// +immutable
	// +optional
	DBClusterIdentifierSelector *xpv1.Selector `json:"dbClusterIdentifierSelector,omitempty"`

	// A list of database security groups to associate with this DB instance
	DBSecurityGroups []string `json:"dbSecurityGroups,omitempty"`

	// DBSubnetGroupNameRef is a reference to a DBSubnetGroup used to set
	// DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameRef *xpv1.Reference `json:"dbSubnetGroupNameRef,omitempty"`

	// DBSubnetGroupNameSelector selects a reference to a DBSubnetGroup used to
	// set DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameSelector *xpv1.Selector `json:"dbSubnetGroupNameSelector,omitempty"`

	// DomainIAMRoleNameRef is a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	// +immutable
	DomainIAMRoleNameRef *xpv1.Reference `json:"domainIAMRoleNameRef,omitempty"`

	// DomainIAMRoleNameSelector selects a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	// +immutable
	DomainIAMRoleNameSelector *xpv1.Selector `json:"domainIAMRoleNameSelector,omitempty"`

	// The version number of the database engine to use.
	//
	// For a list of valid engine versions, use the DescribeDBEngineVersions operation.
	//
	// The following are the database engines and links to information about the
	// major and minor versions that are available with Amazon RDS. Not every database
	// engine is available for every Amazon Web Services Region.
	//
	// Amazon Aurora
	//
	// Not applicable. The version number of the database engine to be used by the
	// DB instance is managed by the DB cluster.
	//
	// Amazon RDS Custom for Oracle
	//
	// A custom engine version (CEV) that you have previously created. This setting
	// is required for RDS Custom for Oracle. The CEV name has the following format:
	// 19.customized_string. A valid CEV name is 19.my_cev1. For more information,
	// see Creating an RDS Custom for Oracle DB instance (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-creating.html#custom-creating.create)
	// in the Amazon RDS User Guide.
	//
	// Amazon RDS Custom for SQL Server
	//
	// See RDS Custom for SQL Server general requirements (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-reqs-limits-MS.html)
	// in the Amazon RDS User Guide.
	//
	// MariaDB
	//
	// For information, see MariaDB on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MariaDB.html#MariaDB.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	//
	// Microsoft SQL Server
	//
	// For information, see Microsoft SQL Server Versions on Amazon RDS (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.VersionSupport)
	// in the Amazon RDS User Guide.
	//
	// MySQL
	//
	// For information, see MySQL on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MySQL.html#MySQL.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	//
	// Oracle
	//
	// For information, see Oracle Database Engine Release Notes (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.Oracle.PatchComposition.html)
	// in the Amazon RDS User Guide.
	//
	// PostgreSQL
	//
	// For information, see Amazon RDS for PostgreSQL versions and extensions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts)
	// in the Amazon RDS User Guide.
	//
	// Note: Downgrades are not allowed by AWS and attempts to set a lower version
	// will be ignored.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The DB instance snapshot identifier of the new DB instance snapshot created
	// when SkipFinalSnapshot is disabled.
	//
	// Specifying this parameter and also skipping the creation of a final DB instance
	// snapshot with the SkipFinalShapshot parameter results in an error.
	//
	// Constraints:
	//
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//
	//    * First character must be a letter
	//
	//    * Can't end with a hyphen or contain two consecutive hyphens
	// +immutable
	// +optional
	FinalDBSnapshotIdentifier string `json:"finalDBSnapshotIdentifier,omitempty"`

	// The password for the master database user. This password can contain any
	// printable ASCII character except "/", """, or "@".
	//
	// Constraints: Must contain from 8 to 41 characters.
	// +optional
	MasterUserPasswordSecretRef *xpv1.SecretKeySelector `json:"masterUserPasswordSecretRef,omitempty"`

	// MonitoringRoleARNRef is a reference to an IAMRole used to set
	// MonitoringRoleARN.
	// +optional
	// +immutable
	MonitoringRoleARNRef *xpv1.Reference `json:"monitoringRoleArnRef,omitempty"`

	// MonitoringRoleARNSelector selects a reference to an IAMRole used to set
	// MonitoringRoleARN.
	// +optional
	// +immutable
	MonitoringRoleARNSelector *xpv1.Selector `json:"monitoringRoleArnSelector,omitempty"`

	// A value that indicates whether to skip the creation of a final DB instance
	// snapshot before the DB instance is deleted. If skip is specified, no DB instance
	// snapshot is created. If skip isn't specified, a DB instance snapshot is created
	// before the DB instance is deleted. By default, skip isn't specified, and the
	// DB instance snapshot is created. By default, this parameter is disabled.
	//
	// You must specify a FinalDBSnapshotIdentifier parameter if SkipFinalSnapshot
	// is disabled.
	// +immutable
	// +optional
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,omitempty"`

	// A list of Amazon EC2 VPC security groups to authorize on this DB instance.
	// This change is asynchronously applied as soon as possible.
	//
	// This setting doesn't apply to RDS Custom.
	//
	// Amazon Aurora
	// Not applicable. The associated list of EC2 VPC security groups is managed
	// by the DB cluster. For more information, see ModifyDBCluster.
	//
	// Constraints:
	//    * If supplied, must match existing VpcSecurityGroupIds.
	VPCSecurityGroupIDs []string `json:"vpcSecurityGroupIDs,omitempty"`

	// VPCSecurityGroupIDRefs are references to VPCSecurityGroups used to set
	// the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDRefs []xpv1.Reference `json:"vpcSecurityGroupIDRefs,omitempty"`

	// VPCSecurityGroupIDSelector selects references to VPCSecurityGroups used
	// to set the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDSelector *xpv1.Selector `json:"vpcSecurityGroupIDSelector,omitempty"`

	// DBParameterGroupNameRef is a reference to a DBParameterGroup used to set
	// DBParameterGroupName.
	// +optional
	DBParameterGroupNameRef *xpv1.Reference `json:"dbParameterGroupNameRef,omitempty"`

	// DBParameterGroupNameSelector selects a reference to a DBParameterGroup used to
	// set DBParameterGroupName.
	// +optional
	DBParameterGroupNameSelector *xpv1.Selector `json:"dbParameterGroupNameSelector,omitempty"`

	// A value that indicates whether the modifications in this request and any
	// pending modifications are asynchronously applied as soon as possible, regardless
	// of the PreferredMaintenanceWindow setting for the DB instance. By default,
	// this parameter is disabled.
	//
	// If this parameter is disabled, changes to the DB instance are applied during
	// the next maintenance window. Some parameter changes can cause an outage and
	// are applied on the next call to RebootDBInstance, or the next failure reboot.
	// Review the table of parameters in Modifying a DB Instance (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
	// in the Amazon RDS User Guide. to see the impact of enabling or disabling
	// ApplyImmediately for each modified parameter and to determine when the changes
	// are applied.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// KMSKeyIDRef is a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIDRef,omitempty"`

	// KMSKeyIDSelector selects a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIDSelector,omitempty"`

	// RestoreFrom specifies the details of the backup to restore when creating a new DBInstance.
	// +optional
	RestoreFrom *RestoreDBInstanceBackupConfiguration `json:"restoreFrom,omitempty"`

	// DeleteAutomatedBackups indicates whether to remove automated backups
	// immediately after the DB instance is deleted. The default is to
	// remove automated backups immediately after the DB instance is
	// deleted.
	// +optional
	DeleteAutomatedBackups *bool `json:"deleteAutomatedBackups,omitempty"`
}

// CustomDBInstanceObservation includes the custom status fields of DBInstance.
type CustomDBInstanceObservation struct{}

// CustomDBInstanceRoleAssociationParameters are custom parameters for the DBInstanceRoleAssociation
type CustomDBInstanceRoleAssociationParameters struct {
	// The name of the DB instance to associate the IAM role with.
	// +crossplane:generate:reference:type=DBInstance
	// +optional
	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	// DBInstanceIdentifierRef is a reference to a DBInstance used to set
	// the DBInstanceIdentifier.
	// +optional
	DBInstanceIdentifierRef *xpv1.Reference `json:"dbInstanceIdentifierRef,omitempty"`

	// DBInstanceIdentifierSelector selects references to a DBInstance used
	// to set the DBInstanceIdentifier.
	// +optional
	DBInstanceIdentifierSelector *xpv1.Selector `json:"dbInstanceIdentifierSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role to associate with the DB instance,
	// for example arn:aws:iam::123456789012:role/AccessRole.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +optional
	RoleARN *string `json:"roleArn,omitempty"`

	// RoleARNRef is a reference to a IAM Role used to set
	// RoleARN.
	// +optional
	RoleARNRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleARNSelector selects a reference to a IAM Role used to
	// set RoleARN.
	// +optional
	RoleARNSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`
}

// CustomDBInstanceRoleAssociationObservation includes the custom status fields of DBInstanceRoleAssociation.
type CustomDBInstanceRoleAssociationObservation struct{}

// CustomOptionGroupParameters are custom parameters for the OptionGroup
type CustomOptionGroupParameters struct {
	// Option in this list are added to the option group or, if already present,
	// the specified configuration is used to update the existing configuration.
	Option []*CustomOptionConfiguration `json:"option,omitempty"`

	// A value that indicates whether to apply the change immediately or during
	// the next maintenance window for each instance associated with the option
	// group.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`
}

// CustomOptionGroupObservation includes the custom status fields of OptionGroup.
type CustomOptionGroupObservation struct{}

// CustomOptionConfiguration are custom parameters for the OptionConfiguration
type CustomOptionConfiguration struct {
	DBSecurityGroupMemberships []*string `json:"dbSecurityGroupMemberships,omitempty"`

	OptionName *string `json:"optionName,omitempty"`

	OptionSettings []*CustomOptionGroupOptionSetting `json:"optionSettings,omitempty"`

	OptionVersion *string `json:"optionVersion,omitempty"`

	Port *int64 `json:"port,omitempty"`

	VPCSecurityGroupMemberships []*string `json:"vpcSecurityGroupMemberships,omitempty"`
}

// CustomOptionGroupOptionSetting are custom parameters for the OptionGroupOptionSetting
type CustomOptionGroupOptionSetting struct {
	Name *string `json:"name,omitempty"`

	Value *string `json:"value,omitempty"`
}

// CustomParameter are custom parameters for the Parameter
type CustomParameter struct {
	// The apply method of the parameter.
	// AWS actually only modifies to value set here, if the parameter value changes too.
	// +kubebuilder:validation:Enum=immediate;pending-reboot
	// +kubebuilder:validation:Required
	ApplyMethod *string `json:"applyMethod"`

	// The name of the parameter.
	// +kubebuilder:validation:Required
	ParameterName *string `json:"parameterName"`

	// The value of the parameter.
	// +kubebuilder:validation:Required
	ParameterValue *string `json:"parameterValue"`
}
