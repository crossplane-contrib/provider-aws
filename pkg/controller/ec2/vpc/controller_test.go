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

package vpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
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
	vpcID          = "some Id"
	cidr           = "192.168.0.0/32"
	tenancyDefault = "default"

	errBoom = errors.New("boom")
)

type args struct {
	vpc  ec2.VPCClient
	kube client.Client
	cr   *v1beta1.VPC
}

type vpcModifier func(*v1beta1.VPC)

func withTags(tagMaps ...map[string]string) vpcModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.VPC) { r.Spec.ForProvider.Tags = tagList }
}

func withExternalName(name string) vpcModifier {
	return func(r *v1beta1.VPC) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) vpcModifier {
	return func(r *v1beta1.VPC) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.VPCParameters) vpcModifier {
	return func(r *v1beta1.VPC) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.VPCObservation) vpcModifier {
	return func(r *v1beta1.VPC) { r.Status.AtProvider = s }
}

func vpc(m ...vpcModifier) *v1beta1.VPC {
	cr := &v1beta1.VPC{
		Spec: v1beta1.VPCSpec{
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

func TestConnect(t *testing.T) {
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ec2.VPCClient, error)
		cr          *v1beta1.VPC
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.VPCClient, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: vpc(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.VPCClient, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: vpc(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: vpc(),
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
				cr: vpc(),
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
				cr: vpc(),
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
		cr     *v1beta1.VPC
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{
									InstanceTenancy: awsec2.TenancyDefault,
									State:           awsec2.VpcStateAvailable,
								}},
							}},
						}
					},
					MockDescribeVpcAttributeRequest: func(input *awsec2.DescribeVpcAttributeInput) awsec2.DescribeVpcAttributeRequest {
						return awsec2.DescribeVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcAttributeOutput{
								EnableDnsHostnames: &awsec2.AttributeBooleanValue{},
								EnableDnsSupport:   &awsec2.AttributeBooleanValue{},
							}},
						}
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withStatus(v1beta1.VPCObservation{
					VPCState: "available",
				}), withExternalName(vpcID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleVpcs": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeVpcsOutput{
								Vpcs: []awsec2.Vpc{{}, {}},
							}},
						}
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				vpc: &fake.MockVPCClient{
					MockDescribe: func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
						return awsec2.DescribeVpcsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
					CIDRBlock:       cidr,
				}), withExternalName(vpcID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr     *v1beta1.VPC
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				vpc: &fake.MockVPCClient{
					MockCreate: func(input *awsec2.CreateVpcInput) awsec2.CreateVpcRequest {
						return awsec2.CreateVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateVpcOutput{
								Vpc: &awsec2.Vpc{
									VpcId:     aws.String(vpcID),
									CidrBlock: aws.String(cidr),
								},
							}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr: vpc(withExternalName(vpcID),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				vpc: &fake.MockVPCClient{
					MockCreate: func(input *awsec2.CreateVpcInput) awsec2.CreateVpcRequest {
						return awsec2.CreateVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr     *v1beta1.VPC
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockModifyTenancy: func(input *awsec2.ModifyVpcTenancyInput) awsec2.ModifyVpcTenancyRequest {
						return awsec2.ModifyVpcTenancyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcTenancyOutput{}},
						}
					},
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockModifyTenancy: func(input *awsec2.ModifyVpcTenancyInput) awsec2.ModifyVpcTenancyRequest {
						return awsec2.ModifyVpcTenancyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
					MockModifyAttribute: func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
						return awsec2.ModifyVpcAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifyVpcAttributeOutput{}},
						}
					},
				},
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
			},
			want: want{
				cr: vpc(withSpec(v1beta1.VPCParameters{
					InstanceTenancy: aws.String(tenancyDefault),
				})),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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
		cr  *v1beta1.VPC
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDelete: func(input *awsec2.DeleteVpcInput) awsec2.DeleteVpcRequest {
						return awsec2.DeleteVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteVpcOutput{}},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr: vpc(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				vpc: &fake.MockVPCClient{
					MockDelete: func(input *awsec2.DeleteVpcInput) awsec2.DeleteVpcRequest {
						return awsec2.DeleteVpcRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: vpc(),
			},
			want: want{
				cr:  vpc(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.vpc}
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

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1beta1.VPC
		kube client.Client
	}
	type want struct {
		cr  *v1beta1.VPC
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   vpc(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: vpc(withTags(resource.GetExternalTags(vpc()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   vpc(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tagger{kube: tc.kube}
			err := e.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.SortSlices(func(a, b v1beta1.Tag) bool { return a.Key > b.Key })); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
