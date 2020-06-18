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

package iamrole

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

const (
	providerName = "aws-creds"
	testRegion   = "us-east-1"
)

var (
	// an arbitrary managed resource
	unexpecedItem resource.Managed
	roleName      = "some arbitrary name"
	description   = "some description"
	policy        = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.RoleClient
	cr  resource.Managed
}

type roleModifier func(*v1beta1.IAMRole)

func withConditions(c ...corev1alpha1.Condition) roleModifier {
	return func(r *v1beta1.IAMRole) { r.Status.ConditionedStatus.Conditions = c }
}

func withRoleName(s *string) roleModifier {
	return func(r *v1beta1.IAMRole) { meta.SetExternalName(r, *s) }
}

func withPolicy() roleModifier {
	return func(r *v1beta1.IAMRole) {
		p, err := awsclients.CompactAndEscapeJSON(policy)
		if err != nil {
			return
		}
		r.Spec.ForProvider.AssumeRolePolicyDocument = p
	}
}

func withDescription() roleModifier {
	return func(r *v1beta1.IAMRole) {
		r.Spec.ForProvider.Description = aws.String(description)
	}
}

func role(m ...roleModifier) *v1beta1.IAMRole {
	cr := &v1beta1.IAMRole{
		Spec: v1beta1.IAMRoleSpec{
			ResourceSpec: corev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {

	type args struct {
		newClientFn func(*aws.Config) (iam.RoleClient, error)
		awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
		cr          resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				newClientFn: func(config *aws.Config) (iam.RoleClient, error) {
					if diff := cmp.Diff(testRegion, config.Region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				awsConfigFn: func(_ context.Context, _ client.Reader, p *corev1.ObjectReference) (*aws.Config, error) {
					if diff := cmp.Diff(providerName, p.Name); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return &aws.Config{Region: testRegion}, nil
				},
				cr: role(),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				err: errors.New(errUnexpectedObject),
			},
		},
		"ProviderFailure": {
			args: args{
				newClientFn: func(config *aws.Config) (iam.RoleClient, error) {
					if diff := cmp.Diff(testRegion, config.Region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, errBoom
				},
				awsConfigFn: func(_ context.Context, _ client.Reader, p *corev1.ObjectReference) (*aws.Config, error) {
					if diff := cmp.Diff(providerName, p.Name); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return &aws.Config{Region: testRegion}, nil
				},
				cr: role(),
			},
			want: want{
				err: errors.Wrap(errBoom, errClient),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{newClientFn: tc.newClientFn, awsConfigFn: tc.awsConfigFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.GetRoleOutput{
								Role: &awsiam.Role{},
							}},
						}
					},
				},
				cr: role(withRoleName(&roleName)),
			},
			want: want{
				cr: role(
					withRoleName(&roleName),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: role(withRoleName(&roleName)),
			},
			want: want{
				cr:  role(withRoleName(&roleName)),
				err: errors.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: role(),
			},
			want: want{
				cr: role(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				iam: &fake.MockRoleClient{
					MockCreateRoleRequest: func(input *awsiam.CreateRoleInput) awsiam.CreateRoleRequest {
						return awsiam.CreateRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.CreateRoleOutput{}},
						}
					},
				},
				cr: role(withRoleName(&roleName)),
			},
			want: want{
				cr: role(
					withRoleName(&roleName),
					withConditions(corev1alpha1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockRoleClient{
					MockCreateRoleRequest: func(input *awsiam.CreateRoleInput) awsiam.CreateRoleRequest {
						return awsiam.CreateRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: role(),
			},
			want: want{
				cr:  role(withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.GetRoleOutput{
								Role: &awsiam.Role{},
							}},
						}
					},
					MockUpdateRoleRequest: func(input *awsiam.UpdateRoleInput) awsiam.UpdateRoleRequest {
						return awsiam.UpdateRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.UpdateRoleOutput{}},
						}
					},
				},
				cr: role(withRoleName(&roleName)),
			},
			want: want{
				cr: role(withRoleName(&roleName)),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientUpdateRoleError": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.GetRoleOutput{
								Role: &awsiam.Role{},
							}},
						}
					},
					MockUpdateRoleRequest: func(input *awsiam.UpdateRoleInput) awsiam.UpdateRoleRequest {
						return awsiam.UpdateRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: role(withDescription()),
			},
			want: want{
				cr:  role(withDescription()),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
		"ClientUpdatePolicyError": {
			args: args{
				iam: &fake.MockRoleClient{
					MockGetRoleRequest: func(input *awsiam.GetRoleInput) awsiam.GetRoleRequest {
						return awsiam.GetRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.GetRoleOutput{
								Role: &awsiam.Role{},
							}},
						}
					},
					MockUpdateRoleRequest: func(input *awsiam.UpdateRoleInput) awsiam.UpdateRoleRequest {
						return awsiam.UpdateRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.UpdateRoleOutput{}},
						}
					},
					MockUpdateAssumeRolePolicyRequest: func(input *awsiam.UpdateAssumeRolePolicyInput) awsiam.UpdateAssumeRolePolicyRequest {
						return awsiam.UpdateAssumeRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: role(withPolicy()),
			},
			want: want{
				cr:  role(withPolicy()),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			o, err := e.Update(context.Background(), tc.args.cr)

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

func TestDelete(t *testing.T) {

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				iam: &fake.MockRoleClient{
					MockDeleteRoleRequest: func(input *awsiam.DeleteRoleInput) awsiam.DeleteRoleRequest {
						return awsiam.DeleteRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.DeleteRoleOutput{}},
						}
					},
				},
				cr: role(withRoleName(&roleName)),
			},
			want: want{
				cr: role(withRoleName(&roleName),
					withConditions(corev1alpha1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockRoleClient{
					MockDeleteRoleRequest: func(input *awsiam.DeleteRoleInput) awsiam.DeleteRoleRequest {
						return awsiam.DeleteRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: role(),
			},
			want: want{
				cr:  role(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRoleClient{
					MockDeleteRoleRequest: func(input *awsiam.DeleteRoleInput) awsiam.DeleteRoleRequest {
						return awsiam.DeleteRoleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: role(),
			},
			want: want{
				cr: role(withConditions(corev1alpha1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
