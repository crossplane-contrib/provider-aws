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

package fake

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudsearch/cloudsearchiface"
)

// this ensures that the mock implements the client interface
var _ svcsdkapi.CloudSearchAPI = (*MockCloudsearchClient)(nil)

// MockCloudsearchClient is a type that implements all the methods for CloudSearchAPI interface
type MockCloudsearchClient struct {
	MockBuildSuggesters                          func(*cloudsearch.BuildSuggestersInput) (*cloudsearch.BuildSuggestersOutput, error)
	MockBuildSuggestersWithContext               func(aws.Context, *cloudsearch.BuildSuggestersInput, ...request.Option) (*cloudsearch.BuildSuggestersOutput, error)
	MockBuildSuggestersRequest                   func(*cloudsearch.BuildSuggestersInput) (*request.Request, *cloudsearch.BuildSuggestersOutput)
	MockCreateDomain                             func(*cloudsearch.CreateDomainInput) (*cloudsearch.CreateDomainOutput, error)
	MockCreateDomainWithContext                  func(aws.Context, *cloudsearch.CreateDomainInput, ...request.Option) (*cloudsearch.CreateDomainOutput, error)
	MockCreateDomainRequest                      func(*cloudsearch.CreateDomainInput) (*request.Request, *cloudsearch.CreateDomainOutput)
	MockDefineAnalysisScheme                     func(*cloudsearch.DefineAnalysisSchemeInput) (*cloudsearch.DefineAnalysisSchemeOutput, error)
	MockDefineAnalysisSchemeWithContext          func(aws.Context, *cloudsearch.DefineAnalysisSchemeInput, ...request.Option) (*cloudsearch.DefineAnalysisSchemeOutput, error)
	MockDefineAnalysisSchemeRequest              func(*cloudsearch.DefineAnalysisSchemeInput) (*request.Request, *cloudsearch.DefineAnalysisSchemeOutput)
	MockDefineExpression                         func(*cloudsearch.DefineExpressionInput) (*cloudsearch.DefineExpressionOutput, error)
	MockDefineExpressionWithContext              func(aws.Context, *cloudsearch.DefineExpressionInput, ...request.Option) (*cloudsearch.DefineExpressionOutput, error)
	MockDefineExpressionRequest                  func(*cloudsearch.DefineExpressionInput) (*request.Request, *cloudsearch.DefineExpressionOutput)
	MockDefineIndexField                         func(*cloudsearch.DefineIndexFieldInput) (*cloudsearch.DefineIndexFieldOutput, error)
	MockDefineIndexFieldWithContext              func(aws.Context, *cloudsearch.DefineIndexFieldInput, ...request.Option) (*cloudsearch.DefineIndexFieldOutput, error)
	MockDefineIndexFieldRequest                  func(*cloudsearch.DefineIndexFieldInput) (*request.Request, *cloudsearch.DefineIndexFieldOutput)
	MockDefineSuggester                          func(*cloudsearch.DefineSuggesterInput) (*cloudsearch.DefineSuggesterOutput, error)
	MockDefineSuggesterWithContext               func(aws.Context, *cloudsearch.DefineSuggesterInput, ...request.Option) (*cloudsearch.DefineSuggesterOutput, error)
	MockDefineSuggesterRequest                   func(*cloudsearch.DefineSuggesterInput) (*request.Request, *cloudsearch.DefineSuggesterOutput)
	MockDeleteAnalysisScheme                     func(*cloudsearch.DeleteAnalysisSchemeInput) (*cloudsearch.DeleteAnalysisSchemeOutput, error)
	MockDeleteAnalysisSchemeWithContext          func(aws.Context, *cloudsearch.DeleteAnalysisSchemeInput, ...request.Option) (*cloudsearch.DeleteAnalysisSchemeOutput, error)
	MockDeleteAnalysisSchemeRequest              func(*cloudsearch.DeleteAnalysisSchemeInput) (*request.Request, *cloudsearch.DeleteAnalysisSchemeOutput)
	MockDeleteDomain                             func(*cloudsearch.DeleteDomainInput) (*cloudsearch.DeleteDomainOutput, error)
	MockDeleteDomainWithContext                  func(aws.Context, *cloudsearch.DeleteDomainInput, ...request.Option) (*cloudsearch.DeleteDomainOutput, error)
	MockDeleteDomainRequest                      func(*cloudsearch.DeleteDomainInput) (*request.Request, *cloudsearch.DeleteDomainOutput)
	MockDeleteExpression                         func(*cloudsearch.DeleteExpressionInput) (*cloudsearch.DeleteExpressionOutput, error)
	MockDeleteExpressionWithContext              func(aws.Context, *cloudsearch.DeleteExpressionInput, ...request.Option) (*cloudsearch.DeleteExpressionOutput, error)
	MockDeleteExpressionRequest                  func(*cloudsearch.DeleteExpressionInput) (*request.Request, *cloudsearch.DeleteExpressionOutput)
	MockDeleteIndexField                         func(*cloudsearch.DeleteIndexFieldInput) (*cloudsearch.DeleteIndexFieldOutput, error)
	MockDeleteIndexFieldWithContext              func(aws.Context, *cloudsearch.DeleteIndexFieldInput, ...request.Option) (*cloudsearch.DeleteIndexFieldOutput, error)
	MockDeleteIndexFieldRequest                  func(*cloudsearch.DeleteIndexFieldInput) (*request.Request, *cloudsearch.DeleteIndexFieldOutput)
	MockDeleteSuggester                          func(*cloudsearch.DeleteSuggesterInput) (*cloudsearch.DeleteSuggesterOutput, error)
	MockDeleteSuggesterWithContext               func(aws.Context, *cloudsearch.DeleteSuggesterInput, ...request.Option) (*cloudsearch.DeleteSuggesterOutput, error)
	MockDeleteSuggesterRequest                   func(*cloudsearch.DeleteSuggesterInput) (*request.Request, *cloudsearch.DeleteSuggesterOutput)
	MockDescribeAnalysisSchemes                  func(*cloudsearch.DescribeAnalysisSchemesInput) (*cloudsearch.DescribeAnalysisSchemesOutput, error)
	MockDescribeAnalysisSchemesWithContext       func(aws.Context, *cloudsearch.DescribeAnalysisSchemesInput, ...request.Option) (*cloudsearch.DescribeAnalysisSchemesOutput, error)
	MockDescribeAnalysisSchemesRequest           func(*cloudsearch.DescribeAnalysisSchemesInput) (*request.Request, *cloudsearch.DescribeAnalysisSchemesOutput)
	MockDescribeAvailabilityOptions              func(*cloudsearch.DescribeAvailabilityOptionsInput) (*cloudsearch.DescribeAvailabilityOptionsOutput, error)
	MockDescribeAvailabilityOptionsWithContext   func(aws.Context, *cloudsearch.DescribeAvailabilityOptionsInput, ...request.Option) (*cloudsearch.DescribeAvailabilityOptionsOutput, error)
	MockDescribeAvailabilityOptionsRequest       func(*cloudsearch.DescribeAvailabilityOptionsInput) (*request.Request, *cloudsearch.DescribeAvailabilityOptionsOutput)
	MockDescribeDomainEndpointOptions            func(*cloudsearch.DescribeDomainEndpointOptionsInput) (*cloudsearch.DescribeDomainEndpointOptionsOutput, error)
	MockDescribeDomainEndpointOptionsWithContext func(aws.Context, *cloudsearch.DescribeDomainEndpointOptionsInput, ...request.Option) (*cloudsearch.DescribeDomainEndpointOptionsOutput, error)
	MockDescribeDomainEndpointOptionsRequest     func(*cloudsearch.DescribeDomainEndpointOptionsInput) (*request.Request, *cloudsearch.DescribeDomainEndpointOptionsOutput)
	MockDescribeDomains                          func(*cloudsearch.DescribeDomainsInput) (*cloudsearch.DescribeDomainsOutput, error)
	MockDescribeDomainsWithContext               func(aws.Context, *cloudsearch.DescribeDomainsInput, ...request.Option) (*cloudsearch.DescribeDomainsOutput, error)
	MockDescribeDomainsRequest                   func(*cloudsearch.DescribeDomainsInput) (*request.Request, *cloudsearch.DescribeDomainsOutput)
	MockDescribeExpressions                      func(*cloudsearch.DescribeExpressionsInput) (*cloudsearch.DescribeExpressionsOutput, error)
	MockDescribeExpressionsWithContext           func(aws.Context, *cloudsearch.DescribeExpressionsInput, ...request.Option) (*cloudsearch.DescribeExpressionsOutput, error)
	MockDescribeExpressionsRequest               func(*cloudsearch.DescribeExpressionsInput) (*request.Request, *cloudsearch.DescribeExpressionsOutput)
	MockDescribeIndexFields                      func(*cloudsearch.DescribeIndexFieldsInput) (*cloudsearch.DescribeIndexFieldsOutput, error)
	MockDescribeIndexFieldsWithContext           func(aws.Context, *cloudsearch.DescribeIndexFieldsInput, ...request.Option) (*cloudsearch.DescribeIndexFieldsOutput, error)
	MockDescribeIndexFieldsRequest               func(*cloudsearch.DescribeIndexFieldsInput) (*request.Request, *cloudsearch.DescribeIndexFieldsOutput)
	MockDescribeScalingParameters                func(*cloudsearch.DescribeScalingParametersInput) (*cloudsearch.DescribeScalingParametersOutput, error)
	MockDescribeScalingParametersWithContext     func(aws.Context, *cloudsearch.DescribeScalingParametersInput, ...request.Option) (*cloudsearch.DescribeScalingParametersOutput, error)
	MockDescribeScalingParametersRequest         func(*cloudsearch.DescribeScalingParametersInput) (*request.Request, *cloudsearch.DescribeScalingParametersOutput)
	MockDescribeServiceAccessPolicies            func(*cloudsearch.DescribeServiceAccessPoliciesInput) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error)
	MockDescribeServiceAccessPoliciesWithContext func(aws.Context, *cloudsearch.DescribeServiceAccessPoliciesInput, ...request.Option) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error)
	MockDescribeServiceAccessPoliciesRequest     func(*cloudsearch.DescribeServiceAccessPoliciesInput) (*request.Request, *cloudsearch.DescribeServiceAccessPoliciesOutput)
	MockDescribeSuggesters                       func(*cloudsearch.DescribeSuggestersInput) (*cloudsearch.DescribeSuggestersOutput, error)
	MockDescribeSuggestersWithContext            func(aws.Context, *cloudsearch.DescribeSuggestersInput, ...request.Option) (*cloudsearch.DescribeSuggestersOutput, error)
	MockDescribeSuggestersRequest                func(*cloudsearch.DescribeSuggestersInput) (*request.Request, *cloudsearch.DescribeSuggestersOutput)
	MockIndexDocuments                           func(*cloudsearch.IndexDocumentsInput) (*cloudsearch.IndexDocumentsOutput, error)
	MockIndexDocumentsWithContext                func(aws.Context, *cloudsearch.IndexDocumentsInput, ...request.Option) (*cloudsearch.IndexDocumentsOutput, error)
	MockIndexDocumentsRequest                    func(*cloudsearch.IndexDocumentsInput) (*request.Request, *cloudsearch.IndexDocumentsOutput)
	MockListDomainNames                          func(*cloudsearch.ListDomainNamesInput) (*cloudsearch.ListDomainNamesOutput, error)
	MockListDomainNamesWithContext               func(aws.Context, *cloudsearch.ListDomainNamesInput, ...request.Option) (*cloudsearch.ListDomainNamesOutput, error)
	MockListDomainNamesRequest                   func(*cloudsearch.ListDomainNamesInput) (*request.Request, *cloudsearch.ListDomainNamesOutput)
	MockUpdateAvailabilityOptions                func(*cloudsearch.UpdateAvailabilityOptionsInput) (*cloudsearch.UpdateAvailabilityOptionsOutput, error)
	MockUpdateAvailabilityOptionsWithContext     func(aws.Context, *cloudsearch.UpdateAvailabilityOptionsInput, ...request.Option) (*cloudsearch.UpdateAvailabilityOptionsOutput, error)
	MockUpdateAvailabilityOptionsRequest         func(*cloudsearch.UpdateAvailabilityOptionsInput) (*request.Request, *cloudsearch.UpdateAvailabilityOptionsOutput)
	MockUpdateDomainEndpointOptions              func(*cloudsearch.UpdateDomainEndpointOptionsInput) (*cloudsearch.UpdateDomainEndpointOptionsOutput, error)
	MockUpdateDomainEndpointOptionsWithContext   func(aws.Context, *cloudsearch.UpdateDomainEndpointOptionsInput, ...request.Option) (*cloudsearch.UpdateDomainEndpointOptionsOutput, error)
	MockUpdateDomainEndpointOptionsRequest       func(*cloudsearch.UpdateDomainEndpointOptionsInput) (*request.Request, *cloudsearch.UpdateDomainEndpointOptionsOutput)
	MockUpdateScalingParameters                  func(*cloudsearch.UpdateScalingParametersInput) (*cloudsearch.UpdateScalingParametersOutput, error)
	MockUpdateScalingParametersWithContext       func(aws.Context, *cloudsearch.UpdateScalingParametersInput, ...request.Option) (*cloudsearch.UpdateScalingParametersOutput, error)
	MockUpdateScalingParametersRequest           func(*cloudsearch.UpdateScalingParametersInput) (*request.Request, *cloudsearch.UpdateScalingParametersOutput)
	MockUpdateServiceAccessPolicies              func(*cloudsearch.UpdateServiceAccessPoliciesInput) (*cloudsearch.UpdateServiceAccessPoliciesOutput, error)
	MockUpdateServiceAccessPoliciesWithContext   func(aws.Context, *cloudsearch.UpdateServiceAccessPoliciesInput, ...request.Option) (*cloudsearch.UpdateServiceAccessPoliciesOutput, error)
	MockUpdateServiceAccessPoliciesRequest       func(*cloudsearch.UpdateServiceAccessPoliciesInput) (*request.Request, *cloudsearch.UpdateServiceAccessPoliciesOutput)
}

