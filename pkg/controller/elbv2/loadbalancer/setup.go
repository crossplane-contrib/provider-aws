package loadbalancer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupLoadBalancer adds a controller that reconciles LoadBalancer.
func SetupLoadBalancer(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.LoadBalancerGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.LoadBalancerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.LoadBalancer{}).
		Complete(r)
}

func postObserve(_ context.Context, cr *svcapitypes.LoadBalancer, resp *svcsdk.DescribeLoadBalancersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.LoadBalancers[0].State.Code) {
	case string(svcapitypes.LoadBalancerStateEnum_active):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.LoadBalancerStateEnum_provisioning):
		cr.SetConditions(xpv1.Creating())
	}
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.LoadBalancer, resp *svcsdk.CreateLoadBalancerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.LoadBalancers[0].LoadBalancerArn))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.LoadBalancer, obj *svcsdk.DeleteLoadBalancerInput) (bool, error) {
	obj.LoadBalancerArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func preCreate(_ context.Context, cr *svcapitypes.LoadBalancer, obj *svcsdk.CreateLoadBalancerInput) error {
	obj.Type = cr.Spec.ForProvider.Type
	return nil
}
