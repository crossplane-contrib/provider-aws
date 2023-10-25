/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package associateresolverrule

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53resolver "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	route53resolvertypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane-contrib/provider-aws/apis/route53resolver/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// Client defines AssociateResolverRule operations
type Client interface {
	AssociateResolverRule(ctx context.Context, input *route53resolver.AssociateResolverRuleInput, opts ...func(*route53resolver.Options)) (*route53resolver.AssociateResolverRuleOutput, error)
	DisassociateResolverRule(ctx context.Context, input *route53resolver.DisassociateResolverRuleInput, opts ...func(*route53resolver.Options)) (*route53resolver.DisassociateResolverRuleOutput, error)
	GetResolverRuleAssociation(ctx context.Context, input *route53resolver.GetResolverRuleAssociationInput, opts ...func(*route53resolver.Options)) (*route53resolver.GetResolverRuleAssociationOutput, error)
}

// NewRoute53ResolverClient creates new AWS client with provided AWS Configuration/Credentials
func NewRoute53ResolverClient(cfg aws.Config) Client {
	return route53resolver.NewFromConfig(cfg)
}

// GenerateCreateAssociateResolverRuleInput returns a route53resolver AssociateResolverRuleOutput
func GenerateCreateAssociateResolverRuleInput(cr *manualv1alpha1.ResolverRuleAssociation) *route53resolver.AssociateResolverRuleInput {
	reqInput := &route53resolver.AssociateResolverRuleInput{
		Name:           pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		VPCId:          cr.Spec.ForProvider.VPCId,
		ResolverRuleId: cr.Spec.ForProvider.ResolverRuleID,
	}
	return reqInput
}

// GenerateDeleteAssociateResolverRuleInput returns a route53resolver DisassociateResolverRuleOutput
func GenerateDeleteAssociateResolverRuleInput(cr *manualv1alpha1.ResolverRuleAssociation) *route53resolver.DisassociateResolverRuleInput {
	reqInput := &route53resolver.DisassociateResolverRuleInput{
		VPCId:          cr.Spec.ForProvider.VPCId,
		ResolverRuleId: cr.Spec.ForProvider.ResolverRuleID,
	}
	return reqInput
}

// GenerateGetAssociateResolverRuleAssociationInput returns a route53resolver AssociateResolverRule
func GenerateGetAssociateResolverRuleAssociationInput(id *string) *route53resolver.GetResolverRuleAssociationInput {
	reqInput := &route53resolver.GetResolverRuleAssociationInput{
		ResolverRuleAssociationId: id,
	}
	return reqInput
}

// IsNotFound returns true if the error code indicates that the requested AssociateResolverRule was not found
func IsNotFound(err error) bool {
	var nshz *route53resolvertypes.ResourceNotFoundException
	return errors.As(err, &nshz)
}
