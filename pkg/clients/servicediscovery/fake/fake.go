package fake

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

// MockServicediscoveryClient is the mocked service client
type MockServicediscoveryClient struct {
	servicediscoveryiface.ServiceDiscoveryAPI
	// MockGetOperation is a function pointer
	MockGetOperation func(*svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error)
	// MockGetNamespace is a function pointer
	MockGetNamespace func(*svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error)
	// MockCreatePrivateDNSNamespace is a function pointer
	MockCreatePrivateDNSNamespace func(*svcsdk.CreatePrivateDnsNamespaceInput) (*svcsdk.CreatePrivateDnsNamespaceOutput, error)
	// MockCreatePublicDNSNamespace is a function pointer
	MockCreatePublicDNSNamespace func(input *svcsdk.CreatePublicDnsNamespaceInput) (*svcsdk.CreatePublicDnsNamespaceOutput, error)
	// MockCreateHTTPNamespace is a function pointer
	MockCreateHTTPNamespace func(input *svcsdk.CreateHttpNamespaceInput) (*svcsdk.CreateHttpNamespaceOutput, error)
	// MockDeleteNamespace is a function pointer
	MockDeleteNamespace func(*svcsdk.DeleteNamespaceInput) (*svcsdk.DeleteNamespaceOutput, error)
	// MockGetOperationRequest is a function pointer
	MockGetOperationRequest func(*svcsdk.GetOperationInput) (*request.Request, *svcsdk.GetOperationOutput)
	// MockGetNamespaceRequest is a function pointer
	MockGetNamespaceRequest func(*svcsdk.GetNamespaceInput) (*request.Request, *svcsdk.GetNamespaceOutput)
	// MockCreatePrivateDNSNamespaceRequest is a function pointer
	MockCreatePrivateDNSNamespaceRequest func(*svcsdk.CreatePrivateDnsNamespaceInput) (*request.Request, *svcsdk.CreatePrivateDnsNamespaceOutput)
	// MockCreatePublicDNSNamespaceRequest is a function pointer
	MockCreatePublicDNSNamespaceRequest func(*svcsdk.CreatePublicDnsNamespaceInput) (*request.Request, *svcsdk.CreatePublicDnsNamespaceOutput)
	// MockCreateHTTPNamespaceRequest is a function pointer
	MockCreateHTTPNamespaceRequest func(*svcsdk.CreateHttpNamespaceInput) (*request.Request, *svcsdk.CreateHttpNamespaceOutput)
	// MockDeleteNamespaceRequest is a function pointer
	MockDeleteNamespaceRequest func(*svcsdk.DeleteNamespaceInput) (*request.Request, *svcsdk.DeleteNamespaceOutput)
	// MockListTagsForResource is a function pointer
	MockListTagsForResource func(*svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error)
	// MockUntagResource is a function pointer
	MockUntagResource func(*svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error)
	// MockTagResource is a function pointer
	MockTagResource func(*svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error)
}

// CreatePrivateDnsNamespace is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePrivateDnsNamespace(input *svcsdk.CreatePrivateDnsNamespaceInput) (*svcsdk.CreatePrivateDnsNamespaceOutput, error) {
	if m.MockCreatePrivateDNSNamespace == nil {
		fmt.Println(".MockCreatePrivateDNSNamespace == nil")
		return &svcsdk.CreatePrivateDnsNamespaceOutput{}, nil
	}
	return m.MockCreatePrivateDNSNamespace(input)
}

// CreatePublicDnsNamespace is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePublicDnsNamespace(input *svcsdk.CreatePublicDnsNamespaceInput) (*svcsdk.CreatePublicDnsNamespaceOutput, error) {
	if m.MockCreatePublicDNSNamespace == nil {
		fmt.Println(".MockCreatePublicDNSNamespace == nil")
		return &svcsdk.CreatePublicDnsNamespaceOutput{}, nil
	}
	return m.MockCreatePublicDNSNamespace(input)
}

