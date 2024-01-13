/*
Copyright 2022 The Crossplane Authors.

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

package job

import (
	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/batch/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func generateJob(resp *svcsdk.DescribeJobsOutput) *svcapitypes.Job { //nolint:gocyclo
	cr := &svcapitypes.Job{}

	for _, elem := range resp.Jobs {
		if elem.ArrayProperties != nil {
			cr.Status.AtProvider.ArrayProperties = &svcapitypes.ArrayPropertiesDetail{
				Index:         elem.ArrayProperties.Index,
				Size:          elem.ArrayProperties.Size,
				StatusSummary: elem.ArrayProperties.StatusSummary,
			}
		}
		if elem.Attempts != nil {
			cr.Status.AtProvider.Attempts = []*svcapitypes.AttemptDetail{}
			for _, attempt := range elem.Attempts {
				apiAttempt := &svcapitypes.AttemptDetail{
					StartedAt:    attempt.StartedAt,
					StatusReason: attempt.StatusReason,
					StoppedAt:    attempt.StoppedAt,
				}
				if attempt.Container != nil {
					apiAttempt.Container = &svcapitypes.AttemptContainerDetail{
						ContainerInstanceArn: attempt.Container.ContainerInstanceArn,
						ExitCode:             attempt.Container.ExitCode,
						LogStreamName:        attempt.Container.LogStreamName,
						Reason:               attempt.Container.Reason,
						TaskArn:              attempt.Container.TaskArn,
					}
					if attempt.Container.NetworkInterfaces != nil {
						apiAttempt.Container.NetworkInterfaces = []*svcapitypes.NetworkInterface{}
						for _, netint := range attempt.Container.NetworkInterfaces {
							apiAttempt.Container.NetworkInterfaces = append(apiAttempt.Container.NetworkInterfaces,
								&svcapitypes.NetworkInterface{
									AttachmentID:       netint.AttachmentId,
									Ipv6Address:        netint.Ipv6Address,
									PrivateIpv4Address: netint.PrivateIpv4Address,
								})
						}
					}
				}
				cr.Status.AtProvider.Attempts = append(cr.Status.AtProvider.Attempts, apiAttempt)
			}
		}
		if elem.CreatedAt != nil {
			cr.Status.AtProvider.CreatedAt = elem.CreatedAt
		}
		if elem.JobArn != nil {
			cr.Status.AtProvider.JobArn = elem.JobArn
		} else {
			cr.Status.AtProvider.JobArn = nil
		}
		if elem.JobId != nil {
			cr.Status.AtProvider.JobID = elem.JobId
		} else {
			cr.Status.AtProvider.JobID = nil
		}
		if elem.StartedAt != nil {
			cr.Status.AtProvider.StartedAt = elem.StartedAt
		}
		if elem.Status != nil {
			cr.Status.AtProvider.Status = elem.Status
		} else {
			cr.Status.AtProvider.Status = nil
		}
		if elem.StatusReason != nil {
			cr.Status.AtProvider.StatusReason = elem.StatusReason
		}
		if elem.StoppedAt != nil {
			cr.Status.AtProvider.StoppedAt = elem.StoppedAt
		}
		if elem.ArrayProperties != nil {
			cr.Spec.ForProvider.ArrayProperties = &svcapitypes.ArrayProperties{Size: elem.ArrayProperties.Size}
		} else {
			cr.Spec.ForProvider.ArrayProperties = nil
		}
		if elem.Container != nil {
			cr.Spec.ForProvider.ContainerOverrides = getContainerOverridesFromDetail(elem.Container)
		}
		if elem.DependsOn != nil {
			jobDeps := []*svcapitypes.JobDependency{}
			for _, dependencies := range elem.DependsOn {
				jobDeps = append(jobDeps, &svcapitypes.JobDependency{
					JobID: dependencies.JobId,
					Type:  dependencies.Type,
				})
			}
			cr.Spec.ForProvider.DependsOn = jobDeps
		} else {
			cr.Spec.ForProvider.DependsOn = nil
		}
		if elem.JobDefinition != nil {
			cr.Spec.ForProvider.JobDefinition = pointer.StringValue(elem.JobDefinition)
		}
		if elem.JobQueue != nil {
			cr.Spec.ForProvider.JobQueue = pointer.StringValue(elem.JobQueue)
		}
		np := elem.NodeProperties
		if np != nil {
			nodeOvers := &svcapitypes.NodeOverrides{}
			if np.NodeRangeProperties != nil {
				noProOver := []*svcapitypes.NodePropertyOverride{}
				for _, noRaProp := range np.NodeRangeProperties {
					apiNoProOver := &svcapitypes.NodePropertyOverride{}
					if noRaProp.Container != nil {
						apiNoProOver.ContainerOverrides = getContainerOverridesFromProperties(noRaProp.Container)
					}
					apiNoProOver.TargetNodes = pointer.StringValue(noRaProp.TargetNodes)
					noProOver = append(noProOver, apiNoProOver)
				}
				nodeOvers.NodePropertyOverrides = noProOver
			}
			if np.NumNodes != nil {
				nodeOvers.NumNodes = np.NumNodes
			}
			cr.Spec.ForProvider.NodeOverrides = nodeOvers
		}
		if elem.Parameters != nil {
			cr.Spec.ForProvider.Parameters = elem.Parameters
		}
		if elem.PropagateTags != nil {
			cr.Spec.ForProvider.PropagateTags = elem.PropagateTags
		}
		if elem.RetryStrategy != nil {
			retStr := &svcapitypes.RetryStrategy{}
			retStr.Attempts = elem.RetryStrategy.Attempts
			if elem.RetryStrategy.EvaluateOnExit != nil {
				eoes := []*svcapitypes.EvaluateOnExit{}
				for _, eoe := range elem.RetryStrategy.EvaluateOnExit {
					eoes = append(eoes, &svcapitypes.EvaluateOnExit{
						Action:         pointer.StringValue(eoe.Action),
						OnExitCode:     eoe.OnExitCode,
						OnReason:       eoe.OnReason,
						OnStatusReason: eoe.OnStatusReason,
					})
				}
				retStr.EvaluateOnExit = eoes
			}
			cr.Spec.ForProvider.RetryStrategy = retStr
		}
		if elem.Tags != nil {
			cr.Spec.ForProvider.Tags = elem.Tags
		}
		if elem.Timeout != nil {
			cr.Spec.ForProvider.Timeout = &svcapitypes.JobTimeout{AttemptDurationSeconds: elem.Timeout.AttemptDurationSeconds}
		}
	}

	return cr
}

// Helper for generateJob() with filling ContainerOverrides from ContainerDetail
func getContainerOverridesFromDetail(co *svcsdk.ContainerDetail) *svcapitypes.ContainerOverrides {
	specco := &svcapitypes.ContainerOverrides{}
	if co != nil {
		if co.Command != nil {
			specco.Command = co.Command
		}
		if co.Environment != nil {
			env := []*svcapitypes.KeyValuePair{}
			for _, pair := range co.Environment {
				env = append(env, &svcapitypes.KeyValuePair{
					Name:  pair.Name,
					Value: pair.Value,
				})
			}
			specco.Environment = env
		}
		if co.InstanceType != nil {
			specco.InstanceType = co.InstanceType
		}
		if co.ResourceRequirements != nil {
			resReqs := []*svcapitypes.ResourceRequirement{}
			for _, resReq := range co.ResourceRequirements {
				resReqs = append(resReqs, &svcapitypes.ResourceRequirement{
					ResourceType: pointer.StringValue(resReq.Type),
					Value:        pointer.StringValue(resReq.Value),
				})
			}
			specco.ResourceRequirements = resReqs
		}
	}
	return specco
}

// Helper for generateJob() with filling ContainerOverrides from ContainerProperties
func getContainerOverridesFromProperties(co *svcsdk.ContainerProperties) *svcapitypes.ContainerOverrides {
	specco := &svcapitypes.ContainerOverrides{}
	if co != nil {
		if co.Command != nil {
			specco.Command = co.Command
		}
		if co.Environment != nil {
			env := []*svcapitypes.KeyValuePair{}
			for _, pair := range co.Environment {
				env = append(env, &svcapitypes.KeyValuePair{
					Name:  pair.Name,
					Value: pair.Value,
				})
			}
			specco.Environment = env
		}
		if co.InstanceType != nil {
			specco.InstanceType = co.InstanceType
		}
		if co.ResourceRequirements != nil {
			resReqs := []*svcapitypes.ResourceRequirement{}
			for _, resReq := range co.ResourceRequirements {
				resReqs = append(resReqs, &svcapitypes.ResourceRequirement{
					ResourceType: pointer.StringValue(resReq.Type),
					Value:        pointer.StringValue(resReq.Value),
				})
			}
			specco.ResourceRequirements = resReqs
		}
	}
	return specco
}

func generateSubmitJobInput(cr *svcapitypes.Job) *svcsdk.SubmitJobInput { //nolint:gocyclo
	res := &svcsdk.SubmitJobInput{}
	res.JobName = pointer.ToOrNilIfZeroValue(cr.Name)

	if cr.Spec.ForProvider.ArrayProperties != nil {
		res.ArrayProperties = &svcsdk.ArrayProperties{Size: cr.Spec.ForProvider.ArrayProperties.Size}
	}

	if cr.Spec.ForProvider.ContainerOverrides != nil {
		res.ContainerOverrides = assignContainerOverrides(cr.Spec.ForProvider.ContainerOverrides)
	}

	if cr.Spec.ForProvider.DependsOn != nil {
		jobDeps := []*svcsdk.JobDependency{}
		for _, dependencies := range cr.Spec.ForProvider.DependsOn {
			jobDeps = append(jobDeps, &svcsdk.JobDependency{
				JobId: dependencies.JobID,
				Type:  dependencies.Type,
			})
		}
		res.DependsOn = jobDeps
	} else {
		res.DependsOn = nil
	}

	res.JobDefinition = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.JobDefinition)
	res.JobQueue = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.JobQueue)

	np := cr.Spec.ForProvider.NodeOverrides
	if np != nil {
		nodeOvers := &svcsdk.NodeOverrides{}
		if np.NodePropertyOverrides != nil {
			noProOver := []*svcsdk.NodePropertyOverride{}
			for _, noProOvers := range np.NodePropertyOverrides {
				sdkNoProOver := &svcsdk.NodePropertyOverride{}
				if noProOvers.ContainerOverrides != nil {
					sdkNoProOver.ContainerOverrides = assignContainerOverrides(noProOvers.ContainerOverrides)
				}
				sdkNoProOver.TargetNodes = pointer.ToOrNilIfZeroValue(noProOvers.TargetNodes)
				noProOver = append(noProOver, sdkNoProOver)
			}
			nodeOvers.NodePropertyOverrides = noProOver
		}
		if np.NumNodes != nil {
			nodeOvers.NumNodes = np.NumNodes
		}
		res.NodeOverrides = nodeOvers
	}

	if cr.Spec.ForProvider.Parameters != nil {
		res.Parameters = cr.Spec.ForProvider.Parameters
	}

	if cr.Spec.ForProvider.PropagateTags != nil {
		res.PropagateTags = cr.Spec.ForProvider.PropagateTags
	}

	if cr.Spec.ForProvider.RetryStrategy != nil {
		retStr := &svcsdk.RetryStrategy{}
		retStr.Attempts = cr.Spec.ForProvider.RetryStrategy.Attempts
		if cr.Spec.ForProvider.RetryStrategy.EvaluateOnExit != nil {
			eoes := []*svcsdk.EvaluateOnExit{}
			for _, eoe := range cr.Spec.ForProvider.RetryStrategy.EvaluateOnExit {
				eoes = append(eoes, &svcsdk.EvaluateOnExit{
					Action:         pointer.ToOrNilIfZeroValue(eoe.Action),
					OnExitCode:     eoe.OnExitCode,
					OnReason:       eoe.OnReason,
					OnStatusReason: eoe.OnStatusReason,
				})
			}
			retStr.EvaluateOnExit = eoes
		}
		res.RetryStrategy = retStr
	}

	if cr.Spec.ForProvider.Tags != nil {
		res.Tags = cr.Spec.ForProvider.Tags
	}

	if cr.Spec.ForProvider.Timeout != nil {
		res.Timeout = &svcsdk.JobTimeout{AttemptDurationSeconds: cr.Spec.ForProvider.Timeout.AttemptDurationSeconds}
	}

	return res
}

// Helper for generateSubmitJobInput() with filling ContainerOverrides
func assignContainerOverrides(co *svcapitypes.ContainerOverrides) *svcsdk.ContainerOverrides {
	specco := &svcsdk.ContainerOverrides{}
	if co != nil {
		if co.Command != nil {
			specco.Command = co.Command
		}
		if co.Environment != nil {
			env := []*svcsdk.KeyValuePair{}
			for _, pair := range co.Environment {
				env = append(env, &svcsdk.KeyValuePair{
					Name:  pair.Name,
					Value: pair.Value,
				})
			}
			specco.Environment = env
		}
		if co.InstanceType != nil {
			specco.InstanceType = co.InstanceType
		}
		if co.ResourceRequirements != nil {
			resReqs := []*svcsdk.ResourceRequirement{}
			for _, resReq := range co.ResourceRequirements {
				resReqs = append(resReqs, &svcsdk.ResourceRequirement{
					Type:  pointer.ToOrNilIfZeroValue(resReq.ResourceType),
					Value: pointer.ToOrNilIfZeroValue(resReq.Value),
				})
			}
			specco.ResourceRequirements = resReqs
		}
	}
	return specco
}

func generateTerminateJobInput(cr *svcapitypes.Job, msg *string) *svcsdk.TerminateJobInput {
	res := &svcsdk.TerminateJobInput{
		JobId:  pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		Reason: msg,
	}

	return res
}
