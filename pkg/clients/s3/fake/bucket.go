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

package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
)

// this ensures that the mock implements the client interface
var _ clientset.BucketClient = (*MockBucketClient)(nil)

// MockBucketClient is a type that implements all the methods for BucketClient interface
type MockBucketClient struct {
	MockHeadBucket   func(ctx context.Context, input *s3.HeadBucketInput, opts []func(*s3.Options)) (*s3.HeadBucketOutput, error)
	MockCreateBucket func(ctx context.Context, input *s3.CreateBucketInput, opts []func(*s3.Options)) (*s3.CreateBucketOutput, error)
	MockDeleteBucket func(ctx context.Context, input *s3.DeleteBucketInput, opts []func(*s3.Options)) (*s3.DeleteBucketOutput, error)

	MockPutBucketEncryption    func(ctx context.Context, input *s3.PutBucketEncryptionInput, opts []func(*s3.Options)) (*s3.PutBucketEncryptionOutput, error)
	MockGetBucketEncryption    func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error)
	MockDeleteBucketEncryption func(ctx context.Context, input *s3.DeleteBucketEncryptionInput, opts []func(*s3.Options)) (*s3.DeleteBucketEncryptionOutput, error)

	MockPutBucketVersioning func(ctx context.Context, input *s3.PutBucketVersioningInput, opts []func(*s3.Options)) (*s3.PutBucketVersioningOutput, error)
	MockGetBucketVersioning func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)

	MockPutBucketAccelerateConfiguration func(ctx context.Context, input *s3.PutBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketAccelerateConfigurationOutput, error)
	MockGetBucketAccelerateConfiguration func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error)

	MockPutBucketCors    func(ctx context.Context, input *s3.PutBucketCorsInput, opts []func(*s3.Options)) (*s3.PutBucketCorsOutput, error)
	MockGetBucketCors    func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error)
	MockDeleteBucketCors func(ctx context.Context, input *s3.DeleteBucketCorsInput, opts []func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error)

	MockPutBucketWebsite    func(ctx context.Context, input *s3.PutBucketWebsiteInput, opts []func(*s3.Options)) (*s3.PutBucketWebsiteOutput, error)
	MockGetBucketWebsite    func(ctx context.Context, input *s3.GetBucketWebsiteInput, opts []func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error)
	MockDeleteBucketWebsite func(ctx context.Context, input *s3.DeleteBucketWebsiteInput, opts []func(*s3.Options)) (*s3.DeleteBucketWebsiteOutput, error)

	MockPutBucketLogging func(ctx context.Context, input *s3.PutBucketLoggingInput, opts []func(*s3.Options)) (*s3.PutBucketLoggingOutput, error)
	MockGetBucketLogging func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error)

	MockPutBucketReplication    func(ctx context.Context, input *s3.PutBucketReplicationInput, opts []func(*s3.Options)) (*s3.PutBucketReplicationOutput, error)
	MockGetBucketReplication    func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error)
	MockDeleteBucketReplication func(ctx context.Context, input *s3.DeleteBucketReplicationInput, opts []func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error)

	MockPutBucketRequestPayment func(ctx context.Context, input *s3.PutBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error)
	MockGetBucketRequestPayment func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error)

	MockPutBucketTagging    func(ctx context.Context, input *s3.PutBucketTaggingInput, opts []func(*s3.Options)) (*s3.PutBucketTaggingOutput, error)
	MockGetBucketTagging    func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error)
	MockDeleteBucketTagging func(ctx context.Context, input *s3.DeleteBucketTaggingInput, opts []func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error)

	MockPutBucketAnalyticsConfiguration func(ctx context.Context, input *s3.PutBucketAnalyticsConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketAnalyticsConfigurationOutput, error)
	MockGetBucketAnalyticsConfiguration func(ctx context.Context, input *s3.GetBucketAnalyticsConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error)

	MockPutBucketLifecycleConfiguration func(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error)
	MockGetBucketLifecycleConfiguration func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error)
	MockDeleteBucketLifecycle           func(ctx context.Context, input *s3.DeleteBucketLifecycleInput, opts []func(*s3.Options)) (*s3.DeleteBucketLifecycleOutput, error)

	MockPutBucketNotificationConfiguration func(ctx context.Context, input *s3.PutBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error)
	MockGetBucketNotificationConfiguration func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error)

	MockGetBucketAcl func(ctx context.Context, input *s3.GetBucketAclInput, opts []func(*s3.Options)) (*s3.GetBucketAclOutput, error)
	MockPutBucketAcl func(ctx context.Context, input *s3.PutBucketAclInput, opts []func(*s3.Options)) (*s3.PutBucketAclOutput, error)

	MockGetPublicAccessBlock    func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error)
	MockPutPublicAccessBlock    func(ctx context.Context, input *s3.PutPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.PutPublicAccessBlockOutput, error)
	MockDeletePublicAccessBlock func(ctx context.Context, input *s3.DeletePublicAccessBlockInput, opts []func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error)

	MockGetBucketOwnershipControls    func(ctx context.Context, input *s3.GetBucketOwnershipControlsInput, opts []func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error)
	MockPutBucketOwnershipControls    func(ctx context.Context, input *s3.PutBucketOwnershipControlsInput, opts []func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error)
	MockDeleteBucketOwnershipControls func(ctx context.Context, input *s3.DeleteBucketOwnershipControlsInput, opts []func(*s3.Options)) (*s3.DeleteBucketOwnershipControlsOutput, error)

	MockBucketPolicyClient
}

