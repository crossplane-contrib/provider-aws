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
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	// InstanceNotFound is the code that is returned by ec2 when the given InstanceID is not valid
	InstanceNotFound = "InvalidInstanceID.NotFound"
)

// InstanceClient is the external client used for Instance Custom Resource
type InstanceClient interface {
	RunInstances(context.Context, *ec2.RunInstancesInput, ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	TerminateInstances(context.Context, *ec2.TerminateInstancesInput, ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeInstanceAttribute(context.Context, *ec2.DescribeInstanceAttributeInput, ...func(*ec2.Options)) (*ec2.DescribeInstanceAttributeOutput, error)
	ModifyInstanceAttribute(context.Context, *ec2.ModifyInstanceAttributeInput, ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error)
	CreateTags(context.Context, *ec2.CreateTagsInput, ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// NewInstanceClient returns a new client using AWS credentials as JSON encoded data.
func NewInstanceClient(cfg aws.Config) InstanceClient {
	return ec2.NewFromConfig(cfg)
}

// IsInstanceNotFoundErr returns true if the error is because the item doesn't exist
func IsInstanceNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == InstanceNotFound
}

// IsInstanceUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsInstanceUpToDate(spec manualv1alpha1.InstanceParameters, instance types.Instance, attributes ec2.DescribeInstanceAttributeOutput) bool {
	// DisableApiTermination
	if pointer.BoolValue(spec.DisableAPITermination) != attributeBoolValue(attributes.DisableApiTermination) {
		return false
	}
	// InstanceInitiatedShutdownBehavior
	if spec.InstanceInitiatedShutdownBehavior != attributeValue(attributes.InstanceInitiatedShutdownBehavior) {
		return false
	}
	// KernalID
	if pointer.StringValue(spec.KernelID) != pointer.StringValue(instance.KernelId) {
		return false
	}
	// RamDiskID
	if pointer.StringValue(spec.RAMDiskID) != pointer.StringValue(instance.RamdiskId) {
		return false
	}
	// UserData
	if pointer.StringValue(spec.UserData) != attributeValue(attributes.UserData) {
		return false
	}
	return CompareGroupIDs(spec.SecurityGroupIDs, instance.SecurityGroups)
}

// GenerateInstanceObservation is used to produce manualv1alpha1.InstanceObservation from
// a []ec2.Instance.
func GenerateInstanceObservation(i types.Instance) manualv1alpha1.InstanceObservation {
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
	case string(types.InstanceStateNameRunning):
		return Available
	case string(types.InstanceStateNameShuttingDown):
		return Deleting
	case string(types.InstanceStateNameStopped):
		return Deleting
	case string(types.InstanceStateNameStopping):
		return Deleting
	case string(types.InstanceStateNameTerminated):
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
func LateInitializeInstance(in *manualv1alpha1.InstanceParameters, instance *types.Instance, attributes *ec2.DescribeInstanceAttributeOutput) { //nolint:gocyclo
	if instance == nil {
		return
	}

	if attributes.DisableApiTermination != nil {
		in.DisableAPITermination = pointer.LateInitialize(in.DisableAPITermination, attributes.DisableApiTermination.Value)
	}

	if attributes.InstanceInitiatedShutdownBehavior != nil {
		in.InstanceInitiatedShutdownBehavior = pointer.LateInitializeValueFromPtr(in.InstanceInitiatedShutdownBehavior, attributes.InstanceInitiatedShutdownBehavior.Value)
	}

	if attributes.InstanceType != nil {
		in.InstanceType = pointer.LateInitializeValueFromPtr(in.InstanceType, attributes.InstanceType.Value)
	}

	if attributes.UserData != nil {
		in.UserData = pointer.LateInitialize(in.UserData, attributes.UserData.Value)
	}

	in.EBSOptimized = pointer.LateInitialize(in.EBSOptimized, instance.EbsOptimized)
	in.KernelID = pointer.LateInitialize(in.KernelID, instance.KernelId)
	in.RAMDiskID = pointer.LateInitialize(in.RAMDiskID, instance.RamdiskId)

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
func GenerateEC2BlockDeviceMappings(mappings []manualv1alpha1.BlockDeviceMapping) []types.BlockDeviceMapping {
	if mappings != nil {
		res := make([]types.BlockDeviceMapping, len(mappings))
		for i, bm := range mappings {
			res[i] = types.BlockDeviceMapping{
				DeviceName: bm.DeviceName,
				Ebs: &types.EbsBlockDevice{
					DeleteOnTermination: bm.EBS.DeleteOnTermination,
					Encrypted:           bm.EBS.Encrypted,
					Iops:                bm.EBS.IOps,
					Throughput:          bm.EBS.Throughput,
					KmsKeyId:            bm.EBS.KmsKeyID,
					SnapshotId:          bm.EBS.SnapshotID,
					VolumeSize:          bm.EBS.VolumeSize,
					VolumeType:          types.VolumeType(bm.EBS.VolumeType),
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
func GenerateEC2CapacityReservationSpecs(spec *manualv1alpha1.CapacityReservationSpecification) *types.CapacityReservationSpecification {
	if spec != nil {
		var capacityReservationID *string
		if spec.CapacityReservationTarget != nil {
			capacityReservationID = spec.CapacityReservationTarget.CapacityReservationID
		}

		return &types.CapacityReservationSpecification{
			CapacityReservationPreference: types.CapacityReservationPreference(spec.CapacityReservationPreference),
			CapacityReservationTarget: &types.CapacityReservationTarget{
				CapacityReservationId: capacityReservationID,
			},
		}
	}
	return nil
}

// GenerateCapacityReservationSpecResponse converts a ec2.CapacityReservationSpecificationResponse into an internal CapacityReservationSpecificationResponse
func GenerateCapacityReservationSpecResponse(resp *types.CapacityReservationSpecificationResponse) *manualv1alpha1.CapacityReservationSpecificationResponse {
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
func GenerateEC2CPUOptions(opts *manualv1alpha1.CPUOptionsRequest) *types.CpuOptionsRequest {
	if opts != nil {
		return &types.CpuOptionsRequest{
			CoreCount:      opts.CoreCount,
			ThreadsPerCore: opts.ThreadsPerCore,
		}
	}
	return nil
}

// GenerateCPUOptionsRequest converts a CpuOptions into a internal CpuOptionsRequest
func GenerateCPUOptionsRequest(opts *types.CpuOptions) *manualv1alpha1.CPUOptionsRequest {
	if opts != nil {
		return &manualv1alpha1.CPUOptionsRequest{
			CoreCount:      opts.CoreCount,
			ThreadsPerCore: opts.ThreadsPerCore,
		}
	}
	return nil
}

// GenerateEC2CreditSpec converts an internal CreditSpecificationRequest into a ec2.CreditSpecificationRequest
func GenerateEC2CreditSpec(spec *manualv1alpha1.CreditSpecificationRequest) *types.CreditSpecificationRequest {
	if spec != nil {
		return &types.CreditSpecificationRequest{
			CpuCredits: spec.CPUCredits,
		}
	}
	return nil
}

// GenerateElasticGPUAssociation coverts a slice of ec2.ElasticGpuAssociation into an internal slice of ElasticGPUAssociation
func GenerateElasticGPUAssociation(assocs []types.ElasticGpuAssociation) []manualv1alpha1.ElasticGPUAssociation {
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
func GenerateEC2ElasticGPUSpecs(specs []manualv1alpha1.ElasticGPUSpecification) []types.ElasticGpuSpecification {
	if specs != nil {
		res := make([]types.ElasticGpuSpecification, len(specs))
		for i, gs := range specs {
			res[i] = types.ElasticGpuSpecification{
				Type: gs.Type,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2ElasticInferenceAccelerators coverts an internal slice of ElasticInferenceAccelerator into a slice of ec2.ElasticInferenceAccelerator
func GenerateEC2ElasticInferenceAccelerators(accs []manualv1alpha1.ElasticInferenceAccelerator) []types.ElasticInferenceAccelerator {
	if accs != nil {
		res := make([]types.ElasticInferenceAccelerator, len(accs))
		for i, a := range accs {
			res[i] = types.ElasticInferenceAccelerator{
				Count: a.Count,
				Type:  a.Type,
			}
		}

		return res
	}
	return nil
}

// GenerateElasticInferenceAcceleratorAssociation coverts a slice of ec2.ElasticInferenceAcceleratorAssociation into an internal slice of ElasticInferenceAcceleratorAssociation
func GenerateElasticInferenceAcceleratorAssociation(assocs []types.ElasticInferenceAcceleratorAssociation) []manualv1alpha1.ElasticInferenceAcceleratorAssociation {
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
func GenerateEC2HibernationOptions(opts *manualv1alpha1.HibernationOptionsRequest) *types.HibernationOptionsRequest {
	if opts != nil {
		return &types.HibernationOptionsRequest{
			Configured: opts.Configured,
		}
	}
	return nil
}

// GenerateGroupIdentifiers coverts a slice of ec2.GroupIdentifier into an internal slice of GroupIdentifier
func GenerateGroupIdentifiers(ids []types.GroupIdentifier) []manualv1alpha1.GroupIdentifier {
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
func GenerateIAMInstanceProfile(p *types.IamInstanceProfile) *manualv1alpha1.IAMInstanceProfile {
	if p != nil {
		return &manualv1alpha1.IAMInstanceProfile{
			ARN: p.Arn,
			ID:  p.Id,
		}
	}
	return nil
}

// GenerateEC2IAMInstanceProfileSpecification converts an internal IamInstanceProfileSpecification into a ec2.IamInstanceProfileSpecification
func GenerateEC2IAMInstanceProfileSpecification(spec *manualv1alpha1.IAMInstanceProfileSpecification) *types.IamInstanceProfileSpecification {
	if spec != nil {
		return &types.IamInstanceProfileSpecification{
			Arn:  spec.ARN,
			Name: spec.Name,
		}
	}
	return nil
}

// GenerateInstanceBlockDeviceMappings coverts a slice of ec2.InstanceBlockDeviceMapping into an internal slice of InstanceBlockDeviceMapping
func GenerateInstanceBlockDeviceMappings(mappings []types.InstanceBlockDeviceMapping) []manualv1alpha1.InstanceBlockDeviceMapping {
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
func GenerateEC2InstanceMarketOptionsRequest(opts *manualv1alpha1.InstanceMarketOptionsRequest) *types.InstanceMarketOptionsRequest {
	if opts != nil {
		var durationMin *int32
		var behavior types.InstanceInterruptionBehavior
		var maxPrice *string
		var instanceType types.SpotInstanceType
		var validUntil *time.Time

		if opts.SpotOptions != nil {
			durationMin = opts.SpotOptions.BlockDurationMinutes
			behavior = types.InstanceInterruptionBehavior(opts.SpotOptions.InstanceInterruptionBehavior)
			maxPrice = opts.SpotOptions.MaxPrice
			instanceType = types.SpotInstanceType(opts.SpotOptions.SpotInstanceType)
			if opts.SpotOptions.ValidUntil != nil {
				validUntil = &opts.SpotOptions.ValidUntil.DeepCopy().Time
			}
		}

		return &types.InstanceMarketOptionsRequest{
			MarketType: types.MarketType(opts.MarketType),
			SpotOptions: &types.SpotMarketOptions{
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
func GenerateInstanceMetadataOptionsRequest(opts *types.InstanceMetadataOptionsResponse) *manualv1alpha1.InstanceMetadataOptionsRequest {
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
func GenerateEC2InstanceIPV6Addresses(addrs []manualv1alpha1.InstanceIPv6Address) []types.InstanceIpv6Address {
	if addrs != nil {
		res := make([]types.InstanceIpv6Address, len(addrs))
		for i, a := range addrs {
			res[i] = types.InstanceIpv6Address{
				Ipv6Address: a.IPv6Address,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2Ipv6PrefixSpecificationRequest coverts an internal slice of Ipv6PrefixSpecificationRequest into a slice of ec2.Ipv6PrefixSpecificationRequest
func GenerateEC2Ipv6PrefixSpecificationRequest(prefixes []manualv1alpha1.Ipv6PrefixSpecificationRequest) []types.Ipv6PrefixSpecificationRequest {
	if len(prefixes) == 0 {
		return nil
	}
	res := make([]types.Ipv6PrefixSpecificationRequest, len(prefixes))
	for i, a := range prefixes {
		res[i] = types.Ipv6PrefixSpecificationRequest{
			Ipv6Prefix: aws.String(a.Ipv6Prefix),
		}
	}
	return res
}

// GenerateInstanceIPV6Addresses coverts a slice of ec2.InstanceIpv6Address into a slice of internal InstanceIPv6Address
func GenerateInstanceIPV6Addresses(addrs []types.InstanceIpv6Address) []manualv1alpha1.InstanceIPv6Address {
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
func GenerateEC2InstanceMetadataOptionsRequest(opts *manualv1alpha1.InstanceMetadataOptionsRequest) *types.InstanceMetadataOptionsRequest {
	if opts != nil {
		return &types.InstanceMetadataOptionsRequest{
			HttpEndpoint:            types.InstanceMetadataEndpointState(opts.HTTPEndpoint),
			HttpPutResponseHopLimit: opts.HTTPPutResponseHopLimit,
			HttpTokens:              types.HttpTokensState(opts.HTTPTokens),
		}
	}
	return nil
}

// GenerateInstanceNetworkInterface coverts a slice of ec2.InstanceNetworkInterface into an internal slice of InstanceNetworkInterface
func GenerateInstanceNetworkInterface(nets []types.InstanceNetworkInterface) []manualv1alpha1.InstanceNetworkInterface {
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
func GenerateEC2InstanceNetworkInterfaceSpecs(specs []manualv1alpha1.InstanceNetworkInterfaceSpecification) []types.InstanceNetworkInterfaceSpecification {
	if specs != nil {
		res := make([]types.InstanceNetworkInterfaceSpecification, len(specs))
		for i, s := range specs {
			res[i] = types.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress:       s.AssociatePublicIPAddress,
				DeleteOnTermination:            s.DeleteOnTermination,
				Description:                    s.Description,
				DeviceIndex:                    s.DeviceIndex,
				Groups:                         s.Groups,
				InterfaceType:                  s.InterfaceType,
				Ipv6AddressCount:               s.IPv6AddressCount,
				Ipv6Addresses:                  GenerateEC2InstanceIPV6Addresses(s.IPv6Addresses),
				Ipv6PrefixCount:                s.Ipv6PrefixCount,
				Ipv6Prefixes:                   GenerateEC2Ipv6PrefixSpecificationRequest(s.Ipv6Prefixes),
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
func GenerateInstancePrivateIPAddresses(addrs []types.InstancePrivateIpAddress) []manualv1alpha1.InstancePrivateIPAddress {
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
func GenerateHibernationOptionsRequest(opts *types.HibernationOptions) *manualv1alpha1.HibernationOptionsRequest {
	if opts != nil {
		return &manualv1alpha1.HibernationOptionsRequest{
			Configured: opts.Configured,
		}
	}
	return nil
}

// GenerateEC2LaunchTemplateSpec converts internal LaunchTemplateSpecification into ec2.LaunchTemplateSpecification
func GenerateEC2LaunchTemplateSpec(spec *manualv1alpha1.LaunchTemplateSpecification) *types.LaunchTemplateSpecification {
	if spec != nil {
		return &types.LaunchTemplateSpecification{
			LaunchTemplateId:   spec.LaunchTemplateID,
			LaunchTemplateName: spec.LaunchTemplateName,
			Version:            spec.Version,
		}
	}
	return nil
}

// GenerateEC2LicenseConfigurationRequest coverts an internal slice of LicenseConfigurationRequest into a slice of ec2.LicenseConfigurationRequest
func GenerateEC2LicenseConfigurationRequest(reqs []manualv1alpha1.LicenseConfigurationRequest) []types.LicenseConfigurationRequest {
	if reqs != nil {
		res := make([]types.LicenseConfigurationRequest, len(reqs))
		for i, r := range reqs {
			res[i] = types.LicenseConfigurationRequest{
				LicenseConfigurationArn: r.LicenseConfigurationARN,
			}
		}

		return res
	}
	return nil
}

// GenerateLicenseConfigurationRequest coverts a slice of ec2.LicenseConfiguration into an internal slice of LicenseConfigurationRequest
func GenerateLicenseConfigurationRequest(reqs []types.LicenseConfiguration) []manualv1alpha1.LicenseConfigurationRequest {
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
func GenerateEC2Monitoring(m *manualv1alpha1.RunInstancesMonitoringEnabled) *types.RunInstancesMonitoringEnabled {
	if m != nil {
		return &types.RunInstancesMonitoringEnabled{
			Enabled: m.Enabled,
		}
	}
	return nil
}

// GenerateMonitoring converts a ec2.Monitoring into a internal Monitoring
func GenerateMonitoring(m *types.Monitoring) *manualv1alpha1.Monitoring {
	if m != nil {
		return &manualv1alpha1.Monitoring{
			State: string(m.State),
		}
	}
	return nil
}

// GenerateEC2Placement converts internal Placement into ec2.Placement
func GenerateEC2Placement(p *manualv1alpha1.Placement) *types.Placement {
	if p != nil {
		return &types.Placement{
			Affinity:             p.Affinity,
			AvailabilityZone:     p.AvailabilityZone,
			GroupName:            p.GroupName,
			HostId:               p.HostID,
			HostResourceGroupArn: p.HostResourceGroupARN,
			PartitionNumber:      p.PartitionNumber,
			SpreadDomain:         p.SpreadDomain,
			Tenancy:              types.Tenancy(p.Tenancy),
		}
	}
	return nil
}

// GeneratePlacement converts ec2.Placement into an internal Placement
func GeneratePlacement(p *types.Placement) *manualv1alpha1.Placement {
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
func GenerateProductCodes(codes []types.ProductCode) []manualv1alpha1.ProductCode {
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
func GenerateEC2PrivateIPAddressSpecs(specs []manualv1alpha1.PrivateIPAddressSpecification) []types.PrivateIpAddressSpecification {
	if specs != nil {
		res := make([]types.PrivateIpAddressSpecification, len(specs))
		for i, s := range specs {
			res[i] = types.PrivateIpAddressSpecification{
				Primary:          s.Primary,
				PrivateIpAddress: s.PrivateIPAddress,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2RunInstancesInput generates a ec2.RunInstanceInput based on the supplied managed resource Name and InstanceParameters
// Note: MaxCount and MinCount are set to 1 each per https://github.com/crossplane-contrib/provider-aws/pull/777#issuecomment-887017783
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
		InstanceInitiatedShutdownBehavior: types.ShutdownBehavior(p.InstanceInitiatedShutdownBehavior),
		InstanceMarketOptions:             GenerateEC2InstanceMarketOptionsRequest(p.InstanceMarketOptions),
		InstanceType:                      types.InstanceType(p.InstanceType),
		Ipv6AddressCount:                  p.IPv6AddressCount,
		Ipv6Addresses:                     GenerateEC2InstanceIPV6Addresses(p.IPv6Addresses),
		KernelId:                          p.KernelID,
		KeyName:                           p.KeyName,
		LaunchTemplate:                    GenerateEC2LaunchTemplateSpec(p.LaunchTemplate),
		LicenseSpecifications:             GenerateEC2LicenseConfigurationRequest(p.LicenseSpecifications),
		MaxCount:                          aws.Int32(1),
		MetadataOptions:                   GenerateEC2InstanceMetadataOptionsRequest(p.MetadataOptions),
		MinCount:                          aws.Int32(1),
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
func GenerateStateReason(r *types.StateReason) *manualv1alpha1.StateReason {
	if r != nil {
		return &manualv1alpha1.StateReason{
			Code:    r.Code,
			Message: r.Message,
		}
	}
	return nil
}

// GenerateTags converts a slice of ec2.Tag into an internal slice of Tag.
func GenerateTags(tags []types.Tag) []manualv1alpha1.Tag {
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
func GenerateEC2TagSpecifications(tagSpecs []manualv1alpha1.TagSpecification) []types.TagSpecification {
	if tagSpecs != nil {
		res := make([]types.TagSpecification, len(tagSpecs))
		for i, ts := range tagSpecs {
			res[i] = types.TagSpecification{
				ResourceType: types.ResourceType(*ts.ResourceType),
			}

			tags := make([]types.Tag, len(ts.Tags))
			for i, t := range ts.Tags {
				tags[i] = types.Tag{
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

	ec2Filters := make([]types.Filter, len(extTags))
	i := 0
	for k, v := range extTags {
		ec2Filters[i] = types.Filter{
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
func attributeBoolValue(v *types.AttributeBooleanValue) bool {
	if v == nil {
		return false
	}
	return pointer.BoolValue(v.Value)
}

// attributeValue helps will comparing string values against nested pointers
func attributeValue(v *types.AttributeValue) string {
	if v == nil {
		return ""
	}
	return pointer.StringValue(v.Value)
}
