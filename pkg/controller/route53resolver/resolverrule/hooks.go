package resolverrule

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	svcapitypes "github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
)

// SetupResolverRule adds a controller that reconciles ResolverRule
func SetupResolverRule(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ResolverRuleGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.preUpdate = preUpdate
			e.postObserve = postObserve
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.ResolverRule{}).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind(v1alpha1.ResolverRuleGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *v1alpha1.ResolverRule, obj *svcsdk.GetResolverRuleInput) error {
	obj.ResolverRuleId = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *v1alpha1.ResolverRule, obj *svcsdk.CreateResolverRuleInput) error {
	obj.CreatorRequestId = aws.String(string(cr.GetObjectMeta().GetUID()))
	return nil
}

func postCreate(_ context.Context, cr *v1alpha1.ResolverRule, obj *svcsdk.CreateResolverRuleOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	meta.SetExternalName(cr, aws.StringValue(obj.ResolverRule.Id))
	return cre, err
}

func preDelete(_ context.Context, cr *v1alpha1.ResolverRule, obj *svcsdk.DeleteResolverRuleInput) (bool, error) {
	obj.ResolverRuleId = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func preUpdate(_ context.Context, cr *v1alpha1.ResolverRule, obj *svcsdk.UpdateResolverRuleInput) error {
	obj.ResolverRuleId = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResolverRule, obj *svcsdk.GetResolverRuleOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(obj.ResolverRule.Status) {
	case string(svcapitypes.ResolverRuleStatus_SDK_COMPLETE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.ResolverRuleStatus_SDK_UPDATING), string(svcapitypes.ResolverRuleStatus_SDK_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.ResolverRuleStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}

	return obs, err
}
