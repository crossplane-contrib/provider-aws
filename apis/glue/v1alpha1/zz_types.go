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

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
)

// +kubebuilder:skipversion
type Action struct {
	Arguments map[string]*string `json:"arguments,omitempty"`

	CrawlerName *string `json:"crawlerName,omitempty"`

	JobName *string `json:"jobName,omitempty"`
	// Specifies configuration properties of a notification.
	NotificationProperty *NotificationProperty `json:"notificationProperty,omitempty"`

	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	Timeout *int64 `json:"timeout,omitempty"`
}

// +kubebuilder:skipversion
type BatchStopJobRunError struct {
	JobName *string `json:"jobName,omitempty"`
}

// +kubebuilder:skipversion
type BatchStopJobRunSuccessfulSubmission struct {
	JobName *string `json:"jobName,omitempty"`
}

// +kubebuilder:skipversion
type CatalogEntry struct {
	DatabaseName *string `json:"databaseName,omitempty"`

	TableName *string `json:"tableName,omitempty"`
}

// +kubebuilder:skipversion
type CatalogImportStatus struct {
	ImportCompleted *bool `json:"importCompleted,omitempty"`

	ImportTime *metav1.Time `json:"importTime,omitempty"`

	ImportedBy *string `json:"importedBy,omitempty"`
}

// +kubebuilder:skipversion
type CatalogTarget struct {
	DatabaseName *string `json:"databaseName,omitempty"`

	Tables []*string `json:"tables,omitempty"`
}

// +kubebuilder:skipversion
type Classifier_SDK struct {
	// A classifier for custom CSV content.
	CsvClassifier *CsvClassifier `json:"csvClassifier,omitempty"`
	// A classifier that uses grok patterns.
	GrokClassifier *GrokClassifier `json:"grokClassifier,omitempty"`
	// A classifier for JSON content.
	JSONClassifier *JSONClassifier `json:"jsonClassifier,omitempty"`
	// A classifier for XML content.
	XMLClassifier *XMLClassifier `json:"xmlClassifier,omitempty"`
}

// +kubebuilder:skipversion
type CloudWatchEncryption struct {
	CloudWatchEncryptionMode *string `json:"cloudWatchEncryptionMode,omitempty"`

	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`
}

// +kubebuilder:skipversion
type CodeGenNodeArg struct {
	Param *bool `json:"param,omitempty"`
}

// +kubebuilder:skipversion
type Column struct {
	Name *string `json:"name,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`
}

// +kubebuilder:skipversion
type ColumnError struct {
	ColumnName *string `json:"columnName,omitempty"`
}

// +kubebuilder:skipversion
type ColumnImportance struct {
	ColumnName *string `json:"columnName,omitempty"`
}

// +kubebuilder:skipversion
type ColumnStatistics struct {
	AnalyzedTime *metav1.Time `json:"analyzedTime,omitempty"`

	ColumnName *string `json:"columnName,omitempty"`
}

// +kubebuilder:skipversion
type Condition struct {
	CrawlerName *string `json:"crawlerName,omitempty"`

	JobName *string `json:"jobName,omitempty"`
}

// +kubebuilder:skipversion
type ConnectionInput struct {
	ConnectionProperties map[string]*string `json:"connectionProperties,omitempty"`

	ConnectionType *string `json:"connectionType,omitempty"`

	Description *string `json:"description,omitempty"`

	MatchCriteria []*string `json:"matchCriteria,omitempty"`

	Name *string `json:"name,omitempty"`
	// Specifies the physical requirements for a connection.
	PhysicalConnectionRequirements *PhysicalConnectionRequirements `json:"physicalConnectionRequirements,omitempty"`
}

// +kubebuilder:skipversion
type ConnectionPasswordEncryption struct {
	AWSKMSKeyID *string `json:"awsKMSKeyID,omitempty"`

	ReturnConnectionPasswordEncrypted *bool `json:"returnConnectionPasswordEncrypted,omitempty"`
}

