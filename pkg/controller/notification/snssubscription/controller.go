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

package snssubscription

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

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	notclient "github.com/crossplane/provider-aws/pkg/clients/notification"
)

const (
	errUnexpectedObject    = "the managed resource is not a SNS Subscription resource"
	errGetSubscriptionAttr = "failed to get SNS Subscription Attributes"
	errCreate              = "failed to create the SNS Subscription"
	errDelete              = "failed to delete the SNS Subscription"
	errUpdate              = "failed to update the SNS Subscription"
)

// SetupSubscription adds a controller than reconciles SNSSubscription
func SetupSubscription(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.SNSSubscriptionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.SNSSubscription{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SNSSubscriptionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: notclient.NewSubscriptionClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) notclient.SubscriptionClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.SNSSubscription)
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
	client notclient.SubscriptionClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Fetch Subscription Attributes with matching SubscriptionARN
	res, err := e.client.GetSubscriptionAttributes(ctx, &awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{},
			awsclient.Wrap(resource.Ignore(notclient.IsSubscriptionNotFound, err), errGetSubscriptionAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	notclient.LateInitializeSubscription(&cr.Spec.ForProvider, res.Attributes)

	// GenerateObservation for SNS Subscription
	cr.Status.AtProvider = notclient.GenerateSubscriptionObservation(res.Attributes)

	// Set Status for SNS Subcription
	switch *cr.Status.AtProvider.Status { //nolint:exhaustive
	case v1alpha1.ConfirmationSuccessful:
		cr.Status.SetConditions(xpv1.Available())
	default:
		cr.Status.SetConditions(xpv1.Creating())
	}

	upToDate := notclient.IsSNSSubscriptionAttributesUpToDate(cr.Spec.ForProvider, res.Attributes)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	input := notclient.GenerateSubscribeInput(&cr.Spec.ForProvider)
	res, err := e.client.Subscribe(ctx, input)

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(res.SubscriptionArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Fetch Subscription Attributes again
	resp, err := e.client.GetSubscriptionAttributes(ctx, &awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
	}
	// Update Subscription
	attrs := notclient.GetChangedSubAttributes(cr.Spec.ForProvider, resp.Attributes)
	for k, v := range attrs {
		_, err := e.client.SetSubscriptionAttributes(ctx, &awssns.SetSubscriptionAttributesInput{
			AttributeName:   aws.String(k),
			AttributeValue:  aws.String(v),
			SubscriptionArn: aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.SetConditions(xpv1.Deleting())
	if meta.GetExternalName(cr) == "" {
		return nil
	}
	_, err := e.client.Unsubscribe(ctx, &awssns.UnsubscribeInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	})
	return awsclient.Wrap(resource.Ignore(notclient.IsSubscriptionNotFound, err), errDelete)
}
