package lifecyclepolicy

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ecr"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ecr/ecriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecr/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupLifecyclePolicy adds a controller that reconciles LifecyclePolicy.
func SetupLifecyclePolicy(mgr ctrl.Manager, o controller.Options) error {
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

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.LifecyclePolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.LifecyclePolicy{}).
		Complete(r)
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

func isUpToDate(_ context.Context, cr *svcapitypes.LifecyclePolicy, obj *svcsdk.GetLifecyclePolicyOutput) (bool, string, error) {
	diff := cmp.Diff(cr.Spec.ForProvider.LifecyclePolicyText, obj.LifecyclePolicyText)
	return diff == "", diff, nil
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
	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
}
