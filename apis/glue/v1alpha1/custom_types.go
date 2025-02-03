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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomJobParameters contains the additional fields for JobParameters.
type CustomJobParameters struct {
	// The connections used for this job.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Connection
	// +crossplane:generate:reference:refFieldName=ConnectionRefs
	// +crossplane:generate:reference:selectorFieldName=ConnectionSelector
	Connections []*string `json:"connections,omitempty"`

	// ConnectionRefs is a list of references to Connections used to set
	// the Connections.
	// +optional
	ConnectionRefs []xpv1.Reference `json:"connectionRefs,omitempty"`

	// ConnectionsSelector selects references to Connections used
	// to set the Connections.
	// +optional
	ConnectionSelector *xpv1.Selector `json:"connectionSelector,omitempty"`

	// The name or Amazon Resource Name (ARN) of the IAM role associated with this
	// job. Role is a required field
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=RoleRef
	// +crossplane:generate:reference:selectorFieldName=RoleSelector
	Role string `json:"role,omitempty"`

	// RoleRef is a reference to an IAMRole used to set
	// the Role.
	// +immutable
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to IAMRole used
	// to set the Role.
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`

	// The name of the SecurityConfiguration structure to be used with this job.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.SecurityConfiguration
	// +crossplane:generate:reference:refFieldName=SecurityConfigurationRef
	// +crossplane:generate:reference:selectorFieldName=SecurityConfigurationSelector
	SecurityConfiguration *string `json:"securityConfiguration,omitempty"`

	// SecurityConfigurationRef is a reference to an SecurityConfiguration used to set
	// the SecurityConfiguration.
	// +optional
	SecurityConfigurationRef *xpv1.Reference `json:"securityConfigurationRef,omitempty"`

	// SecurityConfigurationSelector selects references to SecurityConfiguration used
	// to set the SecurityConfiguration.
	// +optional
	SecurityConfigurationSelector *xpv1.Selector `json:"securityConfigurationSelector,omitempty"`
}

// CustomJobObservation includes the custom status fields of Job.
type CustomJobObservation struct{}

// CustomSecurityConfigurationParameters contains the additional fields for SecurityConfigurationParameters
type CustomSecurityConfigurationParameters struct {
	// The encryption configuration for the new security configuration.
	CustomEncryptionConfiguration *CustomEncryptionConfiguration `json:"encryptionConfiguration"`
}

// CustomJobObservation includes the custom status fields of SecurityConfiguration.
type CustomSecurityConfigurationObservation struct{}

// CustomEncryptionConfiguration contains the additional fields for EncryptionConfiguration
type CustomEncryptionConfiguration struct {
	// Specifies how Amazon CloudWatch data should be encrypted.
	// +optional
	CustomCloudWatchEncryption *CustomCloudWatchEncryption `json:"cloudWatchEncryption,omitempty"`

	// Specifies how job bookmark data should be encrypted.
	// +optional
	CustomJobBookmarksEncryption *CustomJobBookmarksEncryption `json:"jobBookmarksEncryption,omitempty"`

	// Specifies how Amazon Simple Storage Service (Amazon S3) data should be encrypted.
	// +optional
	CustomS3Encryption []*CustomS3Encryption `json:"s3Encryption,omitempty"`
}

// CustomS3Encryption contains the additional fields for S3Encryption
type CustomS3Encryption struct {
	// The encryption mode to use for Amazon S3 data.
	// +kubebuilder:validation:Enum=DISABLED;SSE-KMS;SSE-S3
	S3EncryptionMode *string `json:"s3EncryptionMode,omitempty"`

	// The Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.KMSKeyARN()
	// +crossplane:generate:reference:refFieldName=KMSKeyARNRef
	// +crossplane:generate:reference:selectorFieldName=KMSKeyARNSelector
	KMSKeyARN *string `json:"kmsKeyArn,omitempty"`

	// KMSKeyARNRef is a reference to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyArnRef,omitempty"`

	// KMSKeyARNSelector selects references to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyArnSelector,omitempty"`
}

