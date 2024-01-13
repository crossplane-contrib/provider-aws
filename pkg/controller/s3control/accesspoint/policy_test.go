/*
Copyright 2021 The Crossplane Authors.

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

package accesspoint

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3control/fake"
	s3controlTesting "github.com/crossplane-contrib/provider-aws/pkg/controller/s3control/testing"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	policyV1 = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				SID:    pointer.ToOrNilIfZeroValue("AllowPublicRead"),
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AWSPrincipals: []common.AWSPrincipal{
						{IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::1234567890:role/sso/role")},
					},
				},
				Action:   []string{"s3:GetObject"},
				Resource: []string{"arn:aws:s3:::my-bucket/*"},
			},
		},
	}

	policyV2 = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				SID:    pointer.ToOrNilIfZeroValue("AllowPublicWrite"),
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AWSPrincipals: []common.AWSPrincipal{
						{IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::1234567890:role/sso/role")},
					},
				},
				Action:   []string{"s3:GetObject"},
				Resource: []string{"arn:aws:s3:::my-bucket/*"},
			},
		},
	}

	policyOutputV1 = s3control.GetAccessPointPolicyOutput{
		Policy: aws.String(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Sid": "AllowPublicRead",
							"Effect": "Allow",
							"Principal": {
								"AWS": "arn:aws:iam::1234567890:role/sso/role"
							},
							"Action": "s3:GetObject",
							"Resource": "arn:aws:s3:::my-bucket/*"
						}]
					}`),
	}
)

func TestPolicyClientObserve(t *testing.T) {
	testCases := []struct {
		name           string
		cr             *v1alpha1.AccessPoint
		mockClient     *fake.MockS3ControlClient
		expectedStatus policyStatus
		expectedError  error
	}{
		{
			name: "PolicyNeedsUpdateErrNotFound",
			cr: s3controlTesting.AccessPoint(
				s3controlTesting.WithPolicy(policyV1),
			),
			mockClient: &fake.MockS3ControlClient{
				GetAccessPointPolicyErr: s3controlTesting.NoSuchAccessPointPolicy(),
			},
			expectedStatus: needsUpdate,
		},
		{
			name: "PolicyNeedsUpdate",
			cr: s3controlTesting.AccessPoint(
				s3controlTesting.WithPolicy(policyV2),
			),
			mockClient: &fake.MockS3ControlClient{
				GetAccessPointPolicyOutput: policyOutputV1,
			},
			expectedStatus: needsUpdate,
		},
		{
			name: "PolicyUpdated",
			cr: s3controlTesting.AccessPoint(
				s3controlTesting.WithPolicy(policyV1),
			),
			mockClient: &fake.MockS3ControlClient{
				GetAccessPointPolicyOutput: policyOutputV1,
			},
			expectedStatus: updated,
		},
		{
			name: "NeedsDeletion",
			cr:   s3controlTesting.AccessPoint(),
			mockClient: &fake.MockS3ControlClient{
				GetAccessPointPolicyOutput: testOutputPolicyV1,
			},
			expectedStatus: needsDeletion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := &policyClient{
				client: tc.mockClient,
			}
			status, err := pc.observe(tc.cr)
			if diff := cmp.Diff(tc.expectedError, err, test.EquateErrors()); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
			if status != tc.expectedStatus {
				t.Errorf("expected status: %v, got: %v", tc.expectedStatus, status)
			}
		})
	}
}

func TestPolicyClientCreateOrUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	testCases := []struct {
		name          string
		cr            *v1alpha1.AccessPoint
		mockClient    *fake.MockS3ControlClient
		expectedError error
	}{
		{
			name: "SuccessfulCreateOrUpdate",
			cr: s3controlTesting.AccessPoint(
				s3controlTesting.WithPolicy(policyV1),
			),
			mockClient: &fake.MockS3ControlClient{
				PutAccessPointPolicyWithContextOutput: s3control.PutAccessPointPolicyOutput{},
			},
		},
		{
			name: "FailedCreateOrUpdate",
			cr: s3controlTesting.AccessPoint(
				s3controlTesting.WithPolicy(policyV1),
			),
			mockClient: &fake.MockS3ControlClient{
				PutAccessPointPolicyWithContextOutput: s3control.PutAccessPointPolicyOutput{},
				PutAccessPointPolicyWithContextErr:    errBoom,
			},
			expectedError: errors.Wrap(errBoom, errPutPolicy),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := &policyClient{
				client: tc.mockClient,
			}

			err := pc.createOrUpdate(context.Background(), tc.cr)
			if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestPolicyClientDelete(t *testing.T) {
	errBoom := errors.New("boom")
	testCases := []struct {
		name          string
		cr            *v1alpha1.AccessPoint
		mockClient    *fake.MockS3ControlClient
		expectedError error
	}{
		{
			name: "DeleteSuccessful",
			cr:   s3controlTesting.AccessPoint(),
			mockClient: &fake.MockS3ControlClient{
				DeleteAccessPointPolicyWithContextOutput: s3control.DeleteAccessPointPolicyOutput{},
			},
		},
		{
			name: "DeleteFailed",
			cr:   s3controlTesting.AccessPoint(),
			mockClient: &fake.MockS3ControlClient{
				DeleteAccessPointPolicyWithContextOutput: s3control.DeleteAccessPointPolicyOutput{},
				DeleteAccessPointPolicyWithContextErr:    errBoom,
			},
			expectedError: errors.Wrap(errBoom, errDeletePolicy),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pc := &policyClient{
				client: tc.mockClient,
			}

			err := pc.delete(context.Background(), tc.cr)
			if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}
