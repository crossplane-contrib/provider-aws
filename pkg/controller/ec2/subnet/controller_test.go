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

package subnet

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

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
	subnetID = "some Id"

	errBoom = errors.New("boom")
)

type args struct {
	subnet ec2.SubnetClient
	kube   client.Client
	cr     *v1beta1.Subnet
}

type subnetModifier func(*v1beta1.Subnet)

func withExternalName(name string) subnetModifier {
	return func(r *v1beta1.Subnet) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.SubnetParameters) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.SubnetObservation) subnetModifier {
	return func(r *v1beta1.Subnet) { r.Status.AtProvider = s }
}

func subnet(m ...subnetModifier) *v1beta1.Subnet {
	cr := &v1beta1.Subnet{
		Spec: v1beta1.SubnetSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: runtimev1alpha1.Reference{Name: providerName},
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ec2.SubnetClient, error)
		cr          *v1beta1.Subnet
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.SubnetClient, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: subnet(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.SubnetClient, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: subnet(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: subnet(),
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
				cr: subnet(),
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
				cr: subnet(),
			},
			want: want{
				err: errors.New(errGetProviderSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{client: tc.kube, newClientFn: tc.newClientFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Subnet
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{
									{
										State: awsec2.SubnetStateAvailable,
									},
								},
							}},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetState: "available",
				}), withExternalName(subnetID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleSubnets": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{}, {}},
							}},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errors.New(errMultipleItems),
			},
		},
		"NotUpToDate": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{
									{
										State:               awsec2.SubnetStateAvailable,
										MapPublicIpOnLaunch: aws.Bool(false),
									},
								},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}),
					withExternalName(subnetID)),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetState: string(awsec2.SubnetStateAvailable),
				}), withExternalName(subnetID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"DescribeFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(withExternalName(subnetID)),
			},
			want: want{
				cr:  subnet(withExternalName(subnetID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr     *v1beta1.Subnet
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
				subnet: &fake.MockSubnetClient{
					MockCreate: func(input *awsec2.CreateSubnetInput) awsec2.CreateSubnetRequest {
						return awsec2.CreateSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateSubnetOutput{
								Subnet: &awsec2.Subnet{
									SubnetId: aws.String(subnetID),
								},
							}},
						}
					},
				},
				cr: subnet(),
			},
			want: want{
				cr: subnet(withExternalName(subnetID),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"CreateFailed": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				subnet: &fake.MockSubnetClient{
					MockCreate: func(input *awsec2.CreateSubnetInput) awsec2.CreateSubnetRequest {
						return awsec2.CreateSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(),
			},
			want: want{
				cr:  subnet(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr     *v1beta1.Subnet
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(input *awsec2.ModifySubnetAttributeInput) awsec2.ModifySubnetAttributeRequest {
						return awsec2.ModifySubnetAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifySubnetAttributeOutput{}},
						}
					},
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{
									SubnetId:            aws.String(subnetID),
									MapPublicIpOnLaunch: aws.Bool(false),
								}},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockModify: func(input *awsec2.ModifySubnetAttributeInput) awsec2.ModifySubnetAttributeRequest {
						return awsec2.ModifySubnetAttributeRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ModifySubnetAttributeOutput{}},
						}
					},
					MockDescribe: func(input *awsec2.DescribeSubnetsInput) awsec2.DescribeSubnetsRequest {
						return awsec2.DescribeSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeSubnetsOutput{
								Subnets: []awsec2.Subnet{{
									SubnetId:            aws.String(subnetID),
									MapPublicIpOnLaunch: aws.Bool(false),
								}},
							}},
						}
					},
				},
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withSpec(v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				}), withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
		cr  *v1beta1.Subnet
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(input *awsec2.DeleteSubnetInput) awsec2.DeleteSubnetRequest {
						return awsec2.DeleteSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteSubnetOutput{}},
						}
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				subnet: &fake.MockSubnetClient{
					MockDelete: func(input *awsec2.DeleteSubnetInput) awsec2.DeleteSubnetRequest {
						return awsec2.DeleteSubnetRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				})),
			},
			want: want{
				cr: subnet(withStatus(v1beta1.SubnetObservation{
					SubnetID: subnetID,
				}), withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.subnet}
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
