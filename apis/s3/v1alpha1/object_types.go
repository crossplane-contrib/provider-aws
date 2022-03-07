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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectParameters defines the desired state of Object
type ObjectParameters struct {
	// Region is where the Bucket referenced by this BucketPolicy resides.
	// +immutable
	Region string `json:"region"`

	// Object key for which the PUT action was initiated.
	//
	// This member is required.
	// +immutable
	Key *string `json:"key"`

	// Object data.
	// +optional
	Body *string `json:"body"`

	// The canned ACL to apply to the object. For more information, see Canned ACL
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL).
	// This action is not supported by Amazon S3 on Outposts.
	// +optional
	ACL *string `json:"acl"`

	// The date and time at which the object is no longer cacheable. For more
	// information, see http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.21
	// (http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.21).
	// +optional
	Expires *metav1.Time `json:"expires"`

	// Gives the grantee READ, READ_ACP, and WRITE_ACP permissions on the object. This
	// action is not supported by Amazon S3 on Outposts.
	// +optional
	GrantFullControl *string `json:"grant_full_control"`

	// Allows grantee to read the object data and its metadata. This action is not
	// supported by Amazon S3 on Outposts.
	// +optional
	GrantRead *string `json:"grant_read"`

	// Allows grantee to read the object ACL. This action is not supported by Amazon S3
	// on Outposts.
	// +optional
	GrantReadACP *string `json:"grant_read_acp"`

	// Allows grantee to write the ACL for the applicable object. This action is not
	// supported by Amazon S3 on Outposts.
	// +optional
	GrantWriteACP *string `json:"grant_write_acp"`

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

// ObjectSpec defines the desired state of Object
type ObjectSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ObjectParameters `json:"forProvider"`
}

// ObjectStatus defines the observed state of Object.
type ObjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// Object is the Schema for the Object API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="BUCKET",type="string",JSONPath=".spec.forProvider.key.bucketName"
// +kubebuilder:printcolumn:name="KEY",type="string",JSONPath=".spec.forProvider.key"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Object struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ObjectSpec   `json:"spec"`
	Status            ObjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ObjectList contains a list of Objects
type ObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Object `json:"items"`
}
