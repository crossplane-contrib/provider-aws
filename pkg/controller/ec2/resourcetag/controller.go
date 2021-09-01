/*
Copyright 2021 The Crossplane Authors.

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

package resourcetag

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	ec2client "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not a ResourceTag resource"

	errDescribe = "failed to describe tags"
	errCreate   = "failed to create the ResourceTag resource"
	errDelete   = "failed to delete the ResourceTag resource"
	errUpdate   = "failed to update the ResourceTag resource"
)

// SetupResourceTag adds a controller that reconciles ResourceTags.
func SetupResourceTag(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(manualv1alpha1.ResourceTagGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&manualv1alpha1.ResourceTag{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(manualv1alpha1.ResourceTagGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2client.NewResourceTagClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2client.ResourceTagClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.ResourceTag)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.StringValue(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}

	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2client.ResourceTagClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*manualv1alpha1.ResourceTag)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	tags, err := e.listTags(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	existingTags, missingTags := diffTags(cr, tags)

	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   len(existingTags) > 0,
		ResourceUpToDate: len(missingTags) == 0,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*manualv1alpha1.ResourceTag)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	if err := e.updateTags(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*manualv1alpha1.ResourceTag)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	if err := e.updateTags(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*manualv1alpha1.ResourceTag)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	tags := make([]awsec2.Tag, len(cr.Spec.ForProvider.Tags))
	for i, t := range cr.Spec.ForProvider.Tags {
		tags[i] = awsec2.Tag{Key: awsclient.String(t.Key)}
	}

	_, err := e.client.DeleteTagsRequest(&awsec2.DeleteTagsInput{
		Resources: cr.Spec.ForProvider.ResourceIDs,
		Tags:      tags,
	}).Send(ctx)

	return errors.Wrap(err, errDelete)
}

func (e *external) updateTags(ctx context.Context, cr *manualv1alpha1.ResourceTag) error {
	tags := make([]awsec2.Tag, len(cr.Spec.ForProvider.Tags))
	for i, t := range cr.Spec.ForProvider.Tags {
		tags[i] = awsec2.Tag{Key: awsclient.String(t.Key), Value: awsclient.String(t.Value)}
	}

	_, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
		Resources: cr.Spec.ForProvider.ResourceIDs,
		Tags:      tags,
	}).Send(ctx)

	return err
}

func (e *external) listTags(ctx context.Context, cr *manualv1alpha1.ResourceTag) ([]awsec2.TagDescription, error) {
	res, err := e.client.DescribeTagsRequest(&awsec2.DescribeTagsInput{
		Filters: []awsec2.Filter{
			{Name: awsclient.String("resource-id"), Values: cr.Spec.ForProvider.ResourceIDs},
		},
	}).Send(ctx)

	if err != nil {
		return nil, errors.Wrap(err, errDescribe)
	}

	return res.Tags, nil
}
