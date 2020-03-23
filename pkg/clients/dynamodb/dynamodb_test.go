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

package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/database/v1alpha1"
)

var (
	readCapacityUnits  = 1
	writeCapacityUnits = 1

	tableName = "some name"
	arn       = "some arn"
	tableID   = "some ID"
)

func addRoleOutputFields(r *dynamodb.TableDescription) {
	r.TableArn = aws.String(arn)
	r.TableId = aws.String(tableID)
}

func tableParams(m ...func(*v1alpha1.DynamoTableParameters)) *v1alpha1.DynamoTableParameters {
	o := &v1alpha1.DynamoTableParameters{
		ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
			WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
		},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func table(m ...func(*dynamodb.TableDescription)) *dynamodb.TableDescription {
	o := &dynamodb.TableDescription{
		TableArn: &arn,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func tableObservation(m ...func(*v1alpha1.DynamoTableObservation)) *v1alpha1.DynamoTableObservation {
	o := &v1alpha1.DynamoTableObservation{
		TableArn: arn,
		TableID:  tableID,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestCreatePatch(t *testing.T) {
	type args struct {
		t *dynamodb.TableDescription
		p *v1alpha1.DynamoTableParameters
	}

	type want struct {
		patch *v1alpha1.DynamoTableParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				t: &dynamodb.TableDescription{
					ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
				p: &v1alpha1.DynamoTableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
			},
			want: want{
				patch: &v1alpha1.DynamoTableParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				t: &dynamodb.TableDescription{
					ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
				p: &v1alpha1.DynamoTableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
			want: want{
				patch: &v1alpha1.DynamoTableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.t, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		t dynamodb.TableDescription
		p v1alpha1.DynamoTableParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				t: dynamodb.TableDescription{
					ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
				p: v1alpha1.DynamoTableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				t: dynamodb.TableDescription{
					ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
				p: v1alpha1.DynamoTableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.t)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRoleObservation(t *testing.T) {
	cases := map[string]struct {
		in  dynamodb.TableDescription
		out v1alpha1.DynamoTableObservation
	}{
		"AllFilled": {
			in:  *table(addRoleOutputFields),
			out: *tableObservation(),
		},
		"NoRoleId": {
			in: *table(addRoleOutputFields, func(r *dynamodb.TableDescription) {
				r.TableId = nil
			}),
			out: *tableObservation(func(o *v1alpha1.DynamoTableObservation) {
				o.TableID = ""
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateTableInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.DynamoTableParameters
		out dynamodb.CreateTableInput
	}{
		"FilledInput": {
			in: *tableParams(),
			out: dynamodb.CreateTableInput{
				TableName: aws.String(tableName),
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
					WriteCapacityUnits: aws.Int64(int64(readCapacityUnits)),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateTableInput(tableName, &tc.in)
			if diff := cmp.Diff(r, &tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
