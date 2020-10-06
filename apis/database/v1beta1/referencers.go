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

package v1beta1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	network "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
)

// ResolveReferences of this DBSubnetGroup
func (mg *DBSubnetGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.subnetIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SubnetIDs,
		References:    mg.Spec.ForProvider.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.SubnetIDSelector,
		To:            reference.To{Managed: &network.Subnet{}, List: &network.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.subnetIds")
	}
	mg.Spec.ForProvider.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetIDRefs = mrsp.ResolvedReferences

	return nil
}

// ResolveReferences of this RDSInstance
func (mg *RDSInstance) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.dbSubnetGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBSubnetGroupName),
		Reference:    mg.Spec.ForProvider.DBSubnetGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBSubnetGroupNameSelector,
		To:           reference.To{Managed: &DBSubnetGroup{}, List: &DBSubnetGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbSubnetGroupName")
	}
	mg.Spec.ForProvider.DBSubnetGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBSubnetGroupNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.domainIAMRoleName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DomainIAMRoleName),
		Reference:    mg.Spec.ForProvider.DomainIAMRoleNameRef,
		Selector:     mg.Spec.ForProvider.DomainIAMRoleNameSelector,
		To:           reference.To{Managed: &v1beta1.IAMRole{}, List: &v1beta1.IAMRoleList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.domainIAMRoleName")
	}
	mg.Spec.ForProvider.DomainIAMRoleName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DomainIAMRoleNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.monitoringRoleArn
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.MonitoringRoleARN),
		Reference:    mg.Spec.ForProvider.MonitoringRoleARNRef,
		Selector:     mg.Spec.ForProvider.MonitoringRoleARNSelector,
		To:           reference.To{Managed: &v1beta1.IAMRole{}, List: &v1beta1.IAMRoleList{}},
		Extract:      v1beta1.IAMRoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.monitoringRoleArn")
	}
	mg.Spec.ForProvider.MonitoringRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.MonitoringRoleARNRef = rsp.ResolvedReference

	// Resolve spec.forProvider.vpcSecurityGroupIDs
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

	return nil
}
