package lifecyclepolicy

import (
	"context"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/ecr"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ecr/ecriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupLifecyclePolicy adds a controller that reconciles LifecyclePolicy.
func SetupLifecyclePolicy(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.LifecyclePolicyGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.isUpToDate = isUpToDate
			e.preDelete = preDelete
			u := &updateClient{client: e.client}
			e.update = u.update
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.LifecyclePolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LifecyclePolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.LifecyclePolicy, obj *svcsdk.GetLifecyclePolicyInput) error {
	obj.RepositoryName = cr.Spec.ForProvider.RepositoryName
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.LifecyclePolicy, obj *svcsdk.GetLifecyclePolicyOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.LifecyclePolicy, obj *svcsdk.PutLifecyclePolicyInput) error {
	obj.RepositoryName = cr.Spec.ForProvider.RepositoryName
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.LifecyclePolicy, obj *svcsdk.DeleteLifecyclePolicyInput) (bool, error) {
	obj.RepositoryName = cr.Spec.ForProvider.RepositoryName
	return false, nil
}

func isUpToDate(cr *svcapitypes.LifecyclePolicy, obj *svcsdk.GetLifecyclePolicyOutput) (bool, error) {
	if diff := cmp.Diff(cr.Spec.ForProvider.LifecyclePolicyText, obj.LifecyclePolicyText); diff != "" {
		return false, nil
	}
	return true, nil
}

type updateClient struct {
	client svcsdkapi.ECRAPI
}

func (e *updateClient) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.LifecyclePolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GeneratePutLifecyclePolicyInput(cr)
	input.RepositoryName = cr.Spec.ForProvider.RepositoryName

	_, err := e.client.PutLifecyclePolicyWithContext(ctx, input)
	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}
