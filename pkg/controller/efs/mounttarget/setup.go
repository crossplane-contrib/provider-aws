package mounttarget

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupMountTarget adds a controller that reconciles MountTarget.
func SetupMountTarget(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.MountTargetGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		cpresource.ManagedKind(svcapitypes.MountTargetGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.MountTarget{}).
		Complete(r)
}

func preCreate(_ context.Context, cr *svcapitypes.MountTarget, obj *svcsdk.CreateMountTargetInput) error {
	obj.FileSystemId = cr.Spec.ForProvider.FileSystemID
	obj.SubnetId = cr.Spec.ForProvider.SubnetID

	for i := range cr.Spec.ForProvider.SecurityGroups {
		obj.SecurityGroups = append(obj.SecurityGroups, &cr.Spec.ForProvider.SecurityGroups[i])
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.MountTarget, obj *svcsdk.MountTargetDescription, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.MountTargetId))
	return managed.ExternalCreation{}, nil
}

func preObserve(_ context.Context, cr *svcapitypes.MountTarget, obj *svcsdk.DescribeMountTargetsInput) error {
	// Must specify exactly one mutually exclusive parameter.
	obj.AccessPointId = nil
	obj.Marker = nil
	obj.MaxItems = nil
	obj.FileSystemId = nil
	obj.MountTargetId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.MountTarget, obj *svcsdk.DescribeMountTargetsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if awsclients.StringValue(obj.MountTargets[0].LifeCycleState) == string(svcapitypes.LifeCycleState_available) {
		cr.SetConditions(xpv1.Available())
	}
	return obs, nil
}