// CreateHttpNamespace is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreateHttpNamespace(input *svcsdk.CreateHttpNamespaceInput) (*svcsdk.CreateHttpNamespaceOutput, error) {
	if m.MockCreateHTTPNamespace == nil {
		fmt.Println(".MockCreateHTTPNamespace == nil")
		return &svcsdk.CreateHttpNamespaceOutput{}, nil
	}
	return m.MockCreateHTTPNamespace(input)
}

// CreatePrivateDnsNamespaceWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePrivateDnsNamespaceWithContext(_ context.Context, input *svcsdk.CreatePrivateDnsNamespaceInput, _ ...request.Option) (*svcsdk.CreatePrivateDnsNamespaceOutput, error) {
	if m.MockCreatePrivateDNSNamespace == nil {
		fmt.Println(".MockCreatePrivateDNSNamespace == nil")
		return &svcsdk.CreatePrivateDnsNamespaceOutput{}, nil
	}
	return m.MockCreatePrivateDNSNamespace(input)
}

// CreatePublicDnsNamespaceWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePublicDnsNamespaceWithContext(_ context.Context, input *svcsdk.CreatePublicDnsNamespaceInput, _ ...request.Option) (*svcsdk.CreatePublicDnsNamespaceOutput, error) { //nolint:golint
	if m.MockCreatePublicDNSNamespace == nil {
		fmt.Println(".MockCreatePublicDnsNamespace == nil")
		return &svcsdk.CreatePublicDnsNamespaceOutput{}, nil
	}
	return m.MockCreatePublicDNSNamespace(input)
}

// CreateHttpNamespaceWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreateHttpNamespaceWithContext(_ context.Context, input *svcsdk.CreateHttpNamespaceInput, _ ...request.Option) (*svcsdk.CreateHttpNamespaceOutput, error) { //nolint:golint
	if m.MockCreateHTTPNamespace == nil {
		fmt.Println(".MockCreateHTTPNamespace == nil")
		return &svcsdk.CreateHttpNamespaceOutput{}, nil
	}
	return m.MockCreateHTTPNamespace(input)
}

// GetOperation is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetOperation(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) { //nolint:golint
	if m.MockGetOperation == nil {
		fmt.Println(".MockGetOperation == nil")
		return &svcsdk.GetOperationOutput{}, nil
	}
	return m.MockGetOperation(input)
}

// GetOperationWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetOperationWithContext(_ context.Context, input *svcsdk.GetOperationInput, _ ...request.Option) (*svcsdk.GetOperationOutput, error) { //nolint:golint
	if m.MockGetOperation == nil {
		fmt.Println(".MockGetOperation == nil")
		return &svcsdk.GetOperationOutput{}, nil
	}
	return m.MockGetOperation(input)
}

// GetNamespace is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetNamespace(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) { //nolint:golint
	if m.MockGetNamespace == nil {
		fmt.Println(".MockGetNamespace == nil")
		return &svcsdk.GetNamespaceOutput{}, nil
	}
	return m.MockGetNamespace(input)
}

// GetNamespaceWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetNamespaceWithContext(_ context.Context, input *svcsdk.GetNamespaceInput, _ ...request.Option) (*svcsdk.GetNamespaceOutput, error) { //nolint:golint
	if m.MockGetNamespace == nil {
		fmt.Println(".MockGetNamespace == nil")
		return &svcsdk.GetNamespaceOutput{}, nil
	}
	return m.MockGetNamespace(input)
}

// DeleteNamespace is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) DeleteNamespace(input *svcsdk.DeleteNamespaceInput) (*svcsdk.DeleteNamespaceOutput, error) { //nolint:golint
	if m.MockDeleteNamespace == nil {
		fmt.Println(".MockDeleteNamespace == nil")
		return &svcsdk.DeleteNamespaceOutput{}, nil
	}
	return m.MockDeleteNamespace(input)
}

