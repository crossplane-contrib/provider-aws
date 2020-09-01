package v1beta1

// AccelerateConfiguration configures the transfer acceleration state for an
// Amazon S3 bucket. For more information, see Amazon S3 Transfer Acceleration
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html)
// in the Amazon Simple Storage Service Developer Guide.
type AccelerateConfiguration struct {
	// Status specifies the transfer acceleration status of the bucket.
	// +kubebuilder:validation:Enum=Enabled;Suspended
	Status string `json:"status"`
}
