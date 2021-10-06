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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// RepositoryParameters define the desired state of an AWS Elastic Container Repository
type RepositoryParameters struct {

	// Region is the region you'd like your Repository to be created in.
	Region string `json:"region"`

	// The image scanning configuration for the repository. This determines whether
	// images are scanned for known vulnerabilities after being pushed to the repository.
	// +optional
	ImageScanningConfiguration *ImageScanningConfiguration `json:"imageScanningConfiguration,omitempty"`

	// The tag mutability setting for the repository. If this parameter is omitted,
	// the default setting of MUTABLE will be used which will allow image tags to
	// be overwritten. If IMMUTABLE is specified, all image tags within the repository
	// will be immutable which will prevent them from being overwritten.
	// +optional
	// +kubebuilder:validation:Enum=MUTABLE;IMMUTABLE
	ImageTagMutability *string `json:"imageTagMutability,omitempty"`

	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// If a repository contains images, forces the deletion.
	// +optional
	ForceDelete *bool `json:"forceDelete,omitempty"`

	// LifecyclePolicy contains one or more rules, where each rule defines an action for Amazon ECR.
	// This provides a way to automate the cleaning up of unused images, for example expiring images based on age or count
	// +optional
	LifecyclePolicy *LifecyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// LifecyclePolicy represents the rules to be applied to the repository
// https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html
type LifecyclePolicy struct {
	// Rules are separate lifecycle rules to be applied to the repository.
	Rules []LifecyclePolicyRule `json:"rules,omitempty"`
	// LifecyclePolicyString is a representation of all rules jsonified
	// LifecyclePolicyString *string `json:"lifecyclePolicyString,omitempty"`
}

// LifecyclePolicyRule defines the actions on the ECR repsitory
// https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html
type LifecyclePolicyRule struct {
	// RulePriority is required and represents the order in which the rules are applied
	RulePriority int `json:"rulePriority"`
	// Description describes the purpose of the rule
	// +optional
	Description string `json:"description,omitempty"`
	// Selection provides information about which objects to be handled
	Selection LifecyclePolicySelection `json:"selection"`
	// Action defines an action type, where the only sypported value is 'expire'
	// +kubebuilder:validation:Enum=expire
	Action LifecyclePolicyAction `json:"action"`
}

// LifecyclePolicySelection defines which objects to be handled
type LifecyclePolicySelection struct {
	// TagStatus determines whether the lifecycle policy rule that you are adding specifies a tag for an image.
	// Acceptable options are tagged, untagged, or any. If you specify any, then all images have the rule evaluated against them.
	// If you specify tagged, then you must also specify a tagPrefixList value. If you specify untagged, then you must omit tagPrefixList.
	// +kubebuilder:validation:Enum=tagged;untagged;any
	TagStatus string `json:"tagStatus"`

	// TagPrefixList specifies a comma-separated list of image tag prefixes on which to take action with your lifecycle policy.
	// For example, if your images are tagged as prod, prod1, prod2, and so on, you would use the tag prefix prod to specify all of them.
	// If you specify multiple tags, only the images with all specified tags are selected.
	// Required: yes, only if tagStatus is set to tagged
	TagPrefixList string `json:"tagPrefixList,omitempty"`

	// CountType selects what type to count objects on
	// If countType is set to imageCountMoreThan, you also specify countNumber to create a rule that sets a limit on the number of images that exist in your repository. If countType is set to sinceImagePushed, you also specify countUnit and countNumber to specify a time limit on the images that exist in your repository.
	// +kubebuilder:validation:Enum=imageCountMoreThan;sinceImagePushed
	CountType string `json:"countType"`

	// CountUnit
	// Required: yes, only if countType is set to sinceImagePushed
	// Specify a count unit of days to indicate that as the unit of time, in addition to countNumber, which is the number of days.
	// This should only be specified when countType is sinceImagePushed; an error will occur if you specify a count unit when countType is any other value.
	// +optional
	// +kubebuilder:validation:Enum=days
	CountUnit string `json:"countUnit,omitempty"`

	// CountNumber is the number of images or days based on CountType
	// Specify a count number. Acceptable values are positive integers (0 is not an accepted value).
	// If the countType used is imageCountMoreThan, then the value is the maximum number of images that you want to retain in your repository.
	// If the countType used is sinceImagePushed, then the value is the maximum age limit for your images.
	// +kubebuilder:validation:Minimum=1
	CountNumber int `json:"countNumber"`
}

// LifecyclePolicyAction represents the actions to be taken on a selection of images
type LifecyclePolicyAction struct {
	// Type is what action to be taken when Selection matches
	// +kubebuilder:validation:Enum=expire
	Type string `json:"expore"`
}

// Tag defines a tag
type Tag struct {

	// Key is the name of the tag.
	Key string `json:"key"`

	// Value is the value of the tag.
	Value string `json:"value"`
}

// A RepositorySpec defines the desired state of a Elastic Container Repository.
type RepositorySpec struct {
	xpv1.ResourceSpec `json:",inline"`

	ForProvider RepositoryParameters `json:"forProvider"`
}

// RepositoryObservation keeps the state for the external resource
type RepositoryObservation struct {
	// The date and time, in JavaScript date format, when the repository was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The AWS account ID associated with the registry that contains the repository.
	RegistryID string `json:"registryId"`

	// The Amazon Resource Name (ARN) that identifies the repository. The ARN contains
	// the arn:aws:ecr namespace, followed by the region of the repository, AWS
	// account ID of the repository owner, repository namespace, and repository
	// name. For example, arn:aws:ecr:region:012345678910:repository/test.
	RepositoryArn string `json:"repositoryArn,omitempty"`

	// The name of the repository.
	RepositoryName string `json:"repositoryName,omitempty"`

	// The URI for the repository. You can use this URI for container image push
	// and pull operations.
	RepositoryURI string `json:"repositoryUri,omitempty"`
}

// ImageScanningConfiguration Scanning Configuration
type ImageScanningConfiguration struct {

	// The setting that determines whether images are scanned after being pushed
	// to a repository. If set to true, images will be scanned after being pushed.
	// If this parameter is not specified, it will default to false and images will
	// not be scanned unless a scan is manually started with the StartImageScan
	// API.
	ScanOnPush bool `json:"scanOnPush"`
}

// A RepositoryStatus represents the observed state of a Elastic Container Repository.
type RepositoryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepositoryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Repository is a managed resource that represents an Elastic Container Repository
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="URI",type="string",JSONPath=".status.atProvider.repositoryUri"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepositoryList contains a list of ECRs
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}
