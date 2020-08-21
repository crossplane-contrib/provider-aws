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

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	"github.com/crossplane/provider-aws/apis/s3/v1beta1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this S3BucketPolicy
func (mg *S3BucketPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)
	// Resolve spec.BucketName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.PolicyBody.BucketName),
		Reference:    mg.Spec.PolicyBody.BucketNameRef,
		Selector:     mg.Spec.PolicyBody.BucketNameSelector,
		To:           reference.To{Managed: &v1beta1.Bucket{}, List: &v1beta1.BucketList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.PolicyBody.BucketName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.PolicyBody.BucketNameRef = rsp.ResolvedReference

	// Resolve spec.UserName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.PolicyBody.UserName),
		Reference:    mg.Spec.PolicyBody.UserNameRef,
		Selector:     mg.Spec.PolicyBody.UserNameSelector,
		To:           reference.To{Managed: &v1alpha1.IAMUser{}, List: &v1alpha1.IAMUserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.PolicyBody.UserName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.PolicyBody.UserNameRef = rsp.ResolvedReference

	return nil
}
