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
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	errKubeTopicUpdateFailed = "cannot update SNSTopic custom resource"
	errClient                = "cannot create a new SNSTopic client"
	errUnexpectedObject      = "The managed resource is not a SNSTopic resource"
	errList                  = "failed to list SNS Topics"
	errGetTopic              = "failed to get SNS Topic"
	errListTopic             = "failed to list SNS Topics"
	errGetTopicAttr          = "failed to get SNS Topic Attribute"
	errCreate                = "failed to create the SNS Topic"
	errDelete                = "failed to delete the SNS Topic"
	errUpdate                = "failed to update the SNS Topic"
)

// SetupSNSTopic adds a controller that reconciles SNSTopic.
func SetupSNSTopic(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.SNSTopicGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.SNSTopic{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SNSTopicGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:        mgr.GetClient(),
				newClientFn: sns.NewTopicClient,
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
	newClientFn func(*aws.Config) (sns.TopicClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {

	cr, ok := mgd.(*v1alpha1.SNSTopic)
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
	client snsclient.TopicClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.SNSTopic)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	// AWS SNS resources are uniquely identified by an ARN that is returned
	// on create time; we can't tell whether they exist unless we have recorded
	// their ARN.
	if !awsarn.IsARN(meta.GetExternalName(cr)) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Fetch the list of Topics
	topicList, err := e.client.ListTopicsRequest(&awssns.ListTopicsInput{}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(sns.IsErrorTopicNotFound, err), errList)
	}

	// Filters the list of topic with matching values in CR
	topic, err := snsclient.GetSNSTopic(topicList, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(sns.IsErrorTopicNotFound, err), errGetTopic)
	}

	// Fetch SNS Topic Attributes with matching TopicARN
	res, err := e.client.GetTopicAttributesRequest(&awssns.GetTopicAttributesInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetTopicAttr)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	snsclient.LateInitializeTopic(&cr.Spec.ForProvider, topic, res.Attributes)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeTopicUpdateFailed)
		}
	}

	cr.SetConditions(runtimev1alpha1.Available())

	// GenerateObservation for SNS Topic
	cr.Status.AtProvider = snsclient.GenerateTopicObservation(res.Attributes)

	upToDate := snsclient.IsSNSTopicUpToDate(cr.Spec.ForProvider, res.Attributes)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha1.SNSTopic)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	resp, err := e.client.CreateTopicRequest(snsclient.GenerateCreateTopicInput(&cr.Spec.ForProvider)).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(resp.CreateTopicOutput.TopicArn))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errKubeTopicUpdateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.SNSTopic)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Fetch Topic Attributes again
	resp, err := e.client.GetTopicAttributesRequest(&awssns.GetTopicAttributesInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	// Update Topic Attributes
	attrs := snsclient.GetChangedAttributes(cr.Spec.ForProvider, resp.Attributes)
	for k, v := range attrs {
		_, err := e.client.SetTopicAttributesRequest(&awssns.SetTopicAttributesInput{
			AttributeName:  aws.String(k),
			AttributeValue: aws.String(v),
			TopicArn:       aws.String(meta.GetExternalName(cr)),
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.SNSTopic)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteTopicRequest(&awssns.DeleteTopicInput{
		TopicArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(err, errDelete)
}
