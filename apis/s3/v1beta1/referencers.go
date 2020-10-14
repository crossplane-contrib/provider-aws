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

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
)

// SNSTopicARN returns a function that returns the ARN of the given SNS Topic.
func SNSTopicARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*v1alpha1.SNSTopic)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}

// ResolveReferences of this Bucket
func (mg *Bucket) ResolveReferences(ctx context.Context, c client.Reader) error { // nolint:gocyclo
	r := reference.NewAPIResolver(c, mg)
	// Resolve spec.forProvider.notificationConfiguration.topicConfigurations[].topicArn
	// TODO - need a way to extract arbitrary ARNs from resources - for the
	// topic we are missing a lot of information
	if mg.Spec.ForProvider.NotificationConfiguration != nil {
		for i, v := range mg.Spec.ForProvider.NotificationConfiguration.TopicConfigurations {
			rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
				CurrentValue: v.TopicArn,
				Reference:    v.TopicArnRef,
				Selector:     v.TopicArnSelector,
				To:           reference.To{Managed: &v1alpha1.SNSTopic{}, List: &v1alpha1.SNSTopicList{}},
				Extract:      SNSTopicARN(),
			})
			if err != nil {
				return errors.Wrapf(err, "spec.forProvider.notificationConfiguration.topicConfigurations[%d].topicArn", i)
			}
			mg.Spec.ForProvider.NotificationConfiguration.TopicConfigurations[i].TopicArn = rsp.ResolvedValue
			mg.Spec.ForProvider.NotificationConfiguration.TopicConfigurations[i].TopicArnRef = rsp.ResolvedReference
		}
	}

	// Resolve spec.forProvider.loggingConfiguration.targetBucket
	if mg.Spec.ForProvider.LoggingConfiguration != nil {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.LoggingConfiguration.TargetBucket),
			Reference:    mg.Spec.ForProvider.LoggingConfiguration.TargetBucketRef,
			Selector:     mg.Spec.ForProvider.LoggingConfiguration.TargetBucketSelector,
			To:           reference.To{Managed: &Bucket{}, List: &BucketList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.loggingConfiguration.targetBucket")
		}
		mg.Spec.ForProvider.LoggingConfiguration.TargetBucket = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.LoggingConfiguration.TargetBucketRef = rsp.ResolvedReference
	}

	// Resolve spec.forProvider.replicationConfiguration.role
	if mg.Spec.ForProvider.ReplicationConfiguration != nil {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ReplicationConfiguration.Role),
			Reference:    mg.Spec.ForProvider.ReplicationConfiguration.RoleRef,
			Selector:     mg.Spec.ForProvider.ReplicationConfiguration.RoleSelector,
			To:           reference.To{Managed: &v1beta1.IAMRole{}, List: &v1beta1.IAMRoleList{}},
			Extract:      v1beta1.IAMRoleARN(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.replicationConfiguration.role")
		}
		mg.Spec.ForProvider.ReplicationConfiguration.Role = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.ReplicationConfiguration.RoleRef = rsp.ResolvedReference
	}

	// Resolve spec.forProvider.replicationConfiguration.rules[*].bucket
	if mg.Spec.ForProvider.ReplicationConfiguration != nil {
		for i, v := range mg.Spec.ForProvider.ReplicationConfiguration.Rules {
			rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
				CurrentValue: reference.FromPtrValue(v.Destination.Bucket),
				Reference:    v.Destination.BucketRef,
				Selector:     v.Destination.BucketSelector,
				To:           reference.To{Managed: &Bucket{}, List: &BucketList{}},
				Extract:      reference.ExternalName(),
			})
			if err != nil {
				return errors.Wrapf(err, "spec.forProvider.replicationConfiguration.rules[%d].bucket", i)
			}
			mg.Spec.ForProvider.ReplicationConfiguration.Rules[i].Destination.Bucket = reference.ToPtrValue(rsp.ResolvedValue)
			mg.Spec.ForProvider.ReplicationConfiguration.Rules[i].Destination.BucketRef = rsp.ResolvedReference
		}
	}

	return nil
}
