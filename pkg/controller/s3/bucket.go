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

package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/controller/s3/bucket"
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
func SetupBucket(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.BucketGroupKind)
	logger := l.WithValues("controller", name)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.Bucket{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.BucketGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: s3.NewClient, logger: logger}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(logger),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.LocationConstraint)
	if err != nil {
		return nil, err
	}
	s3client := c.newClientFn(*cfg)
	return &external{s3client: s3client, subresourceClients: bucket.NewSubresourceClients(s3client), kube: c.kube, logger: c.logger}, nil
}

type external struct {
	kube               client.Client
	s3client           s3.BucketClient
	logger             logging.Logger
	subresourceClients []bucket.SubresourceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint: gocyclo
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if _, err := e.s3client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: aws.String(meta.GetExternalName(cr))}); err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(s3.IsNotFound, err), errHead)
	}

	cr.Status.AtProvider = s3.GenerateBucketObservation(meta.GetExternalName(cr))

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
		if obs != bucket.Updated {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false, ResourceLateInitialized: lateInit}, nil
		}
	}

	// TODO: smarter updating for the bucket, we dont need to update the ACL every time
	err := s3.UpdateBucketACL(ctx, e.s3client, cr)
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
	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.s3client.CreateBucket(ctx, s3.GenerateCreateBucketInput(meta.GetExternalName(cr), cr.Spec.ForProvider))
	if resource.Ignore(s3.IsAlreadyExists, err) != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
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
			errs = append(errs, awsclient.Wrap(err, errKubeUpdateFailed))
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
		case bucket.NeedsDeletion:
			err = awsClient.Delete(ctx, cr)
			if err != nil {
				return managed.ExternalUpdate{}, awsclient.Wrap(err, errDelete)
			}
		case bucket.NeedsUpdate:
			if err := awsClient.CreateOrUpdate(ctx, cr); err != nil {
				return managed.ExternalUpdate{}, awsclient.Wrap(err, errCreateOrUpdate)
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
