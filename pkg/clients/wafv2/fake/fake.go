package fake

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/aws/aws-sdk-go/service/wafv2/wafv2iface"
)

var _ wafv2iface.WAFV2API = (*MockWAFV2Client)(nil)

// MockWAFV2Client is a type that implements all the methods for WAFV2API interface
type MockWAFV2Client struct {
	MockAssociateWebACL                                  func(input *svcsdk.AssociateWebACLInput) (*svcsdk.AssociateWebACLOutput, error)
	MockAssociateWebACLWithContext                       func(input *svcsdk.AssociateWebACLInput) (*svcsdk.AssociateWebACLOutput, error)
	MockAssociateWebACLRequest                           func(input *svcsdk.AssociateWebACLInput) (*request.Request, *svcsdk.AssociateWebACLOutput)
	MockCheckCapacity                                    func(input *svcsdk.CheckCapacityInput) (*svcsdk.CheckCapacityOutput, error)
	MockCheckCapacityWithContext                         func(input *svcsdk.CheckCapacityInput) (*svcsdk.CheckCapacityOutput, error)
	MockCapacityRequest                                  func(input *svcsdk.CheckCapacityInput) (*request.Request, *svcsdk.CheckCapacityOutput)
	MockCreateAPIKey                                     func(input *svcsdk.CreateAPIKeyInput) (*svcsdk.CreateAPIKeyOutput, error)
	MockCreateAPIKeyWithContext                          func(input *svcsdk.CreateAPIKeyInput) (*svcsdk.CreateAPIKeyOutput, error)
	MockCreateAPIKeyRequest                              func(input *svcsdk.CreateAPIKeyInput) (*request.Request, *svcsdk.CreateAPIKeyOutput)
	MockDeleteAPIKey                                     func(input *svcsdk.DeleteAPIKeyInput) (*svcsdk.DeleteAPIKeyOutput, error)
	MockDeleteAPIKeyWithContext                          func(input *svcsdk.DeleteAPIKeyInput) (*svcsdk.DeleteAPIKeyOutput, error)
	MockDeleteAPIKeyRequest                              func(input *svcsdk.DeleteAPIKeyInput) (*request.Request, *svcsdk.DeleteAPIKeyOutput)
	MockCreateIPSet                                      func(input *svcsdk.CreateIPSetInput) (*svcsdk.CreateIPSetOutput, error)
	MockCreateIPSetWithContext                           func(input *svcsdk.CreateIPSetInput) (*svcsdk.CreateIPSetOutput, error)
	MockCreateIPSetRequest                               func(input *svcsdk.CreateIPSetInput) (*request.Request, *svcsdk.CreateIPSetOutput)
	MockCreateRegexPatternSet                            func(input *svcsdk.CreateRegexPatternSetInput) (*svcsdk.CreateRegexPatternSetOutput, error)
	MockCreateRegexPatternSetWithContext                 func(input *svcsdk.CreateRegexPatternSetInput) (*svcsdk.CreateRegexPatternSetOutput, error)
	MockCreateRegexPatternSetRequest                     func(input *svcsdk.CreateRegexPatternSetInput) (*request.Request, *svcsdk.CreateRegexPatternSetOutput)
	MockCreateRuleGroup                                  func(input *svcsdk.CreateRuleGroupInput) (*svcsdk.CreateRuleGroupOutput, error)
	MockCreateRuleGroupWithContext                       func(input *svcsdk.CreateRuleGroupInput) (*svcsdk.CreateRuleGroupOutput, error)
	MockCreateRuleGroupRequest                           func(input *svcsdk.CreateRuleGroupInput) (*request.Request, *svcsdk.CreateRuleGroupOutput)
	MockCreateWebACL                                     func(input *svcsdk.CreateWebACLInput) (*svcsdk.CreateWebACLOutput, error)
	MockCreateWebACLWithContext                          func(input *svcsdk.CreateWebACLInput) (*svcsdk.CreateWebACLOutput, error)
	MockCreateWebACLRequest                              func(input *svcsdk.CreateWebACLInput) (*request.Request, *svcsdk.CreateWebACLOutput)
	MockDeleteFirewallManagerRuleGroups                  func(input *svcsdk.DeleteFirewallManagerRuleGroupsInput) (*svcsdk.DeleteFirewallManagerRuleGroupsOutput, error)
	MockDeleteFirewallManagerRuleGroupsRequest           func(input *svcsdk.DeleteFirewallManagerRuleGroupsInput) (*request.Request, *svcsdk.DeleteFirewallManagerRuleGroupsOutput)
	MockDeleteFirewallManagerRuleGroupsWithContext       func(input *svcsdk.DeleteFirewallManagerRuleGroupsInput) (*svcsdk.DeleteFirewallManagerRuleGroupsOutput, error)
	MockDeleteIPSet                                      func(input *svcsdk.DeleteIPSetInput) (*svcsdk.DeleteIPSetOutput, error)
	MockDeleteIPSetWithContext                           func(input *svcsdk.DeleteIPSetInput) (*svcsdk.DeleteIPSetOutput, error)
	MockDeleteIPSetRequest                               func(input *svcsdk.DeleteIPSetInput) (*request.Request, *svcsdk.DeleteIPSetOutput)
	MockDeleteLoggingConfiguration                       func(input *svcsdk.DeleteLoggingConfigurationInput) (*svcsdk.DeleteLoggingConfigurationOutput, error)
	MockDeleteLoggingConfigurationWithContext            func(input *svcsdk.DeleteLoggingConfigurationInput) (*svcsdk.DeleteLoggingConfigurationOutput, error)
	MockDeleteLoggingConfigurationRequest                func(input *svcsdk.DeleteLoggingConfigurationInput) (*request.Request, *svcsdk.DeleteLoggingConfigurationOutput)
	MockDeletePermissionPolicy                           func(input *svcsdk.DeletePermissionPolicyInput) (*svcsdk.DeletePermissionPolicyOutput, error)
	MockDeletePermissionPolicyWithContext                func(input *svcsdk.DeletePermissionPolicyInput) (*svcsdk.DeletePermissionPolicyOutput, error)
	MockDeletePermissionPolicyRequest                    func(input *svcsdk.DeletePermissionPolicyInput) (*request.Request, *svcsdk.DeletePermissionPolicyOutput)
	MockDeleteRegexPatternSet                            func(input *svcsdk.DeleteRegexPatternSetInput) (*svcsdk.DeleteRegexPatternSetOutput, error)
	MockDeleteRegexPatternSetWithContext                 func(input *svcsdk.DeleteRegexPatternSetInput) (*svcsdk.DeleteRegexPatternSetOutput, error)
	MockDeleteRegexPatternSetRequest                     func(input *svcsdk.DeleteRegexPatternSetInput) (*request.Request, *svcsdk.DeleteRegexPatternSetOutput)
	MockDeleteRuleGroup                                  func(input *svcsdk.DeleteRuleGroupInput) (*svcsdk.DeleteRuleGroupOutput, error)
	MockDeleteRuleGroupWithContext                       func(input *svcsdk.DeleteRuleGroupInput) (*svcsdk.DeleteRuleGroupOutput, error)
	MockDeleteRuleGroupRequest                           func(input *svcsdk.DeleteRuleGroupInput) (*request.Request, *svcsdk.DeleteRuleGroupOutput)
	MockDeleteWebACL                                     func(input *svcsdk.DeleteWebACLInput) (*svcsdk.DeleteWebACLOutput, error)
	MockDeleteWebACLWithContext                          func(input *svcsdk.DeleteWebACLInput) (*svcsdk.DeleteWebACLOutput, error)
	MockDeleteWebACLRequest                              func(input *svcsdk.DeleteWebACLInput) (*request.Request, *svcsdk.DeleteWebACLOutput)
	MockDescribeAllManagedProducts                       func(input *svcsdk.DescribeAllManagedProductsInput) (*svcsdk.DescribeAllManagedProductsOutput, error)
	MockDescribeAllManagedProductsWithContext            func(input *svcsdk.DescribeAllManagedProductsInput) (*svcsdk.DescribeAllManagedProductsOutput, error)
	MockDescribeAllManagedProductsRequest                func(input *svcsdk.DescribeAllManagedProductsInput) (*request.Request, *svcsdk.DescribeAllManagedProductsOutput)
	MockDescribeManagedProductsByVendor                  func(input *svcsdk.DescribeManagedProductsByVendorInput) (*svcsdk.DescribeManagedProductsByVendorOutput, error)
	MockDescribeManagedProductsByVendorWithContext       func(input *svcsdk.DescribeManagedProductsByVendorInput) (*svcsdk.DescribeManagedProductsByVendorOutput, error)
	MockDescribeManagedProductsByVendorRequest           func(input *svcsdk.DescribeManagedProductsByVendorInput) (*request.Request, *svcsdk.DescribeManagedProductsByVendorOutput)
	MockDescribeManagedRuleGroup                         func(input *svcsdk.DescribeManagedRuleGroupInput) (*svcsdk.DescribeManagedRuleGroupOutput, error)
	MockDescribeManagedRuleGroupWithContext              func(input *svcsdk.DescribeManagedRuleGroupInput) (*svcsdk.DescribeManagedRuleGroupOutput, error)
	MockDescribeManagedRuleGroupRequest                  func(input *svcsdk.DescribeManagedRuleGroupInput) (*request.Request, *svcsdk.DescribeManagedRuleGroupOutput)
	MockDisassociateWebACL                               func(input *svcsdk.DisassociateWebACLInput) (*svcsdk.DisassociateWebACLOutput, error)
	MockDisassociateWebACLWithContext                    func(input *svcsdk.DisassociateWebACLInput) (*svcsdk.DisassociateWebACLOutput, error)
	MockDisassociateWebACLRequest                        func(input *svcsdk.DisassociateWebACLInput) (*request.Request, *svcsdk.DisassociateWebACLOutput)
	MockGenerateMobileSdkReleaseUrl                      func(input *svcsdk.GenerateMobileSdkReleaseUrlInput) (*svcsdk.GenerateMobileSdkReleaseUrlOutput, error)
	MockGenerateMobileSdkReleaseUrlWithContext           func(input *svcsdk.GenerateMobileSdkReleaseUrlInput) (*svcsdk.GenerateMobileSdkReleaseUrlOutput, error)
	MockGenerateMobileSdkReleaseUrlRequest               func(input *svcsdk.GenerateMobileSdkReleaseUrlInput) (*request.Request, *svcsdk.GenerateMobileSdkReleaseUrlOutput)
	MockGetDecryptedAPIKey                               func(input *svcsdk.GetDecryptedAPIKeyInput) (*svcsdk.GetDecryptedAPIKeyOutput, error)
	MockGetDecryptedAPIKeyWithContext                    func(input *svcsdk.GetDecryptedAPIKeyInput) (*svcsdk.GetDecryptedAPIKeyOutput, error)
	MockGetDecryptedAPIKeyRequest                        func(input *svcsdk.GetDecryptedAPIKeyInput) (*request.Request, *svcsdk.GetDecryptedAPIKeyOutput)
	MockGetManagedRuleSet                                func(input *svcsdk.GetManagedRuleSetInput) (*svcsdk.GetManagedRuleSetOutput, error)
	MockGetManagedRuleSetWithContext                     func(input *svcsdk.GetManagedRuleSetInput) (*svcsdk.GetManagedRuleSetOutput, error)
	MockGetManagedRuleSetRequest                         func(input *svcsdk.GetManagedRuleSetInput) (*request.Request, *svcsdk.GetManagedRuleSetOutput)
	MockGetIPSet                                         func(input *svcsdk.GetIPSetInput) (*svcsdk.GetIPSetOutput, error)
	MockGetIPSetWithContext                              func(input *svcsdk.GetIPSetInput) (*svcsdk.GetIPSetOutput, error)
	MockGetIPSetRequest                                  func(input *svcsdk.GetIPSetInput) (*request.Request, *svcsdk.GetIPSetOutput)
	MockGetLoggingConfiguration                          func(input *svcsdk.GetLoggingConfigurationInput) (*svcsdk.GetLoggingConfigurationOutput, error)
	MockGetLoggingConfigurationWithContext               func(input *svcsdk.GetLoggingConfigurationInput) (*svcsdk.GetLoggingConfigurationOutput, error)
	MockGetLoggingConfigurationRequest                   func(input *svcsdk.GetLoggingConfigurationInput) (*request.Request, *svcsdk.GetLoggingConfigurationOutput)
	MockGetPermissionPolicy                              func(input *svcsdk.GetPermissionPolicyInput) (*svcsdk.GetPermissionPolicyOutput, error)
	MockGetPermissionPolicyWithContext                   func(input *svcsdk.GetPermissionPolicyInput) (*svcsdk.GetPermissionPolicyOutput, error)
	MockGetPermissionPolicyRequest                       func(input *svcsdk.GetPermissionPolicyInput) (*request.Request, *svcsdk.GetPermissionPolicyOutput)
	MockGetRateBasedStatementManagedKeys                 func(input *svcsdk.GetRateBasedStatementManagedKeysInput) (*svcsdk.GetRateBasedStatementManagedKeysOutput, error)
	MockGetRateBasedStatementManagedKeysWithContext      func(input *svcsdk.GetRateBasedStatementManagedKeysInput) (*svcsdk.GetRateBasedStatementManagedKeysOutput, error)
	MockGetRateBasedStatementManagedKeysRequest          func(input *svcsdk.GetRateBasedStatementManagedKeysInput) (*request.Request, *svcsdk.GetRateBasedStatementManagedKeysOutput)
	MockGetRegexPatternSet                               func(input *svcsdk.GetRegexPatternSetInput) (*svcsdk.GetRegexPatternSetOutput, error)
	MockGetRegexPatternSetWithContext                    func(input *svcsdk.GetRegexPatternSetInput) (*svcsdk.GetRegexPatternSetOutput, error)
	MockGetRegexPatternSetRequest                        func(input *svcsdk.GetRegexPatternSetInput) (*request.Request, *svcsdk.GetRegexPatternSetOutput)
	MockGetRuleGroup                                     func(input *svcsdk.GetRuleGroupInput) (*svcsdk.GetRuleGroupOutput, error)
	MockGetRuleGroupWithContext                          func(input *svcsdk.GetRuleGroupInput) (*svcsdk.GetRuleGroupOutput, error)
	MockGetRuleGroupRequest                              func(input *svcsdk.GetRuleGroupInput) (*request.Request, *svcsdk.GetRuleGroupOutput)
	MockGetSampledRequests                               func(input *svcsdk.GetSampledRequestsInput) (*svcsdk.GetSampledRequestsOutput, error)
	MockGetSampledRequestsWithContext                    func(input *svcsdk.GetSampledRequestsInput) (*svcsdk.GetSampledRequestsOutput, error)
	MockGetSampledRequestsRequest                        func(input *svcsdk.GetSampledRequestsInput) (*request.Request, *svcsdk.GetSampledRequestsOutput)
	MockGetMobileSdkRelease                              func(input *svcsdk.GetMobileSdkReleaseInput) (*svcsdk.GetMobileSdkReleaseOutput, error)
	MockGetMobileSdkReleaseWithContext                   func(input *svcsdk.GetMobileSdkReleaseInput) (*svcsdk.GetMobileSdkReleaseOutput, error)
	MockGetMobileSdkReleaseRequest                       func(input *svcsdk.GetMobileSdkReleaseInput) (*request.Request, *svcsdk.GetMobileSdkReleaseOutput)
	MockGetWebACL                                        func(input *svcsdk.GetWebACLInput) (*svcsdk.GetWebACLOutput, error)
	MockGetWebACLWithContext                             func(input *svcsdk.GetWebACLInput) (*svcsdk.GetWebACLOutput, error)
	MockGetWebACLRequest                                 func(input *svcsdk.GetWebACLInput) (*request.Request, *svcsdk.GetWebACLOutput)
	MockGetWebACLForResource                             func(input *svcsdk.GetWebACLForResourceInput) (*svcsdk.GetWebACLForResourceOutput, error)
	MockGetWebACLForResourceWithContext                  func(input *svcsdk.GetWebACLForResourceInput) (*svcsdk.GetWebACLForResourceOutput, error)
	MockGetWebACLForResourceRequest                      func(input *svcsdk.GetWebACLForResourceInput) (*request.Request, *svcsdk.GetWebACLForResourceOutput)
	MockListAPIKeys                                      func(input *svcsdk.ListAPIKeysInput) (*svcsdk.ListAPIKeysOutput, error)
	MockListAPIKeysWithContext                           func(input *svcsdk.ListAPIKeysInput) (*svcsdk.ListAPIKeysOutput, error)
	MockListAPIKeysRequest                               func(input *svcsdk.ListAPIKeysInput) (*request.Request, *svcsdk.ListAPIKeysOutput)
	MockListAvailableManagedRuleGroupVersions            func(input *svcsdk.ListAvailableManagedRuleGroupVersionsInput) (*svcsdk.ListAvailableManagedRuleGroupVersionsOutput, error)
	MockListAvailableManagedRuleGroupVersionsWithContext func(input *svcsdk.ListAvailableManagedRuleGroupVersionsInput) (*svcsdk.ListAvailableManagedRuleGroupVersionsOutput, error)
	MockListAvailableManagedRuleGroupVersionsRequest     func(input *svcsdk.ListAvailableManagedRuleGroupVersionsInput) (*request.Request, *svcsdk.ListAvailableManagedRuleGroupVersionsOutput)
	MockListAvailableManagedRuleGroups                   func(input *svcsdk.ListAvailableManagedRuleGroupsInput) (*svcsdk.ListAvailableManagedRuleGroupsOutput, error)
	MockListAvailableManagedRuleGroupsWithContext        func(input *svcsdk.ListAvailableManagedRuleGroupsInput) (*svcsdk.ListAvailableManagedRuleGroupsOutput, error)
	MockListAvailableManagedRuleGroupsRequest            func(input *svcsdk.ListAvailableManagedRuleGroupsInput) (*request.Request, *svcsdk.ListAvailableManagedRuleGroupsOutput)
	MockListIPSets                                       func(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error)
	MockListIPSetsWithContext                            func(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error)
	MockListIPSetsRequest                                func(input *svcsdk.ListIPSetsInput) (*request.Request, *svcsdk.ListIPSetsOutput)
	MockListLoggingConfigurations                        func(input *svcsdk.ListLoggingConfigurationsInput) (*svcsdk.ListLoggingConfigurationsOutput, error)
	MockListLoggingConfigurationsWithContext             func(input *svcsdk.ListLoggingConfigurationsInput) (*svcsdk.ListLoggingConfigurationsOutput, error)
	MockListLoggingConfigurationsRequest                 func(input *svcsdk.ListLoggingConfigurationsInput) (*request.Request, *svcsdk.ListLoggingConfigurationsOutput)
	MockListManagedRuleSets                              func(input *svcsdk.ListManagedRuleSetsInput) (*svcsdk.ListManagedRuleSetsOutput, error)
	MockListManagedRuleSetsWithContext                   func(input *svcsdk.ListManagedRuleSetsInput) (*svcsdk.ListManagedRuleSetsOutput, error)
	MockListManagedRuleSetsRequest                       func(input *svcsdk.ListManagedRuleSetsInput) (*request.Request, *svcsdk.ListManagedRuleSetsOutput)
	MockListMobileSdkReleases                            func(input *svcsdk.ListMobileSdkReleasesInput) (*svcsdk.ListMobileSdkReleasesOutput, error)
	MockListMobileSdkReleasesWithContext                 func(input *svcsdk.ListMobileSdkReleasesInput) (*svcsdk.ListMobileSdkReleasesOutput, error)
	MockListMobileSdkReleasesRequest                     func(input *svcsdk.ListMobileSdkReleasesInput) (*request.Request, *svcsdk.ListMobileSdkReleasesOutput)
	MockListRegexPatternSets                             func(input *svcsdk.ListRegexPatternSetsInput) (*svcsdk.ListRegexPatternSetsOutput, error)
	MockListRegexPatternSetsWithContext                  func(input *svcsdk.ListRegexPatternSetsInput) (*svcsdk.ListRegexPatternSetsOutput, error)
	MockListRegexPatternSetsRequest                      func(input *svcsdk.ListRegexPatternSetsInput) (*request.Request, *svcsdk.ListRegexPatternSetsOutput)
	MockListResourcesForWebACL                           func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error)
	MockListResourcesForWebACLWithContext                func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error)
	MockListResourcesForWebACLRequest                    func(input *svcsdk.ListResourcesForWebACLInput) (*request.Request, *svcsdk.ListResourcesForWebACLOutput)
	MockListRuleGroups                                   func(input *svcsdk.ListRuleGroupsInput) (*svcsdk.ListRuleGroupsOutput, error)
	MockListRuleGroupsWithContext                        func(input *svcsdk.ListRuleGroupsInput) (*svcsdk.ListRuleGroupsOutput, error)
	MockListRuleGroupsRequest                            func(input *svcsdk.ListRuleGroupsInput) (*request.Request, *svcsdk.ListRuleGroupsOutput)
	MockListTagsForResource                              func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error)
	MockListTagsForResourceWithContext                   func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error)
	MockListTagsForResourceRequest                       func(input *svcsdk.ListTagsForResourceInput) (*request.Request, *svcsdk.ListTagsForResourceOutput)
	MockListWebACLs                                      func(input *svcsdk.ListWebACLsInput) (*svcsdk.ListWebACLsOutput, error)
	MockListWebACLsWithContext                           func(input *svcsdk.ListWebACLsInput) (*svcsdk.ListWebACLsOutput, error)
	MockListWebACLsRequest                               func(input *svcsdk.ListWebACLsInput) (*request.Request, *svcsdk.ListWebACLsOutput)
	MockPutPermissionPolicy                              func(input *svcsdk.PutPermissionPolicyInput) (*svcsdk.PutPermissionPolicyOutput, error)
	MockPutPermissionPolicyWithContext                   func(input *svcsdk.PutPermissionPolicyInput) (*svcsdk.PutPermissionPolicyOutput, error)
	MockPutPermissionPolicyRequest                       func(input *svcsdk.PutPermissionPolicyInput) (*request.Request, *svcsdk.PutPermissionPolicyOutput)
	MockPutLoggingConfiguration                          func(input *svcsdk.PutLoggingConfigurationInput) (*svcsdk.PutLoggingConfigurationOutput, error)
	MockPutLoggingConfigurationWithContext               func(input *svcsdk.PutLoggingConfigurationInput) (*svcsdk.PutLoggingConfigurationOutput, error)
	MockPutLoggingConfigurationRequest                   func(input *svcsdk.PutLoggingConfigurationInput) (*request.Request, *svcsdk.PutLoggingConfigurationOutput)
	MockPutManagedRuleSetVersions                        func(input *svcsdk.PutManagedRuleSetVersionsInput) (*svcsdk.PutManagedRuleSetVersionsOutput, error)
	MockPutManagedRuleSetVersionsWithContext             func(input *svcsdk.PutManagedRuleSetVersionsInput) (*svcsdk.PutManagedRuleSetVersionsOutput, error)
	MockPutManagedRuleSetVersionsRequest                 func(input *svcsdk.PutManagedRuleSetVersionsInput) (*request.Request, *svcsdk.PutManagedRuleSetVersionsOutput)
	MockTagResource                                      func(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error)
	MockTagResourceWithContext                           func(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error)
	MockTagResourceRequest                               func(input *svcsdk.TagResourceInput) (*request.Request, *svcsdk.TagResourceOutput)
	MockUntagResource                                    func(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error)
	MockUntagResourceWithContext                         func(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error)
	MockUntagResourceRequest                             func(input *svcsdk.UntagResourceInput) (*request.Request, *svcsdk.UntagResourceOutput)
	MockUpdateIPSet                                      func(input *svcsdk.UpdateIPSetInput) (*svcsdk.UpdateIPSetOutput, error)
	MockUpdateIPSetWithContext                           func(input *svcsdk.UpdateIPSetInput) (*svcsdk.UpdateIPSetOutput, error)
	MockUpdateIPSetRequest                               func(input *svcsdk.UpdateIPSetInput) (*request.Request, *svcsdk.UpdateIPSetOutput)
	MockUpdateRegexPatternSet                            func(input *svcsdk.UpdateRegexPatternSetInput) (*svcsdk.UpdateRegexPatternSetOutput, error)
	MockUpdateRegexPatternSetWithContext                 func(input *svcsdk.UpdateRegexPatternSetInput) (*svcsdk.UpdateRegexPatternSetOutput, error)
	MockUpdateRegexPatternSetRequest                     func(input *svcsdk.UpdateRegexPatternSetInput) (*request.Request, *svcsdk.UpdateRegexPatternSetOutput)
	MockUpdateManagedRuleSetVersionExpiryDateWithContext func(input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput) (*svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput, error)
	MockUpdateManagedRuleSetVersionExpiryDateRequest     func(input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput) (*request.Request, *svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput)
	MockUpdateManagedRuleSetVersionExpiryDate            func(input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput) (*svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput, error)
	MockUpdateRuleGroup                                  func(input *svcsdk.UpdateRuleGroupInput) (*svcsdk.UpdateRuleGroupOutput, error)
	MockUpdateRuleGroupWithContext                       func(input *svcsdk.UpdateRuleGroupInput) (*svcsdk.UpdateRuleGroupOutput, error)
	MockUpdateRuleGroupRequest                           func(input *svcsdk.UpdateRuleGroupInput) (*request.Request, *svcsdk.UpdateRuleGroupOutput)
	MockUpdateWebACL                                     func(input *svcsdk.UpdateWebACLInput) (*svcsdk.UpdateWebACLOutput, error)
	MockUpdateWebACLWithContext                          func(input *svcsdk.UpdateWebACLInput) (*svcsdk.UpdateWebACLOutput, error)
	MockUpdateWebACLRequest                              func(input *svcsdk.UpdateWebACLInput) (*request.Request, *svcsdk.UpdateWebACLOutput)
}