// BuildSuggesters is a mock
func (m *MockCloudsearchClient) BuildSuggesters(in *cloudsearch.BuildSuggestersInput) (*cloudsearch.BuildSuggestersOutput, error) {
	return m.MockBuildSuggesters(in)
}

// BuildSuggestersWithContext is a mock
func (m *MockCloudsearchClient) BuildSuggestersWithContext(ctx aws.Context, in *cloudsearch.BuildSuggestersInput, opts ...request.Option) (*cloudsearch.BuildSuggestersOutput, error) {
	return m.MockBuildSuggestersWithContext(ctx, in, opts...)
}

// BuildSuggestersRequest is a mock
func (m *MockCloudsearchClient) BuildSuggestersRequest(in *cloudsearch.BuildSuggestersInput) (*request.Request, *cloudsearch.BuildSuggestersOutput) {
	return m.MockBuildSuggestersRequest(in)
}

// CreateDomain is a mock
func (m *MockCloudsearchClient) CreateDomain(in *cloudsearch.CreateDomainInput) (*cloudsearch.CreateDomainOutput, error) {
	return m.MockCreateDomain(in)
}

// CreateDomainWithContext is a mock
func (m *MockCloudsearchClient) CreateDomainWithContext(ctx aws.Context, in *cloudsearch.CreateDomainInput, opts ...request.Option) (*cloudsearch.CreateDomainOutput, error) {
	return m.MockCreateDomainWithContext(ctx, in, opts...)
}

