package v1beta1

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// ServerSideEncryptionConfiguration specifies the default server-side-encryption configuration.
type ServerSideEncryptionConfiguration struct {
	// Container for information about a particular server-side encryption configuration
	// rule.
	Rules []ServerSideEncryptionRule `json:"rules"`
}

// ServerSideEncryptionRule Specifies the default server-side encryption configuration.
type ServerSideEncryptionRule struct {
	// Specifies the default server-side encryption to apply to new objects in the
	// bucket. If a PUT Object request doesn't specify any server-side encryption,
	// this default encryption will be applied.
	ApplyServerSideEncryptionByDefault ServerSideEncryptionByDefault `json:"applyServerSideEncryptionByDefault"`
}

// ServerSideEncryptionByDefault describes the default server-side encryption to
// apply to new objects in the bucket. If a PUT Object request doesn't specify
// any server-side encryption, this default encryption will be applied.
type ServerSideEncryptionByDefault struct {
	// AWS Key Management Service (KMS) customer master key ID to use for the default
	// encryption. This parameter is allowed if and only if SSEAlgorithm is set
	// to aws:kms.
	//
	// You can specify the key ID or the Amazon Resource Name (ARN) of the CMK.
	// However, if you are using encryption with cross-account operations, you must
	// use a fully qualified CMK ARN. For more information, see Using encryption
	// for cross-account operations (https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html#bucket-encryption-update-bucket-policy).
	//
	// For example:
	//
	//    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	//
	//    * Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	//
	// Amazon S3 only supports symmetric CMKs and not asymmetric CMKs. For more
	// information, see Using Symmetric and Asymmetric Keys (https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html)
	// in the AWS Key Management Service Developer Guide.
	// +optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// NOTE(muvaf): aws:kms is not accepted by kubebuilder enum.

	// Server-side encryption algorithm to use for the default encryption.
	// Options are AES256 or aws:kms
	SSEAlgorithm string `json:"sseAlgorithm"`
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (sse *ServerSideEncryptionConfiguration) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (managed.ExternalObservation, error) {
	enc, err := client.GetBucketEncryptionRequest(&awss3.GetBucketEncryptionInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get bucket encryption")
	}

	if len(enc.ServerSideEncryptionConfiguration.Rules) != len(sse.Rules) {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
	}

	for i, Rule := range sse.Rules {
		outputRule := enc.ServerSideEncryptionConfiguration.Rules[i].ApplyServerSideEncryptionByDefault
		if outputRule.KMSMasterKeyID != Rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if string(outputRule.SSEAlgorithm) != Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
	}

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

// GeneratePutBucketEncryptionInput creates the input for PutBucketEncryption for the S3 Client
func (sse *ServerSideEncryptionConfiguration) GeneratePutBucketEncryptionInput(name string) *awss3.PutBucketEncryptionInput {
	bei := &awss3.PutBucketEncryptionInput{
		Bucket:                            aws.String(name),
		ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{},
	}
	for _, rule := range sse.Rules {
		bei.ServerSideEncryptionConfiguration.Rules = append(bei.ServerSideEncryptionConfiguration.Rules, awss3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
				KMSMasterKeyID: rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				SSEAlgorithm:   awss3.ServerSideEncryption(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			},
		})
	}
	return bei
}
