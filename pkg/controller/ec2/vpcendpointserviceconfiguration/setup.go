package vpcendpointserviceconfiguration

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

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupVPCEndpointServiceConfiguration adds a controller that reconciles VPCEndpointServiceConfiguration.
func SetupVPCEndpointServiceConfiguration(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.VPCEndpointServiceConfigurationGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preCreate = preCreate
			e.filterList = filterList
			u := &updater{client: e.client}
			e.delete = u.delete
			e.preUpdate = u.preUpdate
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
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
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		cpresource.ManagedKind(svcapitypes.VPCEndpointServiceConfigurationGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.VPCEndpointServiceConfiguration{}).
		Complete(r)
}

func filterList(cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput) *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput {
	resp := &svcsdk.DescribeVpcEndpointServiceConfigurationsOutput{}
	for _, VPCEndpointServiceConfiguration := range obj.ServiceConfigurations {
		if aws.StringValue(VPCEndpointServiceConfiguration.ServiceId) == meta.GetExternalName(cr) {
			resp.ServiceConfigurations = append(resp.ServiceConfigurations, VPCEndpointServiceConfiguration)
			break
		}
	}
	return resp
}

func postObserve(_ context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider.ServiceConfiguration = GenerateObservation(obj.ServiceConfigurations[0])
	switch pointer.StringValue(obj.ServiceConfigurations[0].ServiceState) {
	case string(svcapitypes.ServiceState_Available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.ServiceState_Pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.ServiceState_Failed):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.ServiceState_Deleting):
		cr.SetConditions(xpv1.Deleting())
	}

	cr.Status.AtProvider.ServiceConfiguration.ServiceID = obj.ServiceConfigurations[0].ServiceId
	cr.Status.AtProvider.ServiceConfiguration.ServiceName = obj.ServiceConfigurations[0].ServiceName
	cr.Status.AtProvider.ServiceConfiguration.ServiceState = obj.ServiceConfigurations[0].ServiceState
	return obs, nil
}

func preCreate(ctx context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.CreateVpcEndpointServiceConfigurationInput) error {
	obj.ClientToken = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.GatewayLoadBalancerArns = append(obj.GatewayLoadBalancerArns, cr.Spec.ForProvider.GatewayLoadBalancerARNs...)
	obj.NetworkLoadBalancerArns = append(obj.NetworkLoadBalancerArns, cr.Spec.ForProvider.NetworkLoadBalancerARNs...)

	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.CreateVpcEndpointServiceConfigurationOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, aws.StringValue(obj.ServiceConfiguration.ServiceId))
	return cre, nil
}

func lateInitialize(cr *svcapitypes.VPCEndpointServiceConfigurationParameters, obj *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput) error {
	if cr.AcceptanceRequired == nil && obj.ServiceConfigurations[0].AcceptanceRequired != nil {
		cr.AcceptanceRequired = obj.ServiceConfigurations[0].AcceptanceRequired
	}
	return nil
}

type updater struct {
	client svcsdkapi.EC2API
}

func isUpToDate(_ context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput) (bool, string, error) {

	createGlbArns, deleteGlbArns := DifferenceARN(cr.Spec.ForProvider.GatewayLoadBalancerARNs, obj.ServiceConfigurations[0].GatewayLoadBalancerArns)
	if len(createGlbArns) != 0 || len(deleteGlbArns) != 0 {
		return false, "", nil
	}

	createNlbArns, deleteNlbArns := DifferenceARN(cr.Spec.ForProvider.NetworkLoadBalancerARNs, obj.ServiceConfigurations[0].NetworkLoadBalancerArns)
	if len(createNlbArns) != 0 || len(deleteNlbArns) != 0 {
		return false, "", nil
	}

	if pointer.StringValue(cr.Spec.ForProvider.PrivateDNSName) != pointer.StringValue(obj.ServiceConfigurations[0].PrivateDnsName) {
		return false, "", nil
	}

	if pointer.BoolValue(cr.Spec.ForProvider.AcceptanceRequired) != pointer.BoolValue(obj.ServiceConfigurations[0].AcceptanceRequired) {
		return false, "", nil
	}

	return true, "", nil
}

