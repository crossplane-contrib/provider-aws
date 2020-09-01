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
}

// CreateSSEConfigurationClient creates the client for Server Side Encryption Configuration
func CreateSSEConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &SSEConfigurationClient{config: parameters.ServerSideEncryptionConfiguration}
}

func (in *SSEConfigurationClient) sseNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "ServerSideEncryptionConfigurationNotFoundError" && in.config == nil {
		return true
	}
	return false
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *SSEConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	enc, err := client.GetBucketEncryptionRequest(&awss3.GetBucketEncryptionInput{Bucket: bucketName}).Send(ctx)
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
func (in *SSEConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketEncryptionRequest(in.GeneratePutBucketEncryptionInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket encryption")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *SSEConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketEncryptionRequest(
		&awss3.DeleteBucketEncryptionInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket encryption configuration")
	}
	return nil
}
