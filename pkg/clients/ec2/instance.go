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
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// InstanceNotFound is the code that is returned by ec2 when the given InstanceID is not valid
	InstanceNotFound = "InvalidInstanceID.NotFound"
)

// InstanceClient is the external client used for Instance Custom Resource
type InstanceClient interface {
	RunInstancesRequest(*ec2.RunInstancesInput) ec2.RunInstancesRequest
	TerminateInstancesRequest(*ec2.TerminateInstancesInput) ec2.TerminateInstancesRequest
	DescribeInstancesRequest(*ec2.DescribeInstancesInput) ec2.DescribeInstancesRequest
	DescribeInstanceAttributeRequest(*ec2.DescribeInstanceAttributeInput) ec2.DescribeInstanceAttributeRequest
	ModifyInstanceAttributeRequest(*ec2.ModifyInstanceAttributeInput) ec2.ModifyInstanceAttributeRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewInstanceClient returns a new client using AWS credentials as JSON encoded data.
func NewInstanceClient(cfg aws.Config) InstanceClient {
	return ec2.New(cfg)
}

// IsInstanceNotFoundErr returns true if the error is because the item doesn't exist
func IsInstanceNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == InstanceNotFound {
			return true
		}
	}

	return false
}

// IsInstanceUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsInstanceUpToDate(spec manualv1alpha1.InstanceParameters, instance ec2.Instance, attributes ec2.DescribeInstanceAttributeOutput) bool {
	// DisableApiTermination
	if aws.BoolValue(spec.DisableAPITermination) != attributeBoolValue(attributes.DisableApiTermination) {
		return false
	}
	// InstanceInitiatedShutdownBehavior
	if spec.InstanceInitiatedShutdownBehavior != attributeValue(attributes.InstanceInitiatedShutdownBehavior) {
		return false
	}
	// KernalID
	if aws.StringValue(spec.KernelID) != aws.StringValue(instance.KernelId) {
		return false
	}
	// RamDiskID
	if aws.StringValue(spec.RAMDiskID) != aws.StringValue(instance.RamdiskId) {
		return false
	}
	// UserData
	if aws.StringValue(spec.UserData) != attributeValue(attributes.UserData) {
		return false
	}
	return manualv1alpha1.CompareGroupIDs(spec.SecurityGroupIDs, instance.SecurityGroups)
}

// GenerateInstanceObservation is used to produce manualv1alpha1.InstanceObservation from
// a []ec2.Instance.
func GenerateInstanceObservation(i ec2.Instance) manualv1alpha1.InstanceObservation {
	return manualv1alpha1.InstanceObservation{
		AmiLaunchIndex:                          i.AmiLaunchIndex,
		Architecture:                            string(i.Architecture),
		BlockDeviceMapping:                      GenerateInstanceBlockDeviceMappings(i.BlockDeviceMappings),
		CapacityReservationID:                   i.CapacityReservationId,
		CapacityReservationSpecification:        GenerateCapacityReservationSpecResponse(i.CapacityReservationSpecification),
		ClientToken:                             i.ClientToken,
		CPUOptons:                               GenerateCPUOptionsRequest(i.CpuOptions),
		EBSOptimized:                            i.EbsOptimized,
		ElasticGPUAssociations:                  GenerateElasticGPUAssociation(i.ElasticGpuAssociations),
		ElasticInferenceAcceleratorAssociations: GenerateElasticInferenceAcceleratorAssociation(i.ElasticInferenceAcceleratorAssociations),
		EnaSupport:                              i.EnaSupport,
		HibernationOptions:                      GenerateHibernationOptionsRequest(i.HibernationOptions),
		Hypervisor:                              string(i.Hypervisor),
		IAMInstanceProfile:                      GenerateIAMInstanceProfile(i.IamInstanceProfile),
		ImageID:                                 i.ImageId,
		InstanceID:                              i.InstanceId,
		InstanceLifecycle:                       string(i.InstanceLifecycle),
		InstanceType:                            string(i.InstanceType),
		KernelID:                                i.KernelId,
		LaunchTime:                              FromTimePtr(i.LaunchTime),
		Licenses:                                GenerateLicenseConfigurationRequest(i.Licenses),
		MetadataOptions:                         GenerateInstanceMetadataOptionsRequest(i.MetadataOptions),
		Monitoring:                              GenerateMonitoring(i.Monitoring),
		NetworkInterfaces:                       GenerateInstanceNetworkInterface(i.NetworkInterfaces),
		OutpostARN:                              i.OutpostArn,
		Placement:                               GeneratePlacement(i.Placement),
		Platform:                                string(i.Platform),
		PrivateDNSName:                          i.PrivateDnsName,
		PrivateIPAddress:                        i.PrivateIpAddress,
		ProductCodes:                            GenerateProductCodes(i.ProductCodes),
		PublicDNSName:                           i.PublicDnsName,
		PublicIPAddress:                         i.PublicIpAddress,
		RAMDiskID:                               i.RamdiskId,
		RootDeviceName:                          i.RootDeviceName,
		RootDeviceType:                          string(i.RootDeviceType),
		SecurityGroups:                          GenerateGroupIdentifiers(i.SecurityGroups),
		SourceDestCheck:                         i.SourceDestCheck,
		SpotInstanceRequestID:                   i.SpotInstanceRequestId,
		SriovNetSupport:                         i.SriovNetSupport,
		State:                                   string(i.State.Name),
		StateReason:                             GenerateStateReason(i.StateReason),
		StateTransitionReason:                   i.StateTransitionReason,
		SubnetID:                                i.SubnetId,
		Tags:                                    GenerateTags(i.Tags),
		VirtualizationType:                      string(i.VirtualizationType),
		VPCID:                                   i.VpcId,
	}
}

