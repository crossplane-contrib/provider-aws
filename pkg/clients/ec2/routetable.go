package ec2

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// DefaultLocalGatewayID is the id for local gateway
	DefaultLocalGatewayID = "local"

	// RouteTableIDNotFound is the code that is returned by ec2 when the given SubnetID is invalid
	RouteTableIDNotFound = "InvalidRouteTableID.NotFound"

	// RouteNotFound is the code that is returned when the given route is not found
	RouteNotFound = "InvalidRoute.NotFound"

	// AssociationIDNotFound is the code that is returned when then given AssociationID is invalid
	AssociationIDNotFound = "InvalidAssociationID.NotFound"
)

// RouteTableClient is the external client used for RouteTable Custom Resource
type RouteTableClient interface {
	CreateRouteTable(ctx context.Context, input *ec2.CreateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.CreateRouteTableOutput, error)
	DeleteRouteTable(ctx context.Context, input *ec2.DeleteRouteTableInput, opts ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error)
	DescribeRouteTables(ctx context.Context, input *ec2.DescribeRouteTablesInput, opts ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	CreateRoute(ctx context.Context, input *ec2.CreateRouteInput, opts ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error)
	DeleteRoute(ctx context.Context, input *ec2.DeleteRouteInput, opts ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error)
	AssociateRouteTable(ctx context.Context, input *ec2.AssociateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.AssociateRouteTableOutput, error)
	DisassociateRouteTable(ctx context.Context, input *ec2.DisassociateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.DisassociateRouteTableOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// NewRouteTableClient returns a new client using AWS credentials as JSON encoded data.
func NewRouteTableClient(cfg aws.Config) RouteTableClient {
	return ec2.NewFromConfig(cfg)
}

// IsRouteTableNotFoundErr returns true if the error is because the route table doesn't exist
func IsRouteTableNotFoundErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == RouteTableIDNotFound {
			return true
		}
	}
	return false
}

// IsRouteNotFoundErr returns true if the error is because the route doesn't exist
func IsRouteNotFoundErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == RouteNotFound {
			return true
		}
	}
	return false
}

// IsAssociationIDNotFoundErr returns true if the error is because the association doesn't exist
func IsAssociationIDNotFoundErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == AssociationIDNotFound {
			return true
		}
	}
	return false
}

// GenerateRTObservation is used to produce v1beta1.RouteTableExternalStatus from
// ec2.RouteTable.
func GenerateRTObservation(rt ec2types.RouteTable) v1beta1.RouteTableObservation {
	o := v1beta1.RouteTableObservation{
		OwnerID:      aws.ToString(rt.OwnerId),
		RouteTableID: aws.ToString(rt.RouteTableId),
	}

	if len(rt.Routes) > 0 {
		o.Routes = make([]v1beta1.RouteState, len(rt.Routes))
		for i, rt := range rt.Routes {
			o.Routes[i] = v1beta1.RouteState{
				State:                    string(rt.State),
				DestinationCIDRBlock:     aws.ToString(rt.DestinationCidrBlock),
				DestinationIPV6CIDRBlock: aws.ToString(rt.DestinationIpv6CidrBlock),
				GatewayID:                aws.ToString(rt.GatewayId),
				InstanceID:               aws.ToString(rt.InstanceId),
				LocalGatewayID:           aws.ToString(rt.LocalGatewayId),
				NatGatewayID:             aws.ToString(rt.NatGatewayId),
				NetworkInterfaceID:       aws.ToString(rt.NetworkInterfaceId),
				TransitGatewayID:         aws.ToString(rt.TransitGatewayId),
				VpcPeeringConnectionID:   aws.ToString(rt.VpcPeeringConnectionId),
			}
		}
	}

	if len(rt.Associations) > 0 {
		o.Associations = make([]v1beta1.AssociationState, len(rt.Associations))
		for i, asc := range rt.Associations {
			o.Associations[i] = v1beta1.AssociationState{
				Main:          asc.Main,
				AssociationID: aws.ToString(asc.RouteTableAssociationId),
				State:         string(asc.AssociationState.State),
				SubnetID:      aws.ToString(asc.SubnetId),
			}
		}
	}

	return o
}

