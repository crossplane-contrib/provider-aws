/*
Copyright 2022 The Crossplane Authors.

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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SNSTopicARN returns a function that returns the ARN of the given SNS Topic.
func SNSTopicARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Topic)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}

// ResolveReferences for SNS Subscription managed type
func (mg *Subscription) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.topicArn
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.TopicARN,
		Reference:    mg.Spec.ForProvider.TopicARNRef,
		Selector:     mg.Spec.ForProvider.TopicARNSelector,
		To:           reference.To{Managed: &Topic{}, List: &TopicList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.topicArn")
	}
	mg.Spec.ForProvider.TopicARN = rsp.ResolvedValue
	mg.Spec.ForProvider.TopicARNRef = rsp.ResolvedReference

	return nil
}
