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

package labelingjob

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeLabelingJobInput returns input for read
// operation.
func GenerateDescribeLabelingJobInput(cr *svcapitypes.LabelingJob) *svcsdk.DescribeLabelingJobInput {
	res := preGenerateDescribeLabelingJobInput(cr, &svcsdk.DescribeLabelingJobInput{})

	if cr.Spec.ForProvider.LabelingJobName != nil {
		res.SetLabelingJobName(*cr.Spec.ForProvider.LabelingJobName)
	}

	return postGenerateDescribeLabelingJobInput(cr, res)
}

// GenerateLabelingJob returns the current state in the form of *svcapitypes.LabelingJob.
func GenerateLabelingJob(resp *svcsdk.DescribeLabelingJobOutput) *svcapitypes.LabelingJob {
	cr := &svcapitypes.LabelingJob{}

	if resp.LabelingJobArn != nil {
		cr.Status.AtProvider.LabelingJobARN = resp.LabelingJobArn
	}

	return cr
}

// GenerateCreateLabelingJobInput returns a create input.
func GenerateCreateLabelingJobInput(cr *svcapitypes.LabelingJob) *svcsdk.CreateLabelingJobInput {
	res := preGenerateCreateLabelingJobInput(cr, &svcsdk.CreateLabelingJobInput{})

	if cr.Spec.ForProvider.HumanTaskConfig != nil {
		f0 := &svcsdk.HumanTaskConfig{}
		if cr.Spec.ForProvider.HumanTaskConfig.AnnotationConsolidationConfig != nil {
			f0f0 := &svcsdk.AnnotationConsolidationConfig{}
			if cr.Spec.ForProvider.HumanTaskConfig.AnnotationConsolidationConfig.AnnotationConsolidationLambdaARN != nil {
				f0f0.SetAnnotationConsolidationLambdaArn(*cr.Spec.ForProvider.HumanTaskConfig.AnnotationConsolidationConfig.AnnotationConsolidationLambdaARN)
			}
			f0.SetAnnotationConsolidationConfig(f0f0)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.MaxConcurrentTaskCount != nil {
			f0.SetMaxConcurrentTaskCount(*cr.Spec.ForProvider.HumanTaskConfig.MaxConcurrentTaskCount)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.NumberOfHumanWorkersPerDataObject != nil {
			f0.SetNumberOfHumanWorkersPerDataObject(*cr.Spec.ForProvider.HumanTaskConfig.NumberOfHumanWorkersPerDataObject)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.PreHumanTaskLambdaARN != nil {
			f0.SetPreHumanTaskLambdaArn(*cr.Spec.ForProvider.HumanTaskConfig.PreHumanTaskLambdaARN)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice != nil {
			f0f4 := &svcsdk.PublicWorkforceTaskPrice{}
			if cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd != nil {
				f0f4f0 := &svcsdk.USD{}
				if cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.Cents != nil {
					f0f4f0.SetCents(*cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.Cents)
				}
				if cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.Dollars != nil {
					f0f4f0.SetDollars(*cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.Dollars)
				}
				if cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.TenthFractionsOfACent != nil {
					f0f4f0.SetTenthFractionsOfACent(*cr.Spec.ForProvider.HumanTaskConfig.PublicWorkforceTaskPrice.AmountInUsd.TenthFractionsOfACent)
				}
				f0f4.SetAmountInUsd(f0f4f0)
			}
			f0.SetPublicWorkforceTaskPrice(f0f4)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.TaskAvailabilityLifetimeInSeconds != nil {
			f0.SetTaskAvailabilityLifetimeInSeconds(*cr.Spec.ForProvider.HumanTaskConfig.TaskAvailabilityLifetimeInSeconds)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.TaskDescription != nil {
			f0.SetTaskDescription(*cr.Spec.ForProvider.HumanTaskConfig.TaskDescription)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.TaskKeywords != nil {
			f0f7 := []*string{}
			for _, f0f7iter := range cr.Spec.ForProvider.HumanTaskConfig.TaskKeywords {
				var f0f7elem string
				f0f7elem = *f0f7iter
				f0f7 = append(f0f7, &f0f7elem)
			}
			f0.SetTaskKeywords(f0f7)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.TaskTimeLimitInSeconds != nil {
			f0.SetTaskTimeLimitInSeconds(*cr.Spec.ForProvider.HumanTaskConfig.TaskTimeLimitInSeconds)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.TaskTitle != nil {
			f0.SetTaskTitle(*cr.Spec.ForProvider.HumanTaskConfig.TaskTitle)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.UiConfig != nil {
			f0f10 := &svcsdk.UiConfig{}
			if cr.Spec.ForProvider.HumanTaskConfig.UiConfig.HumanTaskUiARN != nil {
				f0f10.SetHumanTaskUiArn(*cr.Spec.ForProvider.HumanTaskConfig.UiConfig.HumanTaskUiARN)
			}
			if cr.Spec.ForProvider.HumanTaskConfig.UiConfig.UiTemplateS3URI != nil {
				f0f10.SetUiTemplateS3Uri(*cr.Spec.ForProvider.HumanTaskConfig.UiConfig.UiTemplateS3URI)
			}
			f0.SetUiConfig(f0f10)
		}
		if cr.Spec.ForProvider.HumanTaskConfig.WorkteamARN != nil {
			f0.SetWorkteamArn(*cr.Spec.ForProvider.HumanTaskConfig.WorkteamARN)
		}
		res.SetHumanTaskConfig(f0)
	}
	if cr.Spec.ForProvider.InputConfig != nil {
		f1 := &svcsdk.LabelingJobInputConfig{}
		if cr.Spec.ForProvider.InputConfig.DataAttributes != nil {
			f1f0 := &svcsdk.LabelingJobDataAttributes{}
			if cr.Spec.ForProvider.InputConfig.DataAttributes.ContentClassifiers != nil {
				f1f0f0 := []*string{}
				for _, f1f0f0iter := range cr.Spec.ForProvider.InputConfig.DataAttributes.ContentClassifiers {
					var f1f0f0elem string
					f1f0f0elem = *f1f0f0iter
					f1f0f0 = append(f1f0f0, &f1f0f0elem)
				}
				f1f0.SetContentClassifiers(f1f0f0)
			}
			f1.SetDataAttributes(f1f0)
		}
		if cr.Spec.ForProvider.InputConfig.DataSource != nil {
			f1f1 := &svcsdk.LabelingJobDataSource{}
			if cr.Spec.ForProvider.InputConfig.DataSource.S3DataSource != nil {
				f1f1f0 := &svcsdk.LabelingJobS3DataSource{}
				if cr.Spec.ForProvider.InputConfig.DataSource.S3DataSource.ManifestS3URI != nil {
					f1f1f0.SetManifestS3Uri(*cr.Spec.ForProvider.InputConfig.DataSource.S3DataSource.ManifestS3URI)
				}
				f1f1.SetS3DataSource(f1f1f0)
			}
			if cr.Spec.ForProvider.InputConfig.DataSource.SnsDataSource != nil {
				f1f1f1 := &svcsdk.LabelingJobSnsDataSource{}
				if cr.Spec.ForProvider.InputConfig.DataSource.SnsDataSource.SnsTopicARN != nil {
					f1f1f1.SetSnsTopicArn(*cr.Spec.ForProvider.InputConfig.DataSource.SnsDataSource.SnsTopicARN)
				}
				f1f1.SetSnsDataSource(f1f1f1)
			}
			f1.SetDataSource(f1f1)
		}
		res.SetInputConfig(f1)
	}
	if cr.Spec.ForProvider.LabelAttributeName != nil {
		res.SetLabelAttributeName(*cr.Spec.ForProvider.LabelAttributeName)
	}
	if cr.Spec.ForProvider.LabelCategoryConfigS3URI != nil {
		res.SetLabelCategoryConfigS3Uri(*cr.Spec.ForProvider.LabelCategoryConfigS3URI)
	}
	if cr.Spec.ForProvider.LabelingJobAlgorithmsConfig != nil {
		f4 := &svcsdk.LabelingJobAlgorithmsConfig{}
		if cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.InitialActiveLearningModelARN != nil {
			f4.SetInitialActiveLearningModelArn(*cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.InitialActiveLearningModelARN)
		}
		if cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.LabelingJobAlgorithmSpecificationARN != nil {
			f4.SetLabelingJobAlgorithmSpecificationArn(*cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.LabelingJobAlgorithmSpecificationARN)
		}
		if cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.LabelingJobResourceConfig != nil {
			f4f2 := &svcsdk.LabelingJobResourceConfig{}
			if cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.LabelingJobResourceConfig.VolumeKMSKeyID != nil {
				f4f2.SetVolumeKmsKeyId(*cr.Spec.ForProvider.LabelingJobAlgorithmsConfig.LabelingJobResourceConfig.VolumeKMSKeyID)
			}
			f4.SetLabelingJobResourceConfig(f4f2)
		}
		res.SetLabelingJobAlgorithmsConfig(f4)
	}
	if cr.Spec.ForProvider.LabelingJobName != nil {
		res.SetLabelingJobName(*cr.Spec.ForProvider.LabelingJobName)
	}
	if cr.Spec.ForProvider.OutputConfig != nil {
		f6 := &svcsdk.LabelingJobOutputConfig{}
		if cr.Spec.ForProvider.OutputConfig.KMSKeyID != nil {
			f6.SetKmsKeyId(*cr.Spec.ForProvider.OutputConfig.KMSKeyID)
		}
		if cr.Spec.ForProvider.OutputConfig.S3OutputPath != nil {
			f6.SetS3OutputPath(*cr.Spec.ForProvider.OutputConfig.S3OutputPath)
		}
		if cr.Spec.ForProvider.OutputConfig.SnsTopicARN != nil {
			f6.SetSnsTopicArn(*cr.Spec.ForProvider.OutputConfig.SnsTopicARN)
		}
		res.SetOutputConfig(f6)
	}
	if cr.Spec.ForProvider.RoleARN != nil {
		res.SetRoleArn(*cr.Spec.ForProvider.RoleARN)
	}
	if cr.Spec.ForProvider.StoppingConditions != nil {
		f8 := &svcsdk.LabelingJobStoppingConditions{}
		if cr.Spec.ForProvider.StoppingConditions.MaxHumanLabeledObjectCount != nil {
			f8.SetMaxHumanLabeledObjectCount(*cr.Spec.ForProvider.StoppingConditions.MaxHumanLabeledObjectCount)
		}
		if cr.Spec.ForProvider.StoppingConditions.MaxPercentageOfInputDatasetLabeled != nil {
			f8.SetMaxPercentageOfInputDatasetLabeled(*cr.Spec.ForProvider.StoppingConditions.MaxPercentageOfInputDatasetLabeled)
		}
		res.SetStoppingConditions(f8)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f9 := []*svcsdk.Tag{}
		for _, f9iter := range cr.Spec.ForProvider.Tags {
			f9elem := &svcsdk.Tag{}
			if f9iter.Key != nil {
				f9elem.SetKey(*f9iter.Key)
			}
			if f9iter.Value != nil {
				f9elem.SetValue(*f9iter.Value)
			}
			f9 = append(f9, f9elem)
		}
		res.SetTags(f9)
	}

	return postGenerateCreateLabelingJobInput(cr, res)
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