// CreateDomainRequest is a mock
func (m *MockCloudsearchClient) CreateDomainRequest(in *cloudsearch.CreateDomainInput) (*request.Request, *cloudsearch.CreateDomainOutput) {
	return m.MockCreateDomainRequest(in)
}

// DefineAnalysisScheme is a mock
func (m *MockCloudsearchClient) DefineAnalysisScheme(in *cloudsearch.DefineAnalysisSchemeInput) (*cloudsearch.DefineAnalysisSchemeOutput, error) {
	return m.MockDefineAnalysisScheme(in)
}

// DefineAnalysisSchemeWithContext is a mock
func (m *MockCloudsearchClient) DefineAnalysisSchemeWithContext(ctx aws.Context, in *cloudsearch.DefineAnalysisSchemeInput, opts ...request.Option) (*cloudsearch.DefineAnalysisSchemeOutput, error) {
	return m.MockDefineAnalysisSchemeWithContext(ctx, in, opts...)
}

// DefineAnalysisSchemeRequest is a mock
func (m *MockCloudsearchClient) DefineAnalysisSchemeRequest(in *cloudsearch.DefineAnalysisSchemeInput) (*request.Request, *cloudsearch.DefineAnalysisSchemeOutput) {
	return m.MockDefineAnalysisSchemeRequest(in)
}

