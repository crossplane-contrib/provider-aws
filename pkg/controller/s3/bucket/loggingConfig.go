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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	loggingGetFailed = "cannot get Bucket logging configuration"
	loggingPutFailed = "cannot put Bucket logging configuration"
)

// LoggingConfigurationClient is the client for API methods and reconciling the LoggingConfiguration
type LoggingConfigurationClient struct {
	client s3.BucketClient
}

// NewLoggingConfigurationClient creates the client for Logging Configuration
func NewLoggingConfigurationClient(client s3.BucketClient) *LoggingConfigurationClient {
	return &LoggingConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LoggingConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketLogging(ctx, &awss3.GetBucketLoggingInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return NeedsUpdate, errorutils.Wrap(err, loggingGetFailed)
	}
	if !cmp.Equal(GenerateAWSLogging(bucket.Spec.ForProvider.LoggingConfiguration), external.LoggingEnabled,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}), cmpopts.IgnoreTypes(document.NoSerde{})) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LoggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.LoggingConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketLoggingInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.LoggingConfiguration)
	_, err := in.client.PutBucketLogging(ctx, input)
	return errorutils.Wrap(err, loggingPutFailed)
}

// Delete does nothing because there is no deletion call for logging config.
func (*LoggingConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *LoggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketLogging(ctx, &awss3.GetBucketLoggingInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(err, loggingGetFailed)
	}

	if external == nil || external.LoggingEnabled == nil {
		// There is no value send by AWS to initialize
		return nil
	}

	if bucket.Spec.ForProvider.LoggingConfiguration == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.LoggingConfiguration = &v1beta1.LoggingConfiguration{}
	}

	config := bucket.Spec.ForProvider.LoggingConfiguration
	// Late initialize the target Bucket and target prefix
	config.TargetBucket = pointer.LateInitialize(config.TargetBucket, external.LoggingEnabled.TargetBucket)
	config.TargetPrefix = pointer.LateInitializeValueFromPtr(config.TargetPrefix, external.LoggingEnabled.TargetPrefix)
	// If the there is an external target grant list, and the local one does not exist
	// we create the target grant list
	if len(external.LoggingEnabled.TargetGrants) != 0 && config.TargetGrants == nil {
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

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *LoggingConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.LoggingConfiguration != nil
}

// GeneratePutBucketLoggingInput creates the input for the PutBucketLogging request for the S3 Client
func GeneratePutBucketLoggingInput(name string, config *v1beta1.LoggingConfiguration) *awss3.PutBucketLoggingInput {
	bci := &awss3.PutBucketLoggingInput{
		Bucket: pointer.ToOrNilIfZeroValue(name),
		BucketLoggingStatus: &types.BucketLoggingStatus{LoggingEnabled: &types.LoggingEnabled{
			TargetBucket: config.TargetBucket,
			TargetGrants: make([]types.TargetGrant, 0),
			TargetPrefix: pointer.ToOrNilIfZeroValue(config.TargetPrefix),
		}},
	}
	for _, grant := range config.TargetGrants {
		bci.BucketLoggingStatus.LoggingEnabled.TargetGrants = append(bci.BucketLoggingStatus.LoggingEnabled.TargetGrants, types.TargetGrant{
			Grantee: &types.Grantee{
				DisplayName:  grant.Grantee.DisplayName,
				EmailAddress: grant.Grantee.EmailAddress,
				ID:           grant.Grantee.ID,
				Type:         types.Type(grant.Grantee.Type),
				URI:          grant.Grantee.URI,
			},
			Permission: types.BucketLogsPermission(grant.Permission),
		})
	}
	return bci
}

// GenerateAWSLogging creates an S3 logging enabled struct from the local logging configuration
func GenerateAWSLogging(local *v1beta1.LoggingConfiguration) *types.LoggingEnabled {
	if local == nil {
		return nil
	}
	output := types.LoggingEnabled{
		TargetBucket: local.TargetBucket,
		TargetPrefix: pointer.ToOrNilIfZeroValue(local.TargetPrefix),
	}
	if local.TargetGrants != nil {
		output.TargetGrants = make([]types.TargetGrant, len(local.TargetGrants))
	}
	for i := range local.TargetGrants {
		target := types.TargetGrant{
			Grantee: &types.Grantee{
				DisplayName:  local.TargetGrants[i].Grantee.DisplayName,
				EmailAddress: local.TargetGrants[i].Grantee.EmailAddress,
				ID:           local.TargetGrants[i].Grantee.ID,
				Type:         types.Type(local.TargetGrants[i].Grantee.Type),
				URI:          local.TargetGrants[i].Grantee.URI,
			},
			Permission: types.BucketLogsPermission(local.TargetGrants[i].Permission),
		}

		output.TargetGrants[i] = target
	}
	return &output
}
