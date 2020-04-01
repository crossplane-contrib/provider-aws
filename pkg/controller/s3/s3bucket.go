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

package s3

import (
	"context"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bucketv1alpha3 "github.com/crossplane/provider-aws/apis/storage/v1alpha3"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/controller/utils"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	controllerName = "s3bucket.aws.crossplane.io"
	finalizer      = "finalizer." + controllerName
)

// Amounts of time we wait before requeuing a reconcile.
const (
	aLongWait = 60 * time.Second
)

// Error strings
const (
	errUpdateManagedStatus = "cannot update managed resource status"
)

var (
	ctx           = context.Background()
	result        = reconcile.Result{}
	resultRequeue = reconcile.Result{Requeue: true}
)

// Reconciler reconciles a S3Bucket object
type Reconciler struct {
	client.Client
	scheme *runtime.Scheme
	managed.ReferenceResolver
	managed.ConnectionPublisher

	connect func(*bucketv1alpha3.S3Bucket) (s3.Service, error)
	create  func(*bucketv1alpha3.S3Bucket, s3.Service) (reconcile.Result, error)
	sync    func(*bucketv1alpha3.S3Bucket, s3.Service) (reconcile.Result, error)
	delete  func(*bucketv1alpha3.S3Bucket, s3.Service) (reconcile.Result, error)

	log logging.Logger
}

// SetupS3Bucket adds a controller that reconciles S3Buckets.
func SetupS3Bucket(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(bucketv1alpha3.S3BucketClassGroupKind)

	r := &Reconciler{
		Client:              mgr.GetClient(),
		scheme:              mgr.GetScheme(),
		ReferenceResolver:   managed.NewAPIReferenceResolver(mgr.GetClient()),
		ConnectionPublisher: managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()),
		log:                 l.WithValues("controller", name),
	}
	r.connect = r._connect
	r.create = r._create
	r.delete = r._delete
	r.sync = r._sync

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&bucketv1alpha3.S3Bucket{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// fail - helper function to set fail condition with reason and message
func (r *Reconciler) fail(bucket *bucketv1alpha3.S3Bucket, err error) (reconcile.Result, error) {
	bucket.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
	return reconcile.Result{Requeue: true}, r.Update(context.TODO(), bucket)
}

func (r *Reconciler) _connect(instance *bucketv1alpha3.S3Bucket) (s3.Service, error) {
	config, err := utils.RetrieveAwsConfigFromProvider(ctx, r, instance.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	// NOTE(negz): Buckets must specify a region for creation. They never use
	// the provider's region. This should be addressed per the below issue.
	// https://github.com/crossplane/provider-aws/issues/38
	config.Region = instance.Spec.Region

	// Create new S3 S3Client
	return s3.NewClient(config), nil
}

func (r *Reconciler) _create(bucket *bucketv1alpha3.S3Bucket, client s3.Service) (reconcile.Result, error) {
	bucket.Status.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())
	meta.AddFinalizer(bucket, finalizer)
	err := client.CreateOrUpdateBucket(bucket)
	if err != nil {
		return r.fail(bucket, err)
	}

	// Set username for iam user
	if bucket.Spec.IAMUsername == "" {
		bucket.Spec.IAMUsername = s3.GenerateBucketUsername(bucket)
	}

	// Get access keys for iam user
	accessKeys, currentVersion, err := client.CreateUser(bucket.Spec.IAMUsername, bucket)
	if err != nil {
		return r.fail(bucket, err)
	}

	// Set user policy version in status so we can detect policy drift
	err = bucket.SetUserPolicyVersion(currentVersion)
	if err != nil {
		return r.fail(bucket, err)
	}

	if err := r.PublishConnection(ctx, bucket, managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(aws.StringValue(accessKeys.AccessKeyId)),
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(aws.StringValue(accessKeys.SecretAccessKey)),
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(bucket.Spec.Region),
	}); err != nil {
		return r.fail(bucket, err)
	}

	// No longer creating, we're ready!
	bucket.Status.SetConditions(runtimev1alpha1.Available())
	resource.SetBindable(bucket)
	return result, r.Update(ctx, bucket)
}

func (r *Reconciler) _sync(bucket *bucketv1alpha3.S3Bucket, client s3.Service) (reconcile.Result, error) {
	if bucket.Spec.IAMUsername == "" {
		return r.fail(bucket, errors.New("username not set, .Status.IAMUsername"))
	}
	bucketInfo, err := client.GetBucketInfo(bucket.Spec.IAMUsername, bucket)
	if err != nil {
		return r.fail(bucket, err)
	}

	if bucketInfo.Versioning != bucket.Spec.Versioning {
		err := client.UpdateVersioning(bucket)
		if err != nil {
			return r.fail(bucket, err)
		}
	}

	// TODO: Detect if the bucket CannedACL has changed, possibly by managing grants list directly.
	err = client.UpdateBucketACL(bucket)
	if err != nil {
		return r.fail(bucket, err)
	}

	// Eventually consistent, so we check if this version is newer than our stored version.
	changed, err := bucket.HasPolicyChanged(bucketInfo.UserPolicyVersion)
	if err != nil {
		return r.fail(bucket, err)
	}
	if changed {
		currentVersion, err := client.UpdatePolicyDocument(bucket.Spec.IAMUsername, bucket)
		if err != nil {
			return r.fail(bucket, err)
		}
		err = bucket.SetUserPolicyVersion(currentVersion)
		if err != nil {
			return r.fail(bucket, err)
		}
	}

	bucket.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())
	return result, r.Update(ctx, bucket)
}

func (r *Reconciler) _delete(bucket *bucketv1alpha3.S3Bucket, client s3.Service) (reconcile.Result, error) {
	bucket.Status.SetConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileSuccess())
	if bucket.Spec.ReclaimPolicy == runtimev1alpha1.ReclaimDelete {
		if err := client.DeleteBucket(bucket); err != nil {
			return r.fail(bucket, err)
		}
	}

	meta.RemoveFinalizer(bucket, finalizer)
	return result, r.Update(ctx, bucket)
}

// Reconcile reads that state of the bucket for an Instance object and makes changes based on the state read
// and what is in the Instance.Spec
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Debug("Reconciling", "request", request)

	// Fetch the CRD instance
	bucket := &bucketv1alpha3.S3Bucket{}

	err := r.Get(ctx, request.NamespacedName, bucket)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return result, nil
		}
		return result, err
	}

	s3Client, err := r.connect(bucket)
	if err != nil {
		return r.fail(bucket, err)
	}

	if !resource.IsConditionTrue(bucket.GetCondition(runtimev1alpha1.TypeReferencesResolved)) {
		if err := r.ResolveReferences(ctx, bucket); err != nil {
			condition := runtimev1alpha1.ReconcileError(err)
			if managed.IsReferencesAccessError(err) {
				condition = runtimev1alpha1.ReferenceResolutionBlocked(err)
			}

			bucket.Status.SetConditions(condition)
			return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, bucket), errUpdateManagedStatus)
		}

		// Add ReferenceResolutionSuccess to the conditions
		bucket.Status.SetConditions(runtimev1alpha1.ReferenceResolutionSuccess())
	}

	// Check for deletion
	if bucket.DeletionTimestamp != nil {
		return r.delete(bucket, s3Client)
	}

	// Create s3 bucket
	if bucket.Spec.IAMUsername == "" {
		return r.create(bucket, s3Client)
	}

	// Update the bucket if it's no longer there.
	return r.sync(bucket, s3Client)
}
