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

package bucketresources

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

var _ BucketResource = &WebsiteConfigurationClient{}

// WebsiteConfigurationClient is the client for API methods and reconciling the WebsiteConfiguration
type WebsiteConfigurationClient struct {
	config *v1beta1.WebsiteConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *WebsiteConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	// GetBucketWebsiteRequest throws an error if nothing exists externally
	// Future work can be done to support brownfield initialization for the WebsiteConfiguration
	// TODO
	return nil
}

// NewWebsiteConfigurationClient creates the client for Website Configuration
func NewWebsiteConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *WebsiteConfigurationClient {
	return &WebsiteConfigurationClient{config: bucket.Spec.ForProvider.WebsiteConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *WebsiteConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketWebsiteRequest(&awss3.GetBucketWebsiteInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		if s3.WebsiteConfigurationNotFound(err) && in.config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, "cannot get request bucket website configuration")
	}

	if conf.GetBucketWebsiteOutput != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	source := GenerateWebsiteConfiguration(in)
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

// GenerateWebsiteConfiguration is responsible for creating the Website Configuration for requests.
func GenerateWebsiteConfiguration(in *WebsiteConfigurationClient) *awss3.WebsiteConfiguration {
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
func GeneratePutBucketWebsiteInput(name string, in *WebsiteConfigurationClient) *awss3.PutBucketWebsiteInput {
	wi := &awss3.PutBucketWebsiteInput{
		Bucket:               aws.String(name),
		WebsiteConfiguration: GenerateWebsiteConfiguration(in),
	}
	return wi
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *WebsiteConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketWebsiteRequest(GeneratePutBucketWebsiteInput(meta.GetExternalName(bucket), in)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket website")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *WebsiteConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketWebsiteRequest(
		&awss3.DeleteBucketWebsiteInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket website configuration")
}
