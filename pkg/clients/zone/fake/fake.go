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

package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	clientset "github.com/crossplane/provider-aws/pkg/clients/zone"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockZoneClient)(nil)

// MockZoneClient is a type that implements all the methods for zone Client interface
type MockZoneClient struct {
	MockGetZoneRequest    func(id *string) route53.GetHostedZoneRequest
	MockCreateZoneRequest func(cr *v1alpha3.Zone) route53.CreateHostedZoneRequest
	MockUpdateZoneRequest func(id, comment *string) route53.UpdateHostedZoneCommentRequest
	MockDeleteZoneRequest func(id *string) route53.DeleteHostedZoneRequest
}

// GetZoneRequest mocks GetZoneRequest method
func (m *MockZoneClient) GetZoneRequest(id *string) route53.GetHostedZoneRequest {
	return m.MockGetZoneRequest(id)
}

// CreateZoneRequest mocks CreateZoneRequest method
func (m *MockZoneClient) CreateZoneRequest(cr *v1alpha3.Zone) route53.CreateHostedZoneRequest {
	return m.MockCreateZoneRequest(cr)
}

// UpdateZoneRequest mocks UpdateZoneRequest method
func (m *MockZoneClient) UpdateZoneRequest(id, comment *string) route53.UpdateHostedZoneCommentRequest {
	return m.MockUpdateZoneRequest(id, comment)
}

// DeleteZoneRequest mocks DeleteZoneRequest method
func (m *MockZoneClient) DeleteZoneRequest(id *string) route53.DeleteHostedZoneRequest {
	return m.MockDeleteZoneRequest(id)
}
