package userpooldomain

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
)

func TestPostObserve(t *testing.T) {
	type args struct {
		obj *svcsdk.DescribeUserPoolDomainOutput
		err error
	}
	type want struct {
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"DetectNonExistentResource": {
			args: args{
				obj: &svcsdk.DescribeUserPoolDomainOutput{
					DomainDescription: &svcsdk.DomainDescriptionType{},
				},
				err: nil,
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			cr := svcapitypes.UserPoolDomain{}
			result, err := postObserve(context.Background(), &cr, tc.args.obj, managed.ExternalObservation{}, tc.args.err)

			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
