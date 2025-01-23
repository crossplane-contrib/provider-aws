package server

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
)

var serverParameters = &svcapitypes.Server{
	Spec: svcapitypes.ServerSpec{
		ForProvider: svcapitypes.ServerParameters{
			CustomServerParameters: svcapitypes.CustomServerParameters{
				Certificate: ptr.To("samecertificate"),
				CustomEndpointDetails: &svcapitypes.CustomEndpointDetails{
					AddressAllocationIDs: []*string{
						ptr.To("id"),
					},
					SecurityGroupIDs: []*string{
						ptr.To("id"),
					},
					SubnetIDs: []*string{
						ptr.To("id"),
					},
					VPCEndpointID: ptr.To("id"),
					VPCID:         ptr.To("id"),
				},
				LoggingRole: ptr.To("role"),
			},
			Domain:       ptr.To("S3"),
			EndpointType: ptr.To("VPC_ENDPOINT"),
			IdentityProviderDetails: &svcapitypes.IdentityProviderDetails{
				DirectoryID:               ptr.To("id"),
				Function:                  ptr.To("function"),
				InvocationRole:            ptr.To("role"),
				SftpAuthenticationMethods: ptr.To("method"),
				URL:                       ptr.To("url"),
			},
			IdentityProviderType:          ptr.To("SERVICE_MANAGED"),
			PostAuthenticationLoginBanner: ptr.To("postbanner"),
			PreAuthenticationLoginBanner:  ptr.To("prebanner"),
			ProtocolDetails: &svcapitypes.ProtocolDetails{
				As2Transports: []*string{
					ptr.To("HTTP"),
				},
				PassiveIP:                ptr.To("127.0.0.1"),
				SetStatOption:            ptr.To("SETSTAT"),
				TLSSessionResumptionMode: ptr.To("ENFORCED"),
			},
			Protocols:          []*string{ptr.To("SFTP")},
			SecurityPolicyName: ptr.To("TransferSecurityPolicy-2024-01"),
			StructuredLogDestinations: []*string{
				ptr.To("arn:log"),
			},
			Tags: []*svcapitypes.Tag{
				{Key: ptr.To("key"), Value: ptr.To("value")},
			},
			WorkflowDetails: &svcapitypes.WorkflowDetails{
				OnPartialUpload: []*svcapitypes.WorkflowDetail{
					{
						WorkflowID:    ptr.To("1"),
						ExecutionRole: ptr.To("role"),
					},
					{
						WorkflowID:    ptr.To("2"),
						ExecutionRole: ptr.To("role2"),
					},
				},
			},
		},
	},
}

var describedServer = &svcsdk.DescribeServerOutput{
	Server: &svcsdk.DescribedServer{
		Certificate: ptr.To("samecertificate"),
		Domain:      ptr.To("S3"),
		EndpointDetails: &svcsdk.EndpointDetails{
			AddressAllocationIds: []*string{
				ptr.To("id"),
			},
			SecurityGroupIds: []*string{
				ptr.To("id"),
			},
			SubnetIds: []*string{
				ptr.To("id"),
			},
			VpcEndpointId: ptr.To("id"),
			VpcId:         ptr.To("id"),
		},
		EndpointType: ptr.To("VPC_ENDPOINT"),
		IdentityProviderDetails: &svcsdk.IdentityProviderDetails{
			DirectoryId:               ptr.To("id"),
			Function:                  ptr.To("function"),
			InvocationRole:            ptr.To("role"),
			SftpAuthenticationMethods: ptr.To("method"),
			Url:                       ptr.To("url"),
		},
		IdentityProviderType:          ptr.To("SERVICE_MANAGED"),
		LoggingRole:                   ptr.To("role"),
		PostAuthenticationLoginBanner: ptr.To("postbanner"),
		PreAuthenticationLoginBanner:  ptr.To("prebanner"),
		ProtocolDetails: &svcsdk.ProtocolDetails{
			As2Transports: []*string{
				ptr.To("HTTP"),
			},
			PassiveIp:                ptr.To("127.0.0.1"),
			SetStatOption:            ptr.To("SETSTAT"),
			TlsSessionResumptionMode: ptr.To("ENFORCED"),
		},
		Protocols:          []*string{ptr.To("SFTP")},
		SecurityPolicyName: ptr.To("TransferSecurityPolicy-2024-01"),
		ServerId:           ptr.To("s-1234567890"),
		State:              ptr.To("STARTING"),
		StructuredLogDestinations: []*string{
			ptr.To("arn:log"),
		},
		Tags: []*svcsdk.Tag{
			{Key: ptr.To("key"), Value: ptr.To("value")},
		},
		UserCount: ptr.To(int64(0)),
		WorkflowDetails: &svcsdk.WorkflowDetails{
			OnPartialUpload: []*svcsdk.WorkflowDetail{
				{
					WorkflowId:    ptr.To("1"),
					ExecutionRole: ptr.To("role"),
				},
				{
					WorkflowId:    ptr.To("2"),
					ExecutionRole: ptr.To("role2"),
				},
			},
		},
	},
}

func TestUpToDateEqual(t *testing.T) {
	in := serverParameters.DeepCopy()
	if upToDate, _, _ := isUpToDate(context.TODO(), in, describedServer); !upToDate {
		t.Error("isUpToDate should be true.")
	}
}

func TestUpToDateWorkflowOrder(t *testing.T) {
	in := serverParameters.DeepCopy()
	in.Spec.ForProvider.WorkflowDetails = &svcapitypes.WorkflowDetails{
		OnPartialUpload: []*svcapitypes.WorkflowDetail{
			{
				WorkflowID:    ptr.To("2"),
				ExecutionRole: ptr.To("role2"),
			},
			{
				WorkflowID:    ptr.To("1"),
				ExecutionRole: ptr.To("role"),
			},
		},
	}
	if upToDate, _, _ := isUpToDate(context.TODO(), in, describedServer); !upToDate {
		t.Error("isUpToDate should be true.")
	}
}

func TestUpToDateWorkflow(t *testing.T) {
	in := serverParameters.DeepCopy()
	in.Spec.ForProvider.WorkflowDetails = &svcapitypes.WorkflowDetails{
		OnPartialUpload: []*svcapitypes.WorkflowDetail{
			{
				WorkflowID:    ptr.To("2"),
				ExecutionRole: ptr.To("role2"),
			},
		},
	}
	if upToDate, _, _ := isUpToDate(context.TODO(), in, describedServer); upToDate {
		t.Error("isUpToDate should be false. Different WorkflowDetails.")
	}
}
