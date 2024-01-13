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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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

// NewWebsiteConfigurationClient creates the client for Website Configuration
func NewWebsiteConfigurationClient(client s3.BucketClient) *WebsiteConfigurationClient {
	return &WebsiteConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *WebsiteConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { //nolint:gocyclo
	external, err := in.client.GetBucketWebsite(ctx, &awss3.GetBucketWebsiteInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	config := bucket.Spec.ForProvider.WebsiteConfiguration
	if err != nil {
		if s3.WebsiteConfigurationNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.WebsiteConfigurationNotFound, err), websiteGetFailed)
	}

	switch {
	case external.RoutingRules == nil && external.RedirectAllRequestsTo == nil && external.IndexDocument == nil && external.ErrorDocument == nil && config == nil:
		return Updated, nil
	case external != nil && config == nil:
		return NeedsDeletion, nil
	}

	source := GenerateWebsiteConfiguration(config)
	confBody := &types.WebsiteConfiguration{
		ErrorDocument:         external.ErrorDocument,
		IndexDocument:         external.IndexDocument,
		RedirectAllRequestsTo: external.RedirectAllRequestsTo,
		RoutingRules:          external.RoutingRules,
	}

	if cmp.Equal(confBody, source, cmpopts.IgnoreTypes(document.NoSerde{})) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

// CreateOrUpdate sends a request to have resource created on awsclient.
func (in *WebsiteConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.WebsiteConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketWebsiteInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.WebsiteConfiguration)
	_, err := in.client.PutBucketWebsite(ctx, input)
	return errorutils.Wrap(err, websitePutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *WebsiteConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketWebsite(ctx,
		&awss3.DeleteBucketWebsiteInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, websiteDeleteFailed)
}

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (in *WebsiteConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketWebsite(ctx, &awss3.GetBucketWebsiteInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.WebsiteConfigurationNotFound, err), websiteGetFailed)
	}

	if external == nil || (len(external.RoutingRules) == 0 &&
		external.ErrorDocument == nil &&
		external.IndexDocument == nil &&
		external.RedirectAllRequestsTo == nil) {
		return nil
	}

	if bucket.Spec.ForProvider.WebsiteConfiguration == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.WebsiteConfiguration = &v1beta1.WebsiteConfiguration{}
	}

	createWebsiteConfigFromExternal(external, bucket.Spec.ForProvider.WebsiteConfiguration)
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *WebsiteConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.WebsiteConfiguration != nil
}

// GenerateWebsiteConfiguration is responsible for creating the Website Configuration for requests.
func GenerateWebsiteConfiguration(config *v1beta1.WebsiteConfiguration) *types.WebsiteConfiguration {
	wi := &types.WebsiteConfiguration{}
	if config.ErrorDocument != nil {
		wi.ErrorDocument = &types.ErrorDocument{Key: pointer.ToOrNilIfZeroValue(config.ErrorDocument.Key)}
	}
	if config.IndexDocument != nil {
		wi.IndexDocument = &types.IndexDocument{Suffix: pointer.ToOrNilIfZeroValue(config.IndexDocument.Suffix)}
	}
	if config.RedirectAllRequestsTo != nil {
		wi.RedirectAllRequestsTo = &types.RedirectAllRequestsTo{
			HostName: pointer.ToOrNilIfZeroValue(config.RedirectAllRequestsTo.HostName),
			Protocol: types.Protocol(config.RedirectAllRequestsTo.Protocol),
		}
	}
	if len(config.RoutingRules) > 0 {
		wi.RoutingRules = make([]types.RoutingRule, len(config.RoutingRules))
		for i, rule := range config.RoutingRules {
			rr := types.RoutingRule{
				Redirect: &types.Redirect{
					HostName:             rule.Redirect.HostName,
					HttpRedirectCode:     rule.Redirect.HTTPRedirectCode,
					Protocol:             types.Protocol(rule.Redirect.Protocol),
					ReplaceKeyPrefixWith: rule.Redirect.ReplaceKeyPrefixWith,
					ReplaceKeyWith:       rule.Redirect.ReplaceKeyWith,
				},
			}
			if rule.Condition != nil {
				rr.Condition = &types.Condition{
					HttpErrorCodeReturnedEquals: rule.Condition.HTTPErrorCodeReturnedEquals,
					KeyPrefixEquals:             rule.Condition.KeyPrefixEquals,
				}
				wi.RoutingRules[i] = rr
			}
		}
	}

	return wi
}

