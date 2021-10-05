package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.NatGatewayClient = (*MockNatGatewayClient)(nil)

// MockNatGatewayClient is a type that implements all the methods for NatGatewayClient interface
type MockNatGatewayClient struct {
	MockCreate     func(ctx context.Context, input *ec2.CreateNatGatewayInput, opts []func(*ec2.Options)) (*ec2.CreateNatGatewayOutput, error)
	MockDelete     func(ctx context.Context, input *ec2.DeleteNatGatewayInput, opts []func(*ec2.Options)) (*ec2.DeleteNatGatewayOutput, error)
	MockDescribe   func(ctx context.Context, input *ec2.DescribeNatGatewaysInput, opts []func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	MockCreateTags func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	MockDeleteTags func(ctx context.Context, input *ec2.DeleteTagsInput, opts []func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// CreateNatGateway mocks CreateNatGateway method
func (m *MockNatGatewayClient) CreateNatGateway(ctx context.Context, input *ec2.CreateNatGatewayInput, opts ...func(*ec2.Options)) (*ec2.CreateNatGatewayOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteNatGateway mocks DeleteNatGateway method
func (m *MockNatGatewayClient) DeleteNatGateway(ctx context.Context, input *ec2.DeleteNatGatewayInput, opts ...func(*ec2.Options)) (*ec2.DeleteNatGatewayOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeNatGateways mocks DescribeNatGateways method
func (m *MockNatGatewayClient) DescribeNatGateways(ctx context.Context, input *ec2.DescribeNatGatewaysInput, opts ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockNatGatewayClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}

// DeleteTags mocks DeleteTags method
func (m *MockNatGatewayClient) DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	return m.MockDeleteTags(ctx, input, opts)
}
