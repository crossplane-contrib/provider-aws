package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

func TestLateInitializeUserPolicy(t *testing.T) {
	type args struct {
		spec v1alpha1.UserPolicyAttachmentParameters
		in   iam.AttachedPolicy
	}
	cases := map[string]struct {
		args args
		want v1alpha1.UserPolicyAttachmentParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: v1alpha1.UserPolicyAttachmentParameters{
					PolicyARN: policyARN,
				},
				in: iam.AttachedPolicy{
					PolicyArn: &policyARN,
				},
			},
			want: v1alpha1.UserPolicyAttachmentParameters{
				PolicyARN: policyARN,
			},
		},
		"PartialFilled": {
			args: args{
				spec: v1alpha1.UserPolicyAttachmentParameters{
					PolicyARN: policyARN,
				},
				in: iam.AttachedPolicy{
					PolicyArn: &policyARN,
				},
			},
			want: v1alpha1.UserPolicyAttachmentParameters{
				PolicyARN: policyARN,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeUserPolicy(&tc.args.spec, &tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
