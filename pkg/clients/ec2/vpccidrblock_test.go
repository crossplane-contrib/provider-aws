package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
)

var (
	matchAssociationID = "test"
	otherAssociationID = "other"
	testCidrBlock      = "10.0.0.0/0"
	otherCidrBlock     = "10.0.0.0/16"
	testStatus         = "status"
	testStateString    = string(ec2.VpcCidrBlockStateCodeAssociated)
	testState          = ec2.VpcCidrBlockStateCodeAssociated
)

func TestGenerateVPCCIDRBlockObservation(t *testing.T) {
	cases := map[string]struct {
		associationID string
		in            ec2.Vpc
		out           v1alpha1.VPCCIDRBlockObservation
	}{
		"IPv4": {
			associationID: matchAssociationID,
			in: ec2.Vpc{
				CidrBlockAssociationSet: []ec2.VpcCidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						CidrBlock:     &testCidrBlock,
						CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						CidrBlock:     &otherCidrBlock,
					},
				},
			},
			out: v1alpha1.VPCCIDRBlockObservation{
				AssociationID: &matchAssociationID,
				CIDRBlock:     &testCidrBlock,
				CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
					State:         &testStateString,
					StatusMessage: &testStatus,
				},
			},
		},
		"IPv6": {
			associationID: matchAssociationID,
			in: ec2.Vpc{
				Ipv6CidrBlockAssociationSet: []ec2.VpcIpv6CidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						Ipv6CidrBlock: &testCidrBlock,
						Ipv6CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						Ipv6CidrBlock: &otherCidrBlock,
						Ipv6CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
				},
			},
			out: v1alpha1.VPCCIDRBlockObservation{
				AssociationID: &matchAssociationID,
				IPv6CIDRBlock: &testCidrBlock,
				IPv6CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
					State:         &testStateString,
					StatusMessage: &testStatus,
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
		in            ec2.Vpc
		out           ec2.VpcCidrBlockStateCode
	}{
		"IPv4": {
			associationID: matchAssociationID,
			in: ec2.Vpc{
				CidrBlockAssociationSet: []ec2.VpcCidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						CidrBlock:     &testCidrBlock,
						CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						CidrBlock:     &otherCidrBlock,
						CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeDisassociated,
							StatusMessage: &testStatus,
						},
					},
				},
			},
			out: testState,
		},
		"IPv6": {
			associationID: matchAssociationID,
			in: ec2.Vpc{
				Ipv6CidrBlockAssociationSet: []ec2.VpcIpv6CidrBlockAssociation{
					{
						AssociationId: &matchAssociationID,
						Ipv6CidrBlock: &testCidrBlock,
						Ipv6CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeAssociated,
							StatusMessage: &testStatus,
						},
					},
					{
						AssociationId: &otherAssociationID,
						Ipv6CidrBlock: &otherCidrBlock,
						Ipv6CidrBlockState: &ec2.VpcCidrBlockState{
							State:         ec2.VpcCidrBlockStateCodeDisassociated,
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
		in  v1alpha1.VPCCIDRBlockObservation
		out bool
	}{
		"IPv4": {
			in: v1alpha1.VPCCIDRBlockObservation{
				CIDRBlock: &testCidrBlock,
				CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
					State:         &testStateString,
					StatusMessage: &testStatus,
				},
			},
			out: false,
		},
		"IPv6": {
			in: v1alpha1.VPCCIDRBlockObservation{
				AssociationID: &matchAssociationID,
				IPv6CIDRBlock: &testCidrBlock,
				IPv6CIDRBlockState: &v1alpha1.VPCCIDRBlockState{
					State:         &testStateString,
					StatusMessage: &testStatus,
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
