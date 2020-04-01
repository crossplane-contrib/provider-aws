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
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// S3BucketParameters define the desired state of an AWS S3 Bucket.
type S3BucketParameters struct {
	// NameFormat specifies the name of the external S3Bucket instance. The
	// first instance of the string '%s' will be replaced with the Kubernetes
	// UID of this S3Bucket. Omit this field to use the UID alone as the name.
	// +optional
	NameFormat string `json:"nameFormat,omitempty"`

	// Region of the bucket.
	Region string `json:"region"`

	// CannedACL applies a standard AWS built-in ACL for common bucket use
	// cases.
	// +kubebuilder:validation:Enum=private;public-read;public-read-write;authenticated-read;log-delivery-write;aws-exec-read
	// +optional
	CannedACL *s3.BucketCannedACL `json:"cannedACL,omitempty"`

	// Versioning enables versioning of objects stored in this bucket.
	// +optional
	Versioning bool `json:"versioning,omitempty"`

	// IAMUsername is the name of an IAM user that is automatically created and
	// granted access to this bucket by Crossplane at bucket creation time.
	IAMUsername string `json:"iamUsername,omitempty"`

	// LocalPermission is the permissions granted on the bucket for the provider
	// specific bucket service account that is available in a secret after
	// provisioning.
	// +kubebuilder:validation:Enum=Read;Write;ReadWrite
	LocalPermission *storagev1alpha1.LocalPermissionType `json:"localPermission"`
}

// S3BucketSpec defines the desired state of S3Bucket
type S3BucketSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	S3BucketParameters           `json:",inline"`
}

// S3BucketStatus defines the observed state of S3Bucket
type S3BucketStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// ProviderID is the AWS identifier for this bucket.
	ProviderID string `json:"providerID,omitempty"`

	// LastUserPolicyVersion is the most recent version of the policy associated
	// with this bucket's IAMUser.
	LastUserPolicyVersion int `json:"lastUserPolicyVersion,omitempty"`

	// LastLocalPermission is the most recent local permission that was set for
	// this bucket.
	LastLocalPermission storagev1alpha1.LocalPermissionType `json:"lastLocalPermission,omitempty"`
}

// +kubebuilder:object:root=true

// An S3Bucket is a managed resource that represents an AWS S3 Bucket.
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="PREDEFINED-ACL",type="string",JSONPath=".spec.cannedACL"
// +kubebuilder:printcolumn:name="LOCAL-PERMISSION",type="string",JSONPath=".spec.localPermission"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type S3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketSpec   `json:"spec"`
	Status S3BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// S3BucketList contains a list of S3Buckets
type S3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Bucket `json:"items"`
}

// An S3BucketClassSpecTemplate is a template for the spec of a dynamically
// provisioned S3Bucket.
type S3BucketClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	S3BucketParameters                `json:",inline"`
}

// +kubebuilder:object:root=true

// An S3BucketClass is a resource class. It defines the desired spec of resource
// claims that use it to dynamically provision a managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type S3BucketClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// S3Bucket.
	SpecTemplate S3BucketClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// S3BucketClassList contains a list of cloud memorystore resource classes.
type S3BucketClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3BucketClass `json:"items"`
}

// GetBucketName based on the NameFormat spec value,
// If name format is not provided, bucket name defaults to UID
// If name format provided with '%s' value, bucket name will result in formatted string + UID,
//   NOTE: only single %s substitution is supported
// If name format does not contain '%s' substitution, i.e. a constant string, the
// constant string value is returned back
//
// Examples:
//   For all examples assume "UID" = "test-uid"
//   1. NameFormat = "", BucketName = "test-uid"
//   2. NameFormat = "%s", BucketName = "test-uid"
//   3. NameFormat = "foo", BucketName = "foo"
//   4. NameFormat = "foo-%s", BucketName = "foo-test-uid"
//   5. NameFormat = "foo-%s-bar-%s", BucketName = "foo-test-uid-bar-%!s(MISSING)"
func (b *S3Bucket) GetBucketName() string {
	if b.Spec.NameFormat == "" {
		return string(b.GetUID())
	}
	if strings.Contains(b.Spec.NameFormat, "%s") {
		return fmt.Sprintf(b.Spec.NameFormat, string(b.GetUID()))
	}
	return b.Spec.NameFormat
}

// SetUserPolicyVersion specifies this bucket's policy version.
func (b *S3Bucket) SetUserPolicyVersion(policyVersion string) error {
	policyInt, err := strconv.Atoi(policyVersion[1:])
	if err != nil {
		return err
	}
	b.Status.LastUserPolicyVersion = policyInt
	b.Status.LastLocalPermission = *b.Spec.LocalPermission

	return nil
}

// HasPolicyChanged returns true if the bucket's policy is older than the
// supplied version.
func (b *S3Bucket) HasPolicyChanged(policyVersion string) (bool, error) {
	if *b.Spec.LocalPermission != b.Status.LastLocalPermission {
		return true, nil
	}
	policyInt, err := strconv.Atoi(policyVersion[1:])
	if err != nil {
		return false, err
	}
	if b.Status.LastUserPolicyVersion != policyInt && b.Status.LastUserPolicyVersion < policyInt {
		return true, nil
	}

	return false, nil
}