// DefineExpression is a mock
func (m *MockCloudsearchClient) DefineExpression(in *cloudsearch.DefineExpressionInput) (*cloudsearch.DefineExpressionOutput, error) {
	return m.MockDefineExpression(in)
}

// DefineExpressionWithContext is a mock
func (m *MockCloudsearchClient) DefineExpressionWithContext(ctx aws.Context, in *cloudsearch.DefineExpressionInput, opts ...request.Option) (*cloudsearch.DefineExpressionOutput, error) {
	return m.MockDefineExpressionWithContext(ctx, in, opts...)
}

// DefineExpressionRequest is a mock
func (m *MockCloudsearchClient) DefineExpressionRequest(in *cloudsearch.DefineExpressionInput) (*request.Request, *cloudsearch.DefineExpressionOutput) {
	return m.MockDefineExpressionRequest(in)
}

// DefineIndexField is a mock
func (m *MockCloudsearchClient) DefineIndexField(in *cloudsearch.DefineIndexFieldInput) (*cloudsearch.DefineIndexFieldOutput, error) {
	return m.MockDefineIndexField(in)
}

// DefineIndexFieldWithContext is a mock
func (m *MockCloudsearchClient) DefineIndexFieldWithContext(ctx aws.Context, in *cloudsearch.DefineIndexFieldInput, opts ...request.Option) (*cloudsearch.DefineIndexFieldOutput, error) {
	return m.MockDefineIndexFieldWithContext(ctx, in, opts...)
}

// DefineIndexFieldRequest is a mock
func (m *MockCloudsearchClient) DefineIndexFieldRequest(in *cloudsearch.DefineIndexFieldInput) (*request.Request, *cloudsearch.DefineIndexFieldOutput) {
	return m.MockDefineIndexFieldRequest(in)
}

// DefineSuggester is a mock
func (m *MockCloudsearchClient) DefineSuggester(in *cloudsearch.DefineSuggesterInput) (*cloudsearch.DefineSuggesterOutput, error) {
	return m.MockDefineSuggester(in)
}

// DefineSuggesterWithContext is a mock
func (m *MockCloudsearchClient) DefineSuggesterWithContext(ctx aws.Context, in *cloudsearch.DefineSuggesterInput, opts ...request.Option) (*cloudsearch.DefineSuggesterOutput, error) {
	return m.MockDefineSuggesterWithContext(ctx, in, opts...)
}

// DefineSuggesterRequest is a mock
func (m *MockCloudsearchClient) DefineSuggesterRequest(in *cloudsearch.DefineSuggesterInput) (*request.Request, *cloudsearch.DefineSuggesterOutput) {
	return m.MockDefineSuggesterRequest(in)
}

