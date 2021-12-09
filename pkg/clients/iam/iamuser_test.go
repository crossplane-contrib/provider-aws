package iam

import (
	"testing"
	"time"

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
)

var (
	path   = "/"
	userID = "some id"
)

func userParams(m ...func(*v1beta1.IAMUserParameters)) *v1beta1.IAMUserParameters {
	o := &v1beta1.IAMUserParameters{
		Path: &path,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func user(m ...func(*iamtypes.User)) *iamtypes.User {
	o := &iamtypes.User{
		Path:   &path,
		UserId: &userID,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestLateInitializeUser(t *testing.T) {
	type args struct {
		spec *v1beta1.IAMUserParameters
		in   iamtypes.User
	}
	cases := map[string]struct {
		args args
		want *v1beta1.IAMUserParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: userParams(),
				in:   *user(),
			},
			want: userParams(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: userParams(),
				in: *user(func(r *iamtypes.User) {
					r.CreateDate = &time.Time{}
				}),
			},
			want: userParams(),
		},
		"PartialFilled": {
			args: args{
				spec: userParams(func(p *v1beta1.IAMUserParameters) {
					p.Path = nil
				}),
				in: *user(),
			},
			want: userParams(func(p *v1beta1.IAMUserParameters) {
				p.Path = &path
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeUser(tc.args.spec, &tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
