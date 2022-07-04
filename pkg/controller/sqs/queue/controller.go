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
	awssqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sqs"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/common"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
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
func SetupQueue(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.QueueGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.Queue{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.QueueGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sqs.NewClient}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithInitializers(common.NewTagger(mgr.GetClient(), &v1beta1.Queue{})),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube        client.Client
	newClientFn func(aws.Config) sqs.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Queue)
	if !ok {
		return nil, errors.New(errNotQueue)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mg.(*v1beta1.Queue)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotQueue)
	}

	// Check the existence of the queue.
	getURLOutput, err := e.client.GetQueueUrl(ctx, &awssqs.GetQueueUrlInput{
		QueueName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil || getURLOutput.QueueUrl == nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(sqs.IsNotFound, err), errGetQueueURLFailed)
	}

	// Get all the attributes.
	resAttributes, err := e.client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl:       getURLOutput.QueueUrl,
		AttributeNames: []awssqstypes.QueueAttributeName{awssqstypes.QueueAttributeName(v1beta1.AttributeAll)},
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(sqs.IsNotFound, err), errGetQueueAttributesFailed)
	}

	resTags, err := e.client.ListQueueTags(ctx, &awssqs.ListQueueTagsInput{
		QueueUrl: getURLOutput.QueueUrl,
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errListQueueTagsFailed)
	}

	sqs.LateInitialize(&cr.Spec.ForProvider, resAttributes.Attributes, resTags.Tags)
	current := cr.Spec.ForProvider.DeepCopy()
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.SetConditions(xpv1.Available())

	cr.Status.AtProvider = sqs.GenerateQueueObservation(*getURLOutput.QueueUrl, resAttributes.Attributes)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  sqs.IsUpToDate(cr.Spec.ForProvider, resAttributes.Attributes, resTags.Tags),
		ConnectionDetails: sqs.GetConnectionDetails(*cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Queue)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotQueue)
	}

	cr.SetConditions(xpv1.Creating())

	resp, err := e.client.CreateQueue(ctx, &awssqs.CreateQueueInput{
		Attributes: sqs.GenerateCreateAttributes(&cr.Spec.ForProvider),
		QueueName:  aws.String(meta.GetExternalName(cr)),
		Tags:       cr.Spec.ForProvider.Tags,
	})
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreateFailed)
	}
	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(*resp.QueueUrl),
	}
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Queue)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotQueue)
	}

	if cr.Status.AtProvider.URL == "" {
		return managed.ExternalUpdate{}, nil
	}

	_, err := e.client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(cr.Status.AtProvider.URL),
		Attributes: sqs.GenerateQueueAttributes(&cr.Spec.ForProvider),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateFailed)
	}

	resTags, err := e.client.ListQueueTags(ctx, &awssqs.ListQueueTagsInput{
		QueueUrl: aws.String(cr.Status.AtProvider.URL),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errListQueueTagsFailed)
	}

	removedTags, addedTags := sqs.TagsDiff(resTags.Tags, cr.Spec.ForProvider.Tags)

	if len(removedTags) > 0 {
		removedKeys := []string{}
		for k := range removedTags {
			removedKeys = append(removedKeys, k)
		}

		_, err = e.client.UntagQueue(ctx, &awssqs.UntagQueueInput{
			QueueUrl: aws.String(cr.Status.AtProvider.URL),
			TagKeys:  removedKeys,
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateFailed)
		}
	}

	if len(addedTags) > 0 {
		_, err = e.client.TagQueue(ctx, &awssqs.TagQueueInput{
			QueueUrl: aws.String(cr.Status.AtProvider.URL),
			Tags:     addedTags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errTag)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Queue)
	if !ok {
		return errors.New(errNotQueue)
	}

	cr.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteQueue(ctx, &awssqs.DeleteQueueInput{
		QueueUrl: aws.String(cr.Status.AtProvider.URL),
	})
	return awsclient.Wrap(resource.Ignore(sqs.IsNotFound, err), errDeleteFailed)
}
