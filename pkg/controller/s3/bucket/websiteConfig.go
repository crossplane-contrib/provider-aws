/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bucket

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	websiteGetFailed    = "cannot get Bucket website configuration"
	websitePutFailed    = "cannot put Bucket website configuration"
	websiteDeleteFailed = "cannot delete Bucket website configuration"
)

// WebsiteConfigurationClient is the client for API methods and reconciling the WebsiteConfiguration
type WebsiteConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (*WebsiteConfigurationClient) LateInitialize(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// NewWebsiteConfigurationClient creates the client for Website Configuration
func NewWebsiteConfigurationClient(client s3.BucketClient) *WebsiteConfigurationClient {
	return &WebsiteConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *WebsiteConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { // nolint:gocyclo
	external, err := in.client.GetBucketWebsiteRequest(&awss3.GetBucketWebsiteInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	config := bucket.Spec.ForProvider.WebsiteConfiguration
	if err != nil {
		if s3.WebsiteConfigurationNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, websiteGetFailed)
	}

	switch {
	case external.RoutingRules == nil && external.RedirectAllRequestsTo == nil && external.IndexDocument == nil && external.ErrorDocument == nil && config == nil:
		return Updated, nil
	case external.GetBucketWebsiteOutput != nil && config == nil:
		return NeedsDeletion, nil
	}

	source := GenerateWebsiteConfiguration(config)
	confBody := &awss3.WebsiteConfiguration{
		ErrorDocument:         external.ErrorDocument,
		IndexDocument:         external.IndexDocument,
		RedirectAllRequestsTo: external.RedirectAllRequestsTo,
		RoutingRules:          external.RoutingRules,
	}

	if cmp.Equal(confBody, source) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

// GenerateWebsiteConfiguration is responsible for creating the Website Configuration for requests.
func GenerateWebsiteConfiguration(config *v1beta1.WebsiteConfiguration) *awss3.WebsiteConfiguration {
	wi := &awss3.WebsiteConfiguration{}
	if config.ErrorDocument != nil {
		wi.ErrorDocument = &awss3.ErrorDocument{Key: aws.String(config.ErrorDocument.Key)}
	}
	if config.IndexDocument != nil {
		wi.IndexDocument = &awss3.IndexDocument{Suffix: aws.String(config.IndexDocument.Suffix)}
	}
	if config.RedirectAllRequestsTo != nil {
		wi.RedirectAllRequestsTo = &awss3.RedirectAllRequestsTo{
			HostName: aws.String(config.RedirectAllRequestsTo.HostName),
			Protocol: awss3.Protocol(config.RedirectAllRequestsTo.Protocol),
		}
	}
	wi.RoutingRules = make([]awss3.RoutingRule, len(config.RoutingRules))
	for i, rule := range config.RoutingRules {
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
func GeneratePutBucketWebsiteInput(name string, config *v1beta1.WebsiteConfiguration) *awss3.PutBucketWebsiteInput {
	wi := &awss3.PutBucketWebsiteInput{
		Bucket:               aws.String(name),
		WebsiteConfiguration: GenerateWebsiteConfiguration(config),
	}
	return wi
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *WebsiteConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.WebsiteConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketWebsiteInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.WebsiteConfiguration)
	_, err := in.client.PutBucketWebsiteRequest(input).Send(ctx)
	return errors.Wrap(err, websitePutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *WebsiteConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketWebsiteRequest(
		&awss3.DeleteBucketWebsiteInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, websiteDeleteFailed)
}
