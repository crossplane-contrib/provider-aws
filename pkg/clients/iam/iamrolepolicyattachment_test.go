package iam

import (
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
)

var (
	policyARN = "some arn"
)

func policyParams(m ...func(*v1beta1.IAMRolePolicyAttachmentParameters)) *v1beta1.IAMRolePolicyAttachmentParameters {
	o := &v1beta1.IAMRolePolicyAttachmentParameters{
		PolicyARN: policyARN,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func policy(m ...func(*iamtypes.AttachedPolicy)) *iamtypes.AttachedPolicy {
	o := &iamtypes.AttachedPolicy{
		PolicyArn: &policyARN,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func policyObservation(m ...func(*v1beta1.IAMRolePolicyAttachmentExternalStatus)) *v1beta1.IAMRolePolicyAttachmentExternalStatus {
	o := &v1beta1.IAMRolePolicyAttachmentExternalStatus{
		AttachedPolicyARN: policyARN,
	}

	for _, f := range m {
		f(o)
	}

	return o
}
func TestGeneratePolicyObservation(t *testing.T) {
	cases := map[string]struct {
		in  iamtypes.AttachedPolicy
		out v1beta1.IAMRolePolicyAttachmentExternalStatus
	}{
		"AllFilled": {
			in:  *policy(),
			out: *policyObservation(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRolePolicyObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializePolicy(t *testing.T) {
	type args struct {
		spec *v1beta1.IAMRolePolicyAttachmentParameters
		in   iamtypes.AttachedPolicy
	}
	cases := map[string]struct {
		args args
		want *v1beta1.IAMRolePolicyAttachmentParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: policyParams(),
				in:   *policy(),
			},
			want: policyParams(),
		},
		"PartialFilled": {
			args: args{
				spec: policyParams(func(p *v1beta1.IAMRolePolicyAttachmentParameters) {
					p.PolicyARN = ""
				}),
				in: *policy(),
			},
			want: policyParams(func(p *v1beta1.IAMRolePolicyAttachmentParameters) {
				p.PolicyARN = policyARN
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializePolicy(tc.args.spec, &tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
