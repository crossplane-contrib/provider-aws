package testing

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
)

// CapacityReservationModifier is a function which modifies the CapacityReservation for testing
type CapacityReservationModifier func(capacityReservation *svcapitypes.CapacityReservation)

// WithConditions sets the Conditions for an CapacityReservation
func WithConditions(c ...xpv1.Condition) CapacityReservationModifier {
	return func(r *svcapitypes.CapacityReservation) { r.Status.ConditionedStatus.Conditions = c }
}

// WithSpec sets the Spec for an CapacityReservation
func WithSpec(c svcapitypes.CapacityReservationParameters) CapacityReservationModifier {
	return func(r *svcapitypes.CapacityReservation) { r.Spec.ForProvider = c }
}

// WithStatus sets the Status for an CapacityReservation
func WithStatus(s svcapitypes.CapacityReservationObservation) CapacityReservationModifier {
	return func(r *svcapitypes.CapacityReservation) { r.Status.AtProvider = s }
}

// WithExternalName sets the ExternalName for an CapacityReservation
func WithExternalName(n string) CapacityReservationModifier {
	return func(r *svcapitypes.CapacityReservation) { meta.SetExternalName(r, n) }
}

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
	meta.SetExternalName(cr, "test.capacityReservation.name")
	for _, f := range m {
		f(cr)
	}
	return cr
}
