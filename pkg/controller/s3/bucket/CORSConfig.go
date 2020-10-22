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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	corsGetFailed    = "cannot get Bucket CORS configuration"
	corsPutFailed    = "cannot put Bucket cors"
	corsDeleteFailed = "cannot delete Bucket CORS configuration"
)

// CORSConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type CORSConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize does nothing because CORSConfiguration might have been deleted
// by the user.
func (*CORSConfigurationClient) LateInitialize(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// NewCORSConfigurationClient creates the client for CORS Configuration
func NewCORSConfigurationClient(client s3.BucketClient) *CORSConfigurationClient {
	return &CORSConfigurationClient{client: client}
}

// CompareCORS compares the external and internal representations for the list of CORSRules
func CompareCORS(local *v1beta1.CORSConfiguration, external []awss3.CORSRule) ResourceStatus { // nolint:gocyclo
	switch {
	case local == nil && external != nil:
		return NeedsDeletion
	case local == nil && len(external) == 0:
		return Updated
	case local == nil:
		return NeedsUpdate
	case external == nil:
		return NeedsUpdate
	case len(local.CORSRules) != len(external):
		return NeedsUpdate
	}

	for i := range local.CORSRules {
		outputRule := external[i]
		if !(cmp.Equal(local.CORSRules[i].AllowedHeaders, outputRule.AllowedHeaders) &&
			cmp.Equal(local.CORSRules[i].AllowedMethods, outputRule.AllowedMethods) &&
			cmp.Equal(local.CORSRules[i].AllowedOrigins, outputRule.AllowedOrigins) &&
			cmp.Equal(local.CORSRules[i].ExposeHeaders, outputRule.ExposeHeaders) &&
			cmp.Equal(local.CORSRules[i].MaxAgeSeconds, outputRule.MaxAgeSeconds)) {
			return NeedsUpdate
		}
	}

	return Updated
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *CORSConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketCorsRequest(&awss3.GetBucketCorsInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	config := bucket.Spec.ForProvider.CORSConfiguration
	if err != nil && s3.CORSConfigurationNotFound(err) && config == nil {
		return Updated, nil
	} else if err != nil {
		return NeedsUpdate, errors.Wrap(err, corsGetFailed)
	}

	return CompareCORS(config, external.CORSRules), nil
}

// GeneratePutBucketCorsInput creates the input for the PutBucketCors request for the S3 Client
func GeneratePutBucketCorsInput(name string, config *v1beta1.CORSConfiguration) *awss3.PutBucketCorsInput {
	bci := &awss3.PutBucketCorsInput{
		Bucket:            aws.String(name),
		CORSConfiguration: &awss3.CORSConfiguration{CORSRules: make([]awss3.CORSRule, 0)},
	}
	for _, cors := range config.CORSRules {
		bci.CORSConfiguration.CORSRules = append(bci.CORSConfiguration.CORSRules, awss3.CORSRule{
			AllowedHeaders: cors.AllowedHeaders,
			AllowedMethods: cors.AllowedMethods,
			AllowedOrigins: cors.AllowedOrigins,
			ExposeHeaders:  cors.ExposeHeaders,
			MaxAgeSeconds:  cors.MaxAgeSeconds,
		})
	}
	return bci
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *CORSConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.CORSConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketCorsInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.CORSConfiguration)
	_, err := in.client.PutBucketCorsRequest(input).Send(ctx)
	return errors.Wrap(err, corsPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *CORSConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketCorsRequest(
		&awss3.DeleteBucketCorsInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, corsDeleteFailed)
}