// LateInitializeRT fills the empty fields in *v1beta1.RouteTableParameters with
// the values seen in ec2.RouteTable.
func LateInitializeRT(in *v1beta1.RouteTableParameters, rt *ec2types.RouteTable) { // nolint:gocyclo
	if rt == nil {
		return
	}
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, rt.VpcId)

	if len(in.Routes) == 0 && len(rt.Routes) != 0 {
		in.Routes = make([]v1beta1.Route, len(rt.Routes))
		for i, val := range rt.Routes {
			in.Routes[i] = v1beta1.Route{
				DestinationCIDRBlock:   val.DestinationCidrBlock,
				GatewayID:              val.GatewayId,
				InstanceID:             val.InstanceId,
				LocalGatewayID:         val.LocalGatewayId,
				NatGatewayID:           val.NatGatewayId,
				NetworkInterfaceID:     val.NetworkInterfaceId,
				TransitGatewayID:       val.TransitGatewayId,
				VpcPeeringConnectionID: val.VpcPeeringConnectionId,
			}
		}
	}

	if len(in.Associations) == 0 && len(rt.Associations) != 0 {
		in.Associations = make([]v1beta1.Association, len(rt.Associations))
		for i, val := range rt.Associations {
			in.Associations[i] = v1beta1.Association{
				SubnetID: val.SubnetId,
			}
		}
	}

	if len(in.Tags) == 0 && len(rt.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(rt.Tags)
	}
}

// CreateRTPatch creates a *v1beta1.RouteTableParameters that has only the changed
// values between the target *v1beta1.RouteTableParameters and the current
// *ec2.RouteTable
func CreateRTPatch(in ec2types.RouteTable, target v1beta1.RouteTableParameters) (*v1beta1.RouteTableParameters, error) {
	targetCopy := target.DeepCopy()
	currentParams := &v1beta1.RouteTableParameters{}

	v1beta1.SortTags(target.Tags, in.Tags)

	// Add the default route for fair comparison.
	for _, val := range in.Routes {
		if val.GatewayId != nil && *val.GatewayId == DefaultLocalGatewayID {
			targetCopy.Routes = append([]v1beta1.Route{{
				GatewayID:            val.GatewayId,
				DestinationCIDRBlock: val.DestinationCidrBlock,
			}}, target.Routes...)
		}
	}
	SortRoutes(targetCopy.Routes, in.Routes)

	LateInitializeRT(currentParams, &in)

	for i := range targetCopy.Routes {
		targetCopy.Routes[i].ClearRefSelectors()
	}
	for i := range target.Associations {
		targetCopy.Associations[i].ClearRefSelectors()
	}

	jsonPatch, err := awsclients.CreateJSONPatch(*currentParams, targetCopy)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.RouteTableParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRtUpToDate checks whether there is a change in any of the modifiable fields.
func IsRtUpToDate(p v1beta1.RouteTableParameters, rt ec2types.RouteTable) (bool, error) {
	patch, err := CreateRTPatch(rt, p)
	if err != nil {
		return false, err
	}

	return cmp.Equal(&v1beta1.RouteTableParameters{}, patch,
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}),
		cmpopts.IgnoreFields(v1beta1.RouteTableParameters{}, "Region"),
	), nil
}

// SortRoutes sorts array of Routes on DestinationCIDR
func SortRoutes(route []v1beta1.Route, ec2Route []ec2types.Route) {
	sort.Slice(route, func(i, j int) bool {
		return (route[i].DestinationCIDRBlock != nil && *route[i].DestinationCIDRBlock < *route[j].DestinationCIDRBlock) ||
			(route[i].DestinationIPV6CIDRBlock != nil && *route[i].DestinationIPV6CIDRBlock < *route[j].DestinationIPV6CIDRBlock)
	})

	sort.Slice(ec2Route, func(i, j int) bool {
		return (ec2Route[i].DestinationCidrBlock != nil && *ec2Route[i].DestinationCidrBlock < *ec2Route[j].DestinationCidrBlock) ||
			(ec2Route[i].DestinationIpv6CidrBlock != nil && *ec2Route[i].DestinationIpv6CidrBlock < *ec2Route[j].DestinationIpv6CidrBlock)
	})
}
