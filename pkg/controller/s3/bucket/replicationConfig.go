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
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
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

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (*ReplicationConfigurationClient) LateInitialize(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// NewReplicationConfigurationClient creates the client for Replication Configuration
func NewReplicationConfigurationClient(client s3.BucketClient) *ReplicationConfigurationClient {
	return &ReplicationConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *ReplicationConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { // nolint:gocyclo
	external, err := in.client.GetBucketReplicationRequest(&awss3.GetBucketReplicationInput{Bucket: awsclient.String(meta.GetExternalName(bucket))}).Send(ctx)
	config := bucket.Spec.ForProvider.ReplicationConfiguration
	if err != nil {
		if s3.ReplicationConfigurationNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, awsclient.Wrap(resource.Ignore(s3.ReplicationConfigurationNotFound, err), replicationGetFailed)
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

	sortReplicationRules(external.ReplicationConfiguration.Rules)

	if cmp.Equal(external.ReplicationConfiguration, source) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

func copyDestination(input *v1beta1.ReplicationRule, newRule *awss3.ReplicationRule) {
	newRule.Destination = &awss3.Destination{
		AccessControlTranslation: nil,
		Account:                  input.Destination.Account,
		Bucket:                   input.Destination.Bucket,
		EncryptionConfiguration:  nil,
		Metrics:                  nil,
		ReplicationTime:          nil,
		StorageClass:             awss3.StorageClass(awsclient.StringValue(input.Destination.StorageClass)),
	}
	if input.Destination.AccessControlTranslation != nil {
		newRule.Destination.AccessControlTranslation = &awss3.AccessControlTranslation{
			Owner: awss3.OwnerOverride(input.Destination.AccessControlTranslation.Owner),
		}
	}
	if input.Destination.EncryptionConfiguration != nil {
		newRule.Destination.EncryptionConfiguration = &awss3.EncryptionConfiguration{
			ReplicaKmsKeyID: awsclient.String(input.Destination.EncryptionConfiguration.ReplicaKmsKeyID),
		}
	}
	if input.Destination.Metrics != nil {
		newRule.Destination.Metrics = &awss3.Metrics{
			EventThreshold: &awss3.ReplicationTimeValue{Minutes: &input.Destination.Metrics.EventThreshold.Minutes},
			Status:         awss3.MetricsStatus(input.Destination.Metrics.Status),
		}
	}
	if input.Destination.ReplicationTime != nil {
		newRule.Destination.ReplicationTime = &awss3.ReplicationTime{
			Status: awss3.ReplicationTimeStatus(input.Destination.ReplicationTime.Status),
			Time:   nil,
		}
		if input.Destination.ReplicationTime != nil {
			newRule.Destination.ReplicationTime.Time = &awss3.ReplicationTimeValue{
				Minutes: &input.Destination.ReplicationTime.Time.Minutes,
			}
		}
	}
}

func createRule(input v1beta1.ReplicationRule) awss3.ReplicationRule {
	Rule := input
	newRule := awss3.ReplicationRule{
		ID:       Rule.ID,
		Priority: Rule.Priority,
		Status:   awss3.ReplicationRuleStatus(Rule.Status),
	}
	if Rule.Filter != nil {
		newRule.Filter = &awss3.ReplicationRuleFilter{
			And:    nil,
			Prefix: Rule.Filter.Prefix,
			Tag:    nil,
		}
		if Rule.Filter.And != nil {
			newRule.Filter.And = &awss3.ReplicationRuleAndOperator{
				Prefix: Rule.Filter.And.Prefix,
			}
			if Rule.Filter.And.Tags != nil {
				newRule.Filter.And.Tags = s3.SortS3TagSet(s3.CopyTags(Rule.Filter.And.Tags))
			}
		}
		if Rule.Filter.Tag != nil {
			newRule.Filter.Tag = &awss3.Tag{Key: awsclient.String(Rule.Filter.Tag.Key), Value: awsclient.String(Rule.Filter.Tag.Value)}
		}
	}
	if Rule.SourceSelectionCriteria != nil {
		newRule.SourceSelectionCriteria = &awss3.SourceSelectionCriteria{
			SseKmsEncryptedObjects: &awss3.SseKmsEncryptedObjects{
				Status: awss3.SseKmsEncryptedObjectsStatus(Rule.SourceSelectionCriteria.SseKmsEncryptedObjects.Status),
			},
		}
	}
	if Rule.ExistingObjectReplication != nil {
		newRule.ExistingObjectReplication = &awss3.ExistingObjectReplication{
			Status: awss3.ExistingObjectReplicationStatus(Rule.ExistingObjectReplication.Status),
		}
	}
	if Rule.DeleteMarkerReplication != nil {
		newRule.DeleteMarkerReplication = &awss3.DeleteMarkerReplication{Status: awss3.DeleteMarkerReplicationStatus(Rule.DeleteMarkerReplication.Status)}
	}

	copyDestination(&Rule, &newRule)
	return newRule
}

// GenerateReplicationConfiguration is responsible for creating the Replication Configuration for requests.
func GenerateReplicationConfiguration(config *v1beta1.ReplicationConfiguration) *awss3.ReplicationConfiguration {
	source := &awss3.ReplicationConfiguration{
		Role:  config.Role,
		Rules: make([]awss3.ReplicationRule, len(config.Rules)),
	}

	for i, Rule := range config.Rules {
		source.Rules[i] = createRule(Rule)
	}
	return source
}

// GeneratePutBucketReplicationInput creates the input for the PutBucketReplication request for the S3 Client
func GeneratePutBucketReplicationInput(name string, config *v1beta1.ReplicationConfiguration) *awss3.PutBucketReplicationInput {
	return &awss3.PutBucketReplicationInput{
		Bucket:                   awsclient.String(name),
		ReplicationConfiguration: GenerateReplicationConfiguration(config),
	}
}

// CreateOrUpdate sends a request to have resource created on awsclient.
func (in *ReplicationConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.ReplicationConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketReplicationInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.ReplicationConfiguration)
	_, err := in.client.PutBucketReplicationRequest(input).Send(ctx)
	return awsclient.Wrap(err, replicationPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *ReplicationConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketReplicationRequest(
		&awss3.DeleteBucketReplicationInput{
			Bucket: awsclient.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return awsclient.Wrap(err, replicationDeleteFailed)
}

func sortReplicationRules(rules []awss3.ReplicationRule) {
	for i := range rules {
		if rules[i].Filter != nil && rules[i].Filter.And != nil {
			rules[i].Filter.And.Tags = s3.SortS3TagSet(rules[i].Filter.And.Tags)
		}
	}
}
