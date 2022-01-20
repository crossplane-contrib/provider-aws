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

// Code generated by angryjet. DO NOT EDIT.

package manualv1alpha1

import (
	"context"
	reference "github.com/crossplane/crossplane-runtime/pkg/reference"
	v1beta1 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	v1alpha1 "github.com/crossplane/provider-aws/apis/kms/v1alpha1"
	errors "github.com/pkg/errors"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this Instance.
func (mg *Instance) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var mrsp reference.MultiResolutionResponse
	var err error

	for i3 := 0; i3 < len(mg.Spec.ForProvider.BlockDeviceMappings); i3++ {
		if mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS != nil {
			rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
				CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KmsKeyID),
				Extract:      reference.ExternalName(),
				Reference:    mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KMSKeyIDRef,
				Selector:     mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KMSKeyIDSelector,
				To: reference.To{
					List:    &v1alpha1.KeyList{},
					Managed: &v1alpha1.Key{},
				},
			})
			if err != nil {
				return errors.Wrap(err, "mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KmsKeyID")
			}
			mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KmsKeyID = reference.ToPtrValue(rsp.ResolvedValue)
			mg.Spec.ForProvider.BlockDeviceMappings[i3].EBS.KMSKeyIDRef = rsp.ResolvedReference

		}
	}
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		Extract:       reference.ExternalName(),
		References:    mg.Spec.ForProvider.SecurityGroupRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupSelector,
		To: reference.To{
			List:    &v1beta1.SecurityGroupList{},
			Managed: &v1beta1.SecurityGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.SecurityGroupIDs")
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupRefs = mrsp.ResolvedReferences

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SubnetID),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.SubnetIDRef,
		Selector:     mg.Spec.ForProvider.SubnetIDSelector,
		To: reference.To{
			List:    &v1beta1.SubnetList{},
			Managed: &v1beta1.Subnet{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.SubnetID")
	}
	mg.Spec.ForProvider.SubnetID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SubnetIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this VPCPeeringConnection.
func (mg *VPCPeeringConnection) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCID),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCIDRef,
		Selector:     mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCIDSelector,
		To: reference.To{
			List:    &v1beta1.VPCList{},
			Managed: &v1beta1.VPC{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCID")
	}
	mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomVPCPeeringConnectionParameters.VPCIDRef = rsp.ResolvedReference

	return nil
}
