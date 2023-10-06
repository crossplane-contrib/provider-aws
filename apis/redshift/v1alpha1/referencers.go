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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	network "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

// ResolveReferences of this Cluster
func (mg *Cluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.vpcSecurityGroupIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.VPCSecurityGroupIDs,
		References:    mg.Spec.ForProvider.VPCSecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.VPCSecurityGroupIDSelector,
		To:            reference.To{Managed: &network.SecurityGroup{}, List: &network.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.vpcSecurityGroupIds")
	}
	mg.Spec.ForProvider.VPCSecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.VPCSecurityGroupIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.IAMRoles
	mnrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.IAMRoles,
		References:    mg.Spec.ForProvider.IAMRoleRefs,
		Selector:      mg.Spec.ForProvider.IAMRoleSelector,
		To:            reference.To{Managed: &v1beta1.Role{}, List: &v1beta1.RoleList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.IAMRoles")
	}
	mg.Spec.ForProvider.IAMRoles = mnrsp.ResolvedValues
	mg.Spec.ForProvider.IAMRoleRefs = mnrsp.ResolvedReferences

	// Resolve spec.forProvider.clusterSecurityGroups
	msgrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.ClusterSecurityGroups,
		References:    mg.Spec.ForProvider.ClusterSecurityGroupRefs,
		Selector:      mg.Spec.ForProvider.ClusterSecurityGroupSelector,
		To:            reference.To{Managed: &network.SecurityGroup{}, List: &network.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.clusterSecurityGroups")
	}
	mg.Spec.ForProvider.ClusterSecurityGroups = msgrsp.ResolvedValues
	mg.Spec.ForProvider.ClusterSecurityGroupRefs = msgrsp.ResolvedReferences

	return nil
}
