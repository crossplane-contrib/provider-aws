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

package userprofile

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeUserProfileInput returns input for read
// operation.
func GenerateDescribeUserProfileInput(cr *svcapitypes.UserProfile) *svcsdk.DescribeUserProfileInput {
	res := preGenerateDescribeUserProfileInput(cr, &svcsdk.DescribeUserProfileInput{})

	if cr.Spec.ForProvider.DomainID != nil {
		res.SetDomainId(*cr.Spec.ForProvider.DomainID)
	}
	if cr.Spec.ForProvider.UserProfileName != nil {
		res.SetUserProfileName(*cr.Spec.ForProvider.UserProfileName)
	}

	return postGenerateDescribeUserProfileInput(cr, res)
}

// GenerateUserProfile returns the current state in the form of *svcapitypes.UserProfile.
func GenerateUserProfile(resp *svcsdk.DescribeUserProfileOutput) *svcapitypes.UserProfile {
	cr := &svcapitypes.UserProfile{}

	if resp.UserProfileArn != nil {
		cr.Status.AtProvider.UserProfileARN = resp.UserProfileArn
	}

	return cr
}

// GenerateCreateUserProfileInput returns a create input.
func GenerateCreateUserProfileInput(cr *svcapitypes.UserProfile) *svcsdk.CreateUserProfileInput {
	res := preGenerateCreateUserProfileInput(cr, &svcsdk.CreateUserProfileInput{})

	if cr.Spec.ForProvider.DomainID != nil {
		res.SetDomainId(*cr.Spec.ForProvider.DomainID)
	}
	if cr.Spec.ForProvider.SingleSignOnUserIdentifier != nil {
		res.SetSingleSignOnUserIdentifier(*cr.Spec.ForProvider.SingleSignOnUserIdentifier)
	}
	if cr.Spec.ForProvider.SingleSignOnUserValue != nil {
		res.SetSingleSignOnUserValue(*cr.Spec.ForProvider.SingleSignOnUserValue)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f3 := []*svcsdk.Tag{}
		for _, f3iter := range cr.Spec.ForProvider.Tags {
			f3elem := &svcsdk.Tag{}
			if f3iter.Key != nil {
				f3elem.SetKey(*f3iter.Key)
			}
			if f3iter.Value != nil {
				f3elem.SetValue(*f3iter.Value)
			}
			f3 = append(f3, f3elem)
		}
		res.SetTags(f3)
	}
	if cr.Spec.ForProvider.UserProfileName != nil {
		res.SetUserProfileName(*cr.Spec.ForProvider.UserProfileName)
	}
	if cr.Spec.ForProvider.UserSettings != nil {
		f5 := &svcsdk.UserSettings{}
		if cr.Spec.ForProvider.UserSettings.ExecutionRole != nil {
			f5.SetExecutionRole(*cr.Spec.ForProvider.UserSettings.ExecutionRole)
		}
		if cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings != nil {
			f5f1 := &svcsdk.JupyterServerAppSettings{}
			if cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings.DefaultResourceSpec != nil {
				f5f1f0 := &svcsdk.ResourceSpec{}
				if cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings.DefaultResourceSpec.InstanceType != nil {
					f5f1f0.SetInstanceType(*cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings.DefaultResourceSpec.InstanceType)
				}
				if cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings.DefaultResourceSpec.SageMakerImageARN != nil {
					f5f1f0.SetSageMakerImageArn(*cr.Spec.ForProvider.UserSettings.JupyterServerAppSettings.DefaultResourceSpec.SageMakerImageARN)
				}
				f5f1.SetDefaultResourceSpec(f5f1f0)
			}
			f5.SetJupyterServerAppSettings(f5f1)
		}
		if cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings != nil {
			f5f2 := &svcsdk.KernelGatewayAppSettings{}
			if cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings.DefaultResourceSpec != nil {
				f5f2f0 := &svcsdk.ResourceSpec{}
				if cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings.DefaultResourceSpec.InstanceType != nil {
					f5f2f0.SetInstanceType(*cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings.DefaultResourceSpec.InstanceType)
				}
				if cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings.DefaultResourceSpec.SageMakerImageARN != nil {
					f5f2f0.SetSageMakerImageArn(*cr.Spec.ForProvider.UserSettings.KernelGatewayAppSettings.DefaultResourceSpec.SageMakerImageARN)
				}
				f5f2.SetDefaultResourceSpec(f5f2f0)
			}
			f5.SetKernelGatewayAppSettings(f5f2)
		}
		if cr.Spec.ForProvider.UserSettings.SecurityGroups != nil {
			f5f3 := []*string{}
			for _, f5f3iter := range cr.Spec.ForProvider.UserSettings.SecurityGroups {
				var f5f3elem string
				f5f3elem = *f5f3iter
				f5f3 = append(f5f3, &f5f3elem)
			}
			f5.SetSecurityGroups(f5f3)
		}
		if cr.Spec.ForProvider.UserSettings.SharingSettings != nil {
			f5f4 := &svcsdk.SharingSettings{}
			if cr.Spec.ForProvider.UserSettings.SharingSettings.NotebookOutputOption != nil {
				f5f4.SetNotebookOutputOption(*cr.Spec.ForProvider.UserSettings.SharingSettings.NotebookOutputOption)
			}
			if cr.Spec.ForProvider.UserSettings.SharingSettings.S3KMSKeyID != nil {
				f5f4.SetS3KmsKeyId(*cr.Spec.ForProvider.UserSettings.SharingSettings.S3KMSKeyID)
			}
			if cr.Spec.ForProvider.UserSettings.SharingSettings.S3OutputPath != nil {
				f5f4.SetS3OutputPath(*cr.Spec.ForProvider.UserSettings.SharingSettings.S3OutputPath)
			}
			f5.SetSharingSettings(f5f4)
		}
		if cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings != nil {
			f5f5 := &svcsdk.TensorBoardAppSettings{}
			if cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings.DefaultResourceSpec != nil {
				f5f5f0 := &svcsdk.ResourceSpec{}
				if cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings.DefaultResourceSpec.InstanceType != nil {
					f5f5f0.SetInstanceType(*cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings.DefaultResourceSpec.InstanceType)
				}
				if cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings.DefaultResourceSpec.SageMakerImageARN != nil {
					f5f5f0.SetSageMakerImageArn(*cr.Spec.ForProvider.UserSettings.TensorBoardAppSettings.DefaultResourceSpec.SageMakerImageARN)
				}
				f5f5.SetDefaultResourceSpec(f5f5f0)
			}
			f5.SetTensorBoardAppSettings(f5f5)
		}
		res.SetUserSettings(f5)
	}

	return postGenerateCreateUserProfileInput(cr, res)
}

// GenerateDeleteUserProfileInput returns a deletion input.
func GenerateDeleteUserProfileInput(cr *svcapitypes.UserProfile) *svcsdk.DeleteUserProfileInput {
	res := preGenerateDeleteUserProfileInput(cr, &svcsdk.DeleteUserProfileInput{})

	if cr.Spec.ForProvider.DomainID != nil {
		res.SetDomainId(*cr.Spec.ForProvider.DomainID)
	}
	if cr.Spec.ForProvider.UserProfileName != nil {
		res.SetUserProfileName(*cr.Spec.ForProvider.UserProfileName)
	}

	return postGenerateDeleteUserProfileInput(cr, res)
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "UNKNOWN"
}
