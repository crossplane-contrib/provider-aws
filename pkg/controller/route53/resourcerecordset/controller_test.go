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
package resourcerecordset

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/resourcerecordset"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/resourcerecordset/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	errBoom        = errors.New("Some random error")
	rrName         = "crossplane.io"
	rrtype         = "A"
	TTL            = aws.Int64(300)
	rRecords       = make([]v1alpha1.ResourceRecord, 1)
	zoneID         = aws.String("/hostedzone/XXXXXXXXXXXXXXXXXXX")

	changeFn = func(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
		return &route53.ChangeResourceRecordSetsOutput{}, nil
	}
	changeErrFn = func(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
		return nil, errBoom
	}
)

type rrModifier func(*v1alpha1.ResourceRecordSet)

type args struct {
	kube    client.Client
	route53 resourcerecordset.Client
	cr      resource.Managed
}

func withConditions(c ...xpv1.Condition) rrModifier {
	return func(r *v1alpha1.ResourceRecordSet) { r.Status.ConditionedStatus.Conditions = c }
}

func instance(m ...rrModifier) *v1alpha1.ResourceRecordSet {
	for i := range rRecords {
		rRecords[i].Value = "0.0.0.0"
	}
	cr := &v1alpha1.ResourceRecordSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: rrName,
		},
		Spec: v1alpha1.ResourceRecordSetSpec{
			ForProvider: v1alpha1.ResourceRecordSetParameters{
				Type:            rrtype,
				TTL:             TTL,
				ResourceRecords: rRecords,
				ZoneID:          zoneID,
			},
		},
	}
	meta.SetExternalName(cr, cr.GetName())
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {

	name := rrName + "."
	rrSet := route53types.ResourceRecordSet{
		Name: &name,
		Type: route53types.RRType("A"),
		TTL:  TTL,
		ResourceRecords: []route53types.ResourceRecord{
			{
				Value: aws.String("0.0.0.0"),
			},
		},
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	addWildcardPrefix := func(rr *v1alpha1.ResourceRecordSet) {
		meta.SetExternalName(rr, fmt.Sprintf("*.%s", rr.Name))
	}
	addTwoWildcardPrefix := func(rr *v1alpha1.ResourceRecordSet) {
		meta.SetExternalName(rr, fmt.Sprintf("*.test*.%s", rr.Name))
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
				route53: &fake.MockResourceRecordSetClient{
					MockListResourceRecordSets: func(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
						return &route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []route53types.ResourceRecordSet{rrSet},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
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
				route53: &fake.MockResourceRecordSetClient{
					MockListResourceRecordSets: func(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
						return &route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []route53types.ResourceRecordSet{{
								Name: aws.String(""),
								Type: route53types.RRType(""),
								TTL:  aws.Int64(0),
								ResourceRecords: []route53types.ResourceRecord{
									{
										Value: aws.String(""),
									},
								},
							}},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"WildcardRecord": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				route53: &fake.MockResourceRecordSetClient{
					MockListResourceRecordSets: func(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
						return &route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []route53types.ResourceRecordSet{
								{
									Name: aws.String(fmt.Sprintf("*.%s", name)),
									Type: "A",
									TTL:  TTL,
									ResourceRecords: []route53types.ResourceRecord{
										{
											Value: aws.String("0.0.0.0"),
										},
									},
								},
							},
						}, nil
					},
				},
				cr: instance(addWildcardPrefix),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Available()),
					addWildcardPrefix,
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"RecordWithTwoWildCards": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockSubResourceUpdateFn(nil),
				},
				route53: &fake.MockResourceRecordSetClient{
					MockListResourceRecordSets: func(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
						return &route53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []route53types.ResourceRecordSet{
								{
									Name: aws.String(fmt.Sprintf("*.test*.%s", name)),
									Type: "A",
									TTL:  TTL,
									ResourceRecords: []route53types.ResourceRecord{
										{
											Value: aws.String("0.0.0.0"),
										},
									},
								},
							},
						}, nil
					},
				},
				cr: instance(addTwoWildcardPrefix),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Available()),
					addTwoWildcardPrefix,
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
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
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeFn,
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Creating())),
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
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeErrFn,
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.route53}
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
		"ValidInput": {
			args: args{
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeFn,
				},
				cr: instance(),
			},
			want: want{
				cr: instance(),
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
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeErrFn,
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: errorutils.Wrap(errBoom, errUpdate),
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
		"ValidInput": {
			args: args{
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeFn,
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Deleting())),
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
				route53: &fake.MockResourceRecordSetClient{
					MockChangeResourceRecordSets: changeErrFn,
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
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
