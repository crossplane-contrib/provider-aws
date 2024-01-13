package transitgatewayvpcattachment

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
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupTransitGatewayVPCAttachment adds a controller that reconciles TransitGatewayVPCAttachment.
func SetupTransitGatewayVPCAttachment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.TransitGatewayVPCAttachmentGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preCreate = c.preCreate
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
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithInitializers(),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.TransitGatewayVPCAttachmentGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.TransitGatewayVPCAttachment{}).
		Complete(r)
}

func filterList(cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput) *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput {
	resp := &svcsdk.DescribeTransitGatewayVpcAttachmentsOutput{}
	for _, transitGatewayVpcAttachment := range obj.TransitGatewayVpcAttachments {
		if aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId) == meta.GetExternalName(cr) {
			resp.TransitGatewayVpcAttachments = append(resp.TransitGatewayVpcAttachments, transitGatewayVpcAttachment)
			break
		}
	}
	return resp
}

func postObserve(_ context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(obj.TransitGatewayVpcAttachments[0].State) {
	case string(svcapitypes.TransitGatewayAttachmentState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.TransitGatewayAttachmentState_pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TransitGatewayAttachmentState_modifying):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.TransitGatewayAttachmentState_deleting):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.TransitGatewayAttachmentState_deleted):
		// TransitGatewayAttachment is in status deleted - and really removed after 6 hours in aws
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return obs, nil
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.CreateTransitGatewayVpcAttachmentInput) error {
	// need extra call for error:
	// cannot create TransitGatewayVPCAttachment in AWS: IncorrectState: tgw is in invalid state
	input := &svcsdk.DescribeTransitGatewaysInput{}
	input.TransitGatewayIds = append(input.TransitGatewayIds, cr.Spec.ForProvider.TransitGatewayID)
	tgwState, err := e.client.DescribeTransitGatewaysWithContext(ctx, input)
	if err != nil {
		return err
	}

	if pointer.StringValue(tgwState.TransitGateways[0].State) != string(svcapitypes.TransitGatewayState_available) {
		return errors.New("referenced transitgateway is not available for vpcattachment " + pointer.StringValue(tgwState.TransitGateways[0].State))
	}

	obj.VpcId = cr.Spec.ForProvider.VPCID
	obj.TransitGatewayId = cr.Spec.ForProvider.TransitGatewayID
	obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs

	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.CreateTransitGatewayVpcAttachmentOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	// set transitgatewayvpcattachment id as external name annotation on k8s object after creation
	meta.SetExternalName(cr, aws.StringValue(obj.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	return cre, nil
}
