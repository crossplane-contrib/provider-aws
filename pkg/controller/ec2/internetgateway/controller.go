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

package internetgateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject    = "The managed resource is not an InternetGateway resource"
	errDescribe            = "failed to describe InternetGateway"
	errNotSingleItem       = "either no or multiple InternetGateways retrieved for the given internetGatewayId"
	errMultipleAttachments = "multiple Attachments retrieved for the given internetGatewayId"
	errCreate              = "failed to create the InternetGateway resource"
	errDetach              = "failed to detach the InternetGateway from VPC"
	errDelete              = "failed to delete the InternetGateway resource"
	errUpdate              = "failed to update the InternetGateway resource"
	errStatusUpdate        = "cannot update status of the InternetGateway resource"
	errCreateTags          = "failed to create tags for the InternetGateway resource"
)

// SetupInternetGateway adds a controller that reconciles InternetGateways.
func SetupInternetGateway(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.InternetGatewayGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.InternetGateway{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.InternetGatewayGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewInternetGatewayClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.InternetGatewayClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.InternetGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, aws.StringValue(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.InternetGatewayClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeInternetGatewaysRequest(&awsec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.InternetGateways) != 1 {
		return managed.ExternalObservation{}, errors.New(errNotSingleItem)
	}

	observed := response.InternetGateways[0]

	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeIG(&cr.Spec.ForProvider, &observed)

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = ec2.GenerateIGObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsIgUpToDate(cr.Spec.ForProvider, observed),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	ig, err := e.client.CreateInternetGatewayRequest(&awsec2.CreateInternetGatewayInput{}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(ig.InternetGateway.InternetGatewayId))

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Tagging the created InternetGateway
	if len(cr.Spec.ForProvider.Tags) > 0 {
		if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
		}
	}

	response, err := e.client.DescribeInternetGatewaysRequest(&awsec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDescribe)
	}

	if len(response.InternetGateways) != 1 {
		return managed.ExternalUpdate{}, errors.New(errNotSingleItem)
	}

	observed := response.InternetGateways[0]

	// There can only be one attachment and if that is attached to
	// spec.VpcID, no action is required.
	if len(observed.Attachments) > 1 {
		return managed.ExternalUpdate{}, errors.New(errMultipleAttachments)
	}

	if len(observed.Attachments) == 1 &&
		aws.StringValue(observed.Attachments[0].VpcId) != aws.StringValue(cr.Spec.ForProvider.VPCID) {
		if _, err = e.client.DetachInternetGatewayRequest(&awsec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(meta.GetExternalName(cr)),
			VpcId:             observed.Attachments[0].VpcId,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errDetach)
		}
	}

	// Attach IG to VPC in spec.
	_, err = e.client.AttachInternetGatewayRequest(&awsec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(meta.GetExternalName(cr)),
		VpcId:             cr.Spec.ForProvider.VPCID,
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(ec2.IsInternetGatewayAlreadyAttached, err), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// first detach all vpc attachments
	for _, a := range cr.Status.AtProvider.Attachments {
		_, err := e.client.DetachInternetGatewayRequest(&awsec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(meta.GetExternalName(cr)),
			VpcId:             aws.String(a.VPCID),
		}).Send(ctx)

		if resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err) == nil {
			continue
		}
		return errors.Wrap(err, errDetach)
	}

	// now delete the IG
	_, err := e.client.DeleteInternetGatewayRequest(&awsec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDelete)
}
