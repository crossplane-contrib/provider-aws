/*
Copyright 2020 The Crossplane Authors.

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

package certificateauthoritypermission

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1alpha1 "github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
	acmpca "github.com/crossplane/provider-aws/pkg/clients/acmpca"
	"github.com/crossplane/provider-aws/pkg/clients/acmpca/fake"
)

var (
	// an arbitrary managed resource
	unexpecedItem           resource.Managed
	certificateAuthorityArn = "someauthorityarn"
	nextToken               = "someNextToken"

	errBoom = errors.New("boom")
)

type args struct {
	acmpca acmpca.CAPermissionClient
	cr     resource.Managed
}

type certificateAuthorityPermissionModifier func(*v1alpha1.CertificateAuthorityPermission)

func withConditions(c ...corev1alpha1.Condition) certificateAuthorityPermissionModifier {
	return func(r *v1alpha1.CertificateAuthorityPermission) { r.Status.ConditionedStatus.Conditions = c }
}

func certificateAuthorityPermission(m ...certificateAuthorityPermissionModifier) *v1alpha1.CertificateAuthorityPermission {
	cr := &v1alpha1.CertificateAuthorityPermission{}
	meta.SetExternalName(cr, certificateAuthorityArn)
	for _, f := range m {
		f(cr)
	}
	return cr
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockListPermissionsRequest: func(input *awsacmpca.ListPermissionsInput) awsacmpca.ListPermissionsRequest {
						return awsacmpca.ListPermissionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.ListPermissionsOutput{
								NextToken: aws.String(nextToken),
								Permissions: []awsacmpca.Permission{{
									Actions:                 []awsacmpca.ActionType{awsacmpca.ActionTypeIssueCertificate, awsacmpca.ActionTypeGetCertificate, awsacmpca.ActionTypeListPermissions},
									CertificateAuthorityArn: aws.String(certificateAuthorityArn),
								}},
							}},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr: certificateAuthorityPermission(withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockListPermissionsRequest: func(input *awsacmpca.ListPermissionsInput) awsacmpca.ListPermissionsRequest {
						return awsacmpca.ListPermissionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr:  certificateAuthorityPermission(),
				err: errors.Wrap(errBoom, errGet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acmpca,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			}
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.CreatePermissionOutput{}},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr: certificateAuthorityPermission(withConditions(corev1alpha1.Creating())),
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr:  certificateAuthorityPermission(withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acmpca,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			}
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.DeletePermissionOutput{}},
						}
					},
					MockListPermissionsRequest: func(input *awsacmpca.ListPermissionsInput) awsacmpca.ListPermissionsRequest {
						return awsacmpca.ListPermissionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.ListPermissionsOutput{
								NextToken:   aws.String(nextToken),
								Permissions: []awsacmpca.Permission{{}},
							}},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr: certificateAuthorityPermission(withConditions(corev1alpha1.Deleting())),
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
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
					MockListPermissionsRequest: func(input *awsacmpca.ListPermissionsInput) awsacmpca.ListPermissionsRequest {
						return awsacmpca.ListPermissionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.ListPermissionsOutput{
								NextToken:   aws.String(nextToken),
								Permissions: []awsacmpca.Permission{{}},
							}},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr:  certificateAuthorityPermission(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityPermissionClient{
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil)},
						}
					},
					MockListPermissionsRequest: func(input *awsacmpca.ListPermissionsInput) awsacmpca.ListPermissionsRequest {
						return awsacmpca.ListPermissionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsacmpca.ListPermissionsOutput{
								NextToken:   aws.String(nextToken),
								Permissions: []awsacmpca.Permission{{}},
							}},
						}
					},
				},
				cr: certificateAuthorityPermission(),
			},
			want: want{
				cr:  certificateAuthorityPermission(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil), errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.acmpca}
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
