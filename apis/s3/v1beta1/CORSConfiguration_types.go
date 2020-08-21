package v1beta1

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// CORSConfiguration describes the cross-origin access configuration for objects
// in an Amazon S3 bucket. For more information, see Enabling Cross-Origin Resource Sharing
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) in the Amazon
// Simple Storage Service Developer Guide.
type CORSConfiguration struct {
	// A set of origins and methods (cross-origin access that you want to allow).
	// You can add up to 100 rules to the configuration.
	CORSRules []CORSRule `json:"corsRules"`
}

// CORSRule specifies a cross-origin access rule for an Amazon S3 bucket.
type CORSRule struct {
	// Headers that are specified in the Access-Control-Request-Headers header.
	// These headers are allowed in a preflight OPTIONS request. In response to
	// any preflight OPTIONS request, Amazon S3 returns any requested headers that
	// are allowed.
	// +optional
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`

	// An HTTP method that you allow the origin to execute. Valid values are GET,
	// PUT, HEAD, POST, and DELETE.
	AllowedMethods []string `json:"allowedMethods"`

	// One or more origins you want customers to be able to access the bucket from.
	AllowedOrigins []string `json:"allowedOrigins"`

	// One or more headers in the response that you want customers to be able to
	// access from their applications (for example, from a JavaScript XMLHttpRequest
	// object).
	// +optional
	ExposeHeaders []string `json:"exposeHeaders,omitempty"`

	// The time in seconds that your browser is to cache the preflight response
	// for the specified resource.
	// +optional
	MaxAgeSeconds *int64 `json:"maxAgeSeconds,omitempty"`
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *CORSConfiguration) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (managed.ExternalObservation, error) {
	conf, err := client.GetBucketCorsRequest(&awss3.GetBucketCorsInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get bucket encryption")
	}

	if len(conf.CORSRules) != len(in.CORSRules) {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
	}

	for i, Rule := range in.CORSRules {
		outputRule := conf.CORSRules[i]
		if !cmp.Equal(Rule.AllowedHeaders, outputRule.AllowedHeaders) {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if !cmp.Equal(Rule.AllowedMethods, outputRule.AllowedMethods) {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if !cmp.Equal(Rule.AllowedOrigins, outputRule.AllowedOrigins) {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if !cmp.Equal(Rule.ExposeHeaders, outputRule.ExposeHeaders) {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if !cmp.Equal(Rule.MaxAgeSeconds, outputRule.MaxAgeSeconds) {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
	}

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

// GeneratePutBucketCorsInput creates the input for the PutBucketCors request for the S3 Client
func (in *CORSConfiguration) GeneratePutBucketCorsInput(name string) *awss3.PutBucketCorsInput {
	bci := &awss3.PutBucketCorsInput{
		Bucket:            aws.String(name),
		CORSConfiguration: &awss3.CORSConfiguration{},
	}
	for _, cors := range in.CORSRules {
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
