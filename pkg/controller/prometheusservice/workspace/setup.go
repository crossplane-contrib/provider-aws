package workspace

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/prometheusservice/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errNotWorkspace     = "managed resource is not an Workspace custom resource"
	errKubeUpdateFailed = "cannot update Workspace custom resource"
)

// SetupWorkspace adds a controller that reconciles Workspace for PrometheusService.
func SetupWorkspace(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.WorkspaceGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.postDelete = postDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Workspace{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.WorkspaceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.Workspace, obj *svcsdk.DescribeWorkspaceInput) error {
	obj.WorkspaceId = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Workspace, resp *svcsdk.DescribeWorkspaceOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.Workspace.Status.StatusCode) {
	case string(svcapitypes.WorkspaceStatusCode_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.WorkspaceStatusCode_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.WorkspaceStatusCode_CREATION_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	}

	cr.Status.AtProvider.ARN = resp.Workspace.Arn
	cr.Status.AtProvider.PrometheusEndpoint = resp.Workspace.PrometheusEndpoint
	cr.Status.AtProvider.Status.StatusCode = resp.Workspace.Status.StatusCode

	obs.ConnectionDetails = managed.ConnectionDetails{
		"arn":                []byte(awsclients.StringValue(resp.Workspace.Arn)),
		"prometheusEndpoint": []byte(awsclients.StringValue(resp.Workspace.PrometheusEndpoint)),
		"workspaceId":        []byte(awsclients.StringValue(resp.Workspace.WorkspaceId)),
	}

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Workspace, resp *svcsdk.CreateWorkspaceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.WorkspaceId))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Workspace, obj *svcsdk.DeleteWorkspaceInput) (bool, error) {
	obj.WorkspaceId = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postDelete(_ context.Context, cr *svcapitypes.Workspace, obj *svcsdk.DeleteWorkspaceOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeConflictException) {
			// skip: Can't delete workspace in non-ACTIVE state. Current status is DELETING
			return nil
		}
		return err
	}
	return err
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Workspace)
	if !ok {
		return errors.New(errNotWorkspace)
	}
	if cr.Spec.ForProvider.Tags == nil {
		cr.Spec.ForProvider.Tags = map[string]*string{}
	}
	for k, v := range resource.GetExternalTags(mg) {
		cr.Spec.ForProvider.Tags[k] = awsclients.String(v)
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
