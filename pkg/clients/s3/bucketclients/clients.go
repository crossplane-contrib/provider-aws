package bucketclients

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// BucketResource is the interface all bucket sub-resources must conform to
type BucketResource interface {
	ExistsAndUpdated(ctx context.Context) (ResourceStatus, error)
	CreateResource(ctx context.Context) (managed.ExternalUpdate, error)
	DeleteResource(ctx context.Context) error
}

// MakeClients creates the array of all clients for a given BucketProvider
func MakeClients(bucket *v1beta1.Bucket, client s3.BucketClient) []BucketResource {
	clients := make([]BucketResource, 0)
	clients = append(clients, CreateAccelerateConfigurationClient(bucket, client),
		CreateCORSConfigurationClient(bucket, client),
		CreateLifecycleConfigurationClient(bucket, client),
		CreateLoggingConfigurationClient(bucket, client),
		CreateNotificationConfigurationClient(bucket, client),
		CreateReplicationConfigurationClient(bucket, client),
		CreateRequestPaymentConfigurationClient(bucket, client),
		CreateSSEConfigurationClient(bucket, client),
		CreateTaggingConfigurationClient(bucket, client),
		CreateVersioningConfigurationClient(bucket, client),
		CreateWebsiteConfigurationClient(bucket, client),
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
