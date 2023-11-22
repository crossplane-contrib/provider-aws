package transitgatewayroutetable

import (
	"context"
	"strings"

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
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupTransitGatewayRouteTable adds a controller that reconciles TransitGatewayRouteTable.
func SetupTransitGatewayRouteTable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RouteGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = postObserve
			e.preCreate = c.preCreate
			e.postCreate = postCreate
			e.postDelete = postDelete
			e.filterList = filterList
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
		managed.WithInitializers(),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		cpresource.ManagedKind(svcapitypes.TransitGatewayRouteTableGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.TransitGatewayRouteTable{}).
		Complete(r)
}

func filterList(cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.DescribeTransitGatewayRouteTablesOutput) *svcsdk.DescribeTransitGatewayRouteTablesOutput {
	resp := &svcsdk.DescribeTransitGatewayRouteTablesOutput{}
	for _, TransitGatewayRouteTable := range obj.TransitGatewayRouteTables {
		if pointer.StringValue(TransitGatewayRouteTable.TransitGatewayRouteTableId) == meta.GetExternalName(cr) {
			resp.TransitGatewayRouteTables = append(resp.TransitGatewayRouteTables, TransitGatewayRouteTable)
			break
		}
	}
	return resp
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.CreateTransitGatewayRouteTableInput) error {
	// need extra call for error:
	// cannot create TransitGatewayRouteTable in AWS: IncorrectState: tgw-xxx is in invalid state
	input := &svcsdk.DescribeTransitGatewaysInput{}
	input.TransitGatewayIds = append(input.TransitGatewayIds, cr.Spec.ForProvider.TransitGatewayID)
	tgwState, err := e.client.DescribeTransitGatewaysWithContext(ctx, input)
	if err != nil {
		return err
	}

	if pointer.StringValue(tgwState.TransitGateways[0].State) != string(svcapitypes.TransitGatewayState_available) {
		return errors.New("referenced transitgateway is not available for routetable " + pointer.StringValue(tgwState.TransitGateways[0].State))
	}

	obj.TransitGatewayId = cr.Spec.ForProvider.TransitGatewayID
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.CreateTransitGatewayRouteTableOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(obj.TransitGatewayRouteTable.TransitGatewayRouteTableId))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.DescribeTransitGatewayRouteTablesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(obj.TransitGatewayRouteTables[0].State) {
	case string(svcapitypes.TransitGatewayRouteTableState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.TransitGatewayRouteTableState_pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TransitGatewayRouteTableState_deleting):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.TransitGatewayRouteTableState_deleted):
		// TransitGatewayRouteTable is in status deleted
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return obs, nil
}

func postDelete(_ context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.DeleteTransitGatewayRouteTableOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), string("IncorrectState")) {
			// skip: IncorrectState: tgw-rtb-xxx is in invalid state Error 400
			return nil
		}
		return err
	}
	return err
}
