package resolverrule

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53r "github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
)

var (
	creatorRequestID = "creator request id"
)

func TestPreCreate(t *testing.T) {
	type args struct {
		cr *v1alpha1.ResolverRule
	}

	type want struct {
		result *r53r.CreateResolverRuleInput
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				cr: &v1alpha1.ResolverRule{ObjectMeta: v1.ObjectMeta{UID: types.UID(creatorRequestID)}},
			},
			want: want{
				result: &r53r.CreateResolverRuleInput{
					CreatorRequestId: aws.String(creatorRequestID),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := &r53r.CreateResolverRuleInput{}
			preCreate(context.TODO(), tc.args.cr, result)
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
