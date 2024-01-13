package resolverendpoint

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/route53resolver"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	route53resolverv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupResolverEndpoint adds a controller that reconciles ResolverEndpoints
func SetupResolverEndpoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(route53resolverv1alpha1.ResolverEndpointGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.preUpdate = preUpdate
			e.postObserve = postObserve
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
		cpresource.ManagedKind(route53resolverv1alpha1.ResolverEndpointGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&route53resolverv1alpha1.ResolverEndpoint{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *route53resolverv1alpha1.ResolverEndpoint, obj *svcsdk.GetResolverEndpointInput) error {
	obj.ResolverEndpointId = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *route53resolverv1alpha1.ResolverEndpoint, obj *svcsdk.CreateResolverEndpointInput) error {
	obj.CreatorRequestId = aws.String(string(cr.GetObjectMeta().GetUID()))
	for _, sg := range cr.Spec.ForProvider.SecurityGroupIDs {
		obj.SecurityGroupIds = append(obj.SecurityGroupIds, aws.String(sg))
	}
	if len(cr.Spec.ForProvider.IPAddresses) > 0 {
		obj.IpAddresses = make([]*svcsdk.IpAddressRequest, len(cr.Spec.ForProvider.IPAddresses))
	}
	for i, ip := range cr.Spec.ForProvider.IPAddresses {
		if ip != nil {
			obj.IpAddresses[i] = &svcsdk.IpAddressRequest{
				SubnetId: ip.SubnetID,
				Ip:       ip.IP,
			}
		}
	}
	return nil
}

func postCreate(_ context.Context, cr *route53resolverv1alpha1.ResolverEndpoint, obj *svcsdk.CreateResolverEndpointOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	meta.SetExternalName(cr, aws.StringValue(obj.ResolverEndpoint.Id))
	return cre, err
}

func preDelete(_ context.Context, cr *route53resolverv1alpha1.ResolverEndpoint, obj *svcsdk.DeleteResolverEndpointInput) (bool, error) {
	obj.ResolverEndpointId = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func preUpdate(_ context.Context, cr *route53resolverv1alpha1.ResolverEndpoint, obj *svcsdk.UpdateResolverEndpointInput) error {
	obj.ResolverEndpointId = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResolverEndpoint, obj *svcsdk.GetResolverEndpointOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(obj.ResolverEndpoint.Status) {
	case string(svcapitypes.ResolverEndpointStatus_SDK_OPERATIONAL):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.ResolverEndpointStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.ResolverEndpointStatus_SDK_UPDATING), string(svcapitypes.ResolverEndpointStatus_SDK_ACTION_NEEDED):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.ResolverEndpointStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}

	return obs, err
}
