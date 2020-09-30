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

package bucketresources

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	accelGetFailed           = "cannot get Bucket accelerate configuration"
	accelPutFailed           = "cannot put Bucket acceleration configuration"
	accelDeleteFailed        = "cannot delete Bucket acceleration configuration"
	corsGetFailed            = "cannot get Bucket CORS configuration"
	corsPutFailed            = "cannot put Bucket cors"
	corsDeleteFailed         = "cannot delete Bucket CORS configuration"
	lifecycleGetFailed       = "cannot get Bucket lifecycle"
	lifecyclePutFailed       = "cannot put Bucket lifecycle"
	lifecycleDeleteFailed    = "cannot delete Bucket lifecycle configuration"
	loggingGetFailed         = "cannot get Bucket logging configuration"
	loggingPutFailed         = "cannot put Bucket logging configuration"
	loggingDeleteFailed      = "cannot delete Bucket logging configuration"
	notificationGetFailed    = "cannot get Bucket notification"
	notificationPutFailed    = "cannot put Bucket notification"
	notificationDeleteFailed = "cannot delete Bucket notification"
	replicationGetFailed     = "cannot get replication configuration"
	replicationPutFailed     = "cannot put Bucket replication"
	replicationDeleteFailed  = "cannot delete Bucket replication"
	paymentGetFailed         = "cannot get request payment configuration"
	paymentPutFailed         = "cannot put Bucket payment"
	sseGetFailed             = "cannot get Bucket encryption configuration"
	ssePutFailed             = "cannot put Bucket encryption configuration"
	sseDeleteFailed          = "cannot delete Bucket encryption configuration"
	taggingGetFailed         = "cannot get Bucket tagging set"
	taggingPutFailed         = "cannot put Bucket tagging set"
	taggingDeleteFailed      = "cannot delete Bucket tagging set"
	versioningGetFailed      = "cannot get Bucket versioning configuration"
	versioningPutFailed      = "cannot put Bucket versioning configuration"
	versioningDeleteFailed   = "cannot delete Bucket versioning configuration"
	websiteGetFailed         = "cannot get Bucket website configuration"
	websitePutFailed         = "cannot put Bucket website configuration"
	websiteDeleteFailed      = "cannot delete Bucket website configuration"
)

// BucketResource is the interface all Bucket sub-resources must conform to
type BucketResource interface {
	Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error)
	CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error)
	Delete(ctx context.Context, bucket *v1beta1.Bucket) error
	LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error
}

// MakeControllers creates the array of all clients for a given BucketProvider
func MakeControllers(bucket *v1beta1.Bucket, client s3.BucketClient) []BucketResource {
	clients := make([]BucketResource, 0)
	clients = append(clients, NewAccelerateConfigurationClient(client),
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
	)
	return clients
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
