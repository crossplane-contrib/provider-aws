package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

var (
	path   = "/"
	userID = "some id"
)

func userParams(m ...func(*v1alpha1.IAMUserParameters)) *v1alpha1.IAMUserParameters {
	o := &v1alpha1.IAMUserParameters{
		Path: &path,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func user(m ...func(*iam.User)) *iam.User {
	o := &iam.User{
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
		spec *v1alpha1.IAMUserParameters
		in   iam.User
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.IAMUserParameters
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
				in: *user(func(r *iam.User) {
					r.CreateDate = &time.Time{}
				}),
			},
			want: userParams(),
		},
		"PartialFilled": {
			args: args{
				spec: userParams(func(p *v1alpha1.IAMUserParameters) {
					p.Path = nil
				}),
				in: *user(),
			},
			want: userParams(func(p *v1alpha1.IAMUserParameters) {
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