// +kubebuilder:skipversion
type Connection_SDK struct {
	ConnectionProperties map[string]*string `json:"connectionProperties,omitempty"`

	ConnectionType *string `json:"connectionType,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	Description *string `json:"description,omitempty"`

	LastUpdatedBy *string `json:"lastUpdatedBy,omitempty"`

	LastUpdatedTime *metav1.Time `json:"lastUpdatedTime,omitempty"`

	MatchCriteria []*string `json:"matchCriteria,omitempty"`

	Name *string `json:"name,omitempty"`
	// Specifies the physical requirements for a connection.
	PhysicalConnectionRequirements *PhysicalConnectionRequirements `json:"physicalConnectionRequirements,omitempty"`
}

// +kubebuilder:skipversion
type ConnectionsList struct {
	Connections []*string `json:"connections,omitempty"`
}

// +kubebuilder:skipversion
type Crawl struct {
	CompletedOn *metav1.Time `json:"completedOn,omitempty"`

	ErrorMessage *string `json:"errorMessage,omitempty"`

	LogGroup *string `json:"logGroup,omitempty"`

	LogStream *string `json:"logStream,omitempty"`

	StartedOn *metav1.Time `json:"startedOn,omitempty"`
}

// +kubebuilder:skipversion
type CrawlerMetrics struct {
	CrawlerName *string `json:"crawlerName,omitempty"`

	StillEstimating *bool `json:"stillEstimating,omitempty"`
}

// +kubebuilder:skipversion
type CrawlerTargets struct {
	CatalogTargets []*CatalogTarget `json:"catalogTargets,omitempty"`

	DynamoDBTargets []*DynamoDBTarget `json:"dynamoDBTargets,omitempty"`

	JdbcTargets []*JdbcTarget `json:"jdbcTargets,omitempty"`

	MongoDBTargets []*MongoDBTarget `json:"mongoDBTargets,omitempty"`

	S3Targets []*S3Target `json:"s3Targets,omitempty"`
}

// +kubebuilder:skipversion
type Crawler_SDK struct {
	Classifiers []*string `json:"classifiers,omitempty"`

	Configuration *string `json:"configuration,omitempty"`

	CrawlElapsedTime *int64 `json:"crawlElapsedTime,omitempty"`

	CrawlerSecurityConfiguration *string `json:"crawlerSecurityConfiguration,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	Description *string `json:"description,omitempty"`
	// Status and error information about the most recent crawl.
	LastCrawl *LastCrawlInfo `json:"lastCrawl,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
	// Specifies data lineage configuration settings for the crawler.
	LineageConfiguration *LineageConfiguration `json:"lineageConfiguration,omitempty"`

	Name *string `json:"name,omitempty"`
	// When crawling an Amazon S3 data source after the first crawl is complete,
	// specifies whether to crawl the entire dataset again or to crawl only folders
	// that were added since the last crawler run. For more information, see Incremental
	// Crawls in AWS Glue (https://docs.aws.amazon.com/glue/latest/dg/incremental-crawls.html)
	// in the developer guide.
	RecrawlPolicy *RecrawlPolicy `json:"recrawlPolicy,omitempty"`

	Role *string `json:"role,omitempty"`
	// A policy that specifies update and deletion behaviors for the crawler.
	SchemaChangePolicy *SchemaChangePolicy `json:"schemaChangePolicy,omitempty"`

	State *string `json:"state,omitempty"`

	TablePrefix *string `json:"tablePrefix,omitempty"`
	// Specifies data stores to crawl.
	Targets *CrawlerTargets `json:"targets,omitempty"`

	Version *int64 `json:"version,omitempty"`
}

// +kubebuilder:skipversion
type CreateCsvClassifierRequest struct {
	AllowSingleColumn *bool `json:"allowSingleColumn,omitempty"`

	ContainsHeader *string `json:"containsHeader,omitempty"`

	Delimiter *string `json:"delimiter,omitempty"`

	DisableValueTrimming *bool `json:"disableValueTrimming,omitempty"`

	Header []*string `json:"header,omitempty"`

	Name *string `json:"name,omitempty"`

	QuoteSymbol *string `json:"quoteSymbol,omitempty"`
}

// +kubebuilder:skipversion
type CreateGrokClassifierRequest struct {
	Classification *string `json:"classification,omitempty"`

	CustomPatterns *string `json:"customPatterns,omitempty"`

	GrokPattern *string `json:"grokPattern,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type CreateJSONClassifierRequest struct {
	JSONPath *string `json:"jsonPath,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type CreateXMLClassifierRequest struct {
	Classification *string `json:"classification,omitempty"`

	Name *string `json:"name,omitempty"`

	RowTag *string `json:"rowTag,omitempty"`
}

// +kubebuilder:skipversion
type CsvClassifier struct {
	AllowSingleColumn *bool `json:"allowSingleColumn,omitempty"`

	ContainsHeader *string `json:"containsHeader,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	Delimiter *string `json:"delimiter,omitempty"`

	DisableValueTrimming *bool `json:"disableValueTrimming,omitempty"`

	Header []*string `json:"header,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	Name *string `json:"name,omitempty"`

	QuoteSymbol *string `json:"quoteSymbol,omitempty"`

	Version *int64 `json:"version,omitempty"`
}

// +kubebuilder:skipversion
type DataLakePrincipal struct {
	DataLakePrincipalIdentifier *string `json:"dataLakePrincipalIdentifier,omitempty"`
}

// +kubebuilder:skipversion
type DatabaseIdentifier struct {
	CatalogID *string `json:"catalogID,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`
}

// +kubebuilder:skipversion
type DatabaseInput struct {
	CreateTableDefaultPermissions []*PrincipalPermissions `json:"createTableDefaultPermissions,omitempty"`

	Description *string `json:"description,omitempty"`

	LocationURI *string `json:"locationURI,omitempty"`

	Name *string `json:"name,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`
	// A structure that describes a target database for resource linking.
	TargetDatabase *DatabaseIdentifier `json:"targetDatabase,omitempty"`
}

// +kubebuilder:skipversion
type Database_SDK struct {
	CatalogID *string `json:"catalogID,omitempty"`

	CreateTableDefaultPermissions []*PrincipalPermissions `json:"createTableDefaultPermissions,omitempty"`

	CreateTime *metav1.Time `json:"createTime,omitempty"`

	Description *string `json:"description,omitempty"`

	LocationURI *string `json:"locationURI,omitempty"`

	Name *string `json:"name,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`
	// A structure that describes a target database for resource linking.
	TargetDatabase *DatabaseIdentifier `json:"targetDatabase,omitempty"`
}

// +kubebuilder:skipversion
type DateColumnStatisticsData struct {
	MaximumValue *metav1.Time `json:"maximumValue,omitempty"`

	MinimumValue *metav1.Time `json:"minimumValue,omitempty"`
}

// +kubebuilder:skipversion
type DevEndpoint struct {
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	CreatedTimestamp *metav1.Time `json:"createdTimestamp,omitempty"`

	EndpointName *string `json:"endpointName,omitempty"`

	ExtraJarsS3Path *string `json:"extraJarsS3Path,omitempty"`

	ExtraPythonLibsS3Path *string `json:"extraPythonLibsS3Path,omitempty"`

	FailureReason *string `json:"failureReason,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	LastModifiedTimestamp *metav1.Time `json:"lastModifiedTimestamp,omitempty"`

	LastUpdateStatus *string `json:"lastUpdateStatus,omitempty"`

	NumberOfNodes *int64 `json:"numberOfNodes,omitempty"`

	NumberOfWorkers *int64 `json:"numberOfWorkers,omitempty"`

	PrivateAddress *string `json:"privateAddress,omitempty"`

	PublicAddress *string `json:"publicAddress,omitempty"`

	PublicKey *string `json:"publicKey,omitempty"`

	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	Status *string `json:"status,omitempty"`

	SubnetID *string `json:"subnetID,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`

	WorkerType *string `json:"workerType,omitempty"`

	YarnEndpointAddress *string `json:"yarnEndpointAddress,omitempty"`

	ZeppelinRemoteSparkInterpreterPort *int64 `json:"zeppelinRemoteSparkInterpreterPort,omitempty"`
}

