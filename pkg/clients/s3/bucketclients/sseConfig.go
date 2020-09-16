package bucketclients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// SSEConfigurationClient is the client for API methods and reconciling the ServerSideEncryptionConfiguration
type SSEConfigurationClient struct {
	config *v1beta1.ServerSideEncryptionConfiguration
	bucket *v1beta1.Bucket
	client s3.BucketClient
}

// CreateSSEConfigurationClient creates the client for Server Side Encryption Configuration
func CreateSSEConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *SSEConfigurationClient {
	return &SSEConfigurationClient{config: bucket.Spec.Parameters.ServerSideEncryptionConfiguration, bucket: bucket, client: client}
}

func (in *SSEConfigurationClient) sseNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "ServerSideEncryptionConfigurationNotFoundError" && in.config == nil {
		return true
	}
	return false
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *SSEConfigurationClient) ExistsAndUpdated(ctx context.Context) (ResourceStatus, error) {
	enc, err := in.client.GetBucketEncryptionRequest(&awss3.GetBucketEncryptionInput{Bucket: aws.String(meta.GetExternalName(in.bucket))}).Send(ctx)
	if err != nil && in.sseNotFound(err) {
		return Updated, nil
	} else if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket encryption")
	}

	if enc.ServerSideEncryptionConfiguration != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	if len(enc.ServerSideEncryptionConfiguration.Rules) != len(in.config.Rules) {
		return NeedsUpdate, nil
	}

	for i, Rule := range in.config.Rules {
		outputRule := enc.ServerSideEncryptionConfiguration.Rules[i].ApplyServerSideEncryptionByDefault
		if outputRule.KMSMasterKeyID != Rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID {
			return NeedsUpdate, nil
		}
		if string(outputRule.SSEAlgorithm) != Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm {
			return NeedsUpdate, nil
		}
	}

	return Updated, nil
}

// GeneratePutBucketEncryptionInput creates the input for the PutBucketEncryption request for the S3 Client
func (in *SSEConfigurationClient) GeneratePutBucketEncryptionInput(name string) *awss3.PutBucketEncryptionInput {
	bei := &awss3.PutBucketEncryptionInput{
		Bucket:                            aws.String(name),
		ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{},
	}
	for _, rule := range in.config.Rules {
		bei.ServerSideEncryptionConfiguration.Rules = append(bei.ServerSideEncryptionConfiguration.Rules, awss3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
				KMSMasterKeyID: rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				SSEAlgorithm:   awss3.ServerSideEncryption(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			},
		})
	}
	return bei
}

// CreateResource sends a request to have resource created on AWS.
func (in *SSEConfigurationClient) CreateResource(ctx context.Context) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketEncryptionRequest(in.GeneratePutBucketEncryptionInput(meta.GetExternalName(in.bucket))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket encryption")
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *SSEConfigurationClient) DeleteResource(ctx context.Context) error {
	_, err := in.client.DeleteBucketEncryptionRequest(
		&awss3.DeleteBucketEncryptionInput{
			Bucket: aws.String(meta.GetExternalName(in.bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket encryption configuration")
}
