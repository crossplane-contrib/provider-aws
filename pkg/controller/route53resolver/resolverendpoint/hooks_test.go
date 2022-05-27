package resolverendpoint

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	r53r "github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1"
)

var (
	creatorRequestID = "creator request id"
	ip               = "ip"
	subnetID         = "subnet id"
	securityGroupID  = "security group id"
)

func TestPreCreate(t *testing.T) {
	type args struct {
		cr *v1alpha1.ResolverEndpoint
	}

	type want struct {
		result *r53r.CreateResolverEndpointInput
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Success": {
			args: args{
				cr: &v1alpha1.ResolverEndpoint{
					Spec: v1alpha1.ResolverEndpointSpec{
						ForProvider: v1alpha1.ResolverEndpointParameters{
							CustomResolverEndpointParameters: v1alpha1.CustomResolverEndpointParameters{
								SecurityGroupIDs: []string{securityGroupID},
								IPAddresses: []*v1alpha1.IPAddressRequest{{
									IP:       aws.String(ip),
									SubnetID: aws.String(subnetID)},
								},
							},
						},
					},
					ObjectMeta: v1.ObjectMeta{UID: types.UID(creatorRequestID)},
				},
			},
			want: want{
				result: &r53r.CreateResolverEndpointInput{
					CreatorRequestId: aws.String(creatorRequestID),
					IpAddresses:      []*r53r.IpAddressRequest{{Ip: aws.String(ip), SubnetId: aws.String(subnetID)}},
					SecurityGroupIds: []*string{&securityGroupID},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := &r53r.CreateResolverEndpointInput{}
			preCreate(context.TODO(), tc.args.cr, result)
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
