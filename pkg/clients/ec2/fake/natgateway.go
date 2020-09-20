package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.NatGatewayClient = (*MockNatGatewayClient)(nil)

// MockNatGatewayClient is a type that implements all the methods for NatGatewayClient interface
type MockNatGatewayClient struct {
	MockCreate     func(*ec2.CreateNatGatewayInput) ec2.CreateNatGatewayRequest
	MockDelete     func(*ec2.DeleteNatGatewayInput) ec2.DeleteNatGatewayRequest
	MockDescribe   func(*ec2.DescribeNatGatewaysInput) ec2.DescribeNatGatewaysRequest
	MockCreateTags func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// CreateNatGatewayRequest mocks CreateNatGatewayRequest method
func (m *MockNatGatewayClient) CreateNatGatewayRequest(input *ec2.CreateNatGatewayInput) ec2.CreateNatGatewayRequest {
	return m.MockCreate(input)
}

// DeleteNatGatewayRequest mocks DeleteNatGatewayRequest method
func (m *MockNatGatewayClient) DeleteNatGatewayRequest(input *ec2.DeleteNatGatewayInput) ec2.DeleteNatGatewayRequest {
	return m.MockDelete(input)
}

// DescribeNatGatewaysRequest mocks DescribeNatGatewaysRequest method
func (m *MockNatGatewayClient) DescribeNatGatewaysRequest(input *ec2.DescribeNatGatewaysInput) ec2.DescribeNatGatewaysRequest {
	return m.MockDescribe(input)
}

// CreateTagsRequest mocks CreateTagsRequest method
func (m *MockNatGatewayClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTags(input)
}
