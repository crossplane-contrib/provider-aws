package bucketclients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// WebsiteConfigurationClient is the client for API methods and reconciling the WebsiteConfiguration
type WebsiteConfigurationClient struct {
	config *v1beta1.WebsiteConfiguration
}

// CreateWebsiteConfigurationClient creates the client for Website Configuration
func CreateWebsiteConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &WebsiteConfigurationClient{config: parameters.WebsiteConfiguration}
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *WebsiteConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketWebsiteRequest(&awss3.GetBucketWebsiteInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchWebsiteConfiguration" && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get request bucket website configuration")
	}

	if conf.GetBucketWebsiteOutput != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	source := in.GenerateConfiguration()
	confBody := &awss3.WebsiteConfiguration{
		ErrorDocument:         conf.ErrorDocument,
		IndexDocument:         conf.IndexDocument,
		RedirectAllRequestsTo: conf.RedirectAllRequestsTo,
		RoutingRules:          conf.RoutingRules,
	}

	if cmp.Equal(confBody, source) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

// GenerateConfiguration is responsible for creating the Website Configuration for requests.
func (in *WebsiteConfigurationClient) GenerateConfiguration() *awss3.WebsiteConfiguration {
	wi := &awss3.WebsiteConfiguration{}
	if in.config.ErrorDocument != nil {
		wi.ErrorDocument = &awss3.ErrorDocument{Key: aws.String(in.config.ErrorDocument.Key)}
	}
	if in.config.IndexDocument != nil {
		wi.IndexDocument = &awss3.IndexDocument{Suffix: aws.String(in.config.IndexDocument.Suffix)}
	}
	if in.config.RedirectAllRequestsTo != nil {
		wi.RedirectAllRequestsTo = &awss3.RedirectAllRequestsTo{
			HostName: aws.String(in.config.RedirectAllRequestsTo.HostName),
			Protocol: awss3.Protocol(in.config.RedirectAllRequestsTo.Protocol),
		}
	}
	wi.RoutingRules = make([]awss3.RoutingRule, len(in.config.RoutingRules))
	for i, rule := range in.config.RoutingRules {
		rr := awss3.RoutingRule{
			Redirect: &awss3.Redirect{
				HostName:             rule.Redirect.HostName,
				HttpRedirectCode:     rule.Redirect.HTTPRedirectCode,
				Protocol:             awss3.Protocol(rule.Redirect.Protocol),
				ReplaceKeyPrefixWith: rule.Redirect.ReplaceKeyPrefixWith,
				ReplaceKeyWith:       rule.Redirect.ReplaceKeyWith,
			},
		}
		if rule.Condition != nil {
			rr.Condition = &awss3.Condition{
				HttpErrorCodeReturnedEquals: rule.Condition.HTTPErrorCodeReturnedEquals,
				KeyPrefixEquals:             rule.Condition.KeyPrefixEquals,
			}
		}
		wi.RoutingRules[i] = rr
	}
	return wi
}

// GeneratePutBucketWebsiteInput creates the input for the PutBucketWebsite request for the S3 Client
func (in *WebsiteConfigurationClient) GeneratePutBucketWebsiteInput(name string) *awss3.PutBucketWebsiteInput {
	wi := &awss3.PutBucketWebsiteInput{
		Bucket:               aws.String(name),
		WebsiteConfiguration: in.GenerateConfiguration(),
	}
	return wi
}

// CreateResource sends a request to have resource created on AWS.
func (in *WebsiteConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		if _, err := client.PutBucketWebsiteRequest(in.GeneratePutBucketWebsiteInput(meta.GetExternalName(cr))).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket website")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *WebsiteConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	_, err := client.DeleteBucketWebsiteRequest(
		&awss3.DeleteBucketWebsiteInput{
			Bucket: aws.String(meta.GetExternalName(cr)),
		},
	).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot delete bucket website configuration")
	}
	return nil
}
