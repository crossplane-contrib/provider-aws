package vpcendpoint

import (
	"context"
	"strings"

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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	legacypolicy "github.com/crossplane-contrib/provider-aws/pkg/utils/policy/old"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupVPCEndpoint adds a controller that reconciles VPCEndpoint.
func SetupVPCEndpoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.VPCEndpointGroupKind)
	opts := []option{setupExternal}

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
		cpresource.ManagedKind(svcapitypes.VPCEndpointGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(cpresource.DesiredStateChanged()).
		For(&svcapitypes.VPCEndpoint{}).
		Complete(r)
}

func setupExternal(e *external) {
	c := &custom{client: e.client, kube: e.kube}
	e.delete = c.delete
	e.preCreate = preCreate
	e.postCreate = postCreate
	e.postObserve = postObserve
	e.isUpToDate = isUpToDate
	e.preUpdate = c.preUpdate
	e.postUpdate = postUpdate
	e.filterList = filterList
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func preCreate(_ context.Context, cr *svcapitypes.VPCEndpoint, obj *svcsdk.CreateVpcEndpointInput) error {
	obj.VpcId = cr.Spec.ForProvider.VPCID
	obj.ClientToken = pointer.ToOrNilIfZeroValue(string(cr.UID))
	// Clear SGs, RTs, and Subnets if they're empty
	if len(cr.Spec.ForProvider.SecurityGroupIDs) == 0 {
		obj.SecurityGroupIds = nil
	} else {
		obj.SecurityGroupIds = cr.Spec.ForProvider.SecurityGroupIDs
	}

	if len(cr.Spec.ForProvider.RouteTableIDs) == 0 {
		obj.RouteTableIds = nil
	} else {
		obj.RouteTableIds = cr.Spec.ForProvider.RouteTableIDs
	}

	if len(cr.Spec.ForProvider.SubnetIDs) == 0 {
		obj.SubnetIds = nil
	} else {
		obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs
	}

	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.VPCEndpoint, obj *svcsdk.CreateVpcEndpointOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil || obj.VpcEndpoint == nil {
		return managed.ExternalCreation{}, err
	}

	// set vpc endpoint id as external name annotation on k8s object after creation
	meta.SetExternalName(cr, aws.StringValue(obj.VpcEndpoint.VpcEndpointId))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.VPCEndpoint, resp *svcsdk.DescribeVpcEndpointsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Load DNS Entry as connection detail
	if len(resp.VpcEndpoints[0].DnsEntries) != 0 && pointer.StringValue(resp.VpcEndpoints[0].DnsEntries[0].DnsName) != "" {
		obs.ConnectionDetails = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(pointer.StringValue(resp.VpcEndpoints[0].DnsEntries[0].DnsName)),
		}
	}

	cr.Status.AtProvider = generateVPCEndpointObservation(resp.VpcEndpoints[0])

	switch pointer.StringValue(resp.VpcEndpoints[0].State) {
	case "available":
		cr.SetConditions(xpv1.Available())
	case "pending", "pending-acceptance":
		cr.SetConditions(xpv1.Creating())
	case "deleted":
		cr.SetConditions(xpv1.Unavailable())
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
	}

	return obs, nil
}

