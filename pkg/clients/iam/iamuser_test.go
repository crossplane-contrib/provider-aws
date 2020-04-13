package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

var (
	path      = "/"
	userID    = "some id"
	groupName = "some group"
)

func userParams(m ...func(*v1alpha1.UserParameters)) *v1alpha1.UserParameters {
	o := &v1alpha1.UserParameters{
		Path:      &path,
		GroupList: []string{groupName},
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
		spec *v1alpha1.UserParameters
		in   iam.User
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.UserParameters
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
				spec: userParams(func(p *v1alpha1.UserParameters) {
					p.Path = nil
				}),
				in: *user(),
			},
			want: userParams(func(p *v1alpha1.UserParameters) {
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

func TestCompareGroups(t *testing.T) {
	type args struct {
		spec *v1alpha1.UserParameters
		in   []iam.Group
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"SameGroups": {
			args: args{
				spec: userParams(),
				in: []iam.Group{
					{
						GroupName: &groupName,
					},
				},
			},
			want: true,
		},
		"DifferentGroups": {
			args: args{
				spec: userParams(),
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := CompareGroups(*tc.args.spec, tc.args.in)
			if diff := cmp.Diff(result, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}
