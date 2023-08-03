package fake

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/s3control/s3controliface"
)

// MockS3ControlClient is a mock implementation of the S3ControlAPI interface
// for testing purposes. It allows simulating the behavior of the S3 Control API
// methods and tracking the number of calls made to each method.
type MockS3ControlClient struct {
	s3controliface.S3ControlAPI

	GetAccessPointWithContextOutput s3control.GetAccessPointOutput
	GetAccessPointWithContextErr    error

	CreateAccessPointWithContextOutput s3control.CreateAccessPointOutput
	CreateAccessPointWithContextErr    error

	DeleteAccessPointWithContextOutput s3control.DeleteAccessPointOutput
	DeleteAccessPointWithContextErr    error

	GetAccessPointPolicyOutput s3control.GetAccessPointPolicyOutput
	GetAccessPointPolicyErr    error

	PutAccessPointPolicyWithContextOutput s3control.PutAccessPointPolicyOutput
	PutAccessPointPolicyWithContextErr    error

	DeleteAccessPointPolicyWithContextOutput s3control.DeleteAccessPointPolicyOutput
	DeleteAccessPointPolicyWithContextErr    error
}

// DeleteAccessPointWithContext is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) DeleteAccessPointWithContext(aws.Context, *s3control.DeleteAccessPointInput, ...request.Option) (*s3control.DeleteAccessPointOutput, error) {
	return &m.DeleteAccessPointWithContextOutput, m.DeleteAccessPointWithContextErr
}

// CreateAccessPointWithContext is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) CreateAccessPointWithContext(aws.Context, *s3control.CreateAccessPointInput, ...request.Option) (*s3control.CreateAccessPointOutput, error) {
	return &m.CreateAccessPointWithContextOutput, m.CreateAccessPointWithContextErr
}

// GetAccessPointWithContext is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) GetAccessPointWithContext(aws.Context, *s3control.GetAccessPointInput, ...request.Option) (*s3control.GetAccessPointOutput, error) {
	return &m.GetAccessPointWithContextOutput, m.GetAccessPointWithContextErr
}

// GetAccessPointPolicy is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) GetAccessPointPolicy(_ *s3control.GetAccessPointPolicyInput) (*s3control.GetAccessPointPolicyOutput, error) {
	return &m.GetAccessPointPolicyOutput, m.GetAccessPointPolicyErr
}

// PutAccessPointPolicyWithContext is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) PutAccessPointPolicyWithContext(_ aws.Context, _ *s3control.PutAccessPointPolicyInput, _ ...request.Option) (*s3control.PutAccessPointPolicyOutput, error) {
	return &m.PutAccessPointPolicyWithContextOutput, m.PutAccessPointPolicyWithContextErr
}

// DeleteAccessPointPolicyWithContext is the fake method call to invoke the internal mock method
func (m *MockS3ControlClient) DeleteAccessPointPolicyWithContext(_ aws.Context, _ *s3control.DeleteAccessPointPolicyInput, _ ...request.Option) (*s3control.DeleteAccessPointPolicyOutput, error) {
	return &m.DeleteAccessPointPolicyWithContextOutput, m.DeleteAccessPointPolicyWithContextErr
}