// HeadBucket is the fake method call to invoke the internal mock method
func (m MockBucketClient) HeadBucket(ctx context.Context, input *s3.HeadBucketInput, opts ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return m.MockHeadBucket(ctx, input, opts)
}

// CreateBucket is the fake method call to invoke the internal mock method
func (m MockBucketClient) CreateBucket(ctx context.Context, input *s3.CreateBucketInput, opts ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.MockCreateBucket(ctx, input, opts)
}

// DeleteBucket is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucket(ctx context.Context, input *s3.DeleteBucketInput, opts ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return m.MockDeleteBucket(ctx, input, opts)
}

// PutBucketEncryption is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketEncryption(ctx context.Context, input *s3.PutBucketEncryptionInput, opts ...func(*s3.Options)) (*s3.PutBucketEncryptionOutput, error) {
	return m.MockPutBucketEncryption(ctx, input, opts)
}

// GetBucketEncryption is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketEncryption(ctx context.Context, input *s3.GetBucketEncryptionInput, opts ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	return m.MockGetBucketEncryption(ctx, input, opts)
}

// DeleteBucketEncryption is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketEncryption(ctx context.Context, input *s3.DeleteBucketEncryptionInput, opts ...func(*s3.Options)) (*s3.DeleteBucketEncryptionOutput, error) {
	return m.MockDeleteBucketEncryption(ctx, input, opts)
}

// PutBucketVersioning is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketVersioning(ctx context.Context, input *s3.PutBucketVersioningInput, opts ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return m.MockPutBucketVersioning(ctx, input, opts)
}

// GetBucketVersioning is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketVersioning(ctx context.Context, input *s3.GetBucketVersioningInput, opts ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return m.MockGetBucketVersioning(ctx, input, opts)
}

// PutBucketAccelerateConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAccelerateConfiguration(ctx context.Context, input *s3.PutBucketAccelerateConfigurationInput, opts ...func(*s3.Options)) (*s3.PutBucketAccelerateConfigurationOutput, error) {
	return m.MockPutBucketAccelerateConfiguration(ctx, input, opts)
}

// GetBucketAccelerateConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAccelerateConfiguration(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts ...func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
	return m.MockGetBucketAccelerateConfiguration(ctx, input, opts)
}

// PutBucketCors is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketCors(ctx context.Context, input *s3.PutBucketCorsInput, opts ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
	return m.MockPutBucketCors(ctx, input, opts)
}

// GetBucketCors is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketCors(ctx context.Context, input *s3.GetBucketCorsInput, opts ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
	return m.MockGetBucketCors(ctx, input, opts)
}

// DeleteBucketCors is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketCors(ctx context.Context, input *s3.DeleteBucketCorsInput, opts ...func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error) {
	return m.MockDeleteBucketCors(ctx, input, opts)
}

// PutBucketWebsite is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketWebsite(ctx context.Context, input *s3.PutBucketWebsiteInput, opts ...func(*s3.Options)) (*s3.PutBucketWebsiteOutput, error) {
	return m.MockPutBucketWebsite(ctx, input, opts)
}

// GetBucketWebsite is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketWebsite(ctx context.Context, input *s3.GetBucketWebsiteInput, opts ...func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error) {
	return m.MockGetBucketWebsite(ctx, input, opts)
}

// DeleteBucketWebsite is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketWebsite(ctx context.Context, input *s3.DeleteBucketWebsiteInput, opts ...func(*s3.Options)) (*s3.DeleteBucketWebsiteOutput, error) {
	return m.MockDeleteBucketWebsite(ctx, input, opts)
}

// PutBucketLogging is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketLogging(ctx context.Context, input *s3.PutBucketLoggingInput, opts ...func(*s3.Options)) (*s3.PutBucketLoggingOutput, error) {
	return m.MockPutBucketLogging(ctx, input, opts)
}

// GetBucketLogging is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketLogging(ctx context.Context, input *s3.GetBucketLoggingInput, opts ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
	return m.MockGetBucketLogging(ctx, input, opts)
}

// PutBucketReplication is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketReplication(ctx context.Context, input *s3.PutBucketReplicationInput, opts ...func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
	return m.MockPutBucketReplication(ctx, input, opts)
}

// GetBucketReplication is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketReplication(ctx context.Context, input *s3.GetBucketReplicationInput, opts ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
	return m.MockGetBucketReplication(ctx, input, opts)
}

// DeleteBucketReplication is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketReplication(ctx context.Context, input *s3.DeleteBucketReplicationInput, opts ...func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error) {
	return m.MockDeleteBucketReplication(ctx, input, opts)
}

// PutBucketRequestPayment is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketRequestPayment(ctx context.Context, input *s3.PutBucketRequestPaymentInput, opts ...func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error) {
	return m.MockPutBucketRequestPayment(ctx, input, opts)
}

// GetBucketRequestPayment is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketRequestPayment(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts ...func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
	return m.MockGetBucketRequestPayment(ctx, input, opts)
}

// PutBucketTagging is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketTagging(ctx context.Context, input *s3.PutBucketTaggingInput, opts ...func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
	return m.MockPutBucketTagging(ctx, input, opts)
}

// GetBucketTagging is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketTagging(ctx context.Context, input *s3.GetBucketTaggingInput, opts ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
	return m.MockGetBucketTagging(ctx, input, opts)
}

// DeleteBucketTagging is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketTagging(ctx context.Context, input *s3.DeleteBucketTaggingInput, opts ...func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error) {
	return m.MockDeleteBucketTagging(ctx, input, opts)
}

// PutBucketAnalyticsConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAnalyticsConfiguration(ctx context.Context, input *s3.PutBucketAnalyticsConfigurationInput, opts ...func(*s3.Options)) (*s3.PutBucketAnalyticsConfigurationOutput, error) {
	return m.MockPutBucketAnalyticsConfiguration(ctx, input, opts)
}

// GetBucketAnalyticsConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAnalyticsConfiguration(ctx context.Context, input *s3.GetBucketAnalyticsConfigurationInput, opts ...func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	return m.MockGetBucketAnalyticsConfiguration(ctx, input, opts)
}

// PutBucketLifecycleConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketLifecycleConfiguration(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput, opts ...func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	return m.MockPutBucketLifecycleConfiguration(ctx, input, opts)
}

// GetBucketLifecycleConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketLifecycleConfiguration(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	return m.MockGetBucketLifecycleConfiguration(ctx, input, opts)
}

// DeleteBucketLifecycle is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketLifecycle(ctx context.Context, input *s3.DeleteBucketLifecycleInput, opts ...func(*s3.Options)) (*s3.DeleteBucketLifecycleOutput, error) {
	return m.MockDeleteBucketLifecycle(ctx, input, opts)
}

// PutBucketNotificationConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketNotificationConfiguration(ctx context.Context, input *s3.PutBucketNotificationConfigurationInput, opts ...func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error) {
	return m.MockPutBucketNotificationConfiguration(ctx, input, opts)
}

// GetBucketNotificationConfiguration is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketNotificationConfiguration(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
	return m.MockGetBucketNotificationConfiguration(ctx, input, opts)
}

// GetBucketAcl is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAcl(ctx context.Context, input *s3.GetBucketAclInput, opts ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	return m.MockGetBucketAcl(ctx, input, opts)
}

// PutBucketAcl is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAcl(ctx context.Context, input *s3.PutBucketAclInput, opts ...func(*s3.Options)) (*s3.PutBucketAclOutput, error) {
	return m.MockPutBucketAcl(ctx, input, opts)
}

// GetPublicAccessBlock is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetPublicAccessBlock(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts ...func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
	return m.MockGetPublicAccessBlock(ctx, input, opts)
}

// PutPublicAccessBlock is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutPublicAccessBlock(ctx context.Context, input *s3.PutPublicAccessBlockInput, opts ...func(*s3.Options)) (*s3.PutPublicAccessBlockOutput, error) {
	return m.MockPutPublicAccessBlock(ctx, input, opts)
}

// DeletePublicAccessBlock is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeletePublicAccessBlock(ctx context.Context, input *s3.DeletePublicAccessBlockInput, opts ...func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error) {
	return m.MockDeletePublicAccessBlock(ctx, input, opts)
}

// GetBucketOwnershipControls is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketOwnershipControls(ctx context.Context, input *s3.GetBucketOwnershipControlsInput, opts ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error) {
	return m.MockGetBucketOwnershipControls(ctx, input, opts)
}

// PutBucketOwnershipControls is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketOwnershipControls(ctx context.Context, input *s3.PutBucketOwnershipControlsInput, opts ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
	return m.MockPutBucketOwnershipControls(ctx, input, opts)
}

// DeleteBucketOwnershipControls is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketOwnershipControls(ctx context.Context, input *s3.DeleteBucketOwnershipControlsInput, opts ...func(*s3.Options)) (*s3.DeleteBucketOwnershipControlsOutput, error) {
	return m.MockDeleteBucketOwnershipControls(ctx, input, opts)
}
