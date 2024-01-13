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

package servicecatalog

import (
	"context"
	"errors"

	cfsdkv2 "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfsdkv2types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"
	svcsdkapi "github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface"
)

const (
	cloudformationArnOutputName       = "CloudformationStackARN"
	errCloudformationStackArnNotFound = "provisioned product outputs do not contain cloudformation stack arn"
)

// Client represents a custom client to retrieve information from AWS related to service catalog or cloud formation as resource behind the provisioned product
type Client interface {
	GetCloudformationStackParameters([]*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error)
	GetProvisionedProductOutputs(*svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error)
	DescribeRecord(*svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error)
	DescribeProduct(*svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error)
}

// CustomServiceCatalogClient is the implementation of a Client
type CustomServiceCatalogClient struct {
	CfClient *cfsdkv2.Client
	Client   svcsdkapi.ServiceCatalogAPI
}

// GetCloudformationStackParameters retrieves parameters from cloudformation stack based on outputs from provisioned product
func (c *CustomServiceCatalogClient) GetCloudformationStackParameters(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
	describeCfStacksInput := cfsdkv2.DescribeStacksInput{}
	for i, output := range provisionedProductOutputs {
		if *output.OutputKey == cloudformationArnOutputName {
			describeCfStacksInput.StackName = output.OutputValue
			break
		}
		if i+1 == len(provisionedProductOutputs) {
			return []cfsdkv2types.Parameter{}, errors.New(errCloudformationStackArnNotFound)
		}
	}
	describeCfStacksOutput, err := c.CfClient.DescribeStacks(context.TODO(), &describeCfStacksInput)
	if err != nil {
		return []cfsdkv2types.Parameter{}, err
	}
	return describeCfStacksOutput.Stacks[0].Parameters, nil
}

// GetProvisionedProductOutputs is wrapped (*ServiceCatalog) GetProvisionedProductOutputs from github.com/aws/aws-sdk-go/service/servicecatalog
func (c *CustomServiceCatalogClient) GetProvisionedProductOutputs(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
	getPPOutput, err := c.Client.GetProvisionedProductOutputs(getPPInput)
	return getPPOutput, err
}

// DescribeRecord is wrapped (*ServiceCatalog) DescribeRecord from github.com/aws/aws-sdk-go/service/servicecatalog
func (c *CustomServiceCatalogClient) DescribeRecord(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
	describeRecordOutput, err := c.Client.DescribeRecord(describeRecordInput)
	return describeRecordOutput, err
}

// DescribeProduct is wrapped (*ServiceCatalog) DescribeProduct from github.com/aws/aws-sdk-go/service/servicecatalog
func (c *CustomServiceCatalogClient) DescribeProduct(input *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
	return c.Client.DescribeProduct(input)
}
