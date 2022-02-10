/*
Copyright 2022 The Crossplane Authors.

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

package topic

import (
	"context"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/pkg/errors"
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

	"github.com/crossplane/provider-aws/apis/sns/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/sns"
	snsclient "github.com/crossplane/provider-aws/pkg/clients/sns"
)

const (
	errUnexpectedObject = "the managed resource is not a Topic resource"
	errGetTopicAttr     = "failed to get SNS Topic Attribute"
	errCreate           = "failed to create the SNS Topic"
	errDelete           = "failed to delete the SNS Topic"
	errUpdate           = "failed to update the SNS Topic"
)

// SetupSNSTopic adds a controller that reconciles Topic.
func SetupSNSTopic(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.TopicGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1beta1.Topic{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.TopicGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sns.NewTopicClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) sns.TopicClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Topic)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client snsclient.TopicClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.Topic)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Fetch SNS Topic Attributes with matching TopicARN
	res, err := e.client.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{},
			awsclient.Wrap(resource.Ignore(sns.IsTopicNotFound, err), errGetTopicAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	snsclient.LateInitializeTopicAttr(&cr.Spec.ForProvider, res.Attributes)

	cr.SetConditions(xpv1.Available())

	// GenerateObservation for SNS Topic
	cr.Status.AtProvider = snsclient.GenerateTopicObservation(res.Attributes)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        snsclient.IsSNSTopicUpToDate(cr.Spec.ForProvider, res.Attributes),
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1beta1.Topic)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	resp, err := e.client.CreateTopic(ctx, snsclient.GenerateCreateTopicInput(&cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(resp.TopicArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Topic)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Fetch Topic Attributes again
	resp, err := e.client.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errGetTopicAttr)
	}

	// Update Topic Attributes
	attrs := snsclient.GetChangedAttributes(cr.Spec.ForProvider, resp.Attributes)
	for k, v := range attrs {
		_, err = e.client.SetTopicAttributes(ctx, &awssns.SetTopicAttributesInput{
			AttributeName:  aws.String(k),
			AttributeValue: aws.String(v),
			TopicArn:       aws.String(meta.GetExternalName(cr)),
		})

	}
	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Topic)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteTopic(ctx, &awssns.DeleteTopicInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(sns.IsTopicNotFound, err), errDelete)
}
