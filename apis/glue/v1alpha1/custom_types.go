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
	// The name or Amazon Resource Name (ARN) of the IAM role associated with this
	// job. RoleArn is a required field
	// +immutable
	RoleArn string `json:"roleArn,omitempty"`

	// RoleArnRef is a reference to an IAMRole used to set
	// the RoleArn.
	// +immutable
	// +optional
	RoleArnRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleArnSelector selects references to IAMRole used
	// to set the RoleArn.
	// +optional
	RoleArnSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`
}

// CustomSecurityConfigurationParameters contains the additional fields for SecurityConfigurationParameters
type CustomSecurityConfigurationParameters struct {
	// The encryption configuration for the new security configuration.
	CustomEncryptionConfiguration *CustomEncryptionConfiguration `json:"encryptionConfiguration"`
}

// CustomEncryptionConfiguration contains the additional fields for EncryptionConfiguration
type CustomEncryptionConfiguration struct {
	// Specifies how Amazon CloudWatch data should be encrypted.
	CustomCloudWatchEncryption *CustomCloudWatchEncryption `json:"cloudWatchEncryption,omitempty"`
	// Specifies how job bookmark data should be encrypted.
	CustomJobBookmarksEncryption *CustomJobBookmarksEncryption `json:"jobBookmarksEncryption,omitempty"`
}

// CustomJobBookmarksEncryption contains the additional fields for JobBookmarksEncryption
type CustomJobBookmarksEncryption struct {
	JobBookmarksEncryptionMode *string `json:"jobBookmarksEncryptionMode,omitempty"`

	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`

	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyARNRef,omitempty"`

	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyARNSelector,omitempty"`
}

// CustomCloudWatchEncryption contains the additional fields for CloudWatchEncryption
type CustomCloudWatchEncryption struct {
	CloudWatchEncryptionMode *string `json:"cloudWatchEncryptionMode,omitempty"`

	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`

	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyARNRef,omitempty"`

	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyARNSelector,omitempty"`
}

// CustomConnectionParameters contains the additional fields for ConnectionParameters
type CustomConnectionParameters struct {
	// A ConnectionInput object defining the connection to create.
	CustomConnectionInput *CustomConnectionInput `json:"connectionInput"`
}

// CustomConnectionInput contains the additional fields for ConnectionInput
type CustomConnectionInput struct {

	// These key-value pairs define parameters for the connection.
	ConnectionProperties map[string]*string `json:"connectionProperties,omitempty"`

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
	ConnectionType *string `json:"connectionType,omitempty"`

	// The description of the connection.
	Description *string `json:"description,omitempty"`

	// A list of criteria that can be used in selecting this connection.
	MatchCriteria []*string `json:"matchCriteria,omitempty"`

	// Specifies the physical requirements for a connection.
	CustomPhysicalConnectionRequirements *CustomPhysicalConnectionRequirements `json:"physicalConnectionRequirements,omitempty"`
}

// CustomPhysicalConnectionRequirements contains the additional fields for PhysicalConnectionRequirements
type CustomPhysicalConnectionRequirements struct {
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	SecurityGroupIDList []string `json:"securityGroupIDList,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIDRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIDSelector,omitempty"`

	SubnetID *string `json:"subnetID,omitempty"`

	// +immutable
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIDRef,omitempty"`

	// +immutable
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`
}

// CustomDatabaseParameters contains the additional fields for DatabaseParameters
type CustomDatabaseParameters struct {
	// The metadata for the database.
	CustomDatabaseInput *CustomDatabaseInput `json:"databaseInput,omitempty"`
}

// CustomDatabaseInput contains the fields for DatabaseInput.
type CustomDatabaseInput struct {
	// +optional
	Description *string `json:"description,omitempty"`

	// +optional
	LocationURI *string `json:"locationURI,omitempty"`

	// +optional
	Parameters map[string]*string `json:"parameters,omitempty"`

	// A structure that describes a target database for resource linking.
	// +optional
	TargetDatabase *DatabaseIdentifier `json:"targetDatabase,omitempty"`
}

// CustomCrawlerParameters contains the additional fields for CrawlerParameters
type CustomCrawlerParameters struct {
	// The IAM role or Amazon Resource Name (ARN) of an IAM role used by the new
	// crawler to access customer resources.
	// +immutable
	RoleArn string `json:"roleArn,omitempty"`

	// RoleArnRef is a reference to an IAMRole used to set
	// the RoleArn.
	// +immutable
	// +optional
	RoleArnRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleArnSelector selects references to IAMRole used
	// to set the RoleArn.
	// +optional
	RoleArnSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`
}

// CustomClassifierParameters contains the additional fields for ClassifierParameters
type CustomClassifierParameters struct {
	// A CsvClassifier object specifying the classifier to create.
	CustomCsvClassifier *CustomCreateCsvClassifierRequest `json:"csvClassifier,omitempty"`

	// A CsvClassifier object specifying the classifier to create.
	CustomXMLClassifier *CustomCreateXMLClassifierRequest `json:"xmlClassifier,omitempty"`

	// A GrokClassifier object specifying the classifier to create.
	CustomGrokClassifier *CustomCreateGrokClassifierRequest `json:"grokClassifier,omitempty"`

	// A JsonClassifier object specifying the classifier to create.
	CustomJSONClassifier *CustomCreateJSONClassifierRequest `json:"jsonClassifier,omitempty"`
}

// CustomCreateGrokClassifierRequest contains the fields for CreateGrokClassifierRequest.
type CustomCreateGrokClassifierRequest struct {
	// An identifier of the data format that the classifier matches, such as Twitter,
	// JSON, Omniture logs, Amazon CloudWatch Logs, and so on.
	// +optional
	Classification *string `json:"classification,omitempty"`

	// Optional custom grok patterns used by this classifier.
	// +optional
	CustomPatterns *string `json:"customPatterns,omitempty"`

	// The grok pattern used by this classifier.
	// +optional
	GrokPattern *string `json:"grokPattern,omitempty"`
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
	// +optional
	Classification *string `json:"classification,omitempty"`

	// The XML tag designating the element that contains each record in an XML document
	// being parsed. This can't identify a self-closing element (closed by />).
	// An empty row element that contains only attributes can be parsed as long
	// as it ends with a closing tag (for example, <row item_a="A" item_b="B"></row>
	// is okay, but <row item_a="A" item_b="B" /> is not).
	// +optional
	RowTag *string `json:"rowTag,omitempty"`
}

// CustomCreateCsvClassifierRequest contains the fields for CreateCsvClassifierRequest.
type CustomCreateCsvClassifierRequest struct {
	// Enables the processing of files that contain only one column.
	// +optional
	AllowSingleColumn *bool `json:"allowSingleColumn,omitempty"`

	// Indicates whether the CSV file contains a header.
	// +optional
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