// +kubebuilder:skipversion
type DevEndpointCustomLibraries struct {
	ExtraJarsS3Path *string `json:"extraJarsS3Path,omitempty"`

	ExtraPythonLibsS3Path *string `json:"extraPythonLibsS3Path,omitempty"`
}

// +kubebuilder:skipversion
type DynamoDBTarget struct {
	Path *string `json:"path,omitempty"`

	ScanAll *bool `json:"scanAll,omitempty"`

	ScanRate *float64 `json:"scanRate,omitempty"`
}

// +kubebuilder:skipversion
type Edge struct {
	DestinationID *string `json:"destinationID,omitempty"`

	SourceID *string `json:"sourceID,omitempty"`
}

// +kubebuilder:skipversion
type EncryptionAtRest struct {
	SSEAWSKMSKeyID *string `json:"sseAWSKMSKeyID,omitempty"`
}

// +kubebuilder:skipversion
type EncryptionConfiguration struct {
	// Specifies how Amazon CloudWatch data should be encrypted.
	CloudWatchEncryption *CloudWatchEncryption `json:"cloudWatchEncryption,omitempty"`
	// Specifies how job bookmark data should be encrypted.
	JobBookmarksEncryption *JobBookmarksEncryption `json:"jobBookmarksEncryption,omitempty"`

	S3Encryption []*S3Encryption `json:"s3Encryption,omitempty"`
}

