package launchtemplate

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupLaunchTemplate adds a controller that reconciles LaunchTemplate.
func SetupLaunchTemplate(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.LaunchTemplateGroupKind)
	opts := []option{setupExternal()}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.LaunchTemplateGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.LaunchTemplate{}).
		Complete(r)
}

func setupExternal() option {
	return func(e *external) {
		e.preObserve = preObserve
		e.preUpdate = preUpdate
		e.preDelete = preDelete
		e.postCreate = postCreate
		e.postObserve = postObserve
	}
}

func preObserve(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.DescribeLaunchTemplatesInput) error {
	obj.LaunchTemplateNames = append(obj.LaunchTemplateNames, aws.String(meta.GetExternalName(cr)))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.ModifyLaunchTemplateInput) error {
	obj.LaunchTemplateName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.DeleteLaunchTemplateInput) (bool, error) {
	obj.LaunchTemplateName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.LaunchTemplate, resp *svcsdk.CreateLaunchTemplateOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.LaunchTemplate.LaunchTemplateName))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.LaunchTemplate, resp *svcsdk.DescribeLaunchTemplatesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	lt := resp.LaunchTemplates[0]
	cr.Status.AtProvider.LaunchTemplate = &svcapitypes.LaunchTemplate_SDK{
		CreateTime:           pointer.TimeToMetaTime(lt.CreateTime),
		CreatedBy:            lt.CreatedBy,
		DefaultVersionNumber: lt.DefaultVersionNumber,
		LatestVersionNumber:  lt.LatestVersionNumber,
		LaunchTemplateID:     lt.LaunchTemplateId,
		LaunchTemplateName:   lt.LaunchTemplateName,
	}
	if lt.Tags != nil {
		cr.Status.AtProvider.LaunchTemplate.Tags = make([]*svcapitypes.Tag, len(lt.Tags))
		for i, t := range lt.Tags {
			cr.Status.AtProvider.LaunchTemplate.Tags[i] = &svcapitypes.Tag{
				Key:   t.Key,
				Value: t.Value,
			}
		}
	}

	cr.SetConditions(xpv1.Available())
	return obs, nil
}
