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
type ChangeSetSummary struct {
	ChangeSetID *string `json:"changeSetID,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	Description *string `json:"description,omitempty"`

	ParentChangeSetID *string `json:"parentChangeSetID,omitempty"`

	RootChangeSetID *string `json:"rootChangeSetID,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`
}

// +kubebuilder:skipversion
type Export struct {
	ExportingStackID *string `json:"exportingStackID,omitempty"`

	Name *string `json:"name,omitempty"`
}

// +kubebuilder:skipversion
type Output struct {
	Description *string `json:"description,omitempty"`

	ExportName *string `json:"exportName,omitempty"`

	OutputKey *string `json:"outputKey,omitempty"`

	OutputValue *string `json:"outputValue,omitempty"`
}

// +kubebuilder:skipversion
type Parameter struct {
	ParameterKey *string `json:"parameterKey,omitempty"`

	ParameterValue *string `json:"parameterValue,omitempty"`

	ResolvedValue *string `json:"resolvedValue,omitempty"`

	UsePreviousValue *bool `json:"usePreviousValue,omitempty"`
}

// +kubebuilder:skipversion
type ParameterDeclaration struct {
	DefaultValue *string `json:"defaultValue,omitempty"`

	Description *string `json:"description,omitempty"`

	ParameterKey *string `json:"parameterKey,omitempty"`
}

// +kubebuilder:skipversion
type ResourceChange struct {
	ChangeSetID *string `json:"changeSetID,omitempty"`

	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`
}

// +kubebuilder:skipversion
type ResourceIdentifierSummary struct {
	ResourceType *string `json:"resourceType,omitempty"`
}

// +kubebuilder:skipversion
type ResourceToImport struct {
	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`
}

// +kubebuilder:skipversion
type RollbackConfiguration struct {
	MonitoringTimeInMinutes *int64 `json:"monitoringTimeInMinutes,omitempty"`

	RollbackTriggers []*RollbackTrigger `json:"rollbackTriggers,omitempty"`
}

// +kubebuilder:skipversion
type RollbackTrigger struct {
	ARN *string `json:"arn,omitempty"`

	Type *string `json:"type_,omitempty"`
}

// +kubebuilder:skipversion
type StackDriftInformation struct {
	LastCheckTimestamp *metav1.Time `json:"lastCheckTimestamp,omitempty"`

	StackDriftStatus *string `json:"stackDriftStatus,omitempty"`
}

// +kubebuilder:skipversion
type StackDriftInformationSummary struct {
	LastCheckTimestamp *metav1.Time `json:"lastCheckTimestamp,omitempty"`

	StackDriftStatus *string `json:"stackDriftStatus,omitempty"`
}

