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

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/ecr"
)

// SubresourceClient is the interface all Bucket sub-resources must conform to
type SubresourceClient interface {
	Observe(ctx context.Context, repository *v1alpha1.Repository) (ResourceStatus, error)
	CreateOrUpdate(ctx context.Context, repository *v1alpha1.Repository) error
	Delete(ctx context.Context, repository *v1alpha1.Repository) error
	LateInitialize(ctx context.Context, repository *v1alpha1.Repository) error
	SubresourceExists(repository *v1alpha1.Repository) bool
}

// NewSubresourceClients creates the array of all clients for a given BucketProvider
func NewSubresourceClients(client ecr.RepositoryClient) []SubresourceClient {
	return []SubresourceClient{
		NewLifecyclePolicyClient(client),
	}
}

// ResourceStatus represents the current status  if the resource resource is updated.
type ResourceStatus int

const (
	// Updated is returned if the resource is updated.
	Updated ResourceStatus = iota
	// NeedsUpdate is returned if the resource required updating.
	NeedsUpdate
	// NeedsDeletion is returned if the resource needs to be deleted.
	NeedsDeletion
	// NeedsCreate is returned when the object is not found
	NeedsCreate
)
