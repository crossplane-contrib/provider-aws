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

package v1alpha1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

// IAMPolicyARN returns a function that returns the ARN of the given policy.
func IAMPolicyARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*IAMPolicy)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}

// IAMUserARN returns a function that returns the ARN of the given policy.
func IAMUserARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*IAMUser)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}

// ResolveReferences of this IAMUserPolicyAttachment
func (mg *IAMUserPolicyAttachment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.userName
	user, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.UserName,
		Reference:    mg.Spec.ForProvider.UserNameRef,
		Selector:     mg.Spec.ForProvider.UserNameSelector,
		To:           reference.To{Managed: &IAMUser{}, List: &IAMUserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.userName")
	}
	mg.Spec.ForProvider.UserName = user.ResolvedValue
	mg.Spec.ForProvider.UserNameRef = user.ResolvedReference

	// Resolve spec.forProvider.policyArn
	policy, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.PolicyARN,
		Reference:    mg.Spec.ForProvider.PolicyARNRef,
		Selector:     mg.Spec.ForProvider.PolicyARNSelector,
		To:           reference.To{Managed: &IAMPolicy{}, List: &IAMPolicyList{}},
		Extract:      IAMPolicyARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.policyArn")
	}
	mg.Spec.ForProvider.PolicyARN = policy.ResolvedValue
	mg.Spec.ForProvider.PolicyARNRef = policy.ResolvedReference

	return nil
}

// ResolveReferences of this IAMGroupUserMembership
func (mg *IAMGroupUserMembership) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.userName
	user, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.UserName,
		Reference:    mg.Spec.ForProvider.UserNameRef,
		Selector:     mg.Spec.ForProvider.UserNameSelector,
		To:           reference.To{Managed: &IAMUser{}, List: &IAMUserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.userName")
	}
	mg.Spec.ForProvider.UserName = user.ResolvedValue
	mg.Spec.ForProvider.UserNameRef = user.ResolvedReference

	// Resolve spec.forProvider.groupName
	group, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.GroupName,
		Reference:    mg.Spec.ForProvider.GroupNameRef,
		Selector:     mg.Spec.ForProvider.GroupNameSelector,
		To:           reference.To{Managed: &IAMGroup{}, List: &IAMGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupName")
	}
	mg.Spec.ForProvider.GroupName = group.ResolvedValue
	mg.Spec.ForProvider.GroupNameRef = group.ResolvedReference

	return nil
}

// ResolveReferences of this IAMGroupPolicyAttachment
func (mg *IAMGroupPolicyAttachment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.groupName
	group, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.GroupName,
		Reference:    mg.Spec.ForProvider.GroupNameRef,
		Selector:     mg.Spec.ForProvider.GroupNameSelector,
		To:           reference.To{Managed: &IAMGroup{}, List: &IAMGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.groupName")
	}
	mg.Spec.ForProvider.GroupName = group.ResolvedValue
	mg.Spec.ForProvider.GroupNameRef = group.ResolvedReference

	// Resolve spec.forProvider.policyArn
	policy, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.PolicyARN,
		Reference:    mg.Spec.ForProvider.PolicyARNRef,
		Selector:     mg.Spec.ForProvider.PolicyARNSelector,
		To:           reference.To{Managed: &IAMPolicy{}, List: &IAMPolicyList{}},
		Extract:      IAMPolicyARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.policyArn")
	}
	mg.Spec.ForProvider.PolicyARN = policy.ResolvedValue
	mg.Spec.ForProvider.PolicyARNRef = policy.ResolvedReference

	return nil
}

// ResolveReferences of this IAMAccessKey
func (mg *IAMAccessKey) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.userName
	user, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.IAMUsername,
		Reference:    mg.Spec.ForProvider.IAMUsernameRef,
		Selector:     mg.Spec.ForProvider.IAMUsernameSelector,
		To:           reference.To{Managed: &IAMUser{}, List: &IAMUserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.userName")
	}
	mg.Spec.ForProvider.IAMUsername = user.ResolvedValue
	mg.Spec.ForProvider.IAMUsernameRef = user.ResolvedReference

	return nil
}
