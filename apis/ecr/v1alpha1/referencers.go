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
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	identityv1alpha1 "github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	identityv1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
)

// ResolveReferences of this RepositoryPolicy
func (mg *RepositoryPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.repositoryName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.RepositoryName),
		Reference:    mg.Spec.ForProvider.RepositoryNameRef,
		Selector:     mg.Spec.ForProvider.RepositoryNameSelector,
		To:           reference.To{Managed: &Repository{}, List: &RepositoryList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.repositoryName")
	}
	mg.Spec.ForProvider.RepositoryName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RepositoryNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.userName
	if mg.Spec.ForProvider.Policy != nil {
		for i := range mg.Spec.ForProvider.Policy.Statements {
			statement := mg.Spec.ForProvider.Policy.Statements[i]
			err = ResolvePrincipal(ctx, r, statement.Principal, i)
			if err != nil {
				return err
			}
			err = ResolvePrincipal(ctx, r, statement.NotPrincipal, i)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ResolvePrincipal resolves all the IAMUser and IAMRole references in a RepositoryPrincipal
func ResolvePrincipal(ctx context.Context, r *reference.APIResolver, principal *RepositoryPrincipal, statementIndex int) error {
	if principal == nil {
		return nil
	}
	for i := range principal.AWSPrincipals {

		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(principal.AWSPrincipals[i].IAMUserARN),
			Reference:    principal.AWSPrincipals[i].IAMUserARNRef,
			Selector:     principal.AWSPrincipals[i].IAMUserARNSelector,
			To:           reference.To{Managed: &v1alpha1.IAMUser{}, List: &v1alpha1.IAMUserList{}},
			Extract:      identityv1alpha1.IAMUserARN(),
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("spec.forProvider.statements[%d].principal.awsPrincipals[%d].iamUserArn", statementIndex, i))
		}
		principal.AWSPrincipals[i].IAMUserARN = reference.ToPtrValue(rsp.ResolvedValue)
		principal.AWSPrincipals[i].IAMUserARNRef = rsp.ResolvedReference

		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(principal.AWSPrincipals[i].IAMRoleARN),
			Reference:    principal.AWSPrincipals[i].IAMRoleARNRef,
			Selector:     principal.AWSPrincipals[i].IAMRoleARNSelector,
			To:           reference.To{Managed: &identityv1beta1.IAMRole{}, List: &identityv1beta1.IAMRoleList{}},
			Extract:      identityv1beta1.IAMRoleARN(),
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("spec.forProvider.statement[%d].principal.aws[%d].IAMRoleArn", statementIndex, i))
		}
		principal.AWSPrincipals[i].IAMRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
		principal.AWSPrincipals[i].IAMRoleARNRef = rsp.ResolvedReference

	}
	return nil
}