func (m *MockWAFV2Client) AssociateWebACL(input *svcsdk.AssociateWebACLInput) (*svcsdk.AssociateWebACLOutput, error) {
	return m.MockAssociateWebACL(input)
}

func (m *MockWAFV2Client) AssociateWebACLWithContext(context aws.Context, input *svcsdk.AssociateWebACLInput, option ...request.Option) (*svcsdk.AssociateWebACLOutput, error) {
	return m.MockAssociateWebACLWithContext(input)
}

func (m *MockWAFV2Client) AssociateWebACLRequest(input *svcsdk.AssociateWebACLInput) (*request.Request, *svcsdk.AssociateWebACLOutput) {
	return m.MockAssociateWebACLRequest(input)
}

func (m *MockWAFV2Client) CheckCapacity(input *svcsdk.CheckCapacityInput) (*svcsdk.CheckCapacityOutput, error) {
	return m.MockCheckCapacity(input)
}

func (m *MockWAFV2Client) CheckCapacityWithContext(context aws.Context, input *svcsdk.CheckCapacityInput, option ...request.Option) (*svcsdk.CheckCapacityOutput, error) {
	return m.MockCheckCapacityWithContext(input)
}

func (m *MockWAFV2Client) CheckCapacityRequest(input *svcsdk.CheckCapacityInput) (*request.Request, *svcsdk.CheckCapacityOutput) {
	return m.MockCapacityRequest(input)
}

