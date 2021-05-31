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

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	iamv1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
)

// ResolveReferences of this Stage
func (mg *PlatformApplication) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.successFeedbackRoleArn
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SuccessFeedbackRoleARN),
		Reference:    mg.Spec.ForProvider.SuccessFeedbackRoleARNRef,
		Selector:     mg.Spec.ForProvider.SuccessFeedbackRoleARNSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
		Extract:      iamv1beta1.IAMRoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.successFeedbackRoleArn")
	}
	mg.Spec.ForProvider.SuccessFeedbackRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SuccessFeedbackRoleARNRef = rsp.ResolvedReference

	// Resolve spec.forProvider.failureFeedbackRoleARN
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.FailureFeedbackRoleARN),
		Reference:    mg.Spec.ForProvider.FailureFeedbackRoleARNRef,
		Selector:     mg.Spec.ForProvider.FailureFeedbackRoleARNSelector,
		To:           reference.To{Managed: &iamv1beta1.IAMRole{}, List: &iamv1beta1.IAMRoleList{}},
		Extract:      iamv1beta1.IAMRoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.failureFeedbackRoleARN")
	}
	mg.Spec.ForProvider.FailureFeedbackRoleARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.FailureFeedbackRoleARNRef = rsp.ResolvedReference
	return nil
}
