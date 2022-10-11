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

package launchtemplate

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeLaunchTemplatesInput returns input for read
// operation.
func GenerateDescribeLaunchTemplatesInput(cr *svcapitypes.LaunchTemplate) *svcsdk.DescribeLaunchTemplatesInput {
	res := &svcsdk.DescribeLaunchTemplatesInput{}

	if cr.Spec.ForProvider.LaunchTemplateName != nil {
		f3 := []*string{}
		f3 = append(f3, cr.Spec.ForProvider.LaunchTemplateName)
		res.SetLaunchTemplateNames(f3)
	}

	return res
}

// GenerateLaunchTemplate returns the current state in the form of *svcapitypes.LaunchTemplate.
func GenerateLaunchTemplate(resp *svcsdk.DescribeLaunchTemplatesOutput) *svcapitypes.LaunchTemplate {
	cr := &svcapitypes.LaunchTemplate{}

	found := false
	for _, elem := range resp.LaunchTemplates {
		if elem.LaunchTemplateName != nil {
			cr.Spec.ForProvider.LaunchTemplateName = elem.LaunchTemplateName
		} else {
			cr.Spec.ForProvider.LaunchTemplateName = nil
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateLaunchTemplateInput returns a create input.
func GenerateCreateLaunchTemplateInput(cr *svcapitypes.LaunchTemplate) *svcsdk.CreateLaunchTemplateInput {
	res := &svcsdk.CreateLaunchTemplateInput{}

	if cr.Spec.ForProvider.LaunchTemplateData != nil {
		f0 := &svcsdk.RequestLaunchTemplateData{}
		if cr.Spec.ForProvider.LaunchTemplateData.BlockDeviceMappings != nil {
			f0f0 := []*svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
			for _, f0f0iter := range cr.Spec.ForProvider.LaunchTemplateData.BlockDeviceMappings {
				f0f0elem := &svcsdk.LaunchTemplateBlockDeviceMappingRequest{}
				if f0f0iter.DeviceName != nil {
					f0f0elem.SetDeviceName(*f0f0iter.DeviceName)
				}
				if f0f0iter.EBS != nil {
					f0f0elemf1 := &svcsdk.LaunchTemplateEbsBlockDeviceRequest{}
					if f0f0iter.EBS.DeleteOnTermination != nil {
						f0f0elemf1.SetDeleteOnTermination(*f0f0iter.EBS.DeleteOnTermination)
					}
					if f0f0iter.EBS.Encrypted != nil {
						f0f0elemf1.SetEncrypted(*f0f0iter.EBS.Encrypted)
					}
					if f0f0iter.EBS.IOPS != nil {
						f0f0elemf1.SetIops(*f0f0iter.EBS.IOPS)
					}
					if f0f0iter.EBS.KMSKeyID != nil {
						f0f0elemf1.SetKmsKeyId(*f0f0iter.EBS.KMSKeyID)
					}
					if f0f0iter.EBS.SnapshotID != nil {
						f0f0elemf1.SetSnapshotId(*f0f0iter.EBS.SnapshotID)
					}
					if f0f0iter.EBS.Throughput != nil {
						f0f0elemf1.SetThroughput(*f0f0iter.EBS.Throughput)
					}
					if f0f0iter.EBS.VolumeSize != nil {
						f0f0elemf1.SetVolumeSize(*f0f0iter.EBS.VolumeSize)
					}
					if f0f0iter.EBS.VolumeType != nil {
						f0f0elemf1.SetVolumeType(*f0f0iter.EBS.VolumeType)
					}
					f0f0elem.SetEbs(f0f0elemf1)
				}
				if f0f0iter.NoDevice != nil {
					f0f0elem.SetNoDevice(*f0f0iter.NoDevice)
				}
				if f0f0iter.VirtualName != nil {
					f0f0elem.SetVirtualName(*f0f0iter.VirtualName)
				}
				f0f0 = append(f0f0, f0f0elem)
			}
			f0.SetBlockDeviceMappings(f0f0)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification != nil {
			f0f1 := &svcsdk.LaunchTemplateCapacityReservationSpecificationRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationPreference != nil {
				f0f1.SetCapacityReservationPreference(*cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationPreference)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget != nil {
				f0f1f1 := &svcsdk.CapacityReservationTarget{}
				if cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget.CapacityReservationID != nil {
					f0f1f1.SetCapacityReservationId(*cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget.CapacityReservationID)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget.CapacityReservationResourceGroupARN != nil {
					f0f1f1.SetCapacityReservationResourceGroupArn(*cr.Spec.ForProvider.LaunchTemplateData.CapacityReservationSpecification.CapacityReservationTarget.CapacityReservationResourceGroupARN)
				}
				f0f1.SetCapacityReservationTarget(f0f1f1)
			}
			f0.SetCapacityReservationSpecification(f0f1)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.CPUOptions != nil {
			f0f2 := &svcsdk.LaunchTemplateCpuOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.CPUOptions.CoreCount != nil {
				f0f2.SetCoreCount(*cr.Spec.ForProvider.LaunchTemplateData.CPUOptions.CoreCount)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.CPUOptions.ThreadsPerCore != nil {
				f0f2.SetThreadsPerCore(*cr.Spec.ForProvider.LaunchTemplateData.CPUOptions.ThreadsPerCore)
			}
			f0.SetCpuOptions(f0f2)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.CreditSpecification != nil {
			f0f3 := &svcsdk.CreditSpecificationRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.CreditSpecification.CPUCredits != nil {
				f0f3.SetCpuCredits(*cr.Spec.ForProvider.LaunchTemplateData.CreditSpecification.CPUCredits)
			}
			f0.SetCreditSpecification(f0f3)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.DisableAPITermination != nil {
			f0.SetDisableApiTermination(*cr.Spec.ForProvider.LaunchTemplateData.DisableAPITermination)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.EBSOptimized != nil {
			f0.SetEbsOptimized(*cr.Spec.ForProvider.LaunchTemplateData.EBSOptimized)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.ElasticGPUSpecifications != nil {
			f0f6 := []*svcsdk.ElasticGpuSpecification{}
			for _, f0f6iter := range cr.Spec.ForProvider.LaunchTemplateData.ElasticGPUSpecifications {
				f0f6elem := &svcsdk.ElasticGpuSpecification{}
				if f0f6iter.Type != nil {
					f0f6elem.SetType(*f0f6iter.Type)
				}
				f0f6 = append(f0f6, f0f6elem)
			}
			f0.SetElasticGpuSpecifications(f0f6)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.ElasticInferenceAccelerators != nil {
			f0f7 := []*svcsdk.LaunchTemplateElasticInferenceAccelerator{}
			for _, f0f7iter := range cr.Spec.ForProvider.LaunchTemplateData.ElasticInferenceAccelerators {
				f0f7elem := &svcsdk.LaunchTemplateElasticInferenceAccelerator{}
				if f0f7iter.Count != nil {
					f0f7elem.SetCount(*f0f7iter.Count)
				}
				if f0f7iter.Type != nil {
					f0f7elem.SetType(*f0f7iter.Type)
				}
				f0f7 = append(f0f7, f0f7elem)
			}
			f0.SetElasticInferenceAccelerators(f0f7)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.EnclaveOptions != nil {
			f0f8 := &svcsdk.LaunchTemplateEnclaveOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.EnclaveOptions.Enabled != nil {
				f0f8.SetEnabled(*cr.Spec.ForProvider.LaunchTemplateData.EnclaveOptions.Enabled)
			}
			f0.SetEnclaveOptions(f0f8)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.HibernationOptions != nil {
			f0f9 := &svcsdk.LaunchTemplateHibernationOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.HibernationOptions.Configured != nil {
				f0f9.SetConfigured(*cr.Spec.ForProvider.LaunchTemplateData.HibernationOptions.Configured)
			}
			f0.SetHibernationOptions(f0f9)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.IAMInstanceProfile != nil {
			f0f10 := &svcsdk.LaunchTemplateIamInstanceProfileSpecificationRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.IAMInstanceProfile.ARN != nil {
				f0f10.SetArn(*cr.Spec.ForProvider.LaunchTemplateData.IAMInstanceProfile.ARN)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.IAMInstanceProfile.Name != nil {
				f0f10.SetName(*cr.Spec.ForProvider.LaunchTemplateData.IAMInstanceProfile.Name)
			}
			f0.SetIamInstanceProfile(f0f10)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.ImageID != nil {
			f0.SetImageId(*cr.Spec.ForProvider.LaunchTemplateData.ImageID)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.InstanceInitiatedShutdownBehavior != nil {
			f0.SetInstanceInitiatedShutdownBehavior(*cr.Spec.ForProvider.LaunchTemplateData.InstanceInitiatedShutdownBehavior)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions != nil {
			f0f13 := &svcsdk.LaunchTemplateInstanceMarketOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.MarketType != nil {
				f0f13.SetMarketType(*cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.MarketType)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions != nil {
				f0f13f1 := &svcsdk.LaunchTemplateSpotMarketOptionsRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.BlockDurationMinutes != nil {
					f0f13f1.SetBlockDurationMinutes(*cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.BlockDurationMinutes)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior != nil {
					f0f13f1.SetInstanceInterruptionBehavior(*cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.MaxPrice != nil {
					f0f13f1.SetMaxPrice(*cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.MaxPrice)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.SpotInstanceType != nil {
					f0f13f1.SetSpotInstanceType(*cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.SpotInstanceType)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.ValidUntil != nil {
					f0f13f1.SetValidUntil(cr.Spec.ForProvider.LaunchTemplateData.InstanceMarketOptions.SpotOptions.ValidUntil.Time)
				}
				f0f13.SetSpotOptions(f0f13f1)
			}
			f0.SetInstanceMarketOptions(f0f13)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements != nil {
			f0f14 := &svcsdk.InstanceRequirementsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorCount != nil {
				f0f14f0 := &svcsdk.AcceleratorCountRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorCount.Max != nil {
					f0f14f0.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorCount.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorCount.Min != nil {
					f0f14f0.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorCount.Min)
				}
				f0f14.SetAcceleratorCount(f0f14f0)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorManufacturers != nil {
				f0f14f1 := []*string{}
				for _, f0f14f1iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorManufacturers {
					var f0f14f1elem string
					f0f14f1elem = *f0f14f1iter
					f0f14f1 = append(f0f14f1, &f0f14f1elem)
				}
				f0f14.SetAcceleratorManufacturers(f0f14f1)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorNames != nil {
				f0f14f2 := []*string{}
				for _, f0f14f2iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorNames {
					var f0f14f2elem string
					f0f14f2elem = *f0f14f2iter
					f0f14f2 = append(f0f14f2, &f0f14f2elem)
				}
				f0f14.SetAcceleratorNames(f0f14f2)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTotalMemoryMiB != nil {
				f0f14f3 := &svcsdk.AcceleratorTotalMemoryMiBRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTotalMemoryMiB.Max != nil {
					f0f14f3.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTotalMemoryMiB.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTotalMemoryMiB.Min != nil {
					f0f14f3.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTotalMemoryMiB.Min)
				}
				f0f14.SetAcceleratorTotalMemoryMiB(f0f14f3)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTypes != nil {
				f0f14f4 := []*string{}
				for _, f0f14f4iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.AcceleratorTypes {
					var f0f14f4elem string
					f0f14f4elem = *f0f14f4iter
					f0f14f4 = append(f0f14f4, &f0f14f4elem)
				}
				f0f14.SetAcceleratorTypes(f0f14f4)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BareMetal != nil {
				f0f14.SetBareMetal(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BareMetal)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BaselineEBSBandwidthMbps != nil {
				f0f14f6 := &svcsdk.BaselineEbsBandwidthMbpsRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BaselineEBSBandwidthMbps.Max != nil {
					f0f14f6.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BaselineEBSBandwidthMbps.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BaselineEBSBandwidthMbps.Min != nil {
					f0f14f6.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BaselineEBSBandwidthMbps.Min)
				}
				f0f14.SetBaselineEbsBandwidthMbps(f0f14f6)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BurstablePerformance != nil {
				f0f14.SetBurstablePerformance(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.BurstablePerformance)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.CPUManufacturers != nil {
				f0f14f8 := []*string{}
				for _, f0f14f8iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.CPUManufacturers {
					var f0f14f8elem string
					f0f14f8elem = *f0f14f8iter
					f0f14f8 = append(f0f14f8, &f0f14f8elem)
				}
				f0f14.SetCpuManufacturers(f0f14f8)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.ExcludedInstanceTypes != nil {
				f0f14f9 := []*string{}
				for _, f0f14f9iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.ExcludedInstanceTypes {
					var f0f14f9elem string
					f0f14f9elem = *f0f14f9iter
					f0f14f9 = append(f0f14f9, &f0f14f9elem)
				}
				f0f14.SetExcludedInstanceTypes(f0f14f9)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.InstanceGenerations != nil {
				f0f14f10 := []*string{}
				for _, f0f14f10iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.InstanceGenerations {
					var f0f14f10elem string
					f0f14f10elem = *f0f14f10iter
					f0f14f10 = append(f0f14f10, &f0f14f10elem)
				}
				f0f14.SetInstanceGenerations(f0f14f10)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.LocalStorage != nil {
				f0f14.SetLocalStorage(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.LocalStorage)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.LocalStorageTypes != nil {
				f0f14f12 := []*string{}
				for _, f0f14f12iter := range cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.LocalStorageTypes {
					var f0f14f12elem string
					f0f14f12elem = *f0f14f12iter
					f0f14f12 = append(f0f14f12, &f0f14f12elem)
				}
				f0f14.SetLocalStorageTypes(f0f14f12)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryGiBPerVCPU != nil {
				f0f14f13 := &svcsdk.MemoryGiBPerVCpuRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryGiBPerVCPU.Max != nil {
					f0f14f13.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryGiBPerVCPU.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryGiBPerVCPU.Min != nil {
					f0f14f13.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryGiBPerVCPU.Min)
				}
				f0f14.SetMemoryGiBPerVCpu(f0f14f13)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryMiB != nil {
				f0f14f14 := &svcsdk.MemoryMiBRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryMiB.Max != nil {
					f0f14f14.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryMiB.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryMiB.Min != nil {
					f0f14f14.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.MemoryMiB.Min)
				}
				f0f14.SetMemoryMiB(f0f14f14)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.NetworkInterfaceCount != nil {
				f0f14f15 := &svcsdk.NetworkInterfaceCountRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.NetworkInterfaceCount.Max != nil {
					f0f14f15.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.NetworkInterfaceCount.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.NetworkInterfaceCount.Min != nil {
					f0f14f15.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.NetworkInterfaceCount.Min)
				}
				f0f14.SetNetworkInterfaceCount(f0f14f15)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.OnDemandMaxPricePercentageOverLowestPrice != nil {
				f0f14.SetOnDemandMaxPricePercentageOverLowestPrice(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.OnDemandMaxPricePercentageOverLowestPrice)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.RequireHibernateSupport != nil {
				f0f14.SetRequireHibernateSupport(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.RequireHibernateSupport)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.SpotMaxPricePercentageOverLowestPrice != nil {
				f0f14.SetSpotMaxPricePercentageOverLowestPrice(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.SpotMaxPricePercentageOverLowestPrice)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.TotalLocalStorageGB != nil {
				f0f14f19 := &svcsdk.TotalLocalStorageGBRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.TotalLocalStorageGB.Max != nil {
					f0f14f19.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.TotalLocalStorageGB.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.TotalLocalStorageGB.Min != nil {
					f0f14f19.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.TotalLocalStorageGB.Min)
				}
				f0f14.SetTotalLocalStorageGB(f0f14f19)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.VCPUCount != nil {
				f0f14f20 := &svcsdk.VCpuCountRangeRequest{}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.VCPUCount.Max != nil {
					f0f14f20.SetMax(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.VCPUCount.Max)
				}
				if cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.VCPUCount.Min != nil {
					f0f14f20.SetMin(*cr.Spec.ForProvider.LaunchTemplateData.InstanceRequirements.VCPUCount.Min)
				}
				f0f14.SetVCpuCount(f0f14f20)
			}
			f0.SetInstanceRequirements(f0f14)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.InstanceType != nil {
			f0.SetInstanceType(*cr.Spec.ForProvider.LaunchTemplateData.InstanceType)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.KernelID != nil {
			f0.SetKernelId(*cr.Spec.ForProvider.LaunchTemplateData.KernelID)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.KeyName != nil {
			f0.SetKeyName(*cr.Spec.ForProvider.LaunchTemplateData.KeyName)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.LicenseSpecifications != nil {
			f0f18 := []*svcsdk.LaunchTemplateLicenseConfigurationRequest{}
			for _, f0f18iter := range cr.Spec.ForProvider.LaunchTemplateData.LicenseSpecifications {
				f0f18elem := &svcsdk.LaunchTemplateLicenseConfigurationRequest{}
				if f0f18iter.LicenseConfigurationARN != nil {
					f0f18elem.SetLicenseConfigurationArn(*f0f18iter.LicenseConfigurationARN)
				}
				f0f18 = append(f0f18, f0f18elem)
			}
			f0.SetLicenseSpecifications(f0f18)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions != nil {
			f0f19 := &svcsdk.LaunchTemplateInstanceMetadataOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPEndpoint != nil {
				f0f19.SetHttpEndpoint(*cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPEndpoint)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPProtocolIPv6 != nil {
				f0f19.SetHttpProtocolIpv6(*cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPProtocolIPv6)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPPutResponseHopLimit != nil {
				f0f19.SetHttpPutResponseHopLimit(*cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPPutResponseHopLimit)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPTokens != nil {
				f0f19.SetHttpTokens(*cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.HTTPTokens)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.InstanceMetadataTags != nil {
				f0f19.SetInstanceMetadataTags(*cr.Spec.ForProvider.LaunchTemplateData.MetadataOptions.InstanceMetadataTags)
			}
			f0.SetMetadataOptions(f0f19)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.Monitoring != nil {
			f0f20 := &svcsdk.LaunchTemplatesMonitoringRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.Monitoring.Enabled != nil {
				f0f20.SetEnabled(*cr.Spec.ForProvider.LaunchTemplateData.Monitoring.Enabled)
			}
			f0.SetMonitoring(f0f20)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.NetworkInterfaces != nil {
			f0f21 := []*svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
			for _, f0f21iter := range cr.Spec.ForProvider.LaunchTemplateData.NetworkInterfaces {
				f0f21elem := &svcsdk.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}
				if f0f21iter.AssociateCarrierIPAddress != nil {
					f0f21elem.SetAssociateCarrierIpAddress(*f0f21iter.AssociateCarrierIPAddress)
				}
				if f0f21iter.AssociatePublicIPAddress != nil {
					f0f21elem.SetAssociatePublicIpAddress(*f0f21iter.AssociatePublicIPAddress)
				}
				if f0f21iter.DeleteOnTermination != nil {
					f0f21elem.SetDeleteOnTermination(*f0f21iter.DeleteOnTermination)
				}
				if f0f21iter.Description != nil {
					f0f21elem.SetDescription(*f0f21iter.Description)
				}
				if f0f21iter.DeviceIndex != nil {
					f0f21elem.SetDeviceIndex(*f0f21iter.DeviceIndex)
				}
				if f0f21iter.Groups != nil {
					f0f21elemf5 := []*string{}
					for _, f0f21elemf5iter := range f0f21iter.Groups {
						var f0f21elemf5elem string
						f0f21elemf5elem = *f0f21elemf5iter
						f0f21elemf5 = append(f0f21elemf5, &f0f21elemf5elem)
					}
					f0f21elem.SetGroups(f0f21elemf5)
				}
				if f0f21iter.InterfaceType != nil {
					f0f21elem.SetInterfaceType(*f0f21iter.InterfaceType)
				}
				if f0f21iter.IPv4PrefixCount != nil {
					f0f21elem.SetIpv4PrefixCount(*f0f21iter.IPv4PrefixCount)
				}
				if f0f21iter.IPv4Prefixes != nil {
					f0f21elemf8 := []*svcsdk.Ipv4PrefixSpecificationRequest{}
					for _, f0f21elemf8iter := range f0f21iter.IPv4Prefixes {
						f0f21elemf8elem := &svcsdk.Ipv4PrefixSpecificationRequest{}
						if f0f21elemf8iter.IPv4Prefix != nil {
							f0f21elemf8elem.SetIpv4Prefix(*f0f21elemf8iter.IPv4Prefix)
						}
						f0f21elemf8 = append(f0f21elemf8, f0f21elemf8elem)
					}
					f0f21elem.SetIpv4Prefixes(f0f21elemf8)
				}
				if f0f21iter.IPv6AddressCount != nil {
					f0f21elem.SetIpv6AddressCount(*f0f21iter.IPv6AddressCount)
				}
				if f0f21iter.IPv6Addresses != nil {
					f0f21elemf10 := []*svcsdk.InstanceIpv6AddressRequest{}
					for _, f0f21elemf10iter := range f0f21iter.IPv6Addresses {
						f0f21elemf10elem := &svcsdk.InstanceIpv6AddressRequest{}
						if f0f21elemf10iter.IPv6Address != nil {
							f0f21elemf10elem.SetIpv6Address(*f0f21elemf10iter.IPv6Address)
						}
						f0f21elemf10 = append(f0f21elemf10, f0f21elemf10elem)
					}
					f0f21elem.SetIpv6Addresses(f0f21elemf10)
				}
				if f0f21iter.IPv6PrefixCount != nil {
					f0f21elem.SetIpv6PrefixCount(*f0f21iter.IPv6PrefixCount)
				}
				if f0f21iter.IPv6Prefixes != nil {
					f0f21elemf12 := []*svcsdk.Ipv6PrefixSpecificationRequest{}
					for _, f0f21elemf12iter := range f0f21iter.IPv6Prefixes {
						f0f21elemf12elem := &svcsdk.Ipv6PrefixSpecificationRequest{}
						if f0f21elemf12iter.IPv6Prefix != nil {
							f0f21elemf12elem.SetIpv6Prefix(*f0f21elemf12iter.IPv6Prefix)
						}
						f0f21elemf12 = append(f0f21elemf12, f0f21elemf12elem)
					}
					f0f21elem.SetIpv6Prefixes(f0f21elemf12)
				}
				if f0f21iter.NetworkCardIndex != nil {
					f0f21elem.SetNetworkCardIndex(*f0f21iter.NetworkCardIndex)
				}
				if f0f21iter.NetworkInterfaceID != nil {
					f0f21elem.SetNetworkInterfaceId(*f0f21iter.NetworkInterfaceID)
				}
				if f0f21iter.PrivateIPAddress != nil {
					f0f21elem.SetPrivateIpAddress(*f0f21iter.PrivateIPAddress)
				}
				if f0f21iter.PrivateIPAddresses != nil {
					f0f21elemf16 := []*svcsdk.PrivateIpAddressSpecification{}
					for _, f0f21elemf16iter := range f0f21iter.PrivateIPAddresses {
						f0f21elemf16elem := &svcsdk.PrivateIpAddressSpecification{}
						if f0f21elemf16iter.Primary != nil {
							f0f21elemf16elem.SetPrimary(*f0f21elemf16iter.Primary)
						}
						if f0f21elemf16iter.PrivateIPAddress != nil {
							f0f21elemf16elem.SetPrivateIpAddress(*f0f21elemf16iter.PrivateIPAddress)
						}
						f0f21elemf16 = append(f0f21elemf16, f0f21elemf16elem)
					}
					f0f21elem.SetPrivateIpAddresses(f0f21elemf16)
				}
				if f0f21iter.SecondaryPrivateIPAddressCount != nil {
					f0f21elem.SetSecondaryPrivateIpAddressCount(*f0f21iter.SecondaryPrivateIPAddressCount)
				}
				if f0f21iter.SubnetID != nil {
					f0f21elem.SetSubnetId(*f0f21iter.SubnetID)
				}
				f0f21 = append(f0f21, f0f21elem)
			}
			f0.SetNetworkInterfaces(f0f21)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.Placement != nil {
			f0f22 := &svcsdk.LaunchTemplatePlacementRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.Affinity != nil {
				f0f22.SetAffinity(*cr.Spec.ForProvider.LaunchTemplateData.Placement.Affinity)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.AvailabilityZone != nil {
				f0f22.SetAvailabilityZone(*cr.Spec.ForProvider.LaunchTemplateData.Placement.AvailabilityZone)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.GroupName != nil {
				f0f22.SetGroupName(*cr.Spec.ForProvider.LaunchTemplateData.Placement.GroupName)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.HostID != nil {
				f0f22.SetHostId(*cr.Spec.ForProvider.LaunchTemplateData.Placement.HostID)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.HostResourceGroupARN != nil {
				f0f22.SetHostResourceGroupArn(*cr.Spec.ForProvider.LaunchTemplateData.Placement.HostResourceGroupARN)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.PartitionNumber != nil {
				f0f22.SetPartitionNumber(*cr.Spec.ForProvider.LaunchTemplateData.Placement.PartitionNumber)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.SpreadDomain != nil {
				f0f22.SetSpreadDomain(*cr.Spec.ForProvider.LaunchTemplateData.Placement.SpreadDomain)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.Placement.Tenancy != nil {
				f0f22.SetTenancy(*cr.Spec.ForProvider.LaunchTemplateData.Placement.Tenancy)
			}
			f0.SetPlacement(f0f22)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions != nil {
			f0f23 := &svcsdk.LaunchTemplatePrivateDnsNameOptionsRequest{}
			if cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.EnableResourceNameDNSAAAARecord != nil {
				f0f23.SetEnableResourceNameDnsAAAARecord(*cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.EnableResourceNameDNSAAAARecord)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.EnableResourceNameDNSARecord != nil {
				f0f23.SetEnableResourceNameDnsARecord(*cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.EnableResourceNameDNSARecord)
			}
			if cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.HostnameType != nil {
				f0f23.SetHostnameType(*cr.Spec.ForProvider.LaunchTemplateData.PrivateDNSNameOptions.HostnameType)
			}
			f0.SetPrivateDnsNameOptions(f0f23)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.RAMDiskID != nil {
			f0.SetRamDiskId(*cr.Spec.ForProvider.LaunchTemplateData.RAMDiskID)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.SecurityGroupIDs != nil {
			f0f25 := []*string{}
			for _, f0f25iter := range cr.Spec.ForProvider.LaunchTemplateData.SecurityGroupIDs {
				var f0f25elem string
				f0f25elem = *f0f25iter
				f0f25 = append(f0f25, &f0f25elem)
			}
			f0.SetSecurityGroupIds(f0f25)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.SecurityGroups != nil {
			f0f26 := []*string{}
			for _, f0f26iter := range cr.Spec.ForProvider.LaunchTemplateData.SecurityGroups {
				var f0f26elem string
				f0f26elem = *f0f26iter
				f0f26 = append(f0f26, &f0f26elem)
			}
			f0.SetSecurityGroups(f0f26)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.TagSpecifications != nil {
			f0f27 := []*svcsdk.LaunchTemplateTagSpecificationRequest{}
			for _, f0f27iter := range cr.Spec.ForProvider.LaunchTemplateData.TagSpecifications {
				f0f27elem := &svcsdk.LaunchTemplateTagSpecificationRequest{}
				if f0f27iter.ResourceType != nil {
					f0f27elem.SetResourceType(*f0f27iter.ResourceType)
				}
				if f0f27iter.Tags != nil {
					f0f27elemf1 := []*svcsdk.Tag{}
					for _, f0f27elemf1iter := range f0f27iter.Tags {
						f0f27elemf1elem := &svcsdk.Tag{}
						if f0f27elemf1iter.Key != nil {
							f0f27elemf1elem.SetKey(*f0f27elemf1iter.Key)
						}
						if f0f27elemf1iter.Value != nil {
							f0f27elemf1elem.SetValue(*f0f27elemf1iter.Value)
						}
						f0f27elemf1 = append(f0f27elemf1, f0f27elemf1elem)
					}
					f0f27elem.SetTags(f0f27elemf1)
				}
				f0f27 = append(f0f27, f0f27elem)
			}
			f0.SetTagSpecifications(f0f27)
		}
		if cr.Spec.ForProvider.LaunchTemplateData.UserData != nil {
			f0.SetUserData(*cr.Spec.ForProvider.LaunchTemplateData.UserData)
		}
		res.SetLaunchTemplateData(f0)
	}
	if cr.Spec.ForProvider.LaunchTemplateName != nil {
		res.SetLaunchTemplateName(*cr.Spec.ForProvider.LaunchTemplateName)
	}
	if cr.Spec.ForProvider.TagSpecifications != nil {
		f2 := []*svcsdk.TagSpecification{}
		for _, f2iter := range cr.Spec.ForProvider.TagSpecifications {
			f2elem := &svcsdk.TagSpecification{}
			if f2iter.ResourceType != nil {
				f2elem.SetResourceType(*f2iter.ResourceType)
			}
			if f2iter.Tags != nil {
				f2elemf1 := []*svcsdk.Tag{}
				for _, f2elemf1iter := range f2iter.Tags {
					f2elemf1elem := &svcsdk.Tag{}
					if f2elemf1iter.Key != nil {
						f2elemf1elem.SetKey(*f2elemf1iter.Key)
					}
					if f2elemf1iter.Value != nil {
						f2elemf1elem.SetValue(*f2elemf1iter.Value)
					}
					f2elemf1 = append(f2elemf1, f2elemf1elem)
				}
				f2elem.SetTags(f2elemf1)
			}
			f2 = append(f2, f2elem)
		}
		res.SetTagSpecifications(f2)
	}
	if cr.Spec.ForProvider.VersionDescription != nil {
		res.SetVersionDescription(*cr.Spec.ForProvider.VersionDescription)
	}

	return res
}

// GenerateModifyLaunchTemplateInput returns an update input.
func GenerateModifyLaunchTemplateInput(cr *svcapitypes.LaunchTemplate) *svcsdk.ModifyLaunchTemplateInput {
	res := &svcsdk.ModifyLaunchTemplateInput{}

	if cr.Spec.ForProvider.LaunchTemplateName != nil {
		res.SetLaunchTemplateName(*cr.Spec.ForProvider.LaunchTemplateName)
	}

	return res
}

// GenerateDeleteLaunchTemplateInput returns a deletion input.
func GenerateDeleteLaunchTemplateInput(cr *svcapitypes.LaunchTemplate) *svcsdk.DeleteLaunchTemplateInput {
	res := &svcsdk.DeleteLaunchTemplateInput{}

	if cr.Spec.ForProvider.LaunchTemplateName != nil {
		res.SetLaunchTemplateName(*cr.Spec.ForProvider.LaunchTemplateName)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "InvalidLaunchTemplateName.NotFoundException"
}
