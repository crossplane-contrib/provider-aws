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

package elasticip

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

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha4"
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
	allocationID   = "some Id"
	domainVpc      = "vpc"
	domainStandard = "standard"
	publicIP       = "1.1.1.1"
	errBoom        = errors.New("boom")
)

type args struct {
	elasticIP ec2.ElasticIPClient
	kube      client.Client
	cr        *v1alpha4.ElasticIP
}

type elasticIPModifier func(*v1alpha4.ElasticIP)

func withTags(tagMaps ...map[string]string) elasticIPModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1alpha4.ElasticIP) { r.Spec.ForProvider.Tags = tagList }
}

func withExternalName(name string) elasticIPModifier {
	return func(r *v1alpha4.ElasticIP) { meta.SetExternalName(r, name) }
}

func withConditions(c ...runtimev1alpha1.Condition) elasticIPModifier {
	return func(r *v1alpha4.ElasticIP) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha4.ElasticIPParameters) elasticIPModifier {
	return func(r *v1alpha4.ElasticIP) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha4.ElasticIPObservation) elasticIPModifier {
	return func(r *v1alpha4.ElasticIP) { r.Status.AtProvider = s }
}

func elasticIP(m ...elasticIPModifier) *v1alpha4.ElasticIP {
	cr := &v1alpha4.ElasticIP{
		Spec: v1alpha4.ElasticIPSpec{
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ec2.ElasticIPClient, error)
		cr          *v1alpha4.ElasticIP
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.ElasticIPClient, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: elasticIP(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i ec2.ElasticIPClient, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: elasticIP(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: elasticIP(),
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
				cr: elasticIP(),
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
				cr: elasticIP(),
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
		cr     *v1alpha4.ElasticIP
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
				elasticIP: &fake.MockElasticIPClient{
					MockDescribe: func(input *awsec2.DescribeAddressesInput) awsec2.DescribeAddressesRequest {
						return awsec2.DescribeAddressesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeAddressesOutput{
								Addresses: []awsec2.Address{{
									AllocationId: &allocationID,
								}},
							}},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withStatus(v1alpha4.ElasticIPObservation{
					AllocationID: allocationID,
				}), withExternalName(allocationID),
					withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleAddresses": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elasticIP: &fake.MockElasticIPClient{
					MockDescribe: func(input *awsec2.DescribeAddressesInput) awsec2.DescribeAddressesRequest {
						return awsec2.DescribeAddressesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeAddressesOutput{
								Addresses: []awsec2.Address{{}, {}},
							}},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withExternalName(allocationID)),
				err: errors.New(errMultipleItems),
			},
		},
		"DescribeFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				elasticIP: &fake.MockElasticIPClient{
					MockDescribe: func(input *awsec2.DescribeAddressesInput) awsec2.DescribeAddressesRequest {
						return awsec2.DescribeAddressesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withExternalName(allocationID)),
			},
			want: want{
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				}), withExternalName(allocationID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.elasticIP}
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
		cr     *v1alpha4.ElasticIP
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulVPC": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				elasticIP: &fake.MockElasticIPClient{
					MockAllocate: func(input *awsec2.AllocateAddressInput) awsec2.AllocateAddressRequest {
						return awsec2.AllocateAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AllocateAddressOutput{
								AllocationId: &allocationID,
							}},
						}
					},
				},
				cr: elasticIP(),
			},
			want: want{
				cr: elasticIP(withExternalName(allocationID),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"SuccessfulStandard": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				elasticIP: &fake.MockElasticIPClient{
					MockAllocate: func(input *awsec2.AllocateAddressInput) awsec2.AllocateAddressRequest {
						return awsec2.AllocateAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AllocateAddressOutput{
								PublicIp: &publicIP,
							}},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainStandard,
				})),
			},
			want: want{
				cr: elasticIP(withExternalName(publicIP),
					withConditions(runtimev1alpha1.Creating()),
					withSpec(v1alpha4.ElasticIPParameters{
						Domain: domainStandard,
					})),
			},
		},
		"CreateFail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				elasticIP: &fake.MockElasticIPClient{
					MockAllocate: func(input *awsec2.AllocateAddressInput) awsec2.AllocateAddressRequest {
						return awsec2.AllocateAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: elasticIP(),
			},
			want: want{
				cr:  elasticIP(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.elasticIP}
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
		cr     *v1alpha4.ElasticIP
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				elasticIP: &fake.MockElasticIPClient{

					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				})),
			},
			want: want{
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				})),
			},
		},
		"ModifyFailed": {
			args: args{
				elasticIP: &fake.MockElasticIPClient{
					MockCreateTagsRequest: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				})),
			},
			want: want{
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainVpc,
				})),
				err: errors.Wrap(errBoom, errCreateTags),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.elasticIP}
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

func TestRelease(t *testing.T) {
	type want struct {
		cr  *v1alpha4.ElasticIP
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulVPC": {
			args: args{
				elasticIP: &fake.MockElasticIPClient{
					MockRelease: func(input *awsec2.ReleaseAddressInput) awsec2.ReleaseAddressRequest {
						return awsec2.ReleaseAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ReleaseAddressOutput{}},
						}
					},
				},
				cr: elasticIP(),
			},
			want: want{
				cr: elasticIP(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"SuccessfulStandard": {
			args: args{
				elasticIP: &fake.MockElasticIPClient{
					MockRelease: func(input *awsec2.ReleaseAddressInput) awsec2.ReleaseAddressRequest {
						return awsec2.ReleaseAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.ReleaseAddressOutput{}},
						}
					},
				},
				cr: elasticIP(withSpec(v1alpha4.ElasticIPParameters{
					Domain: domainStandard,
				})),
			},
			want: want{
				cr: elasticIP(withConditions(runtimev1alpha1.Deleting()),
					withSpec(v1alpha4.ElasticIPParameters{
						Domain: domainStandard,
					}),
				),
			},
		},
		"DeleteFailed": {
			args: args{
				elasticIP: &fake.MockElasticIPClient{
					MockRelease: func(input *awsec2.ReleaseAddressInput) awsec2.ReleaseAddressRequest {
						return awsec2.ReleaseAddressRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: elasticIP(),
			},
			want: want{
				cr:  elasticIP(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.elasticIP}
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
		cr   *v1alpha4.ElasticIP
		kube client.Client
	}
	type want struct {
		cr  *v1alpha4.ElasticIP
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   elasticIP(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: elasticIP(withTags(resource.GetExternalTags(elasticIP()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   elasticIP(),
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
