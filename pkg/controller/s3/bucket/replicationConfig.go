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
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	replicationGetFailed    = "cannot get replication configuration"
	replicationPutFailed    = "cannot put Bucket replication"
	replicationDeleteFailed = "cannot delete Bucket replication"
)

// ReplicationConfigurationClient is the client for API methods and reconciling the ReplicationConfiguration
type ReplicationConfigurationClient struct {
	client s3.BucketClient
}

// NewReplicationConfigurationClient creates the client for Replication Configuration
func NewReplicationConfigurationClient(client s3.BucketClient) *ReplicationConfigurationClient {
	return &ReplicationConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *ReplicationConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { //nolint:gocyclo
	external, err := in.client.GetBucketReplication(ctx, &awss3.GetBucketReplicationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	config := bucket.Spec.ForProvider.ReplicationConfiguration
	if err != nil {
		if s3.ReplicationConfigurationNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.ReplicationConfigurationNotFound, err), replicationGetFailed)
	}

	switch {
	case (external == nil || external.ReplicationConfiguration == nil) && config != nil:
		return NeedsUpdate, nil
	case (external == nil || external.ReplicationConfiguration == nil) && config == nil:
		return Updated, nil
	case external.ReplicationConfiguration != nil && config == nil:
		return NeedsDeletion, nil
	}

	source := GenerateReplicationConfiguration(config)

	return IsUpToDate(external.ReplicationConfiguration, source)
}

// IsUpToDate determines whether a replication configuration needs to be updated
func IsUpToDate(external *types.ReplicationConfiguration, source *types.ReplicationConfiguration) (ResourceStatus, error) {
	sortReplicationRules(external.Rules)
	if cmp.Equal(external, source, cmpopts.IgnoreTypes(document.NoSerde{})) {
		return Updated, nil
	}
	return NeedsUpdate, nil
}

// CreateOrUpdate sends a request to have resource created on awsclient.
func (in *ReplicationConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.ReplicationConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketReplicationInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.ReplicationConfiguration)
	_, err := in.client.PutBucketReplication(ctx, input)
	return errorutils.Wrap(err, replicationPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *ReplicationConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketReplication(ctx,
		&awss3.DeleteBucketReplicationInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, replicationDeleteFailed)
}

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (in *ReplicationConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketReplication(ctx, &awss3.GetBucketReplicationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.ReplicationConfigurationNotFound, err), replicationGetFailed)
	}

	if external == nil || external.ReplicationConfiguration == nil || len(external.ReplicationConfiguration.Rules) == 0 {
		return nil
	}

	fp := &bucket.Spec.ForProvider
	if fp.ReplicationConfiguration == nil {
		// We need the configuration to exist so we can initialize
		fp.ReplicationConfiguration = &v1beta1.ReplicationConfiguration{}
	}
	fp.ReplicationConfiguration.Role = pointer.LateInitialize(fp.ReplicationConfiguration.Role, external.ReplicationConfiguration.Role)
	if fp.ReplicationConfiguration.Rules == nil {
		createReplicationRulesFromExternal(external.ReplicationConfiguration, fp.ReplicationConfiguration)
	}
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *ReplicationConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.ReplicationConfiguration != nil
}