func (m *MockWAFV2Client) CreateAPIKey(input *svcsdk.CreateAPIKeyInput) (*svcsdk.CreateAPIKeyOutput, error) {
	return m.MockCreateAPIKey(input)
}

func (m *MockWAFV2Client) CreateAPIKeyWithContext(context aws.Context, input *svcsdk.CreateAPIKeyInput, option ...request.Option) (*svcsdk.CreateAPIKeyOutput, error) {
	return m.MockCreateAPIKeyWithContext(input)
}

func (m *MockWAFV2Client) CreateAPIKeyRequest(input *svcsdk.CreateAPIKeyInput) (*request.Request, *svcsdk.CreateAPIKeyOutput) {
	return m.MockCreateAPIKeyRequest(input)
}

func (m *MockWAFV2Client) DeleteAPIKey(input *svcsdk.DeleteAPIKeyInput) (*svcsdk.DeleteAPIKeyOutput, error) {
	return m.MockDeleteAPIKey(input)
}

func (m *MockWAFV2Client) DeleteAPIKeyWithContext(context aws.Context, input *svcsdk.DeleteAPIKeyInput, option ...request.Option) (*svcsdk.DeleteAPIKeyOutput, error) {
	return m.MockDeleteAPIKeyWithContext(input)
}

func (m *MockWAFV2Client) DeleteAPIKeyRequest(input *svcsdk.DeleteAPIKeyInput) (*request.Request, *svcsdk.DeleteAPIKeyOutput) {
	return m.MockDeleteAPIKeyRequest(input)
}

