/*
Copyright 2019 The Crossplane Authors.

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

package apigateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	svcksdkapi "github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
	jsonpatch "github.com/mattbaird/jsonpatch"
)

// Client represents a custom client to retrieve information from AWS related to apigateway svc
type Client interface {
	GetRestAPIRootResource(context.Context, *string) (*string, error)
	GetRestAPIByID(context.Context, *string) (*svcsdk.RestApi, error)
	GetResource(context.Context, *svcsdk.GetResourceInput, ...request.Option) (*svcsdk.Resource, error)
}

// GatewayClient is the actual implementation of a Client to retrieve information from AWS. This is required because Update*Input shapes in this service expect a jsonpatch as its body. instead of the actual fields of the resource to be passed
type GatewayClient struct {
	Client svcksdkapi.APIGatewayAPI
}

// GetRestAPIByID retrieves the current values of sdk RestApi resource from AWS represented by id
func (c *GatewayClient) GetRestAPIByID(ctx context.Context, id *string) (*svcsdk.RestApi, error) {
	return c.Client.GetRestApiWithContext(ctx, &svcsdk.GetRestApiInput{RestApiId: id})
}

// GetRestAPIRootResource retrieves the root resource for a RestApi from AWS represented by id
func (c *GatewayClient) GetRestAPIRootResource(ctx context.Context, id *string) (*string, error) {
	out, err := c.Client.GetResourcesWithContext(ctx, &svcsdk.GetResourcesInput{RestApiId: id})
	if err != nil {
		return nil, err
	}
	for _, item := range out.Items {
		if item.ParentId == nil {
			return item.Id, nil
		}
	}
	return nil, errors.New("parent not found")
}

// GetResource retrieves the current values for a Resource from AWS using a GetResourceInput. This is just a wrapper to simplify mocking
func (c *GatewayClient) GetResource(ctx context.Context, in *svcsdk.GetResourceInput, opts ...request.Option) (*svcsdk.Resource, error) {
	return c.Client.GetResourceWithContext(ctx, in, opts...)
}

// GetJSONPatch returns a jsonpatch of diff between source and dest
func GetJSONPatch(source interface{}, dest interface{}) ([]jsonpatch.JsonPatchOperation, error) {
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	destinationJSON, err := json.Marshal(dest)
	if err != nil {
		return nil, err
	}

	return jsonpatch.CreatePatch(sourceJSON, destinationJSON)
}

// GetPatchOperations returns a jsonpatch of diff between source and dest adapted to the type expected by the Update*Input shapes of this service
func GetPatchOperations(source interface{}, dest interface{}) ([]*svcsdk.PatchOperation, error) {
	patchJSON, err := GetJSONPatch(source, dest)
	if err != nil {
		return nil, err
	}

	ops := make([]*svcsdk.PatchOperation, 0)
	for _, v := range patchJSON {
		val := v
		op := &svcsdk.PatchOperation{
			Op:   &val.Operation,
			Path: &val.Path,
		}
		if v.Operation == "add" || v.Operation == "replace" || v.Operation == "copy" {
			op.Value = aws.String(fmt.Sprintf("%v", v.Value))
		}
		ops = append(ops, op)
	}

	return ops, nil
}
