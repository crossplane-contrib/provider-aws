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

	network "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	kms "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// ResolveReferences of this AccessPoint
func (mg *AccessPoint) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.fileSystemID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.FileSystemID),
		Reference:    mg.Spec.ForProvider.FileSystemIDRef,
		Selector:     mg.Spec.ForProvider.FileSystemIDSelector,
		To:           reference.To{Managed: &FileSystem{}, List: &FileSystemList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.fileSystemID")
	}
	mg.Spec.ForProvider.FileSystemID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.FileSystemIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this FileSystem
func (mg *FileSystem) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.kmsKeyId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyID),
		Reference:    mg.Spec.ForProvider.KMSKeyIDRef,
		Selector:     mg.Spec.ForProvider.KMSKeyIDSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyId")
	}
	mg.Spec.ForProvider.KMSKeyID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this MountTarget
func (mg *MountTarget) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.securityGroups
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroups,
		References:    mg.Spec.ForProvider.SecurityGroupsRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupsSelector,
		To:            reference.To{Managed: &network.SecurityGroup{}, List: &network.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroups")
	}
	mg.Spec.ForProvider.SecurityGroups = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupsRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.subnetID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SubnetID),
		Reference:    mg.Spec.ForProvider.SubnetIDRef,
		Selector:     mg.Spec.ForProvider.SubnetIDSelector,
		To:           reference.To{Managed: &network.Subnet{}, List: &network.SubnetList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.subnetID")
	}
	mg.Spec.ForProvider.SubnetID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SubnetIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.fileSystemID
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.FileSystemID),
		Reference:    mg.Spec.ForProvider.FileSystemIDRef,
		Selector:     mg.Spec.ForProvider.FileSystemIDSelector,
		To:           reference.To{Managed: &FileSystem{}, List: &FileSystemList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.fileSystemID")
	}
	mg.Spec.ForProvider.FileSystemID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.FileSystemIDRef = rsp.ResolvedReference

	return nil
}
