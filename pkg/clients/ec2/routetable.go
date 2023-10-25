package ec2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == RouteTableIDNotFound
}

// IsRouteNotFoundErr returns true if the error is because the route doesn't exist
func IsRouteNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == RouteNotFound
}

// IsAssociationIDNotFoundErr returns true if the error is because the association doesn't exist
func IsAssociationIDNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == AssociationIDNotFound
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
func LateInitializeRT(in *v1beta1.RouteTableParameters, rt *ec2types.RouteTable) { //nolint:gocyclo
	if rt == nil {
		return
	}
	in.VPCID = pointer.LateInitialize(in.VPCID, rt.VpcId)

	if !pointer.BoolValue(in.IgnoreRoutes) {
		if len(in.Routes) == 0 && len(rt.Routes) != 0 {
			in.Routes = make([]v1beta1.RouteBeta, len(rt.Routes))
			for i, val := range rt.Routes {
				in.Routes[i] = v1beta1.RouteBeta{
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
		in.Tags = BuildFromEC2TagsV1Beta1(rt.Tags)
	}
}

// CreateRTPatch creates a *v1beta1.RouteTableParameters that has only the changed
// values between the target *v1beta1.RouteTableParameters and the current
// *ec2.RouteTable
func CreateRTPatch(in ec2types.RouteTable, target v1beta1.RouteTableParameters) (*v1beta1.RouteTableParameters, error) {
	targetCopy := target.DeepCopy()
	currentParams := &v1beta1.RouteTableParameters{}

	SortTagsV1Beta1(target.Tags, in.Tags)

	if !pointer.BoolValue(target.IgnoreRoutes) {
		// Add the default route for fair comparison.
		for _, val := range in.Routes {
			if val.GatewayId != nil && *val.GatewayId == DefaultLocalGatewayID {
				targetCopy.Routes = append([]v1beta1.RouteBeta{{
					GatewayID:            val.GatewayId,
					DestinationCIDRBlock: val.DestinationCidrBlock,
				}}, target.Routes...)
			}
		}
		SortRoutes(targetCopy.Routes, in.Routes)
	}

	LateInitializeRT(currentParams, &in)

	if !pointer.BoolValue(target.IgnoreRoutes) {
		for i := range targetCopy.Routes {
			targetCopy.Routes[i].ClearRefSelectors()
		}
	}

	for i := range target.Associations {
		targetCopy.Associations[i].ClearRefSelectors()
	}

	jsonPatch, err := jsonpatch.CreateJSONPatch(*currentParams, targetCopy)
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
		cmpopts.IgnoreFields(v1beta1.RouteTableParameters{}, "IgnoreRoutes"),
		cmpopts.IgnoreFields(v1beta1.RouteTableParameters{}, "Region"),
	), nil
}

// SortRoutes sorts array of Routes on DestinationCIDR
func SortRoutes(route []v1beta1.RouteBeta, ec2Route []ec2types.Route) {
	sort.Slice(route, func(i, j int) bool {
		return compareRoutes(i, j, route)
	})

	sort.Slice(ec2Route, func(i, j int) bool {
		return compareEC2Routes(i, j, ec2Route)
	})
}

func compareRoutes(i, j int, route []v1beta1.RouteBeta) bool {
	if route[i].DestinationCIDRBlock != nil && route[j].DestinationCIDRBlock != nil {
		return *route[i].DestinationCIDRBlock < *route[j].DestinationCIDRBlock
	}
	if route[i].DestinationIPV6CIDRBlock != nil && route[j].DestinationIPV6CIDRBlock != nil {
		return *route[i].DestinationIPV6CIDRBlock < *route[j].DestinationIPV6CIDRBlock
	}
	if route[i].DestinationCIDRBlock != nil && route[j].DestinationIPV6CIDRBlock != nil {
		return true
	}
	if route[i].DestinationIPV6CIDRBlock != nil && route[j].DestinationCIDRBlock != nil {
		return false
	}
	return false
}

func compareEC2Routes(i, j int, ec2Route []ec2types.Route) bool {
	if ec2Route[i].DestinationCidrBlock != nil && ec2Route[j].DestinationCidrBlock != nil {
		return *ec2Route[i].DestinationCidrBlock < *ec2Route[j].DestinationCidrBlock
	}
	if ec2Route[i].DestinationIpv6CidrBlock != nil && ec2Route[j].DestinationIpv6CidrBlock != nil {
		return *ec2Route[i].DestinationIpv6CidrBlock < *ec2Route[j].DestinationIpv6CidrBlock
	}
	if ec2Route[i].DestinationCidrBlock != nil && ec2Route[j].DestinationIpv6CidrBlock != nil {
		return true
	}
	if ec2Route[i].DestinationIpv6CidrBlock != nil && ec2Route[j].DestinationCidrBlock != nil {
		return false
	}
	return false
}

// ValidateRoutes on empty cidrs
func ValidateRoutes(route []v1beta1.RouteBeta) error {
	errs := make([]string, 0, len(route))
	for i, r := range route {
		if r.DestinationCIDRBlock == nil && r.DestinationIPV6CIDRBlock == nil {
			errs = append(errs, fmt.Sprintf("route[%d]: both v4 and v6 cidrs are empty", i))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid routes: %s", strings.Join(errs, ";"))
	}
	return nil
}
