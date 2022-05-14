package vpcendpointserviceconfiguration

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/features"
)

const (
	errKubeUpdateFailed = "cannot update VPCEndpointServiceConfiguration"
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

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.VPCEndpointServiceConfiguration{}).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind(svcapitypes.VPCEndpointServiceConfigurationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
	switch awsclients.StringValue(obj.ServiceConfigurations[0].ServiceState) {
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
	obj.ClientToken = awsclients.String(meta.GetExternalName(cr))
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

func isUpToDate(cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.DescribeVpcEndpointServiceConfigurationsOutput) (bool, error) {

	createGlbArns, deleteGlbArns := DifferenceARN(cr.Spec.ForProvider.GatewayLoadBalancerARNs, obj.ServiceConfigurations[0].GatewayLoadBalancerArns)
	if len(createGlbArns) != 0 || len(deleteGlbArns) != 0 {
		return false, nil
	}

	createNlbArns, deleteNlbArns := DifferenceARN(cr.Spec.ForProvider.NetworkLoadBalancerARNs, obj.ServiceConfigurations[0].NetworkLoadBalancerArns)
	if len(createNlbArns) != 0 || len(deleteNlbArns) != 0 {
		return false, nil
	}

	if awsclients.StringValue(cr.Spec.ForProvider.PrivateDNSName) != awsclients.StringValue(obj.ServiceConfigurations[0].PrivateDnsName) {
		return false, nil
	}

	if awsclients.BoolValue(cr.Spec.ForProvider.AcceptanceRequired) != awsclients.BoolValue(obj.ServiceConfigurations[0].AcceptanceRequired) {
		return false, nil
	}

	return true, nil
}

func (u *updater) preUpdate(_ context.Context, cr *svcapitypes.VPCEndpointServiceConfiguration, obj *svcsdk.ModifyVpcEndpointServiceConfigurationInput) error {

	input := &svcsdk.DescribeVpcEndpointServiceConfigurationsInput{}
	input.ServiceIds = append(input.ServiceIds, awsclients.String(meta.GetExternalName(cr)))

	resp, err := u.client.DescribeVpcEndpointServiceConfigurations(input)
	if err != nil {
		return awsclients.Wrap(err, errDescribe)
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

	obj.ServiceId = awsclients.String(meta.GetExternalName(cr))

	return nil
}

func (u *updater) delete(ctx context.Context, mg cpresource.Managed) error {

	cr, ok := mg.(*svcapitypes.VPCEndpointServiceConfiguration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())

	input := &svcsdk.DeleteVpcEndpointServiceConfigurationsInput{}
	input.ServiceIds = append(input.ServiceIds, awsclients.String(meta.GetExternalName(cr)))

	_, err := u.client.DeleteVpcEndpointServiceConfigurationsWithContext(ctx, input)
	return awsclients.Wrap(cpresource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd cpresource.Managed) error {
	cr, ok := mgd.(*svcapitypes.VPCEndpointServiceConfiguration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	var vpcEndpointTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "vpc-endpoint-service" {
			vpcEndpointTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range vpcEndpointTags.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range cpresource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	vpcEndpointTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	vpcEndpointTags.ResourceType = aws.String("vpc-endpoint-service")
	i := 0
	for k, v := range tagMap {
		vpcEndpointTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(vpcEndpointTags.Tags, func(i, j int) bool {
		return aws.StringValue(vpcEndpointTags.Tags[i].Key) < aws.StringValue(vpcEndpointTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&vpcEndpointTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
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

// GenerateObservation is used to produce v1beta1.ClusterObservation from
// ekstypes.Cluster.
func GenerateObservation(obj *svcsdk.ServiceConfiguration) *svcapitypes.ServiceConfiguration { // nolint:gocyclo
	if obj == nil {
		return &svcapitypes.ServiceConfiguration{}
	}

	o := &svcapitypes.ServiceConfiguration{
		AvailabilityZones:    obj.AvailabilityZones,
		BaseEndpointDNSNames: obj.BaseEndpointDnsNames,
		ManagesVPCEndpoints:  obj.ManagesVpcEndpoints,
		PrivateDNSName:       obj.PrivateDnsName,
		PrivateDNSNameConfiguration: &svcapitypes.PrivateDNSNameConfiguration{
			Name:  obj.PrivateDnsNameConfiguration.Name,
			State: obj.PrivateDnsNameConfiguration.State,
			Value: obj.PrivateDnsNameConfiguration.Value,
			Type:  obj.PrivateDnsNameConfiguration.Type,
		},
		ServiceID:    obj.ServiceId,
		ServiceName:  obj.ServiceName,
		ServiceState: obj.ServiceState,
	}

	o.NetworkLoadBalancerARNs = append(o.NetworkLoadBalancerARNs, obj.NetworkLoadBalancerArns...)
	o.GatewayLoadBalancerARNs = append(o.GatewayLoadBalancerARNs, obj.GatewayLoadBalancerArns...)

	return o
}
