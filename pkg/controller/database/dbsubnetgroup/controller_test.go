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

package dbsubnetgroup

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1beta1 "github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	dbsg "github.com/crossplane/provider-aws/pkg/clients/dbsubnetgroup"
	"github.com/crossplane/provider-aws/pkg/clients/dbsubnetgroup/fake"
)

const (
	providerName    = "aws-creds"
	secretNamespace = "crossplane-system"
	testRegion      = "us-east-1"

	connectionSecretName = "my-little-secret"
	secretKey            = "credentials"
	credData             = "confidential!"
)

var (
	dbSubnetGroupDescription = "arbitrary description"
	errBoom                  = errors.New("boom")
)

type args struct {
	client dbsg.Client
	kube   client.Client
	cr     *v1beta1.DBSubnetGroup
}

type dbSubnetGroupModifier func(*v1beta1.DBSubnetGroup)

func withConditions(c ...runtimev1alpha1.Condition) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Status.ConditionedStatus.Conditions = c }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Status.SetBindingPhase(p) }
}

func withDBSubnetGroupStatus(s string) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Status.AtProvider.State = s }
}

func withDBSubnetGroupDescription(s string) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Spec.ForProvider.Description = s }
}

func withDBSubnetGroupTags() dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) {
		sg.Spec.ForProvider.Tags = []v1beta1.Tag{{Key: "arbitrary key", Value: "arbitrary value"}}
	}
}

func mockListTagsForResourceRequest(input *awsrds.ListTagsForResourceInput) awsrds.ListTagsForResourceRequest {
	return awsrds.ListTagsForResourceRequest{
		Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.ListTagsForResourceOutput{TagList: []awsrds.Tag{}}},
	}
}

func dbSubnetGroup(m ...dbSubnetGroupModifier) *v1beta1.DBSubnetGroup {
	cr := &v1beta1.DBSubnetGroup{
		Spec: v1beta1.DBSubnetGroupSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
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

func Test_Connect(t *testing.T) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectionSecretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			secretKey: []byte(credData),
		},
	}

	providerSA := func(saVal bool) awsv1alpha3.Provider {
		return awsv1alpha3.Provider{
			Spec: awsv1alpha3.ProviderSpec{
				Region:            testRegion,
				UseServiceAccount: &saVal,
				ProviderSpec: runtimev1alpha1.ProviderSpec{
					CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
						SecretReference: runtimev1alpha1.SecretReference{
							Namespace: secretNamespace,
							Name:      connectionSecretName,
						},
						Key: secretKey,
					},
				},
			},
		}
	}

	type args struct {
		kube        client.Client
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (dbsg.Client, error)
		cr          *v1beta1.DBSubnetGroup
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i dbsg.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: dbSubnetGroup(),
			},
		},
		"SuccessfulUseServiceAccount": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						if key == (client.ObjectKey{Name: providerName}) {
							p := providerSA(true)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i dbsg.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: dbSubnetGroup(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProvider),
			},
		},
		"SecretGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProviderSecret),
			},
		},
		"SecretGetFailedNil": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.SetCredentialsSecretReference(nil)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				err: errors.New(errGetProviderSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newClientFn: tc.newClientFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{
									{
										SubnetGroupStatus: aws.String(string(v1beta1.DBSubnetGroupStateAvailable)),
									},
								},
							}},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound),
					withDBSubnetGroupStatus(string(v1beta1.DBSubnetGroupStateAvailable)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(runtimev1alpha1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedState": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(runtimev1alpha1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
		"NotFound": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsrds.ErrCodeDBSubnetGroupNotFoundFault)},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{
									{
										DBSubnetGroupDescription: aws.String(dbSubnetGroupDescription),
									},
								},
							}},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
					withConditions(runtimev1alpha1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{
									{
										DBSubnetGroupDescription: aws.String(dbSubnetGroupDescription),
									},
								},
							}},
						}
					},
					MockListTagsForResourceRequest: mockListTagsForResourceRequest,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
				),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockCreateDBSubnetGroupRequest: func(input *awsrds.CreateDBSubnetGroupInput) awsrds.CreateDBSubnetGroupRequest {
						return awsrds.CreateDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.CreateDBSubnetGroupOutput{}},
						}
					},
				},
				cr: dbSubnetGroup(withDBSubnetGroupDescription(dbSubnetGroupDescription)),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockCreateDBSubnetGroupRequest: func(input *awsrds.CreateDBSubnetGroupInput) awsrds.CreateDBSubnetGroupRequest {
						return awsrds.CreateDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroupRequest: func(input *awsrds.ModifyDBSubnetGroupInput) awsrds.ModifyDBSubnetGroupRequest {
						return awsrds.ModifyDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.ModifyDBSubnetGroupOutput{}},
						}
					},
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(),
			},
		},
		"FailedModify": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroupRequest: func(input *awsrds.ModifyDBSubnetGroupInput) awsrds.ModifyDBSubnetGroupRequest {
						return awsrds.ModifyDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
		"SuccessfulWithTags": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroupRequest: func(input *awsrds.ModifyDBSubnetGroupInput) awsrds.ModifyDBSubnetGroupRequest {
						return awsrds.ModifyDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.ModifyDBSubnetGroupOutput{}},
						}
					},
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
					MockAddTagsToResourceRequest: func(input *awsrds.AddTagsToResourceInput) awsrds.AddTagsToResourceRequest {
						return awsrds.AddTagsToResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.AddTagsToResourceOutput{}},
						}
					},
				},
				cr: dbSubnetGroup(withDBSubnetGroupTags()),
			},
			want: want{
				cr: dbSubnetGroup(withDBSubnetGroupTags()),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr  *v1beta1.DBSubnetGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroupRequest: func(input *awsrds.DeleteDBSubnetGroupInput) awsrds.DeleteDBSubnetGroupRequest {
						return awsrds.DeleteDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DeleteDBSubnetGroupOutput{}},
						}
					},
					MockModifyDBSubnetGroupRequest: func(input *awsrds.ModifyDBSubnetGroupInput) awsrds.ModifyDBSubnetGroupRequest {
						return awsrds.ModifyDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.ModifyDBSubnetGroupOutput{}},
						}
					},
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroupRequest: func(input *awsrds.DeleteDBSubnetGroupInput) awsrds.DeleteDBSubnetGroupRequest {
						return awsrds.DeleteDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsrds.ErrCodeDBSubnetGroupNotFoundFault)},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroupRequest: func(input *awsrds.DeleteDBSubnetGroupInput) awsrds.DeleteDBSubnetGroupRequest {
						return awsrds.DeleteDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockModifyDBSubnetGroupRequest: func(input *awsrds.ModifyDBSubnetGroupInput) awsrds.ModifyDBSubnetGroupRequest {
						return awsrds.ModifyDBSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.ModifyDBSubnetGroupOutput{}},
						}
					},
					MockDescribeDBSubnetGroupsRequest: func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
						return awsrds.DescribeDBSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsrds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []awsrds.DBSubnetGroup{{}},
							}},
						}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
