package testing

import (
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

// Client creates a MockBucketClient with default request functions and an optional list of
// ClientModifiers
func Client(m ...ClientModifier) *fake.MockBucketClient {
	client := &fake.MockBucketClient{
		MockCreateBucketRequest: func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
			return awss3.CreateBucketRequest{
				Request: CreateRequest(nil, &awss3.CreateBucketOutput{}),
			}
		},
		MockHeadBucketRequest: func(input *awss3.HeadBucketInput) awss3.HeadBucketRequest {
			return awss3.HeadBucketRequest{
				Request: CreateRequest(nil, &awss3.HeadBucketOutput{}),
			}
		},
		MockGetBucketAccelerateConfigurationRequest: func(input *awss3.GetBucketAccelerateConfigurationInput) awss3.GetBucketAccelerateConfigurationRequest {
			return awss3.GetBucketAccelerateConfigurationRequest{
				Request: CreateRequest(nil, &awss3.GetBucketAccelerateConfigurationOutput{}),
			}
		},
		MockGetBucketCorsRequest: func(input *awss3.GetBucketCorsInput) awss3.GetBucketCorsRequest {
			return awss3.GetBucketCorsRequest{
				Request: CreateRequest(awserr.New(s3.CORSNotFoundErrCode, "", nil), &awss3.GetBucketCorsOutput{}),
			}
		},
		MockGetBucketLifecycleConfigurationRequest: func(input *awss3.GetBucketLifecycleConfigurationInput) awss3.GetBucketLifecycleConfigurationRequest {
			return awss3.GetBucketLifecycleConfigurationRequest{
				Request: CreateRequest(awserr.New(s3.LifecycleNotFoundErrCode, "", nil), &awss3.GetBucketLifecycleConfigurationOutput{}),
			}
		},
		MockGetBucketLoggingRequest: func(input *awss3.GetBucketLoggingInput) awss3.GetBucketLoggingRequest {
			return awss3.GetBucketLoggingRequest{
				Request: CreateRequest(nil, &awss3.GetBucketLoggingOutput{}),
			}
		},
		MockGetBucketNotificationConfigurationRequest: func(input *awss3.GetBucketNotificationConfigurationInput) awss3.GetBucketNotificationConfigurationRequest {
			return awss3.GetBucketNotificationConfigurationRequest{
				Request: CreateRequest(nil, &awss3.GetBucketNotificationConfigurationOutput{}),
			}
		},
		MockGetBucketReplicationRequest: func(input *awss3.GetBucketReplicationInput) awss3.GetBucketReplicationRequest {
			return awss3.GetBucketReplicationRequest{
				Request: CreateRequest(awserr.New(s3.ReplicationNotFoundErrCode, "", nil), &awss3.GetBucketReplicationOutput{}),
			}
		},
		MockGetBucketRequestPaymentRequest: func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
			return awss3.GetBucketRequestPaymentRequest{
				Request: CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{}),
			}
		},
		MockGetBucketEncryptionRequest: func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
			return awss3.GetBucketEncryptionRequest{
				Request: CreateRequest(awserr.New(s3.SSENotFoundErrCode, "", nil), &awss3.GetBucketEncryptionOutput{}),
			}
		},
		MockGetBucketTaggingRequest: func(input *awss3.GetBucketTaggingInput) awss3.GetBucketTaggingRequest {
			return awss3.GetBucketTaggingRequest{
				Request: CreateRequest(awserr.New(s3.TaggingNotFoundErrCode, "", nil), &awss3.GetBucketTaggingOutput{}),
			}
		},
		MockGetBucketVersioningRequest: func(input *awss3.GetBucketVersioningInput) awss3.GetBucketVersioningRequest {
			return awss3.GetBucketVersioningRequest{
				Request: CreateRequest(nil, &awss3.GetBucketVersioningOutput{}),
			}
		},
		MockGetBucketWebsiteRequest: func(input *awss3.GetBucketWebsiteInput) awss3.GetBucketWebsiteRequest {
			return awss3.GetBucketWebsiteRequest{
				Request: CreateRequest(awserr.New(s3.WebsiteNotFoundErrCode, "", nil), &awss3.GetBucketWebsiteOutput{}),
			}
		},
		MockPutBucketAclRequest: func(input *awss3.PutBucketAclInput) awss3.PutBucketAclRequest {
			return awss3.PutBucketAclRequest{
				Request: CreateRequest(nil, &awss3.PutBucketAclOutput{}),
			}
		},
		MockGetPublicAccessBlockRequest: func(input *awss3.GetPublicAccessBlockInput) awss3.GetPublicAccessBlockRequest {
			return awss3.GetPublicAccessBlockRequest{
				Request: CreateRequest(awserr.New(s3.PublicAccessBlockNotFoundErrCode, "error", nil), &awss3.GetPublicAccessBlockOutput{}),
			}
		},
		MockPutPublicAccessBlockRequest: func(input *awss3.PutPublicAccessBlockInput) awss3.PutPublicAccessBlockRequest {
			return awss3.PutPublicAccessBlockRequest{
				Request: CreateRequest(nil, &awss3.PutPublicAccessBlockOutput{}),
			}
		},
		MockDeletePublicAccessBlockRequest: func(input *awss3.DeletePublicAccessBlockInput) awss3.DeletePublicAccessBlockRequest {
			return awss3.DeletePublicAccessBlockRequest{
				Request: CreateRequest(awserr.New(s3.PublicAccessBlockNotFoundErrCode, "error", nil), &awss3.DeletePublicAccessBlockOutput{}),
			}
		},
	}
	for _, v := range m {
		v(client)
	}
	return client
}

// ClientModifier is a function which modifies the S3 Client for testing
type ClientModifier func(client *fake.MockBucketClient)

// WithGetRequestPayment sets the MockGetBucketRequestPaymentRequest of the mock S3 Client
func WithGetRequestPayment(input func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockGetBucketRequestPaymentRequest = input
	}
}

// WithCreateBucket sets the MockCreateBucketRequest of the mock S3 Client
func WithCreateBucket(input func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockCreateBucketRequest = input
	}
}

// WithPutRequestPayment sets the MockPutBucketRequestPaymentRequest of the mock S3 Client
func WithPutRequestPayment(input func(input *awss3.PutBucketRequestPaymentInput) awss3.PutBucketRequestPaymentRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockPutBucketRequestPaymentRequest = input
	}
}

// WithGetSSE sets the MockGetBucketEncryptionRequest of the mock S3 Client
func WithGetSSE(input func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockGetBucketEncryptionRequest = input
	}
}

// WithDeleteSSE sets the MockDeleteBucketEncryptionRequest of the mock S3 Client
func WithDeleteSSE(input func(input *awss3.DeleteBucketEncryptionInput) awss3.DeleteBucketEncryptionRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockDeleteBucketEncryptionRequest = input
	}
}

// WithPutACL sets the MockPutBucketAclRequest of the mock S3 Client
func WithPutACL(input func(input *awss3.PutBucketAclInput) awss3.PutBucketAclRequest) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockPutBucketAclRequest = input
	}
}
