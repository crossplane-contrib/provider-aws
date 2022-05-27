package environment

import (
	svcsdk "github.com/aws/aws-sdk-go/service/mwaa"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mwaa/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

// generateEnvironment the ACK-generated GenerateEnvironment with custom types.
func generateEnvironment(obj *svcsdk.GetEnvironmentOutput) *svcapitypes.Environment {
	if obj.Environment == nil {
		return &svcapitypes.Environment{}
	}

	res := GenerateEnvironment(obj)

	if obj.Environment.NetworkConfiguration != nil {
		res.Spec.ForProvider.NetworkConfiguration = svcapitypes.CustomNetworkConfiguration{
			SecurityGroupIDs: awsclients.StringPtrSliceToValue(obj.Environment.NetworkConfiguration.SecurityGroupIds),
			SubnetIDs:        awsclients.StringPtrSliceToValue(obj.Environment.NetworkConfiguration.SubnetIds),
		}
	}

	res.Spec.ForProvider.SourceBucketARN = obj.Environment.SourceBucketArn
	res.Spec.ForProvider.ExecutionRoleARN = obj.Environment.ExecutionRoleArn
	res.Spec.ForProvider.KMSKey = obj.Environment.KmsKey
	return res
}
