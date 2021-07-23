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
// ec2.Instance.
func GenerateInstanceObservation(vpc ec2.Instance) manualv1alpha1.InstanceObservation {
	o := manualv1alpha1.InstanceObservation{
		// IsDefault:     aws.BoolValue(vpc.IsDefault),
		// DHCPOptionsID: aws.StringValue(vpc.DhcpOptionsId),
		// OwnerID:       aws.StringValue(vpc.OwnerId),
		// VPCState:      string(vpc.State),
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

	return o
}

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
func GenerateEC2InstanceIPV6Addresses(addrs []manualv1alpha1.InstanceIPV6Address) []ec2.InstanceIpv6Address {
	if addrs != nil {
		res := make([]ec2.InstanceIpv6Address, len(addrs))
		for i, a := range addrs {
			res[i] = ec2.InstanceIpv6Address{
				Ipv6Address: a.IPV6Address,
			}
		}

		return res
	}
	return nil
}

// GenerateEC2Monitoring converts internal RunInstancesMonitoringEnabled into ec2.RunInstancesMonitoringEnabled
func GenerateEC2Monitoring(m *manualv1alpha1.RunInstancesMonitoringEnabled) *ec2.RunInstancesMonitoringEnabled {
	if m != nil {
		var res ec2.RunInstancesMonitoringEnabled
		res.Enabled = m.Enabled
		return &res
	}
	return nil
}

// TransformTagSpecifications takes a slice of TagSpecifications, converts it to a slice
// of ec2.TagSpecification and lastly injects the special instance name awsec2.TagSpecification
func TransformTagSpecifications(mdgName string, tagSpecs []manualv1alpha1.TagSpecification) []ec2.TagSpecification {
	ec2Specs := generateEC2TagSpecifications(tagSpecs)
	return injectInstanceNameTagSpecification(mdgName, ec2Specs)
}

// generateEC2TagSpecifications takes a slice of TagSpecifications and converts it to a
// slice of ec2.TagSpecification
func generateEC2TagSpecifications(tagSpecs []manualv1alpha1.TagSpecification) []ec2.TagSpecification {
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

// injectInstanceNameTagSpecification will inject a special TagSpecification of the following
// shape into the give slice, if it does not yet exist:
// resourceType: instance,
// tags:
//   - key: Name
//     value: <name>
// The resulting behavior is that the specified instance name in metadata.Name is reflected
// in the AWS console.
// Note: If the a TagSpecification of resourceType `instance` with key `Name` was supplied,
// this method does not modify the supplied TagSpecification.
func injectInstanceNameTagSpecification(name string, tagSpecs []ec2.TagSpecification) []ec2.TagSpecification {
	instanceNameKey := "Name"

	specialTagSpec := ec2.TagSpecification{
		ResourceType: "instance",
		Tags: []ec2.Tag{
			{
				Key:   aws.String(instanceNameKey), // if the 'N' isn't capitalized AWS treats this as a general tag
				Value: aws.String(name),
			},
		},
	}

	foundSpecial := false
	for _, ts := range tagSpecs {
		if ts.ResourceType == "instance" {
			for _, tag := range ts.Tags {
				if *tag.Key == instanceNameKey {
					foundSpecial = true
					break
				}
			}
		}
	}

	if !foundSpecial {
		tagSpecs = append(tagSpecs, specialTagSpec)
	}

	return tagSpecs
}
