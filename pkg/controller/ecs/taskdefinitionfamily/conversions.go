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

package taskdefinitionfamily

import (
	awsecs "github.com/aws/aws-sdk-go/service/ecs"

	ecs "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
)

// Converts a TaskDefinitionFamily into a TaskDefinition to use auto-generated TaskDefinition functions
func GenerateTaskDefinition(f *ecs.TaskDefinitionFamily) *ecs.TaskDefinition {
	return &ecs.TaskDefinition{
		Spec: ecs.TaskDefinitionSpec{
			ForProvider: ecs.TaskDefinitionParameters{
				Region:                         f.Spec.ForProvider.Region,
				ContainerDefinitions:           f.Spec.ForProvider.ContainerDefinitions,
				CPU:                            f.Spec.ForProvider.CPU,
				EphemeralStorage:               f.Spec.ForProvider.EphemeralStorage,
				Family:                         f.Spec.ForProvider.Family,
				InferenceAccelerators:          f.Spec.ForProvider.InferenceAccelerators,
				IPCMode:                        f.Spec.ForProvider.IPCMode,
				Memory:                         f.Spec.ForProvider.Memory,
				NetworkMode:                    f.Spec.ForProvider.NetworkMode,
				PIDMode:                        f.Spec.ForProvider.PIDMode,
				PlacementConstraints:           f.Spec.ForProvider.PlacementConstraints,
				ProxyConfiguration:             f.Spec.ForProvider.ProxyConfiguration,
				RequiresCompatibilities:        f.Spec.ForProvider.RequiresCompatibilities,
				RuntimePlatform:                f.Spec.ForProvider.RuntimePlatform,
				Tags:                           f.Spec.ForProvider.Tags,
				CustomTaskDefinitionParameters: f.Spec.ForProvider.CustomTaskDefinitionParameters,
			},
		},
		Status: ecs.TaskDefinitionStatus{
			AtProvider: ecs.TaskDefinitionObservation{
				TaskDefinition: f.Status.AtProvider.TaskDefinition,
			},
		},
	}
}

// Converts a TaskDefinition into a TaskDefinitionFamily to use auto-generated TaskDefinition functions
func GenerateTaskDefinitionFamily(f *ecs.TaskDefinition) *ecs.TaskDefinitionFamily {
	return &ecs.TaskDefinitionFamily{
		Spec: ecs.TaskDefinitionFamilySpec{
			ForProvider: ecs.TaskDefinitionFamilyParameters{
				Region:                         f.Spec.ForProvider.Region,
				ContainerDefinitions:           f.Spec.ForProvider.ContainerDefinitions,
				CPU:                            f.Spec.ForProvider.CPU,
				EphemeralStorage:               f.Spec.ForProvider.EphemeralStorage,
				Family:                         f.Spec.ForProvider.Family,
				InferenceAccelerators:          f.Spec.ForProvider.InferenceAccelerators,
				IPCMode:                        f.Spec.ForProvider.IPCMode,
				Memory:                         f.Spec.ForProvider.Memory,
				NetworkMode:                    f.Spec.ForProvider.NetworkMode,
				PIDMode:                        f.Spec.ForProvider.PIDMode,
				PlacementConstraints:           f.Spec.ForProvider.PlacementConstraints,
				ProxyConfiguration:             f.Spec.ForProvider.ProxyConfiguration,
				RequiresCompatibilities:        f.Spec.ForProvider.RequiresCompatibilities,
				RuntimePlatform:                f.Spec.ForProvider.RuntimePlatform,
				Tags:                           f.Spec.ForProvider.Tags,
				CustomTaskDefinitionParameters: f.Spec.ForProvider.CustomTaskDefinitionParameters,
			},
		},
		Status: ecs.TaskDefinitionFamilyApiStatus{
			AtProvider: ecs.TaskDefinitionFamilyObservation{
				TaskDefinition: f.Status.AtProvider.TaskDefinition,
			},
		},
	}
}

// Converts RegisterTaskDefinitionOutput into a DescribeTaskDefinitionOutput to use auto-generated convert functions
func GenerateDescribeTaskDefinitionOutput(registerOutput *awsecs.RegisterTaskDefinitionOutput) *awsecs.DescribeTaskDefinitionOutput {
	return &awsecs.DescribeTaskDefinitionOutput{
		Tags:           registerOutput.Tags,
		TaskDefinition: registerOutput.TaskDefinition,
	}
}
