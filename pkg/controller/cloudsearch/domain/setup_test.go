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

package domain

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	test "github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudsearch/v1alpha1"
	fake "github.com/crossplane-contrib/provider-aws/pkg/clients/cloudsearch/fake"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

var domainName string = "test-domain-name"

func TestLateInitialize(t *testing.T) {

	type args struct {
		spec                svcapitypes.CustomDomainParameters
		statusScaling       cloudsearch.ScalingParameters
		statusPolicies      cloudsearch.AccessPoliciesStatus
		statusScalingError  error
		statusPoliciesError error
	}

	type want struct {
		result svcapitypes.CustomDomainParameters
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NothingToInitialize": {
			args: args{
				spec: svcapitypes.CustomDomainParameters{
					DesiredReplicationCount: pointer.ToIntAsInt64(4),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue("small"),
					DesiredPartitionCount:   pointer.ToIntAsInt64(0),
					AccessPolicies: pointer.ToOrNilIfZeroValue(`{
						"Version": "2012-10-17",
						"Statement": [
						  {
							"Effect": "Allow",
							"Principal": "*",
							"Action": "cloudsearch:*"
						  }
						]
					  }`),
				},
				statusScaling: cloudsearch.ScalingParameters{
					DesiredPartitionCount:   pointer.ToIntAsInt64(4),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue("small"),
					DesiredReplicationCount: pointer.ToIntAsInt64(0),
				},
				statusPolicies: cloudsearch.AccessPoliciesStatus{
					Options: pointer.ToOrNilIfZeroValue(""),
					Status: &cloudsearch.OptionStatus{
						PendingDeletion: pointer.ToOrNilIfZeroValue(false),
					},
				},
				statusScalingError:  nil,
				statusPoliciesError: nil,
			},
			want: want{
				result: svcapitypes.CustomDomainParameters{
					DesiredReplicationCount: pointer.ToIntAsInt64(4),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue("small"),
					DesiredPartitionCount:   pointer.ToIntAsInt64(0),
					AccessPolicies: pointer.ToOrNilIfZeroValue(`{
						"Version": "2012-10-17",
						"Statement": [
						  {
							"Effect": "Allow",
							"Principal": "*",
							"Action": "cloudsearch:*"
						  }
						]
					  }`),
				},
				err: nil,
			},
		},
		"NoSpec": {
			args: args{
				spec: svcapitypes.CustomDomainParameters{},
				statusScaling: cloudsearch.ScalingParameters{
					DesiredPartitionCount:   pointer.ToIntAsInt64(0),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue(""),
					DesiredReplicationCount: pointer.ToIntAsInt64(0),
				},
				statusPolicies: cloudsearch.AccessPoliciesStatus{
					Options: pointer.ToOrNilIfZeroValue(""),
					Status: &cloudsearch.OptionStatus{
						PendingDeletion: pointer.ToOrNilIfZeroValue(false),
					},
				},
				statusScalingError:  nil,
				statusPoliciesError: nil,
			},
			want: want{
				result: svcapitypes.CustomDomainParameters{
					DesiredReplicationCount: pointer.ToIntAsInt64(0),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue(""),
					DesiredPartitionCount:   pointer.ToIntAsInt64(0),
					AccessPolicies:          pointer.ToOrNilIfZeroValue(""),
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := &hooks{
				client: &fake.MockCloudsearchClient{
					MockDescribeServiceAccessPolicies: func(*cloudsearch.DescribeServiceAccessPoliciesInput) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error) {
						return &cloudsearch.DescribeServiceAccessPoliciesOutput{
							AccessPolicies: &tc.args.statusPolicies,
						}, tc.args.statusPoliciesError
					},
					MockDescribeScalingParameters: func(*cloudsearch.DescribeScalingParametersInput) (*cloudsearch.DescribeScalingParametersOutput, error) {
						return &cloudsearch.DescribeScalingParametersOutput{
							ScalingParameters: &cloudsearch.ScalingParametersStatus{
								Options: &tc.statusScaling,
								Status:  &cloudsearch.OptionStatus{PendingDeletion: pointer.ToOrNilIfZeroValue(false)},
							},
						}, tc.args.statusScalingError
					},
				},
			}

			result := tc.args.spec
			err := h.lateInitialize(&svcapitypes.DomainParameters{
				DomainName:             &domainName,
				CustomDomainParameters: result,
			}, &cloudsearch.DescribeDomainsOutput{})
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

		})
	}

}