func createReplicationRulesFromExternal(external *types.ReplicationConfiguration, config *v1beta1.ReplicationConfiguration) { //nolint:gocyclo
	if config.Rules != nil {
		return
	}
	config.Rules = make([]v1beta1.ReplicationRule, len(external.Rules))

	for i, rule := range external.Rules {
		config.Rules[i] = v1beta1.ReplicationRule{
			ID:       rule.ID,
			Priority: rule.Priority,
			Filter:   &v1beta1.ReplicationRuleFilter{},
			Status:   string(rule.Status),
		}

		if rule.Filter != nil {
			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3/types@v1.3.0#ReplicationRuleFilter
			// type switches can be used to check the union value
			union := rule.Filter
			switch v := union.(type) {
			case *types.ReplicationRuleFilterMemberAnd:
				// Value is types.ReplicationRuleAndOperator
				config.Rules[i].Filter.And = &v1beta1.ReplicationRuleAndOperator{}
				config.Rules[i].Filter.And.Prefix = v.Value.Prefix
				config.Rules[i].Filter.And.Tags = GenerateLocalTagging(v.Value.Tags).TagSet
			case *types.ReplicationRuleFilterMemberPrefix:
				// Value is string
				config.Rules[i].Filter = &v1beta1.ReplicationRuleFilter{}
				config.Rules[i].Filter.Prefix = aws.String(v.Value)
			case *types.ReplicationRuleFilterMemberTag:
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

		if rule.DeleteMarkerReplication != nil {
			config.Rules[i].DeleteMarkerReplication = &v1beta1.DeleteMarkerReplication{}
			config.Rules[i].DeleteMarkerReplication.Status = string(rule.DeleteMarkerReplication.Status)
		}
		if rule.Destination != nil {
			config.Rules[i].Destination.Account = rule.Destination.Account
			config.Rules[i].Destination.Bucket = rule.Destination.Bucket
			config.Rules[i].Destination.StorageClass = pointer.ToOrNilIfZeroValue(string(rule.Destination.StorageClass))
			if rule.Destination.AccessControlTranslation != nil {
				config.Rules[i].Destination.AccessControlTranslation = &v1beta1.AccessControlTranslation{}
				config.Rules[i].Destination.AccessControlTranslation.Owner = string(rule.Destination.AccessControlTranslation.Owner)
			}
			if rule.Destination.EncryptionConfiguration != nil {
				config.Rules[i].Destination.EncryptionConfiguration = &v1beta1.EncryptionConfiguration{}
				config.Rules[i].Destination.EncryptionConfiguration.ReplicaKmsKeyID = rule.Destination.EncryptionConfiguration.ReplicaKmsKeyID
			}
			if rule.Destination.Metrics != nil {
				config.Rules[i].Destination.Metrics = &v1beta1.Metrics{}
				if rule.Destination.Metrics.EventThreshold != nil {
					config.Rules[i].Destination.Metrics.EventThreshold = &v1beta1.ReplicationTimeValue{}
					config.Rules[i].Destination.Metrics.EventThreshold.Minutes = rule.Destination.Metrics.EventThreshold.Minutes
				}
				config.Rules[i].Destination.Metrics.Status = string(rule.Destination.Metrics.Status)
			}
			if rule.Destination.ReplicationTime != nil {
				config.Rules[i].Destination.ReplicationTime = &v1beta1.ReplicationTime{}
				config.Rules[i].Destination.ReplicationTime.Status = string(rule.Destination.ReplicationTime.Status)
				if rule.Destination.ReplicationTime.Time != nil {
					config.Rules[i].Destination.ReplicationTime.Time.Minutes = rule.Destination.ReplicationTime.Time.Minutes
				}
			}
		}
		if rule.ExistingObjectReplication != nil {
			config.Rules[i].ExistingObjectReplication = &v1beta1.ExistingObjectReplication{}
			config.Rules[i].ExistingObjectReplication.Status = string(rule.ExistingObjectReplication.Status)
		}
		if rule.SourceSelectionCriteria != nil && rule.SourceSelectionCriteria.SseKmsEncryptedObjects != nil {
			config.Rules[i].SourceSelectionCriteria = &v1beta1.SourceSelectionCriteria{}
			config.Rules[i].SourceSelectionCriteria.SseKmsEncryptedObjects.Status = string(rule.SourceSelectionCriteria.SseKmsEncryptedObjects.Status)
		}
	}
}

func sortReplicationRules(rules []types.ReplicationRule) {

	sort.Slice(rules, func(i, j int) bool {
		// Sort first by Rule ID
		if a, b := rules[i].ID, rules[j].ID; a != b {
			return aws.ToString(a) < aws.ToString(b)
		}
		// AWS won't let you have rules with the same name, but that may be defined
		return true
	})

	for i := range rules {
		andOperator, ok := rules[i].Filter.(*types.ReplicationRuleFilterMemberAnd)
		if ok {
			andOperator.Value.Tags = s3.SortS3TagSet(andOperator.Value.Tags)
		}
	}
}

func copyDestination(input *v1beta1.ReplicationRule, newRule *types.ReplicationRule) {
	newRule.Destination = &types.Destination{
		AccessControlTranslation: nil,
		Account:                  input.Destination.Account,
		Bucket:                   input.Destination.Bucket,
		EncryptionConfiguration:  nil,
		Metrics:                  nil,
		ReplicationTime:          nil,
		StorageClass:             types.StorageClass(pointer.StringValue(input.Destination.StorageClass)),
	}
	if input.Destination.AccessControlTranslation != nil {
		newRule.Destination.AccessControlTranslation = &types.AccessControlTranslation{
			Owner: types.OwnerOverride(input.Destination.AccessControlTranslation.Owner),
		}
	}
	if input.Destination.EncryptionConfiguration != nil {
		newRule.Destination.EncryptionConfiguration = &types.EncryptionConfiguration{
			ReplicaKmsKeyID: input.Destination.EncryptionConfiguration.ReplicaKmsKeyID,
		}
	}
	if input.Destination.Metrics != nil {
		newRule.Destination.Metrics = &types.Metrics{
			Status: types.MetricsStatus(input.Destination.Metrics.Status),
		}
		if input.Destination.Metrics.EventThreshold != nil {
			newRule.Destination.Metrics.EventThreshold = &types.ReplicationTimeValue{Minutes: input.Destination.Metrics.EventThreshold.Minutes}
		}
	}
	if input.Destination.ReplicationTime != nil {
		newRule.Destination.ReplicationTime = &types.ReplicationTime{
			Status: types.ReplicationTimeStatus(input.Destination.ReplicationTime.Status),
			Time:   nil,
		}
		if input.Destination.ReplicationTime != nil {
			newRule.Destination.ReplicationTime.Time = &types.ReplicationTimeValue{
				Minutes: input.Destination.ReplicationTime.Time.Minutes,
			}
		}
	}
}

func createRule(input v1beta1.ReplicationRule) types.ReplicationRule {
	Rule := input
	newRule := types.ReplicationRule{
		ID:       Rule.ID,
		Priority: Rule.Priority,
		Filter:   &types.ReplicationRuleFilterMemberPrefix{Value: ""},
		Status:   types.ReplicationRuleStatus(Rule.Status),
	}
	if Rule.Filter != nil {
		switch {
		case Rule.Filter.And != nil:
			andOperator := &types.ReplicationRuleAndOperator{
				Prefix: Rule.Filter.And.Prefix,
			}
			if Rule.Filter.And.Tags != nil {
				andOperator.Tags = s3.SortS3TagSet(s3.CopyTags(Rule.Filter.And.Tags))
			}
			newRule.Filter = &types.ReplicationRuleFilterMemberAnd{Value: *andOperator}
		case Rule.Filter.Tag != nil:
			newRule.Filter = &types.ReplicationRuleFilterMemberTag{Value: types.Tag{Key: pointer.ToOrNilIfZeroValue(Rule.Filter.Tag.Key), Value: pointer.ToOrNilIfZeroValue(Rule.Filter.Tag.Value)}}
		case Rule.Filter.Prefix != nil:
			newRule.Filter = &types.ReplicationRuleFilterMemberPrefix{Value: *Rule.Filter.Prefix}
		}
	}
	if Rule.SourceSelectionCriteria != nil {
		newRule.SourceSelectionCriteria = &types.SourceSelectionCriteria{
			SseKmsEncryptedObjects: &types.SseKmsEncryptedObjects{
				Status: types.SseKmsEncryptedObjectsStatus(Rule.SourceSelectionCriteria.SseKmsEncryptedObjects.Status),
			},
		}
	}
	if Rule.ExistingObjectReplication != nil {
		newRule.ExistingObjectReplication = &types.ExistingObjectReplication{
			Status: types.ExistingObjectReplicationStatus(Rule.ExistingObjectReplication.Status),
		}
	}
	if Rule.DeleteMarkerReplication != nil {
		newRule.DeleteMarkerReplication = &types.DeleteMarkerReplication{Status: types.DeleteMarkerReplicationStatus(Rule.DeleteMarkerReplication.Status)}
	}

	copyDestination(&Rule, &newRule)
	return newRule
}

// GenerateReplicationConfiguration is responsible for creating the Replication Configuration for requests.
func GenerateReplicationConfiguration(config *v1beta1.ReplicationConfiguration) *types.ReplicationConfiguration {
	source := &types.ReplicationConfiguration{
		Role:  config.Role,
		Rules: make([]types.ReplicationRule, len(config.Rules)),
	}

	for i, Rule := range config.Rules {
		source.Rules[i] = createRule(Rule)
	}
	return source
}

// GeneratePutBucketReplicationInput creates the input for the PutBucketReplication request for the S3 Client
func GeneratePutBucketReplicationInput(name string, config *v1beta1.ReplicationConfiguration) *awss3.PutBucketReplicationInput {
	return &awss3.PutBucketReplicationInput{
		Bucket:                   pointer.ToOrNilIfZeroValue(name),
		ReplicationConfiguration: GenerateReplicationConfiguration(config),
	}
}
