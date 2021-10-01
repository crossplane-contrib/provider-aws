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

package manualv1alpha1

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences of this VPCCIDRBlock
func (mg *VPCCIDRBlock) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.vpcId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.VPCID),
		Reference:    mg.Spec.ForProvider.VPCIDRef,
		Selector:     mg.Spec.ForProvider.VPCIDSelector,
		To:           reference.To{Managed: &v1beta1.VPC{}, List: &v1beta1.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.vpcId")
	}
	mg.Spec.ForProvider.VPCID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.VPCIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this VPCCIDRBlock
func (mg *Instance) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.securityGroups
	rg, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroups,
		References:    mg.Spec.ForProvider.SecurityGroupRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupSelector,
		To:            reference.To{Managed: &v1beta1.SecurityGroup{}, List: &v1beta1.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroups")
	}
	mg.Spec.ForProvider.SecurityGroups = rg.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupRefs = rg.ResolvedReferences

	// Resolve spec.forProvider.subnetId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: aws.StringValue(mg.Spec.ForProvider.SubnetID),
		Reference:    mg.Spec.ForProvider.SubnetIDRef,
		Selector:     mg.Spec.ForProvider.SubnetIDSelector,
		To:           reference.To{Managed: &v1beta1.Subnet{}, List: &v1beta1.SubnetList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrapf(err, "spec.forProvider.subnetId")
	}
	mg.Spec.ForProvider.SubnetID = aws.String(rsp.ResolvedValue)
	mg.Spec.ForProvider.SubnetIDRef = rsp.ResolvedReference

	return nil
}