// CustomJobBookmarksEncryption contains the additional fields for JobBookmarksEncryption
type CustomJobBookmarksEncryption struct {
	// The encryption mode to use for job bookmarks data.
	// +kubebuilder:validation:Enum=DISABLED;CSE-KMS
	JobBookmarksEncryptionMode *string `json:"jobBookmarksEncryptionMode,omitempty"`

	// The Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.KMSKeyARN()
	// +crossplane:generate:reference:refFieldName=KMSKeyARNRef
	// +crossplane:generate:reference:selectorFieldName=KMSKeyARNSelector
	KMSKeyARN *string `json:"kmsKeyArn,omitempty"`

	// KMSKeyARNRef is a reference to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyArnRef,omitempty"`

	// KMSKeyARNSelector selects references to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyArnSelector,omitempty"`
}

// CustomCloudWatchEncryption contains the additional fields for CloudWatchEncryption
type CustomCloudWatchEncryption struct {
	// The encryption mode to use for CloudWatch data.
	// +kubebuilder:validation:Enum=DISABLED;SSE-KMS
	CloudWatchEncryptionMode *string `json:"cloudWatchEncryptionMode,omitempty"`

	// The Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.KMSKeyARN()
	// +crossplane:generate:reference:refFieldName=KMSKeyARNRef
	// +crossplane:generate:reference:selectorFieldName=KMSKeyARNSelector
	KMSKeyARN *string `json:"kmsKeyArn,omitempty"`

	// KMSKeyARNRef is a reference to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyArnRef,omitempty"`

	// KMSKeyARNSelector selects references to an KMSKey used to set the KMSKeyARN.
	// +optional
	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyArnSelector,omitempty"`
}

// CustomConnectionParameters contains the additional fields for ConnectionParameters
type CustomConnectionParameters struct {
	// A ConnectionInput object defining the connection to create.
	CustomConnectionInput *CustomConnectionInput `json:"connectionInput"`
}

// CustomConnectionObservation includes the custom status fields of Connection.
type CustomConnectionObservation struct{}

// CustomConnectionInput contains the additional fields for ConnectionInput
type CustomConnectionInput struct {
	// These key-value pairs define parameters for the connection.
	// Possible keys for connection properties:
	// "HOST"|"PORT"|"USERNAME"|"PASSWORD"|"ENCRYPTED_PASSWORD"|"JDBC_DRIVER_JAR_URI"
	// "JDBC_DRIVER_CLASS_NAME"|"JDBC_ENGINE"|"JDBC_ENGINE_VERSION"|"CONFIG_FILES"
	// "INSTANCE_ID"|"JDBC_CONNECTION_URL"|"JDBC_ENFORCE_SSL"|"CUSTOM_JDBC_CERT"
	// "SKIP_CUSTOM_JDBC_CERT_VALIDATION"|"CUSTOM_JDBC_CERT_STRING"|"CONNECTION_URL"
	// "KAFKA_BOOTSTRAP_SERVERS"|"KAFKA_SSL_ENABLED"|"KAFKA_CUSTOM_CERT"
	// "KAFKA_SKIP_CUSTOM_CERT_VALIDATION"|"KAFKA_CLIENT_KEYSTORE"
	// "KAFKA_CLIENT_KEYSTORE_PASSWORD"|"KAFKA_CLIENT_KEY_PASSWORD"
	// "ENCRYPTED_KAFKA_CLIENT_KEYSTORE_PASSWORD"|"ENCRYPTED_KAFKA_CLIENT_KEY_PASSWORD"
	// "SECRET_ID"|"CONNECTOR_URL"|"CONNECTOR_TYPE"|"CONNECTOR_CLASS_NAME"
	//
	// ConnectionProperties is a required field
	// +kubebuilder:validation:Required
	ConnectionProperties map[string]*string `json:"connectionProperties"`

	// The type of the connection. Currently, these types are supported:
	//
	//    * JDBC - Designates a connection to a database through Java Database Connectivity
	//    (JDBC).
	//
	//    * KAFKA - Designates a connection to an Apache Kafka streaming platform.
	//
	//    * MONGODB - Designates a connection to a MongoDB document database.
	//
	//    * NETWORK - Designates a network connection to a data source within an
	//    Amazon Virtual Private Cloud environment (Amazon VPC).
	//
	//    * MARKETPLACE - Uses configuration settings contained in a connector purchased
	//    from Amazon Web Services Marketplace to read from and write to data stores
	//    that are not natively supported by Glue.
	//
	//    * CUSTOM - Uses configuration settings contained in a custom connector
	//    to read from and write to data stores that are not natively supported
	//    by Glue.
	//
	// SFTP is not supported.
	//
	// ConnectionType is a required field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=JDBC;KAFKA;MONGODB;NETWORK;MARKETPLACE;CUSTOM
	ConnectionType string `json:"connectionType"`

	// The description of the connection.
	Description *string `json:"description,omitempty"`

	// A list of criteria that can be used in selecting this connection.
	MatchCriteria []*string `json:"matchCriteria,omitempty"`

	// Specifies the physical requirements for a connection.
	CustomPhysicalConnectionRequirements *CustomPhysicalConnectionRequirements `json:"physicalConnectionRequirements,omitempty"`
}