// DeleteAnalysisScheme is a mock
func (m *MockCloudsearchClient) DeleteAnalysisScheme(in *cloudsearch.DeleteAnalysisSchemeInput) (*cloudsearch.DeleteAnalysisSchemeOutput, error) {
	return m.MockDeleteAnalysisScheme(in)
}

// DeleteAnalysisSchemeWithContext is a mock
func (m *MockCloudsearchClient) DeleteAnalysisSchemeWithContext(ctx aws.Context, in *cloudsearch.DeleteAnalysisSchemeInput, opts ...request.Option) (*cloudsearch.DeleteAnalysisSchemeOutput, error) {
	return m.MockDeleteAnalysisSchemeWithContext(ctx, in, opts...)
}

// DeleteAnalysisSchemeRequest is a mock
func (m *MockCloudsearchClient) DeleteAnalysisSchemeRequest(in *cloudsearch.DeleteAnalysisSchemeInput) (*request.Request, *cloudsearch.DeleteAnalysisSchemeOutput) {
	return m.MockDeleteAnalysisSchemeRequest(in)
}

// DeleteDomain is a mock
func (m *MockCloudsearchClient) DeleteDomain(in *cloudsearch.DeleteDomainInput) (*cloudsearch.DeleteDomainOutput, error) {
	return m.MockDeleteDomain(in)
}

// DeleteDomainWithContext is a mock
func (m *MockCloudsearchClient) DeleteDomainWithContext(ctx aws.Context, in *cloudsearch.DeleteDomainInput, opts ...request.Option) (*cloudsearch.DeleteDomainOutput, error) {
	return m.MockDeleteDomainWithContext(ctx, in, opts...)
}

// DeleteDomainRequest is a mock
func (m *MockCloudsearchClient) DeleteDomainRequest(in *cloudsearch.DeleteDomainInput) (*request.Request, *cloudsearch.DeleteDomainOutput) {
	return m.MockDeleteDomainRequest(in)
}

// DeleteExpression is a mock
func (m *MockCloudsearchClient) DeleteExpression(in *cloudsearch.DeleteExpressionInput) (*cloudsearch.DeleteExpressionOutput, error) {
	return m.MockDeleteExpression(in)
}

// DeleteExpressionWithContext is a mock
func (m *MockCloudsearchClient) DeleteExpressionWithContext(ctx aws.Context, in *cloudsearch.DeleteExpressionInput, opts ...request.Option) (*cloudsearch.DeleteExpressionOutput, error) {
	return m.MockDeleteExpressionWithContext(ctx, in, opts...)
}

// DeleteExpressionRequest is a mock
func (m *MockCloudsearchClient) DeleteExpressionRequest(in *cloudsearch.DeleteExpressionInput) (*request.Request, *cloudsearch.DeleteExpressionOutput) {
	return m.MockDeleteExpressionRequest(in)
}

// DeleteIndexField is a mock
func (m *MockCloudsearchClient) DeleteIndexField(in *cloudsearch.DeleteIndexFieldInput) (*cloudsearch.DeleteIndexFieldOutput, error) {
	return m.MockDeleteIndexField(in)
}

// DeleteIndexFieldWithContext is a mock
func (m *MockCloudsearchClient) DeleteIndexFieldWithContext(ctx aws.Context, in *cloudsearch.DeleteIndexFieldInput, opts ...request.Option) (*cloudsearch.DeleteIndexFieldOutput, error) {
	return m.MockDeleteIndexFieldWithContext(ctx, in, opts...)
}

// DeleteIndexFieldRequest is a mock
func (m *MockCloudsearchClient) DeleteIndexFieldRequest(in *cloudsearch.DeleteIndexFieldInput) (*request.Request, *cloudsearch.DeleteIndexFieldOutput) {
	return m.MockDeleteIndexFieldRequest(in)
}

// DeleteSuggester is a mock
func (m *MockCloudsearchClient) DeleteSuggester(in *cloudsearch.DeleteSuggesterInput) (*cloudsearch.DeleteSuggesterOutput, error) {
	return m.MockDeleteSuggester(in)
}

// DeleteSuggesterWithContext is a mock
func (m *MockCloudsearchClient) DeleteSuggesterWithContext(ctx aws.Context, in *cloudsearch.DeleteSuggesterInput, opts ...request.Option) (*cloudsearch.DeleteSuggesterOutput, error) {
	return m.MockDeleteSuggesterWithContext(ctx, in, opts...)
}

// DeleteSuggesterRequest is a mock
func (m *MockCloudsearchClient) DeleteSuggesterRequest(in *cloudsearch.DeleteSuggesterInput) (*request.Request, *cloudsearch.DeleteSuggesterOutput) {
	return m.MockDeleteSuggesterRequest(in)
}

// DescribeAnalysisSchemes is a mock
func (m *MockCloudsearchClient) DescribeAnalysisSchemes(in *cloudsearch.DescribeAnalysisSchemesInput) (*cloudsearch.DescribeAnalysisSchemesOutput, error) {
	return m.MockDescribeAnalysisSchemes(in)
}

