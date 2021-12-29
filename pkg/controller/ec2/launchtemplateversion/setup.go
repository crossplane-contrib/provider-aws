package launchtemplateversion

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SetupLaunchTemplateVersion adds a controller that reconciles LaunchTemplateVersion.
func SetupLaunchTemplateVersion(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.LaunchTemplateVersionGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postCreate = postCreate
			e.postObserve = postObserve
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.LaunchTemplateVersion{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LaunchTemplateVersionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, obj *svcsdk.DescribeLaunchTemplateVersionsInput) error {
	obj.LaunchTemplateName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, resp *svcsdk.CreateLaunchTemplateVersionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.LaunchTemplateVersion.LaunchTemplateName))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.LaunchTemplateVersion, _ *svcsdk.DescribeLaunchTemplateVersionsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}
