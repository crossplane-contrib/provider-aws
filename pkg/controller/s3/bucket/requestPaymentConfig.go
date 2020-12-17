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

package bucket

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	paymentGetFailed = "cannot get request payment configuration"
	paymentPutFailed = "cannot put Bucket payment"
)

// RequestPaymentConfigurationClient is the client for API methods and reconciling the PaymentConfiguration
type RequestPaymentConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *RequestPaymentConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: awsclient.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return awsclient.Wrap(err, paymentGetFailed)
	}

	if len(external.Payer) == 0 {
		return nil
	}
	config := bucket.Spec.ForProvider.PayerConfiguration
	if config == nil {
		bucket.Spec.ForProvider.PayerConfiguration = &v1beta1.PaymentConfiguration{}
		config = bucket.Spec.ForProvider.PayerConfiguration
	}
	config.Payer = awsclient.LateInitializeString(config.Payer, awsclient.String(string(external.Payer)))
	return nil
}

// NewRequestPaymentConfigurationClient creates the client for Payment Configuration
func NewRequestPaymentConfigurationClient(client s3.BucketClient) *RequestPaymentConfigurationClient {
	return &RequestPaymentConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *RequestPaymentConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: awsclient.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, awsclient.Wrap(err, paymentGetFailed)
	}
	config := bucket.Spec.ForProvider.PayerConfiguration

	switch {
	case config == nil && len(external.Payer) == 0:
		return Updated, nil
	case config == nil && len(external.Payer) != 0:
		return NeedsUpdate, nil
	case config.Payer != string(external.Payer):
		return NeedsUpdate, nil
	default:
		return Updated, nil
	}
}

// GeneratePutBucketPaymentInput creates the input for the BucketRequestPayment request for the S3 Client
func GeneratePutBucketPaymentInput(name string, config *v1beta1.PaymentConfiguration) *awss3.PutBucketRequestPaymentInput {
	bci := &awss3.PutBucketRequestPaymentInput{
		Bucket:                      awsclient.String(name),
		RequestPaymentConfiguration: &awss3.RequestPaymentConfiguration{Payer: awss3.Payer(config.Payer)},
	}
	return bci
}

// CreateOrUpdate sends a request to have resource created on awsclient.
func (in *RequestPaymentConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.PayerConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketPaymentInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.PayerConfiguration)
	_, err := in.client.PutBucketRequestPaymentRequest(input).Send(ctx)
	return awsclient.Wrap(err, paymentPutFailed)
}

// Delete does nothing since there is no corresponding deletion call in awsclient.
func (*RequestPaymentConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}
