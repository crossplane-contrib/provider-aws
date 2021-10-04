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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	svcapitypes "github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	unexpectedItem resource.Managed
	providerArn    = "arn:123"
	url            = "https://example.com"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.OpenIDConnectProviderClient
	cr  resource.Managed
}

type oidcProviderModifier func(provider *svcapitypes.OpenIDConnectProvider)

func withConditions(c ...xpv1.Condition) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Status.ConditionedStatus.Conditions = c }
}

func withURL(s string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Spec.ForProvider.URL = s }
}
func withExternalName(name string) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { meta.SetExternalName(r, name) }
}

func withAtProvider(s svcapitypes.OpenIDConnectProviderObservation) oidcProviderModifier {
	return func(r *svcapitypes.OpenIDConnectProvider) { r.Status.AtProvider = s }
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
				},
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
			},
			want: want{
				cr: oidcProvider(withURL(url),
					withExternalName(providerArn)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExistName": {
			args: args{
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
						return nil, iam.NewErrorNotFound()
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
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
		"ValidInput": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockCreateOpenIDConnectProvider: func(ctx context.Context, input *awsiam.CreateOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
						return &awsiam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: aws.String(providerArn)}, nil
					},
				},
				cr: oidcProvider(withURL(url)),
			},
			want: want{
				cr: oidcProvider(withURL(url), func(provider *svcapitypes.OpenIDConnectProvider) {
					meta.SetExternalName(provider, providerArn)
				}),
				result: managed.ExternalCreation{
					ExternalNameAssigned: true,
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
				err: awsclient.Wrap(errBoom, errUpdateThumbprint),
			},
		},
		"AddClientError": {
			args: args{
				iam: &fake.MockOpenIDConnectProviderClient{
					MockGetOpenIDConnectProvider: func(ctx context.Context, input *awsiam.GetOpenIDConnectProviderInput, opts []func(*awsiam.Options)) (*awsiam.GetOpenIDConnectProviderOutput, error) {
						return &awsiam.GetOpenIDConnectProviderOutput{
							ThumbprintList: []string{"a"},
						}, nil
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
				err: awsclient.Wrap(errBoom, errAddClientID),
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
				err: awsclient.Wrap(errBoom, errRemoveClientID),
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
				err: awsclient.Wrap(errBoom, errDelete),
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
