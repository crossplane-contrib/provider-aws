package fake

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

// MockRDSClient is a type that implements of some methods for RDSAPI interface
type MockRDSClient struct {
	rdsiface.RDSAPI
	MockCreateDBInstanceReadReplicaWithContext func(
		ctx context.Context,
		input *svcsdk.CreateDBInstanceReadReplicaInput,
		optFns ...request.Option,
	) (*svcsdk.CreateDBInstanceReadReplicaOutput, error)
	MockCreateDBInstanceWithContext func(
		ctx context.Context,
		input *svcsdk.CreateDBInstanceInput,
		optFns ...request.Option,
	) (*svcsdk.CreateDBInstanceOutput, error)
	MockDescribeDBInstancesWithContext func(
		ctx context.Context,
		input *svcsdk.DescribeDBInstancesInput,
		optFns ...request.Option,
	) (*svcsdk.DescribeDBInstancesOutput, error)
}

// CreateDBInstanceReadReplicaWithContext mocks CreateDBInstanceReadReplicaWithContext method for aws-sdk client
func (m *MockRDSClient) CreateDBInstanceReadReplicaWithContext(
	ctx context.Context,
	input *svcsdk.CreateDBInstanceReadReplicaInput,
	optFns ...request.Option,
) (*svcsdk.CreateDBInstanceReadReplicaOutput, error) {
	return m.MockCreateDBInstanceReadReplicaWithContext(ctx, input, optFns...)
}

// CreateDBInstanceWithContext mocks CreateDBInstanceWithContext method for aws-sdk client
func (m *MockRDSClient) CreateDBInstanceWithContext(
	ctx context.Context,
	input *svcsdk.CreateDBInstanceInput,
	optFns ...request.Option,
) (*svcsdk.CreateDBInstanceOutput, error) {
	return m.MockCreateDBInstanceWithContext(ctx, input, optFns...)
}

// DescribeDBInstancesWithContext mocks DescribeDBInstancesWithContext method for aws-sdk client
func (m *MockRDSClient) DescribeDBInstancesWithContext(
	ctx context.Context,
	input *svcsdk.DescribeDBInstancesInput,
	optFns ...request.Option,
) (*svcsdk.DescribeDBInstancesOutput, error) {
	return m.MockDescribeDBInstancesWithContext(ctx, input, optFns...)
}
