package targetgroup

import (
	"context"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"

	svcapitypes "github.com/crossplane/provider-aws/apis/elbv2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupTargetGroup adds a controller that reconciles TargetGroup.
func SetupTargetGroup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.TargetGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.TargetGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TargetGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func postObserve(_ context.Context, cr *svcapitypes.TargetGroup, _ *svcsdk.DescribeTargetGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.TargetGroup, resp *svcsdk.CreateTargetGroupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.TargetGroups[0].TargetGroupArn))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.TargetGroup, obj *svcsdk.DeleteTargetGroupInput) (bool, error) {
	obj.TargetGroupArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