// GenerateInstanceCondition returns an instance Condition depending on the supplied
// observation. Currently, the instance can be denoted as:
// * Available
// * Creating
// * Deleting
func GenerateInstanceCondition(o manualv1alpha1.InstanceObservation) Condition {
	switch o.State {
	case string(ec2.InstanceStateNameRunning):
		return Available
	case string(ec2.InstanceStateNameShuttingDown):
		return Deleting
	case string(ec2.InstanceStateNameStopped):
		return Deleting
	case string(ec2.InstanceStateNameStopping):
		return Deleting
	case string(ec2.InstanceStateNameTerminated):
		return Deleted
	default:
		// ec2.InstanceStateNamePending
		return Creating
	}
}

// Condition denotes the current state across instances
type Condition string

const (
	// Available is the condition that represents all instances are running
	Available Condition = "available"
	// Creating is the condition that represents some instances could be
	// running, but the rest are pending
	Creating Condition = "creating"
	// Deleting is the condition that represents some instances have entered
	// the shutdown/termination state
	Deleting Condition = "deleting"
	// Deleted is the condition that represents all instances have entered
	// the terminated state
	Deleted Condition = "deleted"
)

// LateInitializeInstance fills the empty fields in *manualv1alpha1.InstanceParameters with
// the values seen in ec2.Instance and ec2.DescribeInstanceAttributeOutput.
func LateInitializeInstance(in *manualv1alpha1.InstanceParameters, instance *ec2.Instance, attributes *ec2.DescribeInstanceAttributeOutput) { // nolint:gocyclo
	if instance == nil {
		return
	}

	if attributes.DisableApiTermination != nil {
		in.DisableAPITermination = awsclients.LateInitializeBoolPtr(in.DisableAPITermination, attributes.DisableApiTermination.Value)
	}

	if attributes.InstanceInitiatedShutdownBehavior != nil {
		in.InstanceInitiatedShutdownBehavior = awsclients.LateInitializeString(in.InstanceInitiatedShutdownBehavior, attributes.InstanceInitiatedShutdownBehavior.Value)
	}

	if attributes.InstanceType != nil {
		in.InstanceType = awsclients.LateInitializeString(in.InstanceType, attributes.InstanceType.Value)
	}

	if attributes.UserData != nil {
		in.UserData = awsclients.LateInitializeStringPtr(in.UserData, attributes.UserData.Value)
	}

	in.EBSOptimized = awsclients.LateInitializeBoolPtr(in.EBSOptimized, instance.EbsOptimized)
	in.KernelID = awsclients.LateInitializeStringPtr(in.KernelID, instance.KernelId)
	in.RAMDiskID = awsclients.LateInitializeStringPtr(in.RAMDiskID, instance.RamdiskId)

	if len(in.SecurityGroupIDs) == 0 && len(instance.SecurityGroups) != 0 {
		in.SecurityGroupIDs = make([]string, len(instance.SecurityGroups))
		for i, s := range instance.SecurityGroups {
			in.SecurityGroupIDs[i] = *s.GroupId
		}
	}

	if in.SubnetID == nil || *in.SubnetID == "" && instance.SubnetId != nil {
		in.SubnetID = instance.SubnetId
	}
}

