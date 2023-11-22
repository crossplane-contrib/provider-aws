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

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not a Bucket"
	errHead             = "failed to query Bucket"
	errCreate           = "failed to create the Bucket"
	errCreateOrUpdate   = "cannot create or update"
	errDelete           = "cannot delete"
	errKubeUpdateFailed = "cannot update S3 custom resource"
)

// SetupBucket adds a controller that reconciles Buckets.
func SetupBucket(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.BucketGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: s3.NewClient, logger: o.Logger.WithValues("controller", name)}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.BucketGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Bucket{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) s3.BucketClient
	logger      logging.Logger
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.LocationConstraint)
	if err != nil {
		return nil, err
	}
	s3client := c.newClientFn(*cfg)
	return &external{s3client: s3client, subresourceClients: NewSubresourceClients(s3client), kube: c.kube, logger: c.logger}, nil
}

type external struct {
	kube               client.Client
	s3client           s3.BucketClient
	logger             logging.Logger
	subresourceClients []SubresourceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { //nolint: gocyclo
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if _, err := e.s3client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(meta.GetExternalName(cr))}); err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(s3.IsNotFound, err), errHead)
	}

	// get the proper partitionId for the bucket's region
	resolver := awss3.NewDefaultEndpointResolver()
	endpoint, err1 := resolver.ResolveEndpoint(cr.Spec.ForProvider.LocationConstraint, awss3.EndpointResolverOptions{})
	if err1 != nil {
		return managed.ExternalObservation{}, err1
	}

	cr.Status.AtProvider = s3.GenerateBucketObservation(meta.GetExternalName(cr), endpoint.PartitionID)

	lateInit := false
	current := cr.Spec.ForProvider.DeepCopy()

	for _, awsClient := range e.subresourceClients {
		if awsClient.SubresourceExists(cr) {
			// we need this check, because we do not want to late init resources the user has
			// manually removed, our main late init should happen in the Create method
			err := awsClient.LateInitialize(ctx, cr)
			if err != nil {
				return managed.ExternalObservation{}, err
			}
		}
	}

	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		lateInit = true
	}

	for _, awsClient := range e.subresourceClients {
		obs, err := awsClient.Observe(ctx, cr)
		if err != nil {
			return managed.ExternalObservation{}, err
		}
		if obs != Updated {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: lateInit}, nil
		}
	}

	// See https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAcl.html
	// If your bucket uses the bucket owner enforced setting for S3 Object
	// Ownership, ACLs are disabled and no longer affect permissions. You
	// must use policies to grant access to your bucket and the objects in
	// it. Requests to set ACLs or update ACLs fail and return the
	// AccessControlListNotSupported error code.
	if !s3.BucketHasACLsDisabled(cr) {
		// TODO: smarter updating for the bucket, we don't need to update the ACL every time
		err := s3.UpdateBucketACL(ctx, e.s3client, cr)
		if err != nil {
			return managed.ExternalObservation{}, err
		}
	}

	// TODO: smarter updating for the bucket, we don't need to update the ObjectOwnership every time
	err := s3.UpdateBucketOwnershipControls(ctx, e.s3client, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: lateInit,
		ConnectionDetails: map[string][]byte{
			xpv1.ResourceCredentialsSecretEndpointKey:  []byte(meta.GetExternalName(cr)),
			v1beta1.ResourceCredentialsSecretRegionKey: []byte(cr.Spec.ForProvider.LocationConstraint),
		},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	_, err := e.s3client.CreateBucket(ctx, s3.GenerateCreateBucketInput(meta.GetExternalName(cr), cr.Spec.ForProvider))
	if resource.Ignore(s3.IsAlreadyExists, err) != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	current := cr.Spec.ForProvider.DeepCopy()

	errs := make([]error, 0)
	for _, awsClient := range e.subresourceClients {
		err := awsClient.LateInitialize(ctx, cr)
		if err != nil {
			// aggregate errors since we dont want all late inits to fail if just the first one fails
			// this can only really be run on creation, and we lose fidelty if we let this go into the
			// reconcile loop/Observe func
			errs = append(errs, err)
		}
	}
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			errs = append(errs, errorutils.Wrap(err, errKubeUpdateFailed))
		}
	}
	if len(errs) != 0 {
		return managed.ExternalCreation{}, k8serrors.NewAggregate(errs)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	for _, awsClient := range e.subresourceClients {
		status, err := awsClient.Observe(ctx, cr)
		if err != nil {
			cr.Status.SetConditions(xpv1.ReconcileError(err))
			return managed.ExternalUpdate{}, err
		}
		switch status { //nolint:exhaustive
		case NeedsDeletion:
			err = awsClient.Delete(ctx, cr)
			if err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errDelete)
			}
		case NeedsUpdate:
			if err := awsClient.CreateOrUpdate(ctx, cr); err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreateOrUpdate)
			}
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.s3client.DeleteBucket(ctx, &awss3.DeleteBucketInput{Bucket: aws.String(meta.GetExternalName(cr))})
	return resource.Ignore(s3.IsNotFound, err)
}
