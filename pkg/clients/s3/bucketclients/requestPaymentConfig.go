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
}

// CreateRequestPaymentConfigurationClient creates the client for Payment Configuration
func CreateRequestPaymentConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &RequestPaymentConfigurationClient{config: parameters.PayerConfiguration}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *RequestPaymentConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: bucketName}).Send(ctx)
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
func (in *RequestPaymentConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketRequestPaymentRequest(in.GeneratePutBucketPaymentInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket logging")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *RequestPaymentConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	input := &awss3.PutBucketRequestPaymentInput{
		Bucket:                      aws.String(meta.GetExternalName(cr)),
		RequestPaymentConfiguration: &awss3.RequestPaymentConfiguration{Payer: awss3.PayerBucketOwner},
	}
	if _, err := client.PutBucketRequestPaymentRequest(input).Send(ctx); err != nil {
		return errors.Wrap(err, "cannot delete bucket payment configuration")
	}
	return nil
}
