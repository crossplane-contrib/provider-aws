package v1beta1

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// PaymentConfiguration specifies who pays for the download and request fees.
type PaymentConfiguration struct {
	// Payer is a required field, detailing who pays
	// Valid values are "Requester" and "BucketOwner"
	Payer string `json:"payer"`
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (pay *PaymentConfiguration) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (managed.ExternalObservation, error) {
	conf, err := client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get request payment configuration")
	}

	if pay.Payer != string(conf.Payer) {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
	}

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}
