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

// ReplicationConfigurationClient is the client for API methods and reconciling the ReplicationConfiguration
type ReplicationConfigurationClient struct {
	config *v1beta1.ReplicationConfiguration
}

// CreateReplicationConfigurationClient creates the client for Replication Configuration
func CreateReplicationConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &ReplicationConfigurationClient{config: parameters.ReplicationConfiguration}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *ReplicationConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketReplicationRequest(&awss3.GetBucketReplicationInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "ReplicationConfigurationNotFoundError" && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get request payment configuration")
	}

	if conf.ReplicationConfiguration != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	source := in.GenerateConfiguration()

	if cmp.Equal(conf.ReplicationConfiguration, source) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

func copyDestintation(input *v1beta1.ReplicationRule, newRule *awss3.ReplicationRule) {
	Rule := input
	if Rule.Destination == nil {
		return
	}
	newRule.Destination = &awss3.Destination{
		AccessControlTranslation: nil,
		Account:                  Rule.Destination.Account,
		Bucket:                   Rule.Destination.Bucket,
		EncryptionConfiguration:  nil,
		Metrics:                  nil,
		ReplicationTime:          nil,
		StorageClass:             awss3.StorageClass(Rule.Destination.StorageClass),
	}
	if Rule.Destination.AccessControlTranslation != nil {
		newRule.Destination.AccessControlTranslation = &awss3.AccessControlTranslation{
			Owner: awss3.OwnerOverride(Rule.Destination.AccessControlTranslation.Owner),
		}
	}
	if Rule.Destination.EncryptionConfiguration != nil {
		newRule.Destination.EncryptionConfiguration = &awss3.EncryptionConfiguration{
			ReplicaKmsKeyID: Rule.Destination.EncryptionConfiguration.ReplicaKmsKeyID,
		}
	}
	if Rule.Destination.Metrics != nil {
		newRule.Destination.Metrics = &awss3.Metrics{
			EventThreshold: nil,
			Status:         awss3.MetricsStatus(Rule.Destination.Metrics.Status),
		}
		if Rule.Destination.Metrics.EventThreshold != nil {
			newRule.Destination.Metrics.EventThreshold = &awss3.ReplicationTimeValue{
				Minutes: Rule.Destination.Metrics.EventThreshold.Minutes,
			}
		}
	}
	if Rule.Destination.ReplicationTime != nil {
		newRule.Destination.ReplicationTime = &awss3.ReplicationTime{
			Status: awss3.ReplicationTimeStatus(Rule.Destination.ReplicationTime.Status),
			Time:   nil,
		}
		if Rule.Destination.ReplicationTime.Time != nil {
			newRule.Destination.ReplicationTime.Time = &awss3.ReplicationTimeValue{
				Minutes: Rule.Destination.ReplicationTime.Time.Minutes,
			}
		}
	}
}

func createRule(input v1beta1.ReplicationRule) awss3.ReplicationRule {
	Rule := input
	newRule := awss3.ReplicationRule{
		DeleteMarkerReplication:   nil,
		Destination:               nil,
		ExistingObjectReplication: nil,
		Filter:                    nil,
		ID:                        Rule.ID,
		Priority:                  Rule.Priority,
		SourceSelectionCriteria:   nil,
		Status:                    awss3.ReplicationRuleStatus(Rule.Status),
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
				Tags:   make([]awss3.Tag, len(Rule.Filter.And.Tags)),
			}
			for i, v := range Rule.Filter.And.Tags {
				newRule.Filter.And.Tags[i] = awss3.Tag{
					Key:   v.Key,
					Value: v.Value,
				}
			}
		}
		if Rule.Filter.Tag != nil {
			newRule.Filter.Tag = &awss3.Tag{
				Key:   Rule.Filter.Tag.Key,
				Value: Rule.Filter.Tag.Value,
			}
		}
	}
	if Rule.SourceSelectionCriteria != nil {
		newRule.SourceSelectionCriteria = &awss3.SourceSelectionCriteria{SseKmsEncryptedObjects: nil}
		if Rule.SourceSelectionCriteria.SseKmsEncryptedObjects != nil {
			newRule.SourceSelectionCriteria.SseKmsEncryptedObjects = &awss3.SseKmsEncryptedObjects{
				Status: awss3.SseKmsEncryptedObjectsStatus(Rule.SourceSelectionCriteria.SseKmsEncryptedObjects.Status),
			}
		}
	}
	if Rule.ExistingObjectReplication != nil {
		newRule.ExistingObjectReplication = &awss3.ExistingObjectReplication{
			Status: awss3.ExistingObjectReplicationStatus(Rule.ExistingObjectReplication.Status),
		}
	}
	if Rule.DeleteMarkerReplication != nil {
		newRule.DeleteMarkerReplication.Status = awss3.DeleteMarkerReplicationStatus(Rule.DeleteMarkerReplication.Status)
	}

	if Rule.DeleteMarkerReplication != nil {
		newRule.DeleteMarkerReplication.Status = awss3.DeleteMarkerReplicationStatus(Rule.DeleteMarkerReplication.Status)
	}
	copyDestintation(&Rule, &newRule)
	return newRule
}

// GenerateConfiguration is responsible for creating the Replication Configuration for requests.
func (in *ReplicationConfigurationClient) GenerateConfiguration() *awss3.ReplicationConfiguration {
	source := &awss3.ReplicationConfiguration{
		Role:  in.config.Role,
		Rules: make([]awss3.ReplicationRule, len(in.config.Rules)),
	}

	for i, Rule := range in.config.Rules {
		source.Rules[i] = createRule(Rule)
	}
	return source
}

// GeneratePutBucketReplicationInput creates the input for the PutBucketReplication request for the S3 Client
func (in *ReplicationConfigurationClient) GeneratePutBucketReplicationInput(name string) *awss3.PutBucketReplicationInput {
	return &awss3.PutBucketReplicationInput{
		Bucket:                   aws.String(name),
		ReplicationConfiguration: in.GenerateConfiguration(),
	}
}

// CreateResource sends a request to have resource created on AWS.
func (in *ReplicationConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketReplicationRequest(in.GeneratePutBucketReplicationInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket replication")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *ReplicationConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketReplicationRequest(
		&awss3.DeleteBucketReplicationInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket replication")
	}
	return nil
}
