package fake

import (
	"context"

	clientset "github.com/crossplane/provider-aws/pkg/clients/s3"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// this ensures that the mock implements the client interface
var _ clientset.ObjectClient = (*MockObjectClient)(nil)

// MockObjectClient is a type that implements all the methods for ObjectClient interface
type MockObjectClient struct {
	MockGetObject          func(ctx context.Context, input *s3.GetObjectInput, opts []func(*s3.Options)) (*s3.GetObjectOutput, error)
	MockPutObject          func(ctx context.Context, input *s3.PutObjectInput, opts []func(*s3.Options)) (*s3.PutObjectOutput, error)
	MockDeleteObjects      func(ctx context.Context, input *s3.DeleteObjectsInput, opts []func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	MockListObjectVersions func(ctx context.Context, input *s3.ListObjectVersionsInput, opts []func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
}

// GetObject mocks GetObject method
func (m MockObjectClient) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.MockGetObject(ctx, input, opts)
}

// PutObject mocks PutObject method
func (m MockObjectClient) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.MockPutObject(ctx, input, opts)
}

// DeleteObjects mocks DeleteObjects method
func (m MockObjectClient) DeleteObjects(ctx context.Context, input *s3.DeleteObjectsInput, opts ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return m.MockDeleteObjects(ctx, input, opts)
}

// ListObjectVersions mocks ListObjectVersions method
func (m MockObjectClient) ListObjectVersions(ctx context.Context, input *s3.ListObjectVersionsInput, opts ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return m.MockListObjectVersions(ctx, input, opts)
}
