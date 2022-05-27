package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

var (
	matchAssociationID = "test"
	otherAssociationID = "other"
	testCidrBlock      = "10.0.0.0/0"
	otherCidrBlock     = "10.0.0.0/16"
	testStatus         = "status"
	testStateString    = string(types.VpcCidrBlockStateCodeAssociated)
	testState          = types.VpcCidrBlockStateCodeAssociated
)

func TestGenerateVPCCIDRBlockObservation(t *testing.T) {
	cases := map[string]struct {
		associationID string
		in            types.Vpc
		out           v1beta1.VPCCIDRBlockObservation
	}{
		"IPv4": {
			associationID: matchAssociationID,
			in: types.Vpc{
				CidrBlockAssociationSet: []types.VpcCidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						CidrBlock:     &testCidrBlock,
						CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						CidrBlock:     &otherCidrBlock,
					},
				},
			},
			out: v1beta1.VPCCIDRBlockObservation{
				AssociationID: matchAssociationID,
				CIDRBlock:     testCidrBlock,
				CIDRBlockState: v1beta1.VPCCIDRBlockState{
					State:         testStateString,
					StatusMessage: testStatus,
				},
			},
		},
		"IPv6": {
			associationID: matchAssociationID,
			in: types.Vpc{
				Ipv6CidrBlockAssociationSet: []types.VpcIpv6CidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						Ipv6CidrBlock: &testCidrBlock,
						Ipv6CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						Ipv6CidrBlock: &otherCidrBlock,
						Ipv6CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
				},
			},
			out: v1beta1.VPCCIDRBlockObservation{
				AssociationID: matchAssociationID,
				IPv6CIDRBlock: testCidrBlock,
				IPv6CIDRBlockState: v1beta1.VPCCIDRBlockState{
					State:         testStateString,
					StatusMessage: testStatus,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateVpcCIDRBlockObservation(tc.associationID, tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateVPCCIDRBlockObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFindVPCCIDRBlockStatus(t *testing.T) {
	cases := map[string]struct {
		associationID string
		in            types.Vpc
		out           types.VpcCidrBlockStateCode
	}{
		"IPv4": {
			associationID: matchAssociationID,
			in: types.Vpc{
				CidrBlockAssociationSet: []types.VpcCidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						CidrBlock:     &testCidrBlock,
						CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						CidrBlock:     &otherCidrBlock,
						CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeDisassociated,
							StatusMessage: &testStatus,
						},
					},
				},
			},
			out: testState,
		},
		"IPv6": {
			associationID: matchAssociationID,
			in: types.Vpc{
				Ipv6CidrBlockAssociationSet: []types.VpcIpv6CidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						Ipv6CidrBlock: &testCidrBlock,
						Ipv6CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						Ipv6CidrBlock: &otherCidrBlock,
						Ipv6CidrBlockState: &types.VpcCidrBlockState{
							State:         types.VpcCidrBlockStateCodeDisassociated,
							StatusMessage: &testStatus,
						},
					},
				},
			},
			out: testState,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := FindVPCCIDRBlockStatus(tc.associationID, tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("FindVPCCIDRBlockStatus(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsVpcCidrDeleting(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.VPCCIDRBlockObservation
		out bool
	}{
		"IPv4": {
			in: v1beta1.VPCCIDRBlockObservation{
				CIDRBlock: testCidrBlock,
				CIDRBlockState: v1beta1.VPCCIDRBlockState{
					State:         testStateString,
					StatusMessage: testStatus,
				},
			},
			out: false,
		},
		"IPv6": {
			in: v1beta1.VPCCIDRBlockObservation{
				AssociationID: matchAssociationID,
				IPv6CIDRBlock: testCidrBlock,
				IPv6CIDRBlockState: v1beta1.VPCCIDRBlockState{
					State:         testStateString,
					StatusMessage: testStatus,
				},
			},
			out: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := IsVpcCidrDeleting(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("IsVpcCidrDeleting(...): -want, +got:\n%s", diff)
			}
		})
	}
}
