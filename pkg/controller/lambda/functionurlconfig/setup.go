package functionurlconfig

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/lambda/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupFunctionURL adds a controller that reconciles FunctionURLConfig.
func SetupFunctionURL(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FunctionURLConfigGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postObserve = postObserve
			e.isUpToDate = isUpToDate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.FunctionURLConfigGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.FunctionURLConfig{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.FunctionURLConfig, obj *svcsdk.GetFunctionUrlConfigInput) error {
	obj.FunctionName = cr.Spec.ForProvider.FunctionName

	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.FunctionURLConfig, obj *svcsdk.CreateFunctionUrlConfigInput) error {
	obj.FunctionName = cr.Spec.ForProvider.FunctionName

	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.FunctionURLConfig, obj *svcsdk.UpdateFunctionUrlConfigInput) error {
	obj.FunctionName = cr.Spec.ForProvider.FunctionName

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.FunctionURLConfig, obj *svcsdk.DeleteFunctionUrlConfigInput) (bool, error) {
	obj.FunctionName = cr.Spec.ForProvider.FunctionName

	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.FunctionURLConfig, _ *svcsdk.GetFunctionUrlConfigOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.FunctionURLConfig, obj *svcsdk.GetFunctionUrlConfigOutput) (bool, string, error) {
	if aws.StringValue(cr.Spec.ForProvider.AuthType) != aws.StringValue(obj.AuthType) {
		return false, "", nil
	}

	return isUpToDateCors(cr, obj), "", nil
}

func isUpToDateCors(cr *svcapitypes.FunctionURLConfig, obj *svcsdk.GetFunctionUrlConfigOutput) bool {
	sortCmp := cmpopts.SortSlices(func(x, y *string) bool {
		return *x < *y
	})

	switch {
	case cr.Spec.ForProvider.CORS == nil && obj.Cors == nil, cr.Spec.ForProvider.CORS == nil:
		return true

	case obj.Cors == nil,
		aws.BoolValue(cr.Spec.ForProvider.CORS.AllowCredentials) != aws.BoolValue(obj.Cors.AllowCredentials),
		!cmp.Equal(&cr.Spec.ForProvider.CORS.AllowHeaders, &obj.Cors.AllowHeaders, sortCmp),
		!cmp.Equal(&cr.Spec.ForProvider.CORS.AllowMethods, &obj.Cors.AllowMethods, sortCmp),
		!cmp.Equal(&cr.Spec.ForProvider.CORS.AllowOrigins, &obj.Cors.AllowOrigins, sortCmp),
		!cmp.Equal(&cr.Spec.ForProvider.CORS.ExposeHeaders, &obj.Cors.ExposeHeaders, sortCmp),
		aws.Int64Value(cr.Spec.ForProvider.CORS.MaxAge) != aws.Int64Value(obj.Cors.MaxAge):

		return false
	}

	return true
}