// isUpToDate checks for the following mutable fields for the VPCEndpoint in upstream AWS
func isUpToDate(_ context.Context, cr *svcapitypes.VPCEndpoint, obj *svcsdk.DescribeVpcEndpointsOutput) (bool, string, error) {
	// Check subnets
	if !listCompareStringPtrIsSame(obj.VpcEndpoints[0].SubnetIds, cr.Spec.ForProvider.SubnetIDs) {
		return false, "", nil
	}

	// Check Route Tables
	if !listCompareStringPtrIsSame(obj.VpcEndpoints[0].RouteTableIds, cr.Spec.ForProvider.RouteTableIDs) {
		return false, "", nil
	}

	// Check Security Groups
	upstreamSGs := obj.VpcEndpoints[0].Groups
	if len(upstreamSGs) != len(cr.Spec.ForProvider.SecurityGroupIDs) {
		return false, "", nil
	}

sgCompare:
	for _, declaredSG := range cr.Spec.ForProvider.SecurityGroupIDs {
		for _, upstreamSG := range upstreamSGs {
			if pointer.StringValue(declaredSG) == pointer.StringValue(upstreamSG.GroupId) {
				continue sgCompare
			}
		}
		// declaredSG not found in upstream AWS
		return false, "", nil
	}

	// Check policyDocument
	defaultPolicyEndpoint := aws.String("{\"Statement\":[{\"Action\":\"*\",\"Effect\": \"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}]}")
	defaultPolicyGateway := aws.String("{\"Version\":\"2008-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"*\",\"Resource\":\"*\"}]}")
	declaredPolicy := cr.Spec.ForProvider.PolicyDocument
	upstreamPolicy := obj.VpcEndpoints[0].PolicyDocument

	// If no declared policy, we expect the result to be equivalent to the default policy
	if aws.StringValue(declaredPolicy) == "" {
		return legacypolicy.IsPolicyUpToDate(upstreamPolicy, defaultPolicyEndpoint) || legacypolicy.IsPolicyUpToDate(upstreamPolicy, defaultPolicyGateway), "", nil
	}
	return legacypolicy.IsPolicyUpToDate(upstreamPolicy, declaredPolicy), "", nil
}

// preUpdate adds the mutable fields into the update request input
func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.VPCEndpoint, obj *svcsdk.ModifyVpcEndpointInput) error {
	obj.VpcEndpointId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// Add fields to upstream AWS
	obj.SetAddSecurityGroupIds(cr.Spec.ForProvider.SecurityGroupIDs)
	obj.SetAddSubnetIds(cr.Spec.ForProvider.SubnetIDs)
	obj.SetAddRouteTableIds(cr.Spec.ForProvider.RouteTableIDs)
	obj.SetPolicyDocument(aws.StringValue(cr.Spec.ForProvider.PolicyDocument))

	// Remove fields from upstream AWS
	upstream, err := e.client.DescribeVpcEndpoints(&svcsdk.DescribeVpcEndpointsInput{VpcEndpointIds: []*string{obj.VpcEndpointId}})
	if err != nil {
		return err
	}

	removeSubnets := listSubtractFromStringPtr(upstream.VpcEndpoints[0].SubnetIds, cr.Spec.ForProvider.SubnetIDs)
	removeRTs := listSubtractFromStringPtr(upstream.VpcEndpoints[0].RouteTableIds, cr.Spec.ForProvider.RouteTableIDs)

	removeSGs := []*string{}
sgCompare:
	for _, upstreamSG := range upstream.VpcEndpoints[0].Groups {
		for _, crSG := range cr.Spec.ForProvider.SecurityGroupIDs {
			if aws.StringValue(upstreamSG.GroupId) == aws.StringValue(crSG) {
				continue sgCompare
			}
		}
		removeSGs = append(removeSGs, upstreamSG.GroupId)
	}

	obj.SetRemoveSubnetIds(removeSubnets)
	obj.SetRemoveSecurityGroupIds(removeSGs)
	obj.SetRemoveRouteTableIds(removeRTs)

	formatModifyVpcEndpointInput(obj)
	return nil
}

func (e *custom) delete(_ context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.VPCEndpoint)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	// Generate Deletion Input
	deleteInput := &svcsdk.DeleteVpcEndpointsInput{}
	externalName := meta.GetExternalName(cr)
	deleteInput.SetVpcEndpointIds([]*string{&externalName})

	// Delete
	_, err := e.client.DeleteVpcEndpoints(deleteInput)
	return err
}

func postUpdate(_ context.Context, cr *svcapitypes.VPCEndpoint, resp *svcsdk.ModifyVpcEndpointOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Set it as creating until observation corrects it
	cr.SetConditions(xpv1.Creating())
	return upd, nil
}

