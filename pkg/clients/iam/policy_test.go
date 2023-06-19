package iam

import (
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
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
	document3 = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": ["sts:AssumeRole"]
		  }
		]
	  }`
)

func TestIsPolicyUpToDate(t *testing.T) {
	type args struct {
		p       v1beta1.PolicyParameters
		version iamtypes.PolicyVersion
	}
	type want struct {
		areEqual bool
		err      error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document1,
				},
			},
			want: want{
				areEqual: true,
			},
		},
		"DifferentFields": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document2,
				},
			},
			want: want{
				areEqual: false,
			},
		},
		"EmptyPolicy": {
			args: args{
				p: v1beta1.PolicyParameters{},
				version: iamtypes.PolicyVersion{
					Document: &document2,
				},
			},
			want: want{
				areEqual: false,
			},
		},
		"SameFieldsSingleAction": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document3,
				},
			},
			want: want{
				areEqual: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			areEqual, diff, err := IsPolicyUpToDate(tc.args.p, tc.args.version)
			if diff := cmp.Diff(tc.want.areEqual, areEqual); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff != "" {
				t.Logf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