func (m *MockWAFV2Client) CreateIPSet(input *svcsdk.CreateIPSetInput) (*svcsdk.CreateIPSetOutput, error) {
	return m.MockCreateIPSet(input)
}

func (m *MockWAFV2Client) CreateIPSetWithContext(context aws.Context, input *svcsdk.CreateIPSetInput, option ...request.Option) (*svcsdk.CreateIPSetOutput, error) {
	return m.MockCreateIPSetWithContext(input)
}

func (m *MockWAFV2Client) CreateIPSetRequest(input *svcsdk.CreateIPSetInput) (*request.Request, *svcsdk.CreateIPSetOutput) {
	return m.MockCreateIPSetRequest(input)
}

func (m *MockWAFV2Client) CreateRegexPatternSet(input *svcsdk.CreateRegexPatternSetInput) (*svcsdk.CreateRegexPatternSetOutput, error) {
	return m.MockCreateRegexPatternSet(input)
}

func (m *MockWAFV2Client) CreateRegexPatternSetWithContext(context aws.Context, input *svcsdk.CreateRegexPatternSetInput, option ...request.Option) (*svcsdk.CreateRegexPatternSetOutput, error) {
	return m.MockCreateRegexPatternSetWithContext(input)
}

func (m *MockWAFV2Client) CreateRegexPatternSetRequest(input *svcsdk.CreateRegexPatternSetInput) (*request.Request, *svcsdk.CreateRegexPatternSetOutput) {
	return m.MockCreateRegexPatternSetRequest(input)
}

func (m *MockWAFV2Client) CreateRuleGroup(input *svcsdk.CreateRuleGroupInput) (*svcsdk.CreateRuleGroupOutput, error) {
	return m.MockCreateRuleGroup(input)
}

func (m *MockWAFV2Client) CreateRuleGroupWithContext(context aws.Context, input *svcsdk.CreateRuleGroupInput, option ...request.Option) (*svcsdk.CreateRuleGroupOutput, error) {
	return m.MockCreateRuleGroupWithContext(input)
}

func (m *MockWAFV2Client) CreateRuleGroupRequest(input *svcsdk.CreateRuleGroupInput) (*request.Request, *svcsdk.CreateRuleGroupOutput) {
	return m.MockCreateRuleGroupRequest(input)
}

func (m *MockWAFV2Client) CreateWebACL(input *svcsdk.CreateWebACLInput) (*svcsdk.CreateWebACLOutput, error) {
	return m.MockCreateWebACL(input)
}

func (m *MockWAFV2Client) CreateWebACLWithContext(context aws.Context, input *svcsdk.CreateWebACLInput, option ...request.Option) (*svcsdk.CreateWebACLOutput, error) {
	return m.MockCreateWebACLWithContext(input)
}

func (m *MockWAFV2Client) CreateWebACLRequest(input *svcsdk.CreateWebACLInput) (*request.Request, *svcsdk.CreateWebACLOutput) {
	return m.MockCreateWebACLRequest(input)
}

func (m *MockWAFV2Client) DeleteFirewallManagerRuleGroups(input *svcsdk.DeleteFirewallManagerRuleGroupsInput) (*svcsdk.DeleteFirewallManagerRuleGroupsOutput, error) {
	return m.MockDeleteFirewallManagerRuleGroups(input)
}

func (m *MockWAFV2Client) DeleteFirewallManagerRuleGroupsWithContext(context aws.Context, input *svcsdk.DeleteFirewallManagerRuleGroupsInput, option ...request.Option) (*svcsdk.DeleteFirewallManagerRuleGroupsOutput, error) {
	return m.MockDeleteFirewallManagerRuleGroupsWithContext(input)
}

func (m *MockWAFV2Client) DeleteFirewallManagerRuleGroupsRequest(input *svcsdk.DeleteFirewallManagerRuleGroupsInput) (*request.Request, *svcsdk.DeleteFirewallManagerRuleGroupsOutput) {
	return m.MockDeleteFirewallManagerRuleGroupsRequest(input)
}

func (m *MockWAFV2Client) DeleteIPSet(input *svcsdk.DeleteIPSetInput) (*svcsdk.DeleteIPSetOutput, error) {
	return m.MockDeleteIPSet(input)
}

func (m *MockWAFV2Client) DeleteIPSetWithContext(context aws.Context, input *svcsdk.DeleteIPSetInput, option ...request.Option) (*svcsdk.DeleteIPSetOutput, error) {
	return m.MockDeleteIPSetWithContext(input)
}

func (m *MockWAFV2Client) DeleteIPSetRequest(input *svcsdk.DeleteIPSetInput) (*request.Request, *svcsdk.DeleteIPSetOutput) {
	return m.MockDeleteIPSetRequest(input)
}

func (m *MockWAFV2Client) DeleteLoggingConfiguration(input *svcsdk.DeleteLoggingConfigurationInput) (*svcsdk.DeleteLoggingConfigurationOutput, error) {
	return m.MockDeleteLoggingConfiguration(input)
}

