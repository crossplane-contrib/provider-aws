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

package snstopic

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	notificationv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/notification/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	notclient "github.com/crossplane-contrib/provider-aws/pkg/clients/notification"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUnexpectedObject = "the managed resource is not a SNSTopic resource"
	errGetTopicAttr     = "failed to get SNS Topic Attribute"
	errCreate           = "failed to create the SNS Topic"
	errDelete           = "failed to delete the SNS Topic"
	errUpdate           = "failed to update the SNS Topic"
)

// SetupSNSTopic adds a controller that reconciles SNSTopic.
func SetupSNSTopic(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(notificationv1alpha1.SNSTopicGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&notificationv1alpha1.SNSTopic{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(notificationv1alpha1.SNSTopicGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sns.NewTopicClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) sns.TopicClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*notificationv1alpha1.SNSTopic)
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
	client notclient.TopicClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*notificationv1alpha1.SNSTopic)
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
	notclient.LateInitializeTopicAttr(&cr.Spec.ForProvider, res.Attributes)

	cr.SetConditions(xpv1.Available())

	// GenerateObservation for SNS Topic
	cr.Status.AtProvider = notclient.GenerateTopicObservation(res.Attributes)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        notclient.IsSNSTopicUpToDate(cr.Spec.ForProvider, res.Attributes),
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*notificationv1alpha1.SNSTopic)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	resp, err := e.client.CreateTopic(ctx, notclient.GenerateCreateTopicInput(&cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(resp.TopicArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*notificationv1alpha1.SNSTopic)
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
	attrs := notclient.GetChangedAttributes(cr.Spec.ForProvider, resp.Attributes)
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
	cr, ok := mgd.(*notificationv1alpha1.SNSTopic)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteTopic(ctx, &awssns.DeleteTopicInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(sns.IsTopicNotFound, err), errDelete)
}
