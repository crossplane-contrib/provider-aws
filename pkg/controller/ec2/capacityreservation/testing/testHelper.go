package testing

import (
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
)

// CapacityReservationModifier is a function which modifies the CapacityReservation for testing
type CapacityReservationModifier func(bucket *svcapitypes.CapacityReservation)

// CapacityReservation creates a CapacityReservation for use in testing
func CapacityReservation(m ...CapacityReservationModifier) *svcapitypes.CapacityReservation {
	cr := &svcapitypes.CapacityReservation{
		Spec: svcapitypes.CapacityReservationSpec{
			ForProvider: svcapitypes.CapacityReservationParameters{
				Region: "us-east-1",
			},
		},
		Status: svcapitypes.CapacityReservationStatus{},
	}
	for _, f := range m {
		f(cr)
	}
	meta.SetExternalName(cr, "test.capacityReservation.name")
	return cr
}
