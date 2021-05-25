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
	"github.com/aws/aws-sdk-go-v2/service/s3"

	clientset "github.com/crossplane/provider-aws/pkg/clients/s3"
)

// this ensures that the mock implements the client interface
var _ clientset.BucketClient = (*MockBucketClient)(nil)

// MockBucketClient is a type that implements all the methods for BucketClient interface
type MockBucketClient struct {
	MockHeadBucketRequest   func(input *s3.HeadBucketInput) s3.HeadBucketRequest
	MockCreateBucketRequest func(input *s3.CreateBucketInput) s3.CreateBucketRequest
	MockDeleteBucketRequest func(input *s3.DeleteBucketInput) s3.DeleteBucketRequest

	MockPutBucketEncryptionRequest    func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest
	MockGetBucketEncryptionRequest    func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest
	MockDeleteBucketEncryptionRequest func(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest

	MockPutBucketVersioningRequest func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest
	MockGetBucketVersioningRequest func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest

	MockPutBucketAccelerateConfigurationRequest func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest
	MockGetBucketAccelerateConfigurationRequest func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest

	MockPutBucketCorsRequest    func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest
	MockGetBucketCorsRequest    func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest
	MockDeleteBucketCorsRequest func(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest

	MockPutBucketWebsiteRequest    func(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest
	MockGetBucketWebsiteRequest    func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest
	MockDeleteBucketWebsiteRequest func(input *s3.DeleteBucketWebsiteInput) s3.DeleteBucketWebsiteRequest

	MockPutBucketLoggingRequest func(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest
	MockGetBucketLoggingRequest func(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest

	MockPutBucketReplicationRequest    func(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest
	MockGetBucketReplicationRequest    func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest
	MockDeleteBucketReplicationRequest func(input *s3.DeleteBucketReplicationInput) s3.DeleteBucketReplicationRequest

	MockPutBucketRequestPaymentRequest func(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest
	MockGetBucketRequestPaymentRequest func(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest

	MockPutBucketTaggingRequest    func(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest
	MockGetBucketTaggingRequest    func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest
	MockDeleteBucketTaggingRequest func(input *s3.DeleteBucketTaggingInput) s3.DeleteBucketTaggingRequest

	MockPutBucketAnalyticsConfigurationRequest func(input *s3.PutBucketAnalyticsConfigurationInput) s3.PutBucketAnalyticsConfigurationRequest
	MockGetBucketAnalyticsConfigurationRequest func(input *s3.GetBucketAnalyticsConfigurationInput) s3.GetBucketAnalyticsConfigurationRequest

	MockPutBucketLifecycleConfigurationRequest func(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest
	MockGetBucketLifecycleConfigurationRequest func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest
	MockDeleteBucketLifecycleRequest           func(input *s3.DeleteBucketLifecycleInput) s3.DeleteBucketLifecycleRequest

	MockPutBucketNotificationConfigurationRequest func(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest
	MockGetBucketNotificationConfigurationRequest func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest

	MockGetBucketAclRequest func(*s3.GetBucketAclInput) s3.GetBucketAclRequest //nolint
	MockPutBucketAclRequest func(*s3.PutBucketAclInput) s3.PutBucketAclRequest //nolint

	MockGetPublicAccessBlockRequest    func(*s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest
	MockPutPublicAccessBlockRequest    func(*s3.PutPublicAccessBlockInput) s3.PutPublicAccessBlockRequest
	MockDeletePublicAccessBlockRequest func(*s3.DeletePublicAccessBlockInput) s3.DeletePublicAccessBlockRequest
}

// HeadBucketRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) HeadBucketRequest(input *s3.HeadBucketInput) s3.HeadBucketRequest {
	return m.MockHeadBucketRequest(input)
}

// CreateBucketRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) CreateBucketRequest(input *s3.CreateBucketInput) s3.CreateBucketRequest {
	return m.MockCreateBucketRequest(input)
}

// DeleteBucketRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketRequest(input *s3.DeleteBucketInput) s3.DeleteBucketRequest {
	return m.MockDeleteBucketRequest(input)
}

// PutBucketEncryptionRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketEncryptionRequest(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
	return m.MockPutBucketEncryptionRequest(input)
}

// GetBucketEncryptionRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketEncryptionRequest(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
	return m.MockGetBucketEncryptionRequest(input)
}

// DeleteBucketEncryptionRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketEncryptionRequest(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest {
	return m.MockDeleteBucketEncryptionRequest(input)
}

// PutBucketVersioningRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketVersioningRequest(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
	return m.MockPutBucketVersioningRequest(input)
}

// GetBucketVersioningRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketVersioningRequest(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
	return m.MockGetBucketVersioningRequest(input)
}

// PutBucketAccelerateConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAccelerateConfigurationRequest(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
	return m.MockPutBucketAccelerateConfigurationRequest(input)
}

// GetBucketAccelerateConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAccelerateConfigurationRequest(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
	return m.MockGetBucketAccelerateConfigurationRequest(input)
}

// PutBucketCorsRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketCorsRequest(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
	return m.MockPutBucketCorsRequest(input)
}

// GetBucketCorsRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketCorsRequest(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
	return m.MockGetBucketCorsRequest(input)
}

// DeleteBucketCorsRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketCorsRequest(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest {
	return m.MockDeleteBucketCorsRequest(input)
}

// PutBucketWebsiteRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketWebsiteRequest(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest {
	return m.MockPutBucketWebsiteRequest(input)
}

// GetBucketWebsiteRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketWebsiteRequest(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
	return m.MockGetBucketWebsiteRequest(input)
}

// DeleteBucketWebsiteRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketWebsiteRequest(input *s3.DeleteBucketWebsiteInput) s3.DeleteBucketWebsiteRequest {
	return m.MockDeleteBucketWebsiteRequest(input)
}

// PutBucketLoggingRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketLoggingRequest(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest {
	return m.MockPutBucketLoggingRequest(input)
}

// GetBucketLoggingRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketLoggingRequest(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest {
	return m.MockGetBucketLoggingRequest(input)
}

// PutBucketReplicationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketReplicationRequest(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest {
	return m.MockPutBucketReplicationRequest(input)
}

// GetBucketReplicationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketReplicationRequest(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
	return m.MockGetBucketReplicationRequest(input)
}

// DeleteBucketReplicationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketReplicationRequest(input *s3.DeleteBucketReplicationInput) s3.DeleteBucketReplicationRequest {
	return m.MockDeleteBucketReplicationRequest(input)
}

// PutBucketRequestPaymentRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketRequestPaymentRequest(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest {
	return m.MockPutBucketRequestPaymentRequest(input)
}

// GetBucketRequestPaymentRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketRequestPaymentRequest(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest {
	return m.MockGetBucketRequestPaymentRequest(input)
}

// PutBucketTaggingRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketTaggingRequest(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest {
	return m.MockPutBucketTaggingRequest(input)
}

// GetBucketTaggingRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketTaggingRequest(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
	return m.MockGetBucketTaggingRequest(input)
}

// DeleteBucketTaggingRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketTaggingRequest(input *s3.DeleteBucketTaggingInput) s3.DeleteBucketTaggingRequest {
	return m.MockDeleteBucketTaggingRequest(input)
}

// PutBucketAnalyticsConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAnalyticsConfigurationRequest(input *s3.PutBucketAnalyticsConfigurationInput) s3.PutBucketAnalyticsConfigurationRequest {
	return m.MockPutBucketAnalyticsConfigurationRequest(input)
}

// GetBucketAnalyticsConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAnalyticsConfigurationRequest(input *s3.GetBucketAnalyticsConfigurationInput) s3.GetBucketAnalyticsConfigurationRequest {
	return m.MockGetBucketAnalyticsConfigurationRequest(input)
}

// PutBucketLifecycleConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketLifecycleConfigurationRequest(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest {
	return m.MockPutBucketLifecycleConfigurationRequest(input)
}

// GetBucketLifecycleConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketLifecycleConfigurationRequest(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
	return m.MockGetBucketLifecycleConfigurationRequest(input)
}

// DeleteBucketLifecycleRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeleteBucketLifecycleRequest(input *s3.DeleteBucketLifecycleInput) s3.DeleteBucketLifecycleRequest {
	return m.MockDeleteBucketLifecycleRequest(input)
}

// PutBucketNotificationConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketNotificationConfigurationRequest(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest {
	return m.MockPutBucketNotificationConfigurationRequest(input)
}

// GetBucketNotificationConfigurationRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketNotificationConfigurationRequest(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
	return m.MockGetBucketNotificationConfigurationRequest(input)
}

// GetBucketAclRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetBucketAclRequest(input *s3.GetBucketAclInput) s3.GetBucketAclRequest { //nolint
	return m.MockGetBucketAclRequest(input)
}

// PutBucketAclRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutBucketAclRequest(input *s3.PutBucketAclInput) s3.PutBucketAclRequest { //nolint
	return m.MockPutBucketAclRequest(input)
}

// GetPublicAccessBlockRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) GetPublicAccessBlockRequest(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
	return m.MockGetPublicAccessBlockRequest(input)
}

// PutPublicAccessBlockRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) PutPublicAccessBlockRequest(input *s3.PutPublicAccessBlockInput) s3.PutPublicAccessBlockRequest {
	return m.MockPutPublicAccessBlockRequest(input)
}

// DeletePublicAccessBlockRequest is the fake method call to invoke the internal mock method
func (m MockBucketClient) DeletePublicAccessBlockRequest(input *s3.DeletePublicAccessBlockInput) s3.DeletePublicAccessBlockRequest {
	return m.MockDeletePublicAccessBlockRequest(input)
}
