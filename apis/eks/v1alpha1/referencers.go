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

package v1alpha1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2v1beta1 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	eksv1beta1 "github.com/crossplane/provider-aws/apis/eks/v1beta1"
	iamv1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
)

// ResolveReferences of this NodeGroup
func (mg *NodeGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.clusterName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ClusterName,
		Reference:    mg.Spec.ForProvider.ClusterNameRef,
		Selector:     mg.Spec.ForProvider.ClusterNameSelector,
		To:           reference.To{Managed: &eksv1beta1.Cluster{}, List: &eksv1beta1.ClusterList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.ClusterName = rsp.ResolvedValue
	mg.Spec.ForProvider.ClusterNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.nodeRole
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.NodeRole,
		Reference:    mg.Spec.ForProvider.NodeRoleRef,
		Selector:     mg.Spec.ForProvider.NodeRoleSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
		Extract:      iamv1beta1.IAMRoleARN(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.NodeRole = rsp.ResolvedValue
	mg.Spec.ForProvider.NodeRoleRef = rsp.ResolvedReference

	// Resolve spec.forProvider.subnets
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.Subnets,
		References:    mg.Spec.ForProvider.SubnetRefs,
		Selector:      mg.Spec.ForProvider.SubnetSelector,
		To:            reference.To{Managed: &ec2v1beta1.Subnet{}, List: &ec2v1beta1.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.Subnets = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.remoteAccess.sourceSecurityGroups
	if mg.Spec.ForProvider.RemoteAccess != nil {
		mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
			CurrentValues: mg.Spec.ForProvider.RemoteAccess.SourceSecurityGroups,
			References:    mg.Spec.ForProvider.RemoteAccess.SourceSecurityGroupRefs,
			Selector:      mg.Spec.ForProvider.RemoteAccess.SourceSecurityGroupSelector,
			To:            reference.To{Managed: &ec2v1beta1.SecurityGroup{}, List: &ec2v1beta1.SecurityGroupList{}},
			Extract:       reference.ExternalName(),
		})
		if err != nil {
			return err
		}
		mg.Spec.ForProvider.RemoteAccess.SourceSecurityGroups = mrsp.ResolvedValues
		mg.Spec.ForProvider.RemoteAccess.SourceSecurityGroupRefs = mrsp.ResolvedReferences
	}

	return nil
}
