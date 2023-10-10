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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	sns "github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
)

const errDeprecatedRef = "spec.forProvider.cacheSubnetGroupNameRefs is deprecated - please set only spec.forProvider.cacheSubnetGroupNameRef"

// ResolveReferences of this ReplicationGroup
func (mg *ReplicationGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse

	// We have two variants of this reference because it was originally
	// introduced with a plural JSON tag, when it should be singular. We
	// maintain both in order to avoid making a breaking change.
	switch {
	case mg.Spec.ForProvider.CacheSubnetGroupNameRef != nil && mg.Spec.ForProvider.DeprecatedCacheSubnetGroupNameRef != nil:
		return errors.New(errDeprecatedRef)
	case mg.Spec.ForProvider.DeprecatedCacheSubnetGroupNameRef != nil:
		// Resolve (deprecated) spec.forProvider.cacheSubnetGroupName
		resp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CacheSubnetGroupName),
			Reference:    mg.Spec.ForProvider.DeprecatedCacheSubnetGroupNameRef,
			Selector:     mg.Spec.ForProvider.CacheSubnetGroupNameSelector,
			To:           reference.To{Managed: &v1alpha1.CacheSubnetGroup{}, List: &v1alpha1.CacheSubnetGroupList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.cacheSubnetGroupName")
		}
		mg.Spec.ForProvider.CacheSubnetGroupName = reference.ToPtrValue(resp.ResolvedValue)
		mg.Spec.ForProvider.DeprecatedCacheSubnetGroupNameRef = resp.ResolvedReference
	default:
		// Resolve spec.forProvider.cacheSubnetGroupName
		resp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CacheSubnetGroupName),
			Reference:    mg.Spec.ForProvider.CacheSubnetGroupNameRef,
			Selector:     mg.Spec.ForProvider.CacheSubnetGroupNameSelector,
			To:           reference.To{Managed: &v1alpha1.CacheSubnetGroup{}, List: &v1alpha1.CacheSubnetGroupList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.cacheSubnetGroupName")
		}
		mg.Spec.ForProvider.CacheSubnetGroupName = reference.ToPtrValue(resp.ResolvedValue)
		mg.Spec.ForProvider.CacheSubnetGroupNameRef = resp.ResolvedReference
	}

	// Resolve spec.forProvider.securityGroupIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupIDSelector,
		To:            reference.To{Managed: &v1beta1.SecurityGroup{}, List: &v1beta1.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroupIds")
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.cacheSecurityGroupNames
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.CacheSecurityGroupNames,
		References:    mg.Spec.ForProvider.CacheSecurityGroupNameRefs,
		Selector:      mg.Spec.ForProvider.CacheSecurityGroupNameSelector,
		To:            reference.To{Managed: &v1beta1.SecurityGroup{}, List: &v1beta1.SecurityGroupList{}},
		Extract:       v1beta1.SecurityGroupName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.cacheSecurityGroupNames")
	}
	mg.Spec.ForProvider.CacheSecurityGroupNames = mrsp.ResolvedValues
	mg.Spec.ForProvider.CacheSecurityGroupNameRefs = mrsp.ResolvedReferences

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.NotificationTopicARN),
		Extract:      sns.SNSTopicARN(),
		Reference:    mg.Spec.ForProvider.NotificationTopicARNRef,
		Selector:     mg.Spec.ForProvider.NotificationTopicARNSelector,
		To: reference.To{
			List:    &sns.TopicList{},
			Managed: &sns.Topic{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.NotificationTopicARN")
	}
	mg.Spec.ForProvider.NotificationTopicARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NotificationTopicARNRef = rsp.ResolvedReference

	return nil
}
