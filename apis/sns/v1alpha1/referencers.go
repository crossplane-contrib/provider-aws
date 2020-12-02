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

	// Resolve spec.forProvider.eventDeliveryFailure
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.EventDeliveryFailure),
		Reference:    mg.Spec.ForProvider.EventDeliveryFailureRef,
		Selector:     mg.Spec.ForProvider.EventDeliveryFailureSelector,
		To:           reference.To{Managed: &Topic{}, List: &TopicList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.eventDeliveryFailure")
	}
	mg.Spec.ForProvider.EventDeliveryFailure = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.EventDeliveryFailureRef = rsp.ResolvedReference

	// Resolve spec.forProvider.eventEndpointCreated
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.EventEndpointCreated),
		Reference:    mg.Spec.ForProvider.EventEndpointCreatedRef,
		Selector:     mg.Spec.ForProvider.EventEndpointCreatedSelector,
		To:           reference.To{Managed: &Topic{}, List: &TopicList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.eventEndpointCreated")
	}
	mg.Spec.ForProvider.EventEndpointCreated = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.EventEndpointCreatedRef = rsp.ResolvedReference

	// Resolve spec.forProvider.eventEndpointDeleted
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.EventEndpointDeleted),
		Reference:    mg.Spec.ForProvider.EventEndpointDeletedRef,
		Selector:     mg.Spec.ForProvider.EventEndpointDeletedSelector,
		To:           reference.To{Managed: &Topic{}, List: &TopicList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.eventEndpointDeleted")
	}
	mg.Spec.ForProvider.EventEndpointDeleted = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.EventEndpointDeletedRef = rsp.ResolvedReference

	// Resolve spec.forProvider.eventEndpointUpdated
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.EventEndpointUpdated),
		Reference:    mg.Spec.ForProvider.EventEndpointUpdatedRef,
		Selector:     mg.Spec.ForProvider.EventEndpointUpdatedSelector,
		To:           reference.To{Managed: &Topic{}, List: &TopicList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.eventEndpointUpdated")
	}
	mg.Spec.ForProvider.EventEndpointUpdated = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.EventEndpointUpdatedRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Stage
func (mg *PlatformEndpoint) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.platformApplicationARN
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.PlatformApplicationARN),
		Reference:    mg.Spec.ForProvider.PlatformApplicationARNRef,
		Selector:     mg.Spec.ForProvider.PlatformApplicationARNSelector,
		To:           reference.To{Managed: &PlatformApplication{}, List: &PlatformApplicationList{}},
		Extract:      iamv1beta1.IAMRoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.platformApplicationARN")
	}
	mg.Spec.ForProvider.PlatformApplicationARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.PlatformApplicationARNRef = rsp.ResolvedReference
	return nil
}
