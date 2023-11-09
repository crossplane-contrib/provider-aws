package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomKeyParameters are custom parameters for Key.
type CustomKeyParameters struct {
	// Specifies whether the CMK is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Specifies how many days the Key is retained when scheduled for deletion. Defaults to 30 days.
	PendingWindowInDays *int64 `json:"pendingWindowInDays,omitempty"`

	// Specifies if key rotation is enabled for the corresponding key
	EnableKeyRotation *bool `json:"enableKeyRotation,omitempty"`
}

// CustomGrantParameters are custom parameters for Grant.
type CustomGrantParameters struct {
	// Identifies the KMS key for the grant. The grant gives principals permission
	// to use this KMS key.
	//
	// Specify the key ID or key ARN of the KMS key. To specify a KMS key in a different
	// Amazon Web Services account, you must use the key ARN.
	//
	// For example:
	//
	//    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	//
	//    * Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	//
	// To get the key ID and key ARN for a KMS key, use ListKeys or DescribeKey.
	//
	// KeyID or one of the referencers is a required parameter.
	//
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	KeyID *string `json:"keyId,omitempty"`

	// KeyIDRef is a reference to a KeyID.
	// +optional
	KeyIDRef *xpv1.Reference `json:"keyIdRef,omitempty"`

	// KeyIDSelector selects references to a KeyID.
	// +optional
	KeyIDSelector *xpv1.Selector `json:"keyIdSelector,omitempty"`
}
