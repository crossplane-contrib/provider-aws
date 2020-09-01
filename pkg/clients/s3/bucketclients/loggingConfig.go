package bucketclients

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// LoggingConfigurationClient is the client for API methods and reconciling the LoggingConfiguration
type LoggingConfigurationClient struct {
	config *v1beta1.LoggingConfiguration
}

// CreateLoggingConfigurationClient creates the client for Logging Configuration
func CreateLoggingConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &LoggingConfigurationClient{config: parameters.LoggingConfiguration}
}

// CompareStrings compares pairs of strings passed in
func CompareStrings(strings ...*string) bool {
	if len(strings)%2 != 0 {
		return false
	}
	for i := 0; i < len(strings); i += 2 {
		if aws.StringValue(strings[i]) != aws.StringValue(strings[i+1]) {
			return false
		}
	}
	return true
}

func compareLogging(local *v1beta1.LoggingConfiguration, external *awss3.LoggingEnabled) ResourceStatus {
	if aws.StringValue(external.TargetBucket) != aws.StringValue(local.TargetBucket) {
		return NeedsUpdate
	}

	if aws.StringValue(external.TargetPrefix) != aws.StringValue(local.TargetPrefix) {
		return NeedsUpdate
	}

	for i, grant := range local.TargetGrants {
		outputGrant := external.TargetGrants[i]
		if outputGrant.Grantee != nil {
			oGrant := outputGrant.Grantee
			lGrant := grant.Grantee
			if !CompareStrings(oGrant.DisplayName, lGrant.DisplayName,
				oGrant.EmailAddress, lGrant.EmailAddress,
				oGrant.ID, lGrant.ID,
				oGrant.URI, lGrant.URI) {
				return NeedsUpdate
			}
			if string(oGrant.Type) != lGrant.Type {
				return NeedsUpdate
			}
		}
		if string(outputGrant.Permission) != grant.Permission {
			return NeedsUpdate
		}
	}

	return Updated
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *LoggingConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket encryption")
	}

	if conf.LoggingEnabled == nil && in.config == nil {
		return Updated, nil
	} else if conf.LoggingEnabled != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	return compareLogging(in.config, conf.LoggingEnabled), nil
}

// GeneratePutBucketLoggingInput creates the input for the PutBucketLogging request for the S3 Client
func (in *LoggingConfigurationClient) GeneratePutBucketLoggingInput(name string) *awss3.PutBucketLoggingInput {
	bci := &awss3.PutBucketLoggingInput{
		Bucket: aws.String(name),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{LoggingEnabled: &awss3.LoggingEnabled{
			TargetBucket: in.config.TargetBucket,
			TargetGrants: make([]awss3.TargetGrant, 0),
			TargetPrefix: in.config.TargetPrefix,
		}},
	}
	for _, grant := range in.config.TargetGrants {
		bci.BucketLoggingStatus.LoggingEnabled.TargetGrants = append(bci.BucketLoggingStatus.LoggingEnabled.TargetGrants, awss3.TargetGrant{
			Grantee: &awss3.Grantee{
				DisplayName:  grant.Grantee.DisplayName,
				EmailAddress: grant.Grantee.EmailAddress,
				ID:           grant.Grantee.ID,
				Type:         awss3.Type(grant.Grantee.Type),
				URI:          grant.Grantee.URI,
			},
			Permission: awss3.BucketLogsPermission(grant.Permission),
		})
	}
	return bci
}

// CreateResource sends a request to have resource created on AWS
func (in *LoggingConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketLoggingRequest(in.GeneratePutBucketLoggingInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket logging")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *LoggingConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	input := &awss3.PutBucketLoggingInput{
		Bucket:              aws.String(meta.GetExternalName(cr)),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{}, //  Empty BucketLoggingStatus disables logging
	}
	if _, err := client.PutBucketLoggingRequest(input).Send(ctx); err != nil {
		return errors.Wrap(err, "cannot delete bucket logging")
	}
	return nil
}
