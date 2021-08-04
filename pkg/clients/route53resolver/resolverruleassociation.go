package route53resolver

import (
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// ResolverRuleAssociationNotFound resolver rule association not found
	ResolverRuleAssociationNotFound = "ResourceNotFoundException"
)

// ResolverRuleAssociationClient is the external client used for ResolverRuleAssociation Custom Resource
type ResolverRuleAssociationClient interface {
	GetResolverRuleAssociationRequest(*route53resolver.GetResolverRuleAssociationInput) route53resolver.GetResolverRuleAssociationRequest
	AssociateResolverRuleRequest(*route53resolver.AssociateResolverRuleInput) route53resolver.AssociateResolverRuleRequest
	DisassociateResolverRuleRequest(*route53resolver.DisassociateResolverRuleInput) route53resolver.DisassociateResolverRuleRequest
}

// IsResolverRuleAssociationNotFoundErr returns true when the error is due to the ResolverRuleAssociation not existing
func IsResolverRuleAssociationNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == ResolverRuleAssociationNotFound {
			return true
		}
	}
	return false
}

// GenerateRoute53ResolverObservation is used to produce v1alpha1.ResolverRuleAssociationObservation from the aws ResolverRuleAssociation resource
func GenerateRoute53ResolverObservation(resolverruleassociation route53resolver.ResolverRuleAssociation) v1alpha1.ResolverRuleAssociationObservation {
	o := v1alpha1.ResolverRuleAssociationObservation{
		ID:            aws.StringValue(resolverruleassociation.Id),
		RuleID:        aws.StringValue(resolverruleassociation.ResolverRuleId),
		VPCID:         aws.StringValue(resolverruleassociation.VPCId),
		StatusMessage: aws.StringValue(resolverruleassociation.StatusMessage),
		Status:        string(resolverruleassociation.Status),
	}
	return o
}
