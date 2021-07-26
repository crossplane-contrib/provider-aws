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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
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
	return true
}

// GenerateInstanceObservation is used to produce manualv1alpha1.InstanceObservation from
// a []ec2.Instance.
func GenerateInstanceObservation(instances []ec2.Instance) manualv1alpha1.InstanceObservation {
	// determine the overall statuses for the obeserved instances
	state := manualv1alpha1.InstancesState{
		Total: len(instances),
	}

	for _, i := range instances {
		switch i.State.Name {
		case ec2.InstanceStateNamePending:
			state.Pending++
		case ec2.InstanceStateNameRunning:
			state.Running++
		case ec2.InstanceStateNameShuttingDown:
			state.ShuttingDown++
		case ec2.InstanceStateNameStopped:
			state.Stopped++
		case ec2.InstanceStateNameStopping:
			state.Stopping++
		case ec2.InstanceStateNameTerminated:
			state.Terminated++
		}
	}

	// if len(vpc.CidrBlockAssociationSet) > 0 {
	// 	o.CIDRBlockAssociationSet = make([]v1beta1.VPCCIDRBlockAssociation, len(vpc.CidrBlockAssociationSet))
	// 	for i, v := range vpc.CidrBlockAssociationSet {
	// 		o.CIDRBlockAssociationSet[i] = v1beta1.VPCCIDRBlockAssociation{
	// 			AssociationID: aws.StringValue(v.AssociationId),
	// 			CIDRBlock:     aws.StringValue(v.CidrBlock),
	// 		}
	// 		o.CIDRBlockAssociationSet[i].CIDRBlockState = v1beta1.VPCCIDRBlockState{
	// 			State:         string(v.CidrBlockState.State),
	// 			StatusMessage: aws.StringValue(v.CidrBlockState.StatusMessage),
	// 		}
	// 	}
	// }

	// if len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
	// 	o.IPv6CIDRBlockAssociationSet = make([]v1beta1.VPCIPv6CidrBlockAssociation, len(vpc.Ipv6CidrBlockAssociationSet))
	// 	for i, v := range vpc.Ipv6CidrBlockAssociationSet {
	// 		o.IPv6CIDRBlockAssociationSet[i] = v1beta1.VPCIPv6CidrBlockAssociation{
	// 			AssociationID:      aws.StringValue(v.AssociationId),
	// 			IPv6CIDRBlock:      aws.StringValue(v.Ipv6CidrBlock),
	// 			IPv6Pool:           aws.StringValue(v.Ipv6Pool),
	// 			NetworkBorderGroup: aws.StringValue(v.NetworkBorderGroup),
	// 		}
	// 		o.IPv6CIDRBlockAssociationSet[i].IPv6CIDRBlockState = v1beta1.VPCCIDRBlockState{
	// 			State:         string(v.Ipv6CidrBlockState.State),
	// 			StatusMessage: raws.StringValue(v.Ipv6CidrBlockState.StatusMessage),
	// 		}
	// 	}
	// }

	return manualv1alpha1.InstanceObservation{
		State: state,
	}
}

