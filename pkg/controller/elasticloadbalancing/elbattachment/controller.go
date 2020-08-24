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

	"github.com/aws/aws-sdk-go-v2/aws"
	awselb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
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
func SetupELBAttachment(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ELBAttachmentGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ELBAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ELBAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elb.NewClient, awsConfigFn: awscommon.GetConfig}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elb.Client
	awsConfigFn func(context.Context, client.Client, resource.Managed, string) (*aws.Config, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := c.awsConfigFn(ctx, c.kube, mg, "")
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

	response, err := e.client.DescribeLoadBalancersRequest(&awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{cr.Spec.ForProvider.ELBName},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	var instance string
	for k, v := range observed.Instances {
		if *v.InstanceId == cr.Spec.ForProvider.InstanceID {
			instance = aws.StringValue(observed.Instances[k].InstanceId)
		}
	}

	if instance == "" {
		return managed.ExternalObservation{}, nil
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

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

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.client.RegisterInstancesWithLoadBalancerRequest(&awselb.RegisterInstancesWithLoadBalancerInput{
		Instances:        []awselb.Instance{{InstanceId: aws.String(cr.Spec.ForProvider.InstanceID)}},
		LoadBalancerName: aws.String(cr.Spec.ForProvider.ELBName),
	}).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ELBAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeregisterInstancesFromLoadBalancerRequest(&awselb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []awselb.Instance{{InstanceId: aws.String(cr.Spec.ForProvider.InstanceID)}},
		LoadBalancerName: aws.String(cr.Spec.ForProvider.ELBName),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}
