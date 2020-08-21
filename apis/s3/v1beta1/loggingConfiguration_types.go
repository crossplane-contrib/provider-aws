package v1beta1

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// LoggingConfiguration describes where logs are stored and the prefix that Amazon S3 assigns to
// all log object keys for a bucket. For more information, see PUT Bucket logging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlogging.html)
type LoggingConfiguration struct {
	// A prefix for all log object keys.
	TargetPrefix string `json:"targetPrefix"`

	// Container for granting information.
	TargetGrants []TargetGrant `json:"targetGrants"`
}

// TargetGrant is the container for granting information.
type TargetGrant struct {
	// Container for the person being granted permissions.
	Grantee *TargetGrantee `json:"targetGrantee,omitempty"`

	// Logging permissions assigned to the Grantee for the bucket.
	// Valid values are "FULL_CONTROL", "READ", "WRITE"
	Permission string `json:"bucketLogsPermission"`
}

// TargetGrantee is the container for the person being granted permissions.
type TargetGrantee struct {
	// Screen name of the grantee.
	DisplayName *string `json:"displayName,omitempty"`

	// Email address of the grantee.
	// For a list of all the Amazon S3 supported Regions and endpoints, see Regions
	// and Endpoints (https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region)
	// in the AWS General Reference.
	EmailAddress *string `json:"emailAddress,omitempty"`

	// The canonical user ID of the grantee.
	ID *string `json:"ID,omitempty"`

	// Type of grantee
	// Type is a required field
	Type string `json:"type"`

	// URI of the grantee group.
	URI *string `json:"URI,omitempty"`
}

// CompareStrings compares pairs of strings passed in
func CompareStrings(strings ...*string) bool {
	if len(strings)%2 != 0 {
		return false
	}
	for i := 0; i < len(strings); i += 2 {
		if aws.StringValue(strings[i]) != aws.StringValue(strings[i+1]) {
			return false
		}
	}
	return true
}

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (acc *LoggingConfiguration) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (managed.ExternalObservation, error) {
	conf, err := client.GetBucketLoggingRequest(&awss3.GetBucketLoggingInput{Bucket: bucketName}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get bucket encryption")
	}

	enabled := conf.LoggingEnabled

	if aws.StringValue(enabled.TargetPrefix) != acc.TargetPrefix {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
	}

	for i, grant := range acc.TargetGrants {
		outputGrant := enabled.TargetGrants[i]
		if outputGrant.Grantee != nil && grant.Grantee != nil {
			oGrant := outputGrant.Grantee
			lGrant := grant.Grantee
			if !CompareStrings(oGrant.DisplayName, lGrant.DisplayName,
				oGrant.EmailAddress, lGrant.EmailAddress,
				oGrant.ID, lGrant.ID,
				oGrant.URI, lGrant.URI) {
				return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
			}
			if string(oGrant.Type) != lGrant.Type {
				return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
			}
		}
		if string(outputGrant.Permission) != grant.Permission {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
	}

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}