// +kubebuilder:skipversion
type StackEvent struct {
	ClientRequestToken *string `json:"clientRequestToken,omitempty"`

	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`

	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackInstance struct {
	DriftStatus *string `json:"driftStatus,omitempty"`

	LastDriftCheckTimestamp *metav1.Time `json:"lastDriftCheckTimestamp,omitempty"`

	ParameterOverrides []*Parameter `json:"parameterOverrides,omitempty"`

	StackID *string `json:"stackID,omitempty"`
}

// +kubebuilder:skipversion
type StackInstanceSummary struct {
	DriftStatus *string `json:"driftStatus,omitempty"`

	LastDriftCheckTimestamp *metav1.Time `json:"lastDriftCheckTimestamp,omitempty"`

	StackID *string `json:"stackID,omitempty"`
}

// +kubebuilder:skipversion
type StackResource struct {
	Description *string `json:"description,omitempty"`

	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`

	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackResourceDetail struct {
	Description *string `json:"description,omitempty"`

	LastUpdatedTimestamp *metav1.Time `json:"lastUpdatedTimestamp,omitempty"`

	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`
}

// +kubebuilder:skipversion
type StackResourceDrift struct {
	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackResourceDriftInformation struct {
	LastCheckTimestamp *metav1.Time `json:"lastCheckTimestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackResourceDriftInformationSummary struct {
	LastCheckTimestamp *metav1.Time `json:"lastCheckTimestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackResourceSummary struct {
	LastUpdatedTimestamp *metav1.Time `json:"lastUpdatedTimestamp,omitempty"`

	LogicalResourceID *string `json:"logicalResourceID,omitempty"`

	ResourceType *string `json:"resourceType,omitempty"`
}

// +kubebuilder:skipversion
type StackSet struct {
	AdministrationRoleARN *string `json:"administrationRoleARN,omitempty"`

	Capabilities []*string `json:"capabilities,omitempty"`

	Description *string `json:"description,omitempty"`

	Parameters []*Parameter `json:"parameters,omitempty"`

	Tags []*Tag `json:"tags,omitempty"`

	TemplateBody *string `json:"templateBody,omitempty"`
}

// +kubebuilder:skipversion
type StackSetDriftDetectionDetails struct {
	LastDriftCheckTimestamp *metav1.Time `json:"lastDriftCheckTimestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackSetOperation struct {
	AdministrationRoleARN *string `json:"administrationRoleARN,omitempty"`

	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`

	EndTimestamp *metav1.Time `json:"endTimestamp,omitempty"`

	OperationID *string `json:"operationID,omitempty"`
}

// +kubebuilder:skipversion
type StackSetOperationSummary struct {
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`

	EndTimestamp *metav1.Time `json:"endTimestamp,omitempty"`

	OperationID *string `json:"operationID,omitempty"`
}

// +kubebuilder:skipversion
type StackSetSummary struct {
	Description *string `json:"description,omitempty"`

	DriftStatus *string `json:"driftStatus,omitempty"`

	LastDriftCheckTimestamp *metav1.Time `json:"lastDriftCheckTimestamp,omitempty"`
}

// +kubebuilder:skipversion
type StackSummary struct {
	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	DeletionTime *metav1.Time `json:"deletionTime,omitempty"`

	LastUpdatedTime *metav1.Time `json:"lastUpdatedTime,omitempty"`

	ParentID *string `json:"parentID,omitempty"`

	RootID *string `json:"rootID,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`

	StackStatus *string `json:"stackStatus,omitempty"`

	StackStatusReason *string `json:"stackStatusReason,omitempty"`
}

// +kubebuilder:skipversion
type Stack_SDK struct {
	Capabilities []*string `json:"capabilities,omitempty"`

	ChangeSetID *string `json:"changeSetID,omitempty"`

	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	DeletionTime *metav1.Time `json:"deletionTime,omitempty"`

	Description *string `json:"description,omitempty"`

	DisableRollback *bool `json:"disableRollback,omitempty"`
	// Contains information about whether the stack's actual configuration differs,
	// or has drifted, from its expected configuration, as defined in the stack
	// template and any values specified as template parameters. A stack is considered
	// to have drifted if one or more of its resources have drifted.
	DriftInformation *StackDriftInformation `json:"driftInformation,omitempty"`

	EnableTerminationProtection *bool `json:"enableTerminationProtection,omitempty"`

	LastUpdatedTime *metav1.Time `json:"lastUpdatedTime,omitempty"`

	NotificationARNs []*string `json:"notificationARNs,omitempty"`

	Outputs []*Output `json:"outputs,omitempty"`

	Parameters []*Parameter `json:"parameters,omitempty"`

	ParentID *string `json:"parentID,omitempty"`

	RoleARN *string `json:"roleARN,omitempty"`
	// Structure containing the rollback triggers for CloudFormation to monitor
	// during stack creation and updating operations, and for the specified monitoring
	// period afterwards.
	//
	// Rollback triggers enable you to have CloudFormation monitor the state of
	// your application during stack creation and updating, and to roll back that
	// operation if the application breaches the threshold of any of the alarms
	// you've specified. For more information, see Monitor and Roll Back Stack Operations
	// (http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-rollback-triggers.html).
	RollbackConfiguration *RollbackConfiguration `json:"rollbackConfiguration,omitempty"`

	RootID *string `json:"rootID,omitempty"`

	StackID *string `json:"stackID,omitempty"`

	StackName *string `json:"stackName,omitempty"`

	StackStatus *string `json:"stackStatus,omitempty"`

	StackStatusReason *string `json:"stackStatusReason,omitempty"`

	Tags []*Tag `json:"tags,omitempty"`

	TimeoutInMinutes *int64 `json:"timeoutInMinutes,omitempty"`
}

// +kubebuilder:skipversion
type Tag struct {
	Key *string `json:"key,omitempty"`

	Value *string `json:"value,omitempty"`
}

// +kubebuilder:skipversion
type TemplateParameter struct {
	DefaultValue *string `json:"defaultValue,omitempty"`

	Description *string `json:"description,omitempty"`

	ParameterKey *string `json:"parameterKey,omitempty"`
}

// +kubebuilder:skipversion
type TypeConfigurationDetails struct {
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:skipversion
type TypeSummary struct {
	Description *string `json:"description,omitempty"`

	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:skipversion
type TypeVersionSummary struct {
	Description *string `json:"description,omitempty"`

	TimeCreated *metav1.Time `json:"timeCreated,omitempty"`
}
