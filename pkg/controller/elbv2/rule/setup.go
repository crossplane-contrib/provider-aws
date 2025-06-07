// Package rule provides controllers for managing ELBv2 listener rules.
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
	"github.com/pkg/errors"
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
			e.preUpdate = preUpdate
			e.isUpToDate = isUpToDate
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

	cr.SetConditions(xpv1.Available())

	if len(resp.Rules) > 0 {
		syncRuleToStatus(cr, resp.Rules[0])
	}

	return obs, nil
}

func syncRuleToStatus(cr *svcapitypes.Rule, rule *svcsdk.Rule) {
	if rule.RuleArn == nil {
		return
	}

	if cr.Status.AtProvider.Rules == nil {
		cr.Status.AtProvider.Rules = []*svcapitypes.Rule_SDK{}
	}

	for _, existingRule := range cr.Status.AtProvider.Rules {
		if existingRule.RuleARN != nil && *existingRule.RuleARN == *rule.RuleArn {
			return
		}
	}

	newRule := &svcapitypes.Rule_SDK{
		RuleARN: rule.RuleArn,
	}

	if rule.Priority != nil {
		newRule.Priority = rule.Priority
	}
	if rule.IsDefault != nil {
		newRule.IsDefault = rule.IsDefault
	}

	cr.Status.AtProvider.Rules = append(cr.Status.AtProvider.Rules, newRule)
}

func preCreate(_ context.Context, cr *svcapitypes.Rule, obs *svcsdk.CreateRuleInput) error {
	obs.ListenerArn = cr.Spec.ForProvider.ListenerARN
	obs.Priority = cr.Spec.ForProvider.Priority

	if obs.Conditions != nil {
		for _, condition := range obs.Conditions {
			if condition.Field == nil || *condition.Field == "" {
				condition.Field = inferConditionField(condition)
			}
		}
	}

	return nil
}

func inferConditionField(condition *svcsdk.RuleCondition) *string {
	switch {
	case condition.PathPatternConfig != nil:
		return aws.String("path-pattern")
	case condition.HostHeaderConfig != nil:
		return aws.String("host-header")
	case condition.HttpHeaderConfig != nil:
		return aws.String("http-header")
	case condition.HttpRequestMethodConfig != nil:
		return aws.String("http-request-method")
	case condition.QueryStringConfig != nil:
		return aws.String("query-string")
	case condition.SourceIpConfig != nil:
		return aws.String("source-ip")
	default:
		return nil
	}
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

func preUpdate(_ context.Context, cr *svcapitypes.Rule, obj *svcsdk.ModifyRuleInput) error {
	if meta.GetExternalName(cr) == "" {
		return errors.New("rule ARN is not set")
	}
	obj.RuleArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func isUpToDate(_ context.Context, _ *svcapitypes.Rule, _ *svcsdk.DescribeRulesOutput) (bool, string, error) {
	// TODO: Implement isUpToDate , need to compare actions and conditions correctly
	return true, "listener rule is up to date", nil
}
