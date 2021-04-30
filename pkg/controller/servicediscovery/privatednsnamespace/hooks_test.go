package privatednsnamespace

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane/provider-aws/apis/servicediscovery/v1alpha1"
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   svcapitypes.PrivateDNSNamespace
		resp svcsdk.GetNamespaceOutput
	}

	type want struct {
		result bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				cr: svcapitypes.PrivateDNSNamespace{
					Spec: svcapitypes.PrivateDNSNamespaceSpec{
						ForProvider: svcapitypes.PrivateDNSNamespaceParameters{
							Name: aws.String("test-name"),
						},
					},
				},
				resp: svcsdk.GetNamespaceOutput{
					Namespace: &svcsdk.Namespace{
						Name: aws.String("test-name"),
					},
				},
			},
			want: want{
				result: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isUpToDate(&tc.args.cr, &tc.args.resp)
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
