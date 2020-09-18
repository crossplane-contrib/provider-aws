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

// LifecycleConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type LifecycleConfigurationClient struct {
	config *v1beta1.BucketLifecycleConfiguration
	client s3.BucketClient
}

// NewLifecycleConfigurationClient creates the client for Accelerate Configuration
func NewLifecycleConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *LifecycleConfigurationClient {
	return &LifecycleConfigurationClient{config: bucket.Spec.Parameters.LifecycleConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LifecycleConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketLifecycleConfigurationRequest(&awss3.GetBucketLifecycleConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchLifecycleConfiguration" && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket lifecycle")
	}

	if len(conf.Rules) != 0 && in.config == nil {
		return NeedsDeletion, nil
	}

	if cmp.Equal(conf.Rules, in.generateConfiguration()) {
		return Updated, nil
	}
	return NeedsUpdate, nil
}

func (in *LifecycleConfigurationClient) generateConfiguration() []awss3.LifecycleRule {
	rules := make([]awss3.LifecycleRule, len(in.config.Rules))
	for i, local := range in.config.Rules {
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
		rules[i] = rule
	}
	return rules
}

// Create sends a request to have resource created on AWS
func (in *LifecycleConfigurationClient) Create(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}

	config := &awss3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(meta.GetExternalName(bucket)),
		LifecycleConfiguration: &awss3.BucketLifecycleConfiguration{Rules: in.generateConfiguration()},
	}

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
