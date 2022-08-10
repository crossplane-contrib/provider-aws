package ec2

import (
	"errors"

	"github.com/aws/smithy-go"
)

const (
	// FlowLogNotFound is the code that is returned by ec2 when the given FlowLogId is not valid
	FlowLogNotFound = "InvalidFlowLogId.NotFound"
)

// IsFlowLogsNotFoundErr returns true if the error is because the item doesn't exist
func IsFlowLogsNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == FlowLogNotFound
}
