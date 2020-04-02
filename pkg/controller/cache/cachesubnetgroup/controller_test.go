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
package cachesubnetgroup

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache/fake"
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
	sgDescription = "some description"
	subnetID      = "some ID"

	// replaceMe = "replace-me!"
	errBoom = errors.New("boom")
)

type args struct {
	cache elasticache.Client
	cr    *v1alpha1.CacheSubnetGroup
}

type csgModifier func(*v1alpha1.CacheSubnetGroup)

func withConditions(c ...runtimev1alpha1.Condition) csgModifier {
	return func(r *v1alpha1.CacheSubnetGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.CacheSubnetGroupParameters) csgModifier {
	return func(r *v1alpha1.CacheSubnetGroup) { r.Spec.ForProvider = p }
}

func csg(m ...csgModifier) *v1alpha1.CacheSubnetGroup {
	cr := &v1alpha1.CacheSubnetGroup{
		Spec: v1alpha1.CacheSubnetGroupSpec{
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (elasticache.Client, error)
		cr          *v1alpha1.CacheSubnetGroup
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i elasticache.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: csg(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i elasticache.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: csg(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: csg(),
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
				cr: csg(),
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
				cr: csg(),
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
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroupsRequest: func(input *awscache.DescribeCacheSubnetGroupsInput) awscache.DescribeCacheSubnetGroupsRequest {
						return awscache.DescribeCacheSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awscache.DescribeCacheSubnetGroupsOutput{
								CacheSubnetGroups: []awscache.CacheSubnetGroup{{}},
							}},
						}
					},
				},
				cr: csg(),
			},
			want: want{
				cr: csg(withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"UpToDate": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroupsRequest: func(input *awscache.DescribeCacheSubnetGroupsInput) awscache.DescribeCacheSubnetGroupsRequest {
						return awscache.DescribeCacheSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awscache.DescribeCacheSubnetGroupsOutput{
								CacheSubnetGroups: []awscache.CacheSubnetGroup{{
									CacheSubnetGroupDescription: aws.String(sgDescription),
									Subnets: []awscache.Subnet{
										{
											SubnetIdentifier: aws.String(subnetID),
										},
									},
								}},
							}},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					Description: sgDescription,
					SubnetIds:   []string{subnetID},
				})),
			},
			want: want{
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					Description: sgDescription,
					SubnetIds:   []string{subnetID},
				}), withConditions(runtimev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DescribeFail": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroupsRequest: func(input *awscache.DescribeCacheSubnetGroupsInput) awscache.DescribeCacheSubnetGroupsRequest {
						return awscache.DescribeCacheSubnetGroupsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: csg(),
			},
			want: want{
				cr:  csg(),
				err: errors.Wrap(errBoom, errDescribeSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockCreateCacheSubnetGroupRequest: func(input *awscache.CreateCacheSubnetGroupInput) awscache.CreateCacheSubnetGroupRequest {
						return awscache.CreateCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awscache.CreateCacheSubnetGroupOutput{}},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(runtimev1alpha1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				cache: &fake.MockClient{
					MockCreateCacheSubnetGroupRequest: func(input *awscache.CreateCacheSubnetGroupInput) awscache.CreateCacheSubnetGroupRequest {
						return awscache.CreateCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheSubnetGroupRequest: func(input *awscache.ModifyCacheSubnetGroupInput) awscache.ModifyCacheSubnetGroupRequest {
						return awscache.ModifyCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awscache.ModifyCacheSubnetGroupOutput{}},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				}))),
			},
		},
		"ModifyFailed": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheSubnetGroupRequest: func(input *awscache.ModifyCacheSubnetGroupInput) awscache.ModifyCacheSubnetGroupRequest {
						return awscache.ModifyCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				}))),
				err: errors.Wrap(errBoom, errModifySubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr  *v1alpha1.CacheSubnetGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheSubnetGroupRequest: func(input *awscache.DeleteCacheSubnetGroupInput) awscache.DeleteCacheSubnetGroupRequest {
						return awscache.DeleteCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awscache.DeleteCacheSubnetGroupOutput{}},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				}), withConditions(runtimev1alpha1.Deleting())),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheSubnetGroupRequest: func(input *awscache.DeleteCacheSubnetGroupInput) awscache.DeleteCacheSubnetGroupRequest {
						return awscache.DeleteCacheSubnetGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIds:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
