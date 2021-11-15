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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

// MockClient is a mock of the elb client
type MockClient struct {
	MockDescribeLoadBalancers                   func(ctx context.Context, input *elb.DescribeLoadBalancersInput, opts []func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error)
	MockCreateLoadBalancer                      func(ctx context.Context, input *elb.CreateLoadBalancerInput, opts []func(*elb.Options)) (*elb.CreateLoadBalancerOutput, error)
	MockDeleteLoadBalancer                      func(ctx context.Context, input *elb.DeleteLoadBalancerInput, opts []func(*elb.Options)) (*elb.DeleteLoadBalancerOutput, error)
	MockEnableAvailabilityZonesForLoadBalancer  func(ctx context.Context, input *elb.EnableAvailabilityZonesForLoadBalancerInput, opts []func(*elb.Options)) (*elb.EnableAvailabilityZonesForLoadBalancerOutput, error)
	MockDisableAvailabilityZonesForLoadBalancer func(ctx context.Context, input *elb.DisableAvailabilityZonesForLoadBalancerInput, opts []func(*elb.Options)) (*elb.DisableAvailabilityZonesForLoadBalancerOutput, error)
	MockDetachLoadBalancerFromSubnets           func(ctx context.Context, input *elb.DetachLoadBalancerFromSubnetsInput, opts []func(*elb.Options)) (*elb.DetachLoadBalancerFromSubnetsOutput, error)
	MockAttachLoadBalancerToSubnets             func(ctx context.Context, input *elb.AttachLoadBalancerToSubnetsInput, opts []func(*elb.Options)) (*elb.AttachLoadBalancerToSubnetsOutput, error)
	MockApplySecurityGroupsToLoadBalancer       func(ctx context.Context, input *elb.ApplySecurityGroupsToLoadBalancerInput, opts []func(*elb.Options)) (*elb.ApplySecurityGroupsToLoadBalancerOutput, error)
	MockCreateLoadBalancerListeners             func(ctx context.Context, input *elb.CreateLoadBalancerListenersInput, opts []func(*elb.Options)) (*elb.CreateLoadBalancerListenersOutput, error)
	MockDeleteLoadBalancerListeners             func(ctx context.Context, input *elb.DeleteLoadBalancerListenersInput, opts []func(*elb.Options)) (*elb.DeleteLoadBalancerListenersOutput, error)
	MockRegisterInstancesWithLoadBalancer       func(ctx context.Context, input *elb.RegisterInstancesWithLoadBalancerInput, opts []func(*elb.Options)) (*elb.RegisterInstancesWithLoadBalancerOutput, error)
	MockDeregisterInstancesFromLoadBalancer     func(ctx context.Context, input *elb.DeregisterInstancesFromLoadBalancerInput, opts []func(*elb.Options)) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	MockDescribeTags                            func(ctx context.Context, input *elb.DescribeTagsInput, opts []func(*elb.Options)) (*elb.DescribeTagsOutput, error)
	MockAddTags                                 func(ctx context.Context, input *elb.AddTagsInput, opts []func(*elb.Options)) (*elb.AddTagsOutput, error)
	MockRemoveTags                              func(ctx context.Context, input *elb.RemoveTagsInput, opts []func(*elb.Options)) (*elb.RemoveTagsOutput, error)
	MockConfigureHealthCheck                    func(ctx context.Context, params *elb.ConfigureHealthCheckInput, opts []func(*elb.Options)) (*elb.ConfigureHealthCheckOutput, error)
}

// DescribeLoadBalancers calls the underlying
// MockDescribeLoadBalancers method.
func (c *MockClient) DescribeLoadBalancers(ctx context.Context, i *elasticloadbalancing.DescribeLoadBalancersInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
	return c.MockDescribeLoadBalancers(ctx, i, opts)
}

// CreateLoadBalancer calls the underlying
// MockCreateLoadBalancer method.
func (c *MockClient) CreateLoadBalancer(ctx context.Context, i *elasticloadbalancing.CreateLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.CreateLoadBalancerOutput, error) {
	return c.MockCreateLoadBalancer(ctx, i, opts)
}

// DeleteLoadBalancer calls the underlying
// MockDeleteLoadBalancer method.
func (c *MockClient) DeleteLoadBalancer(ctx context.Context, i *elasticloadbalancing.DeleteLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DeleteLoadBalancerOutput, error) {
	return c.MockDeleteLoadBalancer(ctx, i, opts)
}

// EnableAvailabilityZonesForLoadBalancer calls the underlying
// MockEnableAvailabilityZonesForLoadBalancer method.
func (c *MockClient) EnableAvailabilityZonesForLoadBalancer(ctx context.Context, i *elasticloadbalancing.EnableAvailabilityZonesForLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.EnableAvailabilityZonesForLoadBalancerOutput, error) {
	return c.MockEnableAvailabilityZonesForLoadBalancer(ctx, i, opts)
}

