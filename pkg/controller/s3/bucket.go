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

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/pkg/clients/s3"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3/bucketclients"
)

const (
	errUnexpectedObject   = "The managed resource is not a Bucket"
	errCreateBucketClient = "cannot create the Bucket client"
	errGetProvider        = "cannot get provider"
	errGetProviderSecret  = "cannot get provider secret"
	errHead               = "failed to query Bucket"
	errCreate             = "failed to create the Bucket"

	//errKubeUpdateFailed = "cannot update VPC custom resource"
	//errMultipleItems       = "retrieved multiple VPCs for the given vpcId"
	//errUpdate              = "failed to update VPC resource"
	//errModifyVPCAttributes = "failed to modify the VPC resource attributes"
	//errCreateTags          = "failed to create tags for the VPC resource"
	//errDelete              = "failed to delete the VPC resource"
	//errSpecUpdate          = "cannot update spec of VPC custom resource"
	//errStatusUpdate        = "cannot update status of VPC custom resource"
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
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (s3.BucketClient, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.Spec.ProviderReference.Name}, p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		bucketClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{s3client: bucketClient, kube: c.kube, clients: bucketclients.MakeClients(cr, bucketClient)}, errors.Wrap(err, errCreateBucketClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	bucketClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{s3client: bucketClient, kube: c.kube, clients: bucketclients.MakeClients(cr, bucketClient)}, errors.Wrap(err, errCreateBucketClient)
}

type external struct {
	kube     client.Client
	s3client s3.BucketClient
	clients  []bucketclients.BucketResource
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if _, err := e.s3client.HeadBucketRequest(&awss3.HeadBucketInput{Bucket: aws.String(meta.GetExternalName(cr))}).Send(ctx); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(s3.IsNotFound, err), errHead)
	}
	for _, client := range e.clients {
		updated, err := client.Observe(ctx, cr)
		if err != nil {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, err
		}
		if updated == bucketclients.NeedsUpdate {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
		if updated == bucketclients.NeedsDeletion {
			return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil
		}
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
	_, err := e.s3client.CreateBucketRequest(s3.GenerateCreateBucketInput(meta.GetExternalName(cr), cr.Spec.Parameters)).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	for _, client := range e.clients {
		status, err := client.Observe(ctx, cr)
		if err != nil {
			cr.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
			return managed.ExternalUpdate{}, err
		}
		switch status {
		case bucketclients.NeedsDeletion:
			err = client.Delete(ctx, cr)
			if err != nil {
				cr.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
				return managed.ExternalUpdate{}, err
			}
		case bucketclients.NeedsUpdate:
			update, err := client.Create(ctx, cr)
			if err != nil {
				cr.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
				return update, err
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
