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

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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

// NewLifecycleConfigurationClient creates the client for Accelerate Configuration
func NewLifecycleConfigurationClient(client s3.BucketClient) *LifecycleConfigurationClient {
	return &LifecycleConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LifecycleConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	response, err := in.client.GetBucketLifecycleConfiguration(ctx, &awss3.GetBucketLifecycleConfigurationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if bucket.Spec.ForProvider.LifecycleConfiguration == nil && s3.LifecycleConfigurationNotFound(err) {
		return Updated, nil
	}
	if resource.Ignore(s3.LifecycleConfigurationNotFound, err) != nil {
		return NeedsUpdate, errorutils.Wrap(err, lifecycleGetFailed)
	}
	var local []v1beta1.LifecycleRule
	if bucket.Spec.ForProvider.LifecycleConfiguration != nil {
		local = bucket.Spec.ForProvider.LifecycleConfiguration.Rules
	}
	var external []types.LifecycleRule
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
		cmpopts.IgnoreFields(types.LifecycleRule{}, "ID"), cmpopts.IgnoreTypes(document.NoSerde{})):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LifecycleConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.LifecycleConfiguration == nil {
		return nil
	}
	input := GenerateLifecycleConfiguration(meta.GetExternalName(bucket), bucket.Spec.ForProvider.LifecycleConfiguration)
	_, err := in.client.PutBucketLifecycleConfiguration(ctx, input)
	return errorutils.Wrap(err, lifecyclePutFailed)

}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *LifecycleConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketLifecycle(ctx,
		&awss3.DeleteBucketLifecycleInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, lifecycleDeleteFailed)
}

// LateInitialize does nothing because LifecycleConfiguration might have been be
// deleted by the user.
func (in *LifecycleConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketLifecycleConfiguration(ctx, &awss3.GetBucketLifecycleConfigurationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.LifecycleConfigurationNotFound, err), lifecycleGetFailed)
	}

	// We need the second check here because by default the lifecycle is not set
	if external == nil || len(external.Rules) == 0 {
		return nil
	}

	fp := &bucket.Spec.ForProvider
	if fp.LifecycleConfiguration == nil {
		fp.LifecycleConfiguration = &v1beta1.BucketLifecycleConfiguration{}
	}

	if fp.LifecycleConfiguration.Rules == nil {
		createLifecycleRulesFromExternal(external.Rules, fp.LifecycleConfiguration)
	}

	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *LifecycleConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.LifecycleConfiguration != nil
}

// GenerateLifecycleConfiguration creates the PutBucketLifecycleConfigurationInput for the AWS SDK
func GenerateLifecycleConfiguration(name string, config *v1beta1.BucketLifecycleConfiguration) *awss3.PutBucketLifecycleConfigurationInput {
	if config == nil {
		return nil
	}
	return &awss3.PutBucketLifecycleConfigurationInput{
		Bucket:                 pointer.ToOrNilIfZeroValue(name),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{Rules: GenerateLifecycleRules(config.Rules)},
	}
}

// GenerateLifecycleRules creates the list of LifecycleRules for the AWS SDK
func GenerateLifecycleRules(in []v1beta1.LifecycleRule) []types.LifecycleRule { //nolint:gocyclo
	// NOTE(muvaf): prealloc is disabled due to AWS requiring nil instead
	// of 0-length for empty slices.
	var result []types.LifecycleRule //nolint:prealloc
	for _, local := range in {
		rule := types.LifecycleRule{
			ID:     local.ID,
			Status: types.ExpirationStatus(local.Status),
		}
		if local.AbortIncompleteMultipartUpload != nil {
			rule.AbortIncompleteMultipartUpload = &types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: local.AbortIncompleteMultipartUpload.DaysAfterInitiation,
			}
		}
		if local.Expiration != nil {
			rule.Expiration = &types.LifecycleExpiration{
				Days:                      local.Expiration.Days,
				ExpiredObjectDeleteMarker: local.Expiration.ExpiredObjectDeleteMarker,
			}
			if local.Expiration.Date != nil {
				rule.Expiration.Date = ptr.To(local.Expiration.Date.Time)
			}
		}
		if local.NoncurrentVersionExpiration != nil {
			rule.NoncurrentVersionExpiration = &types.NoncurrentVersionExpiration{NoncurrentDays: local.NoncurrentVersionExpiration.NoncurrentDays}
		}
		if local.NoncurrentVersionTransitions != nil {
			rule.NoncurrentVersionTransitions = make([]types.NoncurrentVersionTransition, len(local.NoncurrentVersionTransitions))
			for tIndex, transition := range local.NoncurrentVersionTransitions {
				rule.NoncurrentVersionTransitions[tIndex] = types.NoncurrentVersionTransition{
					NoncurrentDays: transition.NoncurrentDays,
					StorageClass:   types.TransitionStorageClass(transition.StorageClass),
				}
			}
		}
		if local.Transitions != nil {
			rule.Transitions = make([]types.Transition, len(local.Transitions))
			for tIndex, transition := range local.Transitions {
				rule.Transitions[tIndex] = types.Transition{
					Days:         transition.Days,
					StorageClass: types.TransitionStorageClass(transition.StorageClass),
				}
				if transition.Date != nil {
					rule.Transitions[tIndex].Date = ptr.To(transition.Date.Time)
				}
			}
		}
		// This is done because S3 expects an empty filter, and never nil
		rule.Filter = &types.LifecycleRuleFilterMemberPrefix{}
		if local.Filter != nil {
			if local.Filter.Prefix != nil {
				rule.Filter = &types.LifecycleRuleFilterMemberPrefix{Value: *local.Filter.Prefix}
			}
			if local.Filter.Tag != nil {
				rule.Filter = &types.LifecycleRuleFilterMemberTag{Value: types.Tag{Key: pointer.ToOrNilIfZeroValue(local.Filter.Tag.Key), Value: pointer.ToOrNilIfZeroValue(local.Filter.Tag.Value)}}
			}
			if local.Filter.And != nil {
				andOperator := types.LifecycleRuleAndOperator{
					Prefix: local.Filter.And.Prefix,
				}
				if local.Filter.And.Tags != nil {
					andOperator.Tags = s3.SortS3TagSet(s3.CopyTags(local.Filter.And.Tags))
				}
				rule.Filter = &types.LifecycleRuleFilterMemberAnd{Value: andOperator}
			}
		}
		result = append(result, rule)
	}
	return result
}

