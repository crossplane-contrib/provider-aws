package ec2

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha4"
	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// LocalGatewayID is the id for local gateway
	LocalGatewayID = "local"

	// RouteTableIDNotFound is the code that is returned by ec2 when the given SubnetID is invalid
	RouteTableIDNotFound = "InvalidRouteTableID.NotFound"

	// RouteNotFound is the code that is returned when the given route is not found
	RouteNotFound = "InvalidRoute.NotFound"

	// AssociationIDNotFound is the code that is returned when then given AssociationID is invalid
	AssociationIDNotFound = "InvalidAssociationID.NotFound"
)

// RouteTableClient is the external client used for RouteTable Custom Resource
type RouteTableClient interface {
	CreateRouteTableRequest(*ec2.CreateRouteTableInput) ec2.CreateRouteTableRequest
	DeleteRouteTableRequest(*ec2.DeleteRouteTableInput) ec2.DeleteRouteTableRequest
	DescribeRouteTablesRequest(*ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest
	CreateRouteRequest(*ec2.CreateRouteInput) ec2.CreateRouteRequest
	DeleteRouteRequest(*ec2.DeleteRouteInput) ec2.DeleteRouteRequest
	AssociateRouteTableRequest(*ec2.AssociateRouteTableInput) ec2.AssociateRouteTableRequest
	DisassociateRouteTableRequest(*ec2.DisassociateRouteTableInput) ec2.DisassociateRouteTableRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewRouteTableClient returns a new client using AWS credentials as JSON encoded data.
func NewRouteTableClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (RouteTableClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), nil
}

// IsRouteTableNotFoundErr returns true if the error is because the route table doesn't exist
func IsRouteTableNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == RouteTableIDNotFound {
			return true
		}
	}
	return false
}

// IsRouteNotFoundErr returns true if the error is because the route doesn't exist
func IsRouteNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == RouteNotFound {
			return true
		}
	}
	return false
}

// IsAssociationIDNotFoundErr returns true if the error is because the association doesn't exist
func IsAssociationIDNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == AssociationIDNotFound {
			return true
		}
	}
	return false
}

// GenerateRTObservation is used to produce v1alpha4.RouteTableExternalStatus from
// ec2.RouteTable.
func GenerateRTObservation(rt ec2.RouteTable) v1alpha4.RouteTableObservation {
	o := v1alpha4.RouteTableObservation{
		OwnerID:      aws.StringValue(rt.OwnerId),
		RouteTableID: aws.StringValue(rt.RouteTableId),
	}

	if len(rt.Routes) > 0 {
		o.Routes = make([]v1alpha4.RouteState, len(rt.Routes))
		for i, rt := range rt.Routes {
			o.Routes[i] = v1alpha4.RouteState{
				State:                string(rt.State),
				DestinationCIDRBlock: aws.StringValue(rt.DestinationCidrBlock),
				GatewayID:            aws.StringValue(rt.GatewayId),
			}
		}
	}

	if len(rt.Routes) > 0 {
		o.Associations = make([]v1alpha4.AssociationState, len(rt.Associations))
		for i, asc := range rt.Associations {
			o.Associations[i] = v1alpha4.AssociationState{
				Main:          aws.BoolValue(asc.Main),
				AssociationID: aws.StringValue(asc.RouteTableAssociationId),
				State:         asc.AssociationState.String(),
				SubnetID:      aws.StringValue(asc.SubnetId),
			}
		}
	}

	return o
}

// LateInitializeRT fills the empty fields in *v1alpha4.RouteTableParameters with
// the values seen in ec2.RouteTable.
func LateInitializeRT(in *v1alpha4.RouteTableParameters, rt *ec2.RouteTable) { // nolint:gocyclo
	if rt == nil {
		return
	}
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, rt.VpcId)

	if len(in.Routes) == 0 && len(rt.Routes) != 0 {
		in.Routes = make([]v1alpha4.Route, len(rt.Routes))
		for i, val := range rt.Routes {
			in.Routes[i] = v1alpha4.Route{
				DestinationCIDRBlock: val.DestinationCidrBlock,
				GatewayID:            val.GatewayId,
			}
		}
	}

	if len(in.Associations) == 0 && len(rt.Associations) != 0 {
		in.Associations = make([]v1alpha4.Association, len(rt.Associations))
		for i, val := range rt.Associations {
			in.Associations[i] = v1alpha4.Association{
				SubnetID: val.SubnetId,
			}
		}
	}

	if len(in.Tags) == 0 && len(rt.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(rt.Tags)
	}
}

// CreateRTPatch creates a *v1alpha4.RouteTableParameters that has only the changed
// values between the target *v1alpha4.RouteTableParameters and the current
// *ec2.RouteTable
func CreateRTPatch(in ec2.RouteTable, target v1alpha4.RouteTableParameters) (*v1alpha4.RouteTableParameters, error) {
	currentParams := &v1alpha4.RouteTableParameters{}

	v1beta1.SortTags(target.Tags, in.Tags)

	// Add the default route for fair comparison.
	for _, val := range in.Routes {
		if *val.GatewayId == LocalGatewayID {
			target.Routes = append([]v1alpha4.Route{{
				GatewayID:            val.GatewayId,
				DestinationCIDRBlock: val.DestinationCidrBlock,
			}}, target.Routes...)
		}
	}

	LateInitializeRT(currentParams, &in)

	jsonPatch, err := awsclients.CreateJSONPatch(*currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1alpha4.RouteTableParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRtUpToDate checks whether there is a change in any of the modifiable fields.
func IsRtUpToDate(p v1alpha4.RouteTableParameters, rt ec2.RouteTable) (bool, error) {
	patch, err := CreateRTPatch(rt, p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1alpha4.RouteTableParameters{}, patch, cmpopts.EquateEmpty(), cmpopts.IgnoreTypes(&v1alpha1.Reference{}, &v1alpha1.Selector{})), nil
}
