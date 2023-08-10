package ec2

import (
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func IsCapacityReservationUpToDate(resource *svcapitypes.CapacityReservation, reservations *svcsdk.DescribeCapacityReservationsOutput) bool {

	reservation *svcsdk.CapacityReservation := GetCapacityReservation(resource.arn)

	/*if !cmp.Equal(p.InstanceType, res.InstanceType, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(p.AvailabilityZone, res.AvailabilityZone, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(p.InstancePlatform, res.InstancePlatform, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(p.InstanceCount, res.TotalInstanceCount, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(p.Tenancy, res.Tenancy, cmpopts.EquateEmpty()) {
		return false
	}*/
}

func GetCapacityReservation(arn *string, res *[]svcsdk.CapacityReservation) *svcsdk.CapacityReservation {
	for _, r := range res.CapacityReservations {
		if r.CapacityReservationArn == arn {
			return r
		}
	} 
	return nil
}