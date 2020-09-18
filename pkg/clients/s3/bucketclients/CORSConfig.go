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

package bucketclients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// CORSConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type CORSConfigurationClient struct {
	config *v1beta1.CORSConfiguration
	client s3.BucketClient
}

// NewCORSConfigurationClient creates the client for CORS Configuration
func NewCORSConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *CORSConfigurationClient {
	return &CORSConfigurationClient{config: bucket.Spec.Parameters.CORSConfiguration, client: client}
}

func (in *CORSConfigurationClient) corsNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchCORSConfiguration" && in.config == nil {
		return true
	}
	return false
}

func sameRule(local *v1beta1.CORSRule, external *awss3.CORSRule) bool {
	if !cmp.Equal(local.AllowedHeaders, external.AllowedHeaders) {
		return false
	}
	if !cmp.Equal(local.AllowedMethods, external.AllowedMethods) {
		return false
	}
	if !cmp.Equal(local.AllowedOrigins, external.AllowedOrigins) {
		return false
	}
	if !cmp.Equal(local.ExposeHeaders, external.ExposeHeaders) {
		return false
	}
	if !cmp.Equal(local.MaxAgeSeconds, external.MaxAgeSeconds) {
		return false
	}
	return true
}

func compareCORS(local *v1beta1.CORSConfiguration, external []awss3.CORSRule) ResourceStatus {
	if local == nil && external != nil {
		return NeedsDeletion
	} else if local == nil && len(external) == 0 {
		return Updated
	}

	if len(local.CORSRules) != len(external) {
		return NeedsUpdate
	}

	for i := range local.CORSRules {
		outputRule := external[i]
		if !sameRule(&local.CORSRules[i], &outputRule) {
			return NeedsUpdate
		}
	}
	return Updated
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *CORSConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketCorsRequest(&awss3.GetBucketCorsInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil && in.corsNotFound(err) {
		return Updated, nil
	} else if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket CORS configuration")
	}

	return compareCORS(in.config, conf.CORSRules), nil
}

// GeneratePutBucketCorsInput creates the input for the PutBucketCors request for the S3 Client
func (in *CORSConfigurationClient) GeneratePutBucketCorsInput(name string) *awss3.PutBucketCorsInput {
	bci := &awss3.PutBucketCorsInput{
		Bucket:            aws.String(name),
		CORSConfiguration: &awss3.CORSConfiguration{},
	}
	for _, cors := range in.config.CORSRules {
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

// Create sends a request to have resource created on AWS
func (in *CORSConfigurationClient) Create(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketCorsRequest(in.GeneratePutBucketCorsInput(meta.GetExternalName(bucket))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket cors")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *CORSConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketCorsRequest(
		&awss3.DeleteBucketCorsInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket CORS configuration")
}
