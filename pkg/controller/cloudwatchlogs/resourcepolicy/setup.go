package resourcepolicy

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// SetupResourcePolicy adds a controller that reconciles ResourcePolicy.
func SetupResourcePolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ResourcePolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	opts := []option{
		func(e *external) {
			c := &custom{client: e.client}
			e.filterList = filterList
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.postObserve = postObserve
			e.isUpToDate = isUpToDate
			e.update = c.update
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ResourcePolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ResourcePolicyGroupVersionKind),
			managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type custom struct {
	client svcsdkapi.CloudWatchLogsAPI
}

func preCreate(_ context.Context, cr *svcapitypes.ResourcePolicy, obj *svcsdk.PutResourcePolicyInput) error {
	obj.PolicyName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.ResourcePolicy, obj *svcsdk.DeleteResourcePolicyInput) (bool, error) {
	obj.PolicyName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResourcePolicy, _ *svcsdk.DescribeResourcePoliciesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.ResourcePolicy, obj *svcsdk.DescribeResourcePoliciesOutput) (bool, string, error) {
	// Check if the policy exists
	for _, policy := range obj.ResourcePolicies {
		if policy.PolicyName != nil && *policy.PolicyName == meta.GetExternalName(cr) {
			// Use existing method from iam to compare policy documents
			return iam.IsPolicyDocumentUpToDate(*cr.Spec.ForProvider.PolicyDocument, policy.PolicyDocument)
		}
	}
	return false, "", nil
}

func filterList(cr *svcapitypes.ResourcePolicy, obj *svcsdk.DescribeResourcePoliciesOutput) *svcsdk.DescribeResourcePoliciesOutput {
	resourcePolicyIdentifier := pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeResourcePoliciesOutput{}

	for _, resourcePolicy := range obj.ResourcePolicies {
		if pointer.StringValue(resourcePolicy.PolicyName) == pointer.StringValue(resourcePolicyIdentifier) {
			resp.ResourcePolicies = append(resp.ResourcePolicies, resourcePolicy)
			break
		}
	}
	return resp
}

func preUpdate(_ context.Context, cr *svcapitypes.ResourcePolicy, obj *svcsdk.PutResourcePolicyInput) {
	obj.PolicyName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
}

func postUpdate(_ context.Context, _ *svcapitypes.ResourcePolicy, _ *svcsdk.PutResourcePolicyOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}

func (e *custom) update(ctx context.Context, cr *svcapitypes.ResourcePolicy) (managed.ExternalUpdate, error) {
	input := GeneratePutResourcePolicyInput(cr)
	preUpdate(ctx, cr, input)
	resp, err := e.client.PutResourcePolicyWithContext(ctx, input)
	return postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}
