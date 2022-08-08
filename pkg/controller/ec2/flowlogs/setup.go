package flowlogs

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupFlowLogs adds a controller that reconciles FlowLogs
func SetupFlowLogs(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FlowLogsGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.observe = e.observer
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.FlowLogs{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.FlowLogsGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))

}

func preCreate(_ context.Context, cr *svcapitypes.FlowLogs, obj *svcsdk.CreateFlowLogsInput) error {
	if cr.Spec.ForProvider.DeliverLogsPermissionRole != nil {
		obj.DeliverLogsPermissionArn = cr.Spec.ForProvider.DeliverLogsPermissionRole
	} else {
		obj.DeliverLogsPermissionArn = cr.Spec.ForProvider.DeliverLogsPermissionARN
	}

	obj.ResourceIds, obj.ResourceType = determineResourceIdsAndType(cr)

	return nil
}

func determineResourceIdsAndType(cr *svcapitypes.FlowLogs) ([]*string, *string) {
	vpcResourceType := svcsdk.FlowLogsResourceTypeVpc
	transitGatewayResourceType := "TransitGateway"
	transitGatewayAttachmentResourceType := "TransitGatewayAttachment"
	subnetResourceType := svcsdk.FlowLogsResourceTypeSubnet
	networkInterfaceResourceType := svcsdk.FlowLogsResourceTypeNetworkInterface
	if cr.Spec.ForProvider.VPCID != nil {
		return []*string{cr.Spec.ForProvider.VPCID}, &vpcResourceType
	}

	if cr.Spec.ForProvider.TransitGatewayID != nil {
		return []*string{cr.Spec.ForProvider.TransitGatewayID}, &transitGatewayResourceType
	}

	if cr.Spec.ForProvider.TransitGatewayAttachmentID != nil {
		return []*string{cr.Spec.ForProvider.TransitGatewayAttachmentID}, &transitGatewayAttachmentResourceType
	}

	if cr.Spec.ForProvider.SubnetIDs != nil && len(cr.Spec.ForProvider.SubnetIDs) > 0 {
		return cr.Spec.ForProvider.SubnetIDs, &subnetResourceType
	}

	if cr.Spec.ForProvider.NetworkInterfaceID != nil {
		return []*string{cr.Spec.ForProvider.NetworkInterfaceID}, &networkInterfaceResourceType
	}

	return nil, nil
}

func (e *external) observer(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}
