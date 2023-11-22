package transitgatewayroute

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupTransitGatewayRoute adds a controller that reconciles TransitGatewayRoutes.
func SetupTransitGatewayRoute(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RouteGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.preCreate = c.preCreate
			e.observe = e.observer
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		cpresource.ManagedKind(svcapitypes.TransitGatewayRouteGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.TransitGatewayRoute{}).
		Complete(r)
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.TransitGatewayRoute, obj *svcsdk.CreateTransitGatewayRouteInput) error {
	// need extra call for error:
	// cannot create TransitGatewayRoute in AWS: IncorrectState: tgw-rtb-xxx is in invalid state
	input := &svcsdk.DescribeTransitGatewayRouteTablesInput{}
	input.TransitGatewayRouteTableIds = append(input.TransitGatewayRouteTableIds, cr.Spec.ForProvider.TransitGatewayRouteTableID)
	rtbState, err := e.client.DescribeTransitGatewayRouteTablesWithContext(ctx, input)
	if err != nil {
		return err
	}

	if aws.StringValue(rtbState.TransitGatewayRouteTables[0].State) != string(svcapitypes.TransitGatewayRouteTableState_available) {
		return errors.New("referenced transitgateway-routetable is not available for routes " + aws.StringValue(rtbState.TransitGatewayRouteTables[0].State))
	}

	obj.TransitGatewayAttachmentId = cr.Spec.ForProvider.TransitGatewayAttachmentID
	obj.TransitGatewayRouteTableId = cr.Spec.ForProvider.TransitGatewayRouteTableID
	return nil
}

func (e *external) observer(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.TransitGatewayRoute)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	transitGatewayRoute, err := e.findRouteByDestination(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, nil //nolint:nilerr
	}

	switch aws.StringValue(transitGatewayRoute.State) {
	case string(svcsdk.TransitGatewayRouteStateActive), string(svcsdk.TransitGatewayRouteStateBlackhole):
		cr.SetConditions(xpv1.Available())
	case string(svcsdk.TransitGatewayRouteStatePending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TransitGatewayRouteState_deleting):
		cr.SetConditions(xpv1.Deleting())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func preDelete(_ context.Context, cr *svcapitypes.TransitGatewayRoute, obj *svcsdk.DeleteTransitGatewayRouteInput) (bool, error) {
	obj.TransitGatewayRouteTableId = cr.Spec.ForProvider.TransitGatewayRouteTableID
	return false, nil
}

// findRouteByDestination returns the Route corresponding to the specified DestinationCIDRBlock.
// Returns errUnexpectedObject if no Route is found.
func (e *external) findRouteByDestination(ctx context.Context, cr *svcapitypes.TransitGatewayRoute) (*svcsdk.TransitGatewayRoute, error) {

	response, err := e.client.SearchTransitGatewayRoutesWithContext(ctx, &svcsdk.SearchTransitGatewayRoutesInput{
		Filters: []*svcsdk.Filter{
			{
				Name:   aws.String("type"),
				Values: []*string{aws.String("static")},
			},
		},
		TransitGatewayRouteTableId: cr.Spec.ForProvider.TransitGatewayRouteTableID,
	})

	if err != nil {
		return nil, errorutils.Wrap(cpresource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	for _, route := range response.Routes {
		if route == nil {
			continue
		}

		if ec2.CIDRBlocksEqual(pointer.StringValue(route.DestinationCidrBlock), pointer.StringValue(cr.Spec.ForProvider.DestinationCIDRBlock)) {
			return route, nil
		}
	}

	return nil, errors.New(errUnexpectedObject)
}
