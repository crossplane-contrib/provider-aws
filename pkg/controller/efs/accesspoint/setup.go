package accesspoint

import (
	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	"golang.org/x/net/context"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/efs/v1alpha1"

	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupAccessPoint adds a controller that reconciles AccessPoint.
func SetupAccessPoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AccessPointGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.AccessPoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AccessPointGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.DescribeAccessPointsInput) error {
	obj.FileSystemId = nil
	obj.AccessPointId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.CreateAccessPointInput) error {
	obj.FileSystemId = cr.Spec.ForProvider.FileSystemID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.AccessPoint, obj *svcsdk.CreateAccessPointOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.AccessPointId))
	return managed.ExternalCreation{}, nil
}
