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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences of this Route53ResolverEndpoint
func (mg *ResolverEndpoint) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)
	// Resolve spec.forProvider.securityGroupIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupIDSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroupIds")
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.ipAddresses[].subNetIds
	for i := range mg.Spec.ForProvider.IPAddresses {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IPAddresses[i].SubnetID),
			Reference:    mg.Spec.ForProvider.IPAddresses[i].SubnetIDRef,
			Selector:     mg.Spec.ForProvider.IPAddresses[i].SubnetIDSelector,
			To:           reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("spec.forProvider.ipAddresses[%d].subNetIds", i))
		}
		mg.Spec.ForProvider.IPAddresses[i].SubnetID = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.IPAddresses[i].SubnetIDRef = rsp.ResolvedReference
	}
	return nil
}

// ResolveReferences of this Route53ResolverRule
func (mg *ResolverRule) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.resolverEndpointId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResolverEndpointID),
		Reference:    mg.Spec.ForProvider.ResolverEndpointIDRef,
		Selector:     mg.Spec.ForProvider.ResolverEndpointIDSelector,
		To:           reference.To{Managed: &ResolverEndpoint{}, List: &ResolverEndpointList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.resolverEndpointId")
	}
	mg.Spec.ForProvider.ResolverEndpointID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResolverEndpointIDRef = rsp.ResolvedReference

	return nil
}
