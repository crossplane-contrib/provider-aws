package s3

import (
	"context"
	"strings"

	"github.com/crossplane/provider-aws/apis/s3/v1alpha1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	errors "golang.org/x/xerrors"
)

// ObjectClient is the external client used for Object Custom Resource
type ObjectClient interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObjects(ctx context.Context, input *s3.DeleteObjectsInput, opts ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
}

// NewObjectClient returns a new client given an aws config
func NewObjectClient(cfg aws.Config) ObjectClient {
	return s3.NewFromConfig(cfg)
}

// IsErrorObjectNotFound returns true if the error code indicates that the item was not found
func IsErrorObjectNotFound(err error) bool {
	var nsk *s3types.NoSuchKey
	return errors.As(err, &nsk)
}

// GenerateCreateObjectInput creates the input for CreateObject S3 Client request
func GenerateCreateObjectInput(s v1alpha1.ObjectParameters) *s3.PutObjectInput {
	poi := &s3.PutObjectInput{
		ACL:              s3types.ObjectCannedACL(aws.ToString(s.ACL)),
		Bucket:           s.BucketName,
		GrantFullControl: s.GrantFullControl,
		GrantRead:        s.GrantRead,
		GrantReadACP:     s.GrantReadACP,
		GrantWriteACP:    s.GrantWriteACP,
		Key:              s.Key,
	}
	if s.Body != nil {
		poi.Body = strings.NewReader(*s.Body)
	}
	if s.Expires != nil {
		poi.Expires = &s.Expires.Time
	}
	return poi
}
