package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

var (
	document1 = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`

	document2 = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Deny",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`
)

func TestIsPolicyUpToDate(t *testing.T) {
	type args struct {
		p       v1alpha1.PolicyParameters
		version iam.PolicyVersion
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: v1alpha1.PolicyParameters{
					Document: document1,
				},
				version: iam.PolicyVersion{
					Document: &document1,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: v1alpha1.PolicyParameters{
					Document: document1,
				},
				version: iam.PolicyVersion{
					Document: &document2,
				},
			},
			want: false,
		},
		"EmptyPolicy": {
			args: args{
				p: v1alpha1.PolicyParameters{},
				version: iam.PolicyVersion{
					Document: &document2,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsPolicyUpToDate(tc.args.p, tc.args.version)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
