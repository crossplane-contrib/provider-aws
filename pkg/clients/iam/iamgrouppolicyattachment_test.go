package iam

import (
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

func TestLateInitializeGroupPolicy(t *testing.T) {
	type args struct {
		spec v1alpha1.IAMGroupPolicyAttachmentParameters
		in   iamtypes.AttachedPolicy
	}
	cases := map[string]struct {
		args args
		want v1alpha1.IAMGroupPolicyAttachmentParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: v1alpha1.IAMGroupPolicyAttachmentParameters{
					PolicyARN: policyARN,
				},
				in: iamtypes.AttachedPolicy{
					PolicyArn: &policyARN,
				},
			},
			want: v1alpha1.IAMGroupPolicyAttachmentParameters{
				PolicyARN: policyARN,
			},
		},
		"PartialFilled": {
			args: args{
				spec: v1alpha1.IAMGroupPolicyAttachmentParameters{
					PolicyARN: policyARN,
				},
				in: iamtypes.AttachedPolicy{
					PolicyArn: &policyARN,
				},
			},
			want: v1alpha1.IAMGroupPolicyAttachmentParameters{
				PolicyARN: policyARN,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeGroupPolicy(&tc.args.spec, &tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
