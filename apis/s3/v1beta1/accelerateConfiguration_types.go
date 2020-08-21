package v1beta1

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// AccelerateConfiguration configures the transfer acceleration state for an
// Amazon S3 bucket. For more information, see Amazon S3 Transfer Acceleration
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html)
// in the Amazon Simple Storage Service Developer Guide.
type AccelerateConfiguration struct {
	// Status specifies the transfer acceleration status of the bucket.
	// +kubebuilder:validation:Enum=Enabled;Suspended
	Status string `json:"status"`
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *AccelerateConfiguration) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (managed.ExternalObservation, error) {
	conf, err := client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get bucket encryption")
	}

	if string(conf.Status) != in.Status {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
	}

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

// GenerateAccelerateConfigurationInput creates the input for the AccelerateConfiguration request for the S3 Client
func (in *AccelerateConfiguration) GenerateAccelerateConfigurationInput(name string) *awss3.PutBucketAccelerateConfigurationInput {
	return &awss3.PutBucketAccelerateConfigurationInput{
		Bucket:                  aws.String(name),
		AccelerateConfiguration: &awss3.AccelerateConfiguration{Status: awss3.BucketAccelerateStatus(in.Status)},
	}
}
