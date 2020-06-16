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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences for SNS Topic managed type
func (mg *SNSSubscription) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.TopicID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.TopicArn,
		Reference:    mg.Spec.ForProvider.TopicArnRef,
		Selector:     mg.Spec.ForProvider.TopicArnSelector,
		To:           reference.To{Managed: &SNSTopic{}, List: &SNSTopicList{}},
		Extract:      reference.ExternalName(),
	})

	if err != nil {
		return err
	}
	mg.Spec.ForProvider.TopicArn = rsp.ResolvedValue
	mg.Spec.ForProvider.TopicArnRef = rsp.ResolvedReference

	return nil
}