// CustomPhysicalConnectionRequirements contains the additional fields for PhysicalConnectionRequirements
type CustomPhysicalConnectionRequirements struct {
	// The connection's Availability Zone. This field is redundant because the specified
	// subnet implies the Availability Zone to be used. Currently the field must
	// be populated, but it will be removed in the future.
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// The security group ID list used by the connection.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDList []string `json:"securityGroupIdList,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// The subnet ID used by the connection.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRef
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRef is a reference to SubnetID used to set the SubnetID.
	// +immutable
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector selects a reference to SubnetID used to set the SubnetID.
	// +immutable
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}

// CustomDatabaseParameters contains the additional fields for DatabaseParameters
type CustomDatabaseParameters struct {
	// The metadata for the database.
	CustomDatabaseInput *CustomDatabaseInput `json:"databaseInput,omitempty"`
}

// CustomDatabaseObservation includes the custom status fields of DatabaseParameters.
type CustomDatabaseObservation struct{}

// CustomDatabaseInput contains the fields for DatabaseInput.
type CustomDatabaseInput struct {

	// Creates a set of default permissions on the table for principals.
	// If left empty on creation, AWS defaults it to
	// [Permissions: ["All"], Principal: DataLake Prinicpal Identifier : "IAM_ALLOWED_PRINCIPALS"]
	CreateTableDefaultPermissions []*PrincipalPermissions `json:"createTableDefaultPermissions,omitempty"`

	// A description of the database.
	// +optional
	Description *string `json:"description,omitempty"`

	// The location of the database (for example, an HDFS path).
	// +optional
	LocationURI *string `json:"locationURI,omitempty"`

	// These key-value pairs define parameters and properties of the database.
	// +optional
	Parameters map[string]*string `json:"parameters,omitempty"`

	// A structure that describes a target database for resource linking.
	// +optional
	TargetDatabase *DatabaseIdentifier `json:"targetDatabase,omitempty"`
}

