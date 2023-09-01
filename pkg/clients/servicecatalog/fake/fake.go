/*
Copyright 2023 The Crossplane Authors.

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
	"context"

	cfsdkv2types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/servicecatalog"
)

var _ clientset.Client = (*MockCustomServiceCatalogClient)(nil)

// MockCustomServiceCatalogClient is a type that implements all the methods for Client interface
type MockCustomServiceCatalogClient struct {
	MockGetCloudformationStackParameters    func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error)
	MockGetProvisionedProductOutputs        func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error)
	MockDescribeRecord                      func(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error)
	MockDescribeProduct                     func(*svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error)
	MockDescribeProvisioningArtifact        func(*svcsdk.DescribeProvisioningArtifactInput) (*svcsdk.DescribeProvisioningArtifactOutput, error)
	MockUpdateProvisionedProductWithContext func(context.Context, *svcsdk.UpdateProvisionedProductInput, ...request.Option) (*svcsdk.UpdateProvisionedProductOutput, error)
}

// GetCloudformationStackParameters mocks GetCloudformationStackParameters method
func (m *MockCustomServiceCatalogClient) GetCloudformationStackParameters(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
	return m.MockGetCloudformationStackParameters(provisionedProductOutputs)
}

// GetProvisionedProductOutputs mocks GetProvisionedProductOutputs method
func (m *MockCustomServiceCatalogClient) GetProvisionedProductOutputs(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
	return m.MockGetProvisionedProductOutputs(getPPInput)
}

// DescribeRecord mocks DescribeRecord method
func (m *MockCustomServiceCatalogClient) DescribeRecord(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
	return m.MockDescribeRecord(describeRecordInput)
}

// DescribeProvisioningArtifact mocks DescribeProvisioningArtifact method
func (m *MockCustomServiceCatalogClient) DescribeProvisioningArtifact(input *svcsdk.DescribeProvisioningArtifactInput) (*svcsdk.DescribeProvisioningArtifactOutput, error) {
	return m.MockDescribeProvisioningArtifact(input)
}

// UpdateProvisionedProductWithContext mocks UpdateProvisionedProductWithContext method
func (m *MockCustomServiceCatalogClient) UpdateProvisionedProductWithContext(ctx context.Context, input *svcsdk.UpdateProvisionedProductInput, opts ...request.Option) (*svcsdk.UpdateProvisionedProductOutput, error) {
	return m.MockUpdateProvisionedProductWithContext(ctx, input, opts...)
}

// DescribeProduct mocks DescribeProduct method
func (m *MockCustomServiceCatalogClient) DescribeProduct(input *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
	return m.MockDescribeProduct(input)
}
