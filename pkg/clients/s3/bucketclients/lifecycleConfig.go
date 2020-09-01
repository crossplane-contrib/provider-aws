package bucketclients

import (
	"context"
	"time"

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
}

// CreateLifecycleConfigurationClient creates the client for Accelerate Configuration
func CreateLifecycleConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &LifecycleConfigurationClient{config: parameters.LifecycleConfiguration}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *LifecycleConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketLifecycleConfigurationRequest(&awss3.GetBucketLifecycleConfigurationInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchLifecycleConfiguration" && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket lifecycle")
	}

	if len(conf.Rules) != 0 && in.config == nil {
		return NeedsDeletion, nil
	}

	rules, err := in.generateConfiguration()
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "unable to create rules for bucket lifecycle reconcile")
	}
	if cmp.Equal(conf.Rules, rules) {
		return Updated, nil
	}
	return NeedsUpdate, nil
}

func (in *LifecycleConfigurationClient) createLifecycleRule(other v1beta1.LifecycleRule) (*awss3.LifecycleRule, error) {
	rule := awss3.LifecycleRule{
		ID:     other.ID,
		Status: awss3.ExpirationStatus(other.Status),
	}
	if other.AbortIncompleteMultipartUpload != nil {
		rule.AbortIncompleteMultipartUpload = &awss3.AbortIncompleteMultipartUpload{
			DaysAfterInitiation: other.AbortIncompleteMultipartUpload.DaysAfterInitiation,
		}
	}
	if other.Expiration != nil {
		date, err := time.Parse("2006-01-02T15:04:05.000Z", aws.StringValue(other.Expiration.Date))
		if err != nil {
			return nil, err
		}
		rule.Expiration = &awss3.LifecycleExpiration{
			Date:                      &date,
			Days:                      other.Expiration.Days,
			ExpiredObjectDeleteMarker: other.Expiration.ExpiredObjectDeleteMarker,
		}
	}
	if other.NoncurrentVersionExpiration != nil {
		rule.NoncurrentVersionExpiration = &awss3.NoncurrentVersionExpiration{NoncurrentDays: other.NoncurrentVersionExpiration.NoncurrentDays}
	}
	if other.NoncurrentVersionTransitions != nil {
		rule.NoncurrentVersionTransitions = make([]awss3.NoncurrentVersionTransition, len(other.NoncurrentVersionTransitions))
		for tIndex, transition := range other.NoncurrentVersionTransitions {
			rule.NoncurrentVersionTransitions[tIndex] = awss3.NoncurrentVersionTransition{
				NoncurrentDays: transition.NoncurrentDays,
				StorageClass:   awss3.TransitionStorageClass(transition.StorageClass),
			}
		}
	}
	if other.Transitions != nil {
		rule.Transitions = make([]awss3.Transition, len(other.Transitions))
		for tIndex, transition := range other.Transitions {
			date, err := time.Parse("2006-01-02T15:04:05.000Z", aws.StringValue(transition.Date))
			if err != nil {
				return nil, err
			}
			rule.Transitions[tIndex] = awss3.Transition{
				Date:         &date,
				Days:         transition.Days,
				StorageClass: awss3.TransitionStorageClass(transition.StorageClass),
			}
		}
	}
	return &rule, nil
}

func (in *LifecycleConfigurationClient) generateConfiguration() ([]awss3.LifecycleRule, error) {
	rules := make([]awss3.LifecycleRule, len(in.config.Rules))
	for i, v := range in.config.Rules {
		rule, err := in.createLifecycleRule(v)
		if err != nil {
			return nil, err
		}
		rules[i] = *rule
	}
	return rules, nil
}

// GenerateLifecycleConfigurationInput creates the input for the LifecycleConfiguration request for the S3 Client
func (in *LifecycleConfigurationClient) GenerateLifecycleConfigurationInput(name string) (*awss3.PutBucketLifecycleConfigurationInput, error) {
	rules, err := in.generateConfiguration()
	if err != nil {
		return nil, err
	}
	return &awss3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(name),
		LifecycleConfiguration: &awss3.BucketLifecycleConfiguration{Rules: rules},
	}, nil
}

// CreateResource sends a request to have resource created on AWS
func (in *LifecycleConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		input, err := in.GenerateLifecycleConfigurationInput(meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "unable to create input for bucket lifecycle request")
		}
		if _, err := client.PutBucketLifecycleConfigurationRequest(input).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket lifecycle")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *LifecycleConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketLifecycleRequest(
		&awss3.DeleteBucketLifecycleInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket lifecycle configuration")
	}
	return nil
}
