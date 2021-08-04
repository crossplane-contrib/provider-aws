/*
Copyright 2020 The Crossplane Authors.
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

package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"

	clientset "github.com/crossplane/provider-aws/pkg/clients/route53resolver"
)

// this ensures that the mock implements the client interface
var _ clientset.ResolverRuleAssociationClient = (*MockResolverRuleAssociationClient)(nil)

// MockResolverRuleAssociationClient is a type that implements all the methods for ResolverRuleAssociationClient interface
type MockResolverRuleAssociationClient struct {
	MockGet          func(*route53resolver.GetResolverRuleAssociationInput) route53resolver.GetResolverRuleAssociationRequest
	MockAssociate    func(*route53resolver.AssociateResolverRuleInput) route53resolver.AssociateResolverRuleRequest
	MockDisassociate func(*route53resolver.DisassociateResolverRuleInput) route53resolver.DisassociateResolverRuleRequest
}

// GetResolverRuleAssociationRequest mocks GetResolverRuleAssociationRequest method
func (m *MockResolverRuleAssociationClient) GetResolverRuleAssociationRequest(input *route53resolver.GetResolverRuleAssociationInput) route53resolver.GetResolverRuleAssociationRequest {
	return m.MockGet(input)
}

// AssociateResolverRuleRequest mocks AssociateResolverRuleRequest method
func (m *MockResolverRuleAssociationClient) AssociateResolverRuleRequest(input *route53resolver.AssociateResolverRuleInput) route53resolver.AssociateResolverRuleRequest {
	return m.MockAssociate(input)
}

// DisassociateResolverRuleRequest mocks DisassociateResolverRuleRequest method
func (m *MockResolverRuleAssociationClient) DisassociateResolverRuleRequest(input *route53resolver.DisassociateResolverRuleInput) route53resolver.DisassociateResolverRuleRequest {
	return m.MockDisassociate(input)
}
