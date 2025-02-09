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

package accesskey

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	userName       = "some arbitrary name"
	activeStatus   = awsiamtypes.StatusTypeActive
	inactiveStatus = awsiamtypes.StatusTypeInactive
	accessKeyID    = "accessKeyID"
	secretKeyID    = "secretKeyID"

	errBoom = errors.New("boom")
)

type args struct {
	iam  iam.AccessClient
	cr   resource.Managed
	kube client.Client
}

type accessModifier func(*v1beta1.AccessKey)

func withConditions(c ...xpv1.Condition) accessModifier {
	return func(r *v1beta1.AccessKey) { r.Status.ConditionedStatus.Conditions = c }
}

func withUsername(username string) accessModifier {
	return func(r *v1beta1.AccessKey) {
		r.Spec.ForProvider.Username = username
	}
}

func withStatus(status string) accessModifier {
	return func(r *v1beta1.AccessKey) {
		r.Spec.ForProvider.Status = status
	}
}

func withAccessKey(keyid string) accessModifier {
	return func(r *v1beta1.AccessKey) {
		meta.SetExternalName(r, keyid)
	}
}

func accesskey(m ...accessModifier) *v1beta1.AccessKey {
	cr := &v1beta1.AccessKey{}
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
					MockListAccessKeys: func(ctx context.Context, input *awsiam.ListAccessKeysInput, opts []func(*awsiam.Options)) (*awsiam.ListAccessKeysOutput, error) {
						return &awsiam.ListAccessKeysOutput{
							AccessKeyMetadata: []awsiamtypes.AccessKeyMetadata{{
								AccessKeyId: aws.String(accessKeyID),
								Status:      activeStatus,
								UserName:    aws.String(userName),
							}},
						}, nil
					},
				},
				cr: accesskey(withUsername(userName), withAccessKey(accessKeyID), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"ValidInputNeedsUpdate": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeys: func(ctx context.Context, input *awsiam.ListAccessKeysInput, opts []func(*awsiam.Options)) (*awsiam.ListAccessKeysOutput, error) {
						return &awsiam.ListAccessKeysOutput{
							AccessKeyMetadata: []awsiamtypes.AccessKeyMetadata{{
								AccessKeyId: aws.String(accessKeyID),
								Status:      inactiveStatus,
								UserName:    aws.String(userName),
							}},
						}, nil
					},
				},
				cr: accesskey(withUsername(userName), withAccessKey(accessKeyID), withStatus(string(activeStatus))),
			},
			want: want{
				cr: accesskey(withUsername(userName),
					withAccessKey(accessKeyID),
					withStatus(string(activeStatus)),
					withConditions(xpv1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"ValidInputNotExists": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeys: func(ctx context.Context, input *awsiam.ListAccessKeysInput, opts []func(*awsiam.Options)) (*awsiam.ListAccessKeysOutput, error) {
						return &awsiam.ListAccessKeysOutput{
							AccessKeyMetadata: []awsiamtypes.AccessKeyMetadata{},
						}, nil
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
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ListError": {
			args: args{
				iam: &fake.MockAccessClient{
					MockListAccessKeys: func(ctx context.Context, input *awsiam.ListAccessKeysInput, opts []func(*awsiam.Options)) (*awsiam.ListAccessKeysOutput, error) {
						return nil, errBoom
					},
				},
				cr: accesskey(withAccessKey("test")),
			},
			want: want{
				cr:  accesskey(withAccessKey("test")),
				err: errorutils.Wrap(errBoom, errList),
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
					MockCreateAccessKey: func(ctx context.Context, input *awsiam.CreateAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.CreateAccessKeyOutput, error) {
						return &awsiam.CreateAccessKeyOutput{
							AccessKey: &awsiamtypes.AccessKey{
								AccessKeyId:     aws.String(accessKeyID),
								SecretAccessKey: aws.String(secretKeyID),
								Status:          activeStatus,
								UserName:        aws.String(userName),
							},
						}, nil
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
					withAccessKey(accessKeyID)),
				result: managed.ExternalCreation{

					ConnectionDetails: map[string][]byte{
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(secretKeyID),
						xpv1.ResourceCredentialsSecretUserKey:     []byte(accessKeyID),
					}},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockAccessClient{
					MockCreateAccessKey: func(ctx context.Context, input *awsiam.CreateAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.CreateAccessKeyOutput, error) {
						return nil, errBoom
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr:  accesskey(),
				err: errorutils.Wrap(errBoom, errCreate),
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
					MockDeleteAccessKey: func(ctx context.Context, input *awsiam.DeleteAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteAccessKeyOutput, error) {
						return &awsiam.DeleteAccessKeyOutput{}, nil
					},
				},
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName)),
			},
			want: want{
				cr: accesskey(withAccessKey(accessKeyID), withUsername(userName),
					withConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockAccessClient{
					MockDeleteAccessKey: func(ctx context.Context, input *awsiam.DeleteAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteAccessKeyOutput, error) {
						return nil, errBoom
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr:  accesskey(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockAccessClient{
					MockDeleteAccessKey: func(ctx context.Context, input *awsiam.DeleteAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteAccessKeyOutput, error) {
						return nil, &awsiamtypes.NoSuchEntityException{}
					},
				},
				cr: accesskey(),
			},
			want: want{
				cr: accesskey(withConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			_, err := e.Delete(context.Background(), tc.args.cr)

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
					MockUpdateAccessKey: func(ctx context.Context, input *awsiam.UpdateAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.UpdateAccessKeyOutput, error) {
						return &awsiam.UpdateAccessKeyOutput{}, nil
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
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockAccessClient{
					MockUpdateAccessKey: func(ctx context.Context, input *awsiam.UpdateAccessKeyInput, opts []func(*awsiam.Options)) (*awsiam.UpdateAccessKeyOutput, error) {
						return nil, errBoom
					},
				},
				cr: accesskey(withStatus(string(activeStatus))),
			},
			want: want{
				cr:  accesskey(withStatus(string(activeStatus))),
				err: errorutils.Wrap(errBoom, errUpdate),
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
