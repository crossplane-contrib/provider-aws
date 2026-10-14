/*
Copyright 2019 The Crossplane Authors.

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

package table

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	kmstypes "github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	mockkms "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/kmsiface"
)

var (
	readCapacityUnits  = 1
	writeCapacityUnits = 1

	errListTagsFailed = errors.New("ListTagsOfResource boom")
)

type kmsAPIModifier func(mock *mockkms.MockKMSAPI)

func TestCreatePatch(t *testing.T) {
	type args struct {
		kmsClient kmsAPIModifier
		t         *svcsdk.DescribeTableOutput
		p         *svcapitypes.TableParameters
	}

	type want struct {
		patch *svcapitypes.TableParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				t: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
							WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
						},
					},
				},
				p: &svcapitypes.TableParameters{
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
			},
			want: want{
				patch: &svcapitypes.TableParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				t: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
							WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
						},
					},
				},
				p: &svcapitypes.TableParameters{
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
			want: want{
				patch: &svcapitypes.TableParameters{
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
		},
		"SameKMSMasterKeyButDifferentIDs": {
			args: args{
				kmsClient: func(mock *mockkms.MockKMSAPI) {
					mock.EXPECT().DescribeKeyWithContext(context.Background(), &kmstypes.DescribeKeyInput{
						KeyId: ptr.To("alias/test-key"),
					}).Return(&kmstypes.DescribeKeyOutput{
						KeyMetadata: &kmstypes.KeyMetadata{
							Arn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					}, nil)
				},
				t: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						SSEDescription: &svcsdk.SSEDescription{
							KMSMasterKeyArn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					},
				},
				p: &svcapitypes.TableParameters{
					SSESpecification: &svcapitypes.SSESpecification{
						KMSMasterKeyID: ptr.To("alias/test-key"),
					},
				},
			},
			want: want{
				patch: &svcapitypes.TableParameters{},
			},
		},
		"DifferentKMSMasterKeyIDs": {
			args: args{
				kmsClient: func(mock *mockkms.MockKMSAPI) {
					mock.EXPECT().DescribeKeyWithContext(context.Background(), &kmstypes.DescribeKeyInput{
						KeyId: ptr.To("alias/test-key"),
					}).Return(&kmstypes.DescribeKeyOutput{
						KeyMetadata: &kmstypes.KeyMetadata{
							Arn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-bbbb-bbbbbbbb"),
						},
					}, nil)
				},
				t: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						SSEDescription: &svcsdk.SSEDescription{
							KMSMasterKeyArn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					},
				},
				p: &svcapitypes.TableParameters{
					SSESpecification: &svcapitypes.SSESpecification{
						KMSMasterKeyID: ptr.To("alias/test-key"),
					},
				},
			},
			want: want{
				patch: &svcapitypes.TableParameters{
					SSESpecification: &svcapitypes.SSESpecification{
						KMSMasterKeyID: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-bbbb-bbbbbbbb"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockKms := mockkms.NewMockKMSAPI(gomock.NewController(t))
			if tc.args.kmsClient != nil {
				tc.args.kmsClient(mockKms)
			}
			updater := updateClient{
				clientkms: mockKms,
			}

			result, _ := updater.createPatch(context.Background(), tc.args.t, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsCoreResourceUpToDate(t *testing.T) {
	type args struct {
		kmsClient kmsAPIModifier
		t         svcsdk.DescribeTableOutput
		p         svcapitypes.Table
	}

	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				t: svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
							WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
						},
					},
				},
				p: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
								ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
								WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
							},
						},
					},
				},
			},
			want: want{
				result: true,
			},
		},
		"DifferentFields": {
			args: args{
				t: svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
							WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
						},
					},
				},
				p: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
								ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
								WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
							},
						},
					},
				},
			},
			want: want{
				result: false,
			},
		},
		"SameKMSMasterKeyButDifferentIDs": {
			args: args{
				kmsClient: func(mock *mockkms.MockKMSAPI) {
					mock.EXPECT().DescribeKeyWithContext(context.Background(), &kmstypes.DescribeKeyInput{
						KeyId: ptr.To("alias/test-key"),
					}).Return(&kmstypes.DescribeKeyOutput{
						KeyMetadata: &kmstypes.KeyMetadata{
							Arn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					}, nil)
				},
				t: svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						SSEDescription: &svcsdk.SSEDescription{
							KMSMasterKeyArn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					},
				},
				p: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							SSESpecification: &svcapitypes.SSESpecification{
								KMSMasterKeyID: ptr.To("alias/test-key"),
							},
						},
					},
				},
			},
			want: want{
				result: true,
			},
		},
		"DifferentKMSMasterKeyIDs": {
			args: args{
				kmsClient: func(mock *mockkms.MockKMSAPI) {
					mock.EXPECT().DescribeKeyWithContext(context.Background(), &kmstypes.DescribeKeyInput{
						KeyId: ptr.To("alias/test-key"),
					}).Return(&kmstypes.DescribeKeyOutput{
						KeyMetadata: &kmstypes.KeyMetadata{
							Arn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-bbbb-bbbbbbbb"),
						},
					}, nil)
				},
				t: svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						SSEDescription: &svcsdk.SSEDescription{
							KMSMasterKeyArn: ptr.To("arn:aws:kms:us-east-1:123456789123:key/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa"),
						},
					},
				},
				p: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							SSESpecification: &svcapitypes.SSESpecification{
								KMSMasterKeyID: ptr.To("alias/test-key"),
							},
						},
					},
				},
			},
			want: want{
				result: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockKms := mockkms.NewMockKMSAPI(gomock.NewController(t))
			if tc.args.kmsClient != nil {
				tc.args.kmsClient(mockKms)
			}
			updater := updateClient{
				clientkms: mockKms,
			}

			got, err := updater.isCoreResourceUpToDate(context.Background(), &tc.args.p, &tc.args.t)
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsPitrUpToDate(t *testing.T) {
	type args struct {
		t              svcapitypes.Table
		pitrStatusBool bool
	}

	type want struct {
		result bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				t: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							PointInTimeRecoveryEnabled: aws.Bool(true),
						},
					},
				},
				pitrStatusBool: true,
			},
			want: want{
				result: true,
			},
		},
		"DifferentFields": {
			args: args{
				t: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							PointInTimeRecoveryEnabled: aws.Bool(false),
						},
					},
				},
				pitrStatusBool: true,
			},
			want: want{
				result: false,
			},
		},
		"UnsetButTrueInAws": {
			args: args{
				t: svcapitypes.Table{
					Spec: svcapitypes.TableSpec{
						ForProvider: svcapitypes.TableParameters{
							PointInTimeRecoveryEnabled: nil,
						},
					},
				},
				pitrStatusBool: true,
			},
			want: want{
				result: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isPitrUpToDate(&tc.args.t, tc.args.pitrStatusBool)
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		p  *svcapitypes.TableParameters
		in *svcsdk.DescribeTableOutput
	}
	type want struct {
		p   *svcapitypes.TableParameters
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NilOutput": {
			args: args{
				p: &svcapitypes.TableParameters{},
			},
			want: want{
				p: &svcapitypes.TableParameters{},
			},
		},
		"ImpliedValues": {
			args: args{
				p: &svcapitypes.TableParameters{},
				in: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{},
				},
			},
			want: want{
				p: &svcapitypes.TableParameters{
					BillingMode:         aws.String(svcsdk.BillingModeProvisioned),
					StreamSpecification: &svcapitypes.StreamSpecification{StreamEnabled: aws.Bool(false)},
				},
			},
		},
		"EmptyParams": {
			args: args{
				p: &svcapitypes.TableParameters{},
				in: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						AttributeDefinitions: []*svcsdk.AttributeDefinition{{
							AttributeName: aws.String("N"),
							AttributeType: aws.String("T"),
						}},
						GlobalSecondaryIndexes: []*svcsdk.GlobalSecondaryIndexDescription{{
							IndexName: aws.String("cool-index"),
						}},
						LocalSecondaryIndexes: []*svcsdk.LocalSecondaryIndexDescription{{
							IndexName: aws.String("cool-index"),
						}},
						KeySchema: []*svcsdk.KeySchemaElement{{
							AttributeName: aws.String("N"),
							KeyType:       aws.String("T"),
						}},
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(42),
							WriteCapacityUnits: aws.Int64(42),
						},
						SSEDescription: &svcsdk.SSEDescription{
							Status:          aws.String(string(svcapitypes.SSEStatus_ENABLED)),
							KMSMasterKeyArn: aws.String("some-arn"),
							SSEType:         aws.String("very-secure"),
						},
						StreamSpecification: &svcsdk.StreamSpecification{
							StreamEnabled:  aws.Bool(true),
							StreamViewType: aws.String("the-good-type"),
						},
						BillingModeSummary: &svcsdk.BillingModeSummary{
							BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
						},
					},
				},
			},
			want: want{
				p: &svcapitypes.TableParameters{
					BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
					AttributeDefinitions: []*svcapitypes.AttributeDefinition{{
						AttributeName: aws.String("N"),
						AttributeType: aws.String("T"),
					}},
					GlobalSecondaryIndexes: []*svcapitypes.GlobalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					LocalSecondaryIndexes: []*svcapitypes.LocalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					KeySchema: []*svcapitypes.KeySchemaElement{{
						AttributeName: aws.String("N"),
						KeyType:       aws.String("T"),
					}},
					// ProvisionedThroughput is NOT late-initialized when
					// BillingMode is PAY_PER_REQUEST. AWS returns zeros for
					// on-demand tables which would cause CreateTable to fail
					// if the table needs to be recreated.
					SSESpecification: &svcapitypes.SSESpecification{
						Enabled:        aws.Bool(true),
						KMSMasterKeyID: aws.String("some-arn"),
						SSEType:        aws.String("very-secure"),
					},
					StreamSpecification: &svcapitypes.StreamSpecification{
						StreamEnabled:  aws.Bool(true),
						StreamViewType: aws.String("the-good-type"),
					},
				},
			},
		},
		"EmptyParamsProvisioned": {
			args: args{
				p: &svcapitypes.TableParameters{},
				in: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						AttributeDefinitions: []*svcsdk.AttributeDefinition{{
							AttributeName: aws.String("N"),
							AttributeType: aws.String("T"),
						}},
						KeySchema: []*svcsdk.KeySchemaElement{{
							AttributeName: aws.String("N"),
							KeyType:       aws.String("T"),
						}},
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(5),
							WriteCapacityUnits: aws.Int64(5),
						},
					},
				},
			},
			want: want{
				p: &svcapitypes.TableParameters{
					BillingMode: aws.String(svcsdk.BillingModeProvisioned),
					AttributeDefinitions: []*svcapitypes.AttributeDefinition{{
						AttributeName: aws.String("N"),
						AttributeType: aws.String("T"),
					}},
					KeySchema: []*svcapitypes.KeySchemaElement{{
						AttributeName: aws.String("N"),
						KeyType:       aws.String("T"),
					}},
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
					StreamSpecification: &svcapitypes.StreamSpecification{StreamEnabled: aws.Bool(false)},
				},
			},
		},
		"ExistingParams": {
			args: args{
				p: &svcapitypes.TableParameters{
					BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
					AttributeDefinitions: []*svcapitypes.AttributeDefinition{{
						AttributeName: aws.String("N"),
						AttributeType: aws.String("T"),
					}},
					GlobalSecondaryIndexes: []*svcapitypes.GlobalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					LocalSecondaryIndexes: []*svcapitypes.LocalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					KeySchema: []*svcapitypes.KeySchemaElement{{
						AttributeName: aws.String("N"),
						KeyType:       aws.String("T"),
					}},
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(42),
						WriteCapacityUnits: aws.Int64(42),
					},
					SSESpecification: &svcapitypes.SSESpecification{
						Enabled:        aws.Bool(true),
						KMSMasterKeyID: aws.String("some-arn"),
						SSEType:        aws.String("very-secure"),
					},
					StreamSpecification: &svcapitypes.StreamSpecification{
						StreamEnabled:  aws.Bool(true),
						StreamViewType: aws.String("the-good-type"),
					},
				},
				in: &svcsdk.DescribeTableOutput{
					Table: &svcsdk.TableDescription{
						AttributeDefinitions: []*svcsdk.AttributeDefinition{{
							AttributeName: aws.String("X"),
							AttributeType: aws.String("Y"),
						}},
						GlobalSecondaryIndexes: []*svcsdk.GlobalSecondaryIndexDescription{{
							IndexName: aws.String("cooler-index"),
						}},
						LocalSecondaryIndexes: []*svcsdk.LocalSecondaryIndexDescription{{
							IndexName: aws.String("cooler-index"),
						}},
						KeySchema: []*svcsdk.KeySchemaElement{{
							AttributeName: aws.String("X"),
							KeyType:       aws.String("Y"),
						}},
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(24),
							WriteCapacityUnits: aws.Int64(24),
						},
						SSEDescription: &svcsdk.SSEDescription{
							Status:          aws.String(string(svcapitypes.SSEStatus_DISABLED)),
							KMSMasterKeyArn: aws.String("some-other-arn"),
							SSEType:         aws.String("kinda-secure"),
						},
						StreamSpecification: &svcsdk.StreamSpecification{
							StreamEnabled:  aws.Bool(false),
							StreamViewType: aws.String("the-other-type"),
						},
						BillingModeSummary: &svcsdk.BillingModeSummary{
							BillingMode: aws.String(svcsdk.BillingModeProvisioned),
						},
					},
				},
			},
			want: want{
				p: &svcapitypes.TableParameters{
					BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
					AttributeDefinitions: []*svcapitypes.AttributeDefinition{{
						AttributeName: aws.String("N"),
						AttributeType: aws.String("T"),
					}},
					GlobalSecondaryIndexes: []*svcapitypes.GlobalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					LocalSecondaryIndexes: []*svcapitypes.LocalSecondaryIndex{{
						IndexName: aws.String("cool-index"),
					}},
					KeySchema: []*svcapitypes.KeySchemaElement{{
						AttributeName: aws.String("N"),
						KeyType:       aws.String("T"),
					}},
					ProvisionedThroughput: &svcapitypes.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(42),
						WriteCapacityUnits: aws.Int64(42),
					},
					SSESpecification: &svcapitypes.SSESpecification{
						Enabled:        aws.Bool(true),
						KMSMasterKeyID: aws.String("some-arn"),
						SSEType:        aws.String("very-secure"),
					},
					StreamSpecification: &svcapitypes.StreamSpecification{
						StreamEnabled:  aws.Bool(true),
						StreamViewType: aws.String("the-good-type"),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := lateInitialize(tc.args.p, tc.args.in)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("lateInitialize(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.p, tc.args.p); diff != "" {
				t.Errorf("lateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffGlobalSecondaryIndexes(t *testing.T) {
	type args struct {
		spec []*svcsdk.GlobalSecondaryIndexDescription
		obs  []*svcsdk.GlobalSecondaryIndexDescription
	}
	type want struct {
		result []*svcsdk.GlobalSecondaryIndexUpdate
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NoOp": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("one"),
					},
				},
				obs: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("one"),
					},
				},
			},
		},
		"Create": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("newone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(10),
							WriteCapacityUnits: aws.Int64(10),
						},
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Create: &svcsdk.CreateGlobalSecondaryIndexAction{
							IndexName: aws.String("newone"),
							ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
								ReadCapacityUnits:  aws.Int64(10),
								WriteCapacityUnits: aws.Int64(10),
							},
						},
					},
				},
			},
		},
		"CreateOnlyOne": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("newone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(10),
							WriteCapacityUnits: aws.Int64(10),
						},
					},
					{
						IndexName: aws.String("secondnewone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(10),
							WriteCapacityUnits: aws.Int64(10),
						},
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Create: &svcsdk.CreateGlobalSecondaryIndexAction{
							IndexName: aws.String("newone"),
							ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
								ReadCapacityUnits:  aws.Int64(10),
								WriteCapacityUnits: aws.Int64(10),
							},
						},
					},
				},
			},
		},
		"AddNewToExisting": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("newone"),
					},
				},
				obs: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("oldone"),
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Create: &svcsdk.CreateGlobalSecondaryIndexAction{
							IndexName: aws.String("newone"),
						},
					},
				},
			},
		},
		"UpdateExistingOnes": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("newone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits: aws.Int64(20),
						},
					},
					{
						IndexName: aws.String("oldone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits: aws.Int64(20),
						},
					},
				},
				obs: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("newone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits: aws.Int64(10),
						},
					},
					{
						IndexName: aws.String("oldone"),
						ProvisionedThroughput: &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits: aws.Int64(5),
						},
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Update: &svcsdk.UpdateGlobalSecondaryIndexAction{
							IndexName: aws.String("newone"),
							ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
								ReadCapacityUnits: aws.Int64(20),
							},
						},
					},
					{
						Update: &svcsdk.UpdateGlobalSecondaryIndexAction{
							IndexName: aws.String("oldone"),
							ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
								ReadCapacityUnits: aws.Int64(20),
							},
						},
					},
				},
			},
		},
		"Delete": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{},
				obs: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("oldone"),
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Delete: &svcsdk.DeleteGlobalSecondaryIndexAction{
							IndexName: aws.String("oldone"),
						},
					},
				},
			},
		},
		"DeleteOnlyOne": {
			args: args{
				spec: []*svcsdk.GlobalSecondaryIndexDescription{},
				obs: []*svcsdk.GlobalSecondaryIndexDescription{
					{
						IndexName: aws.String("oldone"),
					},
					{
						IndexName: aws.String("secondoldone"),
					},
				},
			},
			want: want{
				result: []*svcsdk.GlobalSecondaryIndexUpdate{
					{
						Delete: &svcsdk.DeleteGlobalSecondaryIndexAction{
							IndexName: aws.String("oldone"),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := diffGlobalSecondaryIndexes(tc.args.spec, tc.args.obs)
			if diff := cmp.Diff(got, tc.want.result); diff != "" {
				t.Errorf("diffGlobalSecondaryIndexes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPreCreate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.Table
		obj *svcsdk.CreateTableInput
	}
	type want struct {
		obj *svcsdk.CreateTableInput
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"PayPerRequestStripsProvisionedThroughput": {
			args: args{
				cr: &svcapitypes.Table{},
				obj: &svcsdk.CreateTableInput{
					BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
					ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(0),
						WriteCapacityUnits: aws.Int64(0),
					},
				},
			},
			want: want{
				obj: &svcsdk.CreateTableInput{
					BillingMode:           aws.String(svcsdk.BillingModePayPerRequest),
					ProvisionedThroughput: nil,
				},
			},
		},
		"ProvisionedKeepsProvisionedThroughput": {
			args: args{
				cr: &svcapitypes.Table{},
				obj: &svcsdk.CreateTableInput{
					BillingMode: aws.String(svcsdk.BillingModeProvisioned),
					ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
				},
			},
			want: want{
				obj: &svcsdk.CreateTableInput{
					BillingMode: aws.String(svcsdk.BillingModeProvisioned),
					ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(5),
						WriteCapacityUnits: aws.Int64(5),
					},
				},
			},
		},
		"NilBillingModeKeepsProvisionedThroughput": {
			args: args{
				cr: &svcapitypes.Table{},
				obj: &svcsdk.CreateTableInput{
					ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(1),
						WriteCapacityUnits: aws.Int64(1),
					},
				},
			},
			want: want{
				obj: &svcsdk.CreateTableInput{
					ProvisionedThroughput: &svcsdk.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(1),
						WriteCapacityUnits: aws.Int64(1),
					},
				},
			},
		},
		"PayPerRequestWithNilProvisionedThroughput": {
			args: args{
				cr: &svcapitypes.Table{},
				obj: &svcsdk.CreateTableInput{
					BillingMode: aws.String(svcsdk.BillingModePayPerRequest),
				},
			},
			want: want{
				obj: &svcsdk.CreateTableInput{
					BillingMode:           aws.String(svcsdk.BillingModePayPerRequest),
					ProvisionedThroughput: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := preCreate(context.Background(), tc.args.cr, tc.args.obj)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("preCreate(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obj, tc.args.obj); diff != "" {
				t.Errorf("preCreate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// fakeTagLister implements svcsdkapi.DynamoDBAPI just enough for tag tests.
// All methods not used by the tag helpers panic to surface unexpected calls.
type fakeTagLister struct {
	svcsdkapi.DynamoDBAPI // embed to satisfy the interface
	tags                  []*svcsdk.Tag
	err                   error
}

func (f *fakeTagLister) ListTagsOfResourceWithContext(_ aws.Context, _ *svcsdk.ListTagsOfResourceInput, _ ...request.Option) (*svcsdk.ListTagsOfResourceOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &svcsdk.ListTagsOfResourceOutput{Tags: f.tags}, nil
}

func TestSortDynamoDBTags(t *testing.T) {
	cases := map[string]struct {
		input []*svcsdk.Tag
		want  []*svcsdk.Tag
	}{
		"AlreadySorted": {
			input: []*svcsdk.Tag{
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("b"), Value: aws.String("2")},
			},
			want: []*svcsdk.Tag{
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("b"), Value: aws.String("2")},
			},
		},
		"ReverseOrder": {
			input: []*svcsdk.Tag{
				{Key: aws.String("z"), Value: aws.String("9")},
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("m"), Value: aws.String("5")},
			},
			want: []*svcsdk.Tag{
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("m"), Value: aws.String("5")},
				{Key: aws.String("z"), Value: aws.String("9")},
			},
		},
		"SameKeySortByValue": {
			input: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("env"), Value: aws.String("dev")},
			},
			want: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("dev")},
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
		},
		"EmptySlice": {
			input: []*svcsdk.Tag{},
			want:  []*svcsdk.Tag{},
		},
		"NilSlice": {
			input: nil,
			want:  []*svcsdk.Tag{},
		},
		"OriginalNotMutated": {
			// sortDynamoDBTags must not mutate the original slice
			input: []*svcsdk.Tag{
				{Key: aws.String("b"), Value: aws.String("2")},
				{Key: aws.String("a"), Value: aws.String("1")},
			},
			want: []*svcsdk.Tag{
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("b"), Value: aws.String("2")},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// keep a copy of the original to verify it wasn't mutated
			origLen := len(tc.input)
			origKeys := make([]string, origLen)
			for i, tag := range tc.input {
				if tag != nil && tag.Key != nil {
					origKeys[i] = *tag.Key
				}
			}

			got := sortDynamoDBTags(tc.input)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("sortDynamoDBTags(%v): -want, +got:\n%s", name, diff)
			}

			// verify original slice key order was not changed
			for i, tag := range tc.input {
				if tag != nil && tag.Key != nil && *tag.Key != origKeys[i] {
					t.Errorf("sortDynamoDBTags mutated the original slice at index %d", i)
				}
			}
		})
	}
}

func TestAreDynamoDBTagsEqual(t *testing.T) {
	tableArn := aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/test-table")

	cases := map[string]struct {
		observedTags []*svcsdk.Tag
		specTags     []*svcapitypes.Tag
		listErr      error
		wantEqual    bool
		wantErr      bool
	}{
		// Core behaviour: equal sets regardless of arrival order
		"EqualTagsAlreadySorted": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("team"), Value: aws.String("platform")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("team"), Value: aws.String("platform")},
			},
			wantEqual: true,
		},
		// KEY FIX: unsorted tags from AWS must still compare equal to sorted spec
		"EqualTagsUnsortedFromAWS": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("team"), Value: aws.String("platform")},
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("cost-center"), Value: aws.String("eng")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("cost-center"), Value: aws.String("eng")},
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("team"), Value: aws.String("platform")},
			},
			wantEqual: true,
		},
		// KEY FIX: unsorted spec must still match sorted AWS response
		"EqualTagsUnsortedSpec": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("b"), Value: aws.String("2")},
				{Key: aws.String("c"), Value: aws.String("3")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("c"), Value: aws.String("3")},
				{Key: aws.String("a"), Value: aws.String("1")},
				{Key: aws.String("b"), Value: aws.String("2")},
			},
			wantEqual: true,
		},
		"DifferentTagValues": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("staging")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			wantEqual: false,
		},
		"DifferentTagKeys": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("environment"), Value: aws.String("prod")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			wantEqual: false,
		},
		"ExtraTagOnAWS": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("extra"), Value: aws.String("val")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			wantEqual: false,
		},
		"ExtraTagInSpec": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("extra"), Value: aws.String("val")},
			},
			wantEqual: false,
		},
		"BothEmpty": {
			observedTags: []*svcsdk.Tag{},
			specTags:     []*svcapitypes.Tag{},
			wantEqual:    true,
		},
		"BothNil": {
			observedTags: nil,
			specTags:     nil,
			wantEqual:    true,
		},
		"ListTagsError": {
			listErr: errListTagsFailed,
			wantErr: true,
		},
		// AWS-managed tags (aws: prefix) must be ignored; their presence on the
		// observed side should not cause a false drift signal.
		"AWSManagedTagsIgnored": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("aws:cloudformation:stack-name"), Value: aws.String("my-stack")},
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			specTags: []*svcapitypes.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
			},
			wantEqual: true,
		},
		// Only aws: tags differ — no user-managed drift, must be equal.
		"OnlyAWSManagedTagsOnAWSSide": {
			observedTags: []*svcsdk.Tag{
				{Key: aws.String("aws:eks:cluster-name"), Value: aws.String("cluster")},
			},
			specTags:  []*svcapitypes.Tag{},
			wantEqual: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			u := &updateClient{
				client: &fakeTagLister{
					tags: tc.observedTags,
					err:  tc.listErr,
				},
			}

			got, err := u.areDynamoDBTagsEqual(context.Background(), tableArn, tc.specTags)
			if (err != nil) != tc.wantErr {
				t.Errorf("areDynamoDBTagsEqual() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if diff := cmp.Diff(tc.wantEqual, got); diff != "" {
					t.Errorf("areDynamoDBTagsEqual(): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
