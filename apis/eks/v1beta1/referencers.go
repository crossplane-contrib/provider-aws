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

package v1beta1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

// ResolveReferences of this Cluster
func (mg *Cluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.arnRole
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.RoleArn),
		Reference:    mg.Spec.ForProvider.RoleArnRef,
		Selector:     mg.Spec.ForProvider.RoleArnSelector,
		To:           reference.To{Managed: &v1beta1.IAMRole{}, List: &v1beta1.IAMRoleList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.RoleArn = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RoleArnRef = rsp.ResolvedReference

	// Resolve spec.forProvider.resourcesVpcConfig.subnetIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.ResourcesVpcConfig.SubnetIDs,
		References:    mg.Spec.ForProvider.ResourcesVpcConfig.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.ResourcesVpcConfig.SubnetIDSelector,
		To:            reference.To{Managed: &v1alpha3.Subnet{}, List: &v1alpha3.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.ResourcesVpcConfig.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.ResourcesVpcConfig.SubnetIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.resourcesVpcConfig.securityGroupIds
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.ResourcesVpcConfig.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.ResourcesVpcConfig.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.ResourcesVpcConfig.SecurityGroupIDSelector,
		To:            reference.To{Managed: &v1alpha3.SecurityGroup{}, List: &v1alpha3.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.ResourcesVpcConfig.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.ResourcesVpcConfig.SecurityGroupIDRefs = mrsp.ResolvedReferences

	return nil
}
