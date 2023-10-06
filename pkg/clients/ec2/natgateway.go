package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

const (
	// NatGatewayNotFound is the code that is returned by ec2 when the given NATGatewayID is not valid
	// ref: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html#api-error-codes-table-client
	NatGatewayNotFound = "NatGatewayNotFound"
)

// NatGatewayClient is the external client used for NatGateway Custom Resource
type NatGatewayClient interface {
	CreateNatGateway(ctx context.Context, input *ec2.CreateNatGatewayInput, opts ...func(*ec2.Options)) (*ec2.CreateNatGatewayOutput, error)
	DeleteNatGateway(ctx context.Context, input *ec2.DeleteNatGatewayInput, opts ...func(*ec2.Options)) (*ec2.DeleteNatGatewayOutput, error)
	DescribeNatGateways(ctx context.Context, input *ec2.DescribeNatGatewaysInput, opts ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// NewNatGatewayClient returns a new client using AWS credentials as JSON encoded data.
func NewNatGatewayClient(cfg aws.Config) NatGatewayClient {
	return ec2.NewFromConfig(cfg)
}

// IsNatGatewayNotFoundErr returns true if the error is because the item doesn't exist
func IsNatGatewayNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == NatGatewayNotFound
}

// GenerateNATGatewayObservation is used to produce v1beta1.NatGatewayObservation from
// ec2types.NatGateway.
func GenerateNATGatewayObservation(nat ec2types.NatGateway) v1beta1.NATGatewayObservation {
	addresses := make([]v1beta1.NATGatewayAddress, len(nat.NatGatewayAddresses))
	for k, a := range nat.NatGatewayAddresses {
		addresses[k] = v1beta1.NATGatewayAddress{
			AllocationID:       aws.ToString(a.AllocationId),
			NetworkInterfaceID: aws.ToString(a.NetworkInterfaceId),
			PrivateIP:          aws.ToString(a.PrivateIp),
			PublicIP:           aws.ToString(a.PublicIp),
		}
	}
	observation := v1beta1.NATGatewayObservation{
		CreateTime:          &metav1.Time{Time: *nat.CreateTime},
		NatGatewayAddresses: addresses,
		NatGatewayID:        aws.ToString(nat.NatGatewayId),
		State:               string(nat.State),
		VpcID:               aws.ToString(nat.VpcId),
	}
	if nat.DeleteTime != nil {
		observation.DeleteTime = &metav1.Time{Time: *nat.DeleteTime}
	}
	if nat.State == ec2types.NatGatewayStateFailed {
		observation.FailureCode = aws.ToString(nat.FailureCode)
		observation.FailureMessage = aws.ToString(nat.FailureMessage)
	}
	return observation
}
