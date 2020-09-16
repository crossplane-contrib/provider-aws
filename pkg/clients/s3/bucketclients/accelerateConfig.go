package bucketclients

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// AccelerateConfigurationClient is the client for API methods and reconciling the AccelerateConfiguration
type AccelerateConfigurationClient struct {
	config *v1beta1.AccelerateConfiguration
	bucket *v1beta1.Bucket
	client s3.BucketClient
}

// CreateAccelerateConfigurationClient creates the client for Accelerate Configuration
func CreateAccelerateConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *AccelerateConfigurationClient {
	return &AccelerateConfigurationClient{config: bucket.Spec.Parameters.AccelerateConfiguration, bucket: bucket, client: client}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *AccelerateConfigurationClient) ExistsAndUpdated(ctx context.Context) (ResourceStatus, error) {
	conf, err := in.client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: aws.String(meta.GetExternalName(in.bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket accelerate configuration")
	}

	if conf.Status == "" && in.config == nil {
		return Updated, nil
	} else if conf.Status != "" && in.config == nil {
		return NeedsDeletion, nil
	}

	if string(conf.Status) != in.config.Status {
		return NeedsUpdate, nil
	}

	return Updated, nil
}

// GenerateAccelerateConfigurationInput creates the input for the AccelerateConfiguration request for the S3 Client
func (in *AccelerateConfigurationClient) GenerateAccelerateConfigurationInput(name string) *awss3.PutBucketAccelerateConfigurationInput {
	return &awss3.PutBucketAccelerateConfigurationInput{
		Bucket:                  aws.String(name),
		AccelerateConfiguration: &awss3.AccelerateConfiguration{Status: awss3.BucketAccelerateStatus(in.config.Status)},
	}
}

// CreateResource sends a request to have resource created on AWS
func (in *AccelerateConfigurationClient) CreateResource(ctx context.Context) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketAccelerateConfigurationRequest(in.GenerateAccelerateConfigurationInput(meta.GetExternalName(in.bucket))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket acceleration configuration")
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *AccelerateConfigurationClient) DeleteResource(ctx context.Context) error {
	_, err := in.client.PutBucketAccelerateConfigurationRequest(
		&awss3.PutBucketAccelerateConfigurationInput{
			Bucket: aws.String(meta.GetExternalName(in.bucket)),
			AccelerateConfiguration: &awss3.AccelerateConfiguration{
				Status: awss3.BucketAccelerateStatusSuspended,
			},
		},
	).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket acceleration configuration")
}
