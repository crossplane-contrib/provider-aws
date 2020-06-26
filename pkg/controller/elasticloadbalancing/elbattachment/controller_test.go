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

package elbattachment

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
	elbName    = "some-elb"
	instanceID = "someID"

	errBoom = errors.New("boom")

	loadBalancer = awselb.LoadBalancerDescription{
		Instances: []awselb.Instance{
			{
				InstanceId: &instanceID,
			},
		},
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
	elb elb.Client
	cr  resource.Managed
}

type elbAttachmentModifier func(*v1alpha1.ELBAttachment)

func withConditions(c ...corev1alpha1.Condition) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.ELBAttachmentParameters) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { r.Spec.ForProvider = p }
}

func withExternalName(name string) elbAttachmentModifier {
	return func(r *v1alpha1.ELBAttachment) { meta.SetExternalName(r, name) }
}

func elbAttachmentResource(m ...elbAttachmentModifier) *v1alpha1.ELBAttachment {
	cr := &v1alpha1.ELBAttachment{
		Spec: v1alpha1.ELBAttachmentSpec{
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
		cr          *v1alpha1.ELBAttachment
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
				cr: elbAttachmentResource(),
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
				cr: elbAttachmentResource(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: elbAttachmentResource(),
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
				cr: elbAttachmentResource(),
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
				cr: elbAttachmentResource(),
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
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				}),
					withExternalName(elbName),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NoAttachment": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DescribeLoadBalancersOutput{
								LoadBalancerDescriptions: []awselb.LoadBalancerDescription{loadBalancer},
							}},
						}
					},
				},
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName: elbName,
				})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName: elbName,
				})),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
		"DescribeError": {
			args: args{
				elb: &fake.MockClient{
					MockDescribeLoadBalancersRequest: func(input *awselb.DescribeLoadBalancersInput) awselb.DescribeLoadBalancersRequest {
						return awselb.DescribeLoadBalancersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				})),
			},
			want: want{
				cr: elbAttachmentResource(withSpec(v1alpha1.ELBAttachmentParameters{
					ELBName:    elbName,
					InstanceID: instanceID,
				})),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.elb}
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
					MockRegisterInstancesWithLoadBalancerRequest: func(input *awselb.RegisterInstancesWithLoadBalancerInput) awselb.RegisterInstancesWithLoadBalancerRequest {
						return awselb.RegisterInstancesWithLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.RegisterInstancesWithLoadBalancerOutput{}},
						}
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					}),
					withConditions(corev1alpha1.Creating())),
			},
		},
		"CreateError": {
			args: args{
				elb: &fake.MockClient{
					MockRegisterInstancesWithLoadBalancerRequest: func(input *awselb.RegisterInstancesWithLoadBalancerInput) awselb.RegisterInstancesWithLoadBalancerRequest {
						return awselb.RegisterInstancesWithLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
					})),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withSpec(v1alpha1.ELBAttachmentParameters{
						ELBName:    elbName,
						InstanceID: instanceID,
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
					MockDeregisterInstancesFromLoadBalancerRequest: func(input *awselb.DeregisterInstancesFromLoadBalancerInput) awselb.DeregisterInstancesFromLoadBalancerRequest {
						return awselb.DeregisterInstancesFromLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awselb.DeregisterInstancesFromLoadBalancerOutput{}},
						}
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
					withConditions(corev1alpha1.Deleting())),
			},
		},
		"DeleteError": {
			args: args{
				elb: &fake.MockClient{
					MockDeregisterInstancesFromLoadBalancerRequest: func(input *awselb.DeregisterInstancesFromLoadBalancerInput) awselb.DeregisterInstancesFromLoadBalancerRequest {
						return awselb.DeregisterInstancesFromLoadBalancerRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: elbAttachmentResource(withExternalName(elbName)),
			},
			want: want{
				cr: elbAttachmentResource(withExternalName(elbName),
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
