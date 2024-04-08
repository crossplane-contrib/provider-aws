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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
)

const (
	managedName           = "sample-instance"
	managedKind           = "instance.ec2.aws.crossplane.io"
	managedProviderConfig = "example"

	arch                              = "x86_64"
	assocID                           = "assocId"
	assocState                        = "assocState"
	attachmentID                      = "attachId"
	blockDeviceName                   = "/dev/xvda"
	capacityReservationID             = "capResId"
	clientToken                       = "clientToken"
	cpuCredits                        = "cpuCredits"
	description                       = "desc"
	elasticInferAccARN                = "inferArn"
	elasticInferAccID                 = "inferId"
	elasticInferAccState              = "inferState"
	gpuID                             = "gpuId"
	gpuType                           = "gpuType"
	groupID                           = "groupId"
	groupName                         = "groupName"
	hostID                            = "hostId"
	hostGroupARN                      = "hostGroupArn"
	iamARN                            = "iamArn"
	iamID                             = "iamId"
	imageID                           = "imageId"
	instanceInitiatedShutdownBehavior = "stop"
	interfaceType                     = "intType"
	ipOwnerID                         = "ipOwnerId"
	ipv6Address                       = "ipv6Address"
	ipv6Prefix                        = "ipv6Prefix"
	kernelID                          = "kernelId"
	keyName                           = "keyName"
	launchTemplateID                  = "launchTemplateId"
	launchTemplateName                = "launchTemplateName"
	licenseConfig                     = "licenseConfig"
	macAddress                        = "macAddress"
	outpostARN                        = "outpostArn"
	placementAff                      = "affinity"
	privateDNSName                    = "privDnsName"
	privateIPAddress                  = "privIpAddress"
	productCodeID                     = "productCodeId"
	publicDNSName                     = "publicDnsName"
	publicIPAddress                   = "publicIp"
	ramDiskID                         = "ramDiskId"
	rootDeviceName                    = "rootDeviceName"
	snapshotID                        = "snapshotId"
	spotInstanceReqID                 = "spotInstacneId"
	spotMarketType                    = "spotMarketType"
	spreadDomain                      = "spreadDomain"
	sriovNetSupport                   = "sriovNetSupport"
	stateReason                       = "stateReason"
	tagResourceType                   = "instance"
	tagsKey                           = "key"
	tagsVal                           = "value"
	userData                          = "userData"
	volumeID                          = "volId"
	volumeType                        = "gp2"
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
					State: string(types.InstanceStateNameRunning),
				},
			},
			want: Available,
		},
		"InstanceIsPending": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(types.InstanceStateNamePending),
				},
			},
			want: Creating,
		},
		"InstanceIsStopping": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(types.InstanceStateNameStopping),
				},
			},
			want: Deleting,
		},
		"InstanceIsShuttingDown": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(types.InstanceStateNameShuttingDown),
				},
			},
			want: Deleting,
		},
		"InstanceIsTerminated": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: string(types.InstanceStateNameTerminated),
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
				Filters: []types.Filter{
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

			if diff := cmp.Diff(tc.want, input, test.EquateConditions(), cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateInstanceObservation(t *testing.T) {
	cases := map[string]struct {
		in         types.Instance
		attributes ec2.DescribeInstanceAttributeOutput
		out        manualv1alpha1.InstanceObservation
	}{
		"AllFilled": {
			in: types.Instance{
				AmiLaunchIndex: aws.Int32(0),
				Architecture:   arch,
				BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						Ebs: &types.EbsInstanceBlockDevice{
							AttachTime:          nil,
							DeleteOnTermination: aws.Bool(false),
							Status:              types.AttachmentStatusAttached,
							VolumeId:            aws.String(volumeID),
						},
					},
				},
				CapacityReservationId: aws.String(capacityReservationID),
				CapacityReservationSpecification: &types.CapacityReservationSpecificationResponse{
					CapacityReservationPreference: types.CapacityReservationPreferenceNone,
					CapacityReservationTarget: &types.CapacityReservationTargetResponse{
						CapacityReservationId: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CpuOptions: &types.CpuOptions{
					CoreCount:      aws.Int32(1),
					ThreadsPerCore: aws.Int32(1),
				},
				EbsOptimized: aws.Bool(false),
				ElasticGpuAssociations: []types.ElasticGpuAssociation{
					{
						ElasticGpuAssociationId:    aws.String(assocID),
						ElasticGpuAssociationState: aws.String(assocState),
						ElasticGpuAssociationTime:  aws.String("now"),
						ElasticGpuId:               aws.String(gpuID),
					},
				},
				ElasticInferenceAcceleratorAssociations: []types.ElasticInferenceAcceleratorAssociation{
					{
						ElasticInferenceAcceleratorArn:              aws.String(elasticInferAccARN),
						ElasticInferenceAcceleratorAssociationId:    aws.String(elasticInferAccID),
						ElasticInferenceAcceleratorAssociationState: aws.String(elasticInferAccState),
						ElasticInferenceAcceleratorAssociationTime:  nil,
					},
				},
				EnaSupport: aws.Bool(false),
				HibernationOptions: &types.HibernationOptions{
					Configured: aws.Bool(false),
				},
				Hypervisor: types.HypervisorTypeOvm,
				IamInstanceProfile: &types.IamInstanceProfile{
					Arn: aws.String(iamARN),
					Id:  aws.String(iamID),
				},
				ImageId:           aws.String(imageID),
				InstanceId:        aws.String(instanceID),
				InstanceLifecycle: types.InstanceLifecycleTypeScheduled,
				InstanceType:      types.InstanceTypeM1Small,
				KernelId:          aws.String(kernelID),
				LaunchTime:        nil,
				Licenses: []types.LicenseConfiguration{
					{
						LicenseConfigurationArn: aws.String(licenseConfig),
					},
				},
				MetadataOptions: &types.InstanceMetadataOptionsResponse{
					HttpEndpoint:            types.InstanceMetadataEndpointStateEnabled,
					HttpPutResponseHopLimit: aws.Int32(0),
					HttpTokens:              types.HttpTokensStateOptional,
					State:                   types.InstanceMetadataOptionsStateApplied,
				},
				Monitoring: &types.Monitoring{
					State: types.MonitoringStateEnabled,
				},
				NetworkInterfaces: []types.InstanceNetworkInterface{
					{
						Association: &types.InstanceNetworkInterfaceAssociation{
							IpOwnerId:     aws.String(ipOwnerID),
							PublicDnsName: aws.String(publicDNSName),
							PublicIp:      aws.String(publicIPAddress),
						},
						Attachment: &types.InstanceNetworkInterfaceAttachment{
							AttachTime:          nil,
							AttachmentId:        aws.String(attachmentID),
							DeleteOnTermination: aws.Bool(false),
							DeviceIndex:         aws.Int32(0),
							Status:              types.AttachmentStatusAttached,
						},
						Description: aws.String(description),
						Groups: []types.GroupIdentifier{
							{
								GroupId:   aws.String(groupID),
								GroupName: aws.String(groupName),
							},
						},
						InterfaceType: aws.String(interfaceType),
						Ipv6Addresses: []types.InstanceIpv6Address{
							{
								Ipv6Address: aws.String(ipv6Address),
							},
						},
						MacAddress:         aws.String(macAddress),
						NetworkInterfaceId: aws.String(natNetworkInterfaceID),
						OwnerId:            aws.String(ownerID),
						PrivateDnsName:     aws.String(privateDNSName),
						PrivateIpAddress:   aws.String(privateIPAddress),
						PrivateIpAddresses: []types.InstancePrivateIpAddress{
							{
								Association: &types.InstanceNetworkInterfaceAssociation{
									IpOwnerId:     aws.String(ipOwnerID),
									PublicDnsName: aws.String(publicDNSName),
									PublicIp:      aws.String(publicIPAddress),
								},
							},
						},
						SourceDestCheck: aws.Bool(false),
						Status:          types.NetworkInterfaceStatusAvailable,
						SubnetId:        aws.String(subnetID),
						VpcId:           aws.String(vpcID),
					},
				},
				OutpostArn: aws.String(outpostARN),
				Placement: &types.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostId:               aws.String(hostID),
					HostResourceGroupArn: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int32(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              types.TenancyHost,
				},
				Platform:         types.PlatformValuesWindows,
				PrivateDnsName:   aws.String(privateDNSName),
				PrivateIpAddress: aws.String(privateIPAddress),
				ProductCodes: []types.ProductCode{
					{
						ProductCodeId:   aws.String(productCodeID),
						ProductCodeType: types.ProductCodeValuesMarketplace,
					},
				},
				PublicDnsName:   aws.String(publicDNSName),
				PublicIpAddress: aws.String(publicIPAddress),
				RamdiskId:       aws.String(ramDiskID),
				RootDeviceName:  aws.String(rootDeviceName),
				RootDeviceType:  types.DeviceTypeEbs,
				SecurityGroups: []types.GroupIdentifier{
					{
						GroupId:   aws.String(groupID),
						GroupName: aws.String(groupName),
					},
				},
				SourceDestCheck:       aws.Bool(false),
				SpotInstanceRequestId: aws.String(spotInstanceReqID),
				SriovNetSupport:       aws.String(sriovNetSupport),
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
				StateReason: &types.StateReason{
					Message: aws.String(stateReason),
				},
				StateTransitionReason: aws.String(stateReason),
				SubnetId:              aws.String(subnetID),
				Tags: []types.Tag{
					{
						Key:   aws.String(tagsKey),
						Value: aws.String(tagsVal),
					},
				},
				VirtualizationType: types.VirtualizationTypeHvm,
				VpcId:              aws.String(vpcID),
			},
			attributes: ec2.DescribeInstanceAttributeOutput{
				DisableApiTermination:             &types.AttributeBooleanValue{Value: aws.Bool(true)},
				InstanceInitiatedShutdownBehavior: &types.AttributeValue{Value: aws.String(instanceInitiatedShutdownBehavior)},
				KernelId:                          &types.AttributeValue{Value: aws.String(kernelID)},
				RamdiskId:                         &types.AttributeValue{Value: aws.String(ramDiskID)},
				UserData:                          &types.AttributeValue{Value: aws.String(userData)},
			},
			out: manualv1alpha1.InstanceObservation{
				AmiLaunchIndex: aws.Int32(0),
				Architecture:   arch,
				BlockDeviceMapping: []manualv1alpha1.InstanceBlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						EBS: &manualv1alpha1.EBSInstanceBlockDevice{
							AttachTime:          nil,
							DeleteOnTermination: aws.Bool(false),
							Status:              string(types.AttachmentStatusAttached),
							VolumeID:            aws.String(volumeID),
						},
					},
				},
				CapacityReservationID: aws.String(capacityReservationID),
				CapacityReservationSpecification: &manualv1alpha1.CapacityReservationSpecificationResponse{
					CapacityReservationPreference: string(types.CapacityReservationPreferenceNone),
					CapacityReservationTarget: &manualv1alpha1.CapacityReservationTarget{
						CapacityReservationID: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CPUOptons: &manualv1alpha1.CPUOptionsRequest{
					CoreCount:      aws.Int32(1),
					ThreadsPerCore: aws.Int32(1),
				},
				DisableAPITermination: aws.Bool(true),
				EBSOptimized:          aws.Bool(false),
				EnaSupport:            aws.Bool(false),
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
				Hypervisor: string(types.HypervisorTypeOvm),
				IAMInstanceProfile: &manualv1alpha1.IAMInstanceProfile{
					ARN: aws.String(iamARN),
					ID:  aws.String(iamID),
				},
				ImageID:                           aws.String(imageID),
				InstanceID:                        aws.String(instanceID),
				InstanceInitiatedShutdownBehavior: aws.String(instanceInitiatedShutdownBehavior),
				InstanceLifecycle:                 string(types.InstanceLifecycleTypeScheduled),
				InstanceType:                      string(types.InstanceTypeM1Small),
				KernelID:                          aws.String(kernelID),
				Licenses: []manualv1alpha1.LicenseConfigurationRequest{
					{
						LicenseConfigurationARN: aws.String(licenseConfig),
					},
				},
				MetadataOptions: &manualv1alpha1.InstanceMetadataOptionsRequest{
					HTTPEndpoint:            string(types.InstanceMetadataEndpointStateEnabled),
					HTTPPutResponseHopLimit: aws.Int32(0),
					HTTPTokens:              string(types.HttpTokensStateOptional),
				},
				Monitoring: &manualv1alpha1.Monitoring{
					State: string(types.MonitoringStateEnabled),
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
							DeviceIndex:         aws.Int32(0),
							Status:              string(types.AttachmentStatusAttached),
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
						Status:          string(types.NetworkInterfaceStatusAvailable),
						SubnetID:        aws.String(subnetID),
						VPCID:           aws.String(vpcID),
					},
				},
				OutpostARN: aws.String(outpostARN),
				Platform:   string(types.PlatformValuesWindows),
				Placement: &manualv1alpha1.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostID:               aws.String(hostID),
					HostResourceGroupARN: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int32(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              string(types.TenancyHost),
				},
				PrivateDNSName:   aws.String(privateDNSName),
				PrivateIPAddress: aws.String(privateIPAddress),
				ProductCodes: []manualv1alpha1.ProductCode{
					{
						ProductCodeID:   aws.String(productCodeID),
						ProductCodeType: string(types.ProductCodeValuesMarketplace),
					},
				},
				PublicDNSName:   aws.String(publicDNSName),
				PublicIPAddress: aws.String(publicIPAddress),
				RAMDiskID:       aws.String(ramDiskID),
				RootDeviceName:  aws.String(rootDeviceName),
				RootDeviceType:  string(types.DeviceTypeEbs),
				SecurityGroups: []manualv1alpha1.GroupIdentifier{
					{
						GroupID:   groupID,
						GroupName: groupName,
					},
				},
				SourceDestCheck:       aws.Bool(false),
				SpotInstanceRequestID: aws.String(spotInstanceReqID),
				SriovNetSupport:       aws.String(sriovNetSupport),
				State:                 string(types.InstanceStateNameRunning),
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
				UserData:           aws.String(userData),
				VirtualizationType: string(types.VirtualizationTypeHvm),
				VPCID:              aws.String(vpcID),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateInstanceObservation(tc.in, &tc.attributes)
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
				MaxCount: aws.Int32(1),
				MinCount: aws.Int32(1),
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
							IOps:                aws.Int32(1),
							Throughput:          aws.Int32(100),
							KmsKeyID:            aws.String(keyName),
							SnapshotID:          aws.String(snapshotID),
							VolumeSize:          aws.Int32(1),
							VolumeType:          volumeType,
						},
					},
				},
				CapacityReservationSpecification: &manualv1alpha1.CapacityReservationSpecification{
					CapacityReservationPreference: string(types.CapacityReservationPreferenceNone),
					CapacityReservationTarget: &manualv1alpha1.CapacityReservationTarget{
						CapacityReservationID: aws.String(capacityReservationID),
					},
				},
				ClientToken: aws.String(clientToken),
				CPUOptions: &manualv1alpha1.CPUOptionsRequest{
					CoreCount:      aws.Int32(1),
					ThreadsPerCore: aws.Int32(1),
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
						Count: aws.Int32(1),
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
				InstanceInitiatedShutdownBehavior: string(types.ShutdownBehaviorStop),
				InstanceMarketOptions: &manualv1alpha1.InstanceMarketOptionsRequest{
					MarketType: spotMarketType,
					SpotOptions: &manualv1alpha1.SpotMarketOptions{
						BlockDurationMinutes:         aws.Int32(1),
						InstanceInterruptionBehavior: string(types.InstanceInterruptionBehaviorHibernate),
						MaxPrice:                     aws.String("1"),
						SpotInstanceType:             string(types.SpotInstanceTypeOneTime),
					},
				},
				InstanceType:     string(types.InstanceTypeA12xlarge),
				IPv6AddressCount: aws.Int32(1),
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
					HTTPEndpoint:            string(types.InstanceMetadataEndpointStateEnabled),
					HTTPPutResponseHopLimit: aws.Int32(0),
					HTTPTokens:              string(types.HttpTokensStateOptional),
				},
				Monitoring: &manualv1alpha1.RunInstancesMonitoringEnabled{
					Enabled: aws.Bool(false),
				},
				NetworkInterfaces: []manualv1alpha1.InstanceNetworkInterfaceSpecification{
					{
						AssociatePublicIPAddress: aws.Bool(false),
						DeleteOnTermination:      aws.Bool(false),
						Description:              aws.String(description),
						DeviceIndex:              aws.Int32(0),
						Groups: []string{
							groupID,
						},
						InterfaceType:    aws.String(interfaceType),
						IPv6AddressCount: aws.Int32(1),
						IPv6Addresses: []manualv1alpha1.InstanceIPv6Address{
							{
								IPv6Address: aws.String(ipv6Address),
							},
						},
						Ipv6PrefixCount: aws.Int32(1),
						Ipv6Prefixes: []manualv1alpha1.Ipv6PrefixSpecificationRequest{
							{
								Ipv6Prefix: ipv6Prefix,
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
						SecondaryPrivateIPAddressCount: aws.Int32(0),
						SubnetID:                       aws.String(subnetID),
					},
				},
				Placement: &manualv1alpha1.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostID:               aws.String(hostID),
					HostResourceGroupARN: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int32(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              string(types.TenancyHost),
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
				BlockDeviceMappings: []types.BlockDeviceMapping{
					{
						DeviceName: aws.String(blockDeviceName),
						Ebs: &types.EbsBlockDevice{
							DeleteOnTermination: aws.Bool(false),
							Encrypted:           aws.Bool(false),
							Iops:                aws.Int32(1),
							Throughput:          aws.Int32(100),
							KmsKeyId:            aws.String(keyName),
							SnapshotId:          aws.String(snapshotID),
							VolumeSize:          aws.Int32(1),
							VolumeType:          volumeType,
						},
					},
				},
				CapacityReservationSpecification: &types.CapacityReservationSpecification{
					CapacityReservationPreference: types.CapacityReservationPreferenceNone,
					CapacityReservationTarget: &types.CapacityReservationTarget{
						CapacityReservationId: aws.String(capacityReservationID),
					},
				},
				CpuOptions: &types.CpuOptionsRequest{
					CoreCount:      aws.Int32(1),
					ThreadsPerCore: aws.Int32(1),
				},
				CreditSpecification: &types.CreditSpecificationRequest{
					CpuCredits: aws.String(cpuCredits),
				},
				ClientToken:           aws.String(clientToken),
				DisableApiTermination: aws.Bool(false),
				EbsOptimized:          aws.Bool(false),
				ElasticGpuSpecification: []types.ElasticGpuSpecification{
					{
						Type: aws.String(gpuType),
					},
				},
				ElasticInferenceAccelerators: []types.ElasticInferenceAccelerator{
					{
						Count: aws.Int32(1),
						Type:  aws.String(gpuType),
					},
				},
				HibernationOptions: &types.HibernationOptionsRequest{
					Configured: aws.Bool(false),
				},
				IamInstanceProfile: &types.IamInstanceProfileSpecification{
					Arn:  aws.String(iamARN),
					Name: aws.String(iamID),
				},
				ImageId: aws.String(imageID),
				InstanceMarketOptions: &types.InstanceMarketOptionsRequest{
					MarketType: spotMarketType,
					SpotOptions: &types.SpotMarketOptions{
						BlockDurationMinutes:         aws.Int32(1),
						InstanceInterruptionBehavior: types.InstanceInterruptionBehaviorHibernate,
						MaxPrice:                     aws.String("1"),
						SpotInstanceType:             types.SpotInstanceTypeOneTime,
					},
				},
				InstanceType:                      types.InstanceTypeA12xlarge,
				InstanceInitiatedShutdownBehavior: types.ShutdownBehaviorStop,
				Ipv6AddressCount:                  aws.Int32(1),
				Ipv6Addresses: []types.InstanceIpv6Address{
					{
						Ipv6Address: aws.String(ipv6Address),
					},
				},
				KernelId: aws.String(kernelID),
				KeyName:  aws.String(keyName),
				LaunchTemplate: &types.LaunchTemplateSpecification{
					LaunchTemplateId:   aws.String(launchTemplateID),
					LaunchTemplateName: aws.String(launchTemplateName),
					Version:            aws.String("1"),
				},
				LicenseSpecifications: []types.LicenseConfigurationRequest{
					{
						LicenseConfigurationArn: aws.String(licenseConfig),
					},
				},
				MinCount: aws.Int32(1),
				MaxCount: aws.Int32(1),
				MetadataOptions: &types.InstanceMetadataOptionsRequest{
					HttpEndpoint:            types.InstanceMetadataEndpointStateEnabled,
					HttpPutResponseHopLimit: aws.Int32(0),
					HttpTokens:              types.HttpTokensStateOptional,
				},
				Monitoring: &types.RunInstancesMonitoringEnabled{
					Enabled: aws.Bool(false),
				},
				NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
					{
						AssociatePublicIpAddress: aws.Bool(false),
						DeleteOnTermination:      aws.Bool(false),
						Description:              aws.String(description),
						DeviceIndex:              aws.Int32(0),
						Groups: []string{
							groupID,
						},
						InterfaceType:    aws.String(interfaceType),
						Ipv6AddressCount: aws.Int32(1),
						Ipv6Addresses: []types.InstanceIpv6Address{
							{
								Ipv6Address: aws.String(ipv6Address),
							},
						},
						Ipv6PrefixCount: aws.Int32(1),
						Ipv6Prefixes: []types.Ipv6PrefixSpecificationRequest{{
							Ipv6Prefix: aws.String(ipv6Prefix),
						}},
						NetworkInterfaceId: aws.String(networkInterfaceID),
						PrivateIpAddress:   aws.String(privateIPAddress),
						PrivateIpAddresses: []types.PrivateIpAddressSpecification{
							{
								Primary:          aws.Bool(false),
								PrivateIpAddress: aws.String(privateIPAddress),
							},
						},
						SecondaryPrivateIpAddressCount: aws.Int32(0),
						SubnetId:                       aws.String(subnetID),
					},
				},
				Placement: &types.Placement{
					Affinity:             aws.String(placementAff),
					GroupName:            aws.String(groupName),
					HostId:               aws.String(hostID),
					HostResourceGroupArn: aws.String(hostGroupARN),
					PartitionNumber:      aws.Int32(0),
					SpreadDomain:         aws.String(spreadDomain),
					Tenancy:              types.TenancyHost,
				},
				PrivateIpAddress: aws.String(privateIPAddress),
				RamdiskId:        aws.String(ramDiskID),
				SecurityGroupIds: []string{
					groupID,
				},
				SubnetId: aws.String(subnetID),
				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceTypeInstance,
						Tags: []types.Tag{
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
			if diff := cmp.Diff(tc.out, r, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateEC2RunInstancesInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}
