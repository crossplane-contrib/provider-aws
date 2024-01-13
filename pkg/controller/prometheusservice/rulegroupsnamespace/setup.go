package rulegroupsnamespace

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

// SetupRuleGroupsNamespace adds a controller that reconciles RuleGroupsNamespace.
func SetupRuleGroupsNamespace(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RuleGroupsNamespaceGroupKind)
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
		resource.ManagedKind(svcapitypes.RuleGroupsNamespaceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.RuleGroupsNamespace{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, obj *svcsdk.DescribeRuleGroupsNamespaceInput) error {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	obj.Name = cr.Spec.ForProvider.Name
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, obj *svcsdk.CreateRuleGroupsNamespaceInput) error {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	obj.Name = cr.Spec.ForProvider.Name
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, obj *svcsdk.DeleteRuleGroupsNamespaceInput) (bool, error) {
	obj.WorkspaceId = cr.Spec.ForProvider.WorkspaceID
	obj.Name = cr.Spec.ForProvider.Name
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, resp *svcsdk.CreateRuleGroupsNamespaceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.Arn))
	return cre, nil
}

func postDelete(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, obj *svcsdk.DeleteRuleGroupsNamespaceOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeConflictException) {
			// skip: Can't delete rulegroupsnamespace in non-ACTIVE state. Current status is DELETING
			return nil
		}
		return err
	}
	return err
}

func postObserve(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, resp *svcsdk.DescribeRuleGroupsNamespaceOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.RuleGroupsNamespace.Status.StatusCode) {
	case string(svcapitypes.RuleGroupsNamespaceStatusCode_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.RuleGroupsNamespaceStatusCode_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.RuleGroupsNamespaceStatusCode_CREATION_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.RuleGroupsNamespaceStatusCode_UPDATE_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	}

	cr.Status.AtProvider.ARN = resp.RuleGroupsNamespace.Arn
	cr.Status.AtProvider.Status.StatusCode = resp.RuleGroupsNamespace.Status.StatusCode

	return obs, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.RuleGroupsNamespace, resp *svcsdk.DescribeRuleGroupsNamespaceOutput) (bool, string, error) {
	// A rule that's currently creating, deleting, or updating can't be
	// updated, so we temporarily consider it to be up-to-date no matter
	// what.
	switch aws.StringValue(cr.Status.AtProvider.Status.StatusCode) {
	case string(svcapitypes.RuleGroupsNamespaceStatusCode_CREATING), string(svcapitypes.RuleGroupsNamespaceStatusCode_UPDATING), string(svcapitypes.RuleGroupsNamespaceStatusCode_DELETING):
		return true, "", nil
	}

	if cmp := bytes.Compare(cr.Spec.ForProvider.Data, resp.RuleGroupsNamespace.Data); cmp != 0 {
		return false, "", nil
	}
	return true, "", nil
}

type updateClient struct {
	client svcsdkapi.PrometheusServiceAPI
}

// GeneratePutRuleGroupsNamespaceInput returns a update input.
func GeneratePutRuleGroupsNamespaceInput(cr *svcapitypes.RuleGroupsNamespace) *svcsdk.PutRuleGroupsNamespaceInput {
	res := &svcsdk.PutRuleGroupsNamespaceInput{}

	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.WorkspaceID != nil {
		res.SetWorkspaceId(*cr.Spec.ForProvider.WorkspaceID)
	}
	if cr.Spec.ForProvider.Data != nil {
		res.SetData(cr.Spec.ForProvider.Data)
	}

	return res
}

func (e *updateClient) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.RuleGroupsNamespace)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GeneratePutRuleGroupsNamespaceInput(cr)
	_, err := e.client.PutRuleGroupsNamespaceWithContext(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "update failed")
	}
	return managed.ExternalUpdate{}, nil
}
