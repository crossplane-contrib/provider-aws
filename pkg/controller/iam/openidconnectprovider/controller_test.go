/*
Copyright 2021 The Crossplane Authors.

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

package openidconnectprovider

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	providerArn    = "arn:123"
	url            = "https://example.com"
	name           = "oidcProvider"

	errBoom = errors.New("boom")

	key1   = "foo1"
	value1 = "bar1"
	key2   = "foo2"
	value2 = "bar2"

	tagComparer = cmp.Comparer(func(expected, actual iamtypes.Tag) bool {
		return cmp.Equal(expected.Key, actual.Key) &&
			cmp.Equal(expected.Value, actual.Value)
	})

	createInputComparer = cmp.Comparer(func(expected, actual *awsiam.CreateOpenIDConnectProviderInput) bool {
		return cmp.Equal(expected.Url, actual.Url) &&
			cmp.Equal(expected.ClientIDList, actual.ClientIDList, test.EquateConditions()) &&
			cmp.Equal(expected.ThumbprintList, actual.ThumbprintList, test.EquateConditions()) &&
			cmp.Equal(expected.Tags, actual.Tags, tagComparer, sortIAMTags)
	})

	tagInputComparer = cmp.Comparer(func(expected, actual *awsiam.TagOpenIDConnectProviderInput) bool {
		return cmp.Equal(expected.OpenIDConnectProviderArn, actual.OpenIDConnectProviderArn) &&
			cmp.Equal(expected.Tags, actual.Tags, tagComparer, sortIAMTags)
	})

	untagInputComparer = cmp.Comparer(func(expected, actual *awsiam.UntagOpenIDConnectProviderInput) bool {
		return cmp.Equal(expected.OpenIDConnectProviderArn, actual.OpenIDConnectProviderArn) &&
			cmp.Equal(expected.TagKeys, actual.TagKeys, sortStrings)
	})

	sortTags = cmpopts.SortSlices(func(a, b v1beta1.Tag) bool {
		return a.Key > b.Key
	})
	sortIAMTags = cmpopts.SortSlices(func(a, b iamtypes.Tag) bool {
		return *a.Key > *b.Key
	})
	sortStrings = cmpopts.SortSlices(func(x, y string) bool {
		return x < y
	})
)

type args struct {
	iam  *fake.MockOpenIDConnectProviderClient
	kube client.Client
	cr   resource.Managed
}

type oidcProviderModifier func(provider *svcapitypes.OpenIDConnectProvider)

func withConditions(c ...xpv1.Condition) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Status.ConditionedStatus.Conditions = c }
}

func withURL(s string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Spec.ForProvider.URL = s }
}

func withName(name string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Name = name }
}

func withExternalName(name string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { meta.SetExternalName(r, name) }
}

func withAtProvider(s svcapitypes.OpenIDConnectProviderObservation) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Status.AtProvider = s }
}

func withTags(tagMaps ...map[string]string) oidcProviderModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.OpenIDConnectProvider) {
		r.Spec.ForProvider.Tags = tagList
	}
}

func withClientIDList(l []string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) {
		r.Spec.ForProvider.ClientIDList = l
	}
}

func withThumbprintList(l []string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) {
		r.Spec.ForProvider.ThumbprintList = l
	}
}

func oidcProvider(m ...oidcProviderModifier) *svcapitypes.OpenIDConnectProvider {
	cr := &svcapitypes.OpenIDConnectProvider{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	now := metav1.Now()
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
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
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return nil, errBoom
					},
					MockListOpenIDConnectProviders: func(ctx context.Context, input *awsiam.ListOpenIDConnectProvidersInput, opts []func(*awsiam.Options)) (*awsiam.ListOpenIDConnectProvidersOutput, error) {
						return &awsiam.ListOpenIDConnectProvidersOutput{}, nil
					},
				},
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"NoExternalNameExistingResource": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockListOpenIDConnectProviders: func(ctx context.Context, input *awsiam.ListOpenIDConnectProvidersInput, opts []func(*awsiam.Options)) (*awsiam.ListOpenIDConnectProvidersOutput, error) {
						return &awsiam.ListOpenIDConnectProvidersOutput{
							OpenIDConnectProviderList: []iamtypes.OpenIDConnectProviderListEntry{
								{Arn: aws.String(providerArn)},
							},
						}, nil
					},
					MockListOpenIDConnectProviderTags: func(ctx context.Context, input *awsiam.ListOpenIDConnectProviderTagsInput, opts []func(*awsiam.Options)) (*awsiam.ListOpenIDConnectProviderTagsOutput, error) {
						return &awsiam.ListOpenIDConnectProviderTagsOutput{
							Tags: []iamtypes.Tag{
								{Key: aws.String(resource.ExternalResourceTagKeyName), Value: aws.String(name)},
							},
						}, nil
					},
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							CreateDate: &now.Time,
						}, nil
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: oidcProvider(withName(name), withURL(url)),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withName(name),
					withExternalName(providerArn),
					withConditions(xpv1.Available()),
					withAtProvider(svcapitypes.OpenIDConnectProviderObservation{
						CreateDate: &now,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NoExternalNameClientError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockListOpenIDConnectProviders: func(ctx context.Context, input *awsiam.ListOpenIDConnectProvidersInput, opts []func(*awsiam.Options)) (*awsiam.ListOpenIDConnectProvidersOutput, error) {
						return nil, errBoom
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr:  oidcProvider(withURL(url)),
				err: errorutils.Wrap(errBoom, errList),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ResourceDoesNotExistName": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockListOpenIDConnectProviders: func(ctx context.Context, input *awsiam.ListOpenIDConnectProvidersInput, opts []func(*awsiam.Options)) (*awsiam.ListOpenIDConnectProvidersOutput, error) {
						return &awsiam.ListOpenIDConnectProvidersOutput{}, nil
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr: oidcProvider(withURL(url)),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ResourceDoesNotExistAWS": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return nil, &iamtypes.NoSuchEntityException{}
					},
				},
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ValidInput": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							CreateDate: &now.Time,
						}, nil
					},
				},
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn),
					withAtProvider(svcapitypes.OpenIDConnectProviderObservation{
						CreateDate: &now,
					}),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.iam}
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
		input  *awsiam.CreateOpenIDConnectProviderInput
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidInput": {
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
				iam: &fake.MockOpenIDConnectProviderClient{
					MockCreateOpenIDConnectProvider: func(ctx context.Context, input *awsiam.CreateOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
						return &awsiam.CreateOpenIDConnectProviderOutput{}, errBoom
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr:  oidcProvider(withURL(url)),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
		"ValidInput": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockCreateOpenIDConnectProvider: func(ctx context.Context, input *awsiam.CreateOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
						return &awsiam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: aws.String(providerArn)}, nil
					},
				},
				cr: oidcProvider(withURL(url),
					withThumbprintList([]string{"thumbs1", "thumbs2"}),
					withClientIDList([]string{"client1", "client2"}),
					withTags(map[string]string{key1: value1, key2: value2})),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withThumbprintList([]string{"thumbs1", "thumbs2"}),
					withClientIDList([]string{"client1", "client2"}),
					withTags(map[string]string{key1: value1, key2: value2}),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						meta.SetExternalName(provider, providerArn)
					}),
				result: managed.ExternalCreation{},
				input: &awsiam.CreateOpenIDConnectProviderInput{
					ThumbprintList: []string{"thumbs1", "thumbs2"},
					Url:            &url,
					ClientIDList:   []string{"client1", "client2"},
					Tags: []iamtypes.Tag{
						{
							Key:   &key1,
							Value: &value1,
						},
						{
							Key:   &key2,
							Value: &value2,
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), sortTags); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.want.input != nil {
				actual := tc.args.iam.MockOpenIDConnectProviderInput.CreateOIDCProviderInput
				if diff := cmp.Diff(tc.want.input, actual, createInputComparer, sortTags); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
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
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ThumbprintUpdateError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							ThumbprintList: []string{"b"},
						}, nil
					},
					MockUpdateOpenIDConnectProviderThumbprint: func(ctx context.Context, input *awsiam.UpdateOpenIDConnectProviderThumbprintInput, opts []func(*awsiam.Options)) (*awsiam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
						return &awsiam.UpdateOpenIDConnectProviderThumbprintOutput{}, errBoom
					},
				},
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ThumbprintList = []string{"a"}
					},
				),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ThumbprintList = []string{"a"}
					}),
				err: errorutils.Wrap(errBoom, errUpdateThumbprint),
			},
		},
		"AddClientError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{}, nil
					},
					MockAddClientIDToOpenIDConnectProvider: func(ctx context.Context, input *awsiam.AddClientIDToOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.AddClientIDToOpenIDConnectProviderOutput, error) {
						return &awsiam.AddClientIDToOpenIDConnectProviderOutput{}, errBoom
					},
				},
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ClientIDList = []string{"a", "b"}
					},
				),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ClientIDList = []string{"a", "b"}
					}),
				err: errorutils.Wrap(errBoom, errAddClientID),
			},
		},
		"RemoveClientError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							ClientIDList: []string{"a", "b"},
						}, nil
					},
					MockRemoveClientIDFromOpenIDConnectProvider: func(ctx context.Context, input *awsiam.RemoveClientIDFromOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.RemoveClientIDFromOpenIDConnectProviderOutput, error) {
						return &awsiam.RemoveClientIDFromOpenIDConnectProviderOutput{}, errBoom
					},
				},
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ClientIDList = []string{"a"}
					},
				),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ClientIDList = []string{"a"}
					}),
				err: errorutils.Wrap(errBoom, errRemoveClientID),
			},
		},
		"SuccessfulUpdate": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{

					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							ThumbprintList: []string{"b"},
							ClientIDList:   []string{"b"},
						}, nil
					},
					MockUpdateOpenIDConnectProviderThumbprint: func(ctx context.Context, input *awsiam.UpdateOpenIDConnectProviderThumbprintInput, opts []func(*awsiam.Options)) (*awsiam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
						return &awsiam.UpdateOpenIDConnectProviderThumbprintOutput{}, nil
					},
					MockAddClientIDToOpenIDConnectProvider: func(ctx context.Context, input *awsiam.AddClientIDToOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.AddClientIDToOpenIDConnectProviderOutput, error) {
						return &awsiam.AddClientIDToOpenIDConnectProviderOutput{}, nil
					},
					MockRemoveClientIDFromOpenIDConnectProvider: func(ctx context.Context, input *awsiam.RemoveClientIDFromOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.RemoveClientIDFromOpenIDConnectProviderOutput, error) {
						return &awsiam.RemoveClientIDFromOpenIDConnectProviderOutput{}, nil
					},
				},
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ThumbprintList = []string{"a"}
						provider.Spec.ForProvider.ClientIDList = []string{"a", "c"}
					},
				),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					func(provider *svcapitypes.OpenIDConnectProvider) {
						provider.Spec.ForProvider.ThumbprintList = []string{"a"}
						provider.Spec.ForProvider.ClientIDList = []string{"a", "c"}
					}),
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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

func TestUpdate_Tags(t *testing.T) {
	type want struct {
		cr         resource.Managed
		result     managed.ExternalUpdate
		err        error
		tagInput   *awsiam.TagOpenIDConnectProviderInput
		untagInput *awsiam.UntagOpenIDConnectProviderInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddTagsError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{}, nil
					},
					MockTagOpenIDConnectProvider: func(ctx context.Context, input *awsiam.TagOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.TagOpenIDConnectProviderOutput, error) {
						return nil, errBoom
					},
				},
				cr: oidcProvider(withTags(map[string]string{key1: value1})),
			},
			want: want{
				cr:  oidcProvider(withTags(map[string]string{key1: value1})),
				err: errorutils.Wrap(errBoom, errAddTags),
			},
		},
		"AddTagsSuccess": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{}, nil
					},
					MockTagOpenIDConnectProvider: func(ctx context.Context, input *awsiam.TagOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.TagOpenIDConnectProviderOutput, error) {
						return &awsiam.TagOpenIDConnectProviderOutput{}, nil
					},
				},
				cr: oidcProvider(
					withTags(map[string]string{key1: value1, key2: value2}),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(
					withTags(map[string]string{key1: value1, key2: value2}),
					withExternalName(providerArn)),
				tagInput: &awsiam.TagOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: &providerArn,
					Tags: []iamtypes.Tag{
						{Key: &key1, Value: &value1},
						{Key: &key2, Value: &value2},
					}},
			},
		},
		"UpdateTagsSuccess": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							Tags: []iamtypes.Tag{
								{Key: &key1, Value: &value1},
								{Key: &key2, Value: &value2},
							}}, nil
					},
					MockTagOpenIDConnectProvider: func(ctx context.Context, input *awsiam.TagOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.TagOpenIDConnectProviderOutput, error) {
						return &awsiam.TagOpenIDConnectProviderOutput{}, nil
					},
				},
				cr: oidcProvider(
					withTags(map[string]string{key1: value2, key2: value2}),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(
					withTags(map[string]string{key1: value2, key2: value2}),
					withExternalName(providerArn)),
				tagInput: &awsiam.TagOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: &providerArn,
					Tags: []iamtypes.Tag{
						{Key: &key1, Value: &value2},
					}},
			},
		},
		"RemoveTagsError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							Tags: []iamtypes.Tag{
								{Key: &key1, Value: &value1},
								{Key: &key2, Value: &value2},
							}}, nil
					},
					MockUntagOpenIDConnectProvider: func(ctx context.Context, input *awsiam.UntagOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.UntagOpenIDConnectProviderOutput, error) {
						return nil, errBoom
					},
				},
				cr: oidcProvider(withTags(map[string]string{key1: value1})),
			},
			want: want{
				cr:  oidcProvider(withTags(map[string]string{key1: value1})),
				err: errorutils.Wrap(errBoom, errRemoveTags),
			},
		},
		"RemoveTagsSuccess": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							Tags: []iamtypes.Tag{
								{Key: &key1, Value: &value1},
								{Key: &key2, Value: &value2},
							}}, nil
					},
					MockUntagOpenIDConnectProvider: func(ctx context.Context, input *awsiam.UntagOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.UntagOpenIDConnectProviderOutput, error) {
						return nil, nil
					},
				},
				cr: oidcProvider(withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(withExternalName(providerArn)),
				untagInput: &awsiam.UntagOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: &providerArn,
					TagKeys:                  []string{key1, key2},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), sortTags); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.want.tagInput != nil {
				if diff := cmp.Diff(tc.want.tagInput, tc.iam.MockOpenIDConnectProviderInput.TagOpenIDConnectProviderInput, tagInputComparer, sortIAMTags); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
			}
			if tc.want.untagInput != nil {
				if diff := cmp.Diff(tc.want.untagInput, tc.iam.MockOpenIDConnectProviderInput.UntagOpenIDConnectProviderInput, untagInputComparer, sortStrings); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
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
		"InvalidInput": {
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
				iam: &fake.MockOpenIDConnectProviderClient{
					MockDeleteOpenIDConnectProvider: func(ctx context.Context, input *awsiam.DeleteOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.DeleteOpenIDConnectProviderOutput, error) {
						return &awsiam.DeleteOpenIDConnectProviderOutput{}, errBoom
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr:  oidcProvider(withURL(url)),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ValidInput": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockDeleteOpenIDConnectProvider: func(ctx context.Context, input *awsiam.DeleteOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.DeleteOpenIDConnectProviderOutput, error) {
						return &awsiam.DeleteOpenIDConnectProviderOutput{}, nil
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr: oidcProvider(withURL(url)),
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
