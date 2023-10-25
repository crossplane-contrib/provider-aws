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
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	sseGetFailed    = "cannot get encryption configuration"
	ssePutFailed    = "cannot put encryption configuration"
	sseDeleteFailed = "cannot delete encryption configuration"
)

// SSEConfigurationClient is the client for API methods and reconciling the ServerSideEncryptionConfiguration
type SSEConfigurationClient struct {
	client s3.BucketClient
}

// NewSSEConfigurationClient creates the client for Server Side Encryption Configuration
func NewSSEConfigurationClient(client s3.BucketClient) *SSEConfigurationClient {
	return &SSEConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *SSEConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { //nolint:gocyclo
	config := bucket.Spec.ForProvider.ServerSideEncryptionConfiguration
	external, err := in.client.GetBucketEncryption(ctx, &awss3.GetBucketEncryptionInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		if s3.SSEConfigurationNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.SSEConfigurationNotFound, err), sseGetFailed)
	}

	switch {
	case external.ServerSideEncryptionConfiguration != nil && config == nil:
		return NeedsDeletion, nil
	case external.ServerSideEncryptionConfiguration == nil && config == nil:
		return Updated, nil
	case external.ServerSideEncryptionConfiguration == nil && config != nil:
		return NeedsUpdate, nil
	case len(external.ServerSideEncryptionConfiguration.Rules) != len(config.Rules):
		return NeedsUpdate, nil
	}

	for i, Rule := range config.Rules {
		outputRule := external.ServerSideEncryptionConfiguration.Rules[i].ApplyServerSideEncryptionByDefault
		if pointer.StringValue(outputRule.KMSMasterKeyID) != pointer.StringValue(Rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID) {
			return NeedsUpdate, nil
		}
		if string(outputRule.SSEAlgorithm) != Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm {
			return NeedsUpdate, nil
		}
		if external.ServerSideEncryptionConfiguration.Rules[i].BucketKeyEnabled != Rule.BucketKeyEnabled {
			return NeedsUpdate, nil
		}
	}

	return Updated, nil
}

// CreateOrUpdate sends a request to have resource created on awsclient.
func (in *SSEConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.ServerSideEncryptionConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketEncryptionInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.ServerSideEncryptionConfiguration)
	_, err := in.client.PutBucketEncryption(ctx, input)
	return errorutils.Wrap(err, ssePutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *SSEConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketEncryption(ctx,
		&awss3.DeleteBucketEncryptionInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, sseDeleteFailed)
}

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (in *SSEConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketEncryption(ctx, &awss3.GetBucketEncryptionInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.SSEConfigurationNotFound, err), sseGetFailed)
	}

	// We need the second check here because by default the SSE is not set
	if external == nil || external.ServerSideEncryptionConfiguration == nil || len(external.ServerSideEncryptionConfiguration.Rules) == 0 {
		return nil
	}

	fp := &bucket.Spec.ForProvider
	if fp.ServerSideEncryptionConfiguration == nil {
		fp.ServerSideEncryptionConfiguration = &v1beta1.ServerSideEncryptionConfiguration{}
	}

	if fp.ServerSideEncryptionConfiguration.Rules == nil {
		fp.ServerSideEncryptionConfiguration.Rules = GenerateLocalBucketEncryption(external.ServerSideEncryptionConfiguration)
	}

	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *SSEConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.ServerSideEncryptionConfiguration != nil
}

// GeneratePutBucketEncryptionInput creates the input for the PutBucketEncryption request for the S3 Client
func GeneratePutBucketEncryptionInput(name string, config *v1beta1.ServerSideEncryptionConfiguration) *awss3.PutBucketEncryptionInput {
	bei := &awss3.PutBucketEncryptionInput{
		Bucket: pointer.ToOrNilIfZeroValue(name),
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: make([]types.ServerSideEncryptionRule, len(config.Rules)),
		},
	}
	for i, rule := range config.Rules {
		bei.ServerSideEncryptionConfiguration.Rules[i] = types.ServerSideEncryptionRule{
			BucketKeyEnabled: rule.BucketKeyEnabled,
			ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
				KMSMasterKeyID: rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				SSEAlgorithm:   types.ServerSideEncryption(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			},
		}
	}
	return bei
}

// GenerateLocalBucketEncryption creates the local ServerSideEncryptionConfiguration from the S3 Client request
func GenerateLocalBucketEncryption(config *types.ServerSideEncryptionConfiguration) []v1beta1.ServerSideEncryptionRule {
	rules := make([]v1beta1.ServerSideEncryptionRule, len(config.Rules))
	for i, rule := range config.Rules {
		rules[i] = v1beta1.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
				KMSMasterKeyID: rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				SSEAlgorithm:   string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			},
		}
	}
	return rules
}
