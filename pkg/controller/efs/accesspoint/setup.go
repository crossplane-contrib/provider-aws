package accesspoint

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAccessPoint adds a controller that reconciles AccessPoint.
func SetupAccessPoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AccessPointGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
		},
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.AccessPointGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.AccessPoint{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.DescribeAccessPointsInput) error {
	obj.FileSystemId = nil
	obj.AccessPointId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.AccessPoint, resp *svcsdk.DescribeAccessPointsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch pointer.StringValue(resp.AccessPoints[0].LifeCycleState) {
	case string(svcapitypes.LifeCycleState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.LifeCycleState_creating):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.LifeCycleState_deleting):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.LifeCycleState_error):
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.CreateAccessPointInput) error {
	obj.FileSystemId = cr.Spec.ForProvider.FileSystemID
	obj.ClientToken = pointer.ToOrNilIfZeroValue(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.CreateAccessPointOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.AccessPointId))
	return managed.ExternalCreation{}, nil
}
