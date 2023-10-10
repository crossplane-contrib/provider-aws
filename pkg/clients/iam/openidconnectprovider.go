/*
Copyright 2021 The Crossplane Authors.

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

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

// NewOpenIDConnectProviderClient returns a new client using AWS credentials as JSON encoded data.
func NewOpenIDConnectProviderClient(cfg aws.Config) OpenIDConnectProviderClient {
	return iam.NewFromConfig(cfg)
}

// OpenIDConnectProviderClient is the external client used for IAM OpenIDConnectProvide Custom Resource
type OpenIDConnectProviderClient interface {
	GetOpenIDConnectProvider(ctx context.Context, input *iam.GetOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error)
	CreateOpenIDConnectProvider(ctx context.Context, input *iam.CreateOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	AddClientIDToOpenIDConnectProvider(ctx context.Context, input *iam.AddClientIDToOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	RemoveClientIDFromOpenIDConnectProvider(ctx context.Context, input *iam.RemoveClientIDFromOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error)
	UpdateOpenIDConnectProviderThumbprint(ctx context.Context, input *iam.UpdateOpenIDConnectProviderThumbprintInput, opts ...func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error)
	DeleteOpenIDConnectProvider(ctx context.Context, input *iam.DeleteOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error)
	TagOpenIDConnectProvider(ctx context.Context, input *iam.TagOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.TagOpenIDConnectProviderOutput, error)
	UntagOpenIDConnectProvider(ctx context.Context, input *iam.UntagOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.UntagOpenIDConnectProviderOutput, error)
	ListOpenIDConnectProviders(ctx context.Context, input *iam.ListOpenIDConnectProvidersInput, optFns ...func(*iam.Options)) (*iam.ListOpenIDConnectProvidersOutput, error)
	ListOpenIDConnectProviderTags(ctx context.Context, input *iam.ListOpenIDConnectProviderTagsInput, optFns ...func(*iam.Options)) (*iam.ListOpenIDConnectProviderTagsOutput, error)
}

// GenerateOIDCProviderObservation is used to produce v1alpha1.OpenIDConnectProvider
// from iam.OpenIDConnectProvider
func GenerateOIDCProviderObservation(observed iam.GetOpenIDConnectProviderOutput) svcapitypes.OpenIDConnectProviderObservation {
	o := svcapitypes.OpenIDConnectProviderObservation{}
	if observed.CreateDate != nil {
		createdTime := metav1.NewTime(*observed.CreateDate)
		o.CreateDate = &createdTime
	}
	return o
}

// IsOIDCProviderUpToDate checks whether there is a change in any of the modifiable fields in OpenIDConnectProvider.
func IsOIDCProviderUpToDate(in svcapitypes.OpenIDConnectProviderParameters, observed iam.GetOpenIDConnectProviderOutput) bool {
	sortSlicesOpt := cmpopts.SortSlices(func(x, y string) bool {
		return x < y
	})
	sortSliceTags := cmpopts.SortSlices(func(x, y svcapitypes.Tag) bool {
		return x.Key < y.Key
	})
	if !cmp.Equal(in.ClientIDList, observed.ClientIDList, sortSlicesOpt, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(in.ThumbprintList, observed.ThumbprintList, sortSlicesOpt, cmpopts.EquateEmpty()) {
		return false
	}

	nTags := len(observed.Tags)
	if nTags == 0 {
		return true
	}
	cmpTags := make([]svcapitypes.Tag, nTags)
	for i := range observed.Tags {
		cmpTags[i] = svcapitypes.Tag{Key: *observed.Tags[i].Key, Value: *observed.Tags[i].Value}
	}
	return cmp.Equal(in.Tags, cmpTags, sortSliceTags, cmpopts.EquateEmpty())
}

// SliceDifference returns the elements to added and removed between the
// current and desired slices
func SliceDifference(current, desired []string) (add, remove []string) {
	currentMap := make(map[string]struct{}, len(current))
	for _, val := range current {
		currentMap[val] = struct{}{}
	}

	for _, val := range desired {
		_, exists := currentMap[val]
		if !exists {
			add = append(add, val)
			continue
		}
		delete(currentMap, val)
	}

	for val := range currentMap {
		remove = append(remove, val)
	}
	return
}
