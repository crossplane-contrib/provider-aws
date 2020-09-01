package bucketclients

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// BucketResource is the interface all bucket sub-resources must conform to
type BucketResource interface {
	ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error)
	CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error)
	DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error
}

// MakeClients creates the array of all clients for a given BucketProvider
func MakeClients(parameters v1beta1.BucketParameters) []BucketResource {
	clients := make([]BucketResource, 0)
	clients = append(clients, CreateAccelerateConfigurationClient(parameters),
		CreateCORSConfigurationClient(parameters),
		CreateLifecycleConfigurationClient(parameters),
		CreateLoggingConfigurationClient(parameters),
		CreateNotificationConfigurationClient(parameters),
		CreateReplicationConfigurationClient(parameters),
		CreateRequestPaymentConfigurationClient(parameters),
		CreateSSEConfigurationClient(parameters),
		CreateTaggingConfigurationClient(parameters),
		CreateVersioningConfigurationClient(parameters),
		CreateWebsiteConfigurationClient(parameters),
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
