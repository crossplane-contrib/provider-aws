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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/storage/v1beta1"
	bucketv1beta1 "github.com/crossplane/provider-aws/apis/storage/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/rds"
	"github.com/crossplane/provider-aws/pkg/clients/s3"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Error strings
const (
	errUnexpectedObject = "The managed resource is not a S3Bucket resource"

	errCreates3Client = "cannot create RDS client"
	errCreateBucket   = "cannot create S3 bucket"
	errGetBucket      = "cannot get S3 bucket"
	errDeleteBucket   = "cannot delete S3 bucket"
	errUpdate         = "cannot update S3 update"
	errPolicyAttach   = "cannot attached policy to S3 bucket"

	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"
)

var (
	ctx = context.Background()
)

// SetupS3Bucket adds a controller that reconciles S3Buckets.
func SetupS3Bucket(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(bucketv1beta1.S3BucketClassGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&bucketv1beta1.S3Bucket{}).
		Owns(&corev1.Secret{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(bucketv1beta1.S3BucketGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: s3.NewClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (s3.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*bucketv1beta1.S3Bucket)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		s3Client, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: s3Client, kube: c.kube}, errors.Wrap(err, errCreates3Client)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	s3Client, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: s3Client, kube: c.kube}, errors.Wrap(err, errCreates3Client)
}

type external struct {
	client s3.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.S3Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	name := meta.GetExternalName(cr)

	// Check if bucket in spec exists
	_, err := e.client.HeadBucketRequest(&awsS3.HeadBucketInput{
		Bucket: aws.String(name),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(s3.IsErrorNotFound, err), errGetBucket)
	}

	policy := e.getBucketPolicy(meta.GetExternalName(cr))

	cr.Status.AtProvider = s3.GenerateObservation(policy)

	cr.Status.SetConditions(runtimev1alpha1.Available())

	upToDate, err := s3.IsUpToDate(cr.Spec.ForProvider, policy)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*bucketv1beta1.S3Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Creating())

	// Create the bucket
	_, err := e.client.CreateBucketRequest(s3.GenerateCreateBucketInput(meta.GetExternalName(cr), &cr.Spec.ForProvider)).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateBucket)
	}

	// Attach the policy
	if cr.Spec.ForProvider.Policy != nil {
		err = e.putBucketPolicy(cr.Spec.ForProvider.Policy, meta.GetExternalName(cr))
	}

	return managed.ExternalCreation{}, errors.Wrap(err, errPolicyAttach)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*bucketv1beta1.S3Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	policy := e.getBucketPolicy(meta.GetExternalName(cr))

	patch, err := s3.CreatePatch(cr.Spec.ForProvider, &policy)

	if patch != nil && patch.Policy != nil {
		// Update the policy
		if cr.Spec.ForProvider.Policy != nil {
			err = e.putBucketPolicy(patch.Policy, meta.GetExternalName(cr))
		}
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*bucketv1beta1.S3Bucket)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.DeleteBucketRequest(&awsS3.DeleteBucketInput{
		Bucket: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteBucket)
}

func (e *external) getBucketPolicy(bucket string) string {
	o, err := e.client.GetBucketPolicyRequest(&awsS3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	}).Send(ctx)

	if err != nil {
		return ""
	}

	var policy string
	if o.Policy != nil {
		policy = *o.Policy
	}

	return policy
}

func (e *external) putBucketPolicy(policy *string, bucket string) error {
	_, err := e.client.PutBucketPolicyRequest(&awsS3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: policy,
	}).Send(ctx)

	return err
}
