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
	"reflect"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/pkg/clients/s3"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
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
func SetupBucket(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.BucketGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Bucket{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.BucketGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: s3.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) s3.BucketClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.LocationConstraint)
	if err != nil {
		return nil, err
	}
	s3client := c.newClientFn(*cfg)
	return &external{s3client: s3client, subresourceClients: bucket.NewSubresourceClients(s3client), kube: c.kube}, nil
}

type external struct {
	kube               client.Client
	s3client           s3.BucketClient
	subresourceClients []bucket.SubresourceClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if _, err := e.s3client.HeadBucketRequest(&awss3.HeadBucketInput{Bucket: aws.String(meta.GetExternalName(cr))}).Send(ctx); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(s3.IsNotFound, err), errHead)
	}

	cr.Status.AtProvider = s3.GenerateBucketObservation(meta.GetExternalName(cr))

	current := cr.Spec.ForProvider.DeepCopy()
	for _, awsClient := range e.subresourceClients {
		err := awsClient.LateInitialize(ctx, cr)
		if err != nil {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, err
		}
		obs, err := awsClient.Observe(ctx, cr)
		if err != nil {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, err
		}
		if obs != bucket.Updated {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
	}

	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	// TODO: smarter updating for the bucket, we dont need to update the ACL every time
	err := s3.UpdateBucketACL(ctx, e.s3client, cr)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, err
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	_, err := e.s3client.CreateBucketRequest(s3.GenerateCreateBucketInput(meta.GetExternalName(cr), cr.Spec.ForProvider)).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	for _, awsClient := range e.subresourceClients {
		status, err := awsClient.Observe(ctx, cr)
		if err != nil {
			cr.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
			return managed.ExternalUpdate{}, err
		}
		switch status { //nolint:exhaustive
		case bucket.NeedsDeletion:
			err = awsClient.Delete(ctx, cr)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, errDelete)
			}
		case bucket.NeedsUpdate:
			if err := awsClient.CreateOrUpdate(ctx, cr); err != nil {
				// TODO(muvaf): let the user know which client failed.
				return managed.ExternalUpdate{}, errors.Wrap(err, errCreateOrUpdate)
			}
		}
	}
	cr.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.s3client.DeleteBucketRequest(&awss3.DeleteBucketInput{Bucket: aws.String(meta.GetExternalName(cr))}).Send(ctx)
	return resource.Ignore(s3.IsNotFound, err)
}
