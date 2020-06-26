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
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/elasticloadbalancingiface"
)

var _ elasticloadbalancingiface.ClientAPI = &MockClient{}

// MockClient is a fake implementation of elasticloadbalancingiface.ClientAPI.
type MockClient struct {
	elasticloadbalancingiface.ClientAPI

	MockDescribeLoadBalancersRequest                   func(*elb.DescribeLoadBalancersInput) elb.DescribeLoadBalancersRequest
	MockCreateLoadBalancerRequest                      func(*elb.CreateLoadBalancerInput) elb.CreateLoadBalancerRequest
	MockDeleteLoadBalancerRequest                      func(*elb.DeleteLoadBalancerInput) elb.DeleteLoadBalancerRequest
	MockEnableAvailabilityZonesForLoadBalancerRequest  func(*elb.EnableAvailabilityZonesForLoadBalancerInput) elb.EnableAvailabilityZonesForLoadBalancerRequest
	MockDisableAvailabilityZonesForLoadBalancerRequest func(*elb.DisableAvailabilityZonesForLoadBalancerInput) elb.DisableAvailabilityZonesForLoadBalancerRequest
	MockDetachLoadBalancerFromSubnetsRequest           func(*elb.DetachLoadBalancerFromSubnetsInput) elb.DetachLoadBalancerFromSubnetsRequest
	MockAttachLoadBalancerToSubnetsRequest             func(*elb.AttachLoadBalancerToSubnetsInput) elb.AttachLoadBalancerToSubnetsRequest
	MockApplySecurityGroupsToLoadBalancerRequest       func(*elb.ApplySecurityGroupsToLoadBalancerInput) elb.ApplySecurityGroupsToLoadBalancerRequest
	MockCreateLoadBalancerListenersRequest             func(*elb.CreateLoadBalancerListenersInput) elb.CreateLoadBalancerListenersRequest
	MockDeleteLoadBalancerListenersRequest             func(*elb.DeleteLoadBalancerListenersInput) elb.DeleteLoadBalancerListenersRequest
	MockRegisterInstancesWithLoadBalancerRequest       func(*elb.RegisterInstancesWithLoadBalancerInput) elb.RegisterInstancesWithLoadBalancerRequest
	MockDeregisterInstancesFromLoadBalancerRequest     func(*elb.DeregisterInstancesFromLoadBalancerInput) elb.DeregisterInstancesFromLoadBalancerRequest
}

// DescribeLoadBalancersRequest calls the underlying
// MockDescribeLoadBalancersRequest method.
func (c *MockClient) DescribeLoadBalancersRequest(i *elasticloadbalancing.DescribeLoadBalancersInput) elasticloadbalancing.DescribeLoadBalancersRequest {
	return c.MockDescribeLoadBalancersRequest(i)
}

// CreateLoadBalancerRequest calls the underlying
// MockCreateLoadBalancerRequest method.
func (c *MockClient) CreateLoadBalancerRequest(i *elasticloadbalancing.CreateLoadBalancerInput) elasticloadbalancing.CreateLoadBalancerRequest {
	return c.MockCreateLoadBalancerRequest(i)
}

// DeleteLoadBalancerRequest calls the underlying
// MockDeleteLoadBalancerRequest method.
func (c *MockClient) DeleteLoadBalancerRequest(i *elasticloadbalancing.DeleteLoadBalancerInput) elasticloadbalancing.DeleteLoadBalancerRequest {
	return c.MockDeleteLoadBalancerRequest(i)
}

// EnableAvailabilityZonesForLoadBalancerRequest calls the underlying
// MockEnableAvailabilityZonesForLoadBalancerRequest method.
func (c *MockClient) EnableAvailabilityZonesForLoadBalancerRequest(i *elasticloadbalancing.EnableAvailabilityZonesForLoadBalancerInput) elasticloadbalancing.EnableAvailabilityZonesForLoadBalancerRequest {
	return c.MockEnableAvailabilityZonesForLoadBalancerRequest(i)
}

// DisableAvailabilityZonesForLoadBalancerRequest calls the underlying
// MockDisableAvailabilityZonesForLoadBalancerRequest method.
func (c *MockClient) DisableAvailabilityZonesForLoadBalancerRequest(i *elasticloadbalancing.DisableAvailabilityZonesForLoadBalancerInput) elasticloadbalancing.DisableAvailabilityZonesForLoadBalancerRequest {
	return c.MockDisableAvailabilityZonesForLoadBalancerRequest(i)
}

// DetachLoadBalancerFromSubnetsRequest calls the underlying
// MockDetachLoadBalancerFromSubnetsRequest method.
func (c *MockClient) DetachLoadBalancerFromSubnetsRequest(i *elasticloadbalancing.DetachLoadBalancerFromSubnetsInput) elasticloadbalancing.DetachLoadBalancerFromSubnetsRequest {
	return c.MockDetachLoadBalancerFromSubnetsRequest(i)
}

// AttachLoadBalancerToSubnetsRequest calls the underlying
// MockAttachLoadBalancerToSubnetsRequest method.
func (c *MockClient) AttachLoadBalancerToSubnetsRequest(i *elasticloadbalancing.AttachLoadBalancerToSubnetsInput) elasticloadbalancing.AttachLoadBalancerToSubnetsRequest {
	return c.MockAttachLoadBalancerToSubnetsRequest(i)
}

// ApplySecurityGroupsToLoadBalancerRequest calls the underlying
// MockApplySecurityGroupsToLoadBalancerRequest method.
func (c *MockClient) ApplySecurityGroupsToLoadBalancerRequest(i *elasticloadbalancing.ApplySecurityGroupsToLoadBalancerInput) elasticloadbalancing.ApplySecurityGroupsToLoadBalancerRequest {
	return c.MockApplySecurityGroupsToLoadBalancerRequest(i)
}

// CreateLoadBalancerListenersRequest calls the underlying
// MockCreateLoadBalancerListenersRequest method.
func (c *MockClient) CreateLoadBalancerListenersRequest(i *elasticloadbalancing.CreateLoadBalancerListenersInput) elasticloadbalancing.CreateLoadBalancerListenersRequest {
	return c.MockCreateLoadBalancerListenersRequest(i)
}

// DeleteLoadBalancerListenersRequest calls the underlying
// MockDeleteLoadBalancerListenersRequest method.
func (c *MockClient) DeleteLoadBalancerListenersRequest(i *elasticloadbalancing.DeleteLoadBalancerListenersInput) elasticloadbalancing.DeleteLoadBalancerListenersRequest {
	return c.MockDeleteLoadBalancerListenersRequest(i)
}

// RegisterInstancesWithLoadBalancerRequest calls the underlying
// MockRegisterInstancesWithLoadBalancerRequest method.
func (c *MockClient) RegisterInstancesWithLoadBalancerRequest(i *elasticloadbalancing.RegisterInstancesWithLoadBalancerInput) elasticloadbalancing.RegisterInstancesWithLoadBalancerRequest {
	return c.MockRegisterInstancesWithLoadBalancerRequest(i)
}

// DeregisterInstancesFromLoadBalancerRequest calls the underlying
// MockDeregisterInstancesWithLoadBalancerRequest method.
func (c *MockClient) DeregisterInstancesFromLoadBalancerRequest(i *elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput) elasticloadbalancing.DeregisterInstancesFromLoadBalancerRequest {
	return c.MockDeregisterInstancesFromLoadBalancerRequest(i)
}
