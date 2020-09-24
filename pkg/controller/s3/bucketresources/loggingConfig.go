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

var _ BucketResource = &LoggingConfigurationClient{}

// LoggingConfigurationClient is the client for API methods and reconciling the LoggingConfiguration
type LoggingConfigurationClient struct {
	config *v1beta1.LoggingConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
// this function support brownfield initialization, but it does not reconcile subsequent external updates.
// TODO: This could be the subject for future work, pending further discussion with the maintainers
func (in *LoggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	conf, err := in.client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get bucket accelerate configuration")
	}
	if conf.LoggingEnabled == nil {
		// There is no value send by AWS to initialize
		return nil
	}
	if in.config == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.LoggingConfiguration = &v1beta1.LoggingConfiguration{}
		in.config = bucket.Spec.ForProvider.LoggingConfiguration
	}
	// Late initialize the target bucket and target prefix
	in.config.TargetBucket = aws.LateInitializeStringPtr(in.config.TargetBucket, conf.LoggingEnabled.TargetBucket)
	in.config.TargetPrefix = aws.LateInitializeStringPtr(in.config.TargetPrefix, conf.LoggingEnabled.TargetPrefix)
	// If the there is an external target grant list, and the local one does not exist
	// we create the target grant list
	if conf.LoggingEnabled.TargetGrants != nil && len(in.config.TargetGrants) == 0 {
		in.config.TargetGrants = make([]v1beta1.TargetGrant, len(conf.LoggingEnabled.TargetGrants))
		for i, v := range conf.LoggingEnabled.TargetGrants {
			in.config.TargetGrants[i] = v1beta1.TargetGrant{
				Grantee: v1beta1.TargetGrantee{
					DisplayName:  v.Grantee.DisplayName,
					EmailAddress: v.Grantee.EmailAddress,
					ID:           v.Grantee.ID,
					Type:         string(v.Grantee.Type),
					URI:          v.Grantee.URI,
				},
				Permission: string(v.Permission),
			}
		}
	}
	return nil
}

// NewLoggingConfigurationClient creates the client for Logging Configuration
func NewLoggingConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *LoggingConfigurationClient {
	return &LoggingConfigurationClient{config: bucket.Spec.ForProvider.LoggingConfiguration, client: client}
}

// CompareStrings compares pairs of strings passed in
func CompareStrings(strings ...*string) bool {
	if len(strings)%2 != 0 {
		return false
	}
	for i := 0; i < len(strings); i += 2 {
		if aws.StringValue(strings[i]) != aws.StringValue(strings[i+1]) {
			return false
		}
	}
	return true
}

func compareLogging(local *v1beta1.LoggingConfiguration, external *awss3.LoggingEnabled) ResourceStatus {
	if aws.StringValue(external.TargetBucket) != aws.StringValue(local.TargetBucket) {
		return NeedsUpdate
	}

	if aws.StringValue(external.TargetPrefix) != aws.StringValue(local.TargetPrefix) {
		return NeedsUpdate
	}

	for i, grant := range local.TargetGrants {
		outputGrant := external.TargetGrants[i]
		if outputGrant.Grantee != nil {
			oGrant := outputGrant.Grantee
			lGrant := grant.Grantee
			if !CompareStrings(oGrant.DisplayName, lGrant.DisplayName,
				oGrant.EmailAddress, lGrant.EmailAddress,
				oGrant.ID, lGrant.ID,
				oGrant.URI, lGrant.URI) {
				return NeedsUpdate
			}
			if string(oGrant.Type) != lGrant.Type {
				return NeedsUpdate
			}
		}
		if string(outputGrant.Permission) != grant.Permission {
			return NeedsUpdate
		}
	}

	return Updated
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LoggingConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket logging")
	}

	if conf.LoggingEnabled == nil && in.config == nil {
		return Updated, nil
	} else if conf.LoggingEnabled != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	return compareLogging(in.config, conf.LoggingEnabled), nil
}

// GeneratePutBucketLoggingInput creates the input for the PutBucketLogging request for the S3 Client
func GeneratePutBucketLoggingInput(name string, in *LoggingConfigurationClient) *awss3.PutBucketLoggingInput {
	bci := &awss3.PutBucketLoggingInput{
		Bucket: aws.String(name),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{LoggingEnabled: &awss3.LoggingEnabled{
			TargetBucket: in.config.TargetBucket,
			TargetGrants: make([]awss3.TargetGrant, 0),
			TargetPrefix: in.config.TargetPrefix,
		}},
	}
	for _, grant := range in.config.TargetGrants {
		bci.BucketLoggingStatus.LoggingEnabled.TargetGrants = append(bci.BucketLoggingStatus.LoggingEnabled.TargetGrants, awss3.TargetGrant{
			Grantee: &awss3.Grantee{
				DisplayName:  grant.Grantee.DisplayName,
				EmailAddress: grant.Grantee.EmailAddress,
				ID:           grant.Grantee.ID,
				Type:         awss3.Type(grant.Grantee.Type),
				URI:          grant.Grantee.URI,
			},
			Permission: awss3.BucketLogsPermission(grant.Permission),
		})
	}
	return bci
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *LoggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketLoggingRequest(GeneratePutBucketLoggingInput(meta.GetExternalName(bucket), in)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket logging")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *LoggingConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := &awss3.PutBucketLoggingInput{
		Bucket:              aws.String(meta.GetExternalName(bucket)),
		BucketLoggingStatus: &awss3.BucketLoggingStatus{}, //  Empty BucketLoggingStatus disables logging
	}
	_, err := in.client.PutBucketLoggingRequest(input).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket logging")
}
