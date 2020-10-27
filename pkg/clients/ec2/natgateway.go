package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NatGatewayNotFound is the code that is returned by ec2 when the given NATGatewayID is not valid
	// ref: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html#api-error-codes-table-client
	NatGatewayNotFound = "NatGatewayNotFound"
)

// NatGatewayClient is the external client used for NatGateway Custom Resource
type NatGatewayClient interface {
	CreateNatGatewayRequest(input *ec2.CreateNatGatewayInput) ec2.CreateNatGatewayRequest
	DeleteNatGatewayRequest(input *ec2.DeleteNatGatewayInput) ec2.DeleteNatGatewayRequest
	DescribeNatGatewaysRequest(input *ec2.DescribeNatGatewaysInput) ec2.DescribeNatGatewaysRequest
	CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest
	DeleteTagsRequest(input *ec2.DeleteTagsInput) ec2.DeleteTagsRequest
}

// NewNatGatewayClient returns a new client using AWS credentials as JSON encoded data.
func NewNatGatewayClient(cfg aws.Config) NatGatewayClient {
	return ec2.New(cfg)
}

// IsNatGatewayNotFoundErr returns true if the error is because the item doesn't exist
func IsNatGatewayNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == NatGatewayNotFound {
			return true
		}
	}

	return false
}

// GenerateNATGatewayObservation is used to produce v1beta1.NatGatewayObservation from
// ec2.NatGateway.
func GenerateNATGatewayObservation(nat ec2.NatGateway) v1alpha1.NATGatewayObservation {
	addresses := make([]v1alpha1.NATGatewayAddress, len(nat.NatGatewayAddresses))
	for k, a := range nat.NatGatewayAddresses {
		addresses[k] = v1alpha1.NATGatewayAddress{
			AllocationID:       aws.StringValue(a.AllocationId),
			NetworkInterfaceID: aws.StringValue(a.NetworkInterfaceId),
			PrivateIP:          aws.StringValue(a.PrivateIp),
			PublicIP:           aws.StringValue(a.PublicIp),
		}
	}
	observation := v1alpha1.NATGatewayObservation{
		CreateTime:          &metav1.Time{Time: *nat.CreateTime},
		NatGatewayAddresses: addresses,
		NatGatewayID:        aws.StringValue(nat.NatGatewayId),
		State:               string(nat.State),
		VpcID:               aws.StringValue(nat.VpcId),
	}
	if nat.DeleteTime != nil {
		observation.DeleteTime = &metav1.Time{Time: *nat.DeleteTime}
	}
	if nat.State == ec2.NatGatewayStateFailed {
		observation.FailureCode = aws.StringValue(nat.FailureCode)
		observation.FailureMessage = aws.StringValue(nat.FailureMessage)
	}
	return observation
}
