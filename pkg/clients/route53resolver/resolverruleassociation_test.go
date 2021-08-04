package route53resolver

import (
	"testing"

	awsr53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	id            = "association id"
	ruleid        = "rule id"
	vpcid         = "vpc id"
	statusmessage = "status message"
	status        = awsr53r.ResolverRuleAssociationStatusComplete
)

func TestGenerateResolverRuleAssociationObservation(t *testing.T) {
	cases := map[string]struct {
		in  awsr53r.ResolverRuleAssociation
		out v1alpha1.ResolverRuleAssociationObservation
	}{
		"AllFilled": {
			in: awsr53r.ResolverRuleAssociation{
				Id:             aws.String(id),
				ResolverRuleId: aws.String(ruleid),
				VPCId:          aws.String(vpcid),
				Status:         status,
				StatusMessage:  aws.String(statusmessage),
			},
			out: v1alpha1.ResolverRuleAssociationObservation{
				ID:            id,
				VPCID:         vpcid,
				RuleID:        ruleid,
				Status:        string(status),
				StatusMessage: statusmessage,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRoute53ResolverObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateRoute53ResolverObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
