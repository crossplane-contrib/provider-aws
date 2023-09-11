package capacityreservation

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupCapacityReservation adds a controller that reconciles CapacityReservation.
func SetupCapacityReservation(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CapacityReservationGroupKind)

	opts := createOptions()

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.CapacityReservation{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.CapacityReservationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func createOptions() []option {
	return []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.filterList = filterList
			e.isUpToDate = isUpToDate
		},
	}
}

func filterList(cr *svcapitypes.CapacityReservation, list *svcsdk.DescribeCapacityReservationsOutput) *svcsdk.DescribeCapacityReservationsOutput {
	if len(list.CapacityReservations) == 0 {
		return list
	}
	capacityReservations := []*svcsdk.CapacityReservation{}
	for _, c := range list.CapacityReservations {
		if aws.StringValue(c.CapacityReservationId) == meta.GetExternalName(cr) {
			capacityReservations = append(capacityReservations, c)
		}
	}
	list.CapacityReservations = capacityReservations
	return list
}

func isUpToDate(cr *svcapitypes.CapacityReservation, c *svcsdk.CapacityReservation) (bool, error) {
	if c.AvailabilityZone != cr.Spec.ForProvider.AvailabilityZone ||
		c.InstanceMatchCriteria != cr.Spec.ForProvider.InstanceMatchCriteria ||
		c.InstancePlatform != cr.Spec.ForProvider.InstancePlatform ||
		c.InstanceType != cr.Spec.ForProvider.InstanceType ||
		c.Tenancy != cr.Spec.ForProvider.Tenancy ||
		c.TotalInstanceCount != cr.Status.AtProvider.TotalInstanceCount {
		return false, nil
	} else {
		return true, nil
	}
}

func postCreate(_ context.Context, cr *svcapitypes.CapacityReservation, resp *svcsdk.CreateCapacityReservationOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.CapacityReservation.CapacityReservationId))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.CapacityReservation, resp *svcsdk.DescribeCapacityReservationsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch *resp.CapacityReservations[0].State {
	case svcsdk.CapacityReservationStateActive:
		cr.SetConditions(xpv1.Available())
	case svcsdk.CapacityReservationStateExpired:
		cr.SetConditions(xpv1.Unavailable())
	case svcsdk.CapacityReservationStateCancelled:
		cr.SetConditions(xpv1.Unavailable())
	case svcsdk.CapacityReservationStatePending:
		cr.SetConditions(xpv1.Creating())
	case svcsdk.CapacityReservationStateFailed:
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}
