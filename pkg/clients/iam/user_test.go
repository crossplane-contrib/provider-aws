package iam

import (
	"testing"
	"time"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

var (
	path   = "/"
	userID = "some id"
)

func userParams(m ...func(*v1beta1.UserParameters)) *v1beta1.UserParameters {
	o := &v1beta1.UserParameters{
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
		spec *v1beta1.UserParameters
		in   iamtypes.User
	}
	cases := map[string]struct {
		args args
		want *v1beta1.UserParameters
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
				spec: userParams(func(p *v1beta1.UserParameters) {
					p.Path = nil
				}),
				in: *user(),
			},
			want: userParams(func(p *v1beta1.UserParameters) {
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
