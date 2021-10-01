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

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
)

const (
	managedName           = "sample-instance"
	managedKind           = "instance.ec2.aws.crossplane.io"
	managedProviderConfig = "example"

	arch                  = "x86_64"
	assocID               = "assocId"
	assocState            = "assocState"
	attachmentID          = "attachId"
	blockDeviceName       = "/dev/xvda"
	capacityReservationID = "capResId"
	clientToken           = "clientToken"
	cpuCredits            = "cpuCredits"
	description           = "desc"
	elasticInferAccARN    = "inferArn"
	elasticInferAccID     = "inferId"
	elasticInferAccState  = "inferState"
	gpuID                 = "gpuId"
	gpuType               = "gpuType"
	groupID               = "groupId"
	groupName             = "groupName"
	hostID                = "hostId"
	hostGroupARN          = "hostGroupArn"
	iamARN                = "iamArn"
	iamID                 = "iamId"
	imageID               = "imageId"
	interfaceType         = "intType"
	ipOwnerID             = "ipOwnerId"
	ipv6Address           = "ipv6Address"
	kernelID              = "kernelId"
	keyName               = "keyName"
	launchTemplateID      = "launchTemplateId"
	launchTemplateName    = "launchTemplateName"
	licenseConfig         = "licenseConfig"
	macAddress            = "macAddress"
	outpostARN            = "outpostArn"
	placementAff          = "affinity"
	privateDNSName        = "privDnsName"
	privateIPAddress      = "privIpAddress"
	productCodeID         = "productCodeId"
	publicDNSName         = "publicDnsName"
	publicIPAddress       = "publicIp"
	ramDiskID             = "ramDiskId"
	rootDeviceName        = "rootDeviceName"
	snapshotID            = "snapshotId"
	spotInstanceReqID     = "spotInstacneId"
	spotMarketType        = "spotMarketType"
	spreadDomain          = "spreadDomain"
	sriovNetSupport       = "sriovNetSupport"
	stateReason           = "stateReason"
	tagResourceType       = "instance"
	tagsKey               = "key"
	tagsVal               = "value"
	userData              = "userData"
	volumeID              = "volId"
	volumeType            = "gp2"
)

