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

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// ResolveReferences of this CacheSubnetGroup
func (mg *CacheSubnetGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.subnetIDs
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SubnetIDs,
		References:    mg.Spec.ForProvider.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.SubnetIDSelector,
		To:            reference.To{Managed: &v1beta1.Subnet{}, List: &v1beta1.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetIDRefs = mrsp.ResolvedReferences

	return nil
}

// ResolveReferences of this CacheCluster
func (mg *CacheCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.cacheSubnetGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: aws.StringValue(mg.Spec.ForProvider.CacheSubnetGroupName),
		Reference:    mg.Spec.ForProvider.CacheSubnetGroupNameRef,
		Selector:     mg.Spec.ForProvider.CacheSubnetGroupNameSelector,
		To:           reference.To{Managed: &CacheSubnetGroup{}, List: &CacheSubnetGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.CacheSubnetGroupName = aws.String(rsp.ResolvedValue)
	mg.Spec.ForProvider.CacheSubnetGroupNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.securityGroupIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupIDSelector,
		To:            reference.To{Managed: &v1beta1.SecurityGroup{}, List: &v1beta1.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupIDRefs = mrsp.ResolvedReferences

	return nil
}