func filterList(cr *svcapitypes.VPCEndpoint, obj *svcsdk.DescribeVpcEndpointsOutput) *svcsdk.DescribeVpcEndpointsOutput {
	connectionIdentifier := aws.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeVpcEndpointsOutput{}
	for _, vpcEndpoint := range obj.VpcEndpoints {
		if aws.StringValue(vpcEndpoint.VpcEndpointId) == aws.StringValue(connectionIdentifier) {
			resp.VpcEndpoints = append(resp.VpcEndpoints, vpcEndpoint)
			break
		}
	}
	return resp
}

func generateVPCEndpointObservation(vpcEndpoint *svcsdk.VpcEndpoint) svcapitypes.VPCEndpointObservation {
	vpcEndpointObservation := svcapitypes.VPCEndpointObservation{}

	// Mapping vpcEndpoint -> vpcEndpoint_SDK
	vpcEndpointObservation.CreationTimestamp = &v1.Time{
		Time: *vpcEndpoint.CreationTimestamp,
	}
	vpcEndpointObservation.DNSEntries = []*svcapitypes.DNSEntry{}
	for _, dnsEntry := range vpcEndpoint.DnsEntries {
		dnsEntry := svcapitypes.DNSEntry{
			DNSName:      dnsEntry.DnsName,
			HostedZoneID: dnsEntry.HostedZoneId,
		}
		vpcEndpointObservation.DNSEntries = append(vpcEndpointObservation.DNSEntries, &dnsEntry)
	}
	vpcEndpointObservation.State = vpcEndpoint.State

	return vpcEndpointObservation
}

// formatModifyVpcEndpointInput takes in a ModifyVpcEndpointInput, and sets
// fields containing an empty list to nil
func formatModifyVpcEndpointInput(obj *svcsdk.ModifyVpcEndpointInput) {
	if len(obj.AddSecurityGroupIds) == 0 {
		obj.AddSecurityGroupIds = nil
	}
	if len(obj.RemoveSecurityGroupIds) == 0 {
		obj.RemoveSecurityGroupIds = nil
	}
	if len(obj.AddRouteTableIds) == 0 {
		obj.AddRouteTableIds = nil
	}
	if len(obj.RemoveRouteTableIds) == 0 {
		obj.RemoveRouteTableIds = nil
	}
	if len(obj.AddSubnetIds) == 0 {
		obj.AddSubnetIds = nil
	}
	if len(obj.RemoveSubnetIds) == 0 {
		obj.RemoveSubnetIds = nil
	}
	if strings.TrimSpace(aws.StringValue(obj.PolicyDocument)) == "" {
		obj.PolicyDocument = nil
	}
}

// listSubtractFromStringPtr takes in 2 list of string pointers
// ([]*string) "base", "subtract", and returns a "result" list
// of string pointers where "result" = "base" - "subtract".
// Comparisons of the underlying string is done
//
// Example:
//
//	"base": ["a", "b", "g", "x"]
//	"subtract": ["b", "x", "y"]
//	"result": ["a", "g"]
func listSubtractFromStringPtr(base, subtract []*string) []*string {
	result := []*string{}

compare:
	for _, baseElem := range base {
		for _, subtractElem := range subtract {
			if aws.StringValue(baseElem) == aws.StringValue(subtractElem) {
				continue compare
			}
		}
		result = append(result, baseElem)
	}

	return result
}

// listCompareStringPtrIsSame takes in 2 list of string pointers,
// and returns a true on the following condition:
// 1. The length of both lists are the same
// 2. All values in listA can be found in listB
// Warning:
// This function assumes that the values in both lists are unique, that is,
// all values in listA is distinct, and all values in listB is distinct.
func listCompareStringPtrIsSame(listA, listB []*string) bool {
	if len(listA) != len(listB) {
		return false
	}

compare:
	for _, elemA := range listA {
		for _, elemB := range listB {
			if pointer.StringValue(elemA) == pointer.StringValue(elemB) {
				continue compare
			}
		}
		return false
	}

	return true
}
