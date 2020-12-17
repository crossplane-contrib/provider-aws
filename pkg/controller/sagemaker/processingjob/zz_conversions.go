/*
Copyright 2020 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package processingjob

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeProcessingJobInput returns input for read
// operation.
func GenerateDescribeProcessingJobInput(cr *svcapitypes.ProcessingJob) *svcsdk.DescribeProcessingJobInput {
	res := preGenerateDescribeProcessingJobInput(cr, &svcsdk.DescribeProcessingJobInput{})

	if cr.Spec.ForProvider.ProcessingJobName != nil {
		res.SetProcessingJobName(*cr.Spec.ForProvider.ProcessingJobName)
	}

	return postGenerateDescribeProcessingJobInput(cr, res)
}

// GenerateProcessingJob returns the current state in the form of *svcapitypes.ProcessingJob.
func GenerateProcessingJob(resp *svcsdk.DescribeProcessingJobOutput) *svcapitypes.ProcessingJob {
	cr := &svcapitypes.ProcessingJob{}

	if resp.ProcessingJobArn != nil {
		cr.Status.AtProvider.ProcessingJobARN = resp.ProcessingJobArn
	}

	return cr
}

// GenerateCreateProcessingJobInput returns a create input.
func GenerateCreateProcessingJobInput(cr *svcapitypes.ProcessingJob) *svcsdk.CreateProcessingJobInput {
	res := preGenerateCreateProcessingJobInput(cr, &svcsdk.CreateProcessingJobInput{})

	if cr.Spec.ForProvider.AppSpecification != nil {
		f0 := &svcsdk.AppSpecification{}
		if cr.Spec.ForProvider.AppSpecification.ContainerArguments != nil {
			f0f0 := []*string{}
			for _, f0f0iter := range cr.Spec.ForProvider.AppSpecification.ContainerArguments {
				var f0f0elem string
				f0f0elem = *f0f0iter
				f0f0 = append(f0f0, &f0f0elem)
			}
			f0.SetContainerArguments(f0f0)
		}
		if cr.Spec.ForProvider.AppSpecification.ContainerEntrypoint != nil {
			f0f1 := []*string{}
			for _, f0f1iter := range cr.Spec.ForProvider.AppSpecification.ContainerEntrypoint {
				var f0f1elem string
				f0f1elem = *f0f1iter
				f0f1 = append(f0f1, &f0f1elem)
			}
			f0.SetContainerEntrypoint(f0f1)
		}
		if cr.Spec.ForProvider.AppSpecification.ImageURI != nil {
			f0.SetImageUri(*cr.Spec.ForProvider.AppSpecification.ImageURI)
		}
		res.SetAppSpecification(f0)
	}
	if cr.Spec.ForProvider.Environment != nil {
		f1 := map[string]*string{}
		for f1key, f1valiter := range cr.Spec.ForProvider.Environment {
			var f1val string
			f1val = *f1valiter
			f1[f1key] = &f1val
		}
		res.SetEnvironment(f1)
	}
	if cr.Spec.ForProvider.ExperimentConfig != nil {
		f2 := &svcsdk.ExperimentConfig{}
		if cr.Spec.ForProvider.ExperimentConfig.ExperimentName != nil {
			f2.SetExperimentName(*cr.Spec.ForProvider.ExperimentConfig.ExperimentName)
		}
		if cr.Spec.ForProvider.ExperimentConfig.TrialComponentDisplayName != nil {
			f2.SetTrialComponentDisplayName(*cr.Spec.ForProvider.ExperimentConfig.TrialComponentDisplayName)
		}
		if cr.Spec.ForProvider.ExperimentConfig.TrialName != nil {
			f2.SetTrialName(*cr.Spec.ForProvider.ExperimentConfig.TrialName)
		}
		res.SetExperimentConfig(f2)
	}
	if cr.Spec.ForProvider.NetworkConfig != nil {
		f3 := &svcsdk.NetworkConfig{}
		if cr.Spec.ForProvider.NetworkConfig.EnableInterContainerTrafficEncryption != nil {
			f3.SetEnableInterContainerTrafficEncryption(*cr.Spec.ForProvider.NetworkConfig.EnableInterContainerTrafficEncryption)
		}
		if cr.Spec.ForProvider.NetworkConfig.EnableNetworkIsolation != nil {
			f3.SetEnableNetworkIsolation(*cr.Spec.ForProvider.NetworkConfig.EnableNetworkIsolation)
		}
		if cr.Spec.ForProvider.NetworkConfig.VPCConfig != nil {
			f3f2 := &svcsdk.VpcConfig{}
			if cr.Spec.ForProvider.NetworkConfig.VPCConfig.SecurityGroupIDs != nil {
				f3f2f0 := []*string{}
				for _, f3f2f0iter := range cr.Spec.ForProvider.NetworkConfig.VPCConfig.SecurityGroupIDs {
					var f3f2f0elem string
					f3f2f0elem = *f3f2f0iter
					f3f2f0 = append(f3f2f0, &f3f2f0elem)
				}
				f3f2.SetSecurityGroupIds(f3f2f0)
			}
			if cr.Spec.ForProvider.NetworkConfig.VPCConfig.Subnets != nil {
				f3f2f1 := []*string{}
				for _, f3f2f1iter := range cr.Spec.ForProvider.NetworkConfig.VPCConfig.Subnets {
					var f3f2f1elem string
					f3f2f1elem = *f3f2f1iter
					f3f2f1 = append(f3f2f1, &f3f2f1elem)
				}
				f3f2.SetSubnets(f3f2f1)
			}
			f3.SetVpcConfig(f3f2)
		}
		res.SetNetworkConfig(f3)
	}
	if cr.Spec.ForProvider.ProcessingInputs != nil {
		f4 := []*svcsdk.ProcessingInput{}
		for _, f4iter := range cr.Spec.ForProvider.ProcessingInputs {
			f4elem := &svcsdk.ProcessingInput{}
			if f4iter.InputName != nil {
				f4elem.SetInputName(*f4iter.InputName)
			}
			if f4iter.S3Input != nil {
				f4elemf1 := &svcsdk.ProcessingS3Input{}
				if f4iter.S3Input.LocalPath != nil {
					f4elemf1.SetLocalPath(*f4iter.S3Input.LocalPath)
				}
				if f4iter.S3Input.S3CompressionType != nil {
					f4elemf1.SetS3CompressionType(*f4iter.S3Input.S3CompressionType)
				}
				if f4iter.S3Input.S3DataDistributionType != nil {
					f4elemf1.SetS3DataDistributionType(*f4iter.S3Input.S3DataDistributionType)
				}
				if f4iter.S3Input.S3DataType != nil {
					f4elemf1.SetS3DataType(*f4iter.S3Input.S3DataType)
				}
				if f4iter.S3Input.S3InputMode != nil {
					f4elemf1.SetS3InputMode(*f4iter.S3Input.S3InputMode)
				}
				if f4iter.S3Input.S3URI != nil {
					f4elemf1.SetS3Uri(*f4iter.S3Input.S3URI)
				}
				f4elem.SetS3Input(f4elemf1)
			}
			f4 = append(f4, f4elem)
		}
		res.SetProcessingInputs(f4)
	}
	if cr.Spec.ForProvider.ProcessingJobName != nil {
		res.SetProcessingJobName(*cr.Spec.ForProvider.ProcessingJobName)
	}
	if cr.Spec.ForProvider.ProcessingOutputConfig != nil {
		f6 := &svcsdk.ProcessingOutputConfig{}
		if cr.Spec.ForProvider.ProcessingOutputConfig.KMSKeyID != nil {
			f6.SetKmsKeyId(*cr.Spec.ForProvider.ProcessingOutputConfig.KMSKeyID)
		}
		if cr.Spec.ForProvider.ProcessingOutputConfig.Outputs != nil {
			f6f1 := []*svcsdk.ProcessingOutput{}
			for _, f6f1iter := range cr.Spec.ForProvider.ProcessingOutputConfig.Outputs {
				f6f1elem := &svcsdk.ProcessingOutput{}
				if f6f1iter.OutputName != nil {
					f6f1elem.SetOutputName(*f6f1iter.OutputName)
				}
				if f6f1iter.S3Output != nil {
					f6f1elemf1 := &svcsdk.ProcessingS3Output{}
					if f6f1iter.S3Output.LocalPath != nil {
						f6f1elemf1.SetLocalPath(*f6f1iter.S3Output.LocalPath)
					}
					if f6f1iter.S3Output.S3UploadMode != nil {
						f6f1elemf1.SetS3UploadMode(*f6f1iter.S3Output.S3UploadMode)
					}
					if f6f1iter.S3Output.S3URI != nil {
						f6f1elemf1.SetS3Uri(*f6f1iter.S3Output.S3URI)
					}
					f6f1elem.SetS3Output(f6f1elemf1)
				}
				f6f1 = append(f6f1, f6f1elem)
			}
			f6.SetOutputs(f6f1)
		}
		res.SetProcessingOutputConfig(f6)
	}
	if cr.Spec.ForProvider.ProcessingResources != nil {
		f7 := &svcsdk.ProcessingResources{}
		if cr.Spec.ForProvider.ProcessingResources.ClusterConfig != nil {
			f7f0 := &svcsdk.ProcessingClusterConfig{}
			if cr.Spec.ForProvider.ProcessingResources.ClusterConfig.InstanceCount != nil {
				f7f0.SetInstanceCount(*cr.Spec.ForProvider.ProcessingResources.ClusterConfig.InstanceCount)
			}
			if cr.Spec.ForProvider.ProcessingResources.ClusterConfig.InstanceType != nil {
				f7f0.SetInstanceType(*cr.Spec.ForProvider.ProcessingResources.ClusterConfig.InstanceType)
			}
			if cr.Spec.ForProvider.ProcessingResources.ClusterConfig.VolumeKMSKeyID != nil {
				f7f0.SetVolumeKmsKeyId(*cr.Spec.ForProvider.ProcessingResources.ClusterConfig.VolumeKMSKeyID)
			}
			if cr.Spec.ForProvider.ProcessingResources.ClusterConfig.VolumeSizeInGB != nil {
				f7f0.SetVolumeSizeInGB(*cr.Spec.ForProvider.ProcessingResources.ClusterConfig.VolumeSizeInGB)
			}
			f7.SetClusterConfig(f7f0)
		}
		res.SetProcessingResources(f7)
	}
	if cr.Spec.ForProvider.RoleARN != nil {
		res.SetRoleArn(*cr.Spec.ForProvider.RoleARN)
	}
	if cr.Spec.ForProvider.StoppingCondition != nil {
		f9 := &svcsdk.ProcessingStoppingCondition{}
		if cr.Spec.ForProvider.StoppingCondition.MaxRuntimeInSeconds != nil {
			f9.SetMaxRuntimeInSeconds(*cr.Spec.ForProvider.StoppingCondition.MaxRuntimeInSeconds)
		}
		res.SetStoppingCondition(f9)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f10 := []*svcsdk.Tag{}
		for _, f10iter := range cr.Spec.ForProvider.Tags {
			f10elem := &svcsdk.Tag{}
			if f10iter.Key != nil {
				f10elem.SetKey(*f10iter.Key)
			}
			if f10iter.Value != nil {
				f10elem.SetValue(*f10iter.Value)
			}
			f10 = append(f10, f10elem)
		}
		res.SetTags(f10)
	}

	return postGenerateCreateProcessingJobInput(cr, res)
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
