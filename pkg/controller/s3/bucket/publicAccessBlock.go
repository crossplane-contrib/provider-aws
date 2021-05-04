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
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
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

// Observe checks if the resource exists and if it matches the local configuration
func (in *PublicAccessBlockClient) Observe(ctx context.Context, cr *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetPublicAccessBlockRequest(&awss3.GetPublicAccessBlockInput{Bucket: awsclient.String(meta.GetExternalName(cr))}).Send(ctx)
	if s3.PublicAccessBlockConfigurationNotFound(err) && cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil {
		return Updated, nil
	}
	if err != nil {
		return NeedsUpdate, awsclient.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockGetFailed)
	}
	if cr.Spec.ForProvider.PublicAccessBlockConfiguration != nil {
		switch {
		case awsclient.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls) != awsclient.BoolValue(external.PublicAccessBlockConfiguration.BlockPublicAcls):
			return NeedsUpdate, nil
		case awsclient.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy) != awsclient.BoolValue(external.PublicAccessBlockConfiguration.BlockPublicPolicy):
			return NeedsUpdate, nil
		case awsclient.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets) != awsclient.BoolValue(external.PublicAccessBlockConfiguration.RestrictPublicBuckets):
			return NeedsUpdate, nil
		case awsclient.BoolValue(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls) != awsclient.BoolValue(external.PublicAccessBlockConfiguration.IgnorePublicAcls):
			return NeedsUpdate, nil
		}
	}
	return Updated, nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *PublicAccessBlockClient) CreateOrUpdate(ctx context.Context, cr *v1beta1.Bucket) error {
	if cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil {
		return nil
	}
	input := &awss3.PutPublicAccessBlockInput{
		Bucket: awsclient.String(meta.GetExternalName(cr)),
		PublicAccessBlockConfiguration: &awss3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls,
			BlockPublicPolicy:     cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy,
			RestrictPublicBuckets: cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets,
			IgnorePublicAcls:      cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls,
		},
	}
	_, err := in.client.PutPublicAccessBlockRequest(input).Send(ctx)
	return awsclient.Wrap(err, publicAccessBlockPutFailed)
}

// Delete removes the public access block configuration.
func (in *PublicAccessBlockClient) Delete(ctx context.Context, cr *v1beta1.Bucket) error {
	input := &awss3.DeletePublicAccessBlockInput{
		Bucket: awsclient.String(meta.GetExternalName(cr)),
	}
	_, err := in.client.DeletePublicAccessBlockRequest(input).Send(ctx)
	return errors.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockDeleteFailed)
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *PublicAccessBlockClient) LateInitialize(ctx context.Context, cr *v1beta1.Bucket) error {
	external, err := in.client.GetPublicAccessBlockRequest(&awss3.GetPublicAccessBlockInput{Bucket: awsclient.String(meta.GetExternalName(cr))}).Send(ctx)
	if err != nil {
		return awsclient.Wrap(resource.Ignore(s3.PublicAccessBlockConfigurationNotFound, err), publicAccessBlockGetFailed)
	}

	if cr.Spec.ForProvider.PublicAccessBlockConfiguration == nil {
		cr.Spec.ForProvider.PublicAccessBlockConfiguration = &v1beta1.PublicAccessBlockConfiguration{}
	}
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls = awsclient.LateInitializeBoolPtr(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicAcls, external.PublicAccessBlockConfiguration.BlockPublicAcls)
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy = awsclient.LateInitializeBoolPtr(cr.Spec.ForProvider.PublicAccessBlockConfiguration.BlockPublicPolicy, external.PublicAccessBlockConfiguration.BlockPublicPolicy)
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets = awsclient.LateInitializeBoolPtr(cr.Spec.ForProvider.PublicAccessBlockConfiguration.RestrictPublicBuckets, external.PublicAccessBlockConfiguration.RestrictPublicBuckets)
	cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls = awsclient.LateInitializeBoolPtr(cr.Spec.ForProvider.PublicAccessBlockConfiguration.IgnorePublicAcls, external.PublicAccessBlockConfiguration.IgnorePublicAcls)
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *PublicAccessBlockClient) SubresourceExists(cr *v1beta1.Bucket) bool {
	return cr.Spec.ForProvider.PublicAccessBlockConfiguration != nil
}
