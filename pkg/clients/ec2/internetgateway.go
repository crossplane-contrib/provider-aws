package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	// InternetGatewayIDNotFound is the code that is returned by ec2 when the given InternetGatewayID is not valid
	InternetGatewayIDNotFound = "InvalidInternetGatewayID.NotFound"
	// InternetGatewayAlreadyAttached is code for error returned by AWS API
	// for AttachInternetGatewayRequest when an InternetGatway is already attached to specified VPC in the request.
	InternetGatewayAlreadyAttached = "Resource.AlreadyAssociated"
)

// InternetGatewayClient is the external client used for InternetGateway Custom Resource
type InternetGatewayClient interface {
	CreateInternetGateway(ctx context.Context, input *ec2.CreateInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.CreateInternetGatewayOutput, error)
	DeleteInternetGateway(ctx context.Context, input *ec2.DeleteInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error)
	DescribeInternetGateways(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, opts ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
	AttachInternetGateway(ctx context.Context, input *ec2.AttachInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.AttachInternetGatewayOutput, error)
	DetachInternetGateway(ctx context.Context, input *ec2.DetachInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// NewInternetGatewayClient returns a new client using AWS credentials as JSON encoded data.
func NewInternetGatewayClient(cfg aws.Config) InternetGatewayClient {
	return ec2.NewFromConfig(cfg)
}

// IsInternetGatewayNotFoundErr returns true if the error is because the item doesn't exist
func IsInternetGatewayNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == InternetGatewayIDNotFound
}

// IsInternetGatewayAlreadyAttached returns true if the error is because the item doesn't exist
func IsInternetGatewayAlreadyAttached(err error) bool {
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		if awsErr.ErrorCode() == InternetGatewayAlreadyAttached {
			return true
		}
	}
	return false
}

// GenerateIGObservation is used to produce v1beta1.InternetGatewayExternalStatus from
// ec2types.InternetGateway.
func GenerateIGObservation(ig ec2types.InternetGateway) v1beta1.InternetGatewayObservation {
	attachments := make([]v1beta1.InternetGatewayAttachment, len(ig.Attachments))
	for k, a := range ig.Attachments {
		attachments[k] = v1beta1.InternetGatewayAttachment{
			AttachmentStatus: string(a.State),
			VPCID:            aws.ToString(a.VpcId),
		}
	}

	return v1beta1.InternetGatewayObservation{
		InternetGatewayID: aws.ToString(ig.InternetGatewayId),
		Attachments:       attachments,
		OwnerID:           aws.ToString(ig.OwnerId),
	}
}

// LateInitializeIG fills the empty fields in *v1beta1.InternetGatewayParameters with
// the values seen in ec2types.InternetGateway.
func LateInitializeIG(in *v1beta1.InternetGatewayParameters, ig *ec2types.InternetGateway) {
	if ig == nil {
		return
	}
	if ig.Attachments != nil && len(ig.Attachments) > 0 {
		in.VPCID = pointer.LateInitialize(in.VPCID, ig.Attachments[0].VpcId)
	}
	if len(in.Tags) == 0 && len(ig.Tags) != 0 {
		in.Tags = BuildFromEC2TagsV1Beta1(ig.Tags)
	}
}

// IsIgUpToDate checks whether there is a change in any of the modifiable fields.
func IsIgUpToDate(p v1beta1.InternetGatewayParameters, ig ec2types.InternetGateway) bool {

	// if there are no attachments for observed IG and in spec.
	if len(ig.Attachments) == 0 && p.VPCID != nil {
		return false
	}

	// if the attachment in spec exists in ig.Attachments, compare the tags and return
	for _, a := range ig.Attachments {
		if aws.ToString(p.VPCID) == aws.ToString(a.VpcId) {
			return CompareTagsV1Beta1(p.Tags, ig.Tags)
		}
	}

	return false
}
