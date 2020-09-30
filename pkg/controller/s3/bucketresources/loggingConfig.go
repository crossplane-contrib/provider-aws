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

package bucketresources

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// LoggingConfigurationClient is the client for API methods and reconciling the LoggingConfiguration
type LoggingConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *LoggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, loggingGetFailed)
	}
	config := bucket.Spec.ForProvider.LoggingConfiguration
	if external.LoggingEnabled == nil {
		// There is no value send by AWS to initialize
		return nil
	}
	if config == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.LoggingConfiguration = &v1beta1.LoggingConfiguration{}
		config = bucket.Spec.ForProvider.LoggingConfiguration
	}
	// Late initialize the target Bucket and target prefix
	config.TargetBucket = aws.LateInitializeStringPtr(config.TargetBucket, external.LoggingEnabled.TargetBucket)
	config.TargetPrefix = aws.LateInitializeStringPtr(config.TargetPrefix, external.LoggingEnabled.TargetPrefix)
	// If the there is an external target grant list, and the local one does not exist
	// we create the target grant list
	if external.LoggingEnabled.TargetGrants != nil && len(config.TargetGrants) == 0 {
		config.TargetGrants = make([]v1beta1.TargetGrant, len(external.LoggingEnabled.TargetGrants))
		for i, v := range external.LoggingEnabled.TargetGrants {
			config.TargetGrants[i] = v1beta1.TargetGrant{
				Grantee: v1beta1.TargetGrantee{
					DisplayName:  v.Grantee.DisplayName,
					EmailAddress: v.Grantee.EmailAddress,
					ID:           v.Grantee.ID,
					Type:         string(v.Grantee.Type),
					URI:          v.Grantee.URI,
				},
				Permission: string(v.Permission),
			}
		}
	}
	return nil
}

// NewLoggingConfigurationClient creates the client for Logging Configuration
func NewLoggingConfigurationClient(client s3.BucketClient) *LoggingConfigurationClient {
	return &LoggingConfigurationClient{client: client}
}

// GenerateAWSLogging creates an S3 logging enabled struct from the local logging configuration
func GenerateAWSLogging(local *v1beta1.LoggingConfiguration) *awss3.LoggingEnabled {
	output := awss3.LoggingEnabled{
		TargetBucket: local.TargetBucket,
		TargetPrefix: local.TargetPrefix,
	}
	if local.TargetGrants != nil {
		output.TargetGrants = make([]awss3.TargetGrant, len(local.TargetGrants))
	}
	for i := range local.TargetGrants {
		target := awss3.TargetGrant{
			Grantee: &awss3.Grantee{
				DisplayName:  local.TargetGrants[i].Grantee.DisplayName,
				EmailAddress: local.TargetGrants[i].Grantee.EmailAddress,
				ID:           local.TargetGrants[i].Grantee.ID,
				Type:         awss3.Type(local.TargetGrants[i].Grantee.Type),
				URI:          local.TargetGrants[i].Grantee.URI,
			},
			Permission: awss3.BucketLogsPermission(local.TargetGrants[i].Permission),
		}

		output.TargetGrants[i] = target
	}
	return &output
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LoggingConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, loggingGetFailed)
	}
	config := bucket.Spec.ForProvider.LoggingConfiguration

	switch {
	case external.LoggingEnabled == nil && config == nil:
		return Updated, nil
	case external.LoggingEnabled != nil && config == nil:
		return NeedsDeletion, nil
	case cmp.Equal(GenerateAWSLogging(config), external.LoggingEnabled):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// GeneratePutBucketLoggingInput creates the input for the PutBucketLogging request for the S3 Client
func GeneratePutBucketLoggingInput(name string, config *v1beta1.LoggingConfiguration) *awss3.PutBucketLoggingInput {
	bci := &awss3.PutBucketLoggingInput{
		Bucket: aws.String(name),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{LoggingEnabled: &awss3.LoggingEnabled{
			TargetBucket: config.TargetBucket,
			TargetGrants: make([]awss3.TargetGrant, 0),
			TargetPrefix: config.TargetPrefix,
		}},
	}
	for _, grant := range config.TargetGrants {
		bci.BucketLoggingStatus.LoggingEnabled.TargetGrants = append(bci.BucketLoggingStatus.LoggingEnabled.TargetGrants, awss3.TargetGrant{
			Grantee: &awss3.Grantee{
				DisplayName:  grant.Grantee.DisplayName,
				EmailAddress: grant.Grantee.EmailAddress,
				ID:           grant.Grantee.ID,
				Type:         awss3.Type(grant.Grantee.Type),
				URI:          grant.Grantee.URI,
			},
			Permission: awss3.BucketLogsPermission(grant.Permission),
		})
	}
	return bci
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LoggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	config := bucket.Spec.ForProvider.LoggingConfiguration
	if config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketLoggingRequest(GeneratePutBucketLoggingInput(meta.GetExternalName(bucket), config)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, loggingPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *LoggingConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := &awss3.PutBucketLoggingInput{
		Bucket:              aws.String(meta.GetExternalName(bucket)),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{}, //  Empty BucketLoggingStatus disables logging
	}
	_, err := in.client.PutBucketLoggingRequest(input).Send(ctx)
	return errors.Wrap(err, loggingDeleteFailed)
}
