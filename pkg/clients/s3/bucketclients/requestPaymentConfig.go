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

// RequestPaymentConfigurationClient is the client for API methods and reconciling the PaymentConfiguration
type RequestPaymentConfigurationClient struct {
	config *v1beta1.PaymentConfiguration
	bucket *v1beta1.Bucket
	client s3.BucketClient
}

// CreateRequestPaymentConfigurationClient creates the client for Payment Configuration
func CreateRequestPaymentConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *RequestPaymentConfigurationClient {
	return &RequestPaymentConfigurationClient{config: bucket.Spec.Parameters.PayerConfiguration, bucket: bucket, client: client}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *RequestPaymentConfigurationClient) ExistsAndUpdated(ctx context.Context) (ResourceStatus, error) {
	conf, err := in.client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: aws.String(meta.GetExternalName(in.bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get request payment configuration")
	}

	if conf.GetBucketRequestPaymentOutput.Payer == "" && in.config == nil {
		return Updated, nil
	} else if conf.GetBucketRequestPaymentOutput.Payer != "" && in.config == nil {
		return NeedsDeletion, nil
	}

	if in.config.Payer != string(conf.Payer) {
		return NeedsUpdate, nil
	}

	return Updated, nil
}

// GeneratePutBucketPaymentInput creates the input for the BucketRequestPayment request for the S3 Client
func (in *RequestPaymentConfigurationClient) GeneratePutBucketPaymentInput(name string) *awss3.PutBucketRequestPaymentInput {
	bci := &awss3.PutBucketRequestPaymentInput{
		Bucket:                      aws.String(name),
		RequestPaymentConfiguration: &awss3.RequestPaymentConfiguration{Payer: awss3.Payer(in.config.Payer)},
	}

	return bci
}

// CreateResource sends a request to have resource created on AWS.
func (in *RequestPaymentConfigurationClient) CreateResource(ctx context.Context) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketRequestPaymentRequest(in.GeneratePutBucketPaymentInput(meta.GetExternalName(in.bucket))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket payment")
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *RequestPaymentConfigurationClient) DeleteResource(ctx context.Context) error {
	input := &awss3.PutBucketRequestPaymentInput{
		Bucket:                      aws.String(meta.GetExternalName(in.bucket)),
		RequestPaymentConfiguration: &awss3.RequestPaymentConfiguration{Payer: awss3.PayerBucketOwner},
	}
	_, err := in.client.PutBucketRequestPaymentRequest(input).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket payment configuration")
}