func (u *updater) preUpdate(_ context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.ModifyVpcEndpointServiceConfigurationInput) error {

	input := &svcsdk.DescribeVpcEndpointServiceConfigurationsInput{}
	input.ServiceIds = append(input.ServiceIds, pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)))

	resp, err := u.client.DescribeVpcEndpointServiceConfigurations(input)
	if err != nil {
		return errorutils.Wrap(err, errDescribe)
	}

	createGlbArns, deleteGlbArns := DifferenceARN(cr.Spec.ForProvider.GatewayLoadBalancerARNs, resp.ServiceConfigurations[0].GatewayLoadBalancerArns)
	createNlbArns, deleteNlbArns := DifferenceARN(cr.Spec.ForProvider.NetworkLoadBalancerARNs, resp.ServiceConfigurations[0].NetworkLoadBalancerArns)

	if len(createGlbArns) > 0 {
		obj.AddGatewayLoadBalancerArns = createGlbArns
	}

	if len(deleteGlbArns) > 0 {
		obj.RemoveGatewayLoadBalancerArns = deleteGlbArns
	}

	if len(createNlbArns) > 0 {
		obj.AddNetworkLoadBalancerArns = createNlbArns
	}

	if len(deleteNlbArns) > 0 {
		obj.RemoveNetworkLoadBalancerArns = deleteNlbArns
	}

	if cr.Spec.ForProvider.PrivateDNSName == nil && resp.ServiceConfigurations[0].PrivateDnsName != nil {
		obj.RemovePrivateDnsName = aws.Bool(true)
	}

	obj.ServiceId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	return nil
}

func (u *updater) delete(ctx context.Context, mg cpresource.Managed) error {

	cr, ok := mg.(*svcapitypes.VPCEndpointServiceConfiguration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())

	input := &svcsdk.DeleteVpcEndpointServiceConfigurationsInput{}
	input.ServiceIds = append(input.ServiceIds, pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)))

	_, err := u.client.DeleteVpcEndpointServiceConfigurationsWithContext(ctx, input)
	return errorutils.Wrap(cpresource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}

// DifferenceARN returns the lists of ARNs that need to be removed and added according
// to current and desired states.
func DifferenceARN(local []*string, remote []*string) ([]*string, []*string) {
	createKey := []*string{}
	removeKey := []*string{}
	m := map[string]int{}

	for _, value := range local {
		m[*value] = 1
	}

	for _, value := range remote {
		m[*value] += 2
	}

	for mKey, mVal := range m {
		// need for scopelint
		mKey2 := mKey
		if mVal == 1 {
			createKey = append(createKey, &mKey2)
		}

		if mVal == 2 {
			removeKey = append(removeKey, &mKey2)
		}
	}
	return createKey, removeKey
}

// GenerateObservation is used to produce v1alpha1.vpcendpointserviceconfigurationObservation
func GenerateObservation(obj *svcsdk.ServiceConfiguration) *svcapitypes.ServiceConfiguration {
	if obj == nil {
		return &svcapitypes.ServiceConfiguration{}
	}

	o := &svcapitypes.ServiceConfiguration{
		AvailabilityZones:    obj.AvailabilityZones,
		BaseEndpointDNSNames: obj.BaseEndpointDnsNames,
		ManagesVPCEndpoints:  obj.ManagesVpcEndpoints,
		PrivateDNSName:       obj.PrivateDnsName,
		ServiceID:            obj.ServiceId,
		ServiceName:          obj.ServiceName,
		ServiceState:         obj.ServiceState,
	}

	if obj.PrivateDnsNameConfiguration != nil {
		o.PrivateDNSNameConfiguration = &svcapitypes.PrivateDNSNameConfiguration{
			Name:  obj.PrivateDnsNameConfiguration.Name,
			State: obj.PrivateDnsNameConfiguration.State,
			Value: obj.PrivateDnsNameConfiguration.Value,
			Type:  obj.PrivateDnsNameConfiguration.Type,
		}
	}

	o.NetworkLoadBalancerARNs = append(o.NetworkLoadBalancerARNs, obj.NetworkLoadBalancerArns...)
	o.GatewayLoadBalancerARNs = append(o.GatewayLoadBalancerARNs, obj.GatewayLoadBalancerArns...)

	return o
}
