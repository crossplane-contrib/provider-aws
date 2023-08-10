package fake

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// MockCapacityResourceClient is a mock implementation of the EC2API capacity reservation interface
// for testing purposes. It allows simulating the behavior of the EC2API capacity reservation
// methods and tracking the number of calls made to each method.
type MockCapacityResourceClient struct {
	ec2iface.EC2API

	DescribeCapacityReservationsOutput ec2.DescribeCapacityReservationsOutput
	DescribeCapacityReservationsErr    error

	CreateCapacityReservationOutput ec2.CreateCapacityReservationOutput
	CreateCapacityReservationErr    error

	ModifyCapacityReservationOutput ec2.ModifyCapacityReservationOutput
	ModifyCapacityReservationErr    error

	CancelCapacityReservationOutput ec2.CancelCapacityReservationOutput
	CancelCapacityReservationErr    error
}

// DescribeCapacityReservationsWithContext is the fake method call to invoke the internal mock method
func (m *MockCapacityResourceClient) DescribeCapacityReservationsWithContext(_ aws.Context, _ *ec2.DescribeCapacityReservationsInput, _ ...request.Option) (*ec2.DescribeCapacityReservationsOutput, error) {
	return &m.DescribeCapacityReservationsOutput, m.DescribeCapacityReservationsErr
}

// CreateCapacityReservationWithContext is the fake method call to invoke the internal mock method
func (m *MockCapacityResourceClient) CreateCapacityReservationWithContext(_ aws.Context, _ *ec2.CreateCapacityReservationInput, _ ...request.Option) (*ec2.CreateCapacityReservationOutput, error) {
	return &m.CreateCapacityReservationOutput, m.CreateCapacityReservationErr
}

// ModifyCapacityReservationWithContext is the fake method call to invoke the internal mock method
func (m *MockCapacityResourceClient) ModifyCapacityReservationWithContext(_ aws.Context, _ *ec2.ModifyCapacityReservationInput, _ ...request.Option) (*ec2.ModifyCapacityReservationOutput, error) {
	return &m.ModifyCapacityReservationOutput, m.ModifyCapacityReservationErr
}

// CancelCapacityReservationWithContext is the fake method call to invoke the internal mock method
func (m *MockCapacityResourceClient) CancelCapacityReservationWithContext(_ aws.Context, _ *ec2.CancelCapacityReservationInput, _ ...request.Option) (*ec2.CancelCapacityReservationOutput, error) {
	return &m.CancelCapacityReservationOutput, m.CancelCapacityReservationErr
}
