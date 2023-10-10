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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FargateProfileStatusType is a type of FargateProfile status.
type FargateProfileStatusType string

// Types of FargateProfile status.
const (
	FargateProfileStatusCreating     FargateProfileStatusType = "CREATING"
	FargateProfileStatusActive       FargateProfileStatusType = "ACTIVE"
	FargateProfileStatusDeleting     FargateProfileStatusType = "DELETING"
	FargateProfileStatusCreateFailed FargateProfileStatusType = "CREATE_FAILED"
	FargateProfileStatusDeleteFailed FargateProfileStatusType = "DELETE_FAILED"
)

// FargateProfileSelector is an object representing an AWS Fargate profile selector.
type FargateProfileSelector struct {

	// The Kubernetes labels that the selector should match. A pod must contain
	// all of the labels that are specified in the selector for it to be considered
	// a match.
	Labels map[string]string `json:"labels,omitempty"`

	// The Kubernetes namespace that the selector should match.
	Namespace *string `json:"namespace,omitempty"`
}

// FargateProfileObservation is the observed state of a FargateProfile.
type FargateProfileObservation struct {
	// The Unix epoch timestamp in seconds for when the Fargate profile was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The full Amazon Resource Name (ARN) of the Fargate profile.
	FargateProfileArn string `json:"fargateProfileArn,omitempty"`

	// The current status of the Fargate profile.
	Status FargateProfileStatusType `json:"status,omitempty"`
}

// FargateProfileParameters define the desired state of an AWS Elastic Kubernetes
// Service FargateProfile.
// All fields are immutable as it is not possible to update a Fargate profile.
type FargateProfileParameters struct {

	// Region is the region you'd like  the FargateProfile to be created in.
	// +immutable
	Region string `json:"region"`

	// The name of the Amazon EKS cluster to apply the Fargate profile to.
	//
	// ClusterName is a required field
	// +immutable
	// +crossplane:generate:reference:type=Cluster
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterNameRef is a reference to a Cluster used to set
	// the ClusterName.
	// +immutable
	// +optional
	ClusterNameRef *xpv1.Reference `json:"clusterNameRef,omitempty"`

	// ClusterNameSelector selects references to a Cluster used
	// to set the ClusterName.
	// +optional
	ClusterNameSelector *xpv1.Selector `json:"clusterNameSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the pod execution role to use for pods
	// that match the selectors in the Fargate profile. The pod execution role allows
	// Fargate infrastructure to register with your cluster as a node, and it provides
	// read access to Amazon ECR image repositories. For more information, see Pod
	// Execution Role (https://docs.aws.amazon.com/eks/latest/userguide/pod-execution-role.html)
	// in the Amazon EKS User Guide.
	//
	// At least one of podExecutionRoleArn, podExecutionRoleArnRef or podExecutionRoleArnSelector has to be given
	// +immutable
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	PodExecutionRoleArn string `json:"podExecutionRoleArn,omitempty"`

	// PodExecutionRoleArnRef is a reference to an IAMRole used to set
	// the PodExecutionRoleArn.
	// At least one of podExecutionRoleArn, podExecutionRoleArnRef or podExecutionRoleArnSelector has to be given
	// +immutable
	// +optional
	PodExecutionRoleArnRef *xpv1.Reference `json:"podExecutionRoleArnRef,omitempty"`

	// PodExecutionRoleArnSelector selects references to IAMRole used
	// to set the PodExecutionRoleArn.
	// At least one of podExecutionRoleArn, podExecutionRoleArnRef or podExecutionRoleArnSelector has to be given
	// +optional
	PodExecutionRoleArnSelector *xpv1.Selector `json:"podExecutionRoleArnSelector,omitempty"`

	// The selectors to match for pods to use this Fargate profile. Each selector
	// must have an associated namespace. Optionally, you can also specify labels
	// for a namespace. You may specify up to five selectors in a Fargate profile.
	// +immutable
	Selectors []FargateProfileSelector `json:"selectors,omitempty"`

	// The IDs of subnets to launch your pods into. At this time, pods running on
	// Fargate are not assigned public IP addresses, so only private subnets (with
	// no direct route to an Internet Gateway) are accepted for this parameter.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetSelector
	Subnets []string `json:"subnets,omitempty"`

	// SubnetRefs are references to Subnets used to set the Subnets.
	// +immutable
	// +optional
	SubnetRefs []xpv1.Reference `json:"subnetRefs,omitempty"`

	// SubnetSelector selects references to Subnets used to set the Subnets.
	// +optional
	SubnetSelector *xpv1.Selector `json:"subnetSelector,omitempty"`

	// The metadata to apply to the Fargate profile to assist with categorization
	// and organization. Each tag consists of a key and an optional value, both
	// of which you define. Fargate profile tags do not propagate to any other resources
	// associated with the Fargate profile, such as the pods that are scheduled
	// with it.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// A FargateProfileSpec defines the desired state of an EKS FargateProfile.
type FargateProfileSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       FargateProfileParameters `json:"forProvider"`
}

// A FargateProfileStatus represents the observed state of an EKS FargateProfile.
type FargateProfileStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          FargateProfileObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A FargateProfile is a managed resource that represents an AWS Elastic Kubernetes
// Service FargateProfile.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="CLUSTER",type="string",JSONPath=".spec.forProvider.clusterName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type FargateProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FargateProfileSpec   `json:"spec"`
	Status FargateProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FargateProfileList contains a list of FargateProfile items
type FargateProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FargateProfile `json:"items"`
}
