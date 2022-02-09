package fake

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// MockVPCEndpointClient for testing
type MockVPCEndpointClient struct {
	ec2iface.EC2API

	MockCreateVpcEndpointWithContext    func(context.Context, *ec2.CreateVpcEndpointInput, ...request.Option) (*ec2.CreateVpcEndpointOutput, error)
	MockDeleteVpcEndpoints              func(*ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error)
	MockModifyVpcEndpointWithContext    func(context.Context, *ec2.ModifyVpcEndpointInput, ...request.Option) (*ec2.ModifyVpcEndpointOutput, error)
	MockDescribeVpcEndpoints            func(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error)
	MockDescribeVpcEndpointsWithContext func(context.Context, *ec2.DescribeVpcEndpointsInput, ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error)
}

// CreateVpcEndpointWithContext mocks CreateVpcEndpointWithContext
func (m *MockVPCEndpointClient) CreateVpcEndpointWithContext(ctx context.Context, input *ec2.CreateVpcEndpointInput, req ...request.Option) (*ec2.CreateVpcEndpointOutput, error) {
	return m.MockCreateVpcEndpointWithContext(ctx, input)
}

// DeleteVpcEndpoints mocks DeleteVpcEndpoints
func (m *MockVPCEndpointClient) DeleteVpcEndpoints(input *ec2.DeleteVpcEndpointsInput) (*ec2.DeleteVpcEndpointsOutput, error) {
	return m.MockDeleteVpcEndpoints(input)
}

// ModifyVpcEndpointWithContext mocks ModifyVpcEndpointWithContext
func (m *MockVPCEndpointClient) ModifyVpcEndpointWithContext(ctx context.Context, input *ec2.ModifyVpcEndpointInput, req ...request.Option) (*ec2.ModifyVpcEndpointOutput, error) {
	return m.MockModifyVpcEndpointWithContext(ctx, input)
}

// DescribeVpcEndpoints mocks DescribeVpcEndpoints
func (m *MockVPCEndpointClient) DescribeVpcEndpoints(input *ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
	return m.MockDescribeVpcEndpoints(input)
}

// DescribeVpcEndpointsWithContext mocks DescribeVpcEndpointsWithContext
func (m *MockVPCEndpointClient) DescribeVpcEndpointsWithContext(ctx context.Context, input *ec2.DescribeVpcEndpointsInput, req ...request.Option) (*ec2.DescribeVpcEndpointsOutput, error) {
	return m.MockDescribeVpcEndpointsWithContext(ctx, input)
}