func (m *MockWAFV2Client) DeleteLoggingConfigurationWithContext(context aws.Context, input *svcsdk.DeleteLoggingConfigurationInput, option ...request.Option) (*svcsdk.DeleteLoggingConfigurationOutput, error) {
	return m.MockDeleteLoggingConfigurationWithContext(input)
}

func (m *MockWAFV2Client) DeleteLoggingConfigurationRequest(input *svcsdk.DeleteLoggingConfigurationInput) (*request.Request, *svcsdk.DeleteLoggingConfigurationOutput) {
	return m.MockDeleteLoggingConfigurationRequest(input)
}

func (m *MockWAFV2Client) DeletePermissionPolicy(input *svcsdk.DeletePermissionPolicyInput) (*svcsdk.DeletePermissionPolicyOutput, error) {
	return m.MockDeletePermissionPolicy(input)
}

func (m *MockWAFV2Client) DeletePermissionPolicyWithContext(context aws.Context, input *svcsdk.DeletePermissionPolicyInput, option ...request.Option) (*svcsdk.DeletePermissionPolicyOutput, error) {
	return m.MockDeletePermissionPolicyWithContext(input)
}

func (m *MockWAFV2Client) DeletePermissionPolicyRequest(input *svcsdk.DeletePermissionPolicyInput) (*request.Request, *svcsdk.DeletePermissionPolicyOutput) {
	return m.MockDeletePermissionPolicyRequest(input)
}

func (m *MockWAFV2Client) DeleteRegexPatternSet(input *svcsdk.DeleteRegexPatternSetInput) (*svcsdk.DeleteRegexPatternSetOutput, error) {
	return m.MockDeleteRegexPatternSet(input)
}

func (m *MockWAFV2Client) DeleteRegexPatternSetWithContext(context aws.Context, input *svcsdk.DeleteRegexPatternSetInput, option ...request.Option) (*svcsdk.DeleteRegexPatternSetOutput, error) {
	return m.MockDeleteRegexPatternSetWithContext(input)
}

func (m *MockWAFV2Client) DeleteRegexPatternSetRequest(input *svcsdk.DeleteRegexPatternSetInput) (*request.Request, *svcsdk.DeleteRegexPatternSetOutput) {
	return m.MockDeleteRegexPatternSetRequest(input)
}

func (m *MockWAFV2Client) DeleteRuleGroup(input *svcsdk.DeleteRuleGroupInput) (*svcsdk.DeleteRuleGroupOutput, error) {
	return m.MockDeleteRuleGroup(input)
}

func (m *MockWAFV2Client) DeleteRuleGroupWithContext(context aws.Context, input *svcsdk.DeleteRuleGroupInput, option ...request.Option) (*svcsdk.DeleteRuleGroupOutput, error) {
	return m.MockDeleteRuleGroupWithContext(input)
}

func (m *MockWAFV2Client) DeleteRuleGroupRequest(input *svcsdk.DeleteRuleGroupInput) (*request.Request, *svcsdk.DeleteRuleGroupOutput) {
	return m.MockDeleteRuleGroupRequest(input)
}

func (m *MockWAFV2Client) DeleteWebACL(input *svcsdk.DeleteWebACLInput) (*svcsdk.DeleteWebACLOutput, error) {
	return m.MockDeleteWebACL(input)
}

func (m *MockWAFV2Client) DeleteWebACLWithContext(context aws.Context, input *svcsdk.DeleteWebACLInput, option ...request.Option) (*svcsdk.DeleteWebACLOutput, error) {
	return m.MockDeleteWebACLWithContext(input)
}

func (m *MockWAFV2Client) DeleteWebACLRequest(input *svcsdk.DeleteWebACLInput) (*request.Request, *svcsdk.DeleteWebACLOutput) {
	return m.MockDeleteWebACLRequest(input)
}

func (m *MockWAFV2Client) DescribeAllManagedProducts(input *svcsdk.DescribeAllManagedProductsInput) (*svcsdk.DescribeAllManagedProductsOutput, error) {
	return m.MockDescribeAllManagedProducts(input)
}

func (m *MockWAFV2Client) DescribeAllManagedProductsWithContext(context aws.Context, input *svcsdk.DescribeAllManagedProductsInput, option ...request.Option) (*svcsdk.DescribeAllManagedProductsOutput, error) {
	return m.MockDescribeAllManagedProductsWithContext(input)
}

func (m *MockWAFV2Client) DescribeAllManagedProductsRequest(input *svcsdk.DescribeAllManagedProductsInput) (*request.Request, *svcsdk.DescribeAllManagedProductsOutput) {
	return m.MockDescribeAllManagedProductsRequest(input)
}

func (m *MockWAFV2Client) DescribeManagedProductsByVendor(input *svcsdk.DescribeManagedProductsByVendorInput) (*svcsdk.DescribeManagedProductsByVendorOutput, error) {
	return m.MockDescribeManagedProductsByVendor(input)
}

func (m *MockWAFV2Client) DescribeManagedProductsByVendorWithContext(context aws.Context, input *svcsdk.DescribeManagedProductsByVendorInput, option ...request.Option) (*svcsdk.DescribeManagedProductsByVendorOutput, error) {
	return m.MockDescribeManagedProductsByVendorWithContext(input)
}

func (m *MockWAFV2Client) DescribeManagedProductsByVendorRequest(input *svcsdk.DescribeManagedProductsByVendorInput) (*request.Request, *svcsdk.DescribeManagedProductsByVendorOutput) {
	return m.MockDescribeManagedProductsByVendorRequest(input)
}

func (m *MockWAFV2Client) DescribeManagedRuleGroup(input *svcsdk.DescribeManagedRuleGroupInput) (*svcsdk.DescribeManagedRuleGroupOutput, error) {
	return m.MockDescribeManagedRuleGroup(input)
}

func (m *MockWAFV2Client) DescribeManagedRuleGroupWithContext(context aws.Context, input *svcsdk.DescribeManagedRuleGroupInput, option ...request.Option) (*svcsdk.DescribeManagedRuleGroupOutput, error) {
	return m.MockDescribeManagedRuleGroupWithContext(input)
}

func (m *MockWAFV2Client) DescribeManagedRuleGroupRequest(input *svcsdk.DescribeManagedRuleGroupInput) (*request.Request, *svcsdk.DescribeManagedRuleGroupOutput) {
	return m.MockDescribeManagedRuleGroupRequest(input)
}

func (m *MockWAFV2Client) DisassociateWebACL(input *svcsdk.DisassociateWebACLInput) (*svcsdk.DisassociateWebACLOutput, error) {
	return m.MockDisassociateWebACL(input)
}

func (m *MockWAFV2Client) DisassociateWebACLWithContext(context aws.Context, input *svcsdk.DisassociateWebACLInput, option ...request.Option) (*svcsdk.DisassociateWebACLOutput, error) {
	return m.MockDisassociateWebACLWithContext(input)
}

func (m *MockWAFV2Client) DisassociateWebACLRequest(input *svcsdk.DisassociateWebACLInput) (*request.Request, *svcsdk.DisassociateWebACLOutput) {
	return m.MockDisassociateWebACLRequest(input)
}

func (m *MockWAFV2Client) GenerateMobileSdkReleaseUrl(input *svcsdk.GenerateMobileSdkReleaseUrlInput) (*svcsdk.GenerateMobileSdkReleaseUrlOutput, error) {
	return m.MockGenerateMobileSdkReleaseUrl(input)
}

func (m *MockWAFV2Client) GenerateMobileSdkReleaseUrlWithContext(context aws.Context, input *svcsdk.GenerateMobileSdkReleaseUrlInput, option ...request.Option) (*svcsdk.GenerateMobileSdkReleaseUrlOutput, error) {
	return m.MockGenerateMobileSdkReleaseUrlWithContext(input)
}

func (m *MockWAFV2Client) GenerateMobileSdkReleaseUrlRequest(input *svcsdk.GenerateMobileSdkReleaseUrlInput) (*request.Request, *svcsdk.GenerateMobileSdkReleaseUrlOutput) {
	return m.MockGenerateMobileSdkReleaseUrlRequest(input)
}

func (m *MockWAFV2Client) GetDecryptedAPIKey(input *svcsdk.GetDecryptedAPIKeyInput) (*svcsdk.GetDecryptedAPIKeyOutput, error) {
	return m.MockGetDecryptedAPIKey(input)
}

