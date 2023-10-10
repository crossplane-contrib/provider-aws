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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	database "github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	network "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	kmsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// ResolveReferences of this DBCluster
func (mg *DBCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.domainIAMRoleName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DomainIAMRoleName),
		Reference:    mg.Spec.ForProvider.DomainIAMRoleNameRef,
		Selector:     mg.Spec.ForProvider.DomainIAMRoleNameSelector,
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.domainIAMRoleName")
	}
	mg.Spec.ForProvider.DomainIAMRoleName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DomainIAMRoleNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.kmsKeyID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyID),
		Reference:    mg.Spec.ForProvider.KMSKeyIDRef,
		Selector:     mg.Spec.ForProvider.KMSKeyIDSelector,
		To:           reference.To{Managed: &kmsv1alpha1.Key{}, List: &kmsv1alpha1.KeyList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyID")
	}
	mg.Spec.ForProvider.KMSKeyID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyIDRef = rsp.ResolvedReference

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

	// Resolve spec.forProvider.dbSubnetGroupName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBSubnetGroupName),
		Reference:    mg.Spec.ForProvider.DBSubnetGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBSubnetGroupNameSelector,
		To:           reference.To{Managed: &database.DBSubnetGroup{}, List: &database.DBSubnetGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbSubnetGroupName")
	}
	mg.Spec.ForProvider.DBSubnetGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBSubnetGroupNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBClusterParameterGroupName),
		Reference:    mg.Spec.ForProvider.DBClusterParameterGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBClusterParameterGroupNameSelector,
		To:           reference.To{List: &DBClusterParameterGroupList{}, Managed: &DBClusterParameterGroup{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbClusterParameterGroupName")
	}
	mg.Spec.ForProvider.DBClusterParameterGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBClusterParameterGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this GlobalCluster
func (mg *GlobalCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.sourceDBClusterIdentifier
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SourceDBClusterIdentifier),
		Reference:    mg.Spec.ForProvider.SourceDBClusterIdentifierRef,
		Selector:     mg.Spec.ForProvider.SourceDBClusterIdentifierSelector,
		To:           reference.To{Managed: &DBCluster{}, List: &DBClusterList{}},
		Extract:      DBClusterARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.sourceDBClusterIdentifier")
	}
	mg.Spec.ForProvider.SourceDBClusterIdentifier = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SourceDBClusterIdentifierRef = rsp.ResolvedReference

	return nil
}

// DBClusterARN returns the status.atProvider.ARN of an Role.
func DBClusterARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*DBCluster)
		if !ok {
			return ""
		}
		if r.Status.AtProvider.DBClusterARN == nil {
			return ""
		}
		return *r.Status.AtProvider.DBClusterARN
	}
}

// ResolveReferences of this DBInstance
func (mg *DBInstance) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.dbClusterIdentifier
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBClusterIdentifier),
		Reference:    mg.Spec.ForProvider.DBClusterIdentifierRef,
		Selector:     mg.Spec.ForProvider.DBClusterIdentifierSelector,
		To:           reference.To{Managed: &DBCluster{}, List: &DBClusterList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbClusterIdentifier")
	}
	mg.Spec.ForProvider.DBClusterIdentifier = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBClusterIdentifierRef = rsp.ResolvedReference

	// Resolve spec.forProvider.dbSubnetGroupName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBSubnetGroupName),
		Reference:    mg.Spec.ForProvider.DBSubnetGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBSubnetGroupNameSelector,
		To:           reference.To{Managed: &database.DBSubnetGroup{}, List: &database.DBSubnetGroupList{}},
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
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
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
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      iamv1beta1.RoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.monitoringRoleArn")
	}
	mg.Spec.ForProvider.MonitoringRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.MonitoringRoleARNRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBParameterGroupName),
		Reference:    mg.Spec.ForProvider.DBParameterGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBParameterGroupNameSelector,
		To:           reference.To{List: &DBParameterGroupList{}, Managed: &DBParameterGroup{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbParameterGroupName")
	}
	mg.Spec.ForProvider.DBParameterGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBParameterGroupNameRef = rsp.ResolvedReference

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

	// Resolve spec.forProvider.kmsKeyID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyID),
		Reference:    mg.Spec.ForProvider.KMSKeyIDRef,
		Selector:     mg.Spec.ForProvider.KMSKeyIDSelector,
		To:           reference.To{Managed: &kmsv1alpha1.Key{}, List: &kmsv1alpha1.KeyList{}},
		Extract:      kmsv1alpha1.KMSKeyARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyID")
	}
	mg.Spec.ForProvider.KMSKeyID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyIDRef = rsp.ResolvedReference

	return nil
}

// RDSClusterOrInstance interface to access common fields independent of the type
// See: https://github.com/kubernetes-sigs/controller-tools/issues/471
// +kubebuilder:object:generate=false
type RDSClusterOrInstance interface {
	resource.Managed
	GetMasterUserPasswordSecretRef() *xpv1.SecretKeySelector
}

// GetMasterUserPasswordSecretRef returns the MasterUserPasswordSecretRef
func (mg *DBInstance) GetMasterUserPasswordSecretRef() *xpv1.SecretKeySelector {
	return mg.Spec.ForProvider.MasterUserPasswordSecretRef
}

// GetMasterUserPasswordSecretRef returns the MasterUserPasswordSecretRef
func (mg *DBCluster) GetMasterUserPasswordSecretRef() *xpv1.SecretKeySelector {
	return mg.Spec.ForProvider.MasterUserPasswordSecretRef
}

var _ RDSClusterOrInstance = (*DBInstance)(nil)
var _ RDSClusterOrInstance = (*DBCluster)(nil)
