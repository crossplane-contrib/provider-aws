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
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

var _ BucketResource = &SSEConfigurationClient{}

// SSEConfigurationClient is the client for API methods and reconciling the ServerSideEncryptionConfiguration
type SSEConfigurationClient struct {
	config *v1beta1.ServerSideEncryptionConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *SSEConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	// GetBucketEncryptionRequest throws an error if nothing exists externally
	// Future work can be done to support brownfield initialization for the SSEConfiguration
	// TODO
	return nil
}

// NewSSEConfigurationClient creates the client for Server Side Encryption Configuration
func NewSSEConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *SSEConfigurationClient {
	return &SSEConfigurationClient{config: bucket.Spec.ForProvider.ServerSideEncryptionConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *SSEConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { // nolint:gocyclo
	enc, err := in.client.GetBucketEncryptionRequest(&awss3.GetBucketEncryptionInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil && s3.SSEConfigurationNotFound(err) && in.config == nil {
		return Updated, nil
	} else if err != nil {
		return NeedsUpdate, errors.Wrap(err, sseGetFailed)
	}

	switch {
	case enc.ServerSideEncryptionConfiguration != nil && in.config == nil:
		return NeedsDeletion, nil
	case enc.ServerSideEncryptionConfiguration == nil && in.config == nil:
		return Updated, nil
	case enc.ServerSideEncryptionConfiguration == nil && in.config != nil:
		return NeedsUpdate, nil
	case len(enc.ServerSideEncryptionConfiguration.Rules) != len(in.config.Rules):
		return NeedsUpdate, nil
	}

	for i, Rule := range in.config.Rules {
		outputRule := enc.ServerSideEncryptionConfiguration.Rules[i].ApplyServerSideEncryptionByDefault
		if outputRule.KMSMasterKeyID != Rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID {
			return NeedsUpdate, nil
		}
		if string(outputRule.SSEAlgorithm) != Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm {
			return NeedsUpdate, nil
		}
	}

	return Updated, nil
}

// GeneratePutBucketEncryptionInput creates the input for the PutBucketEncryption request for the S3 Client
func GeneratePutBucketEncryptionInput(name string, in *SSEConfigurationClient) *awss3.PutBucketEncryptionInput {
	bei := &awss3.PutBucketEncryptionInput{
		Bucket:                            aws.String(name),
		ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{},
	}
	for _, rule := range in.config.Rules {
		bei.ServerSideEncryptionConfiguration.Rules = append(bei.ServerSideEncryptionConfiguration.Rules, awss3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
				KMSMasterKeyID: rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				SSEAlgorithm:   awss3.ServerSideEncryption(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm),
			},
		})
	}
	return bei
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *SSEConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketEncryptionRequest(GeneratePutBucketEncryptionInput(meta.GetExternalName(bucket), in)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, ssePutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *SSEConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketEncryptionRequest(
		&awss3.DeleteBucketEncryptionInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, sseDeleteFailed)
}
