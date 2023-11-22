package alertmanagerdefinition

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/prometheusservice"
	svcsdkapi "github.com/aws/aws-sdk-go/service/prometheusservice/prometheusserviceiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/prometheusservice/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAlertManagerDefinition adds a controller that reconciles AlertManagerDefinition.
func SetupAlertManagerDefinition(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AlertManagerDefinitionGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.postDelete = postDelete
			e.postObserve = postObserve
			e.isUpToDate = isUpToDate
			u := &updateClient{client: e.client}
			e.update = u.update
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
		resource.ManagedKind(svcapitypes.AlertManagerDefinitionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.AlertManagerDefinition{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.AlertManagerDefinition, obj *svcsdk.DescribeAlertManagerDefinitionInput) error {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.AlertManagerDefinition, obj *svcsdk.CreateAlertManagerDefinitionInput) error {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.AlertManagerDefinition, obj *svcsdk.DeleteAlertManagerDefinitionInput) (bool, error) {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.AlertManagerDefinition, resp *svcsdk.CreateAlertManagerDefinitionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(cr.Spec.ForProvider.WorkspaceID))
	return cre, nil
}

func postDelete(_ context.Context, cr *svcapitypes.AlertManagerDefinition, obj *svcsdk.DeleteAlertManagerDefinitionOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeConflictException) {
			// skip: Can't delete alertmanagerdefinition in non-ACTIVE state. Current status is DELETING
			return nil
		}
		return err
	}
	return err
}

func postObserve(_ context.Context, cr *svcapitypes.AlertManagerDefinition, resp *svcsdk.DescribeAlertManagerDefinitionOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.AlertManagerDefinition.Status.StatusCode) {
	case string(svcapitypes.AlertManagerDefinitionStatusCode_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.AlertManagerDefinitionStatusCode_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.AlertManagerDefinitionStatusCode_CREATION_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.AlertManagerDefinitionStatusCode_UPDATE_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	}

	cr.Status.AtProvider.StatusCode = resp.AlertManagerDefinition.Status.StatusCode

	return obs, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.AlertManagerDefinition, resp *svcsdk.DescribeAlertManagerDefinitionOutput) (bool, string, error) {
	// An AlertManager Definition that's currently creating, deleting, or updating can't be
	// updated, so we temporarily consider it to be up-to-date no matter
	// what.
	switch aws.StringValue(cr.Status.AtProvider.StatusCode) {
	case string(svcapitypes.AlertManagerDefinitionStatusCode_CREATING), string(svcapitypes.AlertManagerDefinitionStatusCode_UPDATING), string(svcapitypes.AlertManagerDefinitionStatusCode_DELETING):
		return true, "", nil
	}

	if cmp := bytes.Compare(cr.Spec.ForProvider.Data, resp.AlertManagerDefinition.Data); cmp != 0 {
		return false, "", nil
	}
	return true, "", nil
}

type updateClient struct {
	client svcsdkapi.PrometheusServiceAPI
}

// GeneratePutAlertManagerDefinitionInput returns a update input.
func GeneratePutAlertManagerDefinitionInput(cr *svcapitypes.AlertManagerDefinition) *svcsdk.PutAlertManagerDefinitionInput {
	res := &svcsdk.PutAlertManagerDefinitionInput{}

	if cr.Spec.ForProvider.WorkspaceID != nil {
		res.SetWorkspaceId(*cr.Spec.ForProvider.WorkspaceID)
	}
	if cr.Spec.ForProvider.Data != nil {
		res.SetData(cr.Spec.ForProvider.Data)
	}

	return res
}

func (e *updateClient) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.AlertManagerDefinition)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GeneratePutAlertManagerDefinitionInput(cr)
	_, err := e.client.PutAlertManagerDefinitionWithContext(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "update failed")
	}
	return managed.ExternalUpdate{}, nil
}
