package broker

import (
	"context"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/mq/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupBroker adds a controller that reconciles Stage.
func SetupBroker(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.BrokerKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postObserve = postObserve
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Broker{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.BrokerGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerInput) error {
	obj.BrokerId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerResponse, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch awsclients.StringValue(obj.BrokerState) {
	case string(svcapitypes.BrokerState_RUNNING):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.BrokerState_CREATION_IN_PROGRESS):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.BrokerState_REBOOT_IN_PROGRESS):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.BrokerState_DELETION_IN_PROGRESS):
		cr.SetConditions(xpv1.Deleting())
	}
	return obs, err
}

func preUpdate(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.UpdateBrokerRequest) error {
	obj.BrokerId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DeleteBrokerInput) (bool, error) {
	obj.BrokerId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.CreateBrokerResponse, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.BrokerId))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}