func TestIsUpToDate(t *testing.T) {

	defaultPolicy := pointer.ToOrNilIfZeroValue(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": "*",
			"Action": "cloudsearch:*"
		  }
		]
	  }`)

	defaultScalingParameters := cloudsearch.ScalingParameters{
		DesiredPartitionCount:   pointer.ToIntAsInt64(2),
		DesiredInstanceType:     pointer.ToOrNilIfZeroValue("small"),
		DesiredReplicationCount: pointer.ToIntAsInt64(1),
	}

	type args struct {
		policySpec      *string
		policyStatus    *string
		policyStatusErr error

		scalingSpec                  cloudsearch.ScalingParameters
		scalingStatus                cloudsearch.ScalingParameters
		scalingStatusPendingDeletion bool
		scalingStatusErr             error

		requiresIndexing bool
	}

	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args
		want
	}{
		"AlreadyUpToDate": {
			args: args{
				policySpec:                   defaultPolicy,
				policyStatus:                 defaultPolicy,
				policyStatusErr:              nil,
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"UpdateNeededAccessPolicy": {
			args: args{
				policySpec: defaultPolicy,
				policyStatus: pointer.ToOrNilIfZeroValue(`{
					"Version": "2012-10-17",
					"Statement": [
					  {
						"Effect": "Allow",
						"Principal": "*",
						"Action": "cloudsearch:document"
					  }
					]
				  }`),
				policyStatusErr:              nil,
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"UpdateNeededAccessPolicyEmpty": {
			args: args{
				policySpec:                   pointer.ToOrNilIfZeroValue(" "),
				policyStatus:                 defaultPolicy,
				policyStatusErr:              nil,
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"AlreadyUpToDateWithPolicyError": {
			args: args{
				policySpec:                   defaultPolicy,
				policyStatus:                 defaultPolicy,
				policyStatusErr:              errors.New("InternalException"),
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: false,
				err:        errors.Wrap(errors.New("InternalException"), "cannot retrieve service access policies for Domain in AWS"),
			},
		},
		"UpdateNeededScalingParameters": {
			args: args{
				policySpec:      defaultPolicy,
				policyStatus:    defaultPolicy,
				policyStatusErr: nil,
				scalingSpec:     defaultScalingParameters,
				scalingStatus: cloudsearch.ScalingParameters{
					DesiredPartitionCount:   pointer.ToIntAsInt64(1),
					DesiredInstanceType:     pointer.ToOrNilIfZeroValue("small"),
					DesiredReplicationCount: pointer.ToIntAsInt64(1),
				},
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"AlreadyUpToDateWithScalingError": {
			args: args{
				policySpec:                   defaultPolicy,
				policyStatus:                 defaultPolicy,
				policyStatusErr:              nil,
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             errors.New("InternalException"),
				requiresIndexing:             false,
			},
			want: want{
				isUpToDate: false,
				err:        errors.Wrap(errors.New("InternalException"), "cannot retrieve scaling parameters for Domain in AWS"),
			},
		},
		"UpdateNeededIndexing": {
			args: args{
				policySpec:                   defaultPolicy,
				policyStatus:                 defaultPolicy,
				policyStatusErr:              nil,
				scalingSpec:                  defaultScalingParameters,
				scalingStatus:                defaultScalingParameters,
				scalingStatusPendingDeletion: false,
				scalingStatusErr:             nil,
				requiresIndexing:             true,
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := &hooks{
				client: &fake.MockCloudsearchClient{
					MockDescribeServiceAccessPoliciesWithContext: func(_ context.Context, in *cloudsearch.DescribeServiceAccessPoliciesInput, _ ...request.Option) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error) {
						return &cloudsearch.DescribeServiceAccessPoliciesOutput{
							AccessPolicies: &cloudsearch.AccessPoliciesStatus{
								Options: tc.args.policyStatus,
								Status:  &cloudsearch.OptionStatus{},
							},
						}, tc.args.policyStatusErr
					},
					MockDescribeScalingParametersWithContext: func(_ context.Context, in *cloudsearch.DescribeScalingParametersInput, _ ...request.Option) (*cloudsearch.DescribeScalingParametersOutput, error) {
						return &cloudsearch.DescribeScalingParametersOutput{
							ScalingParameters: &cloudsearch.ScalingParametersStatus{
								Options: &tc.args.scalingStatus,
								Status: &cloudsearch.OptionStatus{
									PendingDeletion: pointer.ToOrNilIfZeroValue(tc.args.scalingStatusPendingDeletion),
								},
							},
						}, tc.args.scalingStatusErr
					},
				},
			}

			result, _, err := h.isUpToDate(context.Background(), &svcapitypes.Domain{
				Spec: svcapitypes.DomainSpec{
					ForProvider: svcapitypes.DomainParameters{
						DomainName: &domainName,
						CustomDomainParameters: svcapitypes.CustomDomainParameters{
							AccessPolicies:          tc.args.policySpec,
							DesiredReplicationCount: tc.args.scalingSpec.DesiredReplicationCount,
							DesiredInstanceType:     tc.args.scalingSpec.DesiredInstanceType,
							DesiredPartitionCount:   tc.args.scalingSpec.DesiredPartitionCount,
						},
					},
				},
			}, &cloudsearch.DescribeDomainsOutput{
				DomainStatusList: []*cloudsearch.DomainStatus{{
					Created:                pointer.ToOrNilIfZeroValue(true),
					Deleted:                pointer.ToOrNilIfZeroValue(false),
					RequiresIndexDocuments: &tc.requiresIndexing,
					DomainName:             &domainName,
				}},
			})
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.isUpToDate, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

		})
	}
}
