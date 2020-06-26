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

package elb

import (
	"context"
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticloadbalancing/elb"
	"github.com/crossplane/provider-aws/pkg/clients/elasticloadbalancing/elb/fake"
)

var (
	elbName                 = "some-elb"
	protocol                = "HTTP"
	port80            int64 = 80
	availabilityZones       = []string{"us-east-2a"}
	securityGroups          = []string{"sg-someid"}
	subnets                 = []string{"subnet1"}
	listener                = awselb.Listener{
		InstancePort:     &port80,
		InstanceProtocol: &protocol,
		LoadBalancerPort: &port80,
		Protocol:         &protocol,
	}

	errBoom = errors.New("boom")

	loadBalancer = awselb.LoadBalancerDescription{
		AvailabilityZones: availabilityZones,
	}
)

const (
	providerName    = "aws-creds"
	secretNamespace = "crossplane-system"
	testRegion      = "us-east-1"

	connectionSecretName = "my-little-secret"
	secretKey            = "credentials"
	credData             = "confidential!"
)

type args struct {
	kube client.Client
	elb  elb.Client
	cr   resource.Managed
}

type elbModifier func(*v1alpha1.ELB)

func withConditions(c ...corev1alpha1.Condition) elbModifier {
	return func(r *v1alpha1.ELB) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.ELBParameters) elbModifier {
	return func(r *v1alpha1.ELB) { r.Spec.ForProvider = p }
}

func withExternalName(name string) elbModifier {
	return func(r *v1alpha1.ELB) { meta.SetExternalName(r, name) }
}

func elbResource(m ...elbModifier) *v1alpha1.ELB {
	cr := &v1alpha1.ELB{
		Spec: v1alpha1.ELBSpec{
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (elb.Client, error)
		cr          *v1alpha1.ELB
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i elb.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: elbResource(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i elb.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: elbResource(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: elbResource(),
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
				cr: elbResource(),
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
				cr: elbResource(),
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
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withSpec(v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
				}),
					withExternalName(elbName),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleELB": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer, loadBalancer},
							}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr:  elbResource(withExternalName(elbName)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeELBError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr:  elbResource(withExternalName(elbName)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
		"KubeClientError": {
			args: args{
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
				err: errors.Wrap(errBoom, errSpecUpdate),
			},
		},
		"NotUptoDate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroups: securityGroups,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
						SecurityGroups:    securityGroups,
					}),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb, kube: tc.kube}
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
				elb: &fake.MockClient{
					MockCreateLoadBalancerRequest: func(input *awselb.CreateLoadBalancerInput) awselb.CreateLoadBalancerRequest {
						return awselb.CreateLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.CreateLoadBalancerOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					}),
					withConditions(corev1alpha1.Creating())),
			},
		},
		"CreateError": {
			args: args{
				elb: &fake.MockClient{
					MockCreateLoadBalancerRequest: func(input *awselb.CreateLoadBalancerInput) awselb.CreateLoadBalancerRequest {
						return awselb.CreateLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: availabilityZones,
					}),
					withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
		"UpdateAZ": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
					MockEnableAvailabilityZonesForLoadBalancerRequest: func(input *awselb.EnableAvailabilityZonesForLoadBalancerInput) awselb.EnableAvailabilityZonesForLoadBalancerRequest {
						return awselb.EnableAvailabilityZonesForLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.EnableAvailabilityZonesForLoadBalancerOutput{}},
						}
					},
					MockDisableAvailabilityZonesForLoadBalancerRequest: func(input *awselb.DisableAvailabilityZonesForLoadBalancerInput) awselb.DisableAvailabilityZonesForLoadBalancerRequest {
						return awselb.DisableAvailabilityZonesForLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DisableAvailabilityZonesForLoadBalancerOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: []string{"us-east-2b"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						AvailabilityZones: []string{"us-east-2b"},
					})),
			},
		},
		"UpdateSubnet": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{
									{
										Subnets: subnets,
									},
								},
							}},
						}
					},
					MockAttachLoadBalancerToSubnetsRequest: func(input *awselb.AttachLoadBalancerToSubnetsInput) awselb.AttachLoadBalancerToSubnetsRequest {
						return awselb.AttachLoadBalancerToSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.AttachLoadBalancerToSubnetsOutput{}},
						}
					},
					MockDetachLoadBalancerFromSubnetsRequest: func(input *awselb.DetachLoadBalancerFromSubnetsInput) awselb.DetachLoadBalancerFromSubnetsRequest {
						return awselb.DetachLoadBalancerFromSubnetsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DetachLoadBalancerFromSubnetsOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Subnets: []string{"subnet2"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Subnets: []string{"subnet2"},
					})),
			},
		},
		"UpdateSG": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{
									{
										SecurityGroups: securityGroups,
									},
								},
							}},
						}
					},
					MockApplySecurityGroupsToLoadBalancerRequest: func(input *awselb.ApplySecurityGroupsToLoadBalancerInput) awselb.ApplySecurityGroupsToLoadBalancerRequest {
						return awselb.ApplySecurityGroupsToLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.ApplySecurityGroupsToLoadBalancerOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroups: []string{"sg-other"},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						SecurityGroups: []string{"sg-other"},
					})),
			},
		},
		"UpdateListener": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{
									{
										ListenerDescriptions: []awselb.ListenerDescription{
											{
												Listener: &listener,
											},
										},
									},
								},
							}},
						}
					},
					MockCreateLoadBalancerListenersRequest: func(input *awselb.CreateLoadBalancerListenersInput) awselb.CreateLoadBalancerListenersRequest {
						return awselb.CreateLoadBalancerListenersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.CreateLoadBalancerListenersOutput{}},
						}
					},
					MockDeleteLoadBalancerListenersRequest: func(input *awselb.DeleteLoadBalancerListenersInput) awselb.DeleteLoadBalancerListenersRequest {
						return awselb.DeleteLoadBalancerListenersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DeleteLoadBalancerListenersOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Listeners: []v1alpha1.Listener{
							{
								InstancePort:     8180,
								InstanceProtocol: &protocol,
								LoadBalancerPort: 8180,
								Protocol:         protocol,
							},
						},
					})),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBParameters{
						Listeners: []v1alpha1.Listener{
							{
								InstancePort:     8180,
								InstanceProtocol: &protocol,
								LoadBalancerPort: 8180,
								Protocol:         protocol,
							},
						},
					})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
		"Successful": {
			args: args{
				elb: &fake.MockClient{
					MockDeleteLoadBalancerRequest: func(input *awselb.DeleteLoadBalancerInput) awselb.DeleteLoadBalancerRequest {
						return awselb.DeleteLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DeleteLoadBalancerOutput{}},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withConditions(corev1alpha1.Deleting())),
			},
		},
		"DeleteError": {
			args: args{
				elb: &fake.MockClient{
					MockDeleteLoadBalancerRequest: func(input *awselb.DeleteLoadBalancerInput) awselb.DeleteLoadBalancerRequest {
						return awselb.DeleteLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbResource(withExternalName(elbName),
					withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
