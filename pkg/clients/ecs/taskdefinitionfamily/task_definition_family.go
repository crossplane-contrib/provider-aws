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
	"github.com/aws/aws-sdk-go/aws"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	ecs "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func LateInitialize(in *ecs.TaskDefinitionFamilyParameters, resp *awsecs.DescribeTaskDefinitionOutput) { //nolint:gocyclo
	if in != nil && resp != nil && resp.TaskDefinition != nil {
		if len(in.ContainerDefinitions) == len(resp.TaskDefinition.ContainerDefinitions) {
			for cdi, cd := range in.ContainerDefinitions {
				awscd := resp.TaskDefinition.ContainerDefinitions[cdi]

				cd.CPU = pointer.LateInitialize(cd.CPU, awscd.Cpu)

				if len(cd.PortMappings) == len(awscd.PortMappings) {
					for pmi, pm := range cd.PortMappings {
						pmcd := awscd.PortMappings[pmi]

						pm.HostPort = pointer.LateInitialize(pm.HostPort, pmcd.HostPort)
						pm.Protocol = pointer.LateInitialize(pm.Protocol, pmcd.Protocol)
					}
				}
			}
		}
		if in.Volumes != nil {
			if len(in.Volumes) == len(resp.TaskDefinition.Volumes) {
				for voli, vol := range in.Volumes {
					awsvol := resp.TaskDefinition.Volumes[voli]
					if vol.Host == nil && awsvol.Host != nil {
						vol.Host = &ecs.HostVolumeProperties{
							SourcePath: awsvol.Host.SourcePath,
						}
					}
				}
			}
		}
	}
}

// Modified version of ACK generated conversion function
//
//nolint:gocyclo,gosimple
func GenerateTaskDefinitionFamilyFromDescribe(resp *awsecs.DescribeTaskDefinitionOutput) *ecs.TaskDefinitionFamily {
	cr := &ecs.TaskDefinitionFamily{}
	parameters := &cr.Spec.ForProvider

	if resp.Tags != nil {
		f0 := []*ecs.Tag{}
		for _, f0iter := range resp.Tags {
			f0elem := &ecs.Tag{}
			if f0iter.Key != nil {
				f0elem.Key = f0iter.Key
			}
			if f0iter.Value != nil {
				f0elem.Value = f0iter.Value
			}
			f0 = append(f0, f0elem)
		}
		cr.Spec.ForProvider.Tags = f0
	} else {
		cr.Spec.ForProvider.Tags = nil
	}
	if resp.TaskDefinition != nil {
		if resp.TaskDefinition.ContainerDefinitions != nil {
			f1f1 := []*ecs.ContainerDefinition{}
			for _, f1f1iter := range resp.TaskDefinition.ContainerDefinitions {
				f1f1elem := &ecs.ContainerDefinition{}
				if f1f1iter.Command != nil {
					f1f1elemf0 := []*string{}
					for _, f1f1elemf0iter := range f1f1iter.Command {
						var f1f1elemf0elem string
						f1f1elemf0elem = *f1f1elemf0iter
						f1f1elemf0 = append(f1f1elemf0, &f1f1elemf0elem)
					}
					f1f1elem.Command = f1f1elemf0
				}
				if f1f1iter.Cpu != nil {
					f1f1elem.CPU = f1f1iter.Cpu
				}
				if f1f1iter.CredentialSpecs != nil {
					f1f1elemf2 := []*string{}
					for _, f1f1elemf2iter := range f1f1iter.CredentialSpecs {
						var f1f1elemf2elem string
						f1f1elemf2elem = *f1f1elemf2iter
						f1f1elemf2 = append(f1f1elemf2, &f1f1elemf2elem)
					}
					f1f1elem.CredentialSpecs = f1f1elemf2
				}
				if f1f1iter.DependsOn != nil {
					f1f1elemf3 := []*ecs.ContainerDependency{}
					for _, f1f1elemf3iter := range f1f1iter.DependsOn {
						f1f1elemf3elem := &ecs.ContainerDependency{}
						if f1f1elemf3iter.Condition != nil {
							f1f1elemf3elem.Condition = f1f1elemf3iter.Condition
						}
						if f1f1elemf3iter.ContainerName != nil {
							f1f1elemf3elem.ContainerName = f1f1elemf3iter.ContainerName
						}
						f1f1elemf3 = append(f1f1elemf3, f1f1elemf3elem)
					}
					f1f1elem.DependsOn = f1f1elemf3
				}
				if f1f1iter.DisableNetworking != nil {
					f1f1elem.DisableNetworking = f1f1iter.DisableNetworking
				}
				if f1f1iter.DnsSearchDomains != nil {
					f1f1elemf5 := []*string{}
					for _, f1f1elemf5iter := range f1f1iter.DnsSearchDomains {
						var f1f1elemf5elem string
						f1f1elemf5elem = *f1f1elemf5iter
						f1f1elemf5 = append(f1f1elemf5, &f1f1elemf5elem)
					}
					f1f1elem.DNSSearchDomains = f1f1elemf5
				}
				if f1f1iter.DnsServers != nil {
					f1f1elemf6 := []*string{}
					for _, f1f1elemf6iter := range f1f1iter.DnsServers {
						var f1f1elemf6elem string
						f1f1elemf6elem = *f1f1elemf6iter
						f1f1elemf6 = append(f1f1elemf6, &f1f1elemf6elem)
					}
					f1f1elem.DNSServers = f1f1elemf6
				}
				if f1f1iter.DockerLabels != nil {
					f1f1elemf7 := map[string]*string{}
					for f1f1elemf7key, f1f1elemf7valiter := range f1f1iter.DockerLabels {
						var f1f1elemf7val string
						f1f1elemf7val = *f1f1elemf7valiter
						f1f1elemf7[f1f1elemf7key] = &f1f1elemf7val
					}
					f1f1elem.DockerLabels = f1f1elemf7
				}
				if f1f1iter.DockerSecurityOptions != nil {
					f1f1elemf8 := []*string{}
					for _, f1f1elemf8iter := range f1f1iter.DockerSecurityOptions {
						var f1f1elemf8elem string
						f1f1elemf8elem = *f1f1elemf8iter
						f1f1elemf8 = append(f1f1elemf8, &f1f1elemf8elem)
					}
					f1f1elem.DockerSecurityOptions = f1f1elemf8
				}
				if f1f1iter.EntryPoint != nil {
					f1f1elemf9 := []*string{}
					for _, f1f1elemf9iter := range f1f1iter.EntryPoint {
						var f1f1elemf9elem string
						f1f1elemf9elem = *f1f1elemf9iter
						f1f1elemf9 = append(f1f1elemf9, &f1f1elemf9elem)
					}
					f1f1elem.EntryPoint = f1f1elemf9
				}
				if f1f1iter.Environment != nil {
					f1f1elemf10 := []*ecs.KeyValuePair{}
					for _, f1f1elemf10iter := range f1f1iter.Environment {
						f1f1elemf10elem := &ecs.KeyValuePair{}
						if f1f1elemf10iter.Name != nil {
							f1f1elemf10elem.Name = f1f1elemf10iter.Name
						}
						if f1f1elemf10iter.Value != nil {
							f1f1elemf10elem.Value = f1f1elemf10iter.Value
						}
						f1f1elemf10 = append(f1f1elemf10, f1f1elemf10elem)
					}
					f1f1elem.Environment = f1f1elemf10
				}
				if f1f1iter.EnvironmentFiles != nil {
					f1f1elemf11 := []*ecs.EnvironmentFile{}
					for _, f1f1elemf11iter := range f1f1iter.EnvironmentFiles {
						f1f1elemf11elem := &ecs.EnvironmentFile{}
						if f1f1elemf11iter.Type != nil {
							f1f1elemf11elem.Type = f1f1elemf11iter.Type
						}
						if f1f1elemf11iter.Value != nil {
							f1f1elemf11elem.Value = f1f1elemf11iter.Value
						}
						f1f1elemf11 = append(f1f1elemf11, f1f1elemf11elem)
					}
					f1f1elem.EnvironmentFiles = f1f1elemf11
				}
				if f1f1iter.Essential != nil {
					f1f1elem.Essential = f1f1iter.Essential
				}
				if f1f1iter.ExtraHosts != nil {
					f1f1elemf13 := []*ecs.HostEntry{}
					for _, f1f1elemf13iter := range f1f1iter.ExtraHosts {
						f1f1elemf13elem := &ecs.HostEntry{}
						if f1f1elemf13iter.Hostname != nil {
							f1f1elemf13elem.Hostname = f1f1elemf13iter.Hostname
						}
						if f1f1elemf13iter.IpAddress != nil {
							f1f1elemf13elem.IPAddress = f1f1elemf13iter.IpAddress
						}
						f1f1elemf13 = append(f1f1elemf13, f1f1elemf13elem)
					}
					f1f1elem.ExtraHosts = f1f1elemf13
				}
				if f1f1iter.FirelensConfiguration != nil {
					f1f1elemf14 := &ecs.FirelensConfiguration{}
					if f1f1iter.FirelensConfiguration.Options != nil {
						f1f1elemf14f0 := map[string]*string{}
						for f1f1elemf14f0key, f1f1elemf14f0valiter := range f1f1iter.FirelensConfiguration.Options {
							var f1f1elemf14f0val string
							f1f1elemf14f0val = *f1f1elemf14f0valiter
							f1f1elemf14f0[f1f1elemf14f0key] = &f1f1elemf14f0val
						}
						f1f1elemf14.Options = f1f1elemf14f0
					}
					if f1f1iter.FirelensConfiguration.Type != nil {
						f1f1elemf14.Type = f1f1iter.FirelensConfiguration.Type
					}
					f1f1elem.FirelensConfiguration = f1f1elemf14
				}
				if f1f1iter.HealthCheck != nil {
					f1f1elemf15 := &ecs.HealthCheck{}
					if f1f1iter.HealthCheck.Command != nil {
						f1f1elemf15f0 := []*string{}
						for _, f1f1elemf15f0iter := range f1f1iter.HealthCheck.Command {
							var f1f1elemf15f0elem string
							f1f1elemf15f0elem = *f1f1elemf15f0iter
							f1f1elemf15f0 = append(f1f1elemf15f0, &f1f1elemf15f0elem)
						}
						f1f1elemf15.Command = f1f1elemf15f0
					}
					if f1f1iter.HealthCheck.Interval != nil {
						f1f1elemf15.Interval = f1f1iter.HealthCheck.Interval
					}
					if f1f1iter.HealthCheck.Retries != nil {
						f1f1elemf15.Retries = f1f1iter.HealthCheck.Retries
					}
					if f1f1iter.HealthCheck.StartPeriod != nil {
						f1f1elemf15.StartPeriod = f1f1iter.HealthCheck.StartPeriod
					}
					if f1f1iter.HealthCheck.Timeout != nil {
						f1f1elemf15.Timeout = f1f1iter.HealthCheck.Timeout
					}
					f1f1elem.HealthCheck = f1f1elemf15
				}
				if f1f1iter.Hostname != nil {
					f1f1elem.Hostname = f1f1iter.Hostname
				}
				if f1f1iter.Image != nil {
					f1f1elem.Image = f1f1iter.Image
				}
				if f1f1iter.Interactive != nil {
					f1f1elem.Interactive = f1f1iter.Interactive
				}
				if f1f1iter.Links != nil {
					f1f1elemf19 := []*string{}
					for _, f1f1elemf19iter := range f1f1iter.Links {
						var f1f1elemf19elem string
						f1f1elemf19elem = *f1f1elemf19iter
						f1f1elemf19 = append(f1f1elemf19, &f1f1elemf19elem)
					}
					f1f1elem.Links = f1f1elemf19
				}
				if f1f1iter.LinuxParameters != nil {
					f1f1elemf20 := &ecs.LinuxParameters{}
					if f1f1iter.LinuxParameters.Capabilities != nil {
						f1f1elemf20f0 := &ecs.KernelCapabilities{}
						if f1f1iter.LinuxParameters.Capabilities.Add != nil {
							f1f1elemf20f0f0 := []*string{}
							for _, f1f1elemf20f0f0iter := range f1f1iter.LinuxParameters.Capabilities.Add {
								var f1f1elemf20f0f0elem string
								f1f1elemf20f0f0elem = *f1f1elemf20f0f0iter
								f1f1elemf20f0f0 = append(f1f1elemf20f0f0, &f1f1elemf20f0f0elem)
							}
							f1f1elemf20f0.Add = f1f1elemf20f0f0
						}
						if f1f1iter.LinuxParameters.Capabilities.Drop != nil {
							f1f1elemf20f0f1 := []*string{}
							for _, f1f1elemf20f0f1iter := range f1f1iter.LinuxParameters.Capabilities.Drop {
								var f1f1elemf20f0f1elem string
								f1f1elemf20f0f1elem = *f1f1elemf20f0f1iter
								f1f1elemf20f0f1 = append(f1f1elemf20f0f1, &f1f1elemf20f0f1elem)
							}
							f1f1elemf20f0.Drop = f1f1elemf20f0f1
						}
						f1f1elemf20.Capabilities = f1f1elemf20f0
					}
					if f1f1iter.LinuxParameters.Devices != nil {
						f1f1elemf20f1 := []*ecs.Device{}
						for _, f1f1elemf20f1iter := range f1f1iter.LinuxParameters.Devices {
							f1f1elemf20f1elem := &ecs.Device{}
							if f1f1elemf20f1iter.ContainerPath != nil {
								f1f1elemf20f1elem.ContainerPath = f1f1elemf20f1iter.ContainerPath
							}
							if f1f1elemf20f1iter.HostPath != nil {
								f1f1elemf20f1elem.HostPath = f1f1elemf20f1iter.HostPath
							}
							if f1f1elemf20f1iter.Permissions != nil {
								f1f1elemf20f1elemf2 := []*string{}
								for _, f1f1elemf20f1elemf2iter := range f1f1elemf20f1iter.Permissions {
									var f1f1elemf20f1elemf2elem string
									f1f1elemf20f1elemf2elem = *f1f1elemf20f1elemf2iter
									f1f1elemf20f1elemf2 = append(f1f1elemf20f1elemf2, &f1f1elemf20f1elemf2elem)
								}
								f1f1elemf20f1elem.Permissions = f1f1elemf20f1elemf2
							}
							f1f1elemf20f1 = append(f1f1elemf20f1, f1f1elemf20f1elem)
						}
						f1f1elemf20.Devices = f1f1elemf20f1
					}
					if f1f1iter.LinuxParameters.InitProcessEnabled != nil {
						f1f1elemf20.InitProcessEnabled = f1f1iter.LinuxParameters.InitProcessEnabled
					}
					if f1f1iter.LinuxParameters.MaxSwap != nil {
						f1f1elemf20.MaxSwap = f1f1iter.LinuxParameters.MaxSwap
					}
					if f1f1iter.LinuxParameters.SharedMemorySize != nil {
						f1f1elemf20.SharedMemorySize = f1f1iter.LinuxParameters.SharedMemorySize
					}
					if f1f1iter.LinuxParameters.Swappiness != nil {
						f1f1elemf20.Swappiness = f1f1iter.LinuxParameters.Swappiness
					}
					if f1f1iter.LinuxParameters.Tmpfs != nil {
						f1f1elemf20f6 := []*ecs.Tmpfs{}
						for _, f1f1elemf20f6iter := range f1f1iter.LinuxParameters.Tmpfs {
							f1f1elemf20f6elem := &ecs.Tmpfs{}
							if f1f1elemf20f6iter.ContainerPath != nil {
								f1f1elemf20f6elem.ContainerPath = f1f1elemf20f6iter.ContainerPath
							}
							if f1f1elemf20f6iter.MountOptions != nil {
								f1f1elemf20f6elemf1 := []*string{}
								for _, f1f1elemf20f6elemf1iter := range f1f1elemf20f6iter.MountOptions {
									var f1f1elemf20f6elemf1elem string
									f1f1elemf20f6elemf1elem = *f1f1elemf20f6elemf1iter
									f1f1elemf20f6elemf1 = append(f1f1elemf20f6elemf1, &f1f1elemf20f6elemf1elem)
								}
								f1f1elemf20f6elem.MountOptions = f1f1elemf20f6elemf1
							}
							if f1f1elemf20f6iter.Size != nil {
								f1f1elemf20f6elem.Size = f1f1elemf20f6iter.Size
							}
							f1f1elemf20f6 = append(f1f1elemf20f6, f1f1elemf20f6elem)
						}
						f1f1elemf20.Tmpfs = f1f1elemf20f6
					}
					f1f1elem.LinuxParameters = f1f1elemf20
				}
				if f1f1iter.LogConfiguration != nil {
					f1f1elemf21 := &ecs.LogConfiguration{}
					if f1f1iter.LogConfiguration.LogDriver != nil {
						f1f1elemf21.LogDriver = f1f1iter.LogConfiguration.LogDriver
					}
					if f1f1iter.LogConfiguration.Options != nil {
						f1f1elemf21f1 := map[string]*string{}
						for f1f1elemf21f1key, f1f1elemf21f1valiter := range f1f1iter.LogConfiguration.Options {
							var f1f1elemf21f1val string
							f1f1elemf21f1val = *f1f1elemf21f1valiter
							f1f1elemf21f1[f1f1elemf21f1key] = &f1f1elemf21f1val
						}
						f1f1elemf21.Options = f1f1elemf21f1
					}
					if f1f1iter.LogConfiguration.SecretOptions != nil {
						f1f1elemf21f2 := []*ecs.Secret{}
						for _, f1f1elemf21f2iter := range f1f1iter.LogConfiguration.SecretOptions {
							f1f1elemf21f2elem := &ecs.Secret{}
							if f1f1elemf21f2iter.Name != nil {
								f1f1elemf21f2elem.Name = f1f1elemf21f2iter.Name
							}
							if f1f1elemf21f2iter.ValueFrom != nil {
								f1f1elemf21f2elem.ValueFrom = f1f1elemf21f2iter.ValueFrom
							}
							f1f1elemf21f2 = append(f1f1elemf21f2, f1f1elemf21f2elem)
						}
						f1f1elemf21.SecretOptions = f1f1elemf21f2
					}
					f1f1elem.LogConfiguration = f1f1elemf21
				}
				if f1f1iter.Memory != nil {
					f1f1elem.Memory = f1f1iter.Memory
				}
				if f1f1iter.MemoryReservation != nil {
					f1f1elem.MemoryReservation = f1f1iter.MemoryReservation
				}
				if f1f1iter.MountPoints != nil {
					f1f1elemf24 := []*ecs.MountPoint{}
					for _, f1f1elemf24iter := range f1f1iter.MountPoints {
						f1f1elemf24elem := &ecs.MountPoint{}
						if f1f1elemf24iter.ContainerPath != nil {
							f1f1elemf24elem.ContainerPath = f1f1elemf24iter.ContainerPath
						}
						if f1f1elemf24iter.ReadOnly != nil {
							f1f1elemf24elem.ReadOnly = f1f1elemf24iter.ReadOnly
						}
						if f1f1elemf24iter.SourceVolume != nil {
							f1f1elemf24elem.SourceVolume = f1f1elemf24iter.SourceVolume
						}
						f1f1elemf24 = append(f1f1elemf24, f1f1elemf24elem)
					}
					f1f1elem.MountPoints = f1f1elemf24
				}
				if f1f1iter.Name != nil {
					f1f1elem.Name = f1f1iter.Name
				}
				if f1f1iter.PortMappings != nil {
					f1f1elemf26 := []*ecs.PortMapping{}
					for _, f1f1elemf26iter := range f1f1iter.PortMappings {
						f1f1elemf26elem := &ecs.PortMapping{}
						if f1f1elemf26iter.AppProtocol != nil {
							f1f1elemf26elem.AppProtocol = f1f1elemf26iter.AppProtocol
						}
						if f1f1elemf26iter.ContainerPort != nil {
							f1f1elemf26elem.ContainerPort = f1f1elemf26iter.ContainerPort
						}
						if f1f1elemf26iter.ContainerPortRange != nil {
							f1f1elemf26elem.ContainerPortRange = f1f1elemf26iter.ContainerPortRange
						}
						if f1f1elemf26iter.HostPort != nil {
							f1f1elemf26elem.HostPort = f1f1elemf26iter.HostPort
						}
						if f1f1elemf26iter.Name != nil {
							f1f1elemf26elem.Name = f1f1elemf26iter.Name
						}
						if f1f1elemf26iter.Protocol != nil {
							f1f1elemf26elem.Protocol = f1f1elemf26iter.Protocol
						}
						f1f1elemf26 = append(f1f1elemf26, f1f1elemf26elem)
					}
					f1f1elem.PortMappings = f1f1elemf26
				}
				if f1f1iter.Privileged != nil {
					f1f1elem.Privileged = f1f1iter.Privileged
				}
				if f1f1iter.PseudoTerminal != nil {
					f1f1elem.PseudoTerminal = f1f1iter.PseudoTerminal
				}
				if f1f1iter.ReadonlyRootFilesystem != nil {
					f1f1elem.ReadonlyRootFilesystem = f1f1iter.ReadonlyRootFilesystem
				}
				if f1f1iter.RepositoryCredentials != nil {
					f1f1elemf30 := &ecs.RepositoryCredentials{}
					if f1f1iter.RepositoryCredentials.CredentialsParameter != nil {
						f1f1elemf30.CredentialsParameter = f1f1iter.RepositoryCredentials.CredentialsParameter
					}
					f1f1elem.RepositoryCredentials = f1f1elemf30
				}
				if f1f1iter.ResourceRequirements != nil {
					f1f1elemf31 := []*ecs.ResourceRequirement{}
					for _, f1f1elemf31iter := range f1f1iter.ResourceRequirements {
						f1f1elemf31elem := &ecs.ResourceRequirement{}
						if f1f1elemf31iter.Type != nil {
							f1f1elemf31elem.Type = f1f1elemf31iter.Type
						}
						if f1f1elemf31iter.Value != nil {
							f1f1elemf31elem.Value = f1f1elemf31iter.Value
						}
						f1f1elemf31 = append(f1f1elemf31, f1f1elemf31elem)
					}
					f1f1elem.ResourceRequirements = f1f1elemf31
				}
				if f1f1iter.Secrets != nil {
					f1f1elemf32 := []*ecs.Secret{}
					for _, f1f1elemf32iter := range f1f1iter.Secrets {
						f1f1elemf32elem := &ecs.Secret{}
						if f1f1elemf32iter.Name != nil {
							f1f1elemf32elem.Name = f1f1elemf32iter.Name
						}
						if f1f1elemf32iter.ValueFrom != nil {
							f1f1elemf32elem.ValueFrom = f1f1elemf32iter.ValueFrom
						}
						f1f1elemf32 = append(f1f1elemf32, f1f1elemf32elem)
					}
					f1f1elem.Secrets = f1f1elemf32
				}
				if f1f1iter.StartTimeout != nil {
					f1f1elem.StartTimeout = f1f1iter.StartTimeout
				}
				if f1f1iter.StopTimeout != nil {
					f1f1elem.StopTimeout = f1f1iter.StopTimeout
				}
				if f1f1iter.SystemControls != nil {
					f1f1elemf35 := []*ecs.SystemControl{}
					for _, f1f1elemf35iter := range f1f1iter.SystemControls {
						f1f1elemf35elem := &ecs.SystemControl{}
						if f1f1elemf35iter.Namespace != nil {
							f1f1elemf35elem.Namespace = f1f1elemf35iter.Namespace
						}
						if f1f1elemf35iter.Value != nil {
							f1f1elemf35elem.Value = f1f1elemf35iter.Value
						}
						f1f1elemf35 = append(f1f1elemf35, f1f1elemf35elem)
					}
					f1f1elem.SystemControls = f1f1elemf35
				}
				if f1f1iter.Ulimits != nil {
					f1f1elemf36 := []*ecs.Ulimit{}
					for _, f1f1elemf36iter := range f1f1iter.Ulimits {
						f1f1elemf36elem := &ecs.Ulimit{}
						if f1f1elemf36iter.HardLimit != nil {
							f1f1elemf36elem.HardLimit = f1f1elemf36iter.HardLimit
						}
						if f1f1elemf36iter.Name != nil {
							f1f1elemf36elem.Name = f1f1elemf36iter.Name
						}
						if f1f1elemf36iter.SoftLimit != nil {
							f1f1elemf36elem.SoftLimit = f1f1elemf36iter.SoftLimit
						}
						f1f1elemf36 = append(f1f1elemf36, f1f1elemf36elem)
					}
					f1f1elem.Ulimits = f1f1elemf36
				}
				if f1f1iter.User != nil {
					f1f1elem.User = f1f1iter.User
				}
				if f1f1iter.VolumesFrom != nil {
					f1f1elemf38 := []*ecs.VolumeFrom{}
					for _, f1f1elemf38iter := range f1f1iter.VolumesFrom {
						f1f1elemf38elem := &ecs.VolumeFrom{}
						if f1f1elemf38iter.ReadOnly != nil {
							f1f1elemf38elem.ReadOnly = f1f1elemf38iter.ReadOnly
						}
						if f1f1elemf38iter.SourceContainer != nil {
							f1f1elemf38elem.SourceContainer = f1f1elemf38iter.SourceContainer
						}
						f1f1elemf38 = append(f1f1elemf38, f1f1elemf38elem)
					}
					f1f1elem.VolumesFrom = f1f1elemf38
				}
				if f1f1iter.WorkingDirectory != nil {
					f1f1elem.WorkingDirectory = f1f1iter.WorkingDirectory
				}
				f1f1 = append(f1f1, f1f1elem)
			}
			parameters.ContainerDefinitions = f1f1
		}
		if resp.TaskDefinition.Cpu != nil {
			parameters.CPU = resp.TaskDefinition.Cpu
		}
		if resp.TaskDefinition.EphemeralStorage != nil {
			f1f4 := &ecs.EphemeralStorage{}
			if resp.TaskDefinition.EphemeralStorage.SizeInGiB != nil {
				f1f4.SizeInGiB = resp.TaskDefinition.EphemeralStorage.SizeInGiB
			}
			parameters.EphemeralStorage = f1f4
		}
		if resp.TaskDefinition.ExecutionRoleArn != nil {
			parameters.ExecutionRoleARN = resp.TaskDefinition.ExecutionRoleArn
		}
		if resp.TaskDefinition.Family != nil {
			parameters.Family = resp.TaskDefinition.Family
		}
		if resp.TaskDefinition.InferenceAccelerators != nil {
			f1f7 := []*ecs.InferenceAccelerator{}
			for _, f1f7iter := range resp.TaskDefinition.InferenceAccelerators {
				f1f7elem := &ecs.InferenceAccelerator{}
				if f1f7iter.DeviceName != nil {
					f1f7elem.DeviceName = f1f7iter.DeviceName
				}
				if f1f7iter.DeviceType != nil {
					f1f7elem.DeviceType = f1f7iter.DeviceType
				}
				f1f7 = append(f1f7, f1f7elem)
			}
			parameters.InferenceAccelerators = f1f7
		}
		if resp.TaskDefinition.IpcMode != nil {
			parameters.IPCMode = resp.TaskDefinition.IpcMode
		}
		if resp.TaskDefinition.Memory != nil {
			parameters.Memory = resp.TaskDefinition.Memory
		}
		if resp.TaskDefinition.NetworkMode != nil {
			parameters.NetworkMode = resp.TaskDefinition.NetworkMode
		}
		if resp.TaskDefinition.PidMode != nil {
			parameters.PIDMode = resp.TaskDefinition.PidMode
		}
		if resp.TaskDefinition.PlacementConstraints != nil {
			f1f12 := []*ecs.TaskDefinitionPlacementConstraint{}
			for _, f1f12iter := range resp.TaskDefinition.PlacementConstraints {
				f1f12elem := &ecs.TaskDefinitionPlacementConstraint{}
				if f1f12iter.Expression != nil {
					f1f12elem.Expression = f1f12iter.Expression
				}
				if f1f12iter.Type != nil {
					f1f12elem.Type = f1f12iter.Type
				}
				f1f12 = append(f1f12, f1f12elem)
			}
			parameters.PlacementConstraints = f1f12
		}
		if resp.TaskDefinition.ProxyConfiguration != nil {
			f1f13 := &ecs.ProxyConfiguration{}
			if resp.TaskDefinition.ProxyConfiguration.ContainerName != nil {
				f1f13.ContainerName = resp.TaskDefinition.ProxyConfiguration.ContainerName
			}
			if resp.TaskDefinition.ProxyConfiguration.Properties != nil {
				f1f13f1 := []*ecs.KeyValuePair{}
				for _, f1f13f1iter := range resp.TaskDefinition.ProxyConfiguration.Properties {
					f1f13f1elem := &ecs.KeyValuePair{}
					if f1f13f1iter.Name != nil {
						f1f13f1elem.Name = f1f13f1iter.Name
					}
					if f1f13f1iter.Value != nil {
						f1f13f1elem.Value = f1f13f1iter.Value
					}
					f1f13f1 = append(f1f13f1, f1f13f1elem)
				}
				f1f13.Properties = f1f13f1
			}
			if resp.TaskDefinition.ProxyConfiguration.Type != nil {
				f1f13.Type = resp.TaskDefinition.ProxyConfiguration.Type
			}
			parameters.ProxyConfiguration = f1f13
		}
		if resp.TaskDefinition.RequiresCompatibilities != nil {
			f1f17 := []*string{}
			for _, f1f17iter := range resp.TaskDefinition.RequiresCompatibilities {
				var f1f17elem string
				f1f17elem = *f1f17iter
				f1f17 = append(f1f17, &f1f17elem)
			}
			parameters.RequiresCompatibilities = f1f17
		}
		if resp.TaskDefinition.RuntimePlatform != nil {
			f1f19 := &ecs.RuntimePlatform{}
			if resp.TaskDefinition.RuntimePlatform.CpuArchitecture != nil {
				f1f19.CPUArchitecture = resp.TaskDefinition.RuntimePlatform.CpuArchitecture
			}
			if resp.TaskDefinition.RuntimePlatform.OperatingSystemFamily != nil {
				f1f19.OperatingSystemFamily = resp.TaskDefinition.RuntimePlatform.OperatingSystemFamily
			}
			parameters.RuntimePlatform = f1f19
		}
		if resp.TaskDefinition.TaskRoleArn != nil {
			parameters.TaskRoleARN = resp.TaskDefinition.TaskRoleArn
		}
		if resp.TaskDefinition.Volumes != nil {
			f1f23 := []*ecs.CustomVolume{}
			for _, f1f23iter := range resp.TaskDefinition.Volumes {
				f1f23elem := &ecs.CustomVolume{}
				if f1f23iter.DockerVolumeConfiguration != nil {
					f1f23elemf0 := &ecs.DockerVolumeConfiguration{}
					if f1f23iter.DockerVolumeConfiguration.Autoprovision != nil {
						f1f23elemf0.Autoprovision = f1f23iter.DockerVolumeConfiguration.Autoprovision
					}
					if f1f23iter.DockerVolumeConfiguration.Driver != nil {
						f1f23elemf0.Driver = f1f23iter.DockerVolumeConfiguration.Driver
					}
					if f1f23iter.DockerVolumeConfiguration.DriverOpts != nil {
						f1f23elemf0f2 := map[string]*string{}
						for f1f23elemf0f2key, f1f23elemf0f2valiter := range f1f23iter.DockerVolumeConfiguration.DriverOpts {
							var f1f23elemf0f2val string
							f1f23elemf0f2val = *f1f23elemf0f2valiter
							f1f23elemf0f2[f1f23elemf0f2key] = &f1f23elemf0f2val
						}
						f1f23elemf0.DriverOpts = f1f23elemf0f2
					}
					if f1f23iter.DockerVolumeConfiguration.Labels != nil {
						f1f23elemf0f3 := map[string]*string{}
						for f1f23elemf0f3key, f1f23elemf0f3valiter := range f1f23iter.DockerVolumeConfiguration.Labels {
							var f1f23elemf0f3val string
							f1f23elemf0f3val = *f1f23elemf0f3valiter
							f1f23elemf0f3[f1f23elemf0f3key] = &f1f23elemf0f3val
						}
						f1f23elemf0.Labels = f1f23elemf0f3
					}
					if f1f23iter.DockerVolumeConfiguration.Scope != nil {
						f1f23elemf0.Scope = f1f23iter.DockerVolumeConfiguration.Scope
					}
					f1f23elem.DockerVolumeConfiguration = f1f23elemf0
				}
				if f1f23iter.EfsVolumeConfiguration != nil {
					f1f23elemf1 := &ecs.CustomEFSVolumeConfiguration{}
					if f1f23iter.EfsVolumeConfiguration.AuthorizationConfig != nil {
						f1f23elemf1f0 := &ecs.CustomEFSAuthorizationConfig{}
						if f1f23iter.EfsVolumeConfiguration.AuthorizationConfig.AccessPointId != nil {
							f1f23elemf1f0.AccessPointID = f1f23iter.EfsVolumeConfiguration.AuthorizationConfig.AccessPointId
						}
						if f1f23iter.EfsVolumeConfiguration.AuthorizationConfig.Iam != nil {
							f1f23elemf1f0.IAM = f1f23iter.EfsVolumeConfiguration.AuthorizationConfig.Iam
						}
						f1f23elemf1.AuthorizationConfig = f1f23elemf1f0
					}
					if f1f23iter.EfsVolumeConfiguration.FileSystemId != nil {
						f1f23elemf1.FileSystemID = f1f23iter.EfsVolumeConfiguration.FileSystemId
					}
					if f1f23iter.EfsVolumeConfiguration.RootDirectory != nil {
						f1f23elemf1.RootDirectory = f1f23iter.EfsVolumeConfiguration.RootDirectory
					}
					if f1f23iter.EfsVolumeConfiguration.TransitEncryption != nil {
						f1f23elemf1.TransitEncryption = f1f23iter.EfsVolumeConfiguration.TransitEncryption
					}
					if f1f23iter.EfsVolumeConfiguration.TransitEncryptionPort != nil {
						f1f23elemf1.TransitEncryptionPort = f1f23iter.EfsVolumeConfiguration.TransitEncryptionPort
					}
					f1f23elem.EFSVolumeConfiguration = f1f23elemf1
				}
				if f1f23iter.FsxWindowsFileServerVolumeConfiguration != nil {
					f1f23elemf2 := &ecs.FSxWindowsFileServerVolumeConfiguration{}
					if f1f23iter.FsxWindowsFileServerVolumeConfiguration.AuthorizationConfig != nil {
						f1f23elemf2f0 := &ecs.FSxWindowsFileServerAuthorizationConfig{}
						if f1f23iter.FsxWindowsFileServerVolumeConfiguration.AuthorizationConfig.CredentialsParameter != nil {
							f1f23elemf2f0.CredentialsParameter = f1f23iter.FsxWindowsFileServerVolumeConfiguration.AuthorizationConfig.CredentialsParameter
						}
						if f1f23iter.FsxWindowsFileServerVolumeConfiguration.AuthorizationConfig.Domain != nil {
							f1f23elemf2f0.Domain = f1f23iter.FsxWindowsFileServerVolumeConfiguration.AuthorizationConfig.Domain
						}
						f1f23elemf2.AuthorizationConfig = f1f23elemf2f0
					}
					if f1f23iter.FsxWindowsFileServerVolumeConfiguration.FileSystemId != nil {
						f1f23elemf2.FileSystemID = f1f23iter.FsxWindowsFileServerVolumeConfiguration.FileSystemId
					}
					if f1f23iter.FsxWindowsFileServerVolumeConfiguration.RootDirectory != nil {
						f1f23elemf2.RootDirectory = f1f23iter.FsxWindowsFileServerVolumeConfiguration.RootDirectory
					}
					f1f23elem.FsxWindowsFileServerVolumeConfiguration = f1f23elemf2
				}
				if f1f23iter.Host != nil {
					f1f23elemf3 := &ecs.HostVolumeProperties{}
					if f1f23iter.Host.SourcePath != nil {
						f1f23elemf3.SourcePath = f1f23iter.Host.SourcePath
					}
					f1f23elem.Host = f1f23elemf3
				}
				if f1f23iter.Name != nil {
					f1f23elem.Name = f1f23iter.Name
				}
				f1f23 = append(f1f23, f1f23elem)
			}

			parameters.Volumes = f1f23
		}
	}

	return cr
}

func IsUpToDate(target *ecs.TaskDefinitionFamily, out *awsecs.DescribeTaskDefinitionOutput) (bool, string) {
	t := target.Spec.ForProvider.DeepCopy()
	c := GenerateTaskDefinitionFamilyFromDescribe(out).Spec.ForProvider.DeepCopy()

	tags := func(a, b *ecs.Tag) bool { return aws.StringValue(a.Key) < aws.StringValue(b.Key) }
	stringpointer := func(a, b *string) bool { return aws.StringValue(a) < aws.StringValue(b) }
	keyValuePair := func(a, b *ecs.KeyValuePair) bool { return aws.StringValue(a.Name) < aws.StringValue(b.Name) }
	secret := func(a, b *ecs.Secret) bool { return aws.StringValue(a.Name) < aws.StringValue(b.Name) }

	diff := cmp.Diff(c, t,
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(tags),
		cmpopts.SortSlices(stringpointer),
		cmpopts.SortSlices(keyValuePair),
		cmpopts.SortSlices(secret),
		// Not present in DescribeTaskDefinitionOutput
		cmpopts.IgnoreFields(ecs.TaskDefinitionFamilyParameters{}, "Region"),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}))

	return diff == "", diff
}
