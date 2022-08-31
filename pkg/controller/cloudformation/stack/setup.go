package stack

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudformation"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudformation/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupStack adds a controller that reconciles Stack.
func SetupStack(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.StackGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Stack{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StackGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func postObserve(_ context.Context, cr *svcapitypes.Stack, resp *svcsdk.DescribeStacksOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.Stacks[0].StackStatus) {
	case string(svcapitypes.StackStatus_SDK_CREATE_COMPLETE),
		string(svcapitypes.StackStatus_SDK_UPDATE_COMPLETE),
		string(svcapitypes.StackStatus_SDK_IMPORT_COMPLETE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.StackStatus_SDK_CREATE_IN_PROGRESS),
		string(svcapitypes.StackStatus_SDK_UPDATE_IN_PROGRESS),
		string(svcapitypes.StackStatus_SDK_IMPORT_IN_PROGRESS):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.StackStatus_SDK_DELETE_IN_PROGRESS),
		string(svcapitypes.StackStatus_SDK_ROLLBACK_IN_PROGRESS),
		string(svcapitypes.StackStatus_SDK_IMPORT_ROLLBACK_IN_PROGRESS),
		string(svcapitypes.StackStatus_SDK_UPDATE_ROLLBACK_IN_PROGRESS):
		cr.SetConditions(xpv1.Deleting())
	default:
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}