func TestGenerateInstanceConditions(t *testing.T) {
	type args struct {
		obeserved manualv1alpha1.InstanceObservation
	}
	cases := map[string]struct {
		args args
		want Condition
	}{
		"InstanceIsRunning": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(ec2.InstanceStateNameRunning),
				},
			},
			want: Available,
		},
		"InstanceIsPending": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(ec2.InstanceStateNamePending),
				},
			},
			want: Creating,
		},
		"InstanceIsStopping": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(ec2.InstanceStateNameStopping),
				},
			},
			want: Deleting,
		},
		"InstanceIsShuttingDown": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(ec2.InstanceStateNameShuttingDown),
				},
			},
			want: Deleting,
		},
		"InstanceIsTerminated": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(ec2.InstanceStateNameTerminated),
				},
			},
			want: Deleted,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			condition := GenerateInstanceCondition(tc.args.obeserved)

			if diff := cmp.Diff(tc.want, condition, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDescribeInstancesByExternalTags(t *testing.T) {
	type args struct {
		extTags map[string]string
	}
	cases := map[string]struct {
		args args
		want *ec2.DescribeInstancesInput
	}{
		"TagsAreAddedToFilter": {
			args: args{
				extTags: map[string]string{
					"crossplane-name":           managedName,
					"crossplane-kind":           managedKind,
					"crossplane-providerconfig": managedProviderConfig,
				},
			},
			want: &ec2.DescribeInstancesInput{
				Filters: []ec2.Filter{
					{
						Name:   aws.String("tag:crossplane-kind"),
						Values: []string{managedKind},
					},
					{
						Name:   aws.String("tag:crossplane-name"),
						Values: []string{managedName},
					},
					{
						Name:   aws.String("tag:crossplane-providerconfig"),
						Values: []string{managedProviderConfig},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateDescribeInstancesByExternalTags(tc.args.extTags)

			if diff := cmp.Diff(tc.want, input, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateInstanceObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2.Instance
		out manualv1alpha1.InstanceObservation
	}{
		"AllFilled": {
			in: ec2.Instance{
				AmiLaunchIndex: aws.Int64(0),
				Architecture:   arch,
				BlockDeviceMappings: []ec2.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						Ebs: &ec2.EbsInstanceBlockDevice{
							AttachTime:          nil,
							DeleteOnTermination: aws.Bool(false),
							Status:              ec2.AttachmentStatusAttached,
							VolumeId:            aws.String(volumeID),
						},
					},
				},
				CapacityReservationId: aws.String(capacityReservationID),
				CapacityReservationSpecification: &ec2.CapacityReservationSpecificationResponse{
					CapacityReservationPreference: ec2.CapacityReservationPreferenceNone,
					CapacityReservationTarget: &ec2.CapacityReservationTargetResponse{
						CapacityReservationId: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CpuOptions: &ec2.CpuOptions{
					CoreCount:      aws.Int64(1),
					ThreadsPerCore: aws.Int64(1),
				},
				EbsOptimized: aws.Bool(false),
				ElasticGpuAssociations: []ec2.ElasticGpuAssociation{
					{
						ElasticGpuAssociationId:    aws.String(assocID),
						ElasticGpuAssociationState: aws.String(assocState),
						ElasticGpuAssociationTime:  aws.String("now"),
						ElasticGpuId:               aws.String(gpuID),
					},
				},
				ElasticInferenceAcceleratorAssociations: []ec2.ElasticInferenceAcceleratorAssociation{
					{
						ElasticInferenceAcceleratorArn:              aws.String(elasticInferAccARN),
						ElasticInferenceAcceleratorAssociationId:    aws.String(elasticInferAccID),
						ElasticInferenceAcceleratorAssociationState: aws.String(elasticInferAccState),
						ElasticInferenceAcceleratorAssociationTime:  nil,
					},
				},
				EnaSupport: aws.Bool(false),
				HibernationOptions: &ec2.HibernationOptions{
					Configured: aws.Bool(false),
				},
				Hypervisor: ec2.HypervisorTypeOvm,
				IamInstanceProfile: &ec2.IamInstanceProfile{
					Arn: aws.String(iamARN),
					Id:  aws.String(iamID),
				},
				ImageId:           aws.String(imageID),
				InstanceId:        aws.String(instanceID),
				InstanceLifecycle: ec2.InstanceLifecycleTypeScheduled,
				InstanceType:      ec2.InstanceTypeM1Small,
				KernelId:          aws.String(kernelID),
				LaunchTime:        nil,
				Licenses: []ec2.LicenseConfiguration{
					{
						LicenseConfigurationArn: aws.String(licenseConfig),
					},
				},
				MetadataOptions: &ec2.InstanceMetadataOptionsResponse{
					HttpEndpoint:            ec2.InstanceMetadataEndpointStateEnabled,
					HttpPutResponseHopLimit: aws.Int64(0),
					HttpTokens:              ec2.HttpTokensStateOptional,
					State:                   ec2.InstanceMetadataOptionsStateApplied,
				},
				Monitoring: &ec2.Monitoring{
					State: ec2.MonitoringStateEnabled,
				},
				NetworkInterfaces: []ec2.InstanceNetworkInterface{
					{
						Association: &ec2.InstanceNetworkInterfaceAssociation{
							IpOwnerId:     aws.String(ipOwnerID),
							PublicDnsName: aws.String(publicDNSName),
							PublicIp:      aws.String(publicIPAddress),
						},
						Attachment: &ec2.InstanceNetworkInterfaceAttachment{
							AttachTime:          nil,
							AttachmentId:        aws.String(attachmentID),
							DeleteOnTermination: aws.Bool(false),
							DeviceIndex:         aws.Int64(0),
							Status:              ec2.AttachmentStatusAttached,
						},
						Description: aws.String(description),
						Groups: []ec2.GroupIdentifier{
							{
								GroupId:   aws.String(groupID),
								GroupName: aws.String(groupName),
							},
						},
						InterfaceType: aws.String(interfaceType),
						Ipv6Addresses: []ec2.InstanceIpv6Address{
							{
								Ipv6Address: aws.String(ipv6Address),
							},
						},
						MacAddress:         aws.String(macAddress),
						NetworkInterfaceId: aws.String(natNetworkInterfaceID),
						OwnerId:            aws.String(ownerID),
						PrivateDnsName:     aws.String(privateDNSName),
						PrivateIpAddress:   aws.String(privateIPAddress),
						PrivateIpAddresses: []ec2.InstancePrivateIpAddress{
							{
								Association: &ec2.InstanceNetworkInterfaceAssociation{
									IpOwnerId:     aws.String(ipOwnerID),
									PublicDnsName: aws.String(publicDNSName),
									PublicIp:      aws.String(publicIPAddress),
								},
							},
						},
						SourceDestCheck: aws.Bool(false),
						Status:          ec2.NetworkInterfaceStatusAvailable,
						SubnetId:        aws.String(subnetID),
						VpcId:           aws.String(vpcID),
					},
				},
				OutpostArn: aws.String(outpostARN),
				Placement: &ec2.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostId:               aws.String(hostID),
					HostResourceGroupArn: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int64(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              ec2.TenancyHost,
				},
				Platform:         ec2.PlatformValuesWindows,
				PrivateDnsName:   aws.String(privateDNSName),
				PrivateIpAddress: aws.String(privateIPAddress),
				ProductCodes: []ec2.ProductCode{
					{
						ProductCodeId:   aws.String(productCodeID),
						ProductCodeType: ec2.ProductCodeValuesMarketplace,
					},
				},
				PublicDnsName:   aws.String(publicDNSName),
				PublicIpAddress: aws.String(publicIPAddress),
				RamdiskId:       aws.String(ramDiskID),
				RootDeviceName:  aws.String(rootDeviceName),
				RootDeviceType:  ec2.DeviceTypeEbs,
				SecurityGroups: []ec2.GroupIdentifier{
					{
						GroupId:   aws.String(groupID),
						GroupName: aws.String(groupName),
					},
				},
				SourceDestCheck:       aws.Bool(false),
				SpotInstanceRequestId: aws.String(spotInstanceReqID),
				SriovNetSupport:       aws.String(sriovNetSupport),
				State: &ec2.InstanceState{
					Name: ec2.InstanceStateNameRunning,
				},
				StateReason: &ec2.StateReason{
					Message: aws.String(stateReason),
				},
				StateTransitionReason: aws.String(stateReason),
				SubnetId:              aws.String(subnetID),
				Tags: []ec2.Tag{
					{
						Key:   aws.String(tagsKey),
						Value: aws.String(tagsVal),
					},
				},
				VirtualizationType: ec2.VirtualizationTypeHvm,
				VpcId:              aws.String(vpcID),
			},
			out: manualv1alpha1.InstanceObservation{
				AmiLaunchIndex: aws.Int64(0),
				Architecture:   arch,
				BlockDeviceMapping: []manualv1alpha1.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						EBS: &manualv1alpha1.EBSInstanceBlockDevice{
							AttachTime:          nil,
							DeleteOnTermination: aws.Bool(false),
							Status:              string(ec2.AttachmentStatusAttached),
							VolumeID:            aws.String(volumeID),
						},
					},
				},
				CapacityReservationID: aws.String(capacityReservationID),
				CapacityReservationSpecification: &manualv1alpha1.CapacityReservationSpecificationResponse{
					CapacityReservationPreference: string(ec2.CapacityReservationPreferenceNone),
					CapacityReservationTarget: &manualv1alpha1.CapacityReservationTarget{
						CapacityReservationID: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CPUOptons: &manualv1alpha1.CPUOptionsRequest{
					CoreCount:      aws.Int64(1),
					ThreadsPerCore: aws.Int64(1),
				},
				EBSOptimized: aws.Bool(false),
				EnaSupport:   aws.Bool(false),
				ElasticGPUAssociations: []manualv1alpha1.ElasticGPUAssociation{
					{
						ElasticGPUAssociationID:    aws.String(assocID),
						ElasticGPUAssociationState: aws.String(assocState),
						ElasticGPUAssociationTime:  aws.String("now"),
						ElasticGPUID:               aws.String(gpuID),
					},
				},
				ElasticInferenceAcceleratorAssociations: []manualv1alpha1.ElasticInferenceAcceleratorAssociation{
					{
						ElasticInferenceAcceleratorARN:              aws.String(elasticInferAccARN),
						ElasticInferenceAcceleratorAssociationID:    aws.String(elasticInferAccID),
						ElasticInferenceAcceleratorAssociationState: aws.String(elasticInferAccState),
						ElasticInferenceAcceleratorAssociationTime:  nil,
					},
				},
				HibernationOptions: &manualv1alpha1.HibernationOptionsRequest{
					Configured: aws.Bool(false),
				},
				Hypervisor: string(ec2.HypervisorTypeOvm),
				IAMInstanceProfile: &manualv1alpha1.IAMInstanceProfile{
					ARN: aws.String(iamARN),
					ID:  aws.String(iamID),
				},
				ImageID:           aws.String(imageID),
				InstanceID:        aws.String(instanceID),
				InstanceLifecycle: string(ec2.InstanceLifecycleTypeScheduled),
				InstanceType:      string(ec2.InstanceTypeM1Small),
				KernelID:          aws.String(kernelID),
				Licenses: []manualv1alpha1.LicenseConfigurationRequest{
					{
						LicenseConfigurationARN: aws.String(licenseConfig),
					},
				},
				MetadataOptions: &manualv1alpha1.InstanceMetadataOptionsRequest{
					HTTPEndpoint:            string(ec2.InstanceMetadataEndpointStateEnabled),
					HTTPPutResponseHopLimit: aws.Int64(0),
					HTTPTokens:              string(ec2.HttpTokensStateOptional),
				},
				Monitoring: &manualv1alpha1.Monitoring{
					State: string(ec2.MonitoringStateEnabled),
				},
				NetworkInterfaces: []manualv1alpha1.InstanceNetworkInterface{
					{
						Association: &manualv1alpha1.InstanceNetworkInterfaceAssociation{
							IPOwnerID:     aws.String(ipOwnerID),
							PublicDNSName: aws.String(publicDNSName),
							PublicIP:      aws.String(publicIPAddress),
						},
						Attachment: &manualv1alpha1.InstanceNetworkInterfaceAttachment{
							AttachTime:          nil,
							AttachmentID:        aws.String(attachmentID),
							DeleteOnTermination: aws.Bool(false),
							DeviceIndex:         aws.Int64(0),
							Status:              string(ec2.AttachmentStatusAttached),
						},
						Description: aws.String(description),
						Groups: []manualv1alpha1.GroupIdentifier{
							{
								GroupID:   groupID,
								GroupName: groupName,
							},
						},
						InterfaceType: aws.String(interfaceType),
						IPv6Addresses: []manualv1alpha1.InstanceIPv6Address{
							{
								IPv6Address: aws.String(ipv6Address),
							},
						},
						MacAddress:         aws.String(macAddress),
						NetworkInterfaceID: aws.String(natNetworkInterfaceID),
						OwnerID:            aws.String(ownerID),
						PrivateDNSName:     aws.String(privateDNSName),
						PrivateIPAddress:   aws.String(privateIPAddress),
						PrivateIPAddresses: []manualv1alpha1.InstancePrivateIPAddress{
							{
								Association: &manualv1alpha1.InstanceNetworkInterfaceAssociation{
									IPOwnerID:     aws.String(ipOwnerID),
									PublicDNSName: aws.String(publicDNSName),
									PublicIP:      aws.String(publicIPAddress),
								},
							},
						},
						SourceDestCheck: aws.Bool(false),
						Status:          string(ec2.NetworkInterfaceStatusAvailable),
						SubnetID:        aws.String(subnetID),
						VPCID:           aws.String(vpcID),
					},
				},
				OutpostARN: aws.String(outpostARN),
				Platform:   string(ec2.PlatformValuesWindows),
				Placement: &manualv1alpha1.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostID:               aws.String(hostID),
					HostResourceGroupARN: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int64(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              string(ec2.TenancyHost),
				},
				PrivateDNSName:   aws.String(privateDNSName),
				PrivateIPAddress: aws.String(privateIPAddress),
				ProductCodes: []manualv1alpha1.ProductCode{
					{
						ProductCodeID:   aws.String(productCodeID),
						ProductCodeType: string(ec2.ProductCodeValuesMarketplace),
					},
				},
				PublicDNSName:   aws.String(publicDNSName),
				PublicIPAddress: aws.String(publicIPAddress),
				RAMDiskID:       aws.String(ramDiskID),
				RootDeviceName:  aws.String(rootDeviceName),
				RootDeviceType:  string(ec2.DeviceTypeEbs),
				SecurityGroups: []manualv1alpha1.GroupIdentifier{
					{
						GroupID:   groupID,
						GroupName: groupName,
					},
				},
				SourceDestCheck:       aws.Bool(false),
				SpotInstanceRequestID: aws.String(spotInstanceReqID),
				SriovNetSupport:       aws.String(sriovNetSupport),
				State:                 string(ec2.InstanceStateNameRunning),
				StateReason: &manualv1alpha1.StateReason{
					Message: aws.String(stateReason),
				},
				StateTransitionReason: aws.String(stateReason),
				SubnetID:              aws.String(subnetID),
				Tags: []manualv1alpha1.Tag{
					{
						Key:   tagsKey,
						Value: tagsVal,
					},
				},
				VirtualizationType: string(ec2.VirtualizationTypeHvm),
				VPCID:              aws.String(vpcID),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateInstanceObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateInstanceObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEC2RunInstancesInput(t *testing.T) {
	cases := map[string]struct {
		name string
		in   *manualv1alpha1.InstanceParameters
		out  *ec2.RunInstancesInput
	}{
		"MinFilled": {
			name: managedName,
			in: &manualv1alpha1.InstanceParameters{
				ImageID: aws.String(imageID),
			},
			out: &ec2.RunInstancesInput{
				ImageId:  aws.String(imageID),
				MaxCount: aws.Int64(1),
				MinCount: aws.Int64(1),
			},
		},
		"AllFilled": {
			name: managedName,
			in: &manualv1alpha1.InstanceParameters{
				BlockDeviceMappings: []manualv1alpha1.BlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						EBS: &manualv1alpha1.EBSBlockDevice{
							DeleteOnTermination: aws.Bool(false),
							Encrypted:           aws.Bool(false),
							IOps:                aws.Int64(1),
							KmsKeyID:            aws.String(keyName),
							SnapshotID:          aws.String(snapshotID),
							VolumeSize:          aws.Int64(1),
							VolumeType:          volumeType,
						},
					},
				},
				CapacityReservationSpecification: &manualv1alpha1.CapacityReservationSpecification{
					CapacityReservationPreference: string(ec2.CapacityReservationPreferenceNone),
					CapacityReservationTarget: &manualv1alpha1.CapacityReservationTarget{
						CapacityReservationID: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CPUOptions: &manualv1alpha1.CPUOptionsRequest{
					CoreCount:      aws.Int64(1),
					ThreadsPerCore: aws.Int64(1),
				},
				CreditSpecification: &manualv1alpha1.CreditSpecificationRequest{
					CPUCredits: aws.String(cpuCredits),
				},
				DisableAPITermination: aws.Bool(false),
				EBSOptimized:          aws.Bool(false),
				ElasticGPUSpecification: []manualv1alpha1.ElasticGPUSpecification{
					{
						Type: aws.String(gpuType),
					},
				},
				ElasticInferenceAccelerators: []manualv1alpha1.ElasticInferenceAccelerator{
					{
						Count: aws.Int64(1),
						Type:  aws.String(gpuType),
					},
				},
				HibernationOptions: &manualv1alpha1.HibernationOptionsRequest{
					Configured: aws.Bool(false),
				},
				IAMInstanceProfile: &manualv1alpha1.IAMInstanceProfileSpecification{
					ARN:  aws.String(iamARN),
					Name: aws.String(iamID),
				},
				ImageID:                           aws.String(imageID),
				InstanceInitiatedShutdownBehavior: string(ec2.ShutdownBehaviorStop),
				InstanceMarketOptions: &manualv1alpha1.InstanceMarketOptionsRequest{
					MarketType: spotMarketType,
					SpotOptions: &manualv1alpha1.SpotMarketOptions{
						BlockDurationMinutes:         aws.Int64(1),
						InstanceInterruptionBehavior: string(ec2.InstanceInterruptionBehaviorHibernate),
						MaxPrice:                     aws.String("1"),
						SpotInstanceType:             string(ec2.SpotInstanceTypeOneTime),
					},
				},
				InstanceType:     string(ec2.InstanceTypeA12xlarge),
				IPv6AddressCount: aws.Int64(1),
				IPv6Addresses: []manualv1alpha1.InstanceIPv6Address{
					{
						IPv6Address: aws.String(ipv6Address),
					},
				},
				KernelID: aws.String(kernelID),
				KeyName:  aws.String(keyName),
				LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
					LaunchTemplateID:   aws.String(launchTemplateID),
					LaunchTemplateName: aws.String(launchTemplateName),
					Version:            aws.String("1"),
				},
				LicenseSpecifications: []manualv1alpha1.LicenseConfigurationRequest{
					{
						LicenseConfigurationARN: aws.String(licenseConfig),
					},
				},
				MetadataOptions: &manualv1alpha1.InstanceMetadataOptionsRequest{
					HTTPEndpoint:            string(ec2.InstanceMetadataEndpointStateEnabled),
					HTTPPutResponseHopLimit: aws.Int64(0),
					HTTPTokens:              string(ec2.HttpTokensStateOptional),
				},
				Monitoring: &manualv1alpha1.RunInstancesMonitoringEnabled{
					Enabled: aws.Bool(false),
				},
				NetworkInterfaces: []manualv1alpha1.InstanceNetworkInterfaceSpecification{
					{
						AssociatePublicIPAddress: aws.Bool(false),
						DeleteOnTermination:      aws.Bool(false),
						Description:              aws.String(description),
						DeviceIndex:              aws.Int64(0),
						Groups: []string{
							groupID,
						},
						InterfaceType:    aws.String(interfaceType),
						IPv6AddressCount: aws.Int64(1),
						IPv6Addresses: []manualv1alpha1.InstanceIPv6Address{
							{
								IPv6Address: aws.String(ipv6Address),
							},
						},
						NetworkInterfaceID: aws.String(networkInterfaceID),
						PrivateIPAddress:   aws.String(privateIPAddress),
						PrivateIPAddresses: []manualv1alpha1.PrivateIPAddressSpecification{
							{
								Primary:          aws.Bool(false),
								PrivateIPAddress: aws.String(privateIPAddress),
							},
						},
						SecondaryPrivateIPAddressCount: aws.Int64(0),
						SubnetID:                       aws.String(subnetID),
					},
				},
				Placement: &manualv1alpha1.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostID:               aws.String(hostID),
					HostResourceGroupARN: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int64(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              string(ec2.TenancyHost),
				},
				PrivateIPAddress: aws.String(privateIPAddress),
				RAMDiskID:        aws.String(ramDiskID),
				SecurityGroupIDs: []string{
					groupID,
				},
				SubnetID: aws.String(subnetID),
				TagSpecifications: []manualv1alpha1.TagSpecification{
					{
						ResourceType: aws.String(tagResourceType),
						Tags: []manualv1alpha1.Tag{
							{
								Key:   tagsKey,
								Value: tagsVal,
							},
						},
					},
				},
				UserData: aws.String(userData),
			},
			out: &ec2.RunInstancesInput{
				BlockDeviceMappings: []ec2.BlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						Ebs: &ec2.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(false),
							Encrypted:           aws.Bool(false),
							Iops:                aws.Int64(1),
							KmsKeyId:            aws.String(keyName),
							SnapshotId:          aws.String(snapshotID),
							VolumeSize:          aws.Int64(1),
							VolumeType:          volumeType,
						},
					},
				},
				CapacityReservationSpecification: &ec2.CapacityReservationSpecification{
					CapacityReservationPreference: ec2.CapacityReservationPreferenceNone,
					CapacityReservationTarget: &ec2.CapacityReservationTarget{
						CapacityReservationId: aws.String(capacityReservationID),
					},
				},
				CpuOptions: &ec2.CpuOptionsRequest{
					CoreCount:      aws.Int64(1),
					ThreadsPerCore: aws.Int64(1),
				},
				CreditSpecification: &ec2.CreditSpecificationRequest{
					CpuCredits: aws.String(cpuCredits),
				},
				ClientToken:           aws.String(clientToken),
				DisableApiTermination: aws.Bool(false),
				EbsOptimized:          aws.Bool(false),
				ElasticGpuSpecification: []ec2.ElasticGpuSpecification{
					{
						Type: aws.String(gpuType),
					},
				},
				ElasticInferenceAccelerators: []ec2.ElasticInferenceAccelerator{
					{
						Count: aws.Int64(1),
						Type:  aws.String(gpuType),
					},
				},
				HibernationOptions: &ec2.HibernationOptionsRequest{
					Configured: aws.Bool(false),
				},
				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Arn:  aws.String(iamARN),
					Name: aws.String(iamID),
				},
				ImageId: aws.String(imageID),
				InstanceMarketOptions: &ec2.InstanceMarketOptionsRequest{
					MarketType: spotMarketType,
					SpotOptions: &ec2.SpotMarketOptions{
						BlockDurationMinutes:         aws.Int64(1),
						InstanceInterruptionBehavior: ec2.InstanceInterruptionBehaviorHibernate,
						MaxPrice:                     aws.String("1"),
						SpotInstanceType:             ec2.SpotInstanceTypeOneTime,
					},
				},
				InstanceType:                      ec2.InstanceTypeA12xlarge,
				InstanceInitiatedShutdownBehavior: ec2.ShutdownBehaviorStop,
				Ipv6AddressCount:                  aws.Int64(1),
				Ipv6Addresses: []ec2.InstanceIpv6Address{
					{
						Ipv6Address: aws.String(ipv6Address),
					},
				},
				KernelId: aws.String(kernelID),
				KeyName:  aws.String(keyName),
				LaunchTemplate: &ec2.LaunchTemplateSpecification{
					LaunchTemplateId:   aws.String(launchTemplateID),
					LaunchTemplateName: aws.String(launchTemplateName),
					Version:            aws.String("1"),
				},
				LicenseSpecifications: []ec2.LicenseConfigurationRequest{
					{
						LicenseConfigurationArn: aws.String(licenseConfig),
					},
				},
				MinCount: aws.Int64(1),
				MaxCount: aws.Int64(1),
				MetadataOptions: &ec2.InstanceMetadataOptionsRequest{
					HttpEndpoint:            ec2.InstanceMetadataEndpointStateEnabled,
					HttpPutResponseHopLimit: aws.Int64(0),
					HttpTokens:              ec2.HttpTokensStateOptional,
				},
				Monitoring: &ec2.RunInstancesMonitoringEnabled{
					Enabled: aws.Bool(false),
				},
				NetworkInterfaces: []ec2.InstanceNetworkInterfaceSpecification{
					{
						AssociatePublicIpAddress: aws.Bool(false),
						DeleteOnTermination:      aws.Bool(false),
						Description:              aws.String(description),
						DeviceIndex:              aws.Int64(0),
						Groups: []string{
							groupID,
						},
						InterfaceType:    aws.String(interfaceType),
						Ipv6AddressCount: aws.Int64(1),
						Ipv6Addresses: []ec2.InstanceIpv6Address{
							{
								Ipv6Address: aws.String(ipv6Address),
							},
						},
						NetworkInterfaceId: aws.String(networkInterfaceID),
						PrivateIpAddress:   aws.String(privateIPAddress),
						PrivateIpAddresses: []ec2.PrivateIpAddressSpecification{
							{
								Primary:          aws.Bool(false),
								PrivateIpAddress: aws.String(privateIPAddress),
							},
						},
						SecondaryPrivateIpAddressCount: aws.Int64(0),
						SubnetId:                       aws.String(subnetID),
					},
				},
				Placement: &ec2.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostId:               aws.String(hostID),
					HostResourceGroupArn: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int64(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              ec2.TenancyHost,
				},
				PrivateIpAddress: aws.String(privateIPAddress),
				RamdiskId:        aws.String(ramDiskID),
				SecurityGroupIds: []string{
					groupID,
				},
				SubnetId: aws.String(subnetID),
				TagSpecifications: []ec2.TagSpecification{
					{
						ResourceType: ec2.ResourceTypeInstance,
						Tags: []ec2.Tag{
							{
								Key:   aws.String(tagsKey),
								Value: aws.String(tagsVal),
							},
						},
					},
				},
				UserData: aws.String(userData),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateEC2RunInstancesInput(tc.name, tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateEC2RunInstancesInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}
