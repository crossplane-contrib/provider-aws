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

	v1alpha3 "github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject    = "The managed resource is not an InternetGateway resource"
	errClient              = "cannot create a new InternetGatewayClient"
	errDescribe            = "failed to describe InternetGateway"
	errMultipleItems       = "retrieved multiple InternetGateways for the given internetGatewaysId"
	errCreate              = "failed to create the InternetGateway resource"
	errPersistExternalName = "failed to persist InternetGateway ID"
	errDetach              = "failed to detach the InternetGateway from VPC"
	errDelete              = "failed to delete the InternetGateway resource"
)

// SetupInternetGateway adds a controller that reconciles InternetGateways.
func SetupInternetGateway(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.InternetGatewayGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.InternetGateway{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.InternetGatewayGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewInternetGatewayClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (ec2.InternetGatewayClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.InternetGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}

	return &external{kube: conn.client, client: c}, nil
}

type external struct {
	kube   client.Client
	client ec2.InternetGatewayClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.InternetGateway)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// AWS network resources are uniquely identified by an ID that is returned
	// on create time; we can't tell whether they exist unless we have recorded
	// their ID.
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	req := e.client.DescribeInternetGatewaysRequest(&awsec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{meta.GetExternalName(cr)},
	})

	response, err := req.Send(ctx)
	if ec2.IsInternetGatewayNotFoundErr(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.InternetGateways) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.InternetGateways[0]

	// if non of the attachments are currently in progress, then the IG is available
	isAvailable := true
	for _, a := range observed.Attachments {
		if a.State == awsec2.AttachmentStatusAttaching || a.State == awsec2.AttachmentStatusDetaching {
			isAvailable = false
			break
		}
	}

	if isAvailable {
		cr.SetConditions(runtimev1alpha1.Available())
	}

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.InternetGateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.CreateInternetGatewayRequest(&awsec2.CreateInternetGatewayInput{})

	rsp, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(rsp.InternetGateway.InternetGatewayId))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	cr.UpdateExternalStatus(*rsp.InternetGateway)

	// after creating the IG, attach the VPC
	aReq := e.client.AttachInternetGatewayRequest(&awsec2.AttachInternetGatewayInput{
		InternetGatewayId: rsp.InternetGateway.InternetGatewayId,
		VpcId:             aws.String(cr.Spec.VPCID),
	})

	_, err = aReq.Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.InternetGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	// first detach all vpc attachments
	for _, a := range cr.Status.Attachments {
		// after creating the IG, attach the VPC
		dReq := e.client.DetachInternetGatewayRequest(&awsec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(meta.GetExternalName(cr)),
			VpcId:             aws.String(a.VPCID),
		})

		if _, err := dReq.Send(ctx); err != nil {
			if ec2.IsInternetGatewayNotFoundErr(err) {
				continue
			}
			return errors.Wrap(err, errDetach)
		}
	}

	// now delete the IG
	req := e.client.DeleteInternetGatewayRequest(&awsec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(meta.GetExternalName(cr)),
	})

	_, err := req.Send(ctx)
	return errors.Wrap(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDelete)
}
