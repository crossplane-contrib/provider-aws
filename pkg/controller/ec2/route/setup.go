package route

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errMultipleItems = "retrieved multiple RouteTables for the given routeTableId"
)

// SetupRoute adds a controller that reconciles Route.
func SetupRoute(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RouteGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.observe = e.observer
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Route{},
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.LabelChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			))).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind(svcapitypes.RouteGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preCreate(_ context.Context, cr *svcapitypes.Route, obj *svcsdk.CreateRouteInput) error {
	obj.NatGatewayId = cr.Spec.ForProvider.NATGatewayID
	obj.TransitGatewayId = cr.Spec.ForProvider.TransitGatewayID
	obj.VpcPeeringConnectionId = cr.Spec.ForProvider.VPCPeeringConnectionID
	obj.RouteTableId = cr.Spec.ForProvider.RouteTableID
	obj.InstanceId = cr.Spec.ForProvider.InstanceID
	obj.GatewayId = cr.Spec.ForProvider.GatewayID
	return nil
}

func (e *external) observer(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.Route)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	route, err := e.findRouteByDestination(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if awsclients.StringValue(route.State) == svcsdk.RouteStateActive {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Route, obj *svcsdk.DeleteRouteInput) (bool, error) {
	obj.RouteTableId = cr.Spec.ForProvider.RouteTableID
	return false, nil
}

// findRouteByDestination returns the route corresponding to the specified IPv4/IPv6 destination.
// Returns NotFoundError if no route is found.
func (e *external) findRouteByDestination(ctx context.Context, cr *svcapitypes.Route) (*svcsdk.Route, error) {

	response, err := e.client.DescribeRouteTablesWithContext(ctx, &svcsdk.DescribeRouteTablesInput{
		RouteTableIds: []*string{cr.Spec.ForProvider.RouteTableID},
	})

	if err != nil {
		return nil, awsclients.Wrap(cpresource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.RouteTables) != 1 {
		return nil, errors.New(errMultipleItems)
	}

	for _, route := range response.RouteTables[0].Routes {
		if awsclients.StringValue(route.Origin) == svcsdk.RouteOriginCreateRoute {
			if awsclients.CIDRBlocksEqual(awsclients.StringValue(route.DestinationCidrBlock), awsclients.StringValue(cr.Spec.ForProvider.DestinationCIDRBlock)) {
				return route, nil
			}
		}

	}
	return nil, errors.New(errUnexpectedObject)
}
