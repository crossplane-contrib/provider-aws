package v1beta1

import (
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// VersioningConfiguration describes the versioning state of an Amazon S3 bucket.
type VersioningConfiguration struct {
	// MFADelete specifies whether MFA delete is enabled in the bucket versioning configuration.
	// This element is only returned if the bucket has been configured with MFA
	// delete. If the bucket has never been so configured, this element is not returned.
	// +kubebuilder:validation:Enum=Enabled;Disabled
	MFADelete *string `json:"mfaDelete"`

	// Status is the desired versioning state of the bucket.
	// +kubebuilder:validation:Enum=Enabled;Suspended
	Status *string `json:"status"`
}

// GeneratePutBucketVersioningInput creates the input for the PutBucketVersioning request for the S3 Client
func (conf *VersioningConfiguration) GeneratePutBucketVersioningInput(name string) *awss3.PutBucketVersioningInput {
	return &awss3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &awss3.VersioningConfiguration{
			MFADelete: awss3.MFADelete(aws.StringValue(conf.MFADelete)),
			Status:    awss3.BucketVersioningStatus(aws.StringValue(conf.Status)),
		},
	}
}
