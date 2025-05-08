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
		cmpopts.EquateEmpty(), cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}), cmpopts.IgnoreTypes(document.NoSerde{})) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LoggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := GeneratePutBucketLoggingInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.LoggingConfiguration)
	_, err := in.client.PutBucketLogging(ctx, input)
	return errorutils.Wrap(err, loggingPutFailed)
}

// Delete does nothing because there is no deletion call for logging config.
func (*LoggingConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// LateInitialize is not needed because loggingConfiguration is not something which is created be default
// it means if it is not set in the desired state, but it exists on aws side it should be deleted(by CreateOrUpdate), not late initialized
func (in *LoggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *LoggingConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.LoggingConfiguration != nil
}

// GeneratePutBucketLoggingInput creates the input for the PutBucketLogging request for the S3 Client
func GeneratePutBucketLoggingInput(name string, config *v1beta1.LoggingConfiguration) *awss3.PutBucketLoggingInput {
	bci := &awss3.PutBucketLoggingInput{
		Bucket:              pointer.ToOrNilIfZeroValue(name),
		BucketLoggingStatus: &types.BucketLoggingStatus{},
	}
	if config != nil {
		bci = &awss3.PutBucketLoggingInput{
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
		TargetGrants: []types.TargetGrant{},
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

		output.TargetGrants = append(output.TargetGrants, target)
	}
	return &output
}
