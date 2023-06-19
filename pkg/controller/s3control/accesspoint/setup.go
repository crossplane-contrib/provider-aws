package accesspoint

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcsdk "github.com/aws/aws-sdk-go/service/s3control"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupAccessPoint adds a controller that reconciles Stage.
func SetupAccessPoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AccessPointGroupKind)

	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.AccessPoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AccessPointGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preDelete(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.DeleteAccessPointInput) (bool, error) {
	input.Name = aws.String(meta.GetExternalName(point))
	return point.Spec.DeletionPolicy == xpv1.DeletionOrphan, nil
}

func preCreate(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.CreateAccessPointInput) error {
	input.Name = aws.String(meta.GetExternalName(point))
	return nil
}

func preObserve(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.GetAccessPointInput) error {
	input.Name = aws.String(meta.GetExternalName(point))
	return nil
}

func postObserve(_ context.Context, point *svcapitypes.AccessPoint, _ *svcsdk.GetAccessPointOutput, observation managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	point.SetConditions(xpv1.Available())
	return observation, nil
}
