package v1beta1

// LoggingConfiguration describes where logs are stored and the prefix that Amazon S3 assigns to
// all log object keys for a bucket. For more information, see PUT Bucket logging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlogging.html)
type LoggingConfiguration struct {
	// the target bucket where logs will be stored, it can be the same bucket.
	TargetBucket *string `json:"targetBucket"`

	// A prefix for all log object keys.
	TargetPrefix *string `json:"targetPrefix"`

	// Container for granting information.
	TargetGrants []TargetGrant `json:"targetGrants"`
}

// TargetGrant is the container for granting information.
type TargetGrant struct {
	// Container for the person being granted permissions.
	Grantee TargetGrantee `json:"targetGrantee"`

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
