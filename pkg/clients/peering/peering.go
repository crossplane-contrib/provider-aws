package peering

import (
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/crossplane/provider-aws/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/crossplane/provider-aws/apis/vpcpeering/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeVpcPeeringConnectionsInput returns input for read
// operation.
func GenerateDescribeVpcPeeringConnectionsInput(cr *svcapitypes.VPCPeeringConnection) *ec2.DescribeVpcPeeringConnectionsInput {
	res := &ec2.DescribeVpcPeeringConnectionsInput{
		Filters: []ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{cr.ObjectMeta.Name},
			},
		},
	}

	return res
}

// GenerateVPCPeeringConnection returns the current state in the form of *svcapitypes.VPCPeeringConnection.
func GenerateVPCPeeringConnection(resp *ec2.DescribeVpcPeeringConnectionsResponse) *svcapitypes.VPCPeeringConnection {
	cr := &svcapitypes.VPCPeeringConnection{}

	output := resp.DescribeVpcPeeringConnectionsOutput
	if output == nil {
		return cr
	}

	found := false
	for _, elem := range output.VpcPeeringConnections {
		if elem.AccepterVpcInfo != nil {
			f0 := &svcapitypes.VPCPeeringConnectionVPCInfo{}
			if elem.AccepterVpcInfo.CidrBlock != nil {
				f0.CIDRBlock = elem.AccepterVpcInfo.CidrBlock
			}
			if elem.AccepterVpcInfo.CidrBlockSet != nil {
				f0f1 := []*svcapitypes.CIDRBlock{}
				for _, f0f1iter := range elem.AccepterVpcInfo.CidrBlockSet {
					f0f1elem := &svcapitypes.CIDRBlock{}
					if f0f1iter.CidrBlock != nil {
						f0f1elem.CIDRBlock = f0f1iter.CidrBlock
					}
					f0f1 = append(f0f1, f0f1elem)
				}
				f0.CIDRBlockSet = f0f1
			}
			if elem.AccepterVpcInfo.Ipv6CidrBlockSet != nil {
				f0f2 := []*svcapitypes.IPv6CIDRBlock{}
				for _, f0f2iter := range elem.AccepterVpcInfo.Ipv6CidrBlockSet {
					f0f2elem := &svcapitypes.IPv6CIDRBlock{}
					if f0f2iter.Ipv6CidrBlock != nil {
						f0f2elem.IPv6CIDRBlock = f0f2iter.Ipv6CidrBlock
					}
					f0f2 = append(f0f2, f0f2elem)
				}
				f0.IPv6CIDRBlockSet = f0f2
			}
			if elem.AccepterVpcInfo.OwnerId != nil {
				f0.OwnerID = elem.AccepterVpcInfo.OwnerId
			}
			if elem.AccepterVpcInfo.PeeringOptions != nil {
				f0f4 := &svcapitypes.VPCPeeringConnectionOptionsDescription{}
				if elem.AccepterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc != nil {
					f0f4.AllowDNSResolutionFromRemoteVPC = elem.AccepterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc
				}
				if elem.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc != nil {
					f0f4.AllowEgressFromLocalClassicLinkToRemoteVPC = elem.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc
				}
				if elem.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink != nil {
					f0f4.AllowEgressFromLocalVPCToRemoteClassicLink = elem.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink
				}
				f0.PeeringOptions = f0f4
			}
			if elem.AccepterVpcInfo.Region != nil {
				f0.Region = elem.AccepterVpcInfo.Region
			}
			if elem.AccepterVpcInfo.VpcId != nil {
				f0.VPCID = elem.AccepterVpcInfo.VpcId
			}
			cr.Status.AtProvider.AccepterVPCInfo = f0
		} else {
			cr.Status.AtProvider.AccepterVPCInfo = nil
		}
		if elem.ExpirationTime != nil {
			cr.Status.AtProvider.ExpirationTime = &metav1.Time{*elem.ExpirationTime}
		} else {
			cr.Status.AtProvider.ExpirationTime = nil
		}
		if elem.RequesterVpcInfo != nil {
			f2 := &svcapitypes.VPCPeeringConnectionVPCInfo{}
			if elem.RequesterVpcInfo.CidrBlock != nil {
				f2.CIDRBlock = elem.RequesterVpcInfo.CidrBlock
			}
			if elem.RequesterVpcInfo.CidrBlockSet != nil {
				f2f1 := []*svcapitypes.CIDRBlock{}
				for _, f2f1iter := range elem.RequesterVpcInfo.CidrBlockSet {
					f2f1elem := &svcapitypes.CIDRBlock{}
					if f2f1iter.CidrBlock != nil {
						f2f1elem.CIDRBlock = f2f1iter.CidrBlock
					}
					f2f1 = append(f2f1, f2f1elem)
				}
				f2.CIDRBlockSet = f2f1
			}
			if elem.RequesterVpcInfo.Ipv6CidrBlockSet != nil {
				f2f2 := []*svcapitypes.IPv6CIDRBlock{}
				for _, f2f2iter := range elem.RequesterVpcInfo.Ipv6CidrBlockSet {
					f2f2elem := &svcapitypes.IPv6CIDRBlock{}
					if f2f2iter.Ipv6CidrBlock != nil {
						f2f2elem.IPv6CIDRBlock = f2f2iter.Ipv6CidrBlock
					}
					f2f2 = append(f2f2, f2f2elem)
				}
				f2.IPv6CIDRBlockSet = f2f2
			}
			if elem.RequesterVpcInfo.OwnerId != nil {
				f2.OwnerID = elem.RequesterVpcInfo.OwnerId
			}
			if elem.RequesterVpcInfo.PeeringOptions != nil {
				f2f4 := &svcapitypes.VPCPeeringConnectionOptionsDescription{}
				if elem.RequesterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc != nil {
					f2f4.AllowDNSResolutionFromRemoteVPC = elem.RequesterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc
				}
				if elem.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc != nil {
					f2f4.AllowEgressFromLocalClassicLinkToRemoteVPC = elem.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc
				}
				if elem.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink != nil {
					f2f4.AllowEgressFromLocalVPCToRemoteClassicLink = elem.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink
				}
				f2.PeeringOptions = f2f4
			}
			if elem.RequesterVpcInfo.Region != nil {
				f2.Region = elem.RequesterVpcInfo.Region
			}
			if elem.RequesterVpcInfo.VpcId != nil {
				f2.VPCID = elem.RequesterVpcInfo.VpcId
			}
			cr.Status.AtProvider.RequesterVPCInfo = f2
		} else {
			cr.Status.AtProvider.RequesterVPCInfo = nil
		}
		if elem.Status != nil {
			f3 := &svcapitypes.VPCPeeringConnectionStateReason{}
			f3.Code = aws.String(string(elem.Status.Code))
			if elem.Status.Message != nil {
				f3.Message = elem.Status.Message
			}
			cr.Status.AtProvider.Status = f3
		} else {
			cr.Status.AtProvider.Status = nil
		}
		if elem.Tags != nil {
			f4 := []*svcapitypes.Tag{}
			for _, f4iter := range elem.Tags {
				f4elem := &svcapitypes.Tag{}
				if f4iter.Key != nil {
					f4elem.Key = f4iter.Key
				}
				if f4iter.Value != nil {
					f4elem.Value = f4iter.Value
				}
				f4 = append(f4, f4elem)
			}
			cr.Status.AtProvider.Tags = f4
		} else {
			cr.Status.AtProvider.Tags = nil
		}
		if elem.VpcPeeringConnectionId != nil {
			cr.Status.AtProvider.VPCPeeringConnectionID = elem.VpcPeeringConnectionId
		} else {
			cr.Status.AtProvider.VPCPeeringConnectionID = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateVpcPeeringConnectionInput returns a create input.
func GenerateCreateVpcPeeringConnectionInput(cr *svcapitypes.VPCPeeringConnection) *ec2.CreateVpcPeeringConnectionInput {
	res := &ec2.CreateVpcPeeringConnectionInput{}

	if cr.Spec.ForProvider.PeerOwnerID != nil {
		res.PeerOwnerId = cr.Spec.ForProvider.PeerOwnerID
	}
	if cr.Spec.ForProvider.PeerRegion != nil {
		res.PeerRegion = cr.Spec.ForProvider.PeerRegion
	}
	if cr.Spec.ForProvider.PeerVPCID != nil {
		res.PeerVpcId = cr.Spec.ForProvider.PeerVPCID
	}
	if cr.Spec.ForProvider.VPCID != nil {
		res.VpcId = cr.Spec.ForProvider.VPCID
	}

	return res
}

type EC2Client interface {
	DescribeVpcPeeringConnectionsRequest(*ec2.DescribeVpcPeeringConnectionsInput) ec2.DescribeVpcPeeringConnectionsRequest
	CreateVpcPeeringConnectionRequest(*ec2.CreateVpcPeeringConnectionInput) ec2.CreateVpcPeeringConnectionRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	DescribeRouteTablesRequest(*ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest
	CreateRouteRequest(*ec2.CreateRouteInput) ec2.CreateRouteRequest
	DeleteRouteRequest(*ec2.DeleteRouteInput) ec2.DeleteRouteRequest
	ModifyVpcPeeringConnectionOptionsRequest(*ec2.ModifyVpcPeeringConnectionOptionsInput) ec2.ModifyVpcPeeringConnectionOptionsRequest
	DeleteVpcPeeringConnectionRequest(*ec2.DeleteVpcPeeringConnectionInput) ec2.DeleteVpcPeeringConnectionRequest
}

func NewEc2Client(cfg sdkaws.Config) EC2Client {
	return ec2.New(cfg)
}

func NewRoute53Client(cfg sdkaws.Config) Route53Client {
	return route53.New(cfg)
}

type Route53Client interface {
	CreateVPCAssociationAuthorizationRequest(*route53.CreateVPCAssociationAuthorizationInput) route53.CreateVPCAssociationAuthorizationRequest
	DeleteVPCAssociationAuthorizationRequest(*route53.DeleteVPCAssociationAuthorizationInput) route53.DeleteVPCAssociationAuthorizationRequest
}