package environment

import (
	svcsdk "github.com/aws/aws-sdk-go/service/mwaa"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mwaa/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// generateEnvironment the ACK-generated GenerateEnvironment with custom types.
func generateEnvironment(obj *svcsdk.GetEnvironmentOutput) *svcapitypes.Environment {
	if obj.Environment == nil {
		return &svcapitypes.Environment{}
	}

	res := GenerateEnvironment(obj)

	if obj.Environment.NetworkConfiguration != nil {
		res.Spec.ForProvider.NetworkConfiguration = svcapitypes.CustomNetworkConfiguration{
			SecurityGroupIDs: pointer.SlicePtrToValue(obj.Environment.NetworkConfiguration.SecurityGroupIds),
			SubnetIDs:        pointer.SlicePtrToValue(obj.Environment.NetworkConfiguration.SubnetIds),
		}
	}

	res.Spec.ForProvider.SourceBucketARN = obj.Environment.SourceBucketArn
	res.Spec.ForProvider.ExecutionRoleARN = obj.Environment.ExecutionRoleArn
	res.Spec.ForProvider.KMSKey = obj.Environment.KmsKey
	return res
}
