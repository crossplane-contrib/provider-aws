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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
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

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
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
func SetupInternetGateway(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.InternetGatewayGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.InternetGateway{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.InternetGatewayGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewInternetGatewayClient}),
			managed.WithCreationGracePeriod(3*time.Minute),
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
	newClientFn func(config aws.Config) ec2.InternetGatewayClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.InternetGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
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

	response, err := e.client.DescribeInternetGateways(ctx, &awsec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDescribe)
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

	ig, err := e.client.CreateInternetGateway(ctx, &awsec2.CreateInternetGatewayInput{})
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(ig.InternetGateway.InternetGatewayId))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// Tagging the created InternetGateway
	if len(cr.Spec.ForProvider.Tags) > 0 {
		if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errCreateTags)
		}
	}

	response, err := e.client.DescribeInternetGateways(ctx, &awsec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDescribe)
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
		aws.ToString(observed.Attachments[0].VpcId) != aws.ToString(cr.Spec.ForProvider.VPCID) {
		if _, err = e.client.DetachInternetGateway(ctx, &awsec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(meta.GetExternalName(cr)),
			VpcId:             observed.Attachments[0].VpcId,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errDetach)
		}
	}

	// Attach IG to VPC in spec.
	_, err = e.client.AttachInternetGateway(ctx, &awsec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(meta.GetExternalName(cr)),
		VpcId:             cr.Spec.ForProvider.VPCID,
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ec2.IsInternetGatewayAlreadyAttached, err), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// first detach all vpc attachments
	for _, a := range cr.Status.AtProvider.Attachments {
		_, err := e.client.DetachInternetGateway(ctx, &awsec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(meta.GetExternalName(cr)),
			VpcId:             aws.String(a.VPCID),
		})

		if resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err) == nil {
			continue
		}
		return awsclient.Wrap(err, errDetach)
	}

	// now delete the IG
	_, err := e.client.DeleteInternetGateway(ctx, &awsec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDelete)
}
