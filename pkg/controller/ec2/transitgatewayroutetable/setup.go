package transitgatewayroutetable

import (
	"context"
	"sort"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errKubeUpdateFailed = "cannot update TransitGateway"
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.TransitGatewayRouteTable{}).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind(svcapitypes.TransitGatewayRouteTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func filterList(cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.DescribeTransitGatewayRouteTablesOutput) *svcsdk.DescribeTransitGatewayRouteTablesOutput {
	resp := &svcsdk.DescribeTransitGatewayRouteTablesOutput{}
	for _, TransitGatewayRouteTable := range obj.TransitGatewayRouteTables {
		if aws.StringValue(TransitGatewayRouteTable.TransitGatewayRouteTableId) == meta.GetExternalName(cr) {
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

	if aws.StringValue(tgwState.TransitGateways[0].State) != string(svcapitypes.TransitGatewayState_available) {
		return errors.New("referenced transitgateway is not available for routetable " + aws.StringValue(tgwState.TransitGateways[0].State))
	}

	obj.TransitGatewayId = cr.Spec.ForProvider.TransitGatewayID
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.CreateTransitGatewayRouteTableOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, aws.StringValue(obj.TransitGatewayRouteTable.TransitGatewayRouteTableId))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.TransitGatewayRouteTable, obj *svcsdk.DescribeTransitGatewayRouteTablesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(obj.TransitGatewayRouteTables[0].State) {
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

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd cpresource.Managed) error {
	cr, ok := mgd.(*svcapitypes.TransitGatewayRouteTable)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	var transitGatewayRouteTableTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "transit-gateway-route-table" {
			transitGatewayRouteTableTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range transitGatewayRouteTableTags.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range cpresource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	transitGatewayRouteTableTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	transitGatewayRouteTableTags.ResourceType = aws.String("transit-gateway-route-table")
	i := 0
	for k, v := range tagMap {
		transitGatewayRouteTableTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(transitGatewayRouteTableTags.Tags, func(i, j int) bool {
		return aws.StringValue(transitGatewayRouteTableTags.Tags[i].Key) < aws.StringValue(transitGatewayRouteTableTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&transitGatewayRouteTableTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
