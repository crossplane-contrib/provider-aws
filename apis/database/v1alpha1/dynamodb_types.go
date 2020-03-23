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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// RDS instance states.
const (
	DynamoTableStateAvailable = "ACTIVE"
	DynamoTableStateCreating  = "CREATING"
	DynamoTableStateDeleting  = "DELETING"
	DynamoTableStateModifying = "UPDATING"
)

// AttributeDefinition  represents an attribute for describing the key schema for the table and indexes.
type AttributeDefinition struct {
	// A name for the attribute.
	AttributeName string `json:"attributeName"`

	// The data type for the attribute, where:
	//
	//    * S - the attribute is of type String
	//
	//    * N - the attribute is of type Number
	//
	//    * B - the attribute is of type Binary
	//
	AttributeType string `json:"attributeType"`
}

// GlobalSecondaryIndex represents the properties of a global secondary index.
type GlobalSecondaryIndex struct {
	// The name of the global secondary index. The name must be unique among all
	IndexName *string `json:"indexName,omitempty"`

	// The complete key schema for a global secondary index, which consists of one
	KeySchema []KeySchemaElement `json:"keySchema,omitempty"`

	// Represents attributes that are copied (projected) from the table into the
	// global secondary index. These are in addition to the primary key attributes
	// and index key attributes, which are automatically projected.
	Projection *Projection `json:"projection,omitempty"`

	// Represents the provisioned throughput settings for the specified global secondary
	// index.
	ProvisionedThroughput *ProvisionedThroughput `json:"provisionedThroughput,omitempty"`
}

// KeySchemaElement represents a single element of a key schema which make up the primary key.
type KeySchemaElement struct {
	// The name of a key attribute.
	AttributeName string `json:"attributeName"`

	// The role that this key attribute will assume:
	KeyType string `json:"keyType"`
}

// LocalSecondaryIndex represents the properties of a local secondary index.
type LocalSecondaryIndex struct {
	// The name of the local secondary index. The name must be unique among all
	// other indexes on this table.
	IndexName *string `json:"indexName,omitempty"`

	// The complete key schema for the local secondary index, consisting of one
	KeySchema []KeySchemaElement `json:"keySchema,omitempty"`

	// Represents attributes that are copied (projected) from the table into the
	// local secondary index.
	Projection *Projection `json:"projection,omitempty"`
}

// Projection represents attributes that are projected from the table into an index.
type Projection struct {

	// Represents the non-key attribute names which will be projected into the index.
	// +optional
	NonKeyAttributes []string `json:"keyType"`

	// The set of attributes that are projected into the index:
	ProjectionType string `json:"projectionType"`
}

// ProvisionedThroughput represents the provisioned throughput settings for a specified table or index.
type ProvisionedThroughput struct {

	// The maximum number of strongly consistent reads consumed per second before
	ReadCapacityUnits *int64 `json:"readCapacityUnits,omitempty"`

	// The maximum number of writes consumed per second before DynamoDB returns
	// a ThrottlingException.
	WriteCapacityUnits *int64 `json:"writeCapacityUnits,omitempty"`
}

// SSESpecification specifies the settings used to enable server-side encryption.
type SSESpecification struct {

	// Indicates whether server-side encryption is done using an AWS managed CMK
	// or an AWS owned CMK.
	Enabled *bool `json:"enabled,omitempty"`

	// The AWS KMS customer master key (CMK) that should be used for the AWS KMS
	// encryption.
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// Server-side encryption type.
	SSEType *string `json:"SSEType,omitempty"`
	// contains filtered or unexported fields
}

// StreamSpecification specifies the settings for DynamoDB Streams on the table.
type StreamSpecification struct {

	// Indicates whether DynamoDB Streams is enabled (true) or disabled (false)
	// on the table.
	StreamEnabled *bool `json:"streamEnabled,omitempty"`

	// When an item in the table is modified, StreamViewType determines what information
	// is written to the stream for this table.
	StreamViewType string `json:"StreamViewType,omitempty"`
}