// DescribeAnalysisSchemesWithContext is a mock
func (m *MockCloudsearchClient) DescribeAnalysisSchemesWithContext(ctx aws.Context, in *cloudsearch.DescribeAnalysisSchemesInput, opts ...request.Option) (*cloudsearch.DescribeAnalysisSchemesOutput, error) {
	return m.MockDescribeAnalysisSchemesWithContext(ctx, in, opts...)
}

// DescribeAnalysisSchemesRequest is a mock
func (m *MockCloudsearchClient) DescribeAnalysisSchemesRequest(in *cloudsearch.DescribeAnalysisSchemesInput) (*request.Request, *cloudsearch.DescribeAnalysisSchemesOutput) {
	return m.MockDescribeAnalysisSchemesRequest(in)
}

// DescribeAvailabilityOptions is a mock
func (m *MockCloudsearchClient) DescribeAvailabilityOptions(in *cloudsearch.DescribeAvailabilityOptionsInput) (*cloudsearch.DescribeAvailabilityOptionsOutput, error) {
	return m.MockDescribeAvailabilityOptions(in)
}

// DescribeAvailabilityOptionsWithContext is a mock
func (m *MockCloudsearchClient) DescribeAvailabilityOptionsWithContext(ctx aws.Context, in *cloudsearch.DescribeAvailabilityOptionsInput, opts ...request.Option) (*cloudsearch.DescribeAvailabilityOptionsOutput, error) {
	return m.MockDescribeAvailabilityOptionsWithContext(ctx, in, opts...)
}

// DescribeAvailabilityOptionsRequest is a mock
func (m *MockCloudsearchClient) DescribeAvailabilityOptionsRequest(in *cloudsearch.DescribeAvailabilityOptionsInput) (*request.Request, *cloudsearch.DescribeAvailabilityOptionsOutput) {
	return m.MockDescribeAvailabilityOptionsRequest(in)
}

// DescribeDomainEndpointOptions is a mock
func (m *MockCloudsearchClient) DescribeDomainEndpointOptions(in *cloudsearch.DescribeDomainEndpointOptionsInput) (*cloudsearch.DescribeDomainEndpointOptionsOutput, error) {
	return m.MockDescribeDomainEndpointOptions(in)
}

// DescribeDomainEndpointOptionsWithContext is a mock
func (m *MockCloudsearchClient) DescribeDomainEndpointOptionsWithContext(ctx aws.Context, in *cloudsearch.DescribeDomainEndpointOptionsInput, opts ...request.Option) (*cloudsearch.DescribeDomainEndpointOptionsOutput, error) {
	return m.MockDescribeDomainEndpointOptionsWithContext(ctx, in, opts...)
}

// DescribeDomainEndpointOptionsRequest is a mock
func (m *MockCloudsearchClient) DescribeDomainEndpointOptionsRequest(in *cloudsearch.DescribeDomainEndpointOptionsInput) (*request.Request, *cloudsearch.DescribeDomainEndpointOptionsOutput) {
	return m.MockDescribeDomainEndpointOptionsRequest(in)
}

// DescribeDomains is a mock
func (m *MockCloudsearchClient) DescribeDomains(in *cloudsearch.DescribeDomainsInput) (*cloudsearch.DescribeDomainsOutput, error) {
	return m.MockDescribeDomains(in)
}

// DescribeDomainsWithContext is a mock
func (m *MockCloudsearchClient) DescribeDomainsWithContext(ctx aws.Context, in *cloudsearch.DescribeDomainsInput, opts ...request.Option) (*cloudsearch.DescribeDomainsOutput, error) {
	return m.MockDescribeDomainsWithContext(ctx, in, opts...)
}

// DescribeDomainsRequest is a mock
func (m *MockCloudsearchClient) DescribeDomainsRequest(in *cloudsearch.DescribeDomainsInput) (*request.Request, *cloudsearch.DescribeDomainsOutput) {
	return m.MockDescribeDomainsRequest(in)
}

// DescribeExpressions is a mock
func (m *MockCloudsearchClient) DescribeExpressions(in *cloudsearch.DescribeExpressionsInput) (*cloudsearch.DescribeExpressionsOutput, error) {
	return m.MockDescribeExpressions(in)
}

// DescribeExpressionsWithContext is a mock
func (m *MockCloudsearchClient) DescribeExpressionsWithContext(ctx aws.Context, in *cloudsearch.DescribeExpressionsInput, opts ...request.Option) (*cloudsearch.DescribeExpressionsOutput, error) {
	return m.MockDescribeExpressionsWithContext(ctx, in, opts...)
}

// DescribeExpressionsRequest is a mock
func (m *MockCloudsearchClient) DescribeExpressionsRequest(in *cloudsearch.DescribeExpressionsInput) (*request.Request, *cloudsearch.DescribeExpressionsOutput) {
	return m.MockDescribeExpressionsRequest(in)
}

// DescribeIndexFields is a mock
func (m *MockCloudsearchClient) DescribeIndexFields(in *cloudsearch.DescribeIndexFieldsInput) (*cloudsearch.DescribeIndexFieldsOutput, error) {
	return m.MockDescribeIndexFields(in)
}

