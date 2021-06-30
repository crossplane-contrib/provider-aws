package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// MockEC2Client mock ec2 client
type MockEC2Client struct {
	// DescribeVpcPeeringConnectionsRequestFun
	DescribeVpcPeeringConnectionsRequestFun func(input *ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest
	// CreateVpcPeeringConnectionRequestFun
	CreateVpcPeeringConnectionRequestFun func(input *ec2.CreateVpcPeeringConnectionInput) ec2.CreateVpcPeeringConnectionRequest
	// CreateRouteRequestFun
	CreateRouteRequestFun func(input *ec2.CreateRouteInput) ec2.CreateRouteRequest
	// DescribeRouteTablesRequestFun
	DescribeRouteTablesRequestFun func(input *ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest
	// DeleteRouteRequestFun
	DeleteRouteRequestFun func(input *ec2.DeleteRouteInput) ec2.DeleteRouteRequest
	// ModifyVpcPeeringConnectionOptionsRequestFun
	ModifyVpcPeeringConnectionOptionsRequestFun func(input *ec2.ModifyVpcPeeringConnectionOptionsInput) ec2.ModifyVpcPeeringConnectionOptionsRequest
	// DeleteVpcPeeringConnectionRequestFun
	DeleteVpcPeeringConnectionRequestFun func(input *ec2.DeleteVpcPeeringConnectionInput) ec2.DeleteVpcPeeringConnectionRequest
	// CreateTagsRequestFun
	CreateTagsRequestFun func(input *ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// CreateRouteRequest create route request
func (m *MockEC2Client) CreateRouteRequest(input *ec2.CreateRouteInput) ec2.CreateRouteRequest {
	return m.CreateRouteRequestFun(input)
}

// DescribeVpcPeeringConnectionsRequest describe vpc peering connection
func (m *MockEC2Client) DescribeVpcPeeringConnectionsRequest(input *ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest {
	return m.DescribeVpcPeeringConnectionsRequestFun(input)
}

// CreateVpcPeeringConnectionRequest create vpc peering connection.
func (m *MockEC2Client) CreateVpcPeeringConnectionRequest(input *ec2.CreateVpcPeeringConnectionInput) ec2.CreateVpcPeeringConnectionRequest {
	return m.CreateVpcPeeringConnectionRequestFun(input)
}

// DescribeRouteTablesRequest describe route table.
func (m *MockEC2Client) DescribeRouteTablesRequest(input *ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest {
	return m.DescribeRouteTablesRequestFun(input)
}

// DeleteRouteRequest delete route.
func (m *MockEC2Client) DeleteRouteRequest(input *ec2.DeleteRouteInput) ec2.DeleteRouteRequest {
	return m.DeleteRouteRequestFun(input)
}

// ModifyVpcPeeringConnectionOptionsRequest modify vpc peering
func (m *MockEC2Client) ModifyVpcPeeringConnectionOptionsRequest(input *ec2.ModifyVpcPeeringConnectionOptionsInput) ec2.ModifyVpcPeeringConnectionOptionsRequest {
	return m.ModifyVpcPeeringConnectionOptionsRequestFun(input)
}

// DeleteVpcPeeringConnectionRequest delete vpc peering
func (m *MockEC2Client) DeleteVpcPeeringConnectionRequest(input *ec2.DeleteVpcPeeringConnectionInput) ec2.DeleteVpcPeeringConnectionRequest {
	return m.DeleteVpcPeeringConnectionRequestFun(input)
}

// CreateTagsRequest create tags
func (m *MockEC2Client) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.CreateTagsRequestFun(input)
}

// MockRoute53Client route53 client
type MockRoute53Client struct {
	// CreateVPCAssociationAuthorizationRequestFun mock create vpc AssociationAuthorization
	CreateVPCAssociationAuthorizationRequestFun func(input *route53.CreateVPCAssociationAuthorizationInput) route53.CreateVPCAssociationAuthorizationRequest
	// DeleteVPCAssociationAuthorizationRequestFun mock delete vpc AssociationAuthorization
	DeleteVPCAssociationAuthorizationRequestFun func(input *route53.DeleteVPCAssociationAuthorizationInput) route53.DeleteVPCAssociationAuthorizationRequest
}

// CreateVPCAssociationAuthorizationRequest create AssociationAuthorization
func (m *MockRoute53Client) CreateVPCAssociationAuthorizationRequest(input *route53.CreateVPCAssociationAuthorizationInput) route53.CreateVPCAssociationAuthorizationRequest {
	return m.CreateVPCAssociationAuthorizationRequestFun(input)
}

// DeleteVPCAssociationAuthorizationRequest delete AssociationAuthorization
func (m *MockRoute53Client) DeleteVPCAssociationAuthorizationRequest(input *route53.DeleteVPCAssociationAuthorizationInput) route53.DeleteVPCAssociationAuthorizationRequest {
	return m.DeleteVPCAssociationAuthorizationRequestFun(input)
}
