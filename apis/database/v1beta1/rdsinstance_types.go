/*
Copyright 2019 The Crossplane Authors.

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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SQL database engines.
const (
	MysqlEngine      = "mysql"
	PostgresqlEngine = "postgres"
)

// Tag is a metadata assigned to an Amazon RDS resource consisting of a key-value pair.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/Tag
type Tag struct {
	// A key is the required name of the tag. The string value can be from 1 to
	// 128 Unicode characters in length and can't be prefixed with "aws:" or "rds:".
	// The string can only contain only the set of Unicode letters, digits, white-space,
	// '_', '.', '/', '=', '+', '-' (Java regex: "^([\\p{L}\\p{Z}\\p{N}_.:/=+\\-]*)$").
	Key string `json:"key,omitempty"`

	// A value is the optional value of the tag. The string value can be from 1
	// to 256 Unicode characters in length and can't be prefixed with "aws:" or
	// "rds:". The string can only contain only the set of Unicode letters, digits,
	// white-space, '_', '.', '/', '=', '+', '-' (Java regex: "^([\\p{L}\\p{Z}\\p{N}_.:/=+\\-]*)$").
	Value string `json:"value,omitempty"`
}

// ProcessorFeature is a processor feature entry. For more information, see
// Configuring the Processor of the DB Instance Class
// (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html#USER_ConfigureProcessor)
// in the Amazon RDS User Guide.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/ProcessorFeature
type ProcessorFeature struct {
	// Name of the processor feature. Valid names are coreCount and threadsPerCore.
	Name string `json:"name"`

	// Value of a processor feature name.
	Value string `json:"value"`
}

// CloudwatchLogsExportConfiguration is the configuration setting for the log types to be enabled for export to CloudWatch
// Logs for a specific DB instance or DB cluster.
// The EnableLogTypes and DisableLogTypes arrays determine which logs will be
// exported (or not exported) to CloudWatch Logs. The values within these arrays
// depend on the DB engine being used. For more information, see Publishing
// Database Logs to Amazon CloudWatch Logs  (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.html#USER_LogAccess.Procedural.UploadtoCloudWatch)
// in the Amazon RDS User Guide.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/CloudwatchLogsExportConfiguration
type CloudwatchLogsExportConfiguration struct {
	// DisableLogTypes is the list of log types to disable.
	// +immutable
	DisableLogTypes []string `json:"disableLogTypes,omitempty"`

	// EnableLogTypes is the list of log types to enable.
	// +immutable
	EnableLogTypes []string `json:"enableLogTypes,omitempty"`
} // TODO: remove deprecated field + code. Mapping to EnableCloudwatchLogsExports while in deprecation.

// ScalingConfiguration contains the scaling configuration of an Aurora Serverless DB cluster.
// For more information, see Using Amazon Aurora Serverless (http://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless.html)
// in the Amazon Aurora User Guide.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/ScalingConfiguration
type ScalingConfiguration struct {
	// AutoPause specifies whether to allow or disallow automatic pause for an
	// Aurora DB cluster in serverless DB engine mode. A DB cluster can be paused
	// only when it's idle (it has no connections).
	// If a DB cluster is paused for more than seven days, the DB cluster might
	// be backed up with a snapshot. In this case, the DB cluster is restored when
	// there is a request to connect to it.
	// +optional
	AutoPause *bool `json:"autoPause,omitempty"`

	// MaxCapacity is the maximum capacity for an Aurora DB cluster in serverless DB engine mode.
	// Valid capacity values are 2, 4, 8, 16, 32, 64, 128, and 256.
	// The maximum capacity must be greater than or equal to the minimum capacity.
	// +optional
	MaxCapacity *int `json:"maxCapacity,omitempty"`

	// MinCapacity is the minimum capacity for an Aurora DB cluster in serverless DB engine mode.
	// Valid capacity values are 2, 4, 8, 16, 32, 64, 128, and 256.
	// The minimum capacity must be less than or equal to the maximum capacity.
	// +optional
	MinCapacity *int `json:"minCapacity,omitempty"`

	// SecondsUntilAutoPause is the time, in seconds, before an Aurora DB cluster in serverless mode is paused.
	// +optional
	SecondsUntilAutoPause *int `json:"secondsUntilAutoPause,omitempty"`
}

// S3RestoreBackupConfiguration defines the details of the S3 backup to restore from.
type S3RestoreBackupConfiguration struct {
	// BucketName is the name of the S3 bucket containing the backup to restore.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.Bucket
	BucketName *string `json:"bucketName"`

	// BucketNameRef is a reference to a Bucket used to set
	// BucketName.
	// +immutable
	// +optional
	BucketNameRef *xpv1.Reference `json:"bucketNameRef,omitempty"`

	// BucketNameSelector selects a reference to a Bucket used to
	// set BucketName.
	// +immutable
	// +optional
	BucketNameSelector *xpv1.Selector `json:"bucketNameSelector,omitempty"`

	// IngestionRoleARN is the IAM role RDS can assume that will allow it to access the contents of the S3 bucket.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	IngestionRoleARN *string `json:"ingestionRoleARN,omitempty"`

	// IngestionRoleARNRef is a reference to a IAM Role used to set
	// IngestionRoleARN.
	// +immutable
	// +optional
	IngestionRoleARNRef *xpv1.Reference `json:"ingestionRoleARNRef,omitempty"`

	// IngestionRoleARNSelector selects a reference to a IAM Role used to
	// set IngestionRoleARN.
	// +immutable
	// +optional
	IngestionRoleARNSelector *xpv1.Selector `json:"ingestionRoleARNSelector,omitempty"`

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

// SnapshotRestoreBackupConfiguration defines the details of the database snapshot to restore from.
type SnapshotRestoreBackupConfiguration struct {
	// SnapshotIdentifier is the identifier of the database snapshot to restore.
	SnapshotIdentifier *string `json:"snapshotIdentifier"`
}

// PointInTimeRestoreBackupConfiguration defines the details of the point in time to restore from.
// restoreTime or useLatestRestorableTime must be defined together with one of sourceDBInstanceAutomatedBackupArn,
// sourceDBInstanceIdentifier or sourceDbiResourceId.
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

// RestoreBackupConfiguration defines the backup to restore a new RDS instance from.
type RestoreBackupConfiguration struct {
	// S3 specifies the details of the S3 backup to restore from.
	// +optional
	S3 *S3RestoreBackupConfiguration `json:"s3,omitempty"`

	// Snapshot specifies the details of the database snapshot to restore from.
	// +optional
	Snapshot *SnapshotRestoreBackupConfiguration `json:"snapshot,omitempty"`

	// PointInTime specifies the details of the point in time restore.
	// +optional
	PointInTime *PointInTimeRestoreBackupConfiguration `json:"pointInTime,omitempty"`

	// Source is the type of the backup to restore when creating a new RDS instance.
	// S3, Snapshot and PointInTime are supported.
	// +kubebuilder:validation:Enum=S3;Snapshot;PointInTime
	Source *string `json:"source"`
}

// RDSInstanceParameters define the desired state of an AWS Relational Database
// Service instance.
type RDSInstanceParameters struct {
	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your RDSInstance to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// AllocatedStorage is the amount of storage (in gibibytes) to allocate for the DB instance.
	// Type: Integer
	// Amazon Aurora
	// Not applicable. Aurora cluster volumes automatically grow as the amount of
	// data in your database increases, though you are only charged for the space
	// that you use in an Aurora cluster volume.
	// MySQL
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 16384.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// MariaDB
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 16384.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// PostgreSQL
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 16384.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// Oracle
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 16384.
	//    * Magnetic storage (standard): Must be an integer from 10 to 3072.
	// SQL Server
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2):
	// Enterprise and Standard editions: Must be an integer from 200 to 16384.
	// Web and Express editions: Must be an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1):
	// Enterprise and Standard editions: Must be an integer from 200 to 16384.
	// Web and Express editions: Must be an integer from 100 to 16384.
	//    * Magnetic storage (standard):
	// Enterprise and Standard editions: Must be an integer from 200 to 1024.
	// Web and Express editions: Must be an integer from 20 to 1024.
	// +optional
	AllocatedStorage *int `json:"allocatedStorage,omitempty"`

	// AutoMinorVersionUpgrade indicates that minor engine upgrades are applied automatically to the DB
	// instance during the maintenance window.
	// Default: true
	// +immutable
	// +optional
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// AvailabilityZone is the EC2 Availability Zone that the DB instance is created in. For information
	// on AWS Regions and Availability Zones, see Regions and Availability Zones
	// (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).
	// Default: A random, system-chosen Availability Zone in the endpoint's AWS
	// Region.
	// Example: us-east-1d
	// Constraint: The AvailabilityZone parameter is ignored if the MultiAZ
	// is set to true.
	// The specified Availability Zone must be in the
	// same AWS Region as the current endpoint.
	// +immutable
	// +optional
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// BackupRetentionPeriod is the number of days for which automated backups are retained. Setting this
	// parameter to a positive number enables backups. Setting this parameter to
	// 0 disables automated backups.
	// Amazon Aurora
	// Not applicable. The retention period for automated backups is managed by
	// the DB cluster. For more information, see CreateDBCluster.
	// Default: 1
	// Constraints:
	//    * Must be a value from 0 to 35
	//    * Cannot be set to 0 if the DB instance is a source to Read Replicas
	// +optional
	BackupRetentionPeriod *int `json:"backupRetentionPeriod,omitempty"`

	// CACertificateIdentifier indicates the certificate that needs to be associated with the instance.
	// +optional
	CACertificateIdentifier *string `json:"caCertificateIdentifier,omitempty"`

	// CharacterSetName indicates that the DB instance should be associated
	// with the specified CharacterSet for supported engines,
	// Amazon Aurora
	// Not applicable. The character set is managed by the DB cluster. For more
	// information, see CreateDBCluster.
	// +immutable
	// +optional
	CharacterSetName *string `json:"characterSetName,omitempty"`

	// CopyTagsToSnapshot should be true to copy all tags from the DB instance to snapshots of the DB instance,
	// and otherwise false. The default is false.
	// +optional
	CopyTagsToSnapshot *bool `json:"copyTagsToSnapshot,omitempty"`

	// DBClusterIdentifier is the identifier of the DB cluster that the instance will belong to.
	// For information on creating a DB cluster, see CreateDBCluster.
	// Type: String
	// +immutable
	// +optional
	DBClusterIdentifier *string `json:"dbClusterIdentifier,omitempty"`

	// DBInstanceClass is the compute and memory capacity of the DB instance, for example, db.m4.large.
	// Not all DB instance classes are available in all AWS Regions, or for all
	// database engines. For the full list of DB instance classes, and availability
	// for your engine, see DB Instance Class (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html)
	// in the Amazon RDS User Guide.
	DBInstanceClass string `json:"dbInstanceClass"`

	// DBName is the meaning of this parameter differs according to the database engine you
	// use.
	// Type: String
	// MySQL
	// The name of the database to create when the DB instance is created. If this
	// parameter is not specified, no database is created in the DB instance.
	// Constraints:
	//    * Must contain 1 to 64 letters or numbers.
	//    * Cannot be a word reserved by the specified database engine
	// MariaDB
	// The name of the database to create when the DB instance is created. If this
	// parameter is not specified, no database is created in the DB instance.
	// Constraints:
	//    * Must contain 1 to 64 letters or numbers.
	//    * Cannot be a word reserved by the specified database engine
	// PostgreSQL
	// The name of the database to create when the DB instance is created. If this
	// parameter is not specified, the default "postgres" database is created in
	// the DB instance.
	// Constraints:
	//    * Must contain 1 to 63 letters, numbers, or underscores.
	//    * Must begin with a letter or an underscore. Subsequent characters can
	//    be letters, underscores, or digits (0-9).
	//    * Cannot be a word reserved by the specified database engine
	// Oracle
	// The Oracle System ID (SID) of the created DB instance. If you specify null,
	// the default value ORCL is used. You can't specify the string NULL, or any
	// other reserved word, for DBName.
	// Default: ORCL
	// Constraints:
	//    * Cannot be longer than 8 characters
	// SQL Server
	// Not applicable. Must be null.
	// Amazon Aurora
	// The name of the database to create when the primary instance of the DB cluster
	// is created. If this parameter is not specified, no database is created in
	// the DB instance.
	// Constraints:
	//    * Must contain 1 to 64 letters or numbers.
	//    * Cannot be a word reserved by the specified database engine
	// +immutable
	// +optional
	DBName *string `json:"dbName,omitempty"`

	// DBSecurityGroups is a list of DB security groups to associate with this DB instance.
	// Default: The default DB security group for the database engine.
	// +optional
	DBSecurityGroups []string `json:"dbSecurityGroups,omitempty"`

	// DBSubnetGroupName is a DB subnet group to associate with this DB instance.
	// If there is no DB subnet group, then it is a non-VPC DB instance.
	// +optional
	// +crossplane:generate:reference:type=DBSubnetGroup
	DBSubnetGroupName *string `json:"dbSubnetGroupName,omitempty"`

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

	// DeleteAutomatedBackups indicates whether to remove automated backups
	// immediately after the DB instance is deleted. The default is to
	// remove automated backups immediately after the DB instance is
	// deleted.
	// +optional
	DeleteAutomatedBackups *bool `json:"deleteAutomatedBackups,omitempty"`

	// DeletionProtection indicates if the DB instance should have deletion protection enabled. The
	// database can't be deleted when this value is set to true. The default is
	// false. For more information, see  Deleting a DB Instance (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_DeleteInstance.html).
	// +optional
	DeletionProtection *bool `json:"deletionProtection,omitempty"`

	// EnableCloudwatchLogsExports is the list of log types that need to be enabled for exporting to CloudWatch
	// Logs. The values in the list depend on the DB engine being used. For more
	// information, see Publishing Database Logs to Amazon CloudWatch Logs  (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.html#USER_LogAccess.Procedural.UploadtoCloudWatch)
	// in the Amazon Relational Database Service User Guide.
	// +optional
	EnableCloudwatchLogsExports []string `json:"enableCloudwatchLogsExports"`

	// EnableIAMDatabaseAuthentication should be true to enable mapping of AWS Identity and Access Management (IAM) accounts
	// to database accounts, and otherwise false.
	// You can enable IAM database authentication for the following database engines:
	// Amazon Aurora
	// Not applicable. Mapping AWS IAM accounts to database accounts is managed
	// by the DB cluster. For more information, see CreateDBCluster.
	// MySQL
	//    * For MySQL 5.6, minor version 5.6.34 or higher
	//    * For MySQL 5.7, minor version 5.7.16 or higher
	// Default: false
	// +optional
	EnableIAMDatabaseAuthentication *bool `json:"enableIAMDatabaseAuthentication,omitempty"`

	// EnablePerformanceInsights should be true to enable Performance Insights for the DB instance, and otherwise false.
	// For more information, see Using Amazon Performance Insights (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.html)
	// in the Amazon Relational Database Service User Guide.
	// +optional
	EnablePerformanceInsights *bool `json:"enablePerformanceInsights,omitempty"`

	// Engine is the name of the database engine to be used for this instance.
	// Not every database engine is available for every AWS Region.
	// Valid Values:
	//    * aurora (for MySQL 5.6-compatible Aurora)
	//    * aurora-mysql (for MySQL 5.7-compatible Aurora)
	//    * aurora-postgresql
	//    * mariadb
	//    * mysql
	//    * oracle-ee
	//    * oracle-se2
	//    * oracle-se1
	//    * oracle-se
	//    * postgres
	//    * sqlserver-ee
	//    * sqlserver-se
	//    * sqlserver-ex
	//    * sqlserver-web
	// Engine is a required field
	// +immutable
	Engine string `json:"engine"`

	// EngineVersion is the version number of the database engine to use.
	// For a list of valid engine versions, call DescribeDBEngineVersions.
	// The following are the database engines and links to information about the
	// major and minor versions that are available with Amazon RDS. Not every database
	// engine is available for every AWS Region.
	// Amazon Aurora
	// Not applicable. The version number of the database engine to be used by the
	// DB instance is managed by the DB cluster. For more information, see CreateDBCluster.
	// MariaDB
	// See MariaDB on Amazon RDS Versions (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MariaDB.html#MariaDB.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	// Microsoft SQL Server
	// See Version and Feature Support on Amazon RDS (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.FeatureSupport)
	// in the Amazon RDS User Guide.
	// MySQL
	// See MySQL on Amazon RDS Versions (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MySQL.html#MySQL.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	// Oracle
	// See Oracle Database Engine Release Notes (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.Oracle.PatchComposition.html)
	// in the Amazon RDS User Guide.
	// PostgreSQL
	// See Supported PostgreSQL Database Versions (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts.General.DBVersions)
	// in the Amazon RDS User Guide.
	// +optional
	//
	// Note: Downgrades are not allowed by AWS and attempts to set a lower version
	// will be ignored.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// RestoreFrom specifies the details of the backup to restore when creating a new RDS instance. (If the RDS instance already exists, this property will be ignored.)
	// +optional
	RestoreFrom *RestoreBackupConfiguration `json:"restoreFrom,omitempty"`

	// IOPS is the amount of Provisioned IOPS (input/output operations per second) to be
	// initially allocated for the DB instance. For information about valid IOPS
	// values, see Amazon RDS Provisioned IOPS Storage to Improve Performance
	// (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#USER_PIOPS)
	// in the Amazon RDS User Guide.
	// Constraints: Must be a multiple between 1 and 50 of the storage amount for
	// the DB instance. Must also be an integer multiple of 1000. For example, if
	// the size of your DB instance is 500 GiB, then your IOPS value can be 2000,
	// 3000, 4000, or 5000.
	//
	// For valid IOPS values on DB instances with storage type "gp3",
	// see https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#gp3-storage.
	//
	// Note: controller considers 0 and null as equivalent
	// +optional
	IOPS *int `json:"iops,omitempty"`

	// KMSKeyID for an encrypted DB instance.
	// The KMS key identifier is the Amazon Resource Name (ARN) for the KMS encryption
	// key. If you are creating a DB instance with the same AWS account that owns
	// the KMS encryption key used to encrypt the new DB instance, then you can
	// use the KMS key alias instead of the ARN for the KM encryption key.
	// Amazon Aurora
	// Not applicable. The KMS key identifier is managed by the DB cluster. For
	// more information, see CreateDBCluster.
	// If the StorageEncrypted parameter is true, and you do not specify a value
	// for the KMSKeyID parameter, then Amazon RDS will use your default encryption
	// key. AWS KMS creates the default encryption key for your AWS account. Your
	// AWS account has a different default encryption key for each AWS Region.
	// +immutable
	// +optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// LicenseModel information for this DB instance.
	// Valid values: license-included | bring-your-own-license | general-public-license
	// +optional
	LicenseModel *string `json:"licenseModel,omitempty"`

	// MasterUsername is the name for the master user.
	// Amazon Aurora
	// Not applicable. The name for the master user is managed by the DB cluster.
	// For more information, see CreateDBCluster.
	//
	// Constraints:
	//
	//    * Must be 1 to 16 letters or numbers.
	//
	//    * First character must be a letter.
	//
	//    * Can't be a reserved word for the chosen database engine.
	//
	// +immutable
	// +optional
	MasterUsername *string `json:"masterUsername,omitempty"`

	// MasterPasswordSecretRef references the secret that contains the password used
	// in the creation of this RDS instance. If no reference is given, a password
	// will be auto-generated.
	// +optional
	// +immutable
	MasterPasswordSecretRef *xpv1.SecretKeySelector `json:"masterPasswordSecretRef,omitempty"`

	// The upper limit to which Amazon RDS can automatically scale the storage of
	// the DB instance.
	//
	// For more information about this setting, including limitations that apply
	// to it, see Managing capacity automatically with Amazon RDS storage autoscaling
	// (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PIOPS.StorageTypes.html#USER_PIOPS.Autoscaling)
	// in the Amazon RDS User Guide.
	// +optional
	MaxAllocatedStorage *int `json:"maxAllocatedStorage,omitempty"`

	// MonitoringInterval is the interval, in seconds, between points when Enhanced Monitoring metrics
	// are collected for the DB instance. To disable collecting Enhanced Monitoring
	// metrics, specify 0. The default is 0.
	// If MonitoringRoleARN is specified, then you must also set MonitoringInterval
	// to a value other than 0.
	// Valid Values: 0, 1, 5, 10, 15, 30, 60
	// +optional
	MonitoringInterval *int `json:"monitoringInterval,omitempty"`

	// MonitoringRoleARN is the ARN for the IAM role that permits RDS to send enhanced monitoring metrics
	// to Amazon CloudWatch Logs. For example, arn:aws:iam:123456789012:role/emaccess.
	// For information on creating a monitoring role, go to Setting Up and Enabling
	// Enhanced Monitoring (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.html#USER_Monitoring.OS.Enabling)
	// in the Amazon RDS User Guide.
	// If MonitoringInterval is set to a value other than 0, then you must supply
	// a MonitoringRoleARN value.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	MonitoringRoleARN *string `json:"monitoringRoleArn,omitempty"`

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

	// MultiAZ specifies if the DB instance is a Multi-AZ deployment. You can't set the
	// AvailabilityZone parameter if the MultiAZ parameter is set to true.
	// +optional
	MultiAZ *bool `json:"multiAZ,omitempty"`

	// PerformanceInsightsKMSKeyID is the AWS KMS key identifier for encryption of Performance Insights data. The
	// KMS key ID is the Amazon Resource Name (ARN), KMS key identifier, or the
	// KMS key alias for the KMS encryption key.
	// +optional
	PerformanceInsightsKMSKeyID *string `json:"performanceInsightsKMSKeyId,omitempty"`

	// PerformanceInsightsRetentionPeriod is the amount of time, in days, to
	// retain Performance Insights data. Valid values
	// are 7 or 731 (2 years).
	// +optional
	PerformanceInsightsRetentionPeriod *int `json:"performanceInsightsRetentionPeriod,omitempty"`

	// Port number on which the database accepts connections.
	// MySQL
	// Default: 3306
	// Valid Values: 1150-65535
	// Type: Integer
	// MariaDB
	// Default: 3306
	// Valid Values: 1150-65535
	// Type: Integer
	// PostgreSQL
	// Default: 5432
	// Valid Values: 1150-65535
	// Type: Integer
	// Oracle
	// Default: 1521
	// Valid Values: 1150-65535
	// SQL Server
	// Default: 1433
	// Valid Values: 1150-65535 except for 1434, 3389, 47001, 49152, and 49152 through
	// 49156.
	// Amazon Aurora
	// Default: 3306
	// Valid Values: 1150-65535
	// Type: Integer
	// +optional
	Port *int `json:"port,omitempty"`

	// PreferredBackupWindow is the daily time range during which automated backups are created if automated
	// backups are enabled, using the BackupRetentionPeriod parameter. For more
	// information, see The Backup Window (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithAutomatedBackups.html#USER_WorkingWithAutomatedBackups.BackupWindow)
	// in the Amazon RDS User Guide.
	// Amazon Aurora
	// Not applicable. The daily time range for creating automated backups is managed
	// by the DB cluster. For more information, see CreateDBCluster.
	// The default is a 30-minute window selected at random from an 8-hour block
	// of time for each AWS Region. To see the time blocks available, see  Adjusting
	// the Preferred DB Instance Maintenance Window (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html#AdjustingTheMaintenanceWindow)
	// in the Amazon RDS User Guide.
	// Constraints:
	//    * Must be in the format hh24:mi-hh24:mi.
	//    * Must be in Universal Coordinated Time (UTC).
	//    * Must not conflict with the preferred maintenance window.
	//    * Must be at least 30 minutes.
	// +optional
	PreferredBackupWindow *string `json:"preferredBackupWindow,omitempty"`

	// PreferredMaintenanceWindow is the time range each week during which system maintenance can occur, in Universal
	// Coordinated Time (UTC). For more information, see Amazon RDS Maintenance
	// Window (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html#Concepts.DBMaintenance).
	// Format: ddd:hh24:mi-ddd:hh24:mi
	// The default is a 30-minute window selected at random from an 8-hour block
	// of time for each AWS Region, occurring on a random day of the week.
	// Valid Days: Mon, Tue, Wed, Thu, Fri, Sat, Sun.
	// Constraints: Minimum 30-minute window.
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	// ProcessorFeatures is the number of CPU cores and the number of threads per core for the DB instance
	// class of the DB instance.
	// +optional
	ProcessorFeatures []ProcessorFeature `json:"processorFeatures,omitempty"`

	// PromotionTier specifies the order in which an Aurora Replica is promoted to
	// the primary instance after a failure of the existing primary instance. For
	// more information, see  Fault Tolerance for an Aurora DB Cluster (http://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Managing.Backups.html#Aurora.Managing.FaultTolerance)
	// in the Amazon Aurora User Guide.
	// Default: 1
	// Valid Values: 0 - 15
	// +optional
	PromotionTier *int `json:"promotionTier,omitempty"`

	// PubliclyAccessible specifies the accessibility options for the DB instance. A value of true
	// specifies an Internet-facing instance with a publicly resolvable DNS name,
	// which resolves to a public IP address. A value of false specifies an internal
	// instance with a DNS name that resolves to a private IP address.
	// Default: The default behavior varies depending on whether DBSubnetGroupName
	// is specified.
	// If DBSubnetGroupName is not specified, and PubliclyAccessible is not specified,
	// the following applies:
	//    * If the default VPC in the target region doesn’t have an Internet gateway
	//    attached to it, the DB instance is private.
	//    * If the default VPC in the target region has an Internet gateway attached
	//    to it, the DB instance is public.
	// If DBSubnetGroupName is specified, and PubliclyAccessible is not specified,
	// the following applies:
	//    * If the subnets are part of a VPC that doesn’t have an Internet gateway
	//    attached to it, the DB instance is private.
	//    * If the subnets are part of a VPC that has an Internet gateway attached
	//    to it, the DB instance is public.
	// +optional
	PubliclyAccessible *bool `json:"publiclyAccessible,omitempty"`

	// ScalingConfiguration is the scaling properties of the DB cluster. You can only modify scaling properties
	// for DB clusters in serverless DB engine mode.
	// +immutable
	// +optional
	ScalingConfiguration *ScalingConfiguration `json:"scalingConfiguration,omitempty"`

	// StorageEncrypted specifies whether the DB instance is encrypted.
	// Amazon Aurora
	// Not applicable. The encryption for DB instances is managed by the DB cluster.
	// For more information, see CreateDBCluster.
	// Default: false
	// +immutable
	// +optional
	StorageEncrypted *bool `json:"storageEncrypted,omitempty"`

	// The storage throughput value for the DB instance.
	//
	// This setting applies only to the gp3 storage type.
	//
	// This setting doesn't apply to Amazon Aurora or RDS Custom DB instances.
	//
	// Note: controller considers 0 and null as equivalent
	// +optional
	StorageThroughput *int `json:"storageThroughput,omitempty"`

	// StorageType specifies the storage type to be associated with the DB instance.
	// Valid values: standard | gp2 | io1
	// If you specify io1, you must also include a value for the IOPS parameter.
	// Default: io1 if the IOPS parameter is specified, otherwise standard
	// +optional
	StorageType *string `json:"storageType,omitempty"`

	// Tags. For more information, see Tagging Amazon RDS Resources (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	// +immutable
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// TODO(muvaf): get this password as input when we have a way of supplying
	// sensitive information as input from the user.

	// TDECredentialArn is the ARN from the key store with which to associate the instance for TDE encryption.
	// +optional
	// TDECredentialArn *string `json:"tdeCredentialArn,omitempty"`
	// TDECredentialPassword is the password for the given ARN from the key store in order to access the
	// device.
	// +optional
	// TDECredentialPassword *string `json:"tdeCredentialPassword,omitempty"`

	// Timezone of the DB instance. The time zone parameter is currently supported
	// only by Microsoft SQL Server (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.TimeZone).
	// +immutable
	// +optional
	Timezone *string `json:"timezone,omitempty"`

	// VPCSecurityGroupIDs is a list of EC2 VPC security groups to associate with this DB instance.
	// Amazon Aurora
	// Not applicable. The associated list of EC2 VPC security groups is managed
	// by the DB cluster. For more information, see CreateDBCluster.
	// Default: The default EC2 VPC security group for the DB subnet group's VPC.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=VPCSecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=VPCSecurityGroupIDSelector
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

	// Fields whose value cannot be retrieved from rds.DBInstance object.

	// AllowMajorVersionUpgrade indicates that major version upgrades are allowed. Changing this parameter
	// doesn't result in an outage and the change is asynchronously applied as soon
	// as possible.
	// Constraints: This parameter must be set to true when specifying a value for
	// the EngineVersion parameter that is a different major version than the DB
	// instance's current version.
	// +optional
	AllowMajorVersionUpgrade *bool `json:"allowMajorVersionUpgrade,omitempty"`

	// ApplyModificationsImmediately specifies whether the modifications in this request and any pending modifications
	// are asynchronously applied as soon as possible, regardless of the PreferredMaintenanceWindow
	// setting for the DB instance.
	// If this parameter is set to false, changes to the DB instance are applied
	// during the next maintenance window. Some parameter changes can cause an outage
	// and are applied on the next call to RebootDBInstance, or the next failure
	// reboot. Review the table of parameters in Modifying a DB Instance and Using
	// the Apply Immediately Parameter (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
	// in the Amazon RDS User Guide. to see the impact that setting ApplyImmediately
	// to true or false has for each modified parameter and to determine when the
	// changes are applied.
	// Default: false
	// +optional
	ApplyModificationsImmediately *bool `json:"applyModificationsImmediately,omitempty"`

	// Deprecated: This field will be removed. Use `enableCloudwatchLogsExports` instead.
	// CloudwatchLogsExportConfiguration is the configuration setting for the log types to be enabled for export to CloudWatch
	// Logs for a specific DB instance.
	// +immutable
	// +optional
	CloudwatchLogsExportConfiguration *CloudwatchLogsExportConfiguration `json:"cloudwatchLogsExportConfiguration,omitempty"`
	// TODO: remove deprecated field + code. Mapping to EnableCloudwatchLogsExports while in deprecation.

	// DBParameterGroupName is the name of the DB parameter group to associate with this DB instance. If
	// this argument is omitted, the default DBParameterGroup for the specified
	// engine is used.
	// Constraints:
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//    * First character must be a letter
	//    * Cannot end with a hyphen or contain two consecutive hyphens
	// +optional
	DBParameterGroupName *string `json:"dbParameterGroupName,omitempty"`

	// Domain specifies the Active Directory Domain to create the instance in.
	// +optional
	Domain *string `json:"domain,omitempty"`

	// DomainIAMRoleName specifies the name of the IAM role to be used when making API calls to the
	// Directory Service.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	DomainIAMRoleName *string `json:"domainIAMRoleName,omitempty"`

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

	// OptionGroupName indicates that the DB instance should be associated with the specified option
	// group.
	// Permanent options, such as the TDE option for Oracle Advanced Security TDE,
	// can't be removed from an option group, and that option group can't be removed
	// from a DB instance once it is associated with a DB instance
	// +optional
	OptionGroupName *string `json:"optionGroupName,omitempty"`

	// A value that specifies that the DB instance class of the DB instance uses
	// its default processor features.
	UseDefaultProcessorFeatures *bool `json:"useDefaultProcessorFeatures,omitempty"`

	// Determines whether a final DB snapshot is created before the DB instance
	// is deleted. If true is specified, no DBSnapshot is created. If false is specified,
	// a DB snapshot is created before the DB instance is deleted.
	// Note that when a DB instance is in a failure state and has a status of 'failed',
	// 'incompatible-restore', or 'incompatible-network', it can only be deleted
	// when the SkipFinalSnapshotBeforeDeletion parameter is set to "true".
	// Specify true when deleting a Read Replica.
	// The FinalDBSnapshotIdentifier parameter must be specified if SkipFinalSnapshotBeforeDeletion
	// is false.
	// Default: false
	SkipFinalSnapshotBeforeDeletion *bool `json:"skipFinalSnapshotBeforeDeletion,omitempty"`

	// The DBSnapshotIdentifier of the new DBSnapshot created when SkipFinalSnapshot
	// is set to false.
	// Specifying this parameter and also setting the SkipFinalShapshot parameter
	// to true results in an error.
	// Constraints:
	//    * Must be 1 to 255 letters or numbers.
	//    * First character must be a letter
	//    * Cannot end with a hyphen or contain two consecutive hyphens
	//    * Cannot be specified when deleting a Read Replica.
	FinalDBSnapshotIdentifier *string `json:"finalDBSnapshotIdentifier,omitempty"`
}

// An RDSInstanceSpec defines the desired state of an RDSInstance.
type RDSInstanceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RDSInstanceParameters `json:"forProvider"`
}

// RDSInstanceState represents the state of an RDS instance.
type RDSInstanceState string

// RDS instance states.
const (
	// The instance is healthy and available
	RDSInstanceStateAvailable = "available"
	// The instance is being created. The instance is inaccessible while it is being created.
	RDSInstanceStateCreating = "creating"
	// The instance is being deleted.
	RDSInstanceStateDeleting = "deleting"
	// The instance is being modified.
	RDSInstanceStateModifying = "modifying"
	// The instance is being backed up, but is available
	RDSInstanceStateBackingUp = "backing-up"
	// The instance is being backed up, but is available
	RDSInstanceStateConfiguringEnhancedMonitoring = "configuring-enhanced-monitoring"
	// After you modify the storage size for a DB instance, the status of the DB instance is storage-optimization.
	RDSInstanceStateStorageOptimization = "storage-optimization"
	// The instance has failed and Amazon RDS can't recover it. Perform a point-in-time restore to the latest restorable time of the instance to recover the data.
	RDSInstanceStateFailed = "failed"
)

// DBParameterGroupStatus is the status of the DB parameter group.
// This data type is used as a response element in the following actions:
//   - CreateDBInstance
//   - CreateDBInstanceReadReplica
//   - DeleteDBInstance
//   - ModifyDBInstance
//   - RebootDBInstance
//   - RestoreDBInstanceFromDBSnapshot
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/DBParameterGroupStatus
type DBParameterGroupStatus struct {
	// DBParameterGroupName is the name of the DP parameter group.
	DBParameterGroupName string `json:"dbParameterGroupName,omitempty"`

	// ParameterApplyStatus is the status of parameter updates.
	ParameterApplyStatus string `json:"parameterApplyStatus,omitempty"`
}

// DBSecurityGroupMembership is used as a response element in the following actions:
//   - ModifyDBInstance
//   - RebootDBInstance
//   - RestoreDBInstanceFromDBSnapshot
//   - RestoreDBInstanceToPointInTime
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/DBSecurityGroupMembership
type DBSecurityGroupMembership struct {
	// DBSecurityGroupName is the name of the DB security group.
	DBSecurityGroupName string `json:"dbSecurityGroupName,omitempty"`

	// Status is the status of the DB security group.
	Status string `json:"status,omitempty"`
}

// AvailabilityZone contains Availability Zone information.
// This data type is used as an element in the following data type:
//   - OrderableDBInstanceOption
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/AvailabilityZone
type AvailabilityZone struct {
	// Name of the Availability Zone.
	Name string `json:"name,omitempty"`
}

// SubnetInRDS is used as a response element in the DescribeDBSubnetGroups
// action.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/Subnet
type SubnetInRDS struct {
	// SubnetAvailabilityZone contains Availability Zone information.
	// This data type is used as an element in the following data type:
	//    * OrderableDBInstanceOption
	SubnetAvailabilityZone AvailabilityZone `json:"subnetAvailabilityZone,omitempty"`

	// SubnetIdentifier specifies the identifier of the subnet.
	SubnetIdentifier string `json:"subnetIdentifier,omitempty"`

	// SubnetStatus specifies the status of the subnet.
	SubnetStatus string `json:"subnetStatus,omitempty"`
}

// DBSubnetGroupInRDS contains the details of an Amazon RDS DB subnet group.
// This data type is used as a response element in the DescribeDBSubnetGroups
// action.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/DBSubnetGroup
type DBSubnetGroupInRDS struct {
	// DBSubnetGroupARN is the Amazon Resource Name (ARN) for the DB subnet group.
	DBSubnetGroupARN string `json:"dbSubnetGroupArn,omitempty"`

	// DBSubnetGroupDescription provides the description of the DB subnet group.
	DBSubnetGroupDescription string `json:"dbSubnetGroupDescription,omitempty"`

	// DBSubnetGroupName is the name of the DB subnet group.
	DBSubnetGroupName string `json:"dbSubnetGroupName,omitempty"`

	// SubnetGroupStatus provides the status of the DB subnet group.
	SubnetGroupStatus string `json:"subnetGroupStatus,omitempty"`

	// Subnets contains a list of Subnet elements.
	Subnets []SubnetInRDS `json:"subnets,omitempty"`

	// VPCID provides the VPCID of the DB subnet group.
	VPCID string `json:"vpcId,omitempty"`
}

// DomainMembership is an Active Directory Domain membership record associated
// with the DB instance.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/DomainMembership
type DomainMembership struct {
	// Domain is the identifier of the Active Directory Domain.
	Domain string `json:"domain,omitempty"`

	// FQDN us the fully qualified domain name of the Active Directory Domain.
	FQDN string `json:"fqdn,omitempty"`

	// IAMRoleName is the name of the IAM role to be used when making API calls
	// to the Directory Service.
	IAMRoleName string `json:"iamRoleName,omitempty"`

	// Status of the DB instance's Active Directory Domain membership, such
	// as joined, pending-join, failed etc).
	Status string `json:"status,omitempty"`
}

// Endpoint is used as a response element in the following actions:
//   - CreateDBInstance
//   - DescribeDBInstances
//   - DeleteDBInstance
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/Endpoint
type Endpoint struct {
	// Address specifies the DNS address of the DB instance.
	Address string `json:"address,omitempty"`

	// HostedZoneID specifies the ID that Amazon Route 53 assigns when you create a hosted zone.
	HostedZoneID string `json:"hostedZoneId,omitempty"`

	// Port specifies the port that the database engine is listening on.
	Port int `json:"port,omitempty"`
}

// OptionGroupMembership provides information on the option groups the DB instance is a member of.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/OptionGroupMembership
type OptionGroupMembership struct {
	// OptionGroupName is the name of the option group that the instance belongs to.
	OptionGroupName string `json:"optionGroupName,omitempty"`

	// Status is the status of the DB instance's option group membership. Valid values are:
	// in-sync, pending-apply, pending-removal, pending-maintenance-apply, pending-maintenance-removal,
	// applying, removing, and failed.
	Status string `json:"status,omitempty"`
}

// PendingCloudwatchLogsExports is a list of the log types whose configuration
// is still pending. In other words, these log types are in the process of being
// activated or deactivated.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/PendingCloudwatchLogsExports
type PendingCloudwatchLogsExports struct {
	// LogTypesToDisable is list of log types that are in the process of being
	// enabled. After they are enabled, these log types are exported to
	// CloudWatch Logs.
	LogTypesToDisable []string `json:"logTypesToDisable,omitempty"`

	// LogTypesToEnable is the log types that are in the process of being
	// deactivated. After they are deactivated, these log types aren't exported
	// to CloudWatch Logs.
	LogTypesToEnable []string `json:"logTypesToEnable,omitempty"`
}

// PendingModifiedValues is used as a response element in the ModifyDBInstance action.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/PendingModifiedValues
type PendingModifiedValues struct {
	// AllocatedStorage contains the new AllocatedStorage size for the DB instance that will be applied
	// or is currently being applied.
	AllocatedStorage int `json:"allocatedStorage,omitempty"`

	// BackupRetentionPeriod specifies the pending number of days for which automated backups are retained.
	BackupRetentionPeriod int `json:"backupRetentionPeriod,omitempty"`

	// CACertificateIdentifier specifies the identifier of the CA certificate for the DB instance.
	CACertificateIdentifier string `json:"caCertificateIdentifier,omitempty"`

	// DBInstanceClass contains the new DBInstanceClass for the DB instance that will be applied
	// or is currently being applied.
	DBInstanceClass string `json:"dbInstanceClass,omitempty"`

	// DBSubnetGroupName is the new DB subnet group for the DB instance.
	DBSubnetGroupName string `json:"dbSubnetGroupName,omitempty"`

	// EngineVersion indicates the database engine version.
	EngineVersion string `json:"engineVersion,omitempty"`

	// IOPS specifies the new Provisioned IOPS value for the DB instance that will be
	// applied or is currently being applied.
	IOPS int `json:"iops,omitempty"`

	// LicenseModel is the license model for the DB instance.
	// Valid values: license-included | bring-your-own-license | general-public-license
	LicenseModel string `json:"licenseModel,omitempty"`

	// MultiAZ indicates that the Single-AZ DB instance is to change to a Multi-AZ deployment.
	MultiAZ bool `json:"multiAZ,omitempty"`

	// PendingCloudwatchLogsExports is a list of the log types whose configuration is still pending. In other words,
	// these log types are in the process of being activated or deactivated.
	PendingCloudwatchLogsExports PendingCloudwatchLogsExports `json:"pendingCloudwatchLogsExports,omitempty"`

	// Port specifies the pending port for the DB instance.
	Port int `json:"port,omitempty"`

	// ProcessorFeatures is the number of CPU cores and the number of threads per core for the DB instance
	// class of the DB instance.
	ProcessorFeatures []ProcessorFeature `json:"processorFeatures,omitempty"`

	// StorageThroughput indicates the new storage throughput value for the DB instance
	// that will be applied or is currently being applied.
	StorageThroughput int `json:"storageThroughput,omitempty"`

	// StorageType specifies the storage type to be associated with the DB instance.
	StorageType string `json:"storageType,omitempty"`
}

// DBInstanceStatusInfo provides a list of status information for a DB instance.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/DBInstanceStatusInfo
type DBInstanceStatusInfo struct {
	// Message is the details of the error if there is an error for the instance. If the instance
	// is not in an error state, this value is blank.
	Message string `json:"message,omitempty"`

	// Normal is true if the instance is operating normally, or false
	// if the instance is in an error state.
	Normal bool `json:"normal,omitempty"`

	// Status of the DB instance. For a StatusType of read replica, the values can
	// be replicating, replication stop point set, replication stop point reached,
	// error, stopped, or terminated.
	Status string `json:"status,omitempty"`

	// StatusType is currently "read replication."
	StatusType string `json:"statusType,omitempty"`
}

// VPCSecurityGroupMembership is used as a response element for queries on VPC security
// group membership.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/rds-2014-10-31/VpcSecurityGroupMembership
type VPCSecurityGroupMembership struct {
	// Status is the status of the VPC security group.
	Status string `json:"status,omitempty"`

	// VPCSecurityGroupID is the name of the VPC security group.
	VPCSecurityGroupID string `json:"vpcSecurityGroupId,omitempty"`
}

// RDSInstanceObservation is the representation of the current state that is observed.
type RDSInstanceObservation struct {
	// AWSBackupRecoveryPointARN is the Amazon Resource Name (ARN) of the recovery point in Amazon Web Services Backup.
	AWSBackupRecoveryPointARN string `json:"awsBackupRecoveryPointARN,omitempty"`

	// BackupRetentionPeriod is the number of days for which automated backups are retained.
	BackupRetentionPeriod int `json:"backupRetentionPeriod,omitempty"`

	// DBInstanceStatus specifies the current state of this database.
	DBInstanceStatus string `json:"dbInstanceStatus,omitempty"`

	// DBInstanceArn is the Amazon Resource Name (ARN) for the DB instance.
	DBInstanceArn string `json:"dbInstanceArn,omitempty"`

	// DBParameterGroups provides the list of DB parameter groups applied to this DB instance.
	DBParameterGroups []DBParameterGroupStatus `json:"dbParameterGroups,omitempty"`

	// DBSecurityGroups provides List of DB security group elements containing only DBSecurityGroup.Name
	// and DBSecurityGroup.Status subelements.
	DBSecurityGroups []DBSecurityGroupMembership `json:"dbSecurityGroups,omitempty"`

	// DBSubnetGroup specifies information on the subnet group associated with the DB instance,
	// including the name, description, and subnets in the subnet group.
	DBSubnetGroup DBSubnetGroupInRDS `json:"dbSubnetGroup,omitempty"`

	// DBInstancePort specifies the port that the DB instance listens on. If the DB instance is
	// part of a DB cluster, this can be a different port than the DB cluster port.
	DBInstancePort int `json:"dbInstancePort,omitempty"`

	// DBResourceID is the AWS Region-unique, immutable identifier for the DB instance. This identifier
	// is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB
	// instance is accessed.
	DBResourceID string `json:"dbResourceId,omitempty"`

	// AllocatedStorage is the allocated storage size in gibibytes.
	AllocatedStorage int `json:"allocatedStorage,omitempty"`

	// DomainMemberships is the Active Directory Domain membership records associated with the DB instance.
	DomainMemberships []DomainMembership `json:"domainMemberships,omitempty"`

	// InstanceCreateTime provides the date and time the DB instance was created.
	InstanceCreateTime *metav1.Time `json:"instanceCreateTime,omitempty"`

	// A list of log types that this DB instance is configured to export to CloudWatch
	// Logs. Log types vary by DB engine. For information about the log types for each
	// DB engine, see Amazon RDS Database Log Files
	// (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.html) in
	// the Amazon RDS User Guide.
	EnabledCloudwatchLogsExports []string `json:"enabledCloudwatchLogsExports,omitempty"`

	// Endpoint specifies the connection endpoint.
	Endpoint Endpoint `json:"endpoint,omitempty"`

	// Indicates the database engine version.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// EnhancedMonitoringResourceArn is the Amazon Resource Name (ARN) of the
	// Amazon CloudWatch Logs log stream that receives the Enhanced Monitoring
	// metrics data for the DB instance.
	EnhancedMonitoringResourceArn string `json:"enhancedMonitoringResourceArn,omitempty"`

	// LatestRestorableTime specifies the latest time to which a database can be
	// restored with point-in-time restore.
	LatestRestorableTime *metav1.Time `json:"latestRestorableTime,omitempty"`

	// OptionGroupMemberships provides the list of option group memberships for this DB instance.
	OptionGroupMemberships []OptionGroupMembership `json:"optionGroupMemberships,omitempty"`

	// PendingModifiedValues specifies that changes to the DB instance are pending. This element is only
	// included when changes are pending. Specific changes are identified by subelements.
	PendingModifiedValues PendingModifiedValues `json:"pendingModifiedValues,omitempty"`

	// PerformanceInsightsEnabled is true if Performance Insights is enabled for
	// the DB instance, and otherwise false.
	PerformanceInsightsEnabled bool `json:"performanceInsightsEnabled,omitempty"`

	// ReadReplicaDBClusterIdentifiers contains one or more identifiers of Aurora DB clusters to which the RDS DB
	// instance is replicated as a Read Replica. For example, when you create an
	// Aurora Read Replica of an RDS MySQL DB instance, the Aurora MySQL DB cluster
	// for the Aurora Read Replica is shown. This output does not contain information
	// about cross region Aurora Read Replicas.
	ReadReplicaDBClusterIdentifiers []string `json:"readReplicaDBClusterIdentifiers,omitempty"`

	// ReadReplicaDBInstanceIdentifiers contains one or more identifiers of the Read Replicas associated with this
	// DB instance.
	ReadReplicaDBInstanceIdentifiers []string `json:"readReplicaDBInstanceIdentifiers,omitempty"`

	// ReadReplicaSourceDBInstanceIdentifier contains the identifier of the source DB instance if this DB instance is
	// a Read Replica.
	ReadReplicaSourceDBInstanceIdentifier string `json:"readReplicaSourceDBInstanceIdentifier,omitempty"`

	// SecondaryAvailabilityZone specifies the name of the secondary Availability Zone for a DB
	// instance with multi-AZ support when it is present.
	SecondaryAvailabilityZone string `json:"secondaryAvailabilityZone,omitempty"`

	// StatusInfos is the status of a Read Replica. If the instance is not a Read Replica, this
	// is blank.
	StatusInfos []DBInstanceStatusInfo `json:"statusInfos,omitempty"`

	// VPCSecurityGroups provides a list of VPC security group elements that the DB instance belongs
	// to.
	VPCSecurityGroups []VPCSecurityGroupMembership `json:"vpcSecurityGroups,omitempty"`
}

// An RDSInstanceStatus represents the observed state of an RDSInstance.
type RDSInstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RDSInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An RDSInstance is a managed resource that represents an AWS Relational
// Database Service instance.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.dbInstanceStatus"
// +kubebuilder:printcolumn:name="ENGINE",type="string",JSONPath=".spec.forProvider.engine"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status.atProvider.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type RDSInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDSInstanceSpec   `json:"spec"`
	Status RDSInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RDSInstanceList contains a list of RDSInstance
type RDSInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDSInstance `json:"items"`
}