// +kubebuilder:skipversion
type ErrorDetail struct {
	ErrorCode *string `json:"errorCode,omitempty"`

	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +kubebuilder:skipversion
type ExecutionProperty struct {
	MaxConcurrentRuns *int64 `json:"maxConcurrentRuns,omitempty"`
}

// +kubebuilder:skipversion
type ExportLabelsTaskRunProperties struct {
	OutputS3Path *string `json:"outputS3Path,omitempty"`
}

// +kubebuilder:skipversion
type FindMatchesParameters struct {
	EnforceProvidedLabels *bool `json:"enforceProvidedLabels,omitempty"`
}

// +kubebuilder:skipversion
type FindMatchesTaskRunProperties struct {
	JobName *string `json:"jobName,omitempty"`
}

// +kubebuilder:skipversion
type GetConnectionsFilter struct {
	ConnectionType *string `json:"connectionType,omitempty"`

	MatchCriteria []*string `json:"matchCriteria,omitempty"`
}

// +kubebuilder:skipversion
type GluePolicy struct {
	CreateTime *metav1.Time `json:"createTime,omitempty"`

	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
}

// +kubebuilder:skipversion
type GrokClassifier struct {
	Classification *string `json:"classification,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	CustomPatterns *string `json:"customPatterns,omitempty"`

	GrokPattern *string `json:"grokPattern,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	Name *string `json:"name,omitempty"`

	Version *int64 `json:"version,omitempty"`
}

// +kubebuilder:skipversion
type ImportLabelsTaskRunProperties struct {
	InputS3Path *string `json:"inputS3Path,omitempty"`
}

// +kubebuilder:skipversion
type JSONClassifier struct {
	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	JSONPath *string `json:"jsonPath,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	Name *string `json:"name,omitempty"`

	Version *int64 `json:"version,omitempty"`
}

// +kubebuilder:skipversion
type JdbcTarget struct {
	ConnectionName *string `json:"connectionName,omitempty"`

	Exclusions []*string `json:"exclusions,omitempty"`

	Path *string `json:"path,omitempty"`
}

// +kubebuilder:skipversion
type JobBookmarkEntry struct {
	Attempt *int64 `json:"attempt,omitempty"`

	Run *int64 `json:"run,omitempty"`

	Version *int64 `json:"version,omitempty"`
}

// +kubebuilder:skipversion
type JobBookmarksEncryption struct {
	JobBookmarksEncryptionMode *string `json:"jobBookmarksEncryptionMode,omitempty"`

	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`
}

// +kubebuilder:skipversion
type JobCommand struct {
	Name *string `json:"name,omitempty"`

	PythonVersion *string `json:"pythonVersion,omitempty"`

	ScriptLocation *string `json:"scriptLocation,omitempty"`
}

// +kubebuilder:skipversion
type JobRun struct {
	AllocatedCapacity *int64 `json:"allocatedCapacity,omitempty"`

	Arguments map[string]*string `json:"arguments,omitempty"`

	CompletedOn *metav1.Time `json:"completedOn,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	JobName *string `json:"jobName,omitempty"`

	LastModifiedOn *metav1.Time `json:"lastModifiedOn,omitempty"`

	LogGroupName *string `json:"logGroupName,omitempty"`

	MaxCapacity *float64 `json:"maxCapacity,omitempty"`
	// Specifies configuration properties of a notification.
	NotificationProperty *NotificationProperty `json:"notificationProperty,omitempty"`

	NumberOfWorkers *int64 `json:"numberOfWorkers,omitempty"`

	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	StartedOn *metav1.Time `json:"startedOn,omitempty"`

	Timeout *int64 `json:"timeout,omitempty"`

	TriggerName *string `json:"triggerName,omitempty"`

	WorkerType *string `json:"workerType,omitempty"`
}

// +kubebuilder:skipversion
type JobUpdate struct {
	AllocatedCapacity *int64 `json:"allocatedCapacity,omitempty"`
	// Specifies code executed when a job is run.
	Command *JobCommand `json:"command,omitempty"`
	// Specifies the connections used by a job.
	Connections *ConnectionsList `json:"connections,omitempty"`

	DefaultArguments map[string]*string `json:"defaultArguments,omitempty"`

	Description *string `json:"description,omitempty"`
	// An execution property of a job.
	ExecutionProperty *ExecutionProperty `json:"executionProperty,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	LogURI *string `json:"logURI,omitempty"`

	MaxCapacity *float64 `json:"maxCapacity,omitempty"`

	MaxRetries *int64 `json:"maxRetries,omitempty"`

	NonOverridableArguments map[string]*string `json:"nonOverridableArguments,omitempty"`
	// Specifies configuration properties of a notification.
	NotificationProperty *NotificationProperty `json:"notificationProperty,omitempty"`

	NumberOfWorkers *int64 `json:"numberOfWorkers,omitempty"`

	Role *string `json:"role,omitempty"`

	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	Timeout *int64 `json:"timeout,omitempty"`

	WorkerType *string `json:"workerType,omitempty"`
}

// +kubebuilder:skipversion
type Job_SDK struct {
	AllocatedCapacity *int64 `json:"allocatedCapacity,omitempty"`
	// Specifies code executed when a job is run.
	Command *JobCommand `json:"command,omitempty"`
	// Specifies the connections used by a job.
	Connections *ConnectionsList `json:"connections,omitempty"`

	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	DefaultArguments map[string]*string `json:"defaultArguments,omitempty"`

	Description *string `json:"description,omitempty"`
	// An execution property of a job.
	ExecutionProperty *ExecutionProperty `json:"executionProperty,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	LastModifiedOn *metav1.Time `json:"lastModifiedOn,omitempty"`

	LogURI *string `json:"logURI,omitempty"`

	MaxCapacity *float64 `json:"maxCapacity,omitempty"`

	MaxRetries *int64 `json:"maxRetries,omitempty"`

	Name *string `json:"name,omitempty"`

	NonOverridableArguments map[string]*string `json:"nonOverridableArguments,omitempty"`
	// Specifies configuration properties of a notification.
	NotificationProperty *NotificationProperty `json:"notificationProperty,omitempty"`

	NumberOfWorkers *int64 `json:"numberOfWorkers,omitempty"`

	Role *string `json:"role,omitempty"`

	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	Timeout *int64 `json:"timeout,omitempty"`

	WorkerType *string `json:"workerType,omitempty"`
}

// +kubebuilder:skipversion
type KeySchemaElement struct {
	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type LabelingSetGenerationTaskRunProperties struct {
	OutputS3Path *string `json:"outputS3Path,omitempty"`
}

// +kubebuilder:skipversion
type LastCrawlInfo struct {
	ErrorMessage *string `json:"errorMessage,omitempty"`

	LogGroup *string `json:"logGroup,omitempty"`

	LogStream *string `json:"logStream,omitempty"`

	MessagePrefix *string `json:"messagePrefix,omitempty"`

	StartTime *metav1.Time `json:"startTime,omitempty"`

	Status *string `json:"status,omitempty"`
}

// +kubebuilder:skipversion
type LineageConfiguration struct {
	CrawlerLineageSettings *string `json:"crawlerLineageSettings,omitempty"`
}

// +kubebuilder:skipversion
type MLTransform struct {
	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	Description *string `json:"description,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	LastModifiedOn *metav1.Time `json:"lastModifiedOn,omitempty"`

	MaxCapacity *float64 `json:"maxCapacity,omitempty"`

	MaxRetries *int64 `json:"maxRetries,omitempty"`

	Name *string `json:"name,omitempty"`

	NumberOfWorkers *int64 `json:"numberOfWorkers,omitempty"`

	Role *string `json:"role,omitempty"`

	Timeout *int64 `json:"timeout,omitempty"`

	WorkerType *string `json:"workerType,omitempty"`
}

// +kubebuilder:skipversion
type MLUserDataEncryption struct {
	KMSKeyID *string `json:"kmsKeyID,omitempty"`
}

// +kubebuilder:skipversion
type MongoDBTarget struct {
	ConnectionName *string `json:"connectionName,omitempty"`

	Path *string `json:"path,omitempty"`

	ScanAll *bool `json:"scanAll,omitempty"`
}

// +kubebuilder:skipversion
type Node struct {
	Name *string `json:"name,omitempty"`

	UniqueID *string `json:"uniqueID,omitempty"`
}

// +kubebuilder:skipversion
type NotificationProperty struct {
	NotifyDelayAfter *int64 `json:"notifyDelayAfter,omitempty"`
}

// +kubebuilder:skipversion
type Order struct {
	Column *string `json:"column,omitempty"`
}

// +kubebuilder:skipversion
type Partition struct {
	CatalogID *string `json:"catalogID,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	LastAccessTime *metav1.Time `json:"lastAccessTime,omitempty"`

	LastAnalyzedTime *metav1.Time `json:"lastAnalyzedTime,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`

	TableName *string `json:"tableName,omitempty"`
}

// +kubebuilder:skipversion
type PartitionIndex struct {
	IndexName *string `json:"indexName,omitempty"`
}

// +kubebuilder:skipversion
type PartitionIndexDescriptor struct {
	IndexName *string `json:"indexName,omitempty"`
}

// +kubebuilder:skipversion
type PartitionInput struct {
	LastAccessTime *metav1.Time `json:"lastAccessTime,omitempty"`

	LastAnalyzedTime *metav1.Time `json:"lastAnalyzedTime,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`
}

// +kubebuilder:skipversion
type PhysicalConnectionRequirements struct {
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	SecurityGroupIDList []*string `json:"securityGroupIDList,omitempty"`

	SubnetID *string `json:"subnetID,omitempty"`
}

// +kubebuilder:skipversion
type Predecessor struct {
	JobName *string `json:"jobName,omitempty"`
}

// +kubebuilder:skipversion
type PrincipalPermissions struct {
	Permissions []*string `json:"permissions,omitempty"`
	// The AWS Lake Formation principal.
	Principal *DataLakePrincipal `json:"principal,omitempty"`
}

// +kubebuilder:skipversion
type PropertyPredicate struct {
	Key *string `json:"key,omitempty"`

	Value *string `json:"value,omitempty"`
}

// +kubebuilder:skipversion
type RecrawlPolicy struct {
	RecrawlBehavior *string `json:"recrawlBehavior,omitempty"`
}

// +kubebuilder:skipversion
type RegistryListItem struct {
	Description *string `json:"description,omitempty"`
}

// +kubebuilder:skipversion
type ResourceURI struct {
	URI *string `json:"uri,omitempty"`
}

// +kubebuilder:skipversion
type S3Encryption struct {
	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`

	S3EncryptionMode *string `json:"s3EncryptionMode,omitempty"`
}

// +kubebuilder:skipversion
type S3Target struct {
	ConnectionName *string `json:"connectionName,omitempty"`

	Exclusions []*string `json:"exclusions,omitempty"`

	Path *string `json:"path,omitempty"`
}

// +kubebuilder:skipversion
type Schedule struct {
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// +kubebuilder:skipversion
type SchemaChangePolicy struct {
	DeleteBehavior *string `json:"deleteBehavior,omitempty"`

	UpdateBehavior *string `json:"updateBehavior,omitempty"`
}

// +kubebuilder:skipversion
type SchemaListItem struct {
	Description *string `json:"description,omitempty"`
}

// +kubebuilder:skipversion
type SecurityConfiguration_SDK struct {
	CreatedTimeStamp *metav1.Time `json:"createdTimeStamp,omitempty"`
	// Specifies an encryption configuration.
	EncryptionConfiguration *EncryptionConfiguration `json:"encryptionConfiguration,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type SerDeInfo struct {
	Name *string `json:"name,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`

	SerializationLibrary *string `json:"serializationLibrary,omitempty"`
}

// +kubebuilder:skipversion
type SortCriterion struct {
	FieldName *string `json:"fieldName,omitempty"`
}

// +kubebuilder:skipversion
type StorageDescriptor struct {
	Compressed *bool `json:"compressed,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`

	StoredAsSubDirectories *bool `json:"storedAsSubDirectories,omitempty"`
}

// +kubebuilder:skipversion
type Table struct {
	CatalogID *string `json:"catalogID,omitempty"`

	ConnectionName *string `json:"connectionName,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	TableName *string `json:"tableName,omitempty"`
}

// +kubebuilder:skipversion
type TableData struct {
	CatalogID *string `json:"catalogID,omitempty"`

	CreateTime *metav1.Time `json:"createTime,omitempty"`

	CreatedBy *string `json:"createdBy,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	Description *string `json:"description,omitempty"`

	IsRegisteredWithLakeFormation *bool `json:"isRegisteredWithLakeFormation,omitempty"`

	LastAccessTime *metav1.Time `json:"lastAccessTime,omitempty"`

	LastAnalyzedTime *metav1.Time `json:"lastAnalyzedTime,omitempty"`

	Name *string `json:"name,omitempty"`

	Owner *string `json:"owner,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`

	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
}

// +kubebuilder:skipversion
type TableError struct {
	TableName *string `json:"tableName,omitempty"`
}

// +kubebuilder:skipversion
type TableIdentifier struct {
	CatalogID *string `json:"catalogID,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type TableInput struct {
	Description *string `json:"description,omitempty"`

	LastAccessTime *metav1.Time `json:"lastAccessTime,omitempty"`

	LastAnalyzedTime *metav1.Time `json:"lastAnalyzedTime,omitempty"`

	Name *string `json:"name,omitempty"`

	Owner *string `json:"owner,omitempty"`

	Parameters map[string]*string `json:"parameters,omitempty"`
}

// +kubebuilder:skipversion
type TableVersionError struct {
	TableName *string `json:"tableName,omitempty"`
}

// +kubebuilder:skipversion
type TaskRun struct {
	CompletedOn *metav1.Time `json:"completedOn,omitempty"`

	ErrorString *string `json:"errorString,omitempty"`

	LastModifiedOn *metav1.Time `json:"lastModifiedOn,omitempty"`

	LogGroupName *string `json:"logGroupName,omitempty"`

	StartedOn *metav1.Time `json:"startedOn,omitempty"`
}

// +kubebuilder:skipversion
type TaskRunFilterCriteria struct {
	StartedAfter *metav1.Time `json:"startedAfter,omitempty"`

	StartedBefore *metav1.Time `json:"startedBefore,omitempty"`
}

// +kubebuilder:skipversion
type TransformEncryption struct {
	TaskRunSecurityConfigurationName *string `json:"taskRunSecurityConfigurationName,omitempty"`
}

// +kubebuilder:skipversion
type TransformFilterCriteria struct {
	CreatedAfter *metav1.Time `json:"createdAfter,omitempty"`

	CreatedBefore *metav1.Time `json:"createdBefore,omitempty"`

	GlueVersion *string `json:"glueVersion,omitempty"`

	LastModifiedAfter *metav1.Time `json:"lastModifiedAfter,omitempty"`

	LastModifiedBefore *metav1.Time `json:"lastModifiedBefore,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type Trigger struct {
	Description *string `json:"description,omitempty"`

	Name *string `json:"name,omitempty"`

	Schedule *string `json:"schedule,omitempty"`

	WorkflowName *string `json:"workflowName,omitempty"`
}

// +kubebuilder:skipversion
type TriggerUpdate struct {
	Description *string `json:"description,omitempty"`

	Name *string `json:"name,omitempty"`

	Schedule *string `json:"schedule,omitempty"`
}

// +kubebuilder:skipversion
type UpdateCsvClassifierRequest struct {
	AllowSingleColumn *bool `json:"allowSingleColumn,omitempty"`

	ContainsHeader *string `json:"containsHeader,omitempty"`

	Delimiter *string `json:"delimiter,omitempty"`

	DisableValueTrimming *bool `json:"disableValueTrimming,omitempty"`

	Header []*string `json:"header,omitempty"`

	Name *string `json:"name,omitempty"`

	QuoteSymbol *string `json:"quoteSymbol,omitempty"`
}

// +kubebuilder:skipversion
type UpdateGrokClassifierRequest struct {
	Classification *string `json:"classification,omitempty"`

	CustomPatterns *string `json:"customPatterns,omitempty"`

	GrokPattern *string `json:"grokPattern,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type UpdateJSONClassifierRequest struct {
	JSONPath *string `json:"jsonPath,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type UpdateXMLClassifierRequest struct {
	Classification *string `json:"classification,omitempty"`

	Name *string `json:"name,omitempty"`

	RowTag *string `json:"rowTag,omitempty"`
}

// +kubebuilder:skipversion
type UserDefinedFunction struct {
	CatalogID *string `json:"catalogID,omitempty"`

	ClassName *string `json:"className,omitempty"`

	CreateTime *metav1.Time `json:"createTime,omitempty"`

	DatabaseName *string `json:"databaseName,omitempty"`

	FunctionName *string `json:"functionName,omitempty"`

	OwnerName *string `json:"ownerName,omitempty"`
}

// +kubebuilder:skipversion
type UserDefinedFunctionInput struct {
	ClassName *string `json:"className,omitempty"`

	FunctionName *string `json:"functionName,omitempty"`

	OwnerName *string `json:"ownerName,omitempty"`
}

// +kubebuilder:skipversion
type Workflow struct {
	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	Description *string `json:"description,omitempty"`

	LastModifiedOn *metav1.Time `json:"lastModifiedOn,omitempty"`

	MaxConcurrentRuns *int64 `json:"maxConcurrentRuns,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type WorkflowRun struct {
	CompletedOn *metav1.Time `json:"completedOn,omitempty"`

	Name *string `json:"name,omitempty"`

	StartedOn *metav1.Time `json:"startedOn,omitempty"`
}

// +kubebuilder:skipversion
type WorkflowRunStatistics struct {
	FailedActions *int64 `json:"failedActions,omitempty"`

	RunningActions *int64 `json:"runningActions,omitempty"`

	StoppedActions *int64 `json:"stoppedActions,omitempty"`

	SucceededActions *int64 `json:"succeededActions,omitempty"`

	TimeoutActions *int64 `json:"timeoutActions,omitempty"`

	TotalActions *int64 `json:"totalActions,omitempty"`
}

// +kubebuilder:skipversion
type XMLClassifier struct {
	Classification *string `json:"classification,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	Name *string `json:"name,omitempty"`

	RowTag *string `json:"rowTag,omitempty"`

	Version *int64 `json:"version,omitempty"`
}
