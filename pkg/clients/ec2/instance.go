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

// InstanceClient is the external client used for VPC Custom Resourc
type InstanceClient interface {
	RunInstancesRequest(*ec2.RunInstancesInput) ec2.RunInstancesRequest
	TerminateInstancesRequest(*ec2.TerminateInstancesInput) ec2.TerminateInstancesRequest
	DescribeInstancesRequest(*ec2.DescribeInstancesInput) ec2.DescribeInstancesRequest
	DescribeInstanceAttributeRequest(*ec2.DescribeInstanceAttributeInput) ec2.DescribeInstanceAttributeRequest
	ModifyInstanceAttributeRequest(*ec2.ModifyInstanceAttributeInput) ec2.ModifyInstanceAttributeRequest
	// CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	// ModifyVpcTenancyRequest(*ec2.ModifyVpcTenancyInput) ec2.ModifyVpcTenancyRequest
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
	// if aws.StringValue(spec.InstanceTenancy) != string(vpc.InstanceTenancy) {
	// 	return false
	// }

	// if aws.BoolValue(spec.EnableDNSHostNames) != aws.BoolValue(attributes.EnableDnsHostnames.Value) ||
	// 	aws.BoolValue(spec.EnableDNSSupport) != aws.BoolValue(attributes.EnableDnsSupport.Value) {
	// 	return false
	// }

	// return manualv1alpha1.CompareTags(spec.Tags, vpc.Tags)
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
	// 			StatusMessage: aws.StringValue(v.Ipv6CidrBlockState.StatusMessage),
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
