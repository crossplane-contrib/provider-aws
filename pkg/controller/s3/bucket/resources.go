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

package bucket

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// SubresourceClient is the interface all Bucket sub-resources must conform to
type SubresourceClient interface {
	Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error)
	CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error)
	Delete(ctx context.Context, bucket *v1beta1.Bucket) error
	LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error
}

// NewSubresourceClients creates the array of all clients for a given BucketProvider
func NewSubresourceClients(client s3.BucketClient) []SubresourceClient {
	return []SubresourceClient{
		NewAccelerateConfigurationClient(client),
		NewCORSConfigurationClient(client),
		NewLifecycleConfigurationClient(client),
		NewLoggingConfigurationClient(client),
		NewNotificationConfigurationClient(client),
		NewReplicationConfigurationClient(client),
		NewRequestPaymentConfigurationClient(client),
		NewSSEConfigurationClient(client),
		NewTaggingConfigurationClient(client),
		NewVersioningConfigurationClient(client),
		NewWebsiteConfigurationClient(client),
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
)
