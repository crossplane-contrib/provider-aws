package launchtemplateversion

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupLaunchTemplateVersion adds a controller that reconciles LaunchTemplateVersion.
func SetupLaunchTemplateVersion(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.LaunchTemplateVersionGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postCreate = postCreate
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.delete = e.deleter
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
		cpresource.ManagedKind(svcapitypes.LaunchTemplateVersionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.LaunchTemplateVersion{}).
		Complete(r)
}

func preCreate(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, obj *svcsdk.CreateLaunchTemplateVersionInput) error {
	obj.LaunchTemplateName = cr.Spec.ForProvider.LaunchTemplateName
	obj.LaunchTemplateId = cr.Spec.ForProvider.LaunchTemplateID
	return nil
}

func preObserve(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, obj *svcsdk.DescribeLaunchTemplateVersionsInput) error {
	obj.LaunchTemplateName = cr.Spec.ForProvider.LaunchTemplateName
	obj.LaunchTemplateId = cr.Spec.ForProvider.LaunchTemplateID
	obj.Versions = append(obj.Versions, aws.String(meta.GetExternalName(cr)))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, resp *svcsdk.CreateLaunchTemplateVersionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, strconv.FormatInt(*resp.LaunchTemplateVersion.VersionNumber, 10))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, obj *svcsdk.DescribeLaunchTemplateVersionsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func (e *external) deleter(ctx context.Context, mg cpresource.Managed) error {
	cr, _ := mg.(*svcapitypes.LaunchTemplateVersion)
	input := GenerateDeleteLaunchTemplateVersionInput(cr)
	_, err := e.client.DeleteLaunchTemplateVersionsWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

// GenerateDeleteLaunchTemplateVersionInput returns a deletion input.
func GenerateDeleteLaunchTemplateVersionInput(cr *svcapitypes.LaunchTemplateVersion) *svcsdk.DeleteLaunchTemplateVersionsInput {
	res := &svcsdk.DeleteLaunchTemplateVersionsInput{}
	res.SetDryRun(false)
	if cr.Spec.ForProvider.LaunchTemplateName != nil {
		res.SetLaunchTemplateName(pointer.StringValue(cr.Spec.ForProvider.LaunchTemplateName))
	}
	if cr.Spec.ForProvider.LaunchTemplateID != nil {
		res.SetLaunchTemplateId(pointer.StringValue(cr.Spec.ForProvider.LaunchTemplateID))
	}
	if meta.GetExternalName(cr) != "" {
		res.SetVersions(append(res.Versions, aws.String(meta.GetExternalName(cr))))
	}
	return res
}
