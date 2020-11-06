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

package iamaccesskey

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	// an arbitrary managed resource
	unexpecedItem  resource.Managed
	userName       = "some arbitrary name"
	activeStatus   = awsiam.StatusTypeActive
	inactiveStatus = awsiam.StatusTypeInactive
	accessKeyID    = "accessKeyID"
	secretKeyID    = "secretKeyID"

	errBoom = errors.New("boom")
)

type args struct {
	iam  iam.AccessClient
	cr   resource.Managed
	kube client.Client
}

type accessModifier func(*v1alpha1.IAMAccessKey)

func withConditions(c ...corev1alpha1.Condition) accessModifier {
	return func(r *v1alpha1.IAMAccessKey) { r.Status.ConditionedStatus.Conditions = c }
}

func withUsername(username string) accessModifier {
	return func(r *v1alpha1.IAMAccessKey) {
		r.Spec.ForProvider.IAMUsername = username
	}
}

func withStatus(status string) accessModifier {
	return func(r *v1alpha1.IAMAccessKey) {
		r.Spec.ForProvider.Status = status
	}
}

func withAccessKey(keyid string) accessModifier {
	return func(r *v1alpha1.IAMAccessKey) {
		meta.SetExternalName(r, keyid)
	}
}

func accesskey(m ...accessModifier) *v1alpha1.IAMAccessKey {
	cr := &v1alpha1.IAMAccessKey{}
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
		"ValidInputExists": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeysRequest: func(input *awsiam.ListAccessKeysInput) awsiam.ListAccessKeysRequest {
						return awsiam.ListAccessKeysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAccessKeysOutput{
								AccessKeyMetadata: []awsiam.AccessKeyMetadata{{
									AccessKeyId: aws.String(accessKeyID),
									Status:      activeStatus,
									UserName:    aws.String(userName),
								}},
							}},
						}
					},
				},
				cr: accesskey(withUsername(userName), withAccessKey(accessKeyID), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"ValidInputNeedsUpdate": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeysRequest: func(input *awsiam.ListAccessKeysInput) awsiam.ListAccessKeysRequest {
						return awsiam.ListAccessKeysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAccessKeysOutput{
								AccessKeyMetadata: []awsiam.AccessKeyMetadata{{
									AccessKeyId: aws.String(accessKeyID),
									Status:      inactiveStatus,
									UserName:    aws.String(userName),
								}},
							}},
						}
					},
				},
				cr: accesskey(withUsername(userName), withAccessKey(accessKeyID), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"ValidInputNotExists": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeysRequest: func(input *awsiam.ListAccessKeysInput) awsiam.ListAccessKeysRequest {
						return awsiam.ListAccessKeysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAccessKeysOutput{
								AccessKeyMetadata: []awsiam.AccessKeyMetadata{},
							}},
						}
					},
				},
				cr: accesskey(withUsername(userName), withAccessKey(accessKeyID), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus))),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
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
		"ListError": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeysRequest: func(input *awsiam.ListAccessKeysInput) awsiam.ListAccessKeysRequest {
						return awsiam.ListAccessKeysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom, Retryer: aws.NoOpRetryer{}},
						}
					},
				},
				cr: accesskey(withAccessKey("test")),
			},
			want: want{
				cr:  accesskey(withAccessKey("test")),
				err: errors.Wrap(errBoom, errList),
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
		"ValidInput": {
			args: args{
				iam: &fake.MockAccessClient{
					MockCreateAccessKeyRequest: func(input *awsiam.CreateAccessKeyInput) awsiam.CreateAccessKeyRequest {
						return awsiam.CreateAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.CreateAccessKeyOutput{
								AccessKey: &awsiam.AccessKey{
									AccessKeyId:     aws.String(accessKeyID),
									SecretAccessKey: aws.String(secretKeyID),
									Status:          activeStatus,
									UserName:        aws.String(userName),
								},
							}},
						}
					},
				},
				cr: accesskey(withUsername(userName)),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr: accesskey(
					withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(corev1alpha1.Creating())),
				result: managed.ExternalCreation{ConnectionDetails: map[string][]byte{
					corev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(secretKeyID),
					corev1alpha1.ResourceCredentialsSecretUserKey:     []byte(accessKeyID),
				}},
			},
		},
		"CreateFailedKubeUpdate": {
			args: args{
				iam: &fake.MockAccessClient{
					MockCreateAccessKeyRequest: func(input *awsiam.CreateAccessKeyInput) awsiam.CreateAccessKeyRequest {
						return awsiam.CreateAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.CreateAccessKeyOutput{
								AccessKey: &awsiam.AccessKey{
									AccessKeyId:     aws.String(accessKeyID),
									SecretAccessKey: aws.String(secretKeyID),
									Status:          activeStatus,
									UserName:        aws.String(userName),
								},
							}},
						}
					},
				},
				cr: accesskey(withUsername(userName)),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
			},
			want: want{
				cr: accesskey(
					withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
				iam: &fake.MockAccessClient{
					MockCreateAccessKeyRequest: func(input *awsiam.CreateAccessKeyInput) awsiam.CreateAccessKeyRequest {
						return awsiam.CreateAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr:  accesskey(withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam, kube: tc.kube}
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
		"ValidInput": {
			args: args{
				iam: &fake.MockAccessClient{
					MockDeleteAccessKeyRequest: func(input *awsiam.DeleteAccessKeyInput) awsiam.DeleteAccessKeyRequest {
						return awsiam.DeleteAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.DeleteAccessKeyOutput{}},
						}
					},
				},
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName)),
			},
			want: want{
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName),
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
				iam: &fake.MockAccessClient{
					MockDeleteAccessKeyRequest: func(input *awsiam.DeleteAccessKeyInput) awsiam.DeleteAccessKeyRequest {
						return awsiam.DeleteAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr:  accesskey(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockAccessClient{
					MockDeleteAccessKeyRequest: func(input *awsiam.DeleteAccessKeyInput) awsiam.DeleteAccessKeyRequest {
						return awsiam.DeleteAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr: accesskey(withConditions(corev1alpha1.Deleting())),
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

func TestUpdate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		err    error
		update managed.ExternalUpdate
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				iam: &fake.MockAccessClient{
					MockUpdateAccessKeyRequest: func(input *awsiam.UpdateAccessKeyInput) awsiam.UpdateAccessKeyRequest {
						return awsiam.UpdateAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.UpdateAccessKeyOutput{}},
						}
					},
				},
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName), withStatus(string(activeStatus))),
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
				iam: &fake.MockAccessClient{
					MockUpdateAccessKeyRequest: func(input *awsiam.UpdateAccessKeyInput) awsiam.UpdateAccessKeyRequest {
						return awsiam.UpdateAccessKeyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: accesskey(withStatus(string(activeStatus))),
			},
			want: want{
				cr:  accesskey(withStatus(string(activeStatus))),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			update, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.update, update, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
