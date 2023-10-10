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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	kms "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// ResolveReferences of this DBCluster
func (mg *DBCluster) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.dbClusterParameterGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DBClusterParameterGroupName),
		Reference:    mg.Spec.ForProvider.DBClusterParameterGroupNameRef,
		Selector:     mg.Spec.ForProvider.DBClusterParameterGroupNameSelector,
		To:           reference.To{Managed: &DBClusterParameterGroup{}, List: &DBClusterParameterGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbClusterParameterGroupName")
	}
	mg.Spec.ForProvider.DBClusterParameterGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DBClusterParameterGroupNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.dbSubnetGroupName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
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

	// Resolve spec.forProvider.kmsKeyID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyID),
		Reference:    mg.Spec.ForProvider.KMSKeyIDRef,
		Selector:     mg.Spec.ForProvider.KMSKeyIDSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyID")
	}
	mg.Spec.ForProvider.KMSKeyID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.kmsKeyID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyID),
		Reference:    mg.Spec.ForProvider.KMSKeyIDRef,
		Selector:     mg.Spec.ForProvider.KMSKeyIDSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyID")
	}
	mg.Spec.ForProvider.KMSKeyID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.vpcSecurityGroupIDs
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.VPCSecurityGroupIDs),
		References:    mg.Spec.ForProvider.VPCSecurityGroupIDsRefs,
		Selector:      mg.Spec.ForProvider.VPCSecurityGroupIDsSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.vpcSecurityGroupIDs")
	}
	mg.Spec.ForProvider.VPCSecurityGroupIDs = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.VPCSecurityGroupIDsRefs = mrsp.ResolvedReferences

	return nil
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

	return nil
}

// ResolveReferences of this DBSubnetGroup
func (mg *DBSubnetGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.dbSubnetIDs
	rsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.SubnetIDs),
		References:    mg.Spec.ForProvider.SubnetIDsRefs,
		Selector:      mg.Spec.ForProvider.SUbnetIDsSelector,
		To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.dbSubnetIDs")
	}
	mg.Spec.ForProvider.SubnetIDs = reference.ToPtrValues(rsp.ResolvedValues)
	mg.Spec.ForProvider.SubnetIDsRefs = rsp.ResolvedReferences

	return nil
}
