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

package sqs

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/applicationintegration/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/sqs"
)

const (
	errNotQueue                 = "managed resource is not a Queue custom resource"
	errKubeUpdateFailed         = "cannot update Queue custom resource"
	errQueueClient              = "cannot create Queue client"
	errGetProvider              = "cannot get provider"
	errGetProviderSecret        = "cannot get provider secret"
	errCreateFailed             = "cannot create Queue"
	errInvalidNameForFifoQueue  = "cannot create Queue, FIFO queue name must have .fifo suffix"
	errDeleteFailed             = "cannot delete Queue"
	errGetQueueAttributesFailed = "cannot get Queue attributes"
	errGetQueueURLFailed        = "cannot get Queue URL"
	errListQueueTagsFailed      = "cannot list Queue tags"
	errUpdateFailed             = "failed to update the Queue resource"
	fifoQueueSuffix             = ".fifo"
)

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (sqs.Client, error)
}

type external struct {
	client sqs.Client
	kube   client.Client
}

// SetupQueue adds a controller that reconciles Queue.
func SetupQueue(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.QueueGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Queue{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.QueueGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sqs.NewClient}),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return nil, errors.New(errNotQueue)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		queueClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: queueClient, kube: c.kube}, errors.Wrap(err, errQueueClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	queueClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: queueClient, kube: c.kube}, errors.Wrap(err, errQueueClient)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotQueue)
	}

	// Check the existence of the queue.
	getURLResponse, err := e.client.GetQueueUrlRequest(&awssqs.GetQueueUrlInput{
		QueueName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil || getURLResponse.GetQueueUrlOutput.QueueUrl == nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(sqs.IsNotFound, err), errGetQueueURLFailed)
	}

	// Get all the attributes.
	resAttributes, err := e.client.GetQueueAttributesRequest(&awssqs.GetQueueAttributesInput{
		QueueUrl:       getURLResponse.QueueUrl,
		AttributeNames: []awssqs.QueueAttributeName{awssqs.QueueAttributeNameAll},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(sqs.IsNotFound, err), errGetQueueAttributesFailed)
	}

	resTags, err := e.client.ListQueueTagsRequest(&awssqs.ListQueueTagsInput{
		QueueUrl: getURLResponse.QueueUrl,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errListQueueTagsFailed)
	}

	sqs.LateInitialize(&cr.Spec.ForProvider, resAttributes.Attributes, resTags.Tags)
	current := cr.Spec.ForProvider.DeepCopy()
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = v1alpha1.QueueObservation{
		ARN: resAttributes.Attributes[v1alpha1.AttributeQueueArn],
		URL: *getURLResponse.QueueUrl,
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: sqs.IsUpToDate(cr.Spec.ForProvider, resAttributes.Attributes, resTags.Tags),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotQueue)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	// FIFO queues names should end with ".fifo"
	if aws.BoolValue(cr.Spec.ForProvider.FIFOQueue) && !strings.HasSuffix(meta.GetExternalName(cr), fifoQueueSuffix) {
		return managed.ExternalCreation{}, errors.New(errInvalidNameForFifoQueue)
	}

	createResp, err := e.client.CreateQueueRequest(&awssqs.CreateQueueInput{
		Attributes: sqs.GenerateCreateAttributes(&cr.Spec.ForProvider),
		QueueName:  aws.String(meta.GetExternalName(cr)),
		Tags:       sqs.GenerateQueueTags(cr.Spec.ForProvider.Tags),
	}).Send(ctx)
	if err != nil || createResp.CreateQueueOutput.QueueUrl == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotQueue)
	}

	if cr.Status.AtProvider.URL == "" {
		return managed.ExternalUpdate{}, nil
	}

	_, err := e.client.SetQueueAttributesRequest(&awssqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(cr.Status.AtProvider.URL),
		Attributes: sqs.GenerateUpdateAttributes(&cr.Spec.ForProvider),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	resTags, err := e.client.ListQueueTagsRequest(&awssqs.ListQueueTagsInput{
		QueueUrl: aws.String(cr.Status.AtProvider.URL),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errListQueueTagsFailed)
	}

	removedTags, addedTags := sqs.TagsDiff(resTags.Tags, cr.Spec.ForProvider.Tags)

	if len(removedTags) > 0 {
		removedKeys := []string{}
		for k := range removedTags {
			removedKeys = append(removedKeys, k)
		}

		_, err = e.client.UntagQueueRequest(&awssqs.UntagQueueInput{
			QueueUrl: aws.String(cr.Status.AtProvider.URL),
			TagKeys:  removedKeys,
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
		}
	}

	if len(addedTags) > 0 {
		_, err = e.client.TagQueueRequest(&awssqs.TagQueueInput{
			QueueUrl: aws.String(cr.Status.AtProvider.URL),
			Tags:     addedTags,
		}).Send(ctx)
	}
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return errors.New(errNotQueue)
	}

	if cr.Status.AtProvider.URL == "" {
		return nil
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteQueueRequest(&awssqs.DeleteQueueInput{
		QueueUrl: aws.String(cr.Status.AtProvider.URL),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(sqs.IsNotFound, err), errDeleteFailed)
}