// GeneratePutBucketWebsiteInput creates the input for the PutBucketWebsite request for the S3 Client
func GeneratePutBucketWebsiteInput(name string, config *v1beta1.WebsiteConfiguration) *awss3.PutBucketWebsiteInput {
	wi := &awss3.PutBucketWebsiteInput{
		Bucket:               pointer.ToOrNilIfZeroValue(name),
		WebsiteConfiguration: GenerateWebsiteConfiguration(config),
	}
	return wi
}

func createWebsiteConfigFromExternal(external *awss3.GetBucketWebsiteOutput, config *v1beta1.WebsiteConfiguration) { //nolint:gocyclo
	if external.ErrorDocument != nil {
		if config.ErrorDocument == nil {
			config.ErrorDocument = &v1beta1.ErrorDocument{}
		}
		config.ErrorDocument.Key = pointer.LateInitializeValueFromPtr(config.ErrorDocument.Key, external.ErrorDocument.Key)
	}
	if external.IndexDocument != nil {
		if config.IndexDocument == nil {
			config.IndexDocument = &v1beta1.IndexDocument{}
		}
		config.IndexDocument.Suffix = pointer.LateInitializeValueFromPtr(config.IndexDocument.Suffix, external.IndexDocument.Suffix)
	}
	if external.RedirectAllRequestsTo != nil {
		if config.RedirectAllRequestsTo == nil {
			config.RedirectAllRequestsTo = &v1beta1.RedirectAllRequestsTo{}
		}
		if external.RedirectAllRequestsTo.Protocol != "" {
			config.RedirectAllRequestsTo.Protocol = pointer.LateInitializeValueFromPtr(
				config.RedirectAllRequestsTo.Protocol,
				pointer.ToOrNilIfZeroValue(string(external.RedirectAllRequestsTo.Protocol)),
			)
		}
		config.RedirectAllRequestsTo.HostName = pointer.LateInitializeValueFromPtr(
			config.RedirectAllRequestsTo.HostName,
			external.RedirectAllRequestsTo.HostName,
		)
	}
	if len(external.RoutingRules) != 0 && config.RoutingRules == nil {
		config.RoutingRules = make([]v1beta1.RoutingRule, len(external.RoutingRules))
		for i, rr := range external.RoutingRules {
			if rr.Redirect != nil {
				config.RoutingRules[i].Redirect.HostName = pointer.LateInitialize(
					config.RoutingRules[i].Redirect.HostName,
					rr.Redirect.HostName,
				)
				config.RoutingRules[i].Redirect.HTTPRedirectCode = pointer.LateInitialize(
					config.RoutingRules[i].Redirect.HTTPRedirectCode,
					rr.Redirect.HttpRedirectCode,
				)
				config.RoutingRules[i].Redirect.ReplaceKeyPrefixWith = pointer.LateInitialize(
					config.RoutingRules[i].Redirect.ReplaceKeyPrefixWith,
					rr.Redirect.ReplaceKeyPrefixWith,
				)
				config.RoutingRules[i].Redirect.ReplaceKeyWith = pointer.LateInitialize(
					config.RoutingRules[i].Redirect.ReplaceKeyWith,
					rr.Redirect.ReplaceKeyWith,
				)
				if rr.Redirect.Protocol != "" {
					config.RoutingRules[i].Redirect.Protocol = pointer.LateInitializeValueFromPtr(
						config.RoutingRules[i].Redirect.Protocol,
						pointer.ToOrNilIfZeroValue(string(rr.Redirect.Protocol)),
					)
				}
			}
			if rr.Condition != nil {
				if config.RoutingRules[i].Condition == nil {
					config.RoutingRules[i].Condition = &v1beta1.Condition{}
				}
				config.RoutingRules[i].Condition.HTTPErrorCodeReturnedEquals = pointer.LateInitialize(
					config.RoutingRules[i].Condition.HTTPErrorCodeReturnedEquals,
					rr.Condition.HttpErrorCodeReturnedEquals,
				)
				config.RoutingRules[i].Condition.KeyPrefixEquals = pointer.LateInitialize(
					config.RoutingRules[i].Condition.KeyPrefixEquals,
					rr.Condition.KeyPrefixEquals,
				)
			}
		}
	}
}
