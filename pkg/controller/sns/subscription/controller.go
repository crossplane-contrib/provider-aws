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

package subscription

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
	errUnexpectedObject    = "the managed resource is not a SNS Subscription resource"
	errGetSubscriptionAttr = "failed to get SNS Subscription Attributes"
	errCreate              = "failed to create the SNS Subscription"
	errDelete              = "failed to delete the SNS Subscription"
	errUpdate              = "failed to update the SNS Subscription"
)

// SetupSubscription adds a controller than reconciles Subscription
func SetupSubscription(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.SubscriptionGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: sns.NewSubscriptionClient}),
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
		resource.ManagedKind(v1beta1.SubscriptionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Subscription{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) sns.SubscriptionClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Subscription)
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
	client snsclient.SubscriptionClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.Subscription)
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
			errorutils.Wrap(resource.Ignore(sns.IsSubscriptionNotFound, err), errGetSubscriptionAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	snsclient.LateInitializeSubscription(&cr.Spec.ForProvider, res.Attributes)

	// GenerateObservation for SNS Subscription
	cr.Status.AtProvider = snsclient.GenerateSubscriptionObservation(res.Attributes)

	// Set Status for SNS Subcription
	switch *cr.Status.AtProvider.Status { //nolint:exhaustive
	case v1beta1.ConfirmationSuccessful:
		cr.Status.SetConditions(xpv1.Available())
	default:
		cr.Status.SetConditions(xpv1.Creating())
	}

	upToDate := snsclient.IsSNSSubscriptionAttributesUpToDate(cr.Spec.ForProvider, res.Attributes)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Subscription)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	input := snsclient.GenerateSubscribeInput(&cr.Spec.ForProvider)
	res, err := e.client.Subscribe(ctx, input)

	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(res.SubscriptionArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Subscription)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Fetch Subscription Attributes again
	resp, err := e.client.GetSubscriptionAttributes(ctx, &awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}
	// Update Subscription
	attrs := snsclient.GetChangedSubAttributes(cr.Spec.ForProvider, resp.Attributes)
	for k, v := range attrs {
		_, err := e.client.SetSubscriptionAttributes(ctx, &awssns.SetSubscriptionAttributesInput{
			AttributeName:   aws.String(k),
			AttributeValue:  aws.String(v),
			SubscriptionArn: aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Subscription)
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
	return errorutils.Wrap(resource.Ignore(sns.IsSubscriptionNotFound, err), errDelete)
}
