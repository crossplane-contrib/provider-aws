/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errors

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

const (
	errBoom = "boom"
	errMsg  = "example err msg"
)

func TestWrap(t *testing.T) {
	rootErr := &smithy.GenericAPIError{
		Code:    "InvalidVpcID.NotFound",
		Message: "The vpc ID 'vpc-06f35a4eaed9b4609' does not exist",
		Fault:   smithy.FaultUnknown,
	}
	cases := map[string]struct {
		reason string
		arg    error
		want   error
	}{
		"Nil": {
			arg:  nil,
			want: nil,
		},
		"NonAWSError": {
			reason: "It should not change anything if the error is not coming from AWS",
			arg:    errors.New(errBoom),
			want:   errors.Wrap(errors.New(errBoom), errMsg),
		},
		"AWSError": {
			reason: "Request ID should be removed from the final error if it's an AWS error",
			arg: &smithy.OperationError{
				ServiceID:     "EC2",
				OperationName: "DescribeVpcs",
				Err: &http.ResponseError{
					ResponseError: &smithyhttp.ResponseError{
						Err: rootErr,
					},
					RequestID: "c3dc34d4-b9d6-42a1-9909-7e8f62c6b9cc",
				},
			},
			want: errors.Wrap(rootErr, errMsg),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := Wrap(tc.arg, errMsg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}
