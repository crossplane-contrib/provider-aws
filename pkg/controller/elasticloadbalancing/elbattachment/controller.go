/*
Copyright 2020 The Crossplane Authors.

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

package elbattachment

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awselbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
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

	"github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/elasticloadbalancing/elb"
)

const (
	errUnexpectedObject = "The managed resource is not an ELBAttachment resource"

	errDescribe      = "failed to list instances for given ELB"
	errMultipleItems = "retrieved multiple ELBs for the given name"
	errCreate        = "failed to register instance to ELB"
	errDelete        = "failed to deregister instance from the ELB"
)

// SetupELBAttachment adds a controller that reconciles ELBAttachmets.
func SetupELBAttachment(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ELBAttachmentGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.ELBAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ELBAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elb.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elb.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ELBAttachment)
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
	kube   client.Client
	client elb.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.ELBAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeLoadBalancers(ctx, &awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{cr.Spec.ForProvider.ELBName},
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	var instance string
	for k, v := range observed.Instances {
		if *v.InstanceId == cr.Spec.ForProvider.InstanceID {
			instance = aws.ToString(observed.Instances[k].InstanceId)
		}
	}

	if instance == "" {
		return managed.ExternalObservation{}, nil
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.ELBAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.RegisterInstancesWithLoadBalancer(ctx, &awselb.RegisterInstancesWithLoadBalancerInput{
		Instances:        []awselbtypes.Instance{{InstanceId: aws.String(cr.Spec.ForProvider.InstanceID)}},
		LoadBalancerName: aws.String(cr.Spec.ForProvider.ELBName),
	})

	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ELBAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeregisterInstancesFromLoadBalancer(ctx, &awselb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []awselbtypes.Instance{{InstanceId: aws.String(cr.Spec.ForProvider.InstanceID)}},
		LoadBalancerName: aws.String(cr.Spec.ForProvider.ELBName),
	})

	return awsclient.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}
