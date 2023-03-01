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

package v1alpha3

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
)

// ResolveReferences of this BucketPolicy
func (mg *BucketPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)
	// Resolve spec.forProvider.bucketName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.Parameters.BucketName),
		Reference:    mg.Spec.Parameters.BucketNameRef,
		Selector:     mg.Spec.Parameters.BucketNameSelector,
		To:           reference.To{Managed: &v1beta1.Bucket{}, List: &v1beta1.BucketList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.bucketName")
	}
	mg.Spec.Parameters.BucketName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.Parameters.BucketNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.userName
	if mg.Spec.Parameters.Policy != nil && mg.Spec.Parameters.Policy.Statements != nil {
		for i := range mg.Spec.Parameters.Policy.Statements {
			statement := mg.Spec.Parameters.Policy.Statements[i]
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

// ResolvePrincipal resolves all the User and Role references in a BucketPrincipal
func ResolvePrincipal(ctx context.Context, r *reference.APIResolver, principal *common.BucketPrincipal, statementIndex int) error {
	if principal == nil {
		return nil
	}
	for i := range principal.AWSPrincipals {
		if principal.AWSPrincipals[i].UserARNRef != nil || principal.AWSPrincipals[i].UserARNSelector != nil {
			rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
				CurrentValue: reference.FromPtrValue(principal.AWSPrincipals[i].UserARN),
				Reference:    principal.AWSPrincipals[i].UserARNRef,
				Selector:     principal.AWSPrincipals[i].UserARNSelector,
				To:           reference.To{Managed: &iamv1beta1.User{}, List: &iamv1beta1.UserList{}},
				Extract:      iamv1beta1.UserARN(),
			})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("spec.forProvider.statement[%d].principal.aws[%d].UserARN", statementIndex, i))
			}
			principal.AWSPrincipals[i].UserARN = reference.ToPtrValue(rsp.ResolvedValue)
			principal.AWSPrincipals[i].UserARNRef = rsp.ResolvedReference
		}

		if principal.AWSPrincipals[i].IAMRoleARNRef != nil || principal.AWSPrincipals[i].IAMRoleARNSelector != nil {
			rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
				CurrentValue: reference.FromPtrValue(principal.AWSPrincipals[i].IAMRoleARN),
				Reference:    principal.AWSPrincipals[i].IAMRoleARNRef,
				Selector:     principal.AWSPrincipals[i].IAMRoleARNSelector,
				To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
				Extract:      iamv1beta1.RoleARN(),
			})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("spec.forProvider.statement[%d].principal.aws[%d].IAMRoleArn", statementIndex, i))
			}
			principal.AWSPrincipals[i].IAMRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
			principal.AWSPrincipals[i].IAMRoleARNRef = rsp.ResolvedReference
		}
	}
	return nil
}
