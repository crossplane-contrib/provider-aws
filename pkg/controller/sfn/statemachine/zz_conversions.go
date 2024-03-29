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

// Code generated by ack-generate. DO NOT EDIT.

package statemachine

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/sfn"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/sfn/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeStateMachineInput returns input for read
// operation.
func GenerateDescribeStateMachineInput(cr *svcapitypes.StateMachine) *svcsdk.DescribeStateMachineInput {
	res := &svcsdk.DescribeStateMachineInput{}

	if cr.Status.AtProvider.StateMachineARN != nil {
		res.SetStateMachineArn(*cr.Status.AtProvider.StateMachineARN)
	}

	return res
}

// GenerateStateMachine returns the current state in the form of *svcapitypes.StateMachine.
func GenerateStateMachine(resp *svcsdk.DescribeStateMachineOutput) *svcapitypes.StateMachine {
	cr := &svcapitypes.StateMachine{}

	if resp.CreationDate != nil {
		cr.Status.AtProvider.CreationDate = &metav1.Time{*resp.CreationDate}
	} else {
		cr.Status.AtProvider.CreationDate = nil
	}
	if resp.Definition != nil {
		cr.Spec.ForProvider.Definition = resp.Definition
	} else {
		cr.Spec.ForProvider.Definition = nil
	}
	if resp.LoggingConfiguration != nil {
		f4 := &svcapitypes.LoggingConfiguration{}
		if resp.LoggingConfiguration.Destinations != nil {
			f4f0 := []*svcapitypes.LogDestination{}
			for _, f4f0iter := range resp.LoggingConfiguration.Destinations {
				f4f0elem := &svcapitypes.LogDestination{}
				if f4f0iter.CloudWatchLogsLogGroup != nil {
					f4f0elemf0 := &svcapitypes.CloudWatchLogsLogGroup{}
					if f4f0iter.CloudWatchLogsLogGroup.LogGroupArn != nil {
						f4f0elemf0.LogGroupARN = f4f0iter.CloudWatchLogsLogGroup.LogGroupArn
					}
					f4f0elem.CloudWatchLogsLogGroup = f4f0elemf0
				}
				f4f0 = append(f4f0, f4f0elem)
			}
			f4.Destinations = f4f0
		}
		if resp.LoggingConfiguration.IncludeExecutionData != nil {
			f4.IncludeExecutionData = resp.LoggingConfiguration.IncludeExecutionData
		}
		if resp.LoggingConfiguration.Level != nil {
			f4.Level = resp.LoggingConfiguration.Level
		}
		cr.Spec.ForProvider.LoggingConfiguration = f4
	} else {
		cr.Spec.ForProvider.LoggingConfiguration = nil
	}
	if resp.Name != nil {
		cr.Spec.ForProvider.Name = resp.Name
	} else {
		cr.Spec.ForProvider.Name = nil
	}
	if resp.StateMachineArn != nil {
		cr.Status.AtProvider.StateMachineARN = resp.StateMachineArn
	} else {
		cr.Status.AtProvider.StateMachineARN = nil
	}
	if resp.TracingConfiguration != nil {
		f10 := &svcapitypes.TracingConfiguration{}
		if resp.TracingConfiguration.Enabled != nil {
			f10.Enabled = resp.TracingConfiguration.Enabled
		}
		cr.Spec.ForProvider.TracingConfiguration = f10
	} else {
		cr.Spec.ForProvider.TracingConfiguration = nil
	}

	return cr
}

