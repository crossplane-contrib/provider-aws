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

var _ BucketResource = &LifecycleConfigurationClient{}

// LifecycleConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type LifecycleConfigurationClient struct {
	config *v1beta1.BucketLifecycleConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *LifecycleConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	// GetBucketLifecycleConfigurationRequest throws an error if nothing exists externally
	// Future work can be done to support brownfield initialization for the LifecycleConfiguration
	// TODO
	return nil
}

// NewLifecycleConfigurationClient creates the client for Accelerate Configuration
func NewLifecycleConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *LifecycleConfigurationClient {
	return &LifecycleConfigurationClient{config: bucket.Spec.ForProvider.LifecycleConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LifecycleConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketLifecycleConfigurationRequest(&awss3.GetBucketLifecycleConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		if s3.LifecycleConfigurationNotFound(err) && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, lifecycleGetFailed)
	}

	switch {
	case len(conf.Rules) != 0 && in.config == nil:
		return NeedsDeletion, nil
	case in.config == nil && len(conf.Rules) == 0:
		return Updated, nil
	case cmp.Equal(conf.Rules, GenerateRules(in.config)):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// GenerateLifecycleConfiguration creates the PutBucketLifecycleConfigurationInput for the AWS SDK
func GenerateLifecycleConfiguration(name string, in *LifecycleConfigurationClient) *awss3.PutBucketLifecycleConfigurationInput {
	return &awss3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(name),
		LifecycleConfiguration: &awss3.BucketLifecycleConfiguration{Rules: GenerateRules(in.config)},
	}
}

// GenerateRules creates the list of LifecycleRules for the AWS SDK
func GenerateRules(in *v1beta1.BucketLifecycleConfiguration) []awss3.LifecycleRule { // nolint:gocyclo
	rules := make([]awss3.LifecycleRule, len(in.Rules))
	for i, local := range in.Rules {
		rule := awss3.LifecycleRule{
			ID:     local.ID,
			Status: awss3.ExpirationStatus(local.Status),
		}
		if local.AbortIncompleteMultipartUpload != nil {
			rule.AbortIncompleteMultipartUpload = &awss3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: local.AbortIncompleteMultipartUpload.DaysAfterInitiation,
			}
		}
		if local.Expiration != nil {
			rule.Expiration = &awss3.LifecycleExpiration{
				Date:                      &local.Expiration.Date.Time,
				Days:                      local.Expiration.Days,
				ExpiredObjectDeleteMarker: local.Expiration.ExpiredObjectDeleteMarker,
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
					Date:         &transition.Date.Time,
					Days:         transition.Days,
					StorageClass: awss3.TransitionStorageClass(transition.StorageClass),
				}
			}
		}
		if local.Filter != nil {
			rule.Filter = &awss3.LifecycleRuleFilter{
				Prefix: local.Filter.Prefix,
				Tag:    copyTag(local.Filter.Tag),
			}
			if local.Filter.And != nil {
				rule.Filter.And = &awss3.LifecycleRuleAndOperator{
					Prefix: local.Filter.And.Prefix,
					Tags:   copyTags(local.Filter.And.Tags),
				}
			}
		}
		rules[i] = rule
	}
	return rules
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LifecycleConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}

	config := GenerateLifecycleConfiguration(meta.GetExternalName(bucket), in)

	_, err := in.client.PutBucketLifecycleConfigurationRequest(config).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket lifecycle")

}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *LifecycleConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketLifecycleRequest(
		&awss3.DeleteBucketLifecycleInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket lifecycle configuration")
}
