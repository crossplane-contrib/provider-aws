package rule

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupRule adds a controller that reconciles Rule.
func SetupRule(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RuleGroupKind)

	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
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
		resource.ManagedKind(svcapitypes.RuleGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Rule{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Rule, obj *svcsdk.DescribeRulesInput) error {
	obj.RuleArns = append(obj.RuleArns, aws.String(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Rule, resp *svcsdk.DescribeRulesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Set the condition to Available
	cr.SetConditions(xpv1.Available())

	// Make sure the status fields are properly populated
	if len(resp.Rules) > 0 {
		if cr.Status.AtProvider.Rules == nil {
			cr.Status.AtProvider.Rules = []*svcapitypes.Rule_SDK{}
		}

		// This will keep the status information synchronized with AWS
		rule := resp.Rules[0]
		if rule.RuleArn != nil {
			found := false
			for _, existingRule := range cr.Status.AtProvider.Rules {
				if existingRule.RuleARN != nil && *existingRule.RuleARN == *rule.RuleArn {
					found = true
					break
				}
			}

			if !found {
				ruleSDK := &svcapitypes.Rule_SDK{
					RuleARN: rule.RuleArn,
				}
				if rule.Priority != nil {
					ruleSDK.Priority = rule.Priority
				}
				if rule.IsDefault != nil {
					ruleSDK.IsDefault = rule.IsDefault
				}
				cr.Status.AtProvider.Rules = append(cr.Status.AtProvider.Rules, ruleSDK)
			}
		}
	}

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Rule, obs *svcsdk.CreateRuleInput) error {
	obs.ListenerArn = cr.Spec.ForProvider.ListenerARN

	// Ensure all conditions have the Field property set based on which condition config is present
	if obs.Conditions != nil {
		for _, condition := range obs.Conditions {
			// If Field is not set, determine it from the condition type , change to switch case
			if condition.Field == nil || *condition.Field == "" {
				switch {
				case condition.PathPatternConfig != nil:
					condition.Field = aws.String("path-pattern")
				case condition.HostHeaderConfig != nil:
					condition.Field = aws.String("host-header")
				case condition.HttpHeaderConfig != nil:
					condition.Field = aws.String("http-header")
				case condition.HttpRequestMethodConfig != nil:
					condition.Field = aws.String("http-request-method")
				case condition.QueryStringConfig != nil:
					condition.Field = aws.String("query-string")
				case condition.SourceIpConfig != nil:
					condition.Field = aws.String("source-ip")
				}
			}
		}
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Rule, resp *svcsdk.CreateRuleOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.Rules[0].RuleArn))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Rule, obj *svcsdk.DeleteRuleInput) (bool, error) {
	if meta.GetExternalName(cr) == "" {
		return true, nil
	}
	obj.RuleArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
