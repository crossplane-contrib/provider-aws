// /*
// Copyright 2019 The Crossplane Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package hostedzone

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsroute53 "github.com/aws/aws-sdk-go-v2/service/route53"
	awsroute53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/hostedzone"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/hostedzone/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	uuid                 = "a96abeca-8da3-40fc-a2d5-08d72084eb65"
	errBoom              = errors.New("Some random error")
	id                   = "/hostedzone/XXXXXXXXXXXXXXXXXXX"
	rrCount        int64 = 2
	c                    = new(string)
	b                    = false
)

type zoneModifier func(*v1alpha1.HostedZone)

type args struct {
	kube    client.Client
	route53 hostedzone.Client
	cr      resource.Managed
}

func withExternalName(s string) zoneModifier {
	return func(r *v1alpha1.HostedZone) { meta.SetExternalName(r, s) }
}

func withConditions(c ...xpv1.Condition) zoneModifier {
	return func(r *v1alpha1.HostedZone) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(id string, rr int64) zoneModifier {
	return func(r *v1alpha1.HostedZone) {
		r.Status.AtProvider = v1alpha1.HostedZoneObservation{
			DelegationSet: v1alpha1.DelegationSet{
				NameServers: []string{
					"ns-2048.awsdns-64.com",
					"ns-2049.awsdns-65.net",
					"ns-2050.awsdns-66.org",
					"ns-2051.awsdns-67.co.uk",
				},
			},
			HostedZone: v1alpha1.HostedZoneResponse{
				CallerReference:        uuid,
				ID:                     id,
				ResourceRecordSetCount: rr,
			},
		}
	}
}

func withSpec(s v1alpha1.HostedZoneParameters) zoneModifier {
	return func(r *v1alpha1.HostedZone) {
		r.Spec.ForProvider = s
	}
}

func withComment(c string) zoneModifier {
	return func(r *v1alpha1.HostedZone) { r.Spec.ForProvider.Config.Comment = &c }
}

func instance(m ...zoneModifier) *v1alpha1.HostedZone {
	cr := &v1alpha1.HostedZone{
		Spec: v1alpha1.HostedZoneSpec{
			ForProvider: v1alpha1.HostedZoneParameters{
				Config: &v1alpha1.Config{
					Comment:     c,
					PrivateZone: &b,
				},
				Name: id,
			},
		},
	}
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
		"VaildInput": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				route53: &fake.MockHostedZoneClient{
					MockGetHostedZone: func(ctx context.Context, input *awsroute53.GetHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.GetHostedZoneOutput, error) {
						return &awsroute53.GetHostedZoneOutput{
							DelegationSet: &awsroute53types.DelegationSet{
								NameServers: []string{
									"ns-2048.awsdns-64.com",
									"ns-2049.awsdns-65.net",
									"ns-2050.awsdns-66.org",
									"ns-2051.awsdns-67.co.uk",
								},
							},
							HostedZone: &awsroute53types.HostedZone{
								CallerReference:        &uuid,
								Id:                     &id,
								ResourceRecordSetCount: &rrCount,
								Config: &awsroute53types.HostedZoneConfig{
									Comment:     c,
									PrivateZone: b,
								},
							},
							VPCs: make([]awsroute53types.VPC, 0),
						}, nil
					},
					MockListTagsForResource: func(ctx context.Context, params *awsroute53.ListTagsForResourceInput, opts []func(*awsroute53.Options)) (*awsroute53.ListTagsForResourceOutput, error) {
						return &awsroute53.ListTagsForResourceOutput{
							ResourceTagSet: &awsroute53types.ResourceTagSet{
								Tags: []awsroute53types.Tag{
									{
										Key:   aws.String("foo"),
										Value: aws.String("bar"),
									},
									{
										Key:   aws.String("hello"),
										Value: aws.String("world"),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withStatus(id, rrCount),
					withSpec(v1alpha1.HostedZoneParameters{
						Tags: map[string]string{
							"foo":   "bar",
							"hello": "world",
						},
						Config: &v1alpha1.Config{
							Comment:     c,
							PrivateZone: &b,
						},
					}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withStatus(id, rrCount),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.HostedZoneParameters{
						Tags: map[string]string{
							"foo":   "bar",
							"hello": "world",
						},
						Config: &v1alpha1.Config{
							Comment:     c,
							PrivateZone: &b,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DiffTags": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				route53: &fake.MockHostedZoneClient{
					MockGetHostedZone: func(ctx context.Context, input *awsroute53.GetHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.GetHostedZoneOutput, error) {
						return &awsroute53.GetHostedZoneOutput{
							DelegationSet: &awsroute53types.DelegationSet{
								NameServers: []string{
									"ns-2048.awsdns-64.com",
									"ns-2049.awsdns-65.net",
									"ns-2050.awsdns-66.org",
									"ns-2051.awsdns-67.co.uk",
								},
							},
							HostedZone: &awsroute53types.HostedZone{
								CallerReference:        &uuid,
								Id:                     &id,
								ResourceRecordSetCount: &rrCount,
								Config: &awsroute53types.HostedZoneConfig{
									Comment:     c,
									PrivateZone: b,
								},
							},
							VPCs: make([]awsroute53types.VPC, 0),
						}, nil
					},
					MockListTagsForResource: func(ctx context.Context, params *awsroute53.ListTagsForResourceInput, opts []func(*awsroute53.Options)) (*awsroute53.ListTagsForResourceOutput, error) {
						return &awsroute53.ListTagsForResourceOutput{
							ResourceTagSet: &awsroute53types.ResourceTagSet{
								Tags: []awsroute53types.Tag{
									{
										Key:   aws.String("foo"),
										Value: aws.String("bar"),
									},
									{
										Key:   aws.String("hello"),
										Value: aws.String("world"),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withStatus(id, rrCount),
					withSpec(v1alpha1.HostedZoneParameters{
						Tags: map[string]string{
							"hello": "world",
						},
						Config: &v1alpha1.Config{
							Comment:     c,
							PrivateZone: &b,
						},
					}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withStatus(id, rrCount),
					withConditions(xpv1.Available()),
					withSpec(v1alpha1.HostedZoneParameters{
						Tags: map[string]string{
							"hello": "world",
						},
						Config: &v1alpha1.Config{
							Comment:     c,
							PrivateZone: &b,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
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
		"ResourceDoesNotExist": {
			args: args{
				route53: &fake.MockHostedZoneClient{
					MockGetHostedZone: func(ctx context.Context, input *awsroute53.GetHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.GetHostedZoneOutput, error) {
						return nil, &awsroute53types.NoSuchHostedZone{}
					},
				},
				cr: instance(),
			},
			want: want{
				cr:     instance(),
				result: managed.ExternalObservation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: test.NewMockClient(), client: tc.route53}
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
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				route53: &fake.MockHostedZoneClient{
					MockCreateHostedZone: func(ctx context.Context, input *awsroute53.CreateHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.CreateHostedZoneOutput, error) {
						return &awsroute53.CreateHostedZoneOutput{
							DelegationSet: &awsroute53types.DelegationSet{
								NameServers: []string{
									"ns-2048.awsdns-64.com",
									"ns-2049.awsdns-65.net",
									"ns-2050.awsdns-66.org",
									"ns-2051.awsdns-67.co.uk",
								},
							},
							HostedZone: &awsroute53types.HostedZone{
								CallerReference:        &uuid,
								Id:                     &id,
								ResourceRecordSetCount: &rrCount,
								Config: &awsroute53types.HostedZoneConfig{
									Comment:     c,
									PrivateZone: b,
								},
							},
							Location: aws.String(fmt.Sprintf("%s%s", "https://route53.amazonaws.com/2013-04-01/", id)),
						}, nil
					},
				},
				cr: instance(withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1])),
			},
			want: want{
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1])),
				result: managed.ExternalCreation{},
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
				route53: &fake.MockHostedZoneClient{
					MockCreateHostedZone: func(ctx context.Context, input *awsroute53.CreateHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.CreateHostedZoneOutput, error) {
						return nil, errBoom
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: test.NewMockClient(), client: tc.route53}
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
	type args struct {
		route53      hostedzone.Client
		cr           resource.Managed
		tagsToAdd    []awsroute53types.Tag
		tagsToRemove []string
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				route53: &fake.MockHostedZoneClient{
					MockUpdateHostedZoneComment: func(ctx context.Context, input *awsroute53.UpdateHostedZoneCommentInput, opts []func(*awsroute53.Options)) (*awsroute53.UpdateHostedZoneCommentOutput, error) {
						return &awsroute53.UpdateHostedZoneCommentOutput{
							HostedZone: &awsroute53types.HostedZone{
								CallerReference:        &uuid,
								Id:                     &id,
								ResourceRecordSetCount: &rrCount,
								Config: &awsroute53types.HostedZoneConfig{
									Comment:     c,
									PrivateZone: b,
								},
							},
						}, nil
					},
					MockChangeTagsForResource: func(ctx context.Context, params *awsroute53.ChangeTagsForResourceInput, optFns []func(*awsroute53.Options)) (*awsroute53.ChangeTagsForResourceOutput, error) {
						expected := &awsroute53.ChangeTagsForResourceInput{
							ResourceType: awsroute53types.TagResourceTypeHostedzone,
							AddTags: []awsroute53types.Tag{
								{
									Key:   aws.String("foo"),
									Value: aws.String("bar"),
								},
								{
									Key:   aws.String("hello"),
									Value: aws.String("world"),
								},
							},
							RemoveTagKeys: []string{
								"hello",
							},
						}
						if diff := cmp.Diff(expected, params, cmpopts.IgnoreFields("ResourceID")); diff != "" {
							return nil, errors.Errorf("unexpected params: %s", diff)
						}
						return nil, nil
					},
				},
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withComment("New Comment"),
				),
				tagsToAdd: []awsroute53types.Tag{
					{
						Key:   aws.String("foo"),
						Value: aws.String("bar"),
					},
				},
				tagsToRemove: []string{
					"hello",
				},
			},
			want: want{
				cr: instance(
					withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
					withComment("New Comment"),
				),
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
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
		"VaildInput": {
			args: args{
				route53: &fake.MockHostedZoneClient{
					MockDeleteHostedZone: func(ctx context.Context, input *awsroute53.DeleteHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.DeleteHostedZoneOutput, error) {
						return &awsroute53.DeleteHostedZoneOutput{}, nil
					},
				},
				cr: instance(withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1])),
			},
			want: want{
				cr: instance(withExternalName(strings.SplitAfter(id, hostedzone.IDPrefix)[1]),
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
				route53: &fake.MockHostedZoneClient{
					MockDeleteHostedZone: func(ctx context.Context, input *awsroute53.DeleteHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.DeleteHostedZoneOutput, error) {
						return nil, errBoom
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				route53: &fake.MockHostedZoneClient{
					MockDeleteHostedZone: func(ctx context.Context, input *awsroute53.DeleteHostedZoneInput, opts []func(*awsroute53.Options)) (*awsroute53.DeleteHostedZoneOutput, error) {
						return nil, &awsroute53types.NoSuchHostedZone{}
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
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
