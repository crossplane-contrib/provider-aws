package group

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
)

type functionModifier func(*svcapitypes.Group)

func withSpec(p svcapitypes.GroupParameters) functionModifier {
	return func(r *svcapitypes.Group) { r.Spec.ForProvider = p }
}

func group(m ...functionModifier) *svcapitypes.Group {
	cr := &svcapitypes.Group{}
	cr.Name = "test-group-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

type args struct {
	cr   *svcapitypes.Group
	resp *svcsdk.GetGroupOutput
}

var (
	testDescription        string = "description"
	testDescriptionChanged string = "new description"
	precedence             int64  = 1
	precedenceChanged      int64  = 2
)

func TestIsUpToDate(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpToDate": {
			args: args{
				cr: group(withSpec(svcapitypes.GroupParameters{
					Description: &testDescription,
					Precedence:  &precedence,
				})),
				resp: &svcsdk.GetGroupOutput{Group: &svcsdk.GroupType{
					Description: &testDescription,
					Precedence:  &precedence,
				}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ChangedDescription": {
			args: args{
				cr: group(withSpec(svcapitypes.GroupParameters{
					Description: &testDescriptionChanged,
				})),
				resp: &svcsdk.GetGroupOutput{Group: &svcsdk.GroupType{Description: &testDescription}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedPrecedence": {
			args: args{
				cr: group(withSpec(svcapitypes.GroupParameters{
					Precedence: &precedenceChanged,
				})),
				resp: &svcsdk.GetGroupOutput{Group: &svcsdk.GroupType{Precedence: &precedence}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, _, err := isUpToDate(context.Background(), tc.args.cr, tc.args.resp)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
