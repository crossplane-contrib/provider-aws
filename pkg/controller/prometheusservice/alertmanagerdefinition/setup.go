package alertmanagerdefinition

import (
	"context"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/prometheusservice"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/prometheusservice/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
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
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.AlertManagerDefinition{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AlertManagerDefinitionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