// DescribeIndexFieldsWithContext is a mock
func (m *MockCloudsearchClient) DescribeIndexFieldsWithContext(ctx aws.Context, in *cloudsearch.DescribeIndexFieldsInput, opts ...request.Option) (*cloudsearch.DescribeIndexFieldsOutput, error) {
	return m.MockDescribeIndexFieldsWithContext(ctx, in, opts...)
}

// DescribeIndexFieldsRequest is a mock
func (m *MockCloudsearchClient) DescribeIndexFieldsRequest(in *cloudsearch.DescribeIndexFieldsInput) (*request.Request, *cloudsearch.DescribeIndexFieldsOutput) {
	return m.MockDescribeIndexFieldsRequest(in)
}

// DescribeScalingParameters is a mock
func (m *MockCloudsearchClient) DescribeScalingParameters(in *cloudsearch.DescribeScalingParametersInput) (*cloudsearch.DescribeScalingParametersOutput, error) {
	return m.MockDescribeScalingParameters(in)
}

// DescribeScalingParametersWithContext is a mock
func (m *MockCloudsearchClient) DescribeScalingParametersWithContext(ctx aws.Context, in *cloudsearch.DescribeScalingParametersInput, opts ...request.Option) (*cloudsearch.DescribeScalingParametersOutput, error) {
	return m.MockDescribeScalingParametersWithContext(ctx, in, opts...)
}

// DescribeScalingParametersRequest is a mock
func (m *MockCloudsearchClient) DescribeScalingParametersRequest(in *cloudsearch.DescribeScalingParametersInput) (*request.Request, *cloudsearch.DescribeScalingParametersOutput) {
	return m.MockDescribeScalingParametersRequest(in)
}

// DescribeServiceAccessPolicies is a mock
func (m *MockCloudsearchClient) DescribeServiceAccessPolicies(in *cloudsearch.DescribeServiceAccessPoliciesInput) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error) {
	return m.MockDescribeServiceAccessPolicies(in)
}

// DescribeServiceAccessPoliciesWithContext is a mock
func (m *MockCloudsearchClient) DescribeServiceAccessPoliciesWithContext(ctx aws.Context, in *cloudsearch.DescribeServiceAccessPoliciesInput, opts ...request.Option) (*cloudsearch.DescribeServiceAccessPoliciesOutput, error) {
	return m.MockDescribeServiceAccessPoliciesWithContext(ctx, in, opts...)
}

// DescribeServiceAccessPoliciesRequest is a mock
func (m *MockCloudsearchClient) DescribeServiceAccessPoliciesRequest(in *cloudsearch.DescribeServiceAccessPoliciesInput) (*request.Request, *cloudsearch.DescribeServiceAccessPoliciesOutput) {
	return m.MockDescribeServiceAccessPoliciesRequest(in)
}

// DescribeSuggesters is a mock
func (m *MockCloudsearchClient) DescribeSuggesters(in *cloudsearch.DescribeSuggestersInput) (*cloudsearch.DescribeSuggestersOutput, error) {
	return m.MockDescribeSuggesters(in)
}

// DescribeSuggestersWithContext is a mock
func (m *MockCloudsearchClient) DescribeSuggestersWithContext(ctx aws.Context, in *cloudsearch.DescribeSuggestersInput, opts ...request.Option) (*cloudsearch.DescribeSuggestersOutput, error) {
	return m.MockDescribeSuggestersWithContext(ctx, in, opts...)
}

// DescribeSuggestersRequest is a mock
func (m *MockCloudsearchClient) DescribeSuggestersRequest(in *cloudsearch.DescribeSuggestersInput) (*request.Request, *cloudsearch.DescribeSuggestersOutput) {
	return m.MockDescribeSuggestersRequest(in)
}

// IndexDocuments is a mock
func (m *MockCloudsearchClient) IndexDocuments(in *cloudsearch.IndexDocumentsInput) (*cloudsearch.IndexDocumentsOutput, error) {
	return m.MockIndexDocuments(in)
}

// IndexDocumentsWithContext is a mock
func (m *MockCloudsearchClient) IndexDocumentsWithContext(ctx aws.Context, in *cloudsearch.IndexDocumentsInput, opts ...request.Option) (*cloudsearch.IndexDocumentsOutput, error) {
	return m.MockIndexDocumentsWithContext(ctx, in, opts...)
}

// IndexDocumentsRequest is a mock
func (m *MockCloudsearchClient) IndexDocumentsRequest(in *cloudsearch.IndexDocumentsInput) (*request.Request, *cloudsearch.IndexDocumentsOutput) {
	return m.MockIndexDocumentsRequest(in)
}

// ListDomainNames is a mock
func (m *MockCloudsearchClient) ListDomainNames(in *cloudsearch.ListDomainNamesInput) (*cloudsearch.ListDomainNamesOutput, error) {
	return m.MockListDomainNames(in)
}

