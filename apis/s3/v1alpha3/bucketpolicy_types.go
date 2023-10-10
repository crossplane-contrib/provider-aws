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

package v1alpha3

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
)

// BucketPolicyParameters define the desired state of an AWS BucketPolicy.
type BucketPolicyParameters struct {
	// Region is where the Bucket referenced by this BucketPolicy resides.
	// +immutable
	Region string `json:"region"`

	// RawPolicy is a stringified version of the S3 Bucket Policy.
	// either policy or rawPolicy must be specified in the policy
	// +optional
	RawPolicy *string `json:"rawPolicy,omitempty"`

	// Policy is a well defined type which can be parsed into an JSON S3 Bucket Policy
	// either policy or rawPolicy must be specified in the policy
	// +optional
	Policy *common.BucketPolicyBody `json:"policy,omitempty"`

	// BucketName presents the name of the bucket.
	// +optional
	// +immutable
	BucketName *string `json:"bucketName,omitempty"`

	// BucketNameRef references to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameRef *xpv1.Reference `json:"bucketNameRef,omitempty"`

	// BucketNameSelector selects a reference to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameSelector *xpv1.Selector `json:"bucketNameSelector,omitempty"`
}

// An BucketPolicySpec defines the desired state of an
// BucketPolicy.
type BucketPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	Parameters        BucketPolicyParameters `json:"forProvider"`
}

// An BucketPolicyStatus represents the observed state of an
// BucketPolicy.
type BucketPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An BucketPolicy is a managed resource that represents an AWS Bucket
// policy.
// +kubebuilder:printcolumn:name="BUCKETNAME",type="string",JSONPath=".spec.forProvider.bucketName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:deprecatedversion:warning="BucketPolicy has been deprecated. Use spec.forProvider.policy in Bucket instead."
type BucketPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPolicySpec   `json:"spec"`
	Status BucketPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicyList contains a list of BucketPolicies
type BucketPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPolicy `json:"items"`
}
