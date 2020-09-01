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

// CORSConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type CORSConfigurationClient struct {
	config *v1beta1.CORSConfiguration
}

// CreateCORSConfigurationClient creates the client for CORS Configuration
func CreateCORSConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &CORSConfigurationClient{config: parameters.CORSConfiguration}
}

func (in *CORSConfigurationClient) corsNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchCORSConfiguration" && in.config == nil {
		return true
	}
	return false
}

func compareCORS(local *v1beta1.CORSConfiguration, external *awss3.GetBucketCorsResponse) ResourceStatus {
	if local != nil && external == nil {
		return NeedsDeletion
	}

	if len(local.CORSRules) != len(external.CORSRules) {
		return NeedsUpdate
	}

	for i, Rule := range local.CORSRules {
		outputRule := external.CORSRules[i]
		if !cmp.Equal(Rule.AllowedHeaders, outputRule.AllowedHeaders) {
			return NeedsUpdate
		}
		if !cmp.Equal(Rule.AllowedMethods, outputRule.AllowedMethods) {
			return NeedsUpdate
		}
		if !cmp.Equal(Rule.AllowedOrigins, outputRule.AllowedOrigins) {
			return NeedsUpdate
		}
		if !cmp.Equal(Rule.ExposeHeaders, outputRule.ExposeHeaders) {
			return NeedsUpdate
		}
		if !cmp.Equal(Rule.MaxAgeSeconds, outputRule.MaxAgeSeconds) {
			return NeedsUpdate
		}
	}
	return Updated
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *CORSConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketCorsRequest(&awss3.GetBucketCorsInput{Bucket: bucketName}).Send(ctx)
	if err != nil && in.corsNotFound(err) {
		return Updated, nil
	} else if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket encryption")
	}

	return compareCORS(in.config, conf), nil
}

// GeneratePutBucketCorsInput creates the input for the PutBucketCors request for the S3 Client
func (in *CORSConfigurationClient) GeneratePutBucketCorsInput(name string) *awss3.PutBucketCorsInput {
	bci := &awss3.PutBucketCorsInput{
		Bucket:            aws.String(name),
		CORSConfiguration: &awss3.CORSConfiguration{},
	}
	for _, cors := range in.config.CORSRules {
		bci.CORSConfiguration.CORSRules = append(bci.CORSConfiguration.CORSRules, awss3.CORSRule{
			AllowedHeaders: cors.AllowedHeaders,
			AllowedMethods: cors.AllowedMethods,
			AllowedOrigins: cors.AllowedOrigins,
			ExposeHeaders:  cors.ExposeHeaders,
			MaxAgeSeconds:  cors.MaxAgeSeconds,
		})
	}
	return bci
}

// CreateResource sends a request to have resource created on AWS
func (in *CORSConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketCorsRequest(in.GeneratePutBucketCorsInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket cors")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *CORSConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketCorsRequest(
		&awss3.DeleteBucketCorsInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket CORS configuration")
	}
	return nil
}