// ListDomainNamesWithContext is a mock
func (m *MockCloudsearchClient) ListDomainNamesWithContext(ctx aws.Context, in *cloudsearch.ListDomainNamesInput, opts ...request.Option) (*cloudsearch.ListDomainNamesOutput, error) {
	return m.MockListDomainNamesWithContext(ctx, in, opts...)
}

// ListDomainNamesRequest is a mock
func (m *MockCloudsearchClient) ListDomainNamesRequest(in *cloudsearch.ListDomainNamesInput) (*request.Request, *cloudsearch.ListDomainNamesOutput) {
	return m.MockListDomainNamesRequest(in)
}

// UpdateAvailabilityOptions is a mock
func (m *MockCloudsearchClient) UpdateAvailabilityOptions(in *cloudsearch.UpdateAvailabilityOptionsInput) (*cloudsearch.UpdateAvailabilityOptionsOutput, error) {
	return m.MockUpdateAvailabilityOptions(in)
}

// UpdateAvailabilityOptionsWithContext is a mock
func (m *MockCloudsearchClient) UpdateAvailabilityOptionsWithContext(ctx aws.Context, in *cloudsearch.UpdateAvailabilityOptionsInput, opts ...request.Option) (*cloudsearch.UpdateAvailabilityOptionsOutput, error) {
	return m.MockUpdateAvailabilityOptionsWithContext(ctx, in, opts...)
}

// UpdateAvailabilityOptionsRequest is a mock
func (m *MockCloudsearchClient) UpdateAvailabilityOptionsRequest(in *cloudsearch.UpdateAvailabilityOptionsInput) (*request.Request, *cloudsearch.UpdateAvailabilityOptionsOutput) {
	return m.MockUpdateAvailabilityOptionsRequest(in)
}

// UpdateDomainEndpointOptions is a mock
func (m *MockCloudsearchClient) UpdateDomainEndpointOptions(in *cloudsearch.UpdateDomainEndpointOptionsInput) (*cloudsearch.UpdateDomainEndpointOptionsOutput, error) {
	return m.MockUpdateDomainEndpointOptions(in)
}

// UpdateDomainEndpointOptionsWithContext is a mock
func (m *MockCloudsearchClient) UpdateDomainEndpointOptionsWithContext(ctx aws.Context, in *cloudsearch.UpdateDomainEndpointOptionsInput, opts ...request.Option) (*cloudsearch.UpdateDomainEndpointOptionsOutput, error) {
	return m.MockUpdateDomainEndpointOptionsWithContext(ctx, in, opts...)
}

// UpdateDomainEndpointOptionsRequest is a mock
func (m *MockCloudsearchClient) UpdateDomainEndpointOptionsRequest(in *cloudsearch.UpdateDomainEndpointOptionsInput) (*request.Request, *cloudsearch.UpdateDomainEndpointOptionsOutput) {
	return m.MockUpdateDomainEndpointOptionsRequest(in)
}

// UpdateScalingParameters is a mock
func (m *MockCloudsearchClient) UpdateScalingParameters(in *cloudsearch.UpdateScalingParametersInput) (*cloudsearch.UpdateScalingParametersOutput, error) {
	return m.MockUpdateScalingParameters(in)
}

// UpdateScalingParametersWithContext is a mock
func (m *MockCloudsearchClient) UpdateScalingParametersWithContext(ctx aws.Context, in *cloudsearch.UpdateScalingParametersInput, opts ...request.Option) (*cloudsearch.UpdateScalingParametersOutput, error) {
	return m.MockUpdateScalingParametersWithContext(ctx, in, opts...)
}

// UpdateScalingParametersRequest is a mock
func (m *MockCloudsearchClient) UpdateScalingParametersRequest(in *cloudsearch.UpdateScalingParametersInput) (*request.Request, *cloudsearch.UpdateScalingParametersOutput) {
	return m.MockUpdateScalingParametersRequest(in)
}

// UpdateServiceAccessPolicies is a mock
func (m *MockCloudsearchClient) UpdateServiceAccessPolicies(in *cloudsearch.UpdateServiceAccessPoliciesInput) (*cloudsearch.UpdateServiceAccessPoliciesOutput, error) {
	return m.MockUpdateServiceAccessPolicies(in)
}

// UpdateServiceAccessPoliciesWithContext is a mock
func (m *MockCloudsearchClient) UpdateServiceAccessPoliciesWithContext(ctx aws.Context, in *cloudsearch.UpdateServiceAccessPoliciesInput, opts ...request.Option) (*cloudsearch.UpdateServiceAccessPoliciesOutput, error) {
	return m.MockUpdateServiceAccessPoliciesWithContext(ctx, in, opts...)
}

// UpdateServiceAccessPoliciesRequest is a mock
func (m *MockCloudsearchClient) UpdateServiceAccessPoliciesRequest(in *cloudsearch.UpdateServiceAccessPoliciesInput) (*request.Request, *cloudsearch.UpdateServiceAccessPoliciesOutput) {
	return m.MockUpdateServiceAccessPoliciesRequest(in)
}