func (m *MockWAFV2Client) GetDecryptedAPIKeyWithContext(context aws.Context, input *svcsdk.GetDecryptedAPIKeyInput, option ...request.Option) (*svcsdk.GetDecryptedAPIKeyOutput, error) {
	return m.MockGetDecryptedAPIKeyWithContext(input)
}

func (m *MockWAFV2Client) GetDecryptedAPIKeyRequest(input *svcsdk.GetDecryptedAPIKeyInput) (*request.Request, *svcsdk.GetDecryptedAPIKeyOutput) {
	return m.MockGetDecryptedAPIKeyRequest(input)
}

func (m *MockWAFV2Client) GetIPSet(input *svcsdk.GetIPSetInput) (*svcsdk.GetIPSetOutput, error) {
	return m.MockGetIPSet(input)
}

func (m *MockWAFV2Client) GetIPSetWithContext(context aws.Context, input *svcsdk.GetIPSetInput, option ...request.Option) (*svcsdk.GetIPSetOutput, error) {
	return m.MockGetIPSetWithContext(input)
}

func (m *MockWAFV2Client) GetIPSetRequest(input *svcsdk.GetIPSetInput) (*request.Request, *svcsdk.GetIPSetOutput) {
	return m.MockGetIPSetRequest(input)
}

func (m *MockWAFV2Client) GetLoggingConfiguration(input *svcsdk.GetLoggingConfigurationInput) (*svcsdk.GetLoggingConfigurationOutput, error) {
	return m.MockGetLoggingConfiguration(input)
}

func (m *MockWAFV2Client) GetLoggingConfigurationWithContext(context aws.Context, input *svcsdk.GetLoggingConfigurationInput, option ...request.Option) (*svcsdk.GetLoggingConfigurationOutput, error) {
	return m.MockGetLoggingConfigurationWithContext(input)
}

func (m *MockWAFV2Client) GetLoggingConfigurationRequest(input *svcsdk.GetLoggingConfigurationInput) (*request.Request, *svcsdk.GetLoggingConfigurationOutput) {
	return m.MockGetLoggingConfigurationRequest(input)
}

func (m *MockWAFV2Client) GetManagedRuleSet(input *svcsdk.GetManagedRuleSetInput) (*svcsdk.GetManagedRuleSetOutput, error) {
	return m.MockGetManagedRuleSet(input)
}

func (m *MockWAFV2Client) GetManagedRuleSetWithContext(context aws.Context, input *svcsdk.GetManagedRuleSetInput, option ...request.Option) (*svcsdk.GetManagedRuleSetOutput, error) {
	return m.MockGetManagedRuleSetWithContext(input)
}

func (m *MockWAFV2Client) GetManagedRuleSetRequest(input *svcsdk.GetManagedRuleSetInput) (*request.Request, *svcsdk.GetManagedRuleSetOutput) {
	return m.MockGetManagedRuleSetRequest(input)
}

func (m *MockWAFV2Client) GetMobileSdkRelease(input *svcsdk.GetMobileSdkReleaseInput) (*svcsdk.GetMobileSdkReleaseOutput, error) {
	return m.MockGetMobileSdkRelease(input)
}

func (m *MockWAFV2Client) GetMobileSdkReleaseWithContext(context aws.Context, input *svcsdk.GetMobileSdkReleaseInput, option ...request.Option) (*svcsdk.GetMobileSdkReleaseOutput, error) {
	return m.MockGetMobileSdkReleaseWithContext(input)
}

func (m *MockWAFV2Client) GetMobileSdkReleaseRequest(input *svcsdk.GetMobileSdkReleaseInput) (*request.Request, *svcsdk.GetMobileSdkReleaseOutput) {
	return m.MockGetMobileSdkReleaseRequest(input)
}

func (m *MockWAFV2Client) GetPermissionPolicy(input *svcsdk.GetPermissionPolicyInput) (*svcsdk.GetPermissionPolicyOutput, error) {
	return m.MockGetPermissionPolicy(input)
}

func (m *MockWAFV2Client) GetPermissionPolicyWithContext(context aws.Context, input *svcsdk.GetPermissionPolicyInput, option ...request.Option) (*svcsdk.GetPermissionPolicyOutput, error) {
	return m.MockGetPermissionPolicyWithContext(input)
}

func (m *MockWAFV2Client) GetPermissionPolicyRequest(input *svcsdk.GetPermissionPolicyInput) (*request.Request, *svcsdk.GetPermissionPolicyOutput) {
	return m.MockGetPermissionPolicyRequest(input)
}

func (m *MockWAFV2Client) GetRateBasedStatementManagedKeys(input *svcsdk.GetRateBasedStatementManagedKeysInput) (*svcsdk.GetRateBasedStatementManagedKeysOutput, error) {
	return m.MockGetRateBasedStatementManagedKeys(input)
}

func (m *MockWAFV2Client) GetRateBasedStatementManagedKeysWithContext(context aws.Context, input *svcsdk.GetRateBasedStatementManagedKeysInput, option ...request.Option) (*svcsdk.GetRateBasedStatementManagedKeysOutput, error) {
	return m.MockGetRateBasedStatementManagedKeysWithContext(input)
}

func (m *MockWAFV2Client) GetRateBasedStatementManagedKeysRequest(input *svcsdk.GetRateBasedStatementManagedKeysInput) (*request.Request, *svcsdk.GetRateBasedStatementManagedKeysOutput) {
	return m.MockGetRateBasedStatementManagedKeysRequest(input)
}

func (m *MockWAFV2Client) GetRegexPatternSet(input *svcsdk.GetRegexPatternSetInput) (*svcsdk.GetRegexPatternSetOutput, error) {
	return m.MockGetRegexPatternSet(input)
}

func (m *MockWAFV2Client) GetRegexPatternSetWithContext(context aws.Context, input *svcsdk.GetRegexPatternSetInput, option ...request.Option) (*svcsdk.GetRegexPatternSetOutput, error) {
	return m.MockGetRegexPatternSetWithContext(input)
}

func (m *MockWAFV2Client) GetRegexPatternSetRequest(input *svcsdk.GetRegexPatternSetInput) (*request.Request, *svcsdk.GetRegexPatternSetOutput) {
	return m.MockGetRegexPatternSetRequest(input)
}

func (m *MockWAFV2Client) GetRuleGroup(input *svcsdk.GetRuleGroupInput) (*svcsdk.GetRuleGroupOutput, error) {
	return m.MockGetRuleGroup(input)
}

func (m *MockWAFV2Client) GetRuleGroupWithContext(context aws.Context, input *svcsdk.GetRuleGroupInput, option ...request.Option) (*svcsdk.GetRuleGroupOutput, error) {
	return m.MockGetRuleGroupWithContext(input)
}

func (m *MockWAFV2Client) GetRuleGroupRequest(input *svcsdk.GetRuleGroupInput) (*request.Request, *svcsdk.GetRuleGroupOutput) {
	return m.MockGetRuleGroupRequest(input)
}

func (m *MockWAFV2Client) GetSampledRequests(input *svcsdk.GetSampledRequestsInput) (*svcsdk.GetSampledRequestsOutput, error) {
	return m.MockGetSampledRequests(input)
}

func (m *MockWAFV2Client) GetSampledRequestsWithContext(context aws.Context, input *svcsdk.GetSampledRequestsInput, option ...request.Option) (*svcsdk.GetSampledRequestsOutput, error) {
	return m.MockGetSampledRequestsWithContext(input)
}

func (m *MockWAFV2Client) GetSampledRequestsRequest(input *svcsdk.GetSampledRequestsInput) (*request.Request, *svcsdk.GetSampledRequestsOutput) {
	return m.MockGetSampledRequestsRequest(input)
}

func (m *MockWAFV2Client) GetWebACL(input *svcsdk.GetWebACLInput) (*svcsdk.GetWebACLOutput, error) {
	return m.MockGetWebACL(input)
}

func (m *MockWAFV2Client) GetWebACLWithContext(context aws.Context, input *svcsdk.GetWebACLInput, option ...request.Option) (*svcsdk.GetWebACLOutput, error) {
	return m.MockGetWebACLWithContext(input)
}

func (m *MockWAFV2Client) GetWebACLRequest(input *svcsdk.GetWebACLInput) (*request.Request, *svcsdk.GetWebACLOutput) {
	return m.MockGetWebACLRequest(input)
}

func (m *MockWAFV2Client) GetWebACLForResource(input *svcsdk.GetWebACLForResourceInput) (*svcsdk.GetWebACLForResourceOutput, error) {
	return m.MockGetWebACLForResource(input)
}

