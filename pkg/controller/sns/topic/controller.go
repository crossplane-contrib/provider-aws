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

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns"
	snsclient "github.com/crossplane-contrib/provider-aws/pkg/clients/sns"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "the managed resource is not a Topic resource"
	errGetTopicAttr     = "failed to get SNS Topic Attribute"
	errCreate           = "failed to create the SNS Topic"
	errDelete           = "failed to delete the SNS Topic"
	errUpdate           = "failed to update the SNS Topic"
	errGetChangedAttr   = "failed to get changed topic attributes"
)

// SetupSNSTopic adds a controller that reconciles Topic.
func SetupSNSTopic(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.TopicGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sns.NewTopicClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.TopicGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Topic{}).
		Complete(r)
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
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
			errorutils.Wrap(resource.Ignore(sns.IsTopicNotFound, err), errGetTopicAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	snsclient.LateInitializeTopicAttr(&cr.Spec.ForProvider, res.Attributes)

	cr.SetConditions(xpv1.Available())

	// GenerateObservation for SNS Topic
	cr.Status.AtProvider = snsclient.GenerateTopicObservation(res.Attributes)

	upToDate, err := snsclient.IsSNSTopicUpToDate(cr.Spec.ForProvider, res.Attributes)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
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
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
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
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errGetTopicAttr)
	}

	// Update Topic Attributes
	attrs, err := snsclient.GetChangedAttributes(cr.Spec.ForProvider, resp.Attributes)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetChangedAttr)
	}
	for k, v := range attrs {
		_, err = e.client.SetTopicAttributes(ctx, &awssns.SetTopicAttributesInput{
			AttributeName:  aws.String(k),
			AttributeValue: aws.String(v),
			TopicArn:       aws.String(meta.GetExternalName(cr)),
		})
	}
	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
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

	return errorutils.Wrap(resource.Ignore(sns.IsTopicNotFound, err), errDelete)
}
