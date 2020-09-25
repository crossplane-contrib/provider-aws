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
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/database/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/dynamodb"
	"github.com/crossplane/provider-aws/pkg/clients/dynamodb/fake"
)

const (
	providerName = "aws-creds"
)

var (
	errBoom = errors.New("boom")
)

type args struct {
	dynamo dynamodb.Client
	// kube   client.Client
	cr *v1alpha1.DynamoTable
}

type tableModifier func(*v1alpha1.DynamoTable)

func withConditions(c ...runtimev1alpha1.Condition) tableModifier {
	return func(r *v1alpha1.DynamoTable) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(s v1alpha1.DynamoTableObservation) tableModifier {
	return func(r *v1alpha1.DynamoTable) { r.Status.AtProvider = s }
}

func table(m ...tableModifier) *v1alpha1.DynamoTable {
	cr := &v1alpha1.DynamoTable{
		Spec: v1alpha1.DynamoTableSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &runtimev1alpha1.Reference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DynamoTable
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {

			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsdynamo.DescribeTableOutput{
								Table: &awsdynamo.TableDescription{
									TableStatus: v1alpha1.DynamoTableStateAvailable,
								},
							},
							},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(
					withStatus(v1alpha1.DynamoTableObservation{
						TableStatus: v1alpha1.DynamoTableStateAvailable,
					}),
					withConditions(runtimev1alpha1.Available()),
					withConditions(runtimev1alpha1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsdynamo.DescribeTableOutput{
								Table: &awsdynamo.TableDescription{
									TableStatus: v1alpha1.DynamoTableStateDeleting,
								},
							},
							},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(
					withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.DynamoTableObservation{
						TableStatus: v1alpha1.DynamoTableStateDeleting,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(),
				err: errors.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsdynamo.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DynamoTable
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockCreate: func(input *awsdynamo.CreateTableInput) awsdynamo.CreateTableRequest {
						return awsdynamo.CreateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsdynamo.CreateTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Creating())),
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateCreating,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateCreating,
				}), withConditions(runtimev1alpha1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockCreate: func(input *awsdynamo.CreateTableInput) awsdynamo.CreateTableRequest {
						return awsdynamo.CreateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DynamoTable
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockUpdate: func(input *awsdynamo.UpdateTableInput) awsdynamo.UpdateTableRequest {
						return awsdynamo.UpdateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsdynamo.UpdateTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateModifying,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateModifying,
				})),
			},
		},
		"FailedModify": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockUpdate: func(input *awsdynamo.UpdateTableInput) awsdynamo.UpdateTableRequest {
						return awsdynamo.UpdateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.DynamoTable
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsdynamo.DeleteTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateDeleting,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateDeleting,
				}),
					withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsdynamo.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
