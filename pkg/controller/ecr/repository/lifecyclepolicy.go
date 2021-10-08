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

package repository

import (
	"context"
	"encoding/json"
	"sort"

	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	ecr "github.com/crossplane/provider-aws/pkg/clients/ecr"
)

const (
	errComparingLifecyclePolicy = "failed to compare lifecycle policies"
	errGettingLifecyclePolicy   = "failed to get lifecycle policy for the repository resource"
	errCreateLifecyclePolicy    = "failed to create lifecycle policy for the repository resource"
	errDeletingLifecyclePolicy  = "failed to delete lifecycle policy for the repository resource"
)

// LifecyclePolicyClient is the client for API methods and reconciling the LifecyclePolicy
type LifecyclePolicyClient struct {
	client ecr.RepositoryClient
}

// NewLifecyclePolicyClient creates the client for ECR Repository Lifecycle Policy
func NewLifecyclePolicyClient(client ecr.RepositoryClient) *LifecyclePolicyClient {
	return &LifecyclePolicyClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *LifecyclePolicyClient) Observe(ctx context.Context, cr *v1alpha1.Repository) (ResourceStatus, error) {
	l := cr.Spec.ForProvider.LifecyclePolicy
	external, err := in.client.GetLifecyclePolicy(ctx, &awsecr.GetLifecyclePolicyInput{
		RepositoryName: awsclient.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		if ecr.IsLifecyclePolicyNotFoundErr(err) && l == nil {
			return Updated, nil
		}
		return NeedsCreate, awsclient.Wrap(resource.Ignore(ecr.IsLifecyclePolicyNotFoundErr, err), errGettingLifecyclePolicy)
	}

	if ok := observeEmptyValidation(l, external); ok != -1 {
		return ok, nil
	}

	externalLifecyclePolicy, err := decodeLifecyclePolicy(*external.LifecyclePolicyText)
	if err != nil {
		return NeedsUpdate, awsclient.Wrap(err, errComparingLifecyclePolicy)
	}
	sortRules(l)
	sortRules(externalLifecyclePolicy)

	// Unmarshalled object and cr differ
	if diff := cmp.Diff(l, externalLifecyclePolicy); diff != "" {
		return NeedsUpdate, nil
	}

	// TODO: Sort string slice (tagPrefix)
	return Updated, nil
}

func observeEmptyValidation(local *v1alpha1.LifecyclePolicy, external *awsecr.GetLifecyclePolicyOutput) ResourceStatus {
	switch {
	// Lifecycle policy not configured
	case local == nil && external == nil:
		return Updated

	// lifecycle not configured for CR and external resources is empty even though error not thrown
	case local == nil && (external.LifecyclePolicyText == nil || *external.LifecyclePolicyText == ""):
		return Updated

	// lifecycle not configured for CR but external lifecycle policy is not empty
	case local == nil && external.LifecyclePolicyText != nil:
		return NeedsDeletion
	}

	// Proceed
	return -1
}

// CreateOrUpdate sends a request to have resource created on awsclient
func (in *LifecyclePolicyClient) CreateOrUpdate(ctx context.Context, cr *v1alpha1.Repository) error {
	if cr.Spec.ForProvider.LifecyclePolicy == nil {
		return nil
	}
	policyText, err := json.Marshal(cr.Spec.ForProvider.LifecyclePolicy)
	if err != nil {
		return awsclient.Wrap(err, errCreateLifecyclePolicy)
	}
	policyTextString := string(policyText)
	_, err = in.client.PutLifecyclePolicy(ctx, &awsecr.PutLifecyclePolicyInput{
		LifecyclePolicyText: &policyTextString,
		RepositoryName:      awsclient.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return awsclient.Wrap(err, errCreateLifecyclePolicy)
	}
	return nil
}

// Delete removes the lifecycle policy from the repository
func (in *LifecyclePolicyClient) Delete(ctx context.Context, cr *v1alpha1.Repository) error {
	_, err := in.client.DeleteLifecyclePolicy(ctx, &awsecr.DeleteLifecyclePolicyInput{RepositoryName: awsclient.String(meta.GetExternalName(cr))})
	if err != nil {
		return awsclient.Wrap(err, errDeletingLifecyclePolicy)
	}
	return nil
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *LifecyclePolicyClient) LateInitialize(ctx context.Context, cr *v1alpha1.Repository) error {
	external, err := in.client.GetLifecyclePolicy(ctx, &awsecr.GetLifecyclePolicyInput{RepositoryName: awsclient.String(meta.GetExternalName(cr))})
	if err != nil {
		return awsclient.Wrap(err, errGettingLifecyclePolicy)
	}
	if external == nil || external.LifecyclePolicyText == nil || *external.LifecyclePolicyText == "" {
		return nil
	}

	if cr.Spec.ForProvider.LifecyclePolicy == nil {
		p, err := decodeLifecyclePolicy(*external.LifecyclePolicyText)
		if err != nil {
			return awsclient.Wrap(err, errGettingLifecyclePolicy)
		}
		cr.Spec.ForProvider.LifecyclePolicy = p
	}
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *LifecyclePolicyClient) SubresourceExists(cr *v1alpha1.Repository) bool {
	return cr.Spec.ForProvider.LifecyclePolicy != nil
}

func decodeLifecyclePolicy(lifecyclePolicyText string) (*v1alpha1.LifecyclePolicy, error) {
	var lifecycle v1alpha1.LifecyclePolicy
	if err := json.Unmarshal([]byte(lifecyclePolicyText), &lifecycle); err != nil {
		return nil, err
	}
	return &lifecycle, nil
}

func sortRules(lifecyclePolicy *v1alpha1.LifecyclePolicy) {
	sort.Slice(lifecyclePolicy.Rules, func(i, j int) bool {
		return lifecyclePolicy.Rules[i].RulePriority < lifecyclePolicy.Rules[j].RulePriority
	})
}