func sortFilterTags(rules []types.LifecycleRule) {
	for i := range rules {
		andOperator, ok := rules[i].Filter.(*types.LifecycleRuleFilterMemberAnd)
		if ok {
			andOperator.Value.Tags = s3.SortS3TagSet(andOperator.Value.Tags)
		}
	}
}

func createLifecycleRulesFromExternal(external []types.LifecycleRule, config *v1beta1.BucketLifecycleConfiguration) { //nolint:gocyclo
	if config.Rules != nil {
		return
	}

	config.Rules = make([]v1beta1.LifecycleRule, len(external))

	for i, rule := range external {
		config.Rules[i] = v1beta1.LifecycleRule{
			ID:     rule.ID,
			Status: string(rule.Status),
		}

		if rule.Filter != nil {
			config.Rules[i].Filter = &v1beta1.LifecycleRuleFilter{}
			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3/types@v1.3.0#LifecycleRuleFilter
			// type switches can be used to check the union value
			union := rule.Filter
			switch v := union.(type) {
			case *types.LifecycleRuleFilterMemberAnd:
				// Value is types.ReplicationRuleAndOperator
				config.Rules[i].Filter.And = &v1beta1.LifecycleRuleAndOperator{}
				config.Rules[i].Filter.And.Prefix = v.Value.Prefix
				config.Rules[i].Filter.And.Tags = GenerateLocalTagging(v.Value.Tags).TagSet
			case *types.LifecycleRuleFilterMemberPrefix:
				// Value is string
				config.Rules[i].Filter = &v1beta1.LifecycleRuleFilter{}
				config.Rules[i].Filter.Prefix = aws.String(v.Value)
			case *types.LifecycleRuleFilterMemberTag:
				// Value is types.Tag
				config.Rules[i].Filter.Tag = &v1beta1.Tag{}
				config.Rules[i].Filter.Tag.Key = aws.ToString(v.Value.Key)
				config.Rules[i].Filter.Tag.Value = aws.ToString(v.Value.Value)
			case *types.UnknownUnionMember:
			//	fmt.Println("unknown tag:", v.Tag)
			default:
				//	fmt.Println("union is nil or unknown type")
			}
		}

		if rule.AbortIncompleteMultipartUpload != nil {
			config.Rules[i].AbortIncompleteMultipartUpload = &v1beta1.AbortIncompleteMultipartUpload{}
			config.Rules[i].AbortIncompleteMultipartUpload.DaysAfterInitiation = rule.AbortIncompleteMultipartUpload.DaysAfterInitiation
		}
		if rule.Expiration != nil {
			config.Rules[i].Expiration = &v1beta1.LifecycleExpiration{}
			config.Rules[i].Expiration.Date = pointer.LateInitializeTimePtr(
				config.Rules[i].Expiration.Date,
				rule.Expiration.Date,
			)
			config.Rules[i].Expiration.Days = rule.Expiration.Days
			config.Rules[i].Expiration.ExpiredObjectDeleteMarker = rule.Expiration.ExpiredObjectDeleteMarker
		}
		if rule.NoncurrentVersionExpiration != nil {
			config.Rules[i].NoncurrentVersionExpiration = &v1beta1.NoncurrentVersionExpiration{}
			config.Rules[i].NoncurrentVersionExpiration.NoncurrentDays = rule.NoncurrentVersionExpiration.NoncurrentDays
		}
		if len(rule.NoncurrentVersionTransitions) != 0 {
			config.Rules[i].NoncurrentVersionTransitions = make([]v1beta1.NoncurrentVersionTransition, len(rule.NoncurrentVersionTransitions))

			for j, nvt := range rule.NoncurrentVersionTransitions {
				config.Rules[i].NoncurrentVersionTransitions[j].NoncurrentDays = nvt.NoncurrentDays
				config.Rules[i].NoncurrentVersionTransitions[j].StorageClass = string(nvt.StorageClass)
			}
		}
		if len(rule.Transitions) != 0 {
			config.Rules[i].Transitions = make([]v1beta1.Transition, len(rule.Transitions))
			for j, transition := range rule.Transitions {
				config.Rules[i].Transitions[j].Days = transition.Days
				config.Rules[i].Transitions[j].Date = pointer.LateInitializeTimePtr(
					config.Rules[i].Transitions[j].Date,
					transition.Date,
				)
				config.Rules[i].Transitions[j].StorageClass = string(transition.StorageClass)
			}
		}
	}
}
