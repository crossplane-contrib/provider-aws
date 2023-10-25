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

package jobdefinition

import (
	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/batch/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func generateJobDefinition(resp *svcsdk.DescribeJobDefinitionsOutput) *svcapitypes.JobDefinition { //nolint:gocyclo
	cr := &svcapitypes.JobDefinition{}

	for _, elem := range resp.JobDefinitions {
		if elem.JobDefinitionArn != nil {
			cr.Status.AtProvider.JobDefinitionArn = elem.JobDefinitionArn
		} else {
			cr.Status.AtProvider.JobDefinitionArn = nil
		}
		if elem.Revision != nil {
			cr.Status.AtProvider.Revision = elem.Revision
		} else {
			cr.Status.AtProvider.Revision = nil
		}
		if elem.Status != nil {
			cr.Status.AtProvider.Status = elem.Status
		} else {
			cr.Status.AtProvider.Status = nil
		}
		if elem.Type != nil {
			cr.Spec.ForProvider.JobDefinitionType = pointer.StringValue(elem.Type)
		}
		if elem.ContainerProperties != nil {
			cr.Spec.ForProvider.ContainerProperties = getContainerProperties(elem.ContainerProperties)
		}
		np := elem.NodeProperties
		if np != nil {
			nodeProps := &svcapitypes.NodeProperties{}
			if np.MainNode != nil {
				nodeProps.MainNode = pointer.Int64Value(np.MainNode)
			}
			if np.NodeRangeProperties != nil {
				noRaProps := []svcapitypes.NodeRangeProperty{}
				for _, noRaProp := range np.NodeRangeProperties {
					apiNoRaProp := svcapitypes.NodeRangeProperty{}
					if noRaProp.Container != nil {
						apiNoRaProp.Container = getContainerProperties(noRaProp.Container)
					}
					apiNoRaProp.TargetNodes = pointer.StringValue(noRaProp.TargetNodes)
					noRaProps = append(noRaProps, apiNoRaProp)
				}
				nodeProps.NodeRangeProperties = noRaProps
			}
			if np.NumNodes != nil {
				nodeProps.NumNodes = pointer.Int64Value(np.NumNodes)
			}
			cr.Spec.ForProvider.NodeProperties = nodeProps
		}
		if elem.Parameters != nil {
			cr.Spec.ForProvider.Parameters = elem.Parameters
		}
		if elem.PlatformCapabilities != nil {
			cr.Spec.ForProvider.PlatformCapabilities = elem.PlatformCapabilities
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

// Helper for generateJobDefinition() with filling ContainerProperties
func getContainerProperties(cp *svcsdk.ContainerProperties) *svcapitypes.ContainerProperties { //nolint:gocyclo
	speccp := &svcapitypes.ContainerProperties{}
	if cp != nil {
		if cp.Command != nil {
			speccp.Command = cp.Command
		}
		if cp.Environment != nil {
			env := []*svcapitypes.KeyValuePair{}
			for _, pair := range cp.Environment {
				env = append(env, &svcapitypes.KeyValuePair{
					Name:  pair.Name,
					Value: pair.Value,
				})
			}
			speccp.Environment = env
		}
		if cp.ExecutionRoleArn != nil {
			speccp.ExecutionRoleArn = cp.ExecutionRoleArn
		}
		if cp.FargatePlatformConfiguration != nil {
			speccp.FargatePlatformConfiguration = &svcapitypes.FargatePlatformConfiguration{PlatformVersion: cp.FargatePlatformConfiguration.PlatformVersion}
		}
		if cp.Image != nil {
			speccp.Image = cp.Image
		}
		if cp.InstanceType != nil {
			speccp.InstanceType = cp.InstanceType
		}
		if cp.JobRoleArn != nil {
			speccp.JobRoleArn = cp.JobRoleArn
		}
		if cp.LinuxParameters != nil {
			lipa := &svcapitypes.LinuxParameters{}
			if cp.LinuxParameters.Devices != nil {
				devices := []*svcapitypes.Device{}
				for _, device := range cp.LinuxParameters.Devices {
					devices = append(devices, &svcapitypes.Device{
						ContainerPath: device.ContainerPath,
						HostPath:      pointer.StringValue(device.HostPath),
						Permissions:   device.Permissions,
					})
				}
				lipa.Devices = devices
			}
			if cp.LinuxParameters.InitProcessEnabled != nil {
				lipa.InitProcessEnabled = cp.LinuxParameters.InitProcessEnabled
			}
			if cp.LinuxParameters.MaxSwap != nil {
				lipa.MaxSwap = cp.LinuxParameters.MaxSwap
			}
			if cp.LinuxParameters.SharedMemorySize != nil {
				lipa.SharedMemorySize = cp.LinuxParameters.SharedMemorySize
			}
			if cp.LinuxParameters.Swappiness != nil {
				lipa.Swappiness = cp.LinuxParameters.Swappiness
			}
			if cp.LinuxParameters.Tmpfs != nil {
				tmpfs := []*svcapitypes.Tmpfs{}
				for _, tmpf := range cp.LinuxParameters.Tmpfs {
					tmpfs = append(tmpfs, &svcapitypes.Tmpfs{
						ContainerPath: pointer.StringValue(tmpf.ContainerPath),
						MountOptions:  tmpf.MountOptions,
						Size:          pointer.Int64Value(tmpf.Size),
					})
				}
				lipa.Tmpfs = tmpfs
			}
			speccp.LinuxParameters = lipa
		}
		if cp.LogConfiguration != nil {
			logConfi := &svcapitypes.LogConfiguration{}
			if cp.LogConfiguration.LogDriver != nil {
				logConfi.LogDriver = pointer.StringValue(cp.LogConfiguration.LogDriver)
			}
			if cp.LogConfiguration.Options != nil {
				logConfi.Options = cp.LogConfiguration.Options
			}
			if cp.LogConfiguration.SecretOptions != nil {
				secrets := []*svcapitypes.Secret{}
				for _, secret := range cp.LogConfiguration.SecretOptions {
					secrets = append(secrets, &svcapitypes.Secret{
						Name:      pointer.StringValue(secret.Name),
						ValueFrom: pointer.StringValue(secret.ValueFrom),
					})
				}
				logConfi.SecretOptions = secrets
			}
			speccp.LogConfiguration = logConfi
		}
		if cp.MountPoints != nil {
			moPos := []*svcapitypes.MountPoint{}
			for _, moPo := range cp.MountPoints {
				moPos = append(moPos, &svcapitypes.MountPoint{
					ContainerPath: moPo.ContainerPath,
					ReadOnly:      moPo.ReadOnly,
					SourceVolume:  moPo.SourceVolume,
				})
			}
			speccp.MountPoints = moPos
		}
		if cp.NetworkConfiguration != nil {
			speccp.NetworkConfiguration = &svcapitypes.NetworkConfiguration{
				AssignPublicIP: cp.NetworkConfiguration.AssignPublicIp}
		}
		if cp.Privileged != nil {
			speccp.Privileged = cp.Privileged
		}
		if cp.ReadonlyRootFilesystem != nil {
			speccp.ReadonlyRootFilesystem = cp.ReadonlyRootFilesystem
		}
		if cp.ResourceRequirements != nil {
			resReqs := []*svcapitypes.ResourceRequirement{}
			for _, resReq := range cp.ResourceRequirements {
				resReqs = append(resReqs, &svcapitypes.ResourceRequirement{
					ResourceType: pointer.StringValue(resReq.Type),
					Value:        pointer.StringValue(resReq.Value),
				})
			}
			speccp.ResourceRequirements = resReqs
		}
		if cp.Secrets != nil {
			secrets := []*svcapitypes.Secret{}
			for _, secret := range cp.Secrets {
				secrets = append(secrets, &svcapitypes.Secret{
					Name:      pointer.StringValue(secret.Name),
					ValueFrom: pointer.StringValue(secret.ValueFrom),
				})
			}
			speccp.Secrets = secrets
		}
		if cp.Ulimits != nil {
			ulimits := []*svcapitypes.Ulimit{}
			for _, ulimit := range cp.Ulimits {
				ulimits = append(ulimits, &svcapitypes.Ulimit{
					HardLimit: pointer.Int64Value(ulimit.HardLimit),
					Name:      pointer.StringValue(ulimit.Name),
					SoftLimit: pointer.Int64Value(ulimit.SoftLimit),
				})
			}
			speccp.Ulimits = ulimits
		}
		if cp.User != nil {
			speccp.User = cp.User
		}
		if cp.Volumes != nil {
			volumes := []*svcapitypes.Volume{}
			for _, volume := range cp.Volumes {
				specVolume := &svcapitypes.Volume{}
				if volume.EfsVolumeConfiguration != nil {
					specVolumeConfig := &svcapitypes.EFSVolumeConfiguration{}
					if volume.EfsVolumeConfiguration.AuthorizationConfig != nil {
						specVolumeConfig.AuthorizationConfig = &svcapitypes.EFSAuthorizationConfig{
							AccessPointID: volume.EfsVolumeConfiguration.AuthorizationConfig.AccessPointId,
							IAM:           volume.EfsVolumeConfiguration.AuthorizationConfig.Iam,
						}
					}
					specVolumeConfig.FileSystemID = pointer.StringValue(volume.EfsVolumeConfiguration.FileSystemId)
					specVolumeConfig.RootDirectory = volume.EfsVolumeConfiguration.RootDirectory
					specVolumeConfig.TransitEncryption = volume.EfsVolumeConfiguration.TransitEncryption
					specVolumeConfig.TransitEncryptionPort = volume.EfsVolumeConfiguration.TransitEncryptionPort
					specVolume.EfsVolumeConfiguration = specVolumeConfig
				}
				if volume.Host != nil {
					specVolume.Host = &svcapitypes.Host{SourcePath: volume.Host.SourcePath}
				}
				specVolume.Name = volume.Name
				volumes = append(volumes, specVolume)
			}
			speccp.Volumes = volumes
		}
	}
	return speccp
}

func generateRegisterJobDefinitionInput(cr *svcapitypes.JobDefinition) *svcsdk.RegisterJobDefinitionInput { //nolint:gocyclo
	res := &svcsdk.RegisterJobDefinitionInput{}
	res.JobDefinitionName = pointer.ToOrNilIfZeroValue(cr.Name)
	res.Type = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.JobDefinitionType)

	if cr.Spec.ForProvider.ContainerProperties != nil {
		res.ContainerProperties = assignContainerProperties(cr.Spec.ForProvider.ContainerProperties)
	}

	np := cr.Spec.ForProvider.NodeProperties
	if np != nil {
		nodeProps := &svcsdk.NodeProperties{}

		nodeProps.MainNode = &np.MainNode

		if np.NodeRangeProperties != nil {
			noRaProps := []*svcsdk.NodeRangeProperty{}
			for _, noRaProp := range np.NodeRangeProperties {
				sdkNoRaProp := &svcsdk.NodeRangeProperty{}
				if noRaProp.Container != nil {
					sdkNoRaProp.Container = assignContainerProperties(noRaProp.Container)
				}
				sdkNoRaProp.TargetNodes = pointer.ToOrNilIfZeroValue(noRaProp.TargetNodes)
				noRaProps = append(noRaProps, sdkNoRaProp)
			}
			nodeProps.NodeRangeProperties = noRaProps
		}

		nodeProps.NumNodes = &np.NumNodes

		res.NodeProperties = nodeProps
	}

	if cr.Spec.ForProvider.Parameters != nil {
		res.Parameters = cr.Spec.ForProvider.Parameters
	}

	if cr.Spec.ForProvider.PlatformCapabilities != nil {
		res.PlatformCapabilities = cr.Spec.ForProvider.PlatformCapabilities
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

// Helper for generateRegisterJobDefinitionInput() with filling ContainerProperties
func assignContainerProperties(cp *svcapitypes.ContainerProperties) *svcsdk.ContainerProperties { //nolint:gocyclo
	sdkcp := &svcsdk.ContainerProperties{}
	if cp != nil {
		if cp.Command != nil {
			sdkcp.SetCommand(cp.Command)
		}
		if cp.Environment != nil {
			env := []*svcsdk.KeyValuePair{}
			for _, pair := range cp.Environment {
				env = append(env, &svcsdk.KeyValuePair{
					Name:  pair.Name,
					Value: pair.Value,
				})
			}
			sdkcp.Environment = env
		}
		if cp.ExecutionRoleArn != nil {
			sdkcp.ExecutionRoleArn = cp.ExecutionRoleArn
		}
		if cp.FargatePlatformConfiguration != nil {
			sdkcp.FargatePlatformConfiguration = &svcsdk.FargatePlatformConfiguration{PlatformVersion: cp.FargatePlatformConfiguration.PlatformVersion}
		}
		if cp.Image != nil {
			sdkcp.Image = cp.Image
		}
		if cp.InstanceType != nil {
			sdkcp.InstanceType = cp.InstanceType
		}
		if cp.JobRoleArn != nil {
			sdkcp.JobRoleArn = cp.JobRoleArn
		}
		if cp.LinuxParameters != nil {
			lipa := &svcsdk.LinuxParameters{}
			if cp.LinuxParameters.Devices != nil {
				devices := []*svcsdk.Device{}
				for _, device := range cp.LinuxParameters.Devices {
					devices = append(devices, &svcsdk.Device{
						ContainerPath: device.ContainerPath,
						HostPath:      pointer.ToOrNilIfZeroValue(device.HostPath),
						Permissions:   device.Permissions,
					})
				}
				lipa.Devices = devices
			}
			if cp.LinuxParameters.InitProcessEnabled != nil {
				lipa.InitProcessEnabled = cp.LinuxParameters.InitProcessEnabled
			}
			if cp.LinuxParameters.MaxSwap != nil {
				lipa.MaxSwap = cp.LinuxParameters.MaxSwap
			}
			if cp.LinuxParameters.SharedMemorySize != nil {
				lipa.SharedMemorySize = cp.LinuxParameters.SharedMemorySize
			}
			if cp.LinuxParameters.Swappiness != nil {
				lipa.Swappiness = cp.LinuxParameters.Swappiness
			}
			if cp.LinuxParameters.Tmpfs != nil {
				tmpfs := []*svcsdk.Tmpfs{}
				for _, tmpf := range cp.LinuxParameters.Tmpfs {
					tmpfs = append(tmpfs, &svcsdk.Tmpfs{
						ContainerPath: pointer.ToOrNilIfZeroValue(tmpf.ContainerPath),
						MountOptions:  tmpf.MountOptions,
						Size:          ptr.To(tmpf.Size),
					})
				}
				lipa.Tmpfs = tmpfs
			}
			sdkcp.LinuxParameters = lipa
		}
		if cp.LogConfiguration != nil {
			logConfi := &svcsdk.LogConfiguration{}

			logConfi.LogDriver = pointer.ToOrNilIfZeroValue(cp.LogConfiguration.LogDriver)

			if cp.LogConfiguration.Options != nil {
				logConfi.Options = cp.LogConfiguration.Options
			}
			if cp.LogConfiguration.SecretOptions != nil {
				secrets := []*svcsdk.Secret{}
				for _, secret := range cp.LogConfiguration.SecretOptions {
					secrets = append(secrets, &svcsdk.Secret{
						Name:      pointer.ToOrNilIfZeroValue(secret.Name),
						ValueFrom: pointer.ToOrNilIfZeroValue(secret.ValueFrom),
					})
				}
				logConfi.SecretOptions = secrets
			}
			sdkcp.LogConfiguration = logConfi
		}
		if cp.MountPoints != nil {
			moPos := []*svcsdk.MountPoint{}
			for _, moPo := range cp.MountPoints {
				moPos = append(moPos, &svcsdk.MountPoint{
					ContainerPath: moPo.ContainerPath,
					ReadOnly:      moPo.ReadOnly,
					SourceVolume:  moPo.SourceVolume,
				})
			}
			sdkcp.MountPoints = moPos
		}
		if cp.NetworkConfiguration != nil {
			sdkcp.NetworkConfiguration = &svcsdk.NetworkConfiguration{
				AssignPublicIp: cp.NetworkConfiguration.AssignPublicIP}
		}
		if cp.Privileged != nil {
			sdkcp.Privileged = cp.Privileged
		}
		if cp.ReadonlyRootFilesystem != nil {
			sdkcp.ReadonlyRootFilesystem = cp.ReadonlyRootFilesystem
		}
		if cp.ResourceRequirements != nil {
			resReqs := []*svcsdk.ResourceRequirement{}
			for _, resReq := range cp.ResourceRequirements {
				resReqs = append(resReqs, &svcsdk.ResourceRequirement{
					Type:  pointer.ToOrNilIfZeroValue(resReq.ResourceType),
					Value: pointer.ToOrNilIfZeroValue(resReq.Value),
				})
			}
			sdkcp.ResourceRequirements = resReqs
		}
		if cp.Secrets != nil {
			secrets := []*svcsdk.Secret{}
			for _, secret := range cp.Secrets {
				secrets = append(secrets, &svcsdk.Secret{
					Name:      pointer.ToOrNilIfZeroValue(secret.Name),
					ValueFrom: pointer.ToOrNilIfZeroValue(secret.ValueFrom),
				})
			}
			sdkcp.Secrets = secrets
		}
		if cp.Ulimits != nil {
			ulimits := []*svcsdk.Ulimit{}
			for _, ulimit := range cp.Ulimits {
				ulimits = append(ulimits, &svcsdk.Ulimit{
					HardLimit: ptr.To(ulimit.HardLimit),
					Name:      pointer.ToOrNilIfZeroValue(ulimit.Name),
					SoftLimit: ptr.To(ulimit.SoftLimit),
				})
			}
			sdkcp.Ulimits = ulimits
		}
		if cp.User != nil {
			sdkcp.User = cp.User
		}
		if cp.Volumes != nil {
			volumes := []*svcsdk.Volume{}
			for _, volume := range cp.Volumes {
				sdkVolume := &svcsdk.Volume{}
				if volume.EfsVolumeConfiguration != nil {
					sdkVolumeConfig := &svcsdk.EFSVolumeConfiguration{}
					if volume.EfsVolumeConfiguration.AuthorizationConfig != nil {
						sdkVolumeConfig.AuthorizationConfig = &svcsdk.EFSAuthorizationConfig{
							AccessPointId: volume.EfsVolumeConfiguration.AuthorizationConfig.AccessPointID,
							Iam:           volume.EfsVolumeConfiguration.AuthorizationConfig.IAM,
						}
					}
					sdkVolumeConfig.FileSystemId = pointer.ToOrNilIfZeroValue(volume.EfsVolumeConfiguration.FileSystemID)
					sdkVolumeConfig.RootDirectory = volume.EfsVolumeConfiguration.RootDirectory
					sdkVolumeConfig.TransitEncryption = volume.EfsVolumeConfiguration.TransitEncryption
					sdkVolumeConfig.TransitEncryptionPort = volume.EfsVolumeConfiguration.TransitEncryptionPort
					sdkVolume.EfsVolumeConfiguration = sdkVolumeConfig
				}
				if volume.Host != nil {
					sdkVolume.Host = &svcsdk.Host{SourcePath: volume.Host.SourcePath}
				}
				sdkVolume.Name = volume.Name
				volumes = append(volumes, sdkVolume)
			}
			sdkcp.Volumes = volumes
		}
	}
	return sdkcp
}

func generateDeregisterJobDefinitionInput(cr *svcapitypes.JobDefinition) *svcsdk.DeregisterJobDefinitionInput {
	res := &svcsdk.DeregisterJobDefinitionInput{
		JobDefinition: cr.Status.AtProvider.JobDefinitionArn,
	}

	return res
}
