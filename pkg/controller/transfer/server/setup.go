/*
Copyright 2021 The Crossplane Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	svcsdkapitransfer "github.com/aws/aws-sdk-go/service/transfer/transferiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/ssh"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/transfer/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

type custom struct {
	kube              client.Client
	client            svcsdkapitransfer.TransferAPI
	external          *external
	vpcEndpointClient vpcEndpointClient
}

// SetupServer adds a controller that reconciles Server.
func SetupServer(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServerGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&customConnector{connector: &connector{kube: mgr.GetClient()}, newClientFn: newVPCClient}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ServerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Server{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerInput) error {
	if meta.GetExternalName(cr) != "" {
		obj.ServerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DeleteServerInput) (bool, error) {
	obj.ServerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func (c *custom) postObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) { //nolint:gocyclo
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(obj.Server.State) {
	case string(svcapitypes.State_OFFLINE):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.State_ONLINE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.State_STARTING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.State_STOPPING):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.State_START_FAILED):
		cr.SetConditions(xpv1.ReconcileError(err))
	case string(svcapitypes.State_STOP_FAILED):
		cr.SetConditions(xpv1.ReconcileError(err))
	}
	obs.ConnectionDetails = managed.ConnectionDetails{
		"HostKeyFingerprint": []byte(pointer.StringValue(obj.Server.HostKeyFingerprint)),
	}
	// fetch endpoint details for EndpointType VPC only
	// for EndpointType 'PUBLIC' neither endpoint id nor endpoint name is present in the describe output/aws cli v2 output
	// endpoint name can only be found in the console
	if pointer.StringValue(cr.Spec.ForProvider.EndpointType) == "VPC" {
		vpcEndpointsResults, err := c.DescribeVpcEndpoint(obj)
		if err != nil {
			return managed.ExternalObservation{}, err
		}

		for i, vep := range vpcEndpointsResults {
			for j, dnsEntry := range vep.DnsEntries {
				key := fmt.Sprintf("endpoint.%d.dns.%d", i, j)
				obs.ConnectionDetails[key] = []byte(pointer.StringValue(dnsEntry.DnsName))
			}
		}
	}

	cr.Status.AtProvider.ARN = obj.Server.Arn
	cr.Status.AtProvider.Tags = utils.ConvertTagsFromDescribeServer(obj.Server.Tags)

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.ServerId))
	return managed.ExternalCreation{}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerInput) error {
	obj.Certificate = cr.Spec.ForProvider.Certificate
	obj.LoggingRole = cr.Spec.ForProvider.LoggingRole
	obj.EndpointDetails = &svcsdk.EndpointDetails{}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs) > 0 {
		obj.EndpointDetails.SecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs))
		copy(obj.EndpointDetails.SecurityGroupIds, cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs)
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs) > 0 {
		obj.EndpointDetails.SubnetIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs))
		copy(obj.EndpointDetails.SubnetIds, cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs)
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs) > 0 {
		obj.EndpointDetails.AddressAllocationIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs))
		copy(obj.EndpointDetails.AddressAllocationIds, cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs)
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID != nil {
		obj.EndpointDetails.VpcEndpointId = cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCID != nil {
		obj.EndpointDetails.VpcId = cr.Spec.ForProvider.CustomEndpointDetails.VPCID
	}

	return nil
}

func (c *custom) DescribeVpcEndpoint(obj *svcsdk.DescribeServerOutput) (dnsEntries []*ec2.VpcEndpoint, err error) {
	if obj.Server != nil && obj.Server.EndpointDetails != nil && obj.Server.EndpointDetails.VpcEndpointId != nil {
		describeEndpointOutput, err := c.vpcEndpointClient.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: []*string{
				obj.Server.EndpointDetails.VpcEndpointId,
			},
		})
		if err != nil {
			return []*ec2.VpcEndpoint{}, err
		}
		return describeEndpointOutput.VpcEndpoints, nil
	}
	return []*ec2.VpcEndpoint{}, nil
}

func lateInitialize(spec *svcapitypes.ServerParameters, obj *svcsdk.DescribeServerOutput) error {
	if spec.ProtocolDetails == nil {
		spec.ProtocolDetails = &svcapitypes.ProtocolDetails{}
	}
	if obj.Server.ProtocolDetails != nil {
		if spec.ProtocolDetails.As2Transports == nil {
			spec.ProtocolDetails.As2Transports = obj.Server.ProtocolDetails.As2Transports
		}
		if spec.ProtocolDetails.PassiveIP == nil {
			spec.ProtocolDetails.PassiveIP = obj.Server.ProtocolDetails.PassiveIp
		}
		if spec.ProtocolDetails.SetStatOption == nil {
			spec.ProtocolDetails.SetStatOption = obj.Server.ProtocolDetails.SetStatOption
		}
		if spec.ProtocolDetails.TLSSessionResumptionMode == nil {
			spec.ProtocolDetails.TLSSessionResumptionMode = obj.Server.ProtocolDetails.TlsSessionResumptionMode
		}
	}
	if spec.CustomEndpointDetails != nil && spec.CustomEndpointDetails.VPCEndpointID == nil && obj.Server.EndpointDetails.VpcEndpointId != nil {
		spec.CustomEndpointDetails.VPCEndpointID = obj.Server.EndpointDetails.VpcEndpointId
	}
	return nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Server, cur *svcsdk.DescribeServerOutput) (bool, string, error) {
	in := cr.Spec.ForProvider
	out := cur.Server

	if isNotUpToDate(in, out) {
		return false, "", nil
	}

	return true, "", nil

}

func isNotUpToDate(in svcapitypes.ServerParameters, out *svcsdk.DescribedServer) bool { //nolint:gocyclo
	if !cmp.Equal(in.Certificate, out.Certificate) {
		return true
	}

	if !cmp.Equal(in.EndpointType, out.EndpointType) {
		return true
	}

	if !cmp.Equal(in.IdentityProviderType, out.IdentityProviderType) {
		return true
	}

	if !cmp.Equal(in.LoggingRole, out.LoggingRole) {
		return true
	}

	if !cmp.Equal(in.PostAuthenticationLoginBanner, out.PostAuthenticationLoginBanner) {
		return true
	}

	if !cmp.Equal(in.PreAuthenticationLoginBanner, out.PreAuthenticationLoginBanner) {
		return true
	}

	if !cmp.Equal(in.SecurityPolicyName, out.SecurityPolicyName) {
		return true
	}

	if !cmp.Equal(in.StructuredLogDestinations, out.StructuredLogDestinations) {
		return true
	}

	if !isIdentityProviderDetailsUpToDate(in.IdentityProviderDetails, out.IdentityProviderDetails) {
		return true
	}

	if !isProtocolDetailsUpToDate(in.ProtocolDetails, out.ProtocolDetails) {
		return true
	}

	if !isCustomEndpointUpToDate(in.CustomEndpointDetails, out.EndpointDetails) {
		return true
	}

	if !isWorkflowDetailsUpToDate(in.WorkflowDetails, out.WorkflowDetails) {
		return true
	}

	if upToDate, _, _ := utils.DiffTags(in.Tags, out.Tags); !upToDate {
		return true
	}

	if !isHostKeyUpToDate(in.HostKey, out.HostKeyFingerprint) {
		return true
	}

	return false
}

func isIdentityProviderDetailsUpToDate(in *svcapitypes.IdentityProviderDetails, out *svcsdk.IdentityProviderDetails) bool {
	if in == nil && out == nil {
		return true
	}
	if in == nil || out == nil {
		return false
	}
	if !cmp.Equal(in.DirectoryID, out.DirectoryId) {
		return false
	}
	if !cmp.Equal(in.Function, out.Function) {
		return false
	}
	if !cmp.Equal(in.InvocationRole, out.InvocationRole) {
		return false
	}
	if !cmp.Equal(in.SftpAuthenticationMethods, out.SftpAuthenticationMethods) {
		return false
	}
	if !cmp.Equal(in.URL, out.Url) {
		return false
	}
	return true
}

func isProtocolDetailsUpToDate(in *svcapitypes.ProtocolDetails, out *svcsdk.ProtocolDetails) bool {
	if in == nil && out == nil {
		return true
	}
	if in == nil || out == nil {
		return false
	}
	if !cmp.Equal(in.As2Transports, out.As2Transports) {
		return false
	}
	if !cmp.Equal(in.PassiveIP, out.PassiveIp) {
		return false
	}
	if !cmp.Equal(in.SetStatOption, out.SetStatOption) {
		return false
	}
	if !cmp.Equal(in.TLSSessionResumptionMode, out.TlsSessionResumptionMode) {
		return false
	}
	return true
}

func isCustomEndpointUpToDate(in *svcapitypes.CustomEndpointDetails, out *svcsdk.EndpointDetails) bool {
	if in == nil && out == nil {
		return true
	}
	if in == nil || out == nil {
		return false
	}
	if !cmp.Equal(in.AddressAllocationIDs, out.AddressAllocationIds) {
		return false
	}
	if !cmp.Equal(in.SubnetIDs, out.SubnetIds) {
		return false
	}
	if !cmp.Equal(in.VPCEndpointID, out.VpcEndpointId) {
		return false
	}
	if !cmp.Equal(in.VPCID, out.VpcId) {
		return false
	}
	return true
}

func isWorkflowDetailsUpToDate(in *svcapitypes.WorkflowDetails, out *svcsdk.WorkflowDetails) bool { //nolint:gocyclo
	if in == nil && out == nil {
		return true
	}
	if in == nil || out == nil {
		return false
	}
	if len(in.OnPartialUpload) != len(out.OnPartialUpload) || len(in.OnUpload) != len(out.OnUpload) {
		return false
	}

	if len(in.OnPartialUpload) == 0 && len(in.OnUpload) == 0 {
		return true
	}

	apiTypesSort := func(a *svcapitypes.WorkflowDetail, b *svcapitypes.WorkflowDetail) int {
		return strings.Compare(*a.WorkflowID, *b.WorkflowID)
	}
	sdkSort := func(a *svcsdk.WorkflowDetail, b *svcsdk.WorkflowDetail) int {
		return strings.Compare(*a.WorkflowId, *b.WorkflowId)
	}
	compareApiSdk := func(a *svcapitypes.WorkflowDetail, b *svcsdk.WorkflowDetail) bool {
		return ptr.Deref(a.ExecutionRole, "") == ptr.Deref(b.ExecutionRole, "") && ptr.Deref(a.WorkflowID, "") == ptr.Deref(b.WorkflowId, "")
	}

	slices.SortFunc(in.OnPartialUpload, apiTypesSort)
	slices.SortFunc(in.OnUpload, apiTypesSort)
	slices.SortFunc(out.OnPartialUpload, sdkSort)
	slices.SortFunc(out.OnUpload, sdkSort)

	return slices.EqualFunc(in.OnPartialUpload, out.OnPartialUpload, compareApiSdk) && slices.EqualFunc(in.OnUpload, out.OnUpload, compareApiSdk)
}

func isHostKeyUpToDate(in *string, out *string) bool {
	// if there is no HostKey set, AWS generates a HostKey by itself. if the HostKey gets deleted from the spec, don't update.
	if in == nil {
		return true
	}
	if out == nil {
		return false
	}
	key, err := ssh.ParsePrivateKey([]byte(ptr.Deref(in, "")))
	if err != nil {
		panic(err)
	}
	fingerprint := ssh.FingerprintSHA256(key.PublicKey())
	currentFingerprint := strings.TrimSuffix(ptr.Deref(out, ""), "=")
	return fingerprint == currentFingerprint
}
