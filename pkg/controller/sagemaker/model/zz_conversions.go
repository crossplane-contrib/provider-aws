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

package model

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeModelInput returns input for read
// operation.
func GenerateDescribeModelInput(cr *svcapitypes.Model) *svcsdk.DescribeModelInput {
	res := preGenerateDescribeModelInput(cr, &svcsdk.DescribeModelInput{})

	if cr.Spec.ForProvider.ModelName != nil {
		res.SetModelName(*cr.Spec.ForProvider.ModelName)
	}

	return postGenerateDescribeModelInput(cr, res)
}

// GenerateModel returns the current state in the form of *svcapitypes.Model.
func GenerateModel(resp *svcsdk.DescribeModelOutput) *svcapitypes.Model {
	cr := &svcapitypes.Model{}

	if resp.ModelArn != nil {
		cr.Status.AtProvider.ModelARN = resp.ModelArn
	}

	return cr
}

// GenerateCreateModelInput returns a create input.
func GenerateCreateModelInput(cr *svcapitypes.Model) *svcsdk.CreateModelInput {
	res := preGenerateCreateModelInput(cr, &svcsdk.CreateModelInput{})

	if cr.Spec.ForProvider.Containers != nil {
		f0 := []*svcsdk.ContainerDefinition{}
		for _, f0iter := range cr.Spec.ForProvider.Containers {
			f0elem := &svcsdk.ContainerDefinition{}
			if f0iter.ContainerHostname != nil {
				f0elem.SetContainerHostname(*f0iter.ContainerHostname)
			}
			if f0iter.Environment != nil {
				f0elemf1 := map[string]*string{}
				for f0elemf1key, f0elemf1valiter := range f0iter.Environment {
					var f0elemf1val string
					f0elemf1val = *f0elemf1valiter
					f0elemf1[f0elemf1key] = &f0elemf1val
				}
				f0elem.SetEnvironment(f0elemf1)
			}
			if f0iter.Image != nil {
				f0elem.SetImage(*f0iter.Image)
			}
			if f0iter.ImageConfig != nil {
				f0elemf3 := &svcsdk.ImageConfig{}
				if f0iter.ImageConfig.RepositoryAccessMode != nil {
					f0elemf3.SetRepositoryAccessMode(*f0iter.ImageConfig.RepositoryAccessMode)
				}
				f0elem.SetImageConfig(f0elemf3)
			}
			if f0iter.Mode != nil {
				f0elem.SetMode(*f0iter.Mode)
			}
			if f0iter.ModelDataURL != nil {
				f0elem.SetModelDataUrl(*f0iter.ModelDataURL)
			}
			if f0iter.ModelPackageName != nil {
				f0elem.SetModelPackageName(*f0iter.ModelPackageName)
			}
			f0 = append(f0, f0elem)
		}
		res.SetContainers(f0)
	}
	if cr.Spec.ForProvider.EnableNetworkIsolation != nil {
		res.SetEnableNetworkIsolation(*cr.Spec.ForProvider.EnableNetworkIsolation)
	}
	if cr.Spec.ForProvider.ExecutionRoleARN != nil {
		res.SetExecutionRoleArn(*cr.Spec.ForProvider.ExecutionRoleARN)
	}
	if cr.Spec.ForProvider.ModelName != nil {
		res.SetModelName(*cr.Spec.ForProvider.ModelName)
	}
	if cr.Spec.ForProvider.PrimaryContainer != nil {
		f4 := &svcsdk.ContainerDefinition{}
		if cr.Spec.ForProvider.PrimaryContainer.ContainerHostname != nil {
			f4.SetContainerHostname(*cr.Spec.ForProvider.PrimaryContainer.ContainerHostname)
		}
		if cr.Spec.ForProvider.PrimaryContainer.Environment != nil {
			f4f1 := map[string]*string{}
			for f4f1key, f4f1valiter := range cr.Spec.ForProvider.PrimaryContainer.Environment {
				var f4f1val string
				f4f1val = *f4f1valiter
				f4f1[f4f1key] = &f4f1val
			}
			f4.SetEnvironment(f4f1)
		}
		if cr.Spec.ForProvider.PrimaryContainer.Image != nil {
			f4.SetImage(*cr.Spec.ForProvider.PrimaryContainer.Image)
		}
		if cr.Spec.ForProvider.PrimaryContainer.ImageConfig != nil {
			f4f3 := &svcsdk.ImageConfig{}
			if cr.Spec.ForProvider.PrimaryContainer.ImageConfig.RepositoryAccessMode != nil {
				f4f3.SetRepositoryAccessMode(*cr.Spec.ForProvider.PrimaryContainer.ImageConfig.RepositoryAccessMode)
			}
			f4.SetImageConfig(f4f3)
		}
		if cr.Spec.ForProvider.PrimaryContainer.Mode != nil {
			f4.SetMode(*cr.Spec.ForProvider.PrimaryContainer.Mode)
		}
		if cr.Spec.ForProvider.PrimaryContainer.ModelDataURL != nil {
			f4.SetModelDataUrl(*cr.Spec.ForProvider.PrimaryContainer.ModelDataURL)
		}
		if cr.Spec.ForProvider.PrimaryContainer.ModelPackageName != nil {
			f4.SetModelPackageName(*cr.Spec.ForProvider.PrimaryContainer.ModelPackageName)
		}
		res.SetPrimaryContainer(f4)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f5 := []*svcsdk.Tag{}
		for _, f5iter := range cr.Spec.ForProvider.Tags {
			f5elem := &svcsdk.Tag{}
			if f5iter.Key != nil {
				f5elem.SetKey(*f5iter.Key)
			}
			if f5iter.Value != nil {
				f5elem.SetValue(*f5iter.Value)
			}
			f5 = append(f5, f5elem)
		}
		res.SetTags(f5)
	}
	if cr.Spec.ForProvider.VPCConfig != nil {
		f6 := &svcsdk.VpcConfig{}
		if cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs != nil {
			f6f0 := []*string{}
			for _, f6f0iter := range cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs {
				var f6f0elem string
				f6f0elem = *f6f0iter
				f6f0 = append(f6f0, &f6f0elem)
			}
			f6.SetSecurityGroupIds(f6f0)
		}
		if cr.Spec.ForProvider.VPCConfig.Subnets != nil {
			f6f1 := []*string{}
			for _, f6f1iter := range cr.Spec.ForProvider.VPCConfig.Subnets {
				var f6f1elem string
				f6f1elem = *f6f1iter
				f6f1 = append(f6f1, &f6f1elem)
			}
			f6.SetSubnets(f6f1)
		}
		res.SetVpcConfig(f6)
	}

	return postGenerateCreateModelInput(cr, res)
}

// GenerateDeleteModelInput returns a deletion input.
func GenerateDeleteModelInput(cr *svcapitypes.Model) *svcsdk.DeleteModelInput {
	res := preGenerateDeleteModelInput(cr, &svcsdk.DeleteModelInput{})

	if cr.Spec.ForProvider.ModelName != nil {
		res.SetModelName(*cr.Spec.ForProvider.ModelName)
	}

	return postGenerateDeleteModelInput(cr, res)
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "ValidationException" && strings.HasPrefix(awsErr.Message(), "Could not find model")
}
