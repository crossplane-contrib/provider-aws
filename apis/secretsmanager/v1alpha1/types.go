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

	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SecretParameters define the desired state of an AWS Elastic Kubernetes
// Service Secret.
type SecretParameters struct {
	// Region is the region you'd like  the Secret to be created in.
	Region string `json:"region"`

	// (Optional) Specifies a user-provided description of the secret.
	// +optional
	Description *string `json:"description,omitempty"`

	// (Optional) Specifies the ARN, Key ID, or alias of the AWS KMS customer master
	// key (CMK) to be used to encrypt the SecretString or SecretBinary values in the
	// versions stored in this secret. You can specify any of the supported ways to
	// identify a AWS KMS key ID. If you need to reference a CMK in a different
	// account, you can use only the key ARN or the alias ARN. If you don't specify
	// this value, then Secrets Manager defaults to using the AWS account's default CMK
	// (the one named aws/secretsmanager). If a AWS KMS CMK with that name doesn't yet
	// exist, then Secrets Manager creates it for you automatically the first time it
	// needs to encrypt a version's SecretString or SecretBinary fields. You can use
	// the account default CMK to encrypt and decrypt only if you call this operation
	// using credentials from the same account that owns the secret. If the secret
	// resides in a different account, then you must create a custom CMK and specify
	// the ARN in this field.
	// +optional
	KmsKeyID *string `json:"kmsKeyId,omitempty"`

	// KmsKeyIDRef is a reference to a crossplane managed KMS Key
	// to set the KmsKeyId.
	// +optional
	KmsKeyRef *runtimev1.Reference `json:"kmsKeyIdRef,omitempty"`

	// KmsKeyIDSelector selects references to a crossplane managed KMS Key
	// to set the KmsKeyId.
	// +optional
	KmsKeySelector *runtimev1.Selector `json:"kmsKeyIdSelector,omitempty"`

	// (Optional) Specifies text data that you want to encrypt and store in this new
	// version of the secret. Either SecretString or SecretBinary must have a value,
	// but not both. They cannot both be empty. If you create a secret by using the
	// Secrets Manager console then Secrets Manager puts the protected secret text in
	// only the SecretString parameter.

	// For storing multiple values, we recommend that you
	// use a JSON text string argument and specify key/value pairs. For information on
	// how to format a JSON parameter for the various command line tool environments,
	// see Using JSON for Parameters
	// (https://docs.aws.amazon.com/cli/latest/userguide/cli-using-param.html#cli-using-param-json)
	// in the AWS CLI User Guide. For example:
	// {"username":"bob","password":"abc123xyz456"} If your command-line tool or SDK
	// requires quotation marks around the parameter, you should use single quotes to
	// avoid confusion with the double quotes required in the JSON text.
	SecretRef *SecretSelector `json:"secretRef"`

	// (Optional) Specifies that the secret is to be deleted without any recovery
	// window. You can't use both this parameter and the RecoveryWindowInDays parameter
	// in the same API call. An asynchronous background process performs the actual
	// deletion, so there can be a short delay before the operation completes. If you
	// write code to delete and then immediately recreate a secret with the same name,
	// ensure that your code includes appropriate back off and retry logic. Use this
	// parameter with caution. This parameter causes the operation to skip the normal
	// waiting period before the permanent deletion that AWS would normally impose with
	// the RecoveryWindowInDays parameter. If you delete a secret with the
	// ForceDeleteWithouRecovery parameter, then you have no opportunity to recover the
	// secret. It is permanently lost.
	// +optional
	ForceDeleteWithoutRecovery *bool `json:"forceDeleteWithoutRecovery,omitempty"`

	// (Optional) Specifies the number of days that Secrets Manager waits before it can
	// delete the secret. You can't use both this parameter and the
	// ForceDeleteWithoutRecovery parameter in the same API call. This value can range
	// from 7 to 30 days. The default value is 30.
	// +optional
	// +kubebuilder:validation:Enum=7;8;9;10;11;12;13;14;15;16;17;18;19;20;21;22;23;24;25;26;27;28;29;30
	RecoveryWindowInDays *int64 `json:"recoveryWindowInDays,omitempty"`

	// (Optional) Specifies a list of user-defined tags that are attached to the
	// secret. Each tag is a "Key" and "Value" pair of strings. This operation only
	// appends tags to the existing list of tags. To remove tags, you must use UntagResource.
	//
	//    * Secrets Manager tag key names are case sensitive. A tag with the key
	//    "ABC" is a different tag from one with key "abc".
	//
	//    * If you check tags in IAM policy Condition elements as part of your security
	//    strategy, then adding or removing a tag can change permissions. If the
	//    successful completion of this operation would result in you losing your
	//    permissions for this secret, then this operation is blocked and returns
	//    an Access Denied error.
	//
	// This parameter requires a JSON text string argument. For information on how
	// to format a JSON parameter for the various command line tool environments,
	// see Using JSON for Parameters (https://docs.aws.amazon.com/cli/latest/userguide/cli-using-param.html#cli-using-param-json)
	// in the AWS CLI User Guide. For example:
	//
	// [{"Key":"CostCenter","Value":"12345"},{"Key":"environment","Value":"production"}]
	//
	// If your command-line tool or SDK requires quotation marks around the parameter,
	// you should use single quotes to avoid confusion with the double quotes required
	// in the JSON text.
	//
	// The following basic restrictions apply to tags:
	//
	//    * Maximum number of tags per secret—50
	//
	//    * Maximum key length—127 Unicode characters in UTF-8
	//
	//    * Maximum value length—255 Unicode characters in UTF-8
	//
	//    * Tag keys and values are case sensitive.
	//
	//    * Do not use the aws: prefix in your tag names or values because AWS reserves
	//    it for AWS use. You can't edit or delete tag names or values with this
	//    prefix. Tags with this prefix do not count against your tags per secret
	//    limit.
	//
	//    * If you use your tagging schema across multiple services and resources,
	//    remember other services might have restrictions on allowed characters.
	//    Generally allowed characters: letters, spaces, and numbers representable
	//    in UTF-8, plus the following special characters: + - = . _ : / @.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// A SecretSelector is a reference to a secret in an arbitrary namespace
// as well as an optional key to specify a certain secret value in the secret's data.
type SecretSelector struct {

	// The name of the secret and its namespace.
	*runtimev1.SecretReference `json:",inline"`

	// The key to select.
	// +optional
	Key string `json:"key,omitempty"`
}

// Tag is a structure that contains information about a tag.
type Tag struct {

	// The key identifier, or name, of the tag.
	Key string `json:"key"`

	// The string value associated with the key of the tag.
	Value string `json:"value"`
}

// SecretObservation is the observed state of a Secret.
type SecretObservation struct {

	// The date and time that this version of the secret was created.
	CreatedDate *metav1.Time `json:"createdDate,omitempty"`

	// This value exists if the secret is scheduled for deletion. Some time after
	// the specified date and time, Secrets Manager deletes the secret and all of
	// its versions.
	//
	// If a secret is scheduled for deletion, then its details, including the encrypted
	// secret information, is not accessible. To cancel a scheduled deletion and
	// restore access, use RestoreSecret.
	DeletedDate *metav1.Time `json:"deletedDate,omitempty"`

	// The date and time after which this secret can be deleted by Secrets Manager
	// and can no longer be restored. This value is the date and time of the delete
	// request plus the number of days specified in RecoveryWindowInDays.
	// (SecretObservation.DeletionDate = SecretObservation.DeletedDate + SecretParameters.RecoveryWindowInDays)
	DeletionDate *metav1.Time `json:"deletionDate,omitempty"`
}

// A SecretSpec defines the desired state of an EKS Secret.
type SecretSpec struct {
	runtimev1.ResourceSpec `json:",inline"`
	ForProvider            SecretParameters `json:"forProvider"`
}

// A SecretStatus represents the observed state of an EKS Secret.
type SecretStatus struct {
	runtimev1.ResourceStatus `json:",inline"`
	AtProvider               SecretObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Secret is a managed resource that represents an AWS Secrets Manager Service Secret
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretSpec   `json:"spec"`
	Status SecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretList contains a list of Secret items
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}
