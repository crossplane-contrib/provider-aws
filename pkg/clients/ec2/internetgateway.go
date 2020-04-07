package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// InternetGatewayIDNotFound is the code that is returned by ec2 when the given SubnetID is not valid
	InternetGatewayIDNotFound = "InvalidInternetGatewayID.NotFound"
)

// InternetGatewayClient is the external client used for InternetGateway Custom Resource
type InternetGatewayClient interface {
	CreateInternetGatewayRequest(input *ec2.CreateInternetGatewayInput) ec2.CreateInternetGatewayRequest
	DeleteInternetGatewayRequest(input *ec2.DeleteInternetGatewayInput) ec2.DeleteInternetGatewayRequest
	DescribeInternetGatewaysRequest(input *ec2.DescribeInternetGatewaysInput) ec2.DescribeInternetGatewaysRequest
	AttachInternetGatewayRequest(input *ec2.AttachInternetGatewayInput) ec2.AttachInternetGatewayRequest
	DetachInternetGatewayRequest(input *ec2.DetachInternetGatewayInput) ec2.DetachInternetGatewayRequest
	CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewInternetGatewayClient returns a new client using AWS credentials as JSON encoded data.
func NewInternetGatewayClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (InternetGatewayClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), err
}

// IsInternetGatewayNotFoundErr returns true if the error is because the item doesn't exist
func IsInternetGatewayNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == InternetGatewayIDNotFound {
			return true
		}
	}

	return false
}

// GenerateIGObservation is used to produce v1beta1.InternetGatewayExternalStatus from
// ec2.InternetGateway.
func GenerateIGObservation(ig ec2.InternetGateway) v1beta1.InternetGatewayObservation {
	attachments := make([]v1beta1.InternetGatewayAttachment, len(ig.Attachments))
	for k, a := range ig.Attachments {
		attachments[k] = v1beta1.InternetGatewayAttachment{
			AttachmentStatus: string(a.State),
			VPCID:            aws.StringValue(a.VpcId),
		}
	}

	return v1beta1.InternetGatewayObservation{
		InternetGatewayID: aws.StringValue(ig.InternetGatewayId),
		Attachments:       attachments,
		OwnerID:           aws.StringValue(ig.OwnerId),
	}
}

// IsIgUpToDate checks whether there is a change in any of the modifiable fields.
func IsIgUpToDate(p v1beta1.InternetGatewayParameters, ig ec2.InternetGateway) bool {
	attachments := ig.Attachments
	upToDate := false

	// if there are no attachments for obsreved IG and in spec.
	if len(attachments) == 0 && p.VPCID == "" {
		upToDate = true
	}

	// if the attachment in spec exists in ig.Attachments(if any).
	for _, a := range attachments {
		if p.VPCID == aws.StringValue(a.VpcId) {
			upToDate = true
			break
		}
	}

	return upToDate && v1beta1.CompareTags(p.Tags, ig.Tags)
}
