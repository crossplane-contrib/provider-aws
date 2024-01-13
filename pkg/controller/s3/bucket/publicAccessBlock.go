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
	awss3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	publicAccessBlockGetFailed    = "cannot get Bucket public access block"
	publicAccessBlockPutFailed    = "cannot put Bucket public access block"
	publicAccessBlockDeleteFailed = "cannot delete Bucket public access block"
)

// PublicAccessBlockClient is the client for API methods and reconciling the PublicAccessBlock
type PublicAccessBlockClient struct {
	client s3.BucketClient
}

// NewPublicAccessBlockClient creates the client for Accelerate Configuration
func NewPublicAccessBlockClient(client s3.BucketClient) *PublicAccessBlockClient {
	return &PublicAccessBlockClient{client: client}
}

func isDisabledPublicAccessBlock(cr *v1beta1.Bucket) bool {
	return cr.Spec.ForProvider.PublicAccessBlockConfiguration != nil &&
		!pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls) &&
		!pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy) &&
		!pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets) &&
		!pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls)
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *PublicAccessBlockClient) Observe(ctx context.Context, cr *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetPublicAccessBlock(ctx, &awss3.GetPublicAccessBlockInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	if s3.PublicAccessBlockConfigurationNotFound(err) && (cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil || isDisabledPublicAccessBlock(cr)) {
		return Updated, nil
	}
	if err == nil && isDisabledPublicAccessBlock(cr) {
		return NeedsDeletion, nil
	}
	if err != nil {
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockGetFailed)
	}
	if !isPublicAccessBlockUpToDate(cr, external) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

func isPublicAccessBlockUpToDate(cr *v1beta1.Bucket, external *awss3.GetPublicAccessBlockOutput) bool {
	if cr.Spec.ForProvider.PublicAccessBlockConfiguration != nil {
		switch {
		case pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls) != external.PublicAccessBlockConfiguration.BlockPublicAcls:
			return false
		case pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy) != external.PublicAccessBlockConfiguration.BlockPublicPolicy:
			return false
		case pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets) != external.PublicAccessBlockConfiguration.RestrictPublicBuckets:
			return false
		case pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls) != external.PublicAccessBlockConfiguration.IgnorePublicAcls:
			return false
		}
	}
	return true
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *PublicAccessBlockClient) CreateOrUpdate(ctx context.Context, cr *v1beta1.Bucket) error {
	if cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil {
		return nil
	}
	input := &awss3.PutPublicAccessBlockInput{
		Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		PublicAccessBlockConfiguration: &awss3types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls),
			BlockPublicPolicy:     pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy),
			RestrictPublicBuckets: pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets),
			IgnorePublicAcls:      pointer.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls),
		},
	}
	_, err := in.client.PutPublicAccessBlock(ctx, input)
	return errorutils.Wrap(err, publicAccessBlockPutFailed)
}

// Delete removes the public access block configuration.
func (in *PublicAccessBlockClient) Delete(ctx context.Context, cr *v1beta1.Bucket) error {
	input := &awss3.DeletePublicAccessBlockInput{
		Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}
	_, err := in.client.DeletePublicAccessBlock(ctx, input)
	return errors.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockDeleteFailed)
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *PublicAccessBlockClient) LateInitialize(ctx context.Context, cr *v1beta1.Bucket) error {
	external, err := in.client.GetPublicAccessBlock(ctx, &awss3.GetPublicAccessBlockInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockGetFailed)
	}
	if external.PublicAccessBlockConfiguration == nil {
		return nil
	}

	if cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil {
		cr.Spec.ForProvider.PublicAccessBlockConfiguration = &v1beta1.PublicAccessBlockConfiguration{}
	}
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls = pointer.LateInitialize(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls, pointer.ToOrNilIfZeroValue(external.PublicAccessBlockConfiguration.BlockPublicAcls))
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy = pointer.LateInitialize(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy, pointer.ToOrNilIfZeroValue(external.PublicAccessBlockConfiguration.BlockPublicPolicy))
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets = pointer.LateInitialize(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets, pointer.ToOrNilIfZeroValue(external.PublicAccessBlockConfiguration.RestrictPublicBuckets))
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls = pointer.LateInitialize(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls, pointer.ToOrNilIfZeroValue(external.PublicAccessBlockConfiguration.IgnorePublicAcls))
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *PublicAccessBlockClient) SubresourceExists(cr *v1beta1.Bucket) bool {
	return cr.Spec.ForProvider.PublicAccessBlockConfiguration != nil
}
