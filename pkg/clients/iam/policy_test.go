package iam

import (
	"strings"
	"testing"

	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
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
		upToDate bool
		diff     string
		err      error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: awsclients.NewJSONFromRaw(document1),
				},
				version: iamtypes.PolicyVersion{
					Document: &document1,
				},
			},
			want: want{
				upToDate: true,
				diff:     "",
			},
		},
		"DifferentFields": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: awsclients.NewJSONFromRaw(document1),
				},
				version: iamtypes.PolicyVersion{
					Document: &document2,
				},
			},
			want: want{
				upToDate: false,
				diff: `  &policy.Policy{
  	Version: "2012-10-17",
  	ID:      "",
  	Statements: policy.StatementList{
  		{
  			SID:          "",
- 			Effect:       "Allow",
+ 			Effect:       "Deny",
  			Principal:    &{Service: {"eks.amazonaws.com"}},
  			NotPrincipal: nil,
  			... // 5 identical fields
  		},
  	},
  }
`,
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
				upToDate: false,
				diff: `  &policy.Policy{
- 	Version:    "",
+ 	Version:    "2012-10-17",
  	ID:         "",
- 	Statements: nil,
+ 	Statements: policy.StatementList{
+ 		{
+ 			Effect:    "Deny",
+ 			Principal: &policy.Principal{Service: policy.StringOrArray{...}},
+ 			Action:    policy.StringOrArray{"sts:AssumeRole"},
+ 		},
+ 	},
  }
`,
			},
		},
		"SameFieldsSingleAction": {
			args: args{
				p: v1beta1.PolicyParameters{
					Document: awsclients.NewJSONFromRaw(document3),
				},
				version: iamtypes.PolicyVersion{
					Document: &document3,
				},
			},
			want: want{
				upToDate: true,
				diff:     "",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			upToDate, gotDiff, err := IsPolicyUpToDate(tc.args.p, tc.args.version)
			gotDiff = strings.ReplaceAll(gotDiff, "\u00a0", " ")

			if diff := cmp.Diff(tc.want.diff, gotDiff); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, upToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