// DeleteNamespaceWithContext is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) DeleteNamespaceWithContext(_ context.Context, input *svcsdk.DeleteNamespaceInput, _ ...request.Option) (*svcsdk.DeleteNamespaceOutput, error) { //nolint:golint
	if m.MockDeleteNamespace == nil {
		fmt.Println(".MockDeleteNamespace == nil")
		return &svcsdk.DeleteNamespaceOutput{}, nil
	}
	return m.MockDeleteNamespace(input)
}

// CreatePrivateDnsNamespaceRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePrivateDnsNamespaceRequest(input *svcsdk.CreatePrivateDnsNamespaceInput) (*request.Request, *svcsdk.CreatePrivateDnsNamespaceOutput) {
	if m.MockCreatePrivateDNSNamespaceRequest == nil {
		fmt.Println(".MockCreatePrivateDNSNamespaceRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockCreatePrivateDNSNamespaceRequest(input)
}

// CreatePublicDnsNamespaceRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreatePublicDnsNamespaceRequest(input *svcsdk.CreatePublicDnsNamespaceInput) (*request.Request, *svcsdk.CreatePublicDnsNamespaceOutput) {
	if m.MockCreatePublicDNSNamespaceRequest == nil {
		fmt.Println(".MockCreatePublicDNSNamespaceRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockCreatePublicDNSNamespaceRequest(input)
}

// CreateHttpNamespaceRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) CreateHttpNamespaceRequest(input *svcsdk.CreateHttpNamespaceInput) (*request.Request, *svcsdk.CreateHttpNamespaceOutput) {
	if m.MockCreateHTTPNamespaceRequest == nil {
		fmt.Println(".MockCreateHTTPNamespaceRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockCreateHTTPNamespaceRequest(input)
}

// GetOperationRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetOperationRequest(input *svcsdk.GetOperationInput) (*request.Request, *svcsdk.GetOperationOutput) { //nolint:golint
	if m.MockGetOperationRequest == nil {
		fmt.Println(".MockGetOperationRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockGetOperationRequest(input)
}

// GetOperationWithContextRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetOperationWithContextRequest(_ context.Context, input *svcsdk.GetOperationInput) (*request.Request, *svcsdk.GetOperationOutput) { //nolint:golint
	if m.MockGetOperationRequest == nil {
		fmt.Println(".MockGetOperationRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockGetOperationRequest(input)
}

// GetNamespaceRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetNamespaceRequest(input *svcsdk.GetNamespaceInput) (*request.Request, *svcsdk.GetNamespaceOutput) { //nolint:golint
	if m.MockGetNamespaceRequest == nil {
		fmt.Println(".MockGetNamespaceRequest == nil")
		return &request.Request{}, nil
	}
	return m.MockGetNamespaceRequest(input)
}

// GetNamespaceWithContextRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) GetNamespaceWithContextRequest(_ context.Context, input *svcsdk.GetNamespaceInput) (*request.Request, *svcsdk.GetNamespaceOutput) { //nolint:golint
	if m.MockGetNamespaceRequest == nil {
		return &request.Request{}, nil
	}
	return m.MockGetNamespaceRequest(input)
}

// DeleteNamespaceRequest is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) DeleteNamespaceRequest(input *svcsdk.DeleteNamespaceInput) (*request.Request, *svcsdk.DeleteNamespaceOutput) { //nolint:golint
	if m.MockDeleteNamespaceRequest == nil {
		return &request.Request{}, nil
	}
	return m.MockDeleteNamespaceRequest(input)
}

// ListTagsForResource is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) ListTagsForResource(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) { //nolint:golint
	if m.MockListTagsForResource == nil {
		return &svcsdk.ListTagsForResourceOutput{}, nil
	}
	return m.MockListTagsForResource(input)
}

// UntagResource is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) UntagResource(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error) { //nolint:golint
	if m.MockUntagResource == nil {
		return &svcsdk.UntagResourceOutput{}, nil
	}
	return m.MockUntagResource(input)
}

// TagResource is the interface function to call the mock function pointer
func (m *MockServicediscoveryClient) TagResource(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error) { //nolint:golint
	if m.MockTagResource == nil {
		return &svcsdk.TagResourceOutput{}, nil
	}
	return m.MockTagResource(input)
}
