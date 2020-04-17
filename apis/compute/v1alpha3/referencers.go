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

package v1alpha3

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

// ResolveReferences of this EKSCluster
func (mg *EKSCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.roleARN
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.RoleARN,
		Reference:    mg.Spec.RoleARNRef,
		Selector:     mg.Spec.RoleARNSelector,
		To:           reference.To{Managed: &v1beta1.IAMRole{}, List: &v1beta1.IAMRoleList{}},
		Extract:      v1beta1.IAMRoleARN(),
	})
	if err != nil {
		return err
	}
	mg.Spec.RoleARN = rsp.ResolvedValue
	mg.Spec.RoleARNRef = rsp.ResolvedReference

	// Resolve spec.vpcID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.VPCID,
		Reference:    mg.Spec.VPCIDRef,
		Selector:     mg.Spec.VPCIDSelector,
		To:           reference.To{Managed: &v1alpha3.VPC{}, List: &v1alpha3.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.VPCID = rsp.ResolvedValue
	mg.Spec.VPCIDRef = rsp.ResolvedReference

	// Resolve spec.clusterControlPlaneSecurityGroup
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.WorkerNodes.ClusterControlPlaneSecurityGroup,
		Reference:    mg.Spec.WorkerNodes.ClusterControlPlaneSecurityGroupRef,
		Selector:     mg.Spec.WorkerNodes.ClusterControlPlaneSecurityGroupSelector,
		To:           reference.To{Managed: &v1alpha3.SecurityGroup{}, List: &v1alpha3.SecurityGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.WorkerNodes.ClusterControlPlaneSecurityGroup = rsp.ResolvedValue
	mg.Spec.WorkerNodes.ClusterControlPlaneSecurityGroupRef = rsp.ResolvedReference

	// Resolve spec.subnetIDs
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.SubnetIDs,
		References:    mg.Spec.SubnetIDRefs,
		Selector:      mg.Spec.SubnetIDSelector,
		To:            reference.To{Managed: &v1alpha3.Subnet{}, List: &v1alpha3.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.SubnetIDRefs = mrsp.ResolvedReferences

	// Resolve spec.securityGroupIDs
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.SecurityGroupIDs,
		References:    mg.Spec.SecurityGroupIDRefs,
		Selector:      mg.Spec.SecurityGroupIDSelector,
		To:            reference.To{Managed: &v1alpha3.SecurityGroup{}, List: &v1alpha3.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.SecurityGroupIDRefs = mrsp.ResolvedReferences

	return nil
}
