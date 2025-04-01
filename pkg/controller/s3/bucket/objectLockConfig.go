/*
Copyright 2025 The Crossplane Authors.

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
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	errMsgObjectLockIsNotDisableable    = "after you enable object lock for a bucket, you can't disable object lock that bucket"
	errObjectLockConfigurationGetFailed = "cannot get object lock configuration"
	errObjectLockConfigurationPutFailed = "cannot put object lock configuration"
)

type objectLockConfigurationCache struct {
	objectLockConfiguration *types.ObjectLockConfiguration
}

// ObjectLockConfigurationClient is the client for API methods and reconciling the ObjectLockConfiguration
type ObjectLockConfigurationClient struct {
	client s3.BucketClient
	cache  objectLockConfigurationCache
}

// NewObjectLockConfigurationClient creates the client for Object Lock Configuration
func NewObjectLockConfigurationClient(client s3.BucketClient) *ObjectLockConfigurationClient {
	return &ObjectLockConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *ObjectLockConfigurationClient) Observe(ctx context.Context, cr *v1beta1.Bucket) (ResourceStatus, error) {
	response, err := in.client.GetObjectLockConfiguration(ctx, &awss3.GetObjectLockConfigurationInput{
		Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		if s3.ObjectLockConfigurationNotFound(err) && (cr.Spec.ForProvider.ObjectLockEnabledForBucket == nil || !*cr.Spec.ForProvider.ObjectLockEnabledForBucket) {
			return Updated, nil
		}
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.ObjectLockConfigurationNotFound, err), errObjectLockConfigurationGetFailed)
	}
	objectLockConfiguration := &types.ObjectLockConfiguration{
		ObjectLockEnabled: "Disabled",
	}
	if response != nil && response.ObjectLockConfiguration != nil {
		objectLockConfiguration = response.ObjectLockConfiguration
	}
	in.cache.objectLockConfiguration = objectLockConfiguration
	return isUpToDate(&cr.Spec.ForProvider, objectLockConfiguration)
}

// isUpToDate compares the local configuration with the remote configuration and determines if an update is needed
func isUpToDate(cr *v1beta1.BucketParameters, resp *types.ObjectLockConfiguration) (ResourceStatus, error) {
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(types.ObjectLockConfiguration{}, "noSmithyDocumentSerde",
		"Rule.noSmithyDocumentSerde", "Rule.DefaultRetention.noSmithyDocumentSerde")}
	if !cmp.Equal(*GenerateAWSObjectLockConfiguration(cr), *resp, cmpOpts...) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// CreateOrUpdate updates or creates the Object Lock configuration on AWS side
func (in *ObjectLockConfigurationClient) CreateOrUpdate(ctx context.Context, cr *v1beta1.Bucket) error {
	if in.cache.objectLockConfiguration != nil {
		if in.cache.objectLockConfiguration.ObjectLockEnabled == "Enabled" && !*cr.Spec.ForProvider.ObjectLockEnabledForBucket {
			err := errorutils.Wrap(errors.New(errMsgObjectLockIsNotDisableable), errObjectLockConfigurationPutFailed)
			cr.SetConditions(xpv1.ReconcileError(errors.New(errMsgObjectLockIsNotDisableable)))
			return err
		}
	}
	input := &awss3.PutObjectLockConfigurationInput{
		Bucket:                  aws.String(meta.GetExternalName(cr)),
		ObjectLockConfiguration: GenerateAWSObjectLockConfiguration(&cr.Spec.ForProvider),
	}
	_, err := in.client.PutObjectLockConfiguration(ctx, input)
	if err != nil {
		cr.SetConditions(xpv1.ReconcileError(err))
		return errorutils.Wrap(err, errObjectLockConfigurationPutFailed)
	}
	return nil
}

// Delete does nothing because after you enable Object Lock for a bucket, you can't disable it anymore
func (in *ObjectLockConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// LateInitialize is not needed, because if something is not specified in the current state, it should be deleted on aws side
func (in *ObjectLockConfigurationClient) LateInitialize(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// SubresourceExists checks if the subresource exists
func (in *ObjectLockConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.ObjectLockEnabledForBucket != nil
}

// GenerateAWSObjectLockConfiguration generates the AWS S3 Object Lock Configuration
func GenerateAWSObjectLockConfiguration(in *v1beta1.BucketParameters) *types.ObjectLockConfiguration {
	objectLockConfiguration := &types.ObjectLockConfiguration{
		ObjectLockEnabled: "Disabled",
	}
	if in.ObjectLockEnabledForBucket != nil {
		if *in.ObjectLockEnabledForBucket {
			objectLockConfiguration.ObjectLockEnabled = "Enabled"
		}
	}
	if in.ObjectLockRule != nil {
		objectLockConfiguration.Rule = &types.ObjectLockRule{
			DefaultRetention: &types.DefaultRetention{
				Days:  in.ObjectLockRule.DefaultRetention.Days,
				Mode:  types.ObjectLockRetentionMode(in.ObjectLockRule.DefaultRetention.Mode),
				Years: in.ObjectLockRule.DefaultRetention.Years,
			},
		}
	}
	return objectLockConfiguration
}
