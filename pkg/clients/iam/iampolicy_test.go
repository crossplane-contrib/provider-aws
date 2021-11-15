package iam

import (
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
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
		p       v1alpha1.IAMPolicyParameters
		version iamtypes.PolicyVersion
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: v1alpha1.IAMPolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document1,
				},
			},
			want: true,
		},
		"SameFieldsSingleAction": {
			args: args{
				p: v1alpha1.IAMPolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document3,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: v1alpha1.IAMPolicyParameters{
					Document: document1,
				},
				version: iamtypes.PolicyVersion{
					Document: &document2,
				},
			},
			want: false,
		},
		"EmptyPolicy": {
			args: args{
				p: v1alpha1.IAMPolicyParameters{},
				version: iamtypes.PolicyVersion{
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

func TestReplaceActionArray(t *testing.T) {
	type args struct {
		p string
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"ActionString": {
			args: args{
				p: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":\"sts:AssumeRole\"}]}",
			},
			want: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":\"sts:AssumeRole\"}]}",
		},
		"ActionSingleItem": {
			args: args{
				p: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":[\"sts:AssumeRole\"]}]}",
			},
			want: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":\"sts:AssumeRole\"}]}",
		},
		"ActionMultipleItems": {
			args: args{
				p: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":[\"sts:AssumeRole\",\"sts:*\"]}]}",
			},
			want: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"eks.amazonaws.com\"},\"Action\":[\"sts:AssumeRole\",\"sts:*\"]}]}",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := replaceActionArray(tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
