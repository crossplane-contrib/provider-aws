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

	acm "github.com/crossplane/provider-aws/apis/acm/v1alpha1"
	ec2 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	iamv1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this User
func (mg *User) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.serverID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ServerID),
		Reference:    mg.Spec.ForProvider.ServerIDRef,
		Selector:     mg.Spec.ForProvider.ServerIDSelector,
		To:           reference.To{Managed: &Server{}, List: &ServerList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serverID")
	}
	mg.Spec.ForProvider.ServerID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ServerIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.role
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Role),
		Reference:    mg.Spec.ForProvider.RoleRef,
		Selector:     mg.Spec.ForProvider.RoleSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
		Extract:      iamv1beta1.IAMRoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.role")
	}
	mg.Spec.ForProvider.Role = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RoleRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Server
func (mg *Server) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.endpointDetails.subnetIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDs),
		References:    mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDSelector,
		To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.endpointDetails.subnetIDs")
	}
	mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDs = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.endpointDetails.securityGroupIds
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs),
		References:    mg.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.endpointDetails.securityGroupIDs")
	}
	mg.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.endpointDetails.vpcID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomEndpointDetails.VPCID),
		Reference:    mg.Spec.ForProvider.CustomEndpointDetails.VPCIDRef,
		Selector:     mg.Spec.ForProvider.CustomEndpointDetails.SubnetIDSelector,
		To:           reference.To{Managed: &ec2.VPC{}, List: &ec2.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serverID")
	}
	mg.Spec.ForProvider.CustomEndpointDetails.VPCID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomEndpointDetails.VPCIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.Certificate
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Certificate),
		Reference:    mg.Spec.ForProvider.CertificateRef,
		Selector:     mg.Spec.ForProvider.CertificateSelector,
		To:           reference.To{Managed: &acm.Certificate{}, List: &acm.CertificateList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.certificate")
	}
	mg.Spec.ForProvider.Certificate = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CertificateRef = rsp.ResolvedReference

	// Resolve spec.forProvider.loggingRole
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.LoggingRole),
		Reference:    mg.Spec.ForProvider.LoggingRoleRef,
		Selector:     mg.Spec.ForProvider.LoggingRoleSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.loggingRole")
	}
	mg.Spec.ForProvider.LoggingRole = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.LoggingRoleRef = rsp.ResolvedReference

	return nil
}
