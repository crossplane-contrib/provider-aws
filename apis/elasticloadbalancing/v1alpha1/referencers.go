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
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences of this ELB
func (mg *ELB) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.subnetIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SubnetIDs,
		References:    mg.Spec.ForProvider.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.SubnetIDSelector,
		To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.subnetIds")
	}
	mg.Spec.ForProvider.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.securityGroupIds
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupIDSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroupIds")
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupIDRefs = mrsp.ResolvedReferences

	return nil
}

// ResolveReferences of this ELBAttachment
func (mg *ELBAttachment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.elbName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ELBName,
		Reference:    mg.Spec.ForProvider.ELBNameRef,
		Selector:     mg.Spec.ForProvider.ELBNameSelector,
		To:           reference.To{Managed: &ELB{}, List: &ELBList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.elbName")
	}
	mg.Spec.ForProvider.ELBName = rsp.ResolvedValue
	mg.Spec.ForProvider.ELBNameRef = rsp.ResolvedReference

	return nil
}
