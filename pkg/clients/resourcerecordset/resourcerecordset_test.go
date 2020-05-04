package resourcerecordset

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

func TestCreatePatch(t *testing.T) {

	resourceRecordSetName := "x.y.z."
	var ttl int64 = 300
	var ttl2 int64 = 200

	type args struct {
		rrSet route53.ResourceRecordSet
		p     v1alpha3.ResourceRecordSetParameters
	}

	type want struct {
		patch *v1alpha3.ResourceRecordSetParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				rrSet: route53.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha3.ResourceRecordSetParameters{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
			},
			want: want{
				patch: &v1alpha3.ResourceRecordSetParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				rrSet: route53.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha3.ResourceRecordSetParameters{
					Name: &resourceRecordSetName,
					TTL:  &ttl2,
				},
			},
			want: want{
				patch: &v1alpha3.ResourceRecordSetParameters{
					TTL: &ttl2,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(&tc.args.rrSet, &tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {

	resourceRecordSetName := "x.y.z."
	var ttl int64 = 300
	var ttl2 int64 = 200

	type args struct {
		rrSet route53.ResourceRecordSet
		p     v1alpha3.ResourceRecordSetParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				rrSet: route53.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha3.ResourceRecordSetParameters{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				rrSet: route53.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha3.ResourceRecordSetParameters{
					Name: &resourceRecordSetName,
					TTL:  &ttl2,
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				rrSet: route53.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha3.ResourceRecordSetParameters{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.rrSet)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
