package cloudfrontfunction

import (
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
)

func SetupFunction(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Function{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.FunctionGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube: mgr.GetClient()
			}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}