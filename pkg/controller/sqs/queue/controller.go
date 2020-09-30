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

package queue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/sqs/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/sqs"
)

const (
	errNotQueue                 = "managed resource is not a Queue custom resource"
	errKubeUpdateFailed         = "cannot update Queue custom resource"
	errCreateFailed             = "cannot create Queue"
	errDeleteFailed             = "cannot delete Queue"
	errGetQueueAttributesFailed = "cannot get Queue attributes"
	errTag                      = "cannot tag Queue"
	errGetQueueURLFailed        = "cannot get Queue URL"
	errListQueueTagsFailed      = "cannot list Queue tags"
	errUpdateFailed             = "failed to update the Queue resource"
)

// SetupQueue adds a controller that reconciles Queue.
func SetupQueue(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.QueueGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Queue{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.QueueGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sqs.NewClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(aws.Config) sqs.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return nil, errors.New(errNotQueue)
	}
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(*cfg), c.kube}, nil
}

type external struct {
	client sqs.Client
	kube   client.Client
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
		AttributeNames: []awssqs.QueueAttributeName{awssqs.QueueAttributeName(v1alpha1.AttributeAll)},
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

	cr.Status.AtProvider = sqs.GenerateQueueObservation(*getURLResponse.QueueUrl, resAttributes.Attributes)

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

	_, err := e.client.CreateQueueRequest(&awssqs.CreateQueueInput{
		Attributes: sqs.GenerateCreateAttributes(&cr.Spec.ForProvider),
		QueueName:  aws.String(meta.GetExternalName(cr)),
		Tags:       cr.Spec.ForProvider.Tags,
	}).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
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
		Attributes: sqs.GenerateQueueAttributes(&cr.Spec.ForProvider),
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
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTag)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Queue)
	if !ok {
		return errors.New(errNotQueue)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteQueueRequest(&awssqs.DeleteQueueInput{
		QueueUrl: aws.String(cr.Status.AtProvider.URL),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(sqs.IsNotFound, err), errDeleteFailed)
}
