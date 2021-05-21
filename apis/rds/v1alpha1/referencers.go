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

	database "github.com/crossplane/provider-aws/apis/database/v1beta1"
	network "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	iamv1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
	kmsv1alpha1 "github.com/crossplane/provider-aws/apis/kms/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this DBCluster
func (mg *DBCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.domainIAMRoleName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DomainIAMRoleName),
		Reference:    mg.Spec.ForProvider.DomainIAMRoleNameRef,
		Selector:     mg.Spec.ForProvider.DomainIAMRoleNameSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
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

	return nil
}

// ResolveReferences of this GlobalCluster
func (mg *GlobalCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.sourceDBClusterIdentifier
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SourceDBClusterIdentifier),
		Reference:    mg.Spec.ForProvider.SourceDBClusterIDRef,
		Selector:     mg.Spec.ForProvider.SourceDBClusterIDSelector,
		To:           reference.To{Managed: &DBCluster{}, List: &DBClusterList{}},
		Extract:      DBClusterARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.sourceDBClusterIdentifier")
	}
	mg.Spec.ForProvider.SourceDBClusterIdentifier = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SourceDBClusterIDRef = rsp.ResolvedReference

	return nil
}

// DBClusterARN returns the status.atProvider.ARN of an IAMRole.
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