// GenerateCreateStateMachineInput returns a create input.
func GenerateCreateStateMachineInput(cr *svcapitypes.StateMachine) *svcsdk.CreateStateMachineInput {
	res := &svcsdk.CreateStateMachineInput{}

	if cr.Spec.ForProvider.Definition != nil {
		res.SetDefinition(*cr.Spec.ForProvider.Definition)
	}
	if cr.Spec.ForProvider.LoggingConfiguration != nil {
		f1 := &svcsdk.LoggingConfiguration{}
		if cr.Spec.ForProvider.LoggingConfiguration.Destinations != nil {
			f1f0 := []*svcsdk.LogDestination{}
			for _, f1f0iter := range cr.Spec.ForProvider.LoggingConfiguration.Destinations {
				f1f0elem := &svcsdk.LogDestination{}
				if f1f0iter.CloudWatchLogsLogGroup != nil {
					f1f0elemf0 := &svcsdk.CloudWatchLogsLogGroup{}
					if f1f0iter.CloudWatchLogsLogGroup.LogGroupARN != nil {
						f1f0elemf0.SetLogGroupArn(*f1f0iter.CloudWatchLogsLogGroup.LogGroupARN)
					}
					f1f0elem.SetCloudWatchLogsLogGroup(f1f0elemf0)
				}
				f1f0 = append(f1f0, f1f0elem)
			}
			f1.SetDestinations(f1f0)
		}
		if cr.Spec.ForProvider.LoggingConfiguration.IncludeExecutionData != nil {
			f1.SetIncludeExecutionData(*cr.Spec.ForProvider.LoggingConfiguration.IncludeExecutionData)
		}
		if cr.Spec.ForProvider.LoggingConfiguration.Level != nil {
			f1.SetLevel(*cr.Spec.ForProvider.LoggingConfiguration.Level)
		}
		res.SetLoggingConfiguration(f1)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.Publish != nil {
		res.SetPublish(*cr.Spec.ForProvider.Publish)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f4 := []*svcsdk.Tag{}
		for _, f4iter := range cr.Spec.ForProvider.Tags {
			f4elem := &svcsdk.Tag{}
			if f4iter.Key != nil {
				f4elem.SetKey(*f4iter.Key)
			}
			if f4iter.Value != nil {
				f4elem.SetValue(*f4iter.Value)
			}
			f4 = append(f4, f4elem)
		}
		res.SetTags(f4)
	}
	if cr.Spec.ForProvider.TracingConfiguration != nil {
		f5 := &svcsdk.TracingConfiguration{}
		if cr.Spec.ForProvider.TracingConfiguration.Enabled != nil {
			f5.SetEnabled(*cr.Spec.ForProvider.TracingConfiguration.Enabled)
		}
		res.SetTracingConfiguration(f5)
	}
	if cr.Spec.ForProvider.VersionDescription != nil {
		res.SetVersionDescription(*cr.Spec.ForProvider.VersionDescription)
	}

	return res
}

// GenerateUpdateStateMachineInput returns an update input.
func GenerateUpdateStateMachineInput(cr *svcapitypes.StateMachine) *svcsdk.UpdateStateMachineInput {
	res := &svcsdk.UpdateStateMachineInput{}

	if cr.Spec.ForProvider.Definition != nil {
		res.SetDefinition(*cr.Spec.ForProvider.Definition)
	}
	if cr.Spec.ForProvider.LoggingConfiguration != nil {
		f1 := &svcsdk.LoggingConfiguration{}
		if cr.Spec.ForProvider.LoggingConfiguration.Destinations != nil {
			f1f0 := []*svcsdk.LogDestination{}
			for _, f1f0iter := range cr.Spec.ForProvider.LoggingConfiguration.Destinations {
				f1f0elem := &svcsdk.LogDestination{}
				if f1f0iter.CloudWatchLogsLogGroup != nil {
					f1f0elemf0 := &svcsdk.CloudWatchLogsLogGroup{}
					if f1f0iter.CloudWatchLogsLogGroup.LogGroupARN != nil {
						f1f0elemf0.SetLogGroupArn(*f1f0iter.CloudWatchLogsLogGroup.LogGroupARN)
					}
					f1f0elem.SetCloudWatchLogsLogGroup(f1f0elemf0)
				}
				f1f0 = append(f1f0, f1f0elem)
			}
			f1.SetDestinations(f1f0)
		}
		if cr.Spec.ForProvider.LoggingConfiguration.IncludeExecutionData != nil {
			f1.SetIncludeExecutionData(*cr.Spec.ForProvider.LoggingConfiguration.IncludeExecutionData)
		}
		if cr.Spec.ForProvider.LoggingConfiguration.Level != nil {
			f1.SetLevel(*cr.Spec.ForProvider.LoggingConfiguration.Level)
		}
		res.SetLoggingConfiguration(f1)
	}
	if cr.Spec.ForProvider.Publish != nil {
		res.SetPublish(*cr.Spec.ForProvider.Publish)
	}
	if cr.Status.AtProvider.StateMachineARN != nil {
		res.SetStateMachineArn(*cr.Status.AtProvider.StateMachineARN)
	}
	if cr.Spec.ForProvider.TracingConfiguration != nil {
		f5 := &svcsdk.TracingConfiguration{}
		if cr.Spec.ForProvider.TracingConfiguration.Enabled != nil {
			f5.SetEnabled(*cr.Spec.ForProvider.TracingConfiguration.Enabled)
		}
		res.SetTracingConfiguration(f5)
	}
	if cr.Spec.ForProvider.VersionDescription != nil {
		res.SetVersionDescription(*cr.Spec.ForProvider.VersionDescription)
	}

	return res
}

// GenerateDeleteStateMachineInput returns a deletion input.
func GenerateDeleteStateMachineInput(cr *svcapitypes.StateMachine) *svcsdk.DeleteStateMachineInput {
	res := &svcsdk.DeleteStateMachineInput{}

	if cr.Status.AtProvider.StateMachineARN != nil {
		res.SetStateMachineArn(*cr.Status.AtProvider.StateMachineARN)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "StateMachineDoesNotExist"
}