// CustomCrawlerParameters contains the additional fields for CrawlerParameters
type CustomCrawlerParameters struct {
	// A list of custom classifiers that the user has registered. By default, all
	// built-in classifiers are included in a crawl, but these custom classifiers
	// always override the default classifiers for a given classification.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Classifier
	// +crossplane:generate:reference:refFieldName=ClassifierRefs
	// +crossplane:generate:reference:selectorFieldName=ClassifierSelector
	Classifiers []*string `json:"classifiers,omitempty"`

	// ClassifierRefs is a list of references to Classifiers used to set
	// the Classifiers.
	// +optional
	ClassifierRefs []xpv1.Reference `json:"classifierRefs,omitempty"`

	// ClassifiersSelector selects references to Classifiers used
	// to set the Classifiers.
	// +optional
	ClassifierSelector *xpv1.Selector `json:"classifierSelector,omitempty"`

	// The name of the SecurityConfiguration structure to be used by this crawler.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.SecurityConfiguration
	// +crossplane:generate:reference:refFieldName=CrawlerSecurityConfigurationRef
	// +crossplane:generate:reference:selectorFieldName=CrawlerSecurityConfigurationSelector
	CrawlerSecurityConfiguration *string `json:"crawlerSecurityConfiguration,omitempty"`

	// CrawlerSecurityConfigurationRef is a reference to an SecurityConfiguration used to set
	// the CrawlerSecurityConfiguration.
	// +optional
	CrawlerSecurityConfigurationRef *xpv1.Reference `json:"crawlerSecurityConfigurationRef,omitempty"`

	// CrawlerSecurityConfigurationSelector selects references to SecurityConfiguration used
	// to set the CrawlerSecurityConfiguration.
	// +optional
	CrawlerSecurityConfigurationSelector *xpv1.Selector `json:"crawlerSecurityConfigurationSelector,omitempty"`

	// The Glue database where results are written, such as: arn:aws:daylight:us-east-1::database/sometable/*.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Database
	// +crossplane:generate:reference:refFieldName=DatabaseNameRef
	// +crossplane:generate:reference:selectorFieldName=DatabaseNameSelector
	DatabaseName *string `json:"databaseName,omitempty"`

	// DatabaseNameRef is a reference to an Database used to set
	// the DatabaseName.
	// +optional
	DatabaseNameRef *xpv1.Reference `json:"databaseNameRef,omitempty"`

	// DatabaseNamesSelector selects references to Database used
	// to set the DatabaseName.
	// +optional
	DatabaseNameSelector *xpv1.Selector `json:"databaseNameSelector,omitempty"`

	// The IAM role or Amazon Resource Name (ARN) of an IAM role used by the new
	// crawler to access customer resources.
	// AWS API seems to give just name of the role back (not ARN)
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	// +crossplane:generate:reference:refFieldName=RoleRef
	// +crossplane:generate:reference:selectorFieldName=RoleSelector
	Role string `json:"role,omitempty"`

	// RoleRef is a reference to an IAMRole used to set
	// the Role.
	// +immutable
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to IAMRole used
	// to set the Role.
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`

	// A list of collection of targets to crawl.
	//
	// Targets is a required field
	// +kubebuilder:validation:Required
	Targets CustomCrawlerTargets `json:"targets"`
}

// CustomCrawlerObservation includes the custom status fields of Crawler.
type CustomCrawlerObservation struct{}

// CustomCrawlerTargets contains the additional fields for CrawlerTargets
type CustomCrawlerTargets struct {
	// Specifies Glue Data Catalog targets.
	CatalogTargets []*CustomCatalogTarget `json:"catalogTargets,omitempty"`

	// Specifies Amazon DynamoDB targets.
	DynamoDBTargets []*DynamoDBTarget `json:"dynamoDBTargets,omitempty"`

	// Specifies JDBC targets.
	JDBCTargets []*CustomJDBCTarget `json:"jdbcTargets,omitempty"`

	// Specifies Amazon DocumentDB or MongoDB targets.
	MongoDBTargets []*CustomMongoDBTarget `json:"mongoDBTargets,omitempty"`

	// Specifies Amazon Simple Storage Service (Amazon S3) targets.
	S3Targets []*CustomS3Target `json:"s3Targets,omitempty"`
}

// CustomCatalogTarget contains the additional fields for CatalogTarget
type CustomCatalogTarget struct {
	// The name of the database to be synchronized.
	//
	// DatabaseName is a required field
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Database
	// +crossplane:generate:reference:refFieldName=DatabaseNameRef
	// +crossplane:generate:reference:selectorFieldName=DatabaseNameSelector
	DatabaseName string `json:"databaseName,omitempty"`

	// DatabaseNameRef is a reference to an Database used to set
	// the DatabaseName.
	// +optional
	DatabaseNameRef *xpv1.Reference `json:"databaseNameRef,omitempty"`

	// DatabaseNamesSelector selects references to Database used
	// to set the DatabaseName.
	// +optional
	DatabaseNameSelector *xpv1.Selector `json:"databaseNameSelector,omitempty"`

	// A list of the tables to be synchronized.
	//
	// Tables is a required field
	// +kubebuilder:validation:Required
	Tables []string `json:"tables"`
}

// CustomJDBCTarget contains the additional fields for JdbcTarget
type CustomJDBCTarget struct {
	// The name of the connection to use to connect to the JDBC target.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Connection
	// +crossplane:generate:reference:refFieldName=ConnectionNameRef
	// +crossplane:generate:reference:selectorFieldName=ConnectionNameSelector
	ConnectionName *string `json:"connectionName,omitempty"`

	// ConnectionNameRef is a reference to an Connection used to set
	// the ConnectionName.
	// +optional
	ConnectionNameRef *xpv1.Reference `json:"connectionNameRef,omitempty"`

	// ConnectionNamesSelector selects references to Connection used
	// to set the ConnectionName.
	// +optional
	ConnectionNameSelector *xpv1.Selector `json:"connectionNameSelector,omitempty"`

	// A list of glob patterns used to exclude from the crawl. For more information,
	// see Catalog Tables with a Crawler (https://docs.aws.amazon.com/glue/latest/dg/add-crawler.html).
	Exclusions []*string `json:"exclusions,omitempty"`

	// The path of the JDBC target.
	Path *string `json:"path,omitempty"`
}

// CustomMongoDBTarget contains the additional fields for MongoDBTarget
type CustomMongoDBTarget struct {
	// The name of the connection to use to connect to the Amazon DocumentDB or
	// MongoDB target.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Connection
	// +crossplane:generate:reference:refFieldName=ConnectionNameRef
	// +crossplane:generate:reference:selectorFieldName=ConnectionNameSelector
	ConnectionName *string `json:"connectionName,omitempty"`

	// ConnectionNameRef is a reference to an Connection used to set
	// the ConnectionName.
	// +optional
	ConnectionNameRef *xpv1.Reference `json:"connectionNameRef,omitempty"`

	// ConnectionNamesSelector selects references to Connection used
	// to set the ConnectionName.
	// +optional
	ConnectionNameSelector *xpv1.Selector `json:"connectionNameSelector,omitempty"`

	// The path of the Amazon DocumentDB or MongoDB target (database/collection).
	Path *string `json:"path,omitempty"`

	// Indicates whether to scan all the records, or to sample rows from the table.
	// Scanning all the records can take a long time when the table is not a high
	// throughput table.
	//
	// A value of true means to scan all records, while a value of false means to
	// sample the records. If no value is specified, the value defaults to true.
	ScanAll *bool `json:"scanAll,omitempty"`
}

// CustomS3Target contains the additional fields for S3Target
type CustomS3Target struct {
	// The name of a connection which allows a job or crawler to access data in
	// Amazon S3 within an Amazon Virtual Private Cloud environment (Amazon VPC).
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1.Connection
	// +crossplane:generate:reference:refFieldName=ConnectionNameRef
	// +crossplane:generate:reference:selectorFieldName=ConnectionNameSelector
	ConnectionName *string `json:"connectionName,omitempty"`

	// ConnectionNameRef is a reference to an Connection used to set
	// the ConnectionName.
	// +optional
	ConnectionNameRef *xpv1.Reference `json:"connectionNameRef,omitempty"`

	// ConnectionNamesSelector selects references to Connection used
	// to set the ConnectionName.
	// +optional
	ConnectionNameSelector *xpv1.Selector `json:"connectionNameSelector,omitempty"`

	// A valid Amazon dead-letter SQS ARN. For example, arn:aws:sqs:region:account:deadLetterQueue.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1.QueueARN()
	// +crossplane:generate:reference:refFieldName=DlqEventQueueARNRef
	// +crossplane:generate:reference:selectorFieldName=DlqEventQueueARNSelector
	DlqEventQueueARN *string `json:"dlqEventQueueArn,omitempty"`

	// DlqEventQueueARNRef is a reference to an SQSEventQueue used to set
	// the DlqEventQueueARN.
	// +optional
	DlqEventQueueARNRef *xpv1.Reference `json:"dlqEventQueueArnRef,omitempty"`

	// DlqEventQueueARNSelector selects references to SQSEventQueue used
	// to set the DlqEventQueueARN.
	// +optional
	DlqEventQueueARNSelector *xpv1.Selector `json:"dlqEventQueueArnSelector,omitempty"`

	// A valid Amazon SQS ARN. For example, arn:aws:sqs:region:account:sqs.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1.QueueARN()
	// +crossplane:generate:reference:refFieldName=EventQueueARNRef
	// +crossplane:generate:reference:selectorFieldName=EventQueueARNSelector
	EventQueueARN *string `json:"eventQueueArn,omitempty"`

	// EventQueueARNRef is a reference to an SQSEventQueue used to set
	// the EventQueueARN.
	// +optional
	EventQueueARNRef *xpv1.Reference `json:"eventQueueArnRef,omitempty"`

	// EventQueueARNSelector selects references to SQSEventQueue used
	// to set the EventQueueARN.
	// +optional
	EventQueueARNSelector *xpv1.Selector `json:"eventQueueArnSelector,omitempty"`

	// A list of glob patterns used to exclude from the crawl. For more information,
	// see Catalog Tables with a Crawler (https://docs.aws.amazon.com/glue/latest/dg/add-crawler.html).
	Exclusions []*string `json:"exclusions,omitempty"`

	// The path to the Amazon S3 target.
	Path *string `json:"path,omitempty"`

	// Sets the number of files in each leaf folder to be crawled when crawling
	// sample files in a dataset. If not set, all the files are crawled. A valid
	// value is an integer between 1 and 249.
	SampleSize *int64 `json:"sampleSize,omitempty"`
}

// CustomClassifierParameters contains the additional fields for ClassifierParameters
type CustomClassifierParameters struct {
	// A CSVClassifier object specifying the classifier to create.
	CustomCSVClassifier *CustomCreateCSVClassifierRequest `json:"csvClassifier,omitempty"`

	// A XMLClassifier object specifying the classifier to create.
	CustomXMLClassifier *CustomCreateXMLClassifierRequest `json:"xmlClassifier,omitempty"`

	// A GrokClassifier object specifying the classifier to create.
	CustomGrokClassifier *CustomCreateGrokClassifierRequest `json:"grokClassifier,omitempty"`

	// A JsonClassifier object specifying the classifier to create.
	CustomJSONClassifier *CustomCreateJSONClassifierRequest `json:"jsonClassifier,omitempty"`
}

// CustomClassifierObservation includes the custom status fields of Classifier.
type CustomClassifierObservation struct{}

// CustomCreateGrokClassifierRequest contains the fields for CreateGrokClassifierRequest.
type CustomCreateGrokClassifierRequest struct {
	// An identifier of the data format that the classifier matches, such as Twitter,
	// JSON, Omniture logs, Amazon CloudWatch Logs, and so on.
	// +kubebuilder:validation:Required
	Classification string `json:"classification"`

	// Optional custom grok patterns used by this classifier.
	// +optional
	CustomPatterns *string `json:"customPatterns,omitempty"`

	// The grok pattern used by this classifier.
	// +kubebuilder:validation:Required
	GrokPattern string `json:"grokPattern"`
}

// CustomCreateJSONClassifierRequest contains the fields for CreateJSONClassifierRequest.
type CustomCreateJSONClassifierRequest struct {
	// A JsonPath string defining the JSON data for the classifier to classify.
	// Glue supports a subset of JsonPath, as described in Writing JsonPath Custom
	// Classifiers (https://docs.aws.amazon.com/glue/latest/dg/custom-classifier.html#custom-classifier-json).
	// +optional
	JSONPath *string `json:"jsonPath,omitempty"`
}

// CustomCreateXMLClassifierRequest contains the fields for CreateXMLClassifierRequest.
type CustomCreateXMLClassifierRequest struct {
	// An identifier of the data format that the classifier matches.
	// Classification is a required field
	// +kubebuilder:validation:Required
	Classification string `json:"classification"`

	// The XML tag designating the element that contains each record in an XML document
	// being parsed. This can't identify a self-closing element (closed by />).
	// An empty row element that contains only attributes can be parsed as long
	// as it ends with a closing tag (for example, <row item_a="A" item_b="B"></row>
	// is okay, but <row item_a="A" item_b="B" /> is not).
	RowTag *string `json:"rowTag,omitempty"`
}

// CustomCreateCSVClassifierRequest contains the fields for CreateCSVClassifierRequest.
type CustomCreateCSVClassifierRequest struct {
	// Enables the processing of files that contain only one column.
	// +optional
	AllowSingleColumn *bool `json:"allowSingleColumn,omitempty"`

	// Indicates whether the CSV file contains a header.
	// UNKNOWN = "Detect headings"
	// PRESENT = "Has headings"
	// ABSENT = "No headings"
	// +optional
	// +kubebuilder:validation:Enum=UNKNOWN;PRESENT;ABSENT
	ContainsHeader *string `json:"containsHeader,omitempty"`

	// A custom symbol to denote what separates each column entry in the row.
	// +optional
	Delimiter *string `json:"delimiter,omitempty"`

	// Specifies not to trim values before identifying the type of column values.
	// The default value is true.
	// +optional
	DisableValueTrimming *bool `json:"disableValueTrimming,omitempty"`

	// A list of strings representing column names.
	// +optional
	Header []*string `json:"header,omitempty"`

	// A custom symbol to denote what combines content into a single column value.
	// Must be different from the column delimiter.
	// +optional
	QuoteSymbol *string `json:"quoteSymbol,omitempty"`
}

type CustomTriggerParameters struct{}

// CustomTriggerObservation includes the custom status fields of Trigger.
type CustomTriggerObservation struct{}
