package server

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/transfer/utils"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

func (e *external) UpdateServer(ctx context.Context, cr *svcapitypes.Server) (managed.ExternalUpdate, error) {
	input := GenerateUpdateServerInput(cr)
	_, err := e.client.UpdateServerWithContext(ctx, input)

	isUpToDate, add, remove := utils.DiffTagsFromStatus(cr.Spec.ForProvider.Tags, cr.Status.AtProvider.Tags)
	if isUpToDate {
		return managed.ExternalUpdate{}, nil
	}
	if len(add) > 0 {
		_, err := e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			Arn:  cr.Status.AtProvider.ARN,
			Tags: add,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot tag resource")
		}
	}
	if len(remove) > 0 {
		_, err := e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			Arn:     cr.Status.AtProvider.ARN,
			TagKeys: remove,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot tag resource")
		}
	}

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
}

func GenerateUpdateServerInput(cr *svcapitypes.Server) *svcsdk.UpdateServerInput { //nolint:gocyclo
	res := &svcsdk.UpdateServerInput{}

	if cr.Spec.ForProvider.EndpointType != nil {
		res.SetEndpointType(*cr.Spec.ForProvider.EndpointType)
	}
	if cr.Spec.ForProvider.HostKey != nil {
		res.SetHostKey(*cr.Spec.ForProvider.HostKey)
	}
	if cr.Spec.ForProvider.IdentityProviderDetails != nil {
		f4 := &svcsdk.IdentityProviderDetails{}
		if cr.Spec.ForProvider.IdentityProviderDetails.DirectoryID != nil {
			f4.SetDirectoryId(*cr.Spec.ForProvider.IdentityProviderDetails.DirectoryID)
		}
		if cr.Spec.ForProvider.IdentityProviderDetails.Function != nil {
			f4.SetFunction(*cr.Spec.ForProvider.IdentityProviderDetails.Function)
		}
		if cr.Spec.ForProvider.IdentityProviderDetails.InvocationRole != nil {
			f4.SetInvocationRole(*cr.Spec.ForProvider.IdentityProviderDetails.InvocationRole)
		}
		if cr.Spec.ForProvider.IdentityProviderDetails.SftpAuthenticationMethods != nil {
			f4.SetSftpAuthenticationMethods(*cr.Spec.ForProvider.IdentityProviderDetails.SftpAuthenticationMethods)
		}
		if cr.Spec.ForProvider.IdentityProviderDetails.URL != nil {
			f4.SetUrl(*cr.Spec.ForProvider.IdentityProviderDetails.URL)
		}
		res.SetIdentityProviderDetails(f4)
	}
	if cr.Spec.ForProvider.PostAuthenticationLoginBanner != nil {
		res.SetPostAuthenticationLoginBanner(*cr.Spec.ForProvider.PostAuthenticationLoginBanner)
	}
	if cr.Spec.ForProvider.PreAuthenticationLoginBanner != nil {
		res.SetPreAuthenticationLoginBanner(*cr.Spec.ForProvider.PreAuthenticationLoginBanner)
	}
	if cr.Spec.ForProvider.ProtocolDetails != nil {
		f8 := &svcsdk.ProtocolDetails{}
		if cr.Spec.ForProvider.ProtocolDetails.As2Transports != nil {
			f8f0 := []*string{}
			for _, f8f0iter := range cr.Spec.ForProvider.ProtocolDetails.As2Transports {
				var f8f0elem string = *f8f0iter
				f8f0 = append(f8f0, &f8f0elem)
			}
			f8.SetAs2Transports(f8f0)
		}
		if cr.Spec.ForProvider.ProtocolDetails.PassiveIP != nil {
			f8.SetPassiveIp(*cr.Spec.ForProvider.ProtocolDetails.PassiveIP)
		}
		if cr.Spec.ForProvider.ProtocolDetails.SetStatOption != nil {
			f8.SetSetStatOption(*cr.Spec.ForProvider.ProtocolDetails.SetStatOption)
		}
		if cr.Spec.ForProvider.ProtocolDetails.TLSSessionResumptionMode != nil {
			f8.SetTlsSessionResumptionMode(*cr.Spec.ForProvider.ProtocolDetails.TLSSessionResumptionMode)
		}
		res.SetProtocolDetails(f8)
	}
	if cr.Spec.ForProvider.Protocols != nil {
		f9 := []*string{}
		for _, f9iter := range cr.Spec.ForProvider.Protocols {
			var f9elem string = *f9iter
			f9 = append(f9, &f9elem)
		}
		res.SetProtocols(f9)
	}
	if cr.Spec.ForProvider.SecurityPolicyName != nil {
		res.SetSecurityPolicyName(*cr.Spec.ForProvider.SecurityPolicyName)
	}
	if cr.Status.AtProvider.ServerID != nil {
		res.SetServerId(*cr.Status.AtProvider.ServerID)
	}
	if cr.Spec.ForProvider.StructuredLogDestinations != nil {
		f12 := []*string{}
		for _, f12iter := range cr.Spec.ForProvider.StructuredLogDestinations {
			var f12elem string = *f12iter
			f12 = append(f12, &f12elem)
		}
		res.SetStructuredLogDestinations(f12)
	}
	if cr.Spec.ForProvider.WorkflowDetails != nil {
		f13 := &svcsdk.WorkflowDetails{}
		if cr.Spec.ForProvider.WorkflowDetails.OnPartialUpload != nil {
			f13f0 := []*svcsdk.WorkflowDetail{}
			for _, f13f0iter := range cr.Spec.ForProvider.WorkflowDetails.OnPartialUpload {
				f13f0elem := &svcsdk.WorkflowDetail{}
				if f13f0iter.ExecutionRole != nil {
					f13f0elem.SetExecutionRole(*f13f0iter.ExecutionRole)
				}
				if f13f0iter.WorkflowID != nil {
					f13f0elem.SetWorkflowId(*f13f0iter.WorkflowID)
				}
				f13f0 = append(f13f0, f13f0elem)
			}
			f13.SetOnPartialUpload(f13f0)
		}
		if cr.Spec.ForProvider.WorkflowDetails.OnUpload != nil {
			f13f1 := []*svcsdk.WorkflowDetail{}
			for _, f13f1iter := range cr.Spec.ForProvider.WorkflowDetails.OnUpload {
				f13f1elem := &svcsdk.WorkflowDetail{}
				if f13f1iter.ExecutionRole != nil {
					f13f1elem.SetExecutionRole(*f13f1iter.ExecutionRole)
				}
				if f13f1iter.WorkflowID != nil {
					f13f1elem.SetWorkflowId(*f13f1iter.WorkflowID)
				}
				f13f1 = append(f13f1, f13f1elem)
			}
			f13.SetOnUpload(f13f1)
		}
		res.SetWorkflowDetails(f13)
	}

	return res
}
