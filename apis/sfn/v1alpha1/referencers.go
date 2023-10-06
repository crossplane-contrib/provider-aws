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

	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

// ResolveReferences of this StateMachine
func (mg *StateMachine) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.roleArn
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.RoleARN),
		Reference:    mg.Spec.ForProvider.RoleARNRef,
		Selector:     mg.Spec.ForProvider.RoleARNSelector,
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      iamv1beta1.RoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.roleArn")
	}
	mg.Spec.ForProvider.RoleARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RoleARNRef = rsp.ResolvedReference
	return nil
}