func (m *MockWAFV2Client) GetWebACLForResourceWithContext(context aws.Context, input *svcsdk.GetWebACLForResourceInput, option ...request.Option) (*svcsdk.GetWebACLForResourceOutput, error) {
	return m.MockGetWebACLForResourceWithContext(input)
}

func (m *MockWAFV2Client) GetWebACLForResourceRequest(input *svcsdk.GetWebACLForResourceInput) (*request.Request, *svcsdk.GetWebACLForResourceOutput) {
	return m.MockGetWebACLForResourceRequest(input)
}

func (m *MockWAFV2Client) ListAPIKeys(input *svcsdk.ListAPIKeysInput) (*svcsdk.ListAPIKeysOutput, error) {
	return m.MockListAPIKeys(input)
}

func (m *MockWAFV2Client) ListAPIKeysWithContext(context aws.Context, input *svcsdk.ListAPIKeysInput, option ...request.Option) (*svcsdk.ListAPIKeysOutput, error) {
	return m.MockListAPIKeysWithContext(input)
}

func (m *MockWAFV2Client) ListAPIKeysRequest(input *svcsdk.ListAPIKeysInput) (*request.Request, *svcsdk.ListAPIKeysOutput) {
	return m.MockListAPIKeysRequest(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroupVersions(input *svcsdk.ListAvailableManagedRuleGroupVersionsInput) (*svcsdk.ListAvailableManagedRuleGroupVersionsOutput, error) {
	return m.MockListAvailableManagedRuleGroupVersions(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroupVersionsWithContext(context aws.Context, input *svcsdk.ListAvailableManagedRuleGroupVersionsInput, option ...request.Option) (*svcsdk.ListAvailableManagedRuleGroupVersionsOutput, error) {
	return m.MockListAvailableManagedRuleGroupVersionsWithContext(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroupVersionsRequest(input *svcsdk.ListAvailableManagedRuleGroupVersionsInput) (*request.Request, *svcsdk.ListAvailableManagedRuleGroupVersionsOutput) {
	return m.MockListAvailableManagedRuleGroupVersionsRequest(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroups(input *svcsdk.ListAvailableManagedRuleGroupsInput) (*svcsdk.ListAvailableManagedRuleGroupsOutput, error) {
	return m.MockListAvailableManagedRuleGroups(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroupsWithContext(context aws.Context, input *svcsdk.ListAvailableManagedRuleGroupsInput, option ...request.Option) (*svcsdk.ListAvailableManagedRuleGroupsOutput, error) {
	return m.MockListAvailableManagedRuleGroupsWithContext(input)
}

func (m *MockWAFV2Client) ListAvailableManagedRuleGroupsRequest(input *svcsdk.ListAvailableManagedRuleGroupsInput) (*request.Request, *svcsdk.ListAvailableManagedRuleGroupsOutput) {
	return m.MockListAvailableManagedRuleGroupsRequest(input)
}

func (m *MockWAFV2Client) ListIPSets(input *svcsdk.ListIPSetsInput) (*svcsdk.ListIPSetsOutput, error) {
	return m.MockListIPSets(input)
}

func (m *MockWAFV2Client) ListIPSetsWithContext(context aws.Context, input *svcsdk.ListIPSetsInput, option ...request.Option) (*svcsdk.ListIPSetsOutput, error) {
	return m.MockListIPSetsWithContext(input)
}

func (m *MockWAFV2Client) ListIPSetsRequest(input *svcsdk.ListIPSetsInput) (*request.Request, *svcsdk.ListIPSetsOutput) {
	return m.MockListIPSetsRequest(input)
}

func (m *MockWAFV2Client) ListLoggingConfigurations(input *svcsdk.ListLoggingConfigurationsInput) (*svcsdk.ListLoggingConfigurationsOutput, error) {
	return m.MockListLoggingConfigurations(input)
}

func (m *MockWAFV2Client) ListLoggingConfigurationsWithContext(context aws.Context, input *svcsdk.ListLoggingConfigurationsInput, option ...request.Option) (*svcsdk.ListLoggingConfigurationsOutput, error) {
	return m.MockListLoggingConfigurationsWithContext(input)
}

func (m *MockWAFV2Client) ListLoggingConfigurationsRequest(input *svcsdk.ListLoggingConfigurationsInput) (*request.Request, *svcsdk.ListLoggingConfigurationsOutput) {
	return m.MockListLoggingConfigurationsRequest(input)
}

func (m *MockWAFV2Client) ListManagedRuleSets(input *svcsdk.ListManagedRuleSetsInput) (*svcsdk.ListManagedRuleSetsOutput, error) {
	return m.MockListManagedRuleSets(input)
}

func (m *MockWAFV2Client) ListManagedRuleSetsWithContext(context aws.Context, input *svcsdk.ListManagedRuleSetsInput, option ...request.Option) (*svcsdk.ListManagedRuleSetsOutput, error) {
	return m.MockListManagedRuleSetsWithContext(input)
}

func (m *MockWAFV2Client) ListManagedRuleSetsRequest(input *svcsdk.ListManagedRuleSetsInput) (*request.Request, *svcsdk.ListManagedRuleSetsOutput) {
	return m.MockListManagedRuleSetsRequest(input)
}

func (m *MockWAFV2Client) ListMobileSdkReleases(input *svcsdk.ListMobileSdkReleasesInput) (*svcsdk.ListMobileSdkReleasesOutput, error) {
	return m.MockListMobileSdkReleases(input)
}

func (m *MockWAFV2Client) ListMobileSdkReleasesWithContext(context aws.Context, input *svcsdk.ListMobileSdkReleasesInput, option ...request.Option) (*svcsdk.ListMobileSdkReleasesOutput, error) {
	return m.MockListMobileSdkReleasesWithContext(input)
}

func (m *MockWAFV2Client) ListMobileSdkReleasesRequest(input *svcsdk.ListMobileSdkReleasesInput) (*request.Request, *svcsdk.ListMobileSdkReleasesOutput) {
	return m.MockListMobileSdkReleasesRequest(input)
}

func (m *MockWAFV2Client) ListRegexPatternSets(input *svcsdk.ListRegexPatternSetsInput) (*svcsdk.ListRegexPatternSetsOutput, error) {
	return m.MockListRegexPatternSets(input)
}

func (m *MockWAFV2Client) ListRegexPatternSetsWithContext(context aws.Context, input *svcsdk.ListRegexPatternSetsInput, option ...request.Option) (*svcsdk.ListRegexPatternSetsOutput, error) {
	return m.MockListRegexPatternSetsWithContext(input)
}

func (m *MockWAFV2Client) ListRegexPatternSetsRequest(input *svcsdk.ListRegexPatternSetsInput) (*request.Request, *svcsdk.ListRegexPatternSetsOutput) {
	return m.MockListRegexPatternSetsRequest(input)
}

func (m *MockWAFV2Client) ListResourcesForWebACL(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
	return m.MockListResourcesForWebACL(input)
}

func (m *MockWAFV2Client) ListResourcesForWebACLWithContext(context aws.Context, input *svcsdk.ListResourcesForWebACLInput, option ...request.Option) (*svcsdk.ListResourcesForWebACLOutput, error) {
	return m.MockListResourcesForWebACLWithContext(input)
}

func (m *MockWAFV2Client) ListResourcesForWebACLRequest(input *svcsdk.ListResourcesForWebACLInput) (*request.Request, *svcsdk.ListResourcesForWebACLOutput) {
	return m.MockListResourcesForWebACLRequest(input)
}

func (m *MockWAFV2Client) ListRuleGroups(input *svcsdk.ListRuleGroupsInput) (*svcsdk.ListRuleGroupsOutput, error) {
	return m.MockListRuleGroups(input)
}

func (m *MockWAFV2Client) ListRuleGroupsWithContext(context aws.Context, input *svcsdk.ListRuleGroupsInput, option ...request.Option) (*svcsdk.ListRuleGroupsOutput, error) {
	return m.MockListRuleGroupsWithContext(input)
}

func (m *MockWAFV2Client) ListRuleGroupsRequest(input *svcsdk.ListRuleGroupsInput) (*request.Request, *svcsdk.ListRuleGroupsOutput) {
	return m.MockListRuleGroupsRequest(input)
}

func (m *MockWAFV2Client) ListTagsForResourceWithContext(context aws.Context, input *svcsdk.ListTagsForResourceInput, option ...request.Option) (*svcsdk.ListTagsForResourceOutput, error) {
	return m.MockListTagsForResourceWithContext(input)
}

func (m *MockWAFV2Client) ListTagsForResourceRequest(input *svcsdk.ListTagsForResourceInput) (*request.Request, *svcsdk.ListTagsForResourceOutput) {
	return m.MockListTagsForResourceRequest(input)
}

func (m *MockWAFV2Client) ListWebACLs(input *svcsdk.ListWebACLsInput) (*svcsdk.ListWebACLsOutput, error) {
	return m.MockListWebACLs(input)
}

func (m *MockWAFV2Client) ListWebACLsWithContext(context aws.Context, input *svcsdk.ListWebACLsInput, option ...request.Option) (*svcsdk.ListWebACLsOutput, error) {
	return m.MockListWebACLsWithContext(input)
}

func (m *MockWAFV2Client) ListWebACLsRequest(input *svcsdk.ListWebACLsInput) (*request.Request, *svcsdk.ListWebACLsOutput) {
	return m.MockListWebACLsRequest(input)
}

func (m *MockWAFV2Client) PutLoggingConfiguration(input *svcsdk.PutLoggingConfigurationInput) (*svcsdk.PutLoggingConfigurationOutput, error) {
	return m.MockPutLoggingConfiguration(input)
}

func (m *MockWAFV2Client) PutLoggingConfigurationWithContext(context aws.Context, input *svcsdk.PutLoggingConfigurationInput, option ...request.Option) (*svcsdk.PutLoggingConfigurationOutput, error) {
	return m.MockPutLoggingConfigurationWithContext(input)
}

func (m *MockWAFV2Client) PutLoggingConfigurationRequest(input *svcsdk.PutLoggingConfigurationInput) (*request.Request, *svcsdk.PutLoggingConfigurationOutput) {
	return m.MockPutLoggingConfigurationRequest(input)
}

func (m *MockWAFV2Client) PutManagedRuleSetVersions(input *svcsdk.PutManagedRuleSetVersionsInput) (*svcsdk.PutManagedRuleSetVersionsOutput, error) {
	return m.MockPutManagedRuleSetVersions(input)
}

func (m *MockWAFV2Client) PutManagedRuleSetVersionsWithContext(context aws.Context, input *svcsdk.PutManagedRuleSetVersionsInput, option ...request.Option) (*svcsdk.PutManagedRuleSetVersionsOutput, error) {
	return m.MockPutManagedRuleSetVersionsWithContext(input)
}

func (m *MockWAFV2Client) PutManagedRuleSetVersionsRequest(input *svcsdk.PutManagedRuleSetVersionsInput) (*request.Request, *svcsdk.PutManagedRuleSetVersionsOutput) {
	return m.MockPutManagedRuleSetVersionsRequest(input)
}

func (m *MockWAFV2Client) PutPermissionPolicy(input *svcsdk.PutPermissionPolicyInput) (*svcsdk.PutPermissionPolicyOutput, error) {
	return m.MockPutPermissionPolicy(input)
}

func (m *MockWAFV2Client) PutPermissionPolicyWithContext(context aws.Context, input *svcsdk.PutPermissionPolicyInput, option ...request.Option) (*svcsdk.PutPermissionPolicyOutput, error) {
	return m.MockPutPermissionPolicyWithContext(input)
}

func (m *MockWAFV2Client) PutPermissionPolicyRequest(input *svcsdk.PutPermissionPolicyInput) (*request.Request, *svcsdk.PutPermissionPolicyOutput) {
	return m.MockPutPermissionPolicyRequest(input)
}

func (m *MockWAFV2Client) TagResourceWithContext(context aws.Context, input *svcsdk.TagResourceInput, option ...request.Option) (*svcsdk.TagResourceOutput, error) {
	return m.MockTagResourceWithContext(input)
}

func (m *MockWAFV2Client) TagResourceRequest(input *svcsdk.TagResourceInput) (*request.Request, *svcsdk.TagResourceOutput) {
	return m.MockTagResourceRequest(input)
}

func (m *MockWAFV2Client) UntagResource(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error) {
	return m.MockUntagResource(input)
}

func (m *MockWAFV2Client) UntagResourceWithContext(context aws.Context, input *svcsdk.UntagResourceInput, option ...request.Option) (*svcsdk.UntagResourceOutput, error) {
	return m.MockUntagResourceWithContext(input)
}

func (m *MockWAFV2Client) UntagResourceRequest(input *svcsdk.UntagResourceInput) (*request.Request, *svcsdk.UntagResourceOutput) {
	return m.MockUntagResourceRequest(input)
}

func (m *MockWAFV2Client) UpdateIPSet(input *svcsdk.UpdateIPSetInput) (*svcsdk.UpdateIPSetOutput, error) {
	return m.MockUpdateIPSet(input)
}

func (m *MockWAFV2Client) UpdateIPSetWithContext(context aws.Context, input *svcsdk.UpdateIPSetInput, option ...request.Option) (*svcsdk.UpdateIPSetOutput, error) {
	return m.MockUpdateIPSetWithContext(input)
}

func (m *MockWAFV2Client) UpdateIPSetRequest(input *svcsdk.UpdateIPSetInput) (*request.Request, *svcsdk.UpdateIPSetOutput) {
	return m.MockUpdateIPSetRequest(input)
}

func (m *MockWAFV2Client) UpdateManagedRuleSetVersionExpiryDate(input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput) (*svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput, error) {
	return m.MockUpdateManagedRuleSetVersionExpiryDate(input)
}

func (m *MockWAFV2Client) UpdateManagedRuleSetVersionExpiryDateWithContext(context aws.Context, input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput, option ...request.Option) (*svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput, error) {
	return m.MockUpdateManagedRuleSetVersionExpiryDateWithContext(input)
}

func (m *MockWAFV2Client) UpdateManagedRuleSetVersionExpiryDateRequest(input *svcsdk.UpdateManagedRuleSetVersionExpiryDateInput) (*request.Request, *svcsdk.UpdateManagedRuleSetVersionExpiryDateOutput) {
	return m.MockUpdateManagedRuleSetVersionExpiryDateRequest(input)
}

func (m *MockWAFV2Client) UpdateRegexPatternSet(input *svcsdk.UpdateRegexPatternSetInput) (*svcsdk.UpdateRegexPatternSetOutput, error) {
	return m.MockUpdateRegexPatternSet(input)
}

func (m *MockWAFV2Client) UpdateRegexPatternSetWithContext(context aws.Context, input *svcsdk.UpdateRegexPatternSetInput, option ...request.Option) (*svcsdk.UpdateRegexPatternSetOutput, error) {
	return m.MockUpdateRegexPatternSetWithContext(input)
}

func (m *MockWAFV2Client) UpdateRegexPatternSetRequest(input *svcsdk.UpdateRegexPatternSetInput) (*request.Request, *svcsdk.UpdateRegexPatternSetOutput) {
	return m.MockUpdateRegexPatternSetRequest(input)
}

func (m *MockWAFV2Client) UpdateRuleGroup(input *svcsdk.UpdateRuleGroupInput) (*svcsdk.UpdateRuleGroupOutput, error) {
	return m.MockUpdateRuleGroup(input)
}

func (m *MockWAFV2Client) UpdateRuleGroupWithContext(context aws.Context, input *svcsdk.UpdateRuleGroupInput, option ...request.Option) (*svcsdk.UpdateRuleGroupOutput, error) {
	return m.MockUpdateRuleGroupWithContext(input)
}

func (m *MockWAFV2Client) UpdateRuleGroupRequest(input *svcsdk.UpdateRuleGroupInput) (*request.Request, *svcsdk.UpdateRuleGroupOutput) {
	return m.MockUpdateRuleGroupRequest(input)
}

func (m *MockWAFV2Client) UpdateWebACL(input *svcsdk.UpdateWebACLInput) (*svcsdk.UpdateWebACLOutput, error) {
	return m.MockUpdateWebACL(input)
}

func (m *MockWAFV2Client) UpdateWebACLWithContext(context aws.Context, input *svcsdk.UpdateWebACLInput, option ...request.Option) (*svcsdk.UpdateWebACLOutput, error) {
	return m.MockUpdateWebACLWithContext(input)
}

func (m *MockWAFV2Client) UpdateWebACLRequest(input *svcsdk.UpdateWebACLInput) (*request.Request, *svcsdk.UpdateWebACLOutput) {
	return m.MockUpdateWebACLRequest(input)
}

// ListTagsForResource mocks ListTagsForResource method
func (m *MockWAFV2Client) ListTagsForResource(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
	return m.MockListTagsForResource(input)
}

// TagResource mocks TagResource method
func (m *MockWAFV2Client) TagResource(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error) {
	return m.MockTagResource(input)
}
