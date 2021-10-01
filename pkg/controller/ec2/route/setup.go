package route

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SetupRoute adds a controller that reconciles Route.
func SetupRoute(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.RouteGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Route{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.RouteGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func postCreate(ctx context.Context, cr *svcapitypes.Route, obj *svcsdk.CreateRouteOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if obj.Return == aws.Bool(true) {
		cr.SetConditions(xpv1.Available())
		cr.Status.AtProvider.Return = aws.Bool(true)
	}
	return cre, err
}
