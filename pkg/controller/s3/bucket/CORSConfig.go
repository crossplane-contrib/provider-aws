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
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	corsGetFailed    = "cannot get Bucket CORS configuration"
	corsPutFailed    = "cannot put Bucket cors"
	corsDeleteFailed = "cannot delete Bucket CORS configuration"
)

// CORSConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type CORSConfigurationClient struct {
	client s3.BucketClient
}

// NewCORSConfigurationClient creates the client for CORS Configuration
func NewCORSConfigurationClient(client s3.BucketClient) *CORSConfigurationClient {
	return &CORSConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *CORSConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	result, err := in.client.GetBucketCors(ctx, &awss3.GetBucketCorsInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if resource.Ignore(s3.CORSConfigurationNotFound, err) != nil {
		return NeedsUpdate, errorutils.Wrap(err, corsGetFailed)
	}
	var local []v1beta1.CORSRule
	if bucket.Spec.ForProvider.CORSConfiguration != nil {
		local = bucket.Spec.ForProvider.CORSConfiguration.CORSRules
	}
	var external []types.CORSRule
	if result != nil {
		external = result.CORSRules
	}
	return CompareCORS(local, external), nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *CORSConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.CORSConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketCorsInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.CORSConfiguration)
	_, err := in.client.PutBucketCors(ctx, input)
	return errorutils.Wrap(err, corsPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *CORSConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketCors(ctx,
		&awss3.DeleteBucketCorsInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, corsDeleteFailed)
}

// LateInitialize does nothing because CORSConfiguration might have been deleted
// by the user.
func (in *CORSConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketCors(ctx, &awss3.GetBucketCorsInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.CORSConfigurationNotFound, err), corsGetFailed)
	}

	// We need the second check here because by default the CORS is not set
	if external == nil || len(external.CORSRules) == 0 {
		return nil
	}

	fp := &bucket.Spec.ForProvider
	if fp.CORSConfiguration == nil {
		fp.CORSConfiguration = &v1beta1.CORSConfiguration{}
	}

	if fp.CORSConfiguration.CORSRules == nil {
		// only run late init if the user has not specified CORSRules
		bucket.Spec.ForProvider.CORSConfiguration.CORSRules = GenerateCORSRule(external.CORSRules)
	}

	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *CORSConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.CORSConfiguration != nil
}

// GeneratePutBucketCorsInput creates the input for the PutBucketCors request for the S3 Client
func GeneratePutBucketCorsInput(name string, config *v1beta1.CORSConfiguration) *awss3.PutBucketCorsInput {
	bci := &awss3.PutBucketCorsInput{
		Bucket:            pointer.ToOrNilIfZeroValue(name),
		CORSConfiguration: &types.CORSConfiguration{CORSRules: make([]types.CORSRule, 0)},
	}
	for _, cors := range config.CORSRules {
		bci.CORSConfiguration.CORSRules = append(bci.CORSConfiguration.CORSRules, types.CORSRule{
			AllowedHeaders: cors.AllowedHeaders,
			AllowedMethods: cors.AllowedMethods,
			AllowedOrigins: cors.AllowedOrigins,
			ExposeHeaders:  cors.ExposeHeaders,
			MaxAgeSeconds:  cors.MaxAgeSeconds,
		})
	}
	return bci
}

// CompareCORS compares the external and internal representations for the list of CORSRules
func CompareCORS(local []v1beta1.CORSRule, external []types.CORSRule) ResourceStatus { //nolint:gocyclo
	switch {
	case len(local) == 0 && len(external) != 0:
		return NeedsDeletion
	case len(local) == 0 && len(external) == 0:
		return Updated
	case len(local) != len(external):
		return NeedsUpdate
	}

	for i := range local {
		outputRule := external[i]
		if !(cmp.Equal(local[i].AllowedHeaders, outputRule.AllowedHeaders) &&
			cmp.Equal(local[i].AllowedMethods, outputRule.AllowedMethods) &&
			cmp.Equal(local[i].AllowedOrigins, outputRule.AllowedOrigins) &&
			cmp.Equal(local[i].ExposeHeaders, outputRule.ExposeHeaders) &&
			local[i].MaxAgeSeconds == outputRule.MaxAgeSeconds) {
			return NeedsUpdate
		}
	}

	return Updated
}

// GenerateCORSRule creates the cors rule from a GetBucketCORS request from the S3 Client
func GenerateCORSRule(config []types.CORSRule) []v1beta1.CORSRule {
	output := make([]v1beta1.CORSRule, len(config))
	for i, cors := range config {
		output[i] = v1beta1.CORSRule{
			AllowedHeaders: cors.AllowedHeaders,
			AllowedMethods: cors.AllowedMethods,
			AllowedOrigins: cors.AllowedOrigins,
			ExposeHeaders:  cors.ExposeHeaders,
			MaxAgeSeconds:  cors.MaxAgeSeconds,
		}
	}
	return output
}
