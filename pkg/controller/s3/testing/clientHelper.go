package testing

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	clients3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
)

// Client creates a MockBucketClient with default request functions and an optional list of
// ClientModifiers
func Client(m ...ClientModifier) *fake.MockBucketClient {
	client := &fake.MockBucketClient{
		MockHeadBucket: func(ctx context.Context, input *awss3.HeadBucketInput, opts []func(*awss3.Options)) (*awss3.HeadBucketOutput, error) {
			return &awss3.HeadBucketOutput{}, nil
		},
		MockCreateBucket: func(ctx context.Context, input *awss3.CreateBucketInput, opts []func(*awss3.Options)) (*awss3.CreateBucketOutput, error) {
			return &awss3.CreateBucketOutput{}, nil
		},
		MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *awss3.GetBucketAccelerateConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetBucketAccelerateConfigurationOutput, error) {
			return &awss3.GetBucketAccelerateConfigurationOutput{}, nil
		},
		MockGetBucketCors: func(ctx context.Context, input *awss3.GetBucketCorsInput, opts []func(*awss3.Options)) (*awss3.GetBucketCorsOutput, error) {
			return &awss3.GetBucketCorsOutput{}, &smithy.GenericAPIError{Code: clients3.CORSNotFoundErrCode}
		},
		MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *awss3.GetBucketLifecycleConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetBucketLifecycleConfigurationOutput, error) {
			return &awss3.GetBucketLifecycleConfigurationOutput{}, &smithy.GenericAPIError{Code: clients3.LifecycleNotFoundErrCode}
		},
		MockGetBucketLogging: func(ctx context.Context, input *awss3.GetBucketLoggingInput, opts []func(*awss3.Options)) (*awss3.GetBucketLoggingOutput, error) {
			return &awss3.GetBucketLoggingOutput{}, nil
		},
		MockGetBucketNotificationConfiguration: func(ctx context.Context, input *awss3.GetBucketNotificationConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetBucketNotificationConfigurationOutput, error) {
			return &awss3.GetBucketNotificationConfigurationOutput{}, nil
		},
		MockGetBucketReplication: func(ctx context.Context, input *awss3.GetBucketReplicationInput, opts []func(*awss3.Options)) (*awss3.GetBucketReplicationOutput, error) {
			return nil, &smithy.GenericAPIError{Code: clients3.ReplicationNotFoundErrCode}
		},
		MockGetBucketRequestPayment: func(ctx context.Context, input *awss3.GetBucketRequestPaymentInput, opts []func(*awss3.Options)) (*awss3.GetBucketRequestPaymentOutput, error) {
			return &awss3.GetBucketRequestPaymentOutput{}, nil
		},
		MockGetBucketEncryption: func(ctx context.Context, input *awss3.GetBucketEncryptionInput, opts []func(*awss3.Options)) (*awss3.GetBucketEncryptionOutput, error) {
			return nil, &smithy.GenericAPIError{Code: s3.SSENotFoundErrCode}
		},
		MockGetBucketTagging: func(ctx context.Context, input *awss3.GetBucketTaggingInput, opts []func(*awss3.Options)) (*awss3.GetBucketTaggingOutput, error) {
			return nil, &smithy.GenericAPIError{Code: clients3.TaggingNotFoundErrCode}
		},
		MockGetBucketVersioning: func(ctx context.Context, input *awss3.GetBucketVersioningInput, opts []func(*awss3.Options)) (*awss3.GetBucketVersioningOutput, error) {
			return &awss3.GetBucketVersioningOutput{}, nil
		},
		MockGetBucketWebsite: func(ctx context.Context, input *awss3.GetBucketWebsiteInput, opts []func(*awss3.Options)) (*awss3.GetBucketWebsiteOutput, error) {
			return nil, &smithy.GenericAPIError{Code: clients3.WebsiteNotFoundErrCode}
		},
		MockPutBucketAcl: func(ctx context.Context, input *awss3.PutBucketAclInput, opts []func(*awss3.Options)) (*awss3.PutBucketAclOutput, error) {
			return &awss3.PutBucketAclOutput{}, nil
		},
		MockGetPublicAccessBlock: func(ctx context.Context, input *awss3.GetPublicAccessBlockInput, opts []func(*awss3.Options)) (*awss3.GetPublicAccessBlockOutput, error) {
			return &awss3.GetPublicAccessBlockOutput{}, nil
		},
		MockPutPublicAccessBlock: func(ctx context.Context, input *awss3.PutPublicAccessBlockInput, opts []func(*awss3.Options)) (*awss3.PutPublicAccessBlockOutput, error) {
			return &awss3.PutPublicAccessBlockOutput{}, nil
		},
		MockDeletePublicAccessBlock: func(ctx context.Context, input *awss3.DeletePublicAccessBlockInput, opts []func(*awss3.Options)) (*awss3.DeletePublicAccessBlockOutput, error) {
			return &awss3.DeletePublicAccessBlockOutput{}, nil
		},
		MockPutBucketOwnershipControls: func(ctx context.Context, input *awss3.PutBucketOwnershipControlsInput, opts []func(*awss3.Options)) (*awss3.PutBucketOwnershipControlsOutput, error) {
			return &awss3.PutBucketOwnershipControlsOutput{}, nil
		},
		MockDeleteBucketOwnershipControls: func(ctx context.Context, input *awss3.DeleteBucketOwnershipControlsInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketOwnershipControlsOutput, error) {
			return &awss3.DeleteBucketOwnershipControlsOutput{}, nil
		},
		MockBucketPolicyClient: fake.MockBucketPolicyClient{
			MockGetBucketPolicy: func(ctx context.Context, input *awss3.GetBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.GetBucketPolicyOutput, error) {
				return &awss3.GetBucketPolicyOutput{}, nil
			},
			MockPutBucketPolicy: func(ctx context.Context, input *awss3.PutBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.PutBucketPolicyOutput, error) {
				return &awss3.PutBucketPolicyOutput{}, nil
			},
			MockDeleteBucketPolicy: func(ctx context.Context, input *awss3.DeleteBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketPolicyOutput, error) {
				return &awss3.DeleteBucketPolicyOutput{}, nil
			},
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
func WithGetRequestPayment(input func(ctx context.Context, input *awss3.GetBucketRequestPaymentInput, opts []func(*awss3.Options)) (*awss3.GetBucketRequestPaymentOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockGetBucketRequestPayment = input
	}
}

// WithCreateBucket sets the MockCreateBucketRequest of the mock S3 Client
func WithCreateBucket(input func(ctx context.Context, input *awss3.CreateBucketInput, opts []func(*awss3.Options)) (*awss3.CreateBucketOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockCreateBucket = input
	}
}

// WithPutRequestPayment sets the MockPutBucketRequestPaymentRequest of the mock S3 Client
func WithPutRequestPayment(input func(ctx context.Context, input *awss3.PutBucketRequestPaymentInput, opts []func(*awss3.Options)) (*awss3.PutBucketRequestPaymentOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockPutBucketRequestPayment = input
	}
}

// WithGetSSE sets the MockGetBucketEncryptionRequest of the mock S3 Client
func WithGetSSE(input func(ctx context.Context, input *awss3.GetBucketEncryptionInput, opts []func(*awss3.Options)) (*awss3.GetBucketEncryptionOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockGetBucketEncryption = input
	}
}

// WithDeleteSSE sets the MockDeleteBucketEncryptionRequest of the mock S3 Client
func WithDeleteSSE(input func(ctx context.Context, input *awss3.DeleteBucketEncryptionInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketEncryptionOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockDeleteBucketEncryption = input
	}
}

// WithPutACL sets the MockPutBucketAclRequest of the mock S3 Client
func WithPutACL(input func(ctx context.Context, input *awss3.PutBucketAclInput, opts []func(*awss3.Options)) (*awss3.PutBucketAclOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockPutBucketAcl = input
	}
}

// WithPutOwnershipControls sets the MockPutBucketOwnershipControlsRequest of the mock S3 Client
func WithPutOwnershipControls(input func(ctx context.Context, input *awss3.PutBucketOwnershipControlsInput, opts []func(*awss3.Options)) (*awss3.PutBucketOwnershipControlsOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockPutBucketOwnershipControls = input
	}
}

// WithDeleteOwnershipControls sets the MockDeleteBucketOwnershipControlsRequest of the mock S3 Client
func WithDeleteOwnershipControls(input func(ctx context.Context, input *awss3.DeleteBucketOwnershipControlsInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketOwnershipControlsOutput, error)) ClientModifier {
	return func(client *fake.MockBucketClient) {
		client.MockDeleteBucketOwnershipControls = input
	}
}
