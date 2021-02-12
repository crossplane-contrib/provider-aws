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
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
)

var (
	readCapacityUnits  = 1
	writeCapacityUnits = 1

	arn = "some arn"
)

func tableParams(m ...func(*v1alpha1.TableParameters)) *v1alpha1.TableParameters {
	o := &v1alpha1.TableParameters{
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

func table(m ...func(*svcsdk.TableDescription)) *svcsdk.TableDescription {
	o := &svcsdk.TableDescription{
		TableArn: &arn,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestCreatePatch(t *testing.T) {
	type args struct {
		t *svcsdk.DescribeTableOutput
		p *v1alpha1.TableParameters
	}

	type want struct {
		patch *v1alpha1.TableParameters
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
				p: &v1alpha1.TableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
					},
				},
			},
			want: want{
				patch: &v1alpha1.TableParameters{},
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
				p: &v1alpha1.TableParameters{
					ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits + 1)),
						WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits + 1)),
					},
				},
			},
			want: want{
				patch: &v1alpha1.TableParameters{
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
			result, _ := createPatch(tc.args.t, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		t svcsdk.DescribeTableOutput
		p v1alpha1.Table
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
				p: v1alpha1.Table{
					Spec: v1alpha1.TableSpec{
						ForProvider: v1alpha1.TableParameters{
							ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
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
				p: v1alpha1.Table{
					Spec: v1alpha1.TableSpec{
						ForProvider: v1alpha1.TableParameters{
							ProvisionedThroughput: &v1alpha1.ProvisionedThroughput{
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := isUpToDate(&tc.args.p, &tc.args.t)
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		spec *v1alpha1.TableParameters
		in   *svcsdk.DescribeTableOutput
	}
	type want struct {
		spec *v1alpha1.TableParameters
		err  error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: tableParams(),
				in: &svcsdk.DescribeTableOutput{
					Table: table(),
				},
			},
			want: want{spec: tableParams()},
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: tableParams(),
				in: &svcsdk.DescribeTableOutput{
					Table: table(func(t *svcsdk.TableDescription) {
						t.ItemCount = aws.Int64(1)
					}),
				},
			},
			want: want{spec: tableParams()},
		},
		"PartialFilled": {
			args: args{
				spec: tableParams(func(p *v1alpha1.TableParameters) {
					p.ProvisionedThroughput = nil
				}),
				in: &svcsdk.DescribeTableOutput{
					Table: table(func(t *svcsdk.TableDescription) {
						t.ProvisionedThroughput = &svcsdk.ProvisionedThroughputDescription{
							ReadCapacityUnits:  aws.Int64(int64(readCapacityUnits)),
							WriteCapacityUnits: aws.Int64(int64(writeCapacityUnits)),
						}
					}),
				},
			},
			want: want{spec: tableParams()},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := lateInitialize(tc.args.spec, tc.args.in)
			if diff := cmp.Diff(err, tc.want.err); diff != "" {
				t.Errorf("lateInitialize(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.args.spec, tc.want.spec); diff != "" {
				t.Errorf("lateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}
