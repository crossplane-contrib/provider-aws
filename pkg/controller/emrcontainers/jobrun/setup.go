package jobrun

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/emrcontainers"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/emrcontainers/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	firstObserveJobRunID = "0000000000000000000"
)

// SetupJobRun adds a controller that reconciles JobRun.
func SetupJobRun(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobRunKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.preObserve = preObserve
			e.postCreate = postCreate
			e.postDelete = e.postDeleter
			e.postObserve = postObserve
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
		resource.ManagedKind(svcapitypes.JobRunGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.JobRun{}).
		Complete(r)
}

func preObserve(ctx context.Context, cr *svcapitypes.JobRun, input *svcsdk.DescribeJobRunInput) error {
	externalName := meta.GetExternalName(cr)
	if externalName == cr.Name {
		// ensure 404 is returned on first pass.
		input.Id = aws.String(firstObserveJobRunID)
	} else {
		input.Id = aws.String(externalName)
	}
	input.VirtualClusterId = cr.Spec.ForProvider.VirtualClusterID
	return nil
}
func postObserve(ctx context.Context, cr *svcapitypes.JobRun, resp *svcsdk.DescribeJobRunOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}
	state := *resp.JobRun.State
	// job runs cannot be deleted explicitly. they will be GCed after some time within AWS
	if meta.WasDeleted(cr) && (state == svcsdk.JobRunStateCancelled || state == svcsdk.JobRunStateCompleted || state == svcsdk.JobRunStateFailed) {
		obs.ResourceExists = false
	}
	setResourceCondition(state, cr)
	return obs, nil
}

func setResourceCondition(state string, cr *svcapitypes.JobRun) {
	switch state {
	case svcsdk.JobRunStateCancelled:
		cr.SetConditions(xpv1.Unavailable())
	case svcsdk.JobRunStateCancelPending:
		cr.SetConditions(xpv1.Deleting())
	case svcsdk.JobRunStateCompleted:
		cr.SetConditions(xpv1.Available())
	case svcsdk.JobRunStateFailed:
		cr.SetConditions(xpv1.Unavailable())
	case svcsdk.JobRunStatePending:
		cr.SetConditions(xpv1.Creating())
	case svcsdk.JobRunStateRunning:
		cr.SetConditions(xpv1.Available())
	case svcsdk.JobRunStateSubmitted:
		cr.SetConditions(xpv1.Creating())
	}
}

func preCreate(_ context.Context, cr *svcapitypes.JobRun, input *svcsdk.StartJobRunInput) error {
	input.VirtualClusterId = cr.Spec.ForProvider.VirtualClusterID
	input.Name = &cr.Name
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.JobRun, resp *svcsdk.StartJobRunOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return cre, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.Id))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.JobRun, input *svcsdk.CancelJobRunInput) (bool, error) {
	input.VirtualClusterId = cr.Spec.ForProvider.VirtualClusterID
	return input.VirtualClusterId == nil, nil
}

func (e *external) postDeleter(ctx context.Context, cr *svcapitypes.JobRun, resp *svcsdk.CancelJobRunOutput, err error) error {
	if err == nil {
		return nil
	}

	// error context is stripped. cannot type assert.
	cause := errors.Cause(err)
	if strings.HasPrefix(cause.Error(), svcsdk.ErrCodeValidationException) {
		res, jobErr := e.client.DescribeJobRunWithContext(ctx, &svcsdk.DescribeJobRunInput{
			Id:               resp.Id,
			VirtualClusterId: resp.VirtualClusterId,
		})
		if jobErr != nil {
			return errors.Wrap(jobErr, cause.Error())
		}
		state := *res.JobRun.State
		if state == svcsdk.JobRunStateCompleted || state == svcsdk.JobRunStateCancelled || state == svcsdk.JobRunStateFailed {
			return nil
		}
	}
	return err
}