// DynamoTableParameters define the desired state of an AWS DynomoDBTable
type DynamoTableParameters struct {
	// An array of attributes that describe the key schema for the table and indexes.
	AttributeDefinitions []AttributeDefinition `json:"attributeDefinitions"`

	// One or more global secondary indexes (the maximum is 20) to be created on
	// the table.
	// +optional
	GlobalSecondaryIndexes []GlobalSecondaryIndex `json:"globalSecondaryIndexes,omitempty"`

	// KeySchema specifies the attributes that make up the primary key for a table or an index.
	// +immutable
	KeySchema []KeySchemaElement `json:"keySchema"`

	// One or more local secondary indexes (the maximum is 5) to be created on the
	// table.
	// +optional
	// +immutable
	LocalSecondaryIndexes []LocalSecondaryIndex `json:"localSecondaryIndexes,omitempty"`

	// Represents the provisioned throughput settings for a specified table or index.
	// +optional
	ProvisionedThroughput *ProvisionedThroughput `json:"provisionedThroughput,omitempty"`

	// Represents the settings used to enable server-side encryption.
	// +optional
	SSESpecification *SSESpecification `json:"sseSpecification,omitempty"`

	// The stream settings for DynamoDB Streams on the table. These settings consist of:
	// +optional
	StreamSpecification *StreamSpecification `json:"streamSpecification,omitempty"`

	// A list of key-value pairs to label the table. For more information.
	// +optional
	// Tags []*Tag `json:"tag,omitempty"`
}

// A DynamoTableSpec defines the desired state of a DynamoDB Table.
type DynamoTableSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  DynamoTableParameters `json:"forProvider"`
}

// DynamoTableObservation keeps the state for the external resource
type DynamoTableObservation struct {

	// An array of AttributeDefinition objects. Each of these objects describes
	// one attribute in the table and index key schema.
	AttributeDefinitions []AttributeDefinition `json:"attributeDefinitions,omitempty"`

	// // The date and time when the table was created, in UNIX epoch time (http://www.epochconverter.com/)
	// // format.
	// CreationDateTime *time.Time `json:"creationDateTime,omitempty"`

	// The global secondary indexes, if any, on the table. Each index is scoped
	// to a given partition key value.
	GlobalSecondaryIndexes []GlobalSecondaryIndex `json:"globalSecondaryIndexes,omitempty"`

	// Represents the version of global tables (https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GlobalTables.html)
	// in use, if the table is replicated across AWS Regions.
	GlobalTableVersion string `json:"globalTableVersion,omitempty"`

	// The number of items in the specified table.
	ItemCount int64 `json:"itemCount,omitempty"`

	// The primary key structure for the table. Each KeySchemaElement consists of:
	KeySchema []KeySchemaElement `json:"keySchema,omitempty"`

	// Represents one or more local secondary indexes on the table.
	LocalSecondaryIndexes []LocalSecondaryIndex `json:"localSecondaryIndexes,omitempty"`

	// The provisioned throughput settings for the table, consisting of read and
	// write capacity units, along with data about increases and decreases.
	ProvisionedThroughput ProvisionedThroughput `json:"provisionedThroughput,omitempty"`

	// The Amazon Resource Name (ARN) that uniquely identifies the table.
	TableArn string `json:"tableArn,omitempty"`

	// Unique identifier for the table for which the backup was created.
	TableID string `json:"tableId,omitempty"`

	// The current state of the table:
	TableStatus string `json:"tableStatus,omitempty"`

	// Unique identifier for the table for which the backup was created.
	TableName string `json:"tableName,omitempty"`
}

// A DynamoTableStatus represents the observed state of a DBSubnetGroup.
type DynamoTableStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     DynamoTableObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A DynamoTable is a managed resource that represents an AWS DynamoDB Table
// +kubebuilder:printcolumn:name="TABLE_NAME",type="string",JSONPath=".status.atProvider.tableName"
// +kubebuilder:printcolumn:name="TABLE_STATUS",type="string",JSONPath=".status.atProvider.tableStatus"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type DynamoTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamoTableSpec   `json:"spec"`
	Status DynamoTableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DynamoTableList contains a list of DynamoTable
type DynamoTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamoTable `json:"items"`
}
