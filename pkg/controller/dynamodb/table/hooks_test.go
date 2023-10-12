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
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	kmstypes "github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	mockkms "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/kmsiface"
)

var (
	readCapacityUnits  = 1
	writeCapacityUnits = 1
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
