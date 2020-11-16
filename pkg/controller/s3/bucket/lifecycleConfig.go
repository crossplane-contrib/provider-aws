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

	"github.com/google/go-cmp/cmp/cmpopts"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	lifecycleGetFailed    = "cannot get Bucket lifecycle configuration"
	lifecyclePutFailed    = "cannot put Bucket lifecycle configuration"
	lifecycleDeleteFailed = "cannot delete Bucket lifecycle configuration"
)

// LifecycleConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type LifecycleConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize does nothing because LifecycleConfiguration might have been be
// deleted by the user.
func (*LifecycleConfigurationClient) LateInitialize(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// NewLifecycleConfigurationClient creates the client for Accelerate Configuration
func NewLifecycleConfigurationClient(client s3.BucketClient) *LifecycleConfigurationClient {
	return &LifecycleConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LifecycleConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	response, err := in.client.GetBucketLifecycleConfigurationRequest(&awss3.GetBucketLifecycleConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if bucket.Spec.ForProvider.LifecycleConfiguration == nil && s3.LifecycleConfigurationNotFound(err) {
		return Updated, nil
	}
	if resource.Ignore(s3.LifecycleConfigurationNotFound, err) != nil {
		return NeedsUpdate, errors.Wrap(err, lifecycleGetFailed)
	}
	var local []v1beta1.LifecycleRule
	if bucket.Spec.ForProvider.LifecycleConfiguration != nil {
		local = bucket.Spec.ForProvider.LifecycleConfiguration.Rules
	}
	var external []awss3.LifecycleRule
	if response != nil {
		external = response.Rules
	}
	sortFilterTags(external)
	switch {
	case len(external) != 0 && len(local) == 0:
		return NeedsDeletion, nil
	// NOTE(muvaf): We ignore ID because it might have been auto-assigned by AWS
	// and we don't have late-init for this subresource. Besides, a change in ID
	// is almost never expected.
	case cmp.Equal(external, GenerateLifecycleRules(local),
		cmpopts.IgnoreFields(awss3.LifecycleRule{}, "ID")):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// GenerateLifecycleConfiguration creates the PutBucketLifecycleConfigurationInput for the AWS SDK
func GenerateLifecycleConfiguration(name string, config *v1beta1.BucketLifecycleConfiguration) *awss3.PutBucketLifecycleConfigurationInput {
	if config == nil {
		return nil
	}
	return &awss3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(name),
		LifecycleConfiguration: &awss3.BucketLifecycleConfiguration{Rules: GenerateLifecycleRules(config.Rules)},
	}
}

// GenerateLifecycleRules creates the list of LifecycleRules for the AWS SDK
func GenerateLifecycleRules(in []v1beta1.LifecycleRule) []awss3.LifecycleRule { // nolint:gocyclo
	// NOTE(muvaf): prealloc is disabled due to AWS requiring nil instead
	// of 0-length for empty slices.
	var result []awss3.LifecycleRule // nolint:prealloc
	for _, local := range in {
		rule := awss3.LifecycleRule{
			ID:     local.ID,
			Status: awss3.ExpirationStatus(local.Status),
		}
		if local.AbortIncompleteMultipartUpload != nil {
			rule.AbortIncompleteMultipartUpload = &awss3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: &local.AbortIncompleteMultipartUpload.DaysAfterInitiation,
			}
		}
		if local.Expiration != nil {
			rule.Expiration = &awss3.LifecycleExpiration{
				Days:                      local.Expiration.Days,
				ExpiredObjectDeleteMarker: local.Expiration.ExpiredObjectDeleteMarker,
			}
			if local.Expiration.Date != nil {
				rule.Expiration.Date = &local.Expiration.Date.Time
			}
		}
		if local.NoncurrentVersionExpiration != nil {
			rule.NoncurrentVersionExpiration = &awss3.NoncurrentVersionExpiration{NoncurrentDays: local.NoncurrentVersionExpiration.NoncurrentDays}
		}
		if local.NoncurrentVersionTransitions != nil {
			rule.NoncurrentVersionTransitions = make([]awss3.NoncurrentVersionTransition, len(local.NoncurrentVersionTransitions))
			for tIndex, transition := range local.NoncurrentVersionTransitions {
				rule.NoncurrentVersionTransitions[tIndex] = awss3.NoncurrentVersionTransition{
					NoncurrentDays: transition.NoncurrentDays,
					StorageClass:   awss3.TransitionStorageClass(transition.StorageClass),
				}
			}
		}
		if local.Transitions != nil {
			rule.Transitions = make([]awss3.Transition, len(local.Transitions))
			for tIndex, transition := range local.Transitions {
				rule.Transitions[tIndex] = awss3.Transition{
					Days:         transition.Days,
					StorageClass: awss3.TransitionStorageClass(transition.StorageClass),
				}
				if transition.Date != nil {
					rule.Transitions[tIndex].Date = &transition.Date.Time
				}
			}
		}
		// This is done because S3 expects an empty filter, and never nil
		rule.Filter = &awss3.LifecycleRuleFilter{}
		if local.Filter != nil {
			rule.Filter.Prefix = local.Filter.Prefix
			if local.Filter.Tag != nil {
				rule.Filter.Tag = &awss3.Tag{Key: aws.String(local.Filter.Tag.Key), Value: aws.String(local.Filter.Tag.Value)}
			}
			if local.Filter.And != nil {
				rule.Filter.And = &awss3.LifecycleRuleAndOperator{
					Prefix: local.Filter.And.Prefix,
				}
				if local.Filter.And.Tags != nil {
					rule.Filter.And.Tags = s3.SortS3TagSet(s3.CopyTags(local.Filter.And.Tags))
				}
			}
		}
		result = append(result, rule)
	}
	return result
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LifecycleConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.LifecycleConfiguration == nil {
		return nil
	}
	input := GenerateLifecycleConfiguration(meta.GetExternalName(bucket), bucket.Spec.ForProvider.LifecycleConfiguration)
	_, err := in.client.PutBucketLifecycleConfigurationRequest(input).Send(ctx)
	return errors.Wrap(err, lifecyclePutFailed)

}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *LifecycleConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketLifecycleRequest(
		&awss3.DeleteBucketLifecycleInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, lifecycleDeleteFailed)
}

func sortFilterTags(rules []awss3.LifecycleRule) {
	for i := range rules {
		if rules[i].Filter != nil && rules[i].Filter.And != nil {
			rules[i].Filter.And.Tags = s3.SortS3TagSet(rules[i].Filter.And.Tags)
		}
	}
}