// DisableAvailabilityZonesForLoadBalancer calls the underlying
// MockDisableAvailabilityZonesForLoadBalancer method.
func (c *MockClient) DisableAvailabilityZonesForLoadBalancer(ctx context.Context, i *elasticloadbalancing.DisableAvailabilityZonesForLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DisableAvailabilityZonesForLoadBalancerOutput, error) {
	return c.MockDisableAvailabilityZonesForLoadBalancer(ctx, i, opts)
}

// DetachLoadBalancerFromSubnets calls the underlying
// MockDetachLoadBalancerFromSubnets method.
func (c *MockClient) DetachLoadBalancerFromSubnets(ctx context.Context, i *elasticloadbalancing.DetachLoadBalancerFromSubnetsInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DetachLoadBalancerFromSubnetsOutput, error) {
	return c.MockDetachLoadBalancerFromSubnets(ctx, i, opts)
}

// AttachLoadBalancerToSubnets calls the underlying
// MockAttachLoadBalancerToSubnets method.
func (c *MockClient) AttachLoadBalancerToSubnets(ctx context.Context, i *elasticloadbalancing.AttachLoadBalancerToSubnetsInput, opts ...func(*elb.Options)) (*elasticloadbalancing.AttachLoadBalancerToSubnetsOutput, error) {
	return c.MockAttachLoadBalancerToSubnets(ctx, i, opts)
}

// ApplySecurityGroupsToLoadBalancer calls the underlying
// MockApplySecurityGroupsToLoadBalancer method.
func (c *MockClient) ApplySecurityGroupsToLoadBalancer(ctx context.Context, i *elasticloadbalancing.ApplySecurityGroupsToLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.ApplySecurityGroupsToLoadBalancerOutput, error) {
	return c.MockApplySecurityGroupsToLoadBalancer(ctx, i, opts)
}

// CreateLoadBalancerListeners calls the underlying
// MockCreateLoadBalancerListeners method.
func (c *MockClient) CreateLoadBalancerListeners(ctx context.Context, i *elasticloadbalancing.CreateLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elasticloadbalancing.CreateLoadBalancerListenersOutput, error) {
	return c.MockCreateLoadBalancerListeners(ctx, i, opts)
}

// DeleteLoadBalancerListeners calls the underlying
// MockDeleteLoadBalancerListeners method.
func (c *MockClient) DeleteLoadBalancerListeners(ctx context.Context, i *elasticloadbalancing.DeleteLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DeleteLoadBalancerListenersOutput, error) {
	return c.MockDeleteLoadBalancerListeners(ctx, i, opts)
}

// RegisterInstancesWithLoadBalancer calls the underlying
// MockRegisterInstancesWithLoadBalancer method.
func (c *MockClient) RegisterInstancesWithLoadBalancer(ctx context.Context, i *elasticloadbalancing.RegisterInstancesWithLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.RegisterInstancesWithLoadBalancerOutput, error) {
	return c.MockRegisterInstancesWithLoadBalancer(ctx, i, opts)
}

// DeregisterInstancesFromLoadBalancer calls the underlying
// MockDeregisterInstancesWithLoadBalancer method.
func (c *MockClient) DeregisterInstancesFromLoadBalancer(ctx context.Context, i *elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DeregisterInstancesFromLoadBalancerOutput, error) {
	return c.MockDeregisterInstancesFromLoadBalancer(ctx, i, opts)
}

// DescribeTags calls the underlying
// MockDescribeTags method.
func (c *MockClient) DescribeTags(ctx context.Context, i *elasticloadbalancing.DescribeTagsInput, opts ...func(*elb.Options)) (*elasticloadbalancing.DescribeTagsOutput, error) {
	return c.MockDescribeTags(ctx, i, opts)
}

// AddTags calls the underlying
// MockAddTags method.
func (c *MockClient) AddTags(ctx context.Context, i *elasticloadbalancing.AddTagsInput, opts ...func(*elb.Options)) (*elasticloadbalancing.AddTagsOutput, error) {
	return c.MockAddTags(ctx, i, opts)
}

// RemoveTags calls the underlying
// MockRemoveTags method.
func (c *MockClient) RemoveTags(ctx context.Context, i *elasticloadbalancing.RemoveTagsInput, opts ...func(*elb.Options)) (*elasticloadbalancing.RemoveTagsOutput, error) {
	return c.MockRemoveTags(ctx, i, opts)
}

// ConfigureHealthCheck calls the underlying
// MockConfigureHealthCheck method.
func (c *MockClient) ConfigureHealthCheck(ctx context.Context, i *elasticloadbalancing.ConfigureHealthCheckInput, opts ...func(*elb.Options)) (*elasticloadbalancing.ConfigureHealthCheckOutput, error) {
	return c.MockConfigureHealthCheck(ctx, i, opts)
}
