package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomKeyParameters are custom parameters for Key.
type CustomKeyParameters struct {
	// Specifies whether the CMK is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Specifies how many days the Key is retained when scheduled for deletion. Defaults to 30 days.
	PendingWindowInDays *int64 `json:"pendingWindowInDays,omitempty"`
}

// CustomAliasParameters are custom parameters for Alias.
type CustomAliasParameters struct {
	// Associates the alias with the specified customer managed CMK (https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#customer-cmk). The CMK must be in the same AWS Region. \n A valid CMK ID is required. If you supply a null or empty string value, this operation returns an error. \n For help finding the key ID and ARN, see Finding the Key ID and ARN (https://docs.aws.amazon.com/kms/latest/developerguide/viewing-keys.html#find-cmk-id-arn) in the AWS Key Management Service Developer Guide. \n Specify the key ID or the Amazon Resource Name (ARN) of the CMK. \n For example: \n    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab \n    * Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab \n To get the key ID and key ARN for a CMK, use ListKeys or DescribeKey.
	// +optional
	TargetKeyID *string `json:"targetKeyID,omitempty"`

	// TargetKeyIDRef is a reference to a KMS Key used to set TargetKeyID.
	// +optional
	TargetKeyIDRef *xpv1.Reference `json:"targetKeyIDRef,omitempty"`

	// TargetKeyIDSelector selects a reference to a KMS Key used to set TargetKeyID.
	// +optional
	TargetKeyIDSelector *xpv1.Selector `json:"targetKeyIDSelector,omitempty"`
}
