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

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/sns"
	snsclient "github.com/crossplane/provider-aws/pkg/clients/sns"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errSubscriptionPending          = "cannot delete a subscription in PendingConfirmation state"
	errKubeSubscriptionUpdateFailed = "cannot update SNSSubscription custom resource"
	errClient                       = "cannot create a new SNSSubscription client"
	errUnexpectedObject             = "the managed resource is not a SNS Subscription resource"
	errGetSubscriptionAttr          = "failed to get SNS Subscription Attributes"
	errCreate                       = "failed to create the SNS Subscription"
	errDelete                       = "failed to delete the SNS Subscription"
	errUpdate                       = "failed to update the SNS Subscription"
)

// SetupSubscription adds a controller than reconciles SNSSubscription
func SetupSubscription(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.SNSSubscriptionGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.SNSSubscription{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SNSSubscriptionGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:        mgr.GetClient(),
				newClientFn: sns.NewSubscriptionClient,
				awsConfigFn: utils.RetrieveAwsConfigFromProvider,
			}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(*aws.Config) (sns.SubscriptionClient, error)
	awsConfigFn func(context.Context, client.Reader, runtimev1alpha1.Reference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {

	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.kube, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}
	return &external{c, conn.kube}, nil
}

type external struct {
	client snsclient.SubscriptionClient
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
	res, err := e.client.GetSubscriptionAttributesRequest(&awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(sns.IsSubscriptionNotFound, err), errGetSubscriptionAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	snsclient.LateInitializeSubscription(&cr.Spec.ForProvider, res.Attributes)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{},
				errors.Wrap(err, errKubeSubscriptionUpdateFailed)
		}
	}

	// GenerateObservation for SNS Subscription
	cr.Status.AtProvider = snsclient.GenerateSubscriptionObservation(res.Attributes)

	// Set Status for SNS Subcription
	switch *cr.Status.AtProvider.Status {
	case v1alpha1.ConfirmationSuccessful:
		cr.Status.SetConditions(runtimev1alpha1.Available())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	}

	upToDate := snsclient.IsSNSSubscriptionAttributesUpToDate(cr.Spec.ForProvider, res.Attributes)
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	input := snsclient.GenerateSubscribeInput(&cr.Spec.ForProvider)
	res, err := e.client.SubscribeRequest(input).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(res.SubscribeOutput.SubscriptionArn))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errKubeSubscriptionUpdateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Fetch Subscription Attributes again
	resp, err := e.client.GetSubscriptionAttributesRequest(&awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}
	// Update Subscription
	attrs := snsclient.GetChangedSubAttributes(cr.Spec.ForProvider, resp.Attributes)
	for k, v := range attrs {
		_, err := e.client.SetSubscriptionAttributesRequest(&awssns.SetSubscriptionAttributesInput{
			AttributeName:   aws.String(k),
			AttributeValue:  aws.String(v),
			SubscriptionArn: aws.String(meta.GetExternalName(cr)),
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.SNSSubscription)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())
	if meta.GetExternalName(cr) == "" {
		return nil
	}
	_, err := e.client.UnsubscribeRequest(&awssns.UnsubscribeInput{
		SubscriptionArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(sns.IsSubscriptionNotFound, err), errDelete)
}
