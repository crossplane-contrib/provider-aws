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

// TaggingConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type TaggingConfigurationClient struct {
	config *v1beta1.Tagging
}

// CreateTaggingConfigurationClient creates the client for CORS Configuration
func CreateTaggingConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &TaggingConfigurationClient{config: parameters.BucketTagging}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *TaggingConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketTaggingRequest(&awss3.GetBucketTaggingInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchTagSet" && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket tagging")
	}

	if in.config == nil && len(conf.TagSet) != 0 {
		return NeedsDeletion, nil
	}

	if cmp.Equal(conf.TagSet, in.generateTagging().TagSet) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

func (in *TaggingConfigurationClient) generateTagging() *awss3.Tagging {
	if in.config.TagSet == nil {
		return &awss3.Tagging{TagSet: make([]awss3.Tag, 0)}
	}
	conf := &awss3.Tagging{TagSet: make([]awss3.Tag, len(in.config.TagSet))}
	for i, v := range in.config.TagSet {
		conf.TagSet[i] = awss3.Tag{
			Key:   v.Key,
			Value: v.Value,
		}
	}
	return conf
}

func (in *TaggingConfigurationClient) generatePutBucketTagging(name string) *awss3.PutBucketTaggingInput {
	return &awss3.PutBucketTaggingInput{
		Bucket:  aws.String(name),
		Tagging: in.generateTagging(),
	}
}

// CreateResource sends a request to have resource created on AWS
func (in *TaggingConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketTaggingRequest(in.generatePutBucketTagging(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket tagging")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *TaggingConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketTaggingRequest(
		&awss3.DeleteBucketTaggingInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket tagging configuration")
	}
	return nil
}