// GenerateInstanceCondition returns an instance Condition depending on the supplied
// observation. Currently, the instances (as a group) can be denoted as:
// * Available
// * Creating
// * Deleting
func GenerateInstanceCondition(o manualv1alpha1.InstanceObservation) Condition {
	// if all of the instances are in a running state, then we say the
	// Instance is available
	if o.State.Running == int64(o.State.Total) {
		return Available
	}

	// if all of the instances are terminated, then we say the
	// Instance is deleted
	if o.State.Terminated == int64(o.State.Total) {
		return Deleted
	}

	// if any of the instances are beginning the termination cycle, we
	// say the Instance is deleting
	if o.State.ShuttingDown > 0 ||
		o.State.Stopped > 0 ||
		o.State.Stopping > 0 ||
		o.State.Terminated > 0 {

		return Deleting
	}

	// otherwise we're creating the Instance
	return Creating
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
func LateInitializeInstance(in *manualv1alpha1.InstanceParameters, v *ec2.Instance, attributes *ec2.DescribeInstanceAttributeOutput) { // nolint:gocyclo
	if v == nil {
		return
	}

	// in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, v.CidrBlock)
	// in.InstanceTenancy = awsclients.LateInitializeStringPtr(in.InstanceTenancy, aws.String(string(v.InstanceTenancy)))
	// if attributes.EnableDnsHostnames != nil {
	// 	in.EnableDNSHostNames = awsclients.LateInitializeBoolPtr(in.EnableDNSHostNames, attributes.EnableDnsHostnames.Value)
	// }
	// if attributes.EnableDnsHostnames != nil {
	// 	in.EnableDNSSupport = awsclients.LateInitializeBoolPtr(in.EnableDNSSupport, attributes.EnableDnsSupport.Value)
	// }
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
		return &ec2.CapacityReservationSpecification{
			CapacityReservationPreference: ec2.CapacityReservationPreference(spec.CapacityReservationPreference),
			CapacityReservationTarget: &ec2.CapacityReservationTarget{
				CapacityReservationId: spec.CapacityReservationTarget.CapacityReservationID,
			},
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

// GenerateEC2CreditSpec converts an internal CreditSpecificationRequest into a ec2.CreditSpecificationRequest
func GenerateEC2CreditSpec(spec *manualv1alpha1.CreditSpecificationRequest) *ec2.CreditSpecificationRequest {
	if spec != nil {
		return &ec2.CreditSpecificationRequest{
			CpuCredits: spec.CPUCredits,
		}
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

// GenerateEC2HibernationOptions converts an internal HibernationOptionsRequest into a ec2.HibernationOptionsRequest
func GenerateEC2HibernationOptions(opts *manualv1alpha1.HibernationOptionsRequest) *ec2.HibernationOptionsRequest {
	if opts != nil {
		return &ec2.HibernationOptionsRequest{
			Configured: opts.Configured,
		}
	}
	return nil
}

// GenerateEC2IamInstanceProfileSpecification converts an internal IamInstanceProfileSpecification into a ec2.IamInstanceProfileSpecification
func GenerateEC2IamInstanceProfileSpecification(spec *manualv1alpha1.IamInstanceProfileSpecification) *ec2.IamInstanceProfileSpecification {
	if spec != nil {
		return &ec2.IamInstanceProfileSpecification{
			Arn:  spec.ARN,
			Name: spec.Name,
		}
	}
	return nil
}

// GenerateEC2InstanceMarketOptionsRequest converts an internal InstanceMarketOptionsRequest into a ec2.InstanceMarketOptionsRequest
func GenerateEC2InstanceMarketOptionsRequest(opts *manualv1alpha1.InstanceMarketOptionsRequest) *ec2.InstanceMarketOptionsRequest {
	if opts != nil {
		return &ec2.InstanceMarketOptionsRequest{
			MarketType: ec2.MarketType(opts.MarketType),
			SpotOptions: &ec2.SpotMarketOptions{
				BlockDurationMinutes:         opts.SpotOptions.BlockDurationMinutes,
				InstanceInterruptionBehavior: ec2.InstanceInterruptionBehavior(opts.SpotOptions.InstanceInterruptionBehavior),
				MaxPrice:                     opts.SpotOptions.MaxPrice,
				SpotInstanceType:             ec2.SpotInstanceType(opts.SpotOptions.SpotInstanceType),
				ValidUntil:                   &opts.SpotOptions.ValidUntil.DeepCopy().Time, // need to convert to time.Time (YYYY-MM-DDTHH:MM:SSZ)
			},
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
				Ipv6AddressCount:               s.Ipv6AddressCount,
				Ipv6Addresses:                  GenerateEC2InstanceIPV6Addresses(s.IPV6Addresses),
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

// GenerateEC2Monitoring converts internal RunInstancesMonitoringEnabled into ec2.RunInstancesMonitoringEnabled
func GenerateEC2Monitoring(m *manualv1alpha1.RunInstancesMonitoringEnabled) *ec2.RunInstancesMonitoringEnabled {
	if m != nil {
		return &ec2.RunInstancesMonitoringEnabled{
			Enabled: m.Enabled,
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
func GenerateEC2RunInstancesInput(name string, p *manualv1alpha1.InstanceParameters) *ec2.RunInstancesInput {
	return &ec2.RunInstancesInput{
		BlockDeviceMappings:              GenerateEC2BlockDeviceMappings(p.BlockDeviceMappings),
		CapacityReservationSpecification: GenerateEC2CapacityReservationSpecs(p.CapacityReservationSpecification),
		ClientToken:                      p.ClientToken,
		CpuOptions:                       GenerateEC2CPUOptions(p.CPUOptions),
		CreditSpecification:              GenerateEC2CreditSpec(p.CreditSpecification),
		// DisableApiTermination: cr.Spec.ForProvider.DisableAPITermination, // this setting will have some behavior we need to think through
		DryRun:                            p.DryRun,
		EbsOptimized:                      p.EBSOptimized,
		ElasticGpuSpecification:           GenerateEC2ElasticGPUSpecs(p.ElasticGPUSpecification),
		ElasticInferenceAccelerators:      GenerateEC2ElasticInferenceAccelerators(p.ElasticInferenceAccelerators),
		HibernationOptions:                GenerateEC2HibernationOptions(p.HibernationOptions),
		IamInstanceProfile:                GenerateEC2IamInstanceProfileSpecification(p.IamInstanceProfile),
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
		MaxCount:                          p.MaxCount, // TODO handle the case when we have more than 1 here. If this is not 1, each instance has a different instanceID
		MetadataOptions:                   GenerateEC2InstanceMetadataOptionsRequest(p.MetadataOptions),
		MinCount:                          p.MinCount,
		Monitoring:                        GenerateEC2Monitoring(p.Monitoring),
		NetworkInterfaces:                 GenerateEC2InstanceNetworkInterfaceSpecs(p.NetworkInterfaces),
		Placement:                         GenerateEC2Placement(p.Placement),
		PrivateIpAddress:                  p.PrivateIPAddress,
		RamdiskId:                         p.RAMDiskID,
		SecurityGroupIds:                  p.SecurityGroupIDs,
		SecurityGroups:                    p.SecurityGroups,
		SubnetId:                          p.SubnetID,
		TagSpecifications:                 GenerateEC2TagSpecifications(p.TagSpecifications),
		UserData:                          p.UserData,
	}
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
