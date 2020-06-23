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

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IAMRoleARN returns the status.atProvider.ARN of an IAMRole.
func IAMRoleARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*IAMRole)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN

	}
}

// ResolveReferences of this IAMRolePolicyAttachment
func (mg *IAMRolePolicyAttachment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.roleName
	iamRole, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.RoleName,
		Reference:    mg.Spec.ForProvider.RoleNameRef,
		Selector:     mg.Spec.ForProvider.RoleNameSelector,
		To:           reference.To{Managed: &IAMRole{}, List: &IAMRoleList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.RoleName = iamRole.ResolvedValue
	mg.Spec.ForProvider.RoleNameRef = iamRole.ResolvedReference

	// Resolve spec.forProvider.policyArn
	iamPolicy, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.PolicyARN,
		Reference:    mg.Spec.ForProvider.PolicyARNRef,
		Selector:     mg.Spec.ForProvider.PolicyARNSelector,
		To:           reference.To{Managed: &v1alpha1.IAMPolicy{}, List: &v1alpha1.IAMPolicyList{}},
		Extract:      v1alpha1.IAMPolicyARN(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.PolicyARN = iamPolicy.ResolvedValue
	mg.Spec.ForProvider.PolicyARNRef = iamPolicy.ResolvedReference

	return nil
}
