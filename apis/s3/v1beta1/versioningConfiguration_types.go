package v1beta1

// VersioningConfiguration describes the versioning state of an Amazon S3 bucket.
type VersioningConfiguration struct {
	// MFADelete specifies whether MFA delete is enabled in the bucket versioning configuration.
	// This element is only returned if the bucket has been configured with MFA
	// delete. If the bucket has never been so configured, this element is not returned.
	// +kubebuilder:validation:Enum=Enabled;Disabled
	MFADelete *string `json:"mfaDelete"`

	// Status is the desired versioning state of the bucket.
	// +kubebuilder:validation:Enum=Enabled;Suspended
	Status string `json:"status"`
}