// GenerateEC2BlockDeviceMappings coverts an internal slice of BlockDeviceMapping into a slice of ec2.BlockDeviceMapping
func GenerateEC2BlockDeviceMappings(mappings []manualv1alpha1.BlockDeviceMapping) []ec2.BlockDeviceMapping {
	if mappings != nil {
		res := make([]ec2.BlockDeviceMapping, len(mappings))
		for i, bm := range mappings {
			res[i] = ec2.BlockDeviceMapping{
				DeviceName: bm.DeviceName,
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: bm.EBS.DeleteOnTermination,
					Encrypted:           bm.EBS.Encrypted,
					Iops:                bm.EBS.IOps,
					KmsKeyId:            bm.EBS.KmsKeyID,
					SnapshotId:          bm.EBS.SnapshotID,
					VolumeSize:          bm.EBS.VolumeSize,
					VolumeType:          ec2.VolumeType(bm.EBS.VolumeType),
				},
				NoDevice:    bm.NoDevice,
				VirtualName: bm.VirtualName,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2CapacityReservationSpecs coverts an internal CapacityReservationSpecification into a ec2.CapacityReservationSpecification
func GenerateEC2CapacityReservationSpecs(spec *manualv1alpha1.CapacityReservationSpecification) *ec2.CapacityReservationSpecification {
	if spec != nil {
		var capacityReservationID *string
		if spec.CapacityReservationTarget != nil {
			capacityReservationID = spec.CapacityReservationTarget.CapacityReservationID
		}

		return &ec2.CapacityReservationSpecification{
			CapacityReservationPreference: ec2.CapacityReservationPreference(spec.CapacityReservationPreference),
			CapacityReservationTarget: &ec2.CapacityReservationTarget{
				CapacityReservationId: capacityReservationID,
			},
		}
	}
	return nil
}

// GenerateCapacityReservationSpecResponse converts a ec2.CapacityReservationSpecificationResponse into an internal CapacityReservationSpecificationResponse
func GenerateCapacityReservationSpecResponse(resp *ec2.CapacityReservationSpecificationResponse) *manualv1alpha1.CapacityReservationSpecificationResponse {
	if resp != nil {
		var target manualv1alpha1.CapacityReservationTarget

		if resp.CapacityReservationTarget != nil {
			target.CapacityReservationID = resp.CapacityReservationTarget.CapacityReservationId
		}

		return &manualv1alpha1.CapacityReservationSpecificationResponse{
			CapacityReservationPreference: string(resp.CapacityReservationPreference),
			CapacityReservationTarget:     &target,
		}
	}
	return nil
}

// GenerateEC2CPUOptions converts an internal CPUOptionsRequest into a ec2.CpuOptionsRequest
func GenerateEC2CPUOptions(opts *manualv1alpha1.CPUOptionsRequest) *ec2.CpuOptionsRequest {
	if opts != nil {
		return &ec2.CpuOptionsRequest{
			CoreCount:      opts.CoreCount,
			ThreadsPerCore: opts.ThreadsPerCore,
		}
	}
	return nil
}

// GenerateCPUOptionsRequest converts a CpuOptions into a internal CpuOptionsRequest
func GenerateCPUOptionsRequest(opts *ec2.CpuOptions) *manualv1alpha1.CPUOptionsRequest {
	if opts != nil {
		return &manualv1alpha1.CPUOptionsRequest{
			CoreCount:      opts.CoreCount,
			ThreadsPerCore: opts.ThreadsPerCore,
		}
	}
	return nil
}

// GenerateEC2CreditSpec converts an internal CreditSpecificationRequest into a ec2.CreditSpecificationRequest
func GenerateEC2CreditSpec(spec *manualv1alpha1.CreditSpecificationRequest) *ec2.CreditSpecificationRequest {
	if spec != nil {
		return &ec2.CreditSpecificationRequest{
			CpuCredits: spec.CPUCredits,
		}
	}
	return nil
}

// GenerateElasticGPUAssociation coverts a slice of ec2.ElasticGpuAssociation into an internal slice of ElasticGPUAssociation
func GenerateElasticGPUAssociation(assocs []ec2.ElasticGpuAssociation) []manualv1alpha1.ElasticGPUAssociation {
	if assocs != nil {
		res := make([]manualv1alpha1.ElasticGPUAssociation, len(assocs))
		for i, a := range assocs {
			res[i] = manualv1alpha1.ElasticGPUAssociation{
				ElasticGPUAssociationID:    a.ElasticGpuAssociationId,
				ElasticGPUAssociationState: a.ElasticGpuAssociationState,
				ElasticGPUAssociationTime:  a.ElasticGpuAssociationTime,
				ElasticGPUID:               a.ElasticGpuId,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2ElasticGPUSpecs coverts an internal slice of ElasticGPUSpecification into a slice of ec2.ElasticGpuSpecification
func GenerateEC2ElasticGPUSpecs(specs []manualv1alpha1.ElasticGPUSpecification) []ec2.ElasticGpuSpecification {
	if specs != nil {
		res := make([]ec2.ElasticGpuSpecification, len(specs))
		for i, gs := range specs {
			res[i] = ec2.ElasticGpuSpecification{
				Type: gs.Type,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2ElasticInferenceAccelerators coverts an internal slice of ElasticInferenceAccelerator into a slice of ec2.ElasticInferenceAccelerator
func GenerateEC2ElasticInferenceAccelerators(accs []manualv1alpha1.ElasticInferenceAccelerator) []ec2.ElasticInferenceAccelerator {
	if accs != nil {
		res := make([]ec2.ElasticInferenceAccelerator, len(accs))
		for i, a := range accs {
			res[i] = ec2.ElasticInferenceAccelerator{
				Count: a.Count,
				Type:  a.Type,
			}
		}

		return res
	}
	return nil
}

// GenerateElasticInferenceAcceleratorAssociation coverts a slice of ec2.ElasticInferenceAcceleratorAssociation into an internal slice of ElasticInferenceAcceleratorAssociation
func GenerateElasticInferenceAcceleratorAssociation(assocs []ec2.ElasticInferenceAcceleratorAssociation) []manualv1alpha1.ElasticInferenceAcceleratorAssociation {
	if assocs != nil {
		res := make([]manualv1alpha1.ElasticInferenceAcceleratorAssociation, len(assocs))
		for i, a := range assocs {
			res[i] = manualv1alpha1.ElasticInferenceAcceleratorAssociation{
				ElasticInferenceAcceleratorARN:              a.ElasticInferenceAcceleratorArn,
				ElasticInferenceAcceleratorAssociationID:    a.ElasticInferenceAcceleratorAssociationId,
				ElasticInferenceAcceleratorAssociationState: a.ElasticInferenceAcceleratorAssociationState,
				ElasticInferenceAcceleratorAssociationTime:  FromTimePtr(a.ElasticInferenceAcceleratorAssociationTime),
			}
		}

		return res
	}
	return nil
}

// GenerateEC2HibernationOptions converts an internal HibernationOptionsRequest into a ec2.HibernationOptionsRequest
func GenerateEC2HibernationOptions(opts *manualv1alpha1.HibernationOptionsRequest) *ec2.HibernationOptionsRequest {
	if opts != nil {
		return &ec2.HibernationOptionsRequest{
			Configured: opts.Configured,
		}
	}
	return nil
}

// GenerateGroupIdentifiers coverts a slice of ec2.GroupIdentifier into an internal slice of GroupIdentifier
func GenerateGroupIdentifiers(ids []ec2.GroupIdentifier) []manualv1alpha1.GroupIdentifier {
	if ids != nil {
		res := make([]manualv1alpha1.GroupIdentifier, len(ids))
		for i, id := range ids {
			res[i] = manualv1alpha1.GroupIdentifier{
				GroupID:   *id.GroupId,
				GroupName: *id.GroupName,
			}
		}

		return res
	}
	return nil
}

// GenerateIAMInstanceProfile converts a ec2.IamInstanceProfile into a internal IamInstanceProfile
func GenerateIAMInstanceProfile(p *ec2.IamInstanceProfile) *manualv1alpha1.IAMInstanceProfile {
	if p != nil {
		return &manualv1alpha1.IAMInstanceProfile{
			ARN: p.Arn,
			ID:  p.Id,
		}
	}
	return nil
}

// GenerateEC2IAMInstanceProfileSpecification converts an internal IamInstanceProfileSpecification into a ec2.IamInstanceProfileSpecification
func GenerateEC2IAMInstanceProfileSpecification(spec *manualv1alpha1.IAMInstanceProfileSpecification) *ec2.IamInstanceProfileSpecification {
	if spec != nil {
		return &ec2.IamInstanceProfileSpecification{
			Arn:  spec.ARN,
			Name: spec.Name,
		}
	}
	return nil
}

// GenerateInstanceBlockDeviceMappings coverts a slice of ec2.InstanceBlockDeviceMapping into an internal slice of InstanceBlockDeviceMapping
func GenerateInstanceBlockDeviceMappings(mappings []ec2.InstanceBlockDeviceMapping) []manualv1alpha1.InstanceBlockDeviceMapping {
	if mappings != nil {
		res := make([]manualv1alpha1.InstanceBlockDeviceMapping, len(mappings))
		for i, m := range mappings {
			res[i] = manualv1alpha1.InstanceBlockDeviceMapping{
				DeviceName: m.DeviceName,
				EBS: &manualv1alpha1.EBSInstanceBlockDevice{
					AttachTime:          FromTimePtr(m.Ebs.AttachTime),
					DeleteOnTermination: m.Ebs.DeleteOnTermination,
					Status:              string(m.Ebs.Status),
					VolumeID:            m.Ebs.VolumeId,
				},
			}
		}

		return res
	}
	return nil
}

// GenerateEC2InstanceMarketOptionsRequest converts an internal InstanceMarketOptionsRequest into a ec2.InstanceMarketOptionsRequest
func GenerateEC2InstanceMarketOptionsRequest(opts *manualv1alpha1.InstanceMarketOptionsRequest) *ec2.InstanceMarketOptionsRequest {
	if opts != nil {
		var durationMin *int64
		var behavior ec2.InstanceInterruptionBehavior
		var maxPrice *string
		var instanceType ec2.SpotInstanceType
		var validUntil *time.Time

		if opts.SpotOptions != nil {
			durationMin = opts.SpotOptions.BlockDurationMinutes
			behavior = ec2.InstanceInterruptionBehavior(opts.SpotOptions.InstanceInterruptionBehavior)
			maxPrice = opts.SpotOptions.MaxPrice
			instanceType = ec2.SpotInstanceType(opts.SpotOptions.SpotInstanceType)
			if opts.SpotOptions.ValidUntil != nil {
				validUntil = &opts.SpotOptions.ValidUntil.DeepCopy().Time
			}
		}

		return &ec2.InstanceMarketOptionsRequest{
			MarketType: ec2.MarketType(opts.MarketType),
			SpotOptions: &ec2.SpotMarketOptions{
				BlockDurationMinutes:         durationMin,
				InstanceInterruptionBehavior: behavior,
				MaxPrice:                     maxPrice,
				SpotInstanceType:             instanceType,
				ValidUntil:                   validUntil,
			},
		}
	}
	return nil
}

// GenerateInstanceMetadataOptionsRequest converts an ec2.InstanceMetadataOptionsResponse into an internal InstanceMetadataOptionsRequest
func GenerateInstanceMetadataOptionsRequest(opts *ec2.InstanceMetadataOptionsResponse) *manualv1alpha1.InstanceMetadataOptionsRequest {
	if opts != nil {
		return &manualv1alpha1.InstanceMetadataOptionsRequest{
			HTTPEndpoint:            string(opts.HttpEndpoint),
			HTTPPutResponseHopLimit: opts.HttpPutResponseHopLimit,
			HTTPTokens:              string(opts.HttpTokens),
		}
	}
	return nil
}

// GenerateEC2InstanceIPV6Addresses coverts an internal slice of InstanceIPV6Address into a slice of ec2.InstanceIpv6Address
func GenerateEC2InstanceIPV6Addresses(addrs []manualv1alpha1.InstanceIPv6Address) []ec2.InstanceIpv6Address {
	if addrs != nil {
		res := make([]ec2.InstanceIpv6Address, len(addrs))
		for i, a := range addrs {
			res[i] = ec2.InstanceIpv6Address{
				Ipv6Address: a.IPv6Address,
			}
		}

		return res
	}
	return nil
}

// GenerateInstanceIPV6Addresses coverts a slice of ec2.InstanceIpv6Address into a slice of internal InstanceIPv6Address
func GenerateInstanceIPV6Addresses(addrs []ec2.InstanceIpv6Address) []manualv1alpha1.InstanceIPv6Address {
	if addrs != nil {
		res := make([]manualv1alpha1.InstanceIPv6Address, len(addrs))
		for i, a := range addrs {
			res[i] = manualv1alpha1.InstanceIPv6Address{
				IPv6Address: a.Ipv6Address,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2InstanceMetadataOptionsRequest converts an internal InstanceMetadataOptionsRequest into a ec2.InstanceMetadataOptionsRequest
func GenerateEC2InstanceMetadataOptionsRequest(opts *manualv1alpha1.InstanceMetadataOptionsRequest) *ec2.InstanceMetadataOptionsRequest {
	if opts != nil {
		return &ec2.InstanceMetadataOptionsRequest{
			HttpEndpoint:            ec2.InstanceMetadataEndpointState(opts.HTTPEndpoint),
			HttpPutResponseHopLimit: opts.HTTPPutResponseHopLimit,
			HttpTokens:              ec2.HttpTokensState(opts.HTTPTokens),
		}
	}
	return nil
}

// GenerateInstanceNetworkInterface coverts a slice of ec2.InstanceNetworkInterface into an internal slice of InstanceNetworkInterface
func GenerateInstanceNetworkInterface(nets []ec2.InstanceNetworkInterface) []manualv1alpha1.InstanceNetworkInterface {
	if nets != nil {
		res := make([]manualv1alpha1.InstanceNetworkInterface, len(nets))
		for i, intr := range nets {
			var association manualv1alpha1.InstanceNetworkInterfaceAssociation
			var attachment manualv1alpha1.InstanceNetworkInterfaceAttachment

			if intr.Association != nil {
				association.IPOwnerID = intr.Association.IpOwnerId
				association.PublicDNSName = intr.Association.PublicDnsName
				association.PublicIP = intr.Association.PublicIp
			}

			if intr.Attachment != nil {
				attachment.AttachTime = FromTimePtr(intr.Attachment.AttachTime)
				attachment.AttachmentID = intr.Attachment.AttachmentId
				attachment.DeleteOnTermination = intr.Attachment.DeleteOnTermination
				attachment.DeviceIndex = intr.Attachment.DeviceIndex
				attachment.Status = string(intr.Attachment.Status)
			}

			res[i] = manualv1alpha1.InstanceNetworkInterface{
				Association:        &association,
				Attachment:         &attachment,
				Description:        intr.Description,
				Groups:             GenerateGroupIdentifiers(intr.Groups),
				InterfaceType:      intr.InterfaceType,
				IPv6Addresses:      GenerateInstanceIPV6Addresses(intr.Ipv6Addresses),
				MacAddress:         intr.MacAddress,
				NetworkInterfaceID: intr.NetworkInterfaceId,
				OwnerID:            intr.OwnerId,
				PrivateDNSName:     intr.PrivateDnsName,
				PrivateIPAddress:   intr.PrivateIpAddress,
				PrivateIPAddresses: GenerateInstancePrivateIPAddresses(intr.PrivateIpAddresses),
				SourceDestCheck:    intr.SourceDestCheck,
				Status:             string(intr.Status),
				SubnetID:           intr.SubnetId,
				VPCID:              intr.VpcId,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2InstanceNetworkInterfaceSpecs coverts an internal slice of InstanceNetworkInterfaceSpecification
// into a slice of ec2.InstanceNetworkInterfaceSpecification
func GenerateEC2InstanceNetworkInterfaceSpecs(specs []manualv1alpha1.InstanceNetworkInterfaceSpecification) []ec2.InstanceNetworkInterfaceSpecification {
	if specs != nil {
		res := make([]ec2.InstanceNetworkInterfaceSpecification, len(specs))
		for i, s := range specs {
			res[i] = ec2.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress:       s.AssociatePublicIPAddress,
				DeleteOnTermination:            s.DeleteOnTermination,
				Description:                    s.Description,
				DeviceIndex:                    s.DeviceIndex,
				Groups:                         s.Groups,
				InterfaceType:                  s.InterfaceType,
				Ipv6AddressCount:               s.IPv6AddressCount,
				Ipv6Addresses:                  GenerateEC2InstanceIPV6Addresses(s.IPv6Addresses),
				NetworkInterfaceId:             s.NetworkInterfaceID,
				PrivateIpAddress:               s.PrivateIPAddress,
				PrivateIpAddresses:             GenerateEC2PrivateIPAddressSpecs(s.PrivateIPAddresses),
				SecondaryPrivateIpAddressCount: s.SecondaryPrivateIPAddressCount,
				SubnetId:                       s.SubnetID,
			}
		}

		return res
	}
	return nil
}

// GenerateInstancePrivateIPAddresses coverts a slice of ec2.InstanceIpv6Address into a slice of internal InstanceIPv6Address
func GenerateInstancePrivateIPAddresses(addrs []ec2.InstancePrivateIpAddress) []manualv1alpha1.InstancePrivateIPAddress {
	if addrs != nil {
		res := make([]manualv1alpha1.InstancePrivateIPAddress, len(addrs))
		for i, a := range addrs {
			var association manualv1alpha1.InstanceNetworkInterfaceAssociation
			if a.Association != nil {
				association.IPOwnerID = a.Association.IpOwnerId
				association.PublicDNSName = a.Association.PublicDnsName
				association.PublicIP = a.Association.PublicIp
			}

			res[i] = manualv1alpha1.InstancePrivateIPAddress{
				Association:      &association,
				Primary:          a.Primary,
				PrivateDNSName:   a.PrivateDnsName,
				PrivateIPAddress: a.PrivateIpAddress,
			}
		}

		return res
	}
	return nil
}

// GenerateHibernationOptionsRequest converts a ec2.HibernationOptions into a internal HibernationOptionsRequest
func GenerateHibernationOptionsRequest(opts *ec2.HibernationOptions) *manualv1alpha1.HibernationOptionsRequest {
	if opts != nil {
		return &manualv1alpha1.HibernationOptionsRequest{
			Configured: opts.Configured,
		}
	}
	return nil
}

// GenerateEC2LaunchTemplateSpec converts internal LaunchTemplateSpecification into ec2.LaunchTemplateSpecification
func GenerateEC2LaunchTemplateSpec(spec *manualv1alpha1.LaunchTemplateSpecification) *ec2.LaunchTemplateSpecification {
	if spec != nil {
		return &ec2.LaunchTemplateSpecification{
			LaunchTemplateId:   spec.LaunchTemplateID,
			LaunchTemplateName: spec.LaunchTemplateName,
			Version:            spec.Version,
		}
	}
	return nil
}

// GenerateEC2LicenseConfigurationRequest coverts an internal slice of LicenseConfigurationRequest into a slice of ec2.LicenseConfigurationRequest
func GenerateEC2LicenseConfigurationRequest(reqs []manualv1alpha1.LicenseConfigurationRequest) []ec2.LicenseConfigurationRequest {
	if reqs != nil {
		res := make([]ec2.LicenseConfigurationRequest, len(reqs))
		for i, r := range reqs {
			res[i] = ec2.LicenseConfigurationRequest{
				LicenseConfigurationArn: r.LicenseConfigurationARN,
			}
		}

		return res
	}
	return nil
}

// GenerateLicenseConfigurationRequest coverts a slice of ec2.LicenseConfiguration into an internal slice of LicenseConfigurationRequest
func GenerateLicenseConfigurationRequest(reqs []ec2.LicenseConfiguration) []manualv1alpha1.LicenseConfigurationRequest {
	if reqs != nil {
		res := make([]manualv1alpha1.LicenseConfigurationRequest, len(reqs))
		for i, r := range reqs {
			res[i] = manualv1alpha1.LicenseConfigurationRequest{
				LicenseConfigurationARN: r.LicenseConfigurationArn,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2Monitoring converts internal RunInstancesMonitoringEnabled into ec2.RunInstancesMonitoringEnabled
func GenerateEC2Monitoring(m *manualv1alpha1.RunInstancesMonitoringEnabled) *ec2.RunInstancesMonitoringEnabled {
	if m != nil {
		return &ec2.RunInstancesMonitoringEnabled{
			Enabled: m.Enabled,
		}
	}
	return nil
}

// GenerateMonitoring converts a ec2.Monitoring into a internal Monitoring
func GenerateMonitoring(m *ec2.Monitoring) *manualv1alpha1.Monitoring {
	if m != nil {
		return &manualv1alpha1.Monitoring{
			State: string(m.State),
		}
	}
	return nil
}

// GenerateEC2Placement converts internal Placement into ec2.Placement
func GenerateEC2Placement(p *manualv1alpha1.Placement) *ec2.Placement {
	if p != nil {
		return &ec2.Placement{
			Affinity:             p.Affinity,
			AvailabilityZone:     p.AvailabilityZone,
			GroupName:            p.GroupName,
			HostId:               p.HostID,
			HostResourceGroupArn: p.HostResourceGroupARN,
			PartitionNumber:      p.PartitionNumber,
			SpreadDomain:         p.SpreadDomain,
			Tenancy:              ec2.Tenancy(p.Tenancy),
		}
	}
	return nil
}

// GeneratePlacement converts ec2.Placement into an internal Placement
func GeneratePlacement(p *ec2.Placement) *manualv1alpha1.Placement {
	if p != nil {
		return &manualv1alpha1.Placement{
			Affinity:             p.Affinity,
			AvailabilityZone:     p.AvailabilityZone,
			GroupName:            p.GroupName,
			HostID:               p.HostId,
			HostResourceGroupARN: p.HostResourceGroupArn,
			PartitionNumber:      p.PartitionNumber,
			SpreadDomain:         p.SpreadDomain,
			Tenancy:              string(p.Tenancy),
		}
	}
	return nil
}

// GenerateProductCodes converts ec2.ProductCode into an internal ProductCode
func GenerateProductCodes(codes []ec2.ProductCode) []manualv1alpha1.ProductCode {
	if codes != nil {
		res := make([]manualv1alpha1.ProductCode, len(codes))
		for i, c := range codes {
			res[i] = manualv1alpha1.ProductCode{
				ProductCodeID:   c.ProductCodeId,
				ProductCodeType: string(c.ProductCodeType),
			}
		}
		return res
	}
	return nil
}

// GenerateEC2PrivateIPAddressSpecs coverts an internal slice of PrivateIPAddressSpecification into a slice of ec2.PrivateIpAddressSpecification
func GenerateEC2PrivateIPAddressSpecs(specs []manualv1alpha1.PrivateIPAddressSpecification) []ec2.PrivateIpAddressSpecification {
	if specs != nil {
		res := make([]ec2.PrivateIpAddressSpecification, len(specs))
		for i, s := range specs {
			res[i] = ec2.PrivateIpAddressSpecification{
				Primary:          s.Primary,
				PrivateIpAddress: s.PrivateIPAddress,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2RunInstancesInput generates a ec2.RunInstanceInput based on the supplied managed resource Name and InstanceParameters
// Note: MaxCount and MinCount are set to 1 each per https://github.com/crossplane/provider-aws/pull/777#issuecomment-887017783
func GenerateEC2RunInstancesInput(name string, p *manualv1alpha1.InstanceParameters) *ec2.RunInstancesInput {
	return &ec2.RunInstancesInput{
		BlockDeviceMappings:               GenerateEC2BlockDeviceMappings(p.BlockDeviceMappings),
		CapacityReservationSpecification:  GenerateEC2CapacityReservationSpecs(p.CapacityReservationSpecification),
		ClientToken:                       p.ClientToken,
		CpuOptions:                        GenerateEC2CPUOptions(p.CPUOptions),
		CreditSpecification:               GenerateEC2CreditSpec(p.CreditSpecification),
		DisableApiTermination:             p.DisableAPITermination,
		EbsOptimized:                      p.EBSOptimized,
		ElasticGpuSpecification:           GenerateEC2ElasticGPUSpecs(p.ElasticGPUSpecification),
		ElasticInferenceAccelerators:      GenerateEC2ElasticInferenceAccelerators(p.ElasticInferenceAccelerators),
		HibernationOptions:                GenerateEC2HibernationOptions(p.HibernationOptions),
		IamInstanceProfile:                GenerateEC2IAMInstanceProfileSpecification(p.IAMInstanceProfile),
		ImageId:                           p.ImageID,
		InstanceInitiatedShutdownBehavior: ec2.ShutdownBehavior(p.InstanceInitiatedShutdownBehavior),
		InstanceMarketOptions:             GenerateEC2InstanceMarketOptionsRequest(p.InstanceMarketOptions),
		InstanceType:                      ec2.InstanceType(p.InstanceType),
		Ipv6AddressCount:                  p.IPv6AddressCount,
		Ipv6Addresses:                     GenerateEC2InstanceIPV6Addresses(p.IPv6Addresses),
		KernelId:                          p.KernelID,
		KeyName:                           p.KeyName,
		LaunchTemplate:                    GenerateEC2LaunchTemplateSpec(p.LaunchTemplate),
		LicenseSpecifications:             GenerateEC2LicenseConfigurationRequest(p.LicenseSpecifications),
		MaxCount:                          aws.Int64(1),
		MetadataOptions:                   GenerateEC2InstanceMetadataOptionsRequest(p.MetadataOptions),
		MinCount:                          aws.Int64(1),
		Monitoring:                        GenerateEC2Monitoring(p.Monitoring),
		NetworkInterfaces:                 GenerateEC2InstanceNetworkInterfaceSpecs(p.NetworkInterfaces),
		Placement:                         GenerateEC2Placement(p.Placement),
		PrivateIpAddress:                  p.PrivateIPAddress,
		RamdiskId:                         p.RAMDiskID,
		SecurityGroupIds:                  p.SecurityGroupIDs,
		SubnetId:                          p.SubnetID,
		TagSpecifications:                 GenerateEC2TagSpecifications(p.TagSpecifications),
		UserData:                          p.UserData,
	}
}

// GenerateStateReason converts ec2.StateReason into an internal StateReason
func GenerateStateReason(r *ec2.StateReason) *manualv1alpha1.StateReason {
	if r != nil {
		return &manualv1alpha1.StateReason{
			Code:    r.Code,
			Message: r.Message,
		}
	}
	return nil
}

// GenerateTags converts a slice of ec2.Tag into an internal slice of Tag.
func GenerateTags(tags []ec2.Tag) []manualv1alpha1.Tag {
	if tags != nil {
		res := make([]manualv1alpha1.Tag, len(tags))
		for i, t := range tags {
			res[i] = manualv1alpha1.Tag{
				Key:   *t.Key,
				Value: *t.Value,
			}
		}
		return res
	}
	return nil
}

// GenerateEC2TagSpecifications takes a slice of TagSpecifications and converts it to a
// slice of ec2.TagSpecification
func GenerateEC2TagSpecifications(tagSpecs []manualv1alpha1.TagSpecification) []ec2.TagSpecification {
	if tagSpecs != nil {
		res := make([]ec2.TagSpecification, len(tagSpecs))
		for i, ts := range tagSpecs {
			res[i] = ec2.TagSpecification{
				ResourceType: ec2.ResourceType(*ts.ResourceType),
			}

			tags := make([]ec2.Tag, len(ts.Tags))
			for i, t := range ts.Tags {
				tags[i] = ec2.Tag{
					Key:   aws.String(t.Key),
					Value: aws.String(t.Value),
				}
			}

			res[i].Tags = tags
		}
		return res
	}
	return nil
}

// GenerateDescribeInstancesByExternalTags generates a ec2.DescribeInstancesInput that is used to query for
// Instances by the external labels.
func GenerateDescribeInstancesByExternalTags(extTags map[string]string) *ec2.DescribeInstancesInput {

	ec2Filters := make([]ec2.Filter, len(extTags))
	i := 0
	for k, v := range extTags {
		ec2Filters[i] = ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", k)),
			Values: []string{v},
		}
		i++
	}

	sort.Slice(ec2Filters, func(i, j int) bool {
		return *ec2Filters[i].Name < *ec2Filters[j].Name
	})

	return &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}
}

// FromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func FromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}

// attributeBoolValue helps will comparing bool values against nested pointers
func attributeBoolValue(v *ec2.AttributeBooleanValue) bool {
	if v == nil {
		return false
	}
	return aws.BoolValue(v.Value)
}

// attributeValue helps will comparing string values against nested pointers
func attributeValue(v *ec2.AttributeValue) string {
	if v == nil {
		return ""
	}
	return aws.StringValue(v.Value)
}
