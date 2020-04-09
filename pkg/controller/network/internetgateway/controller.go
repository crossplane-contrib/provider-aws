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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1beta1 "github.com/crossplane/provider-aws/apis/network/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errClient            = "cannot create a new InternetGatewayClient"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errUnexpectedObject    = "The managed resource is not an InternetGateway resource"
	errDescribe            = "failed to describe InternetGateway"
	errNotSingleItem       = "either no or multiple InternetGateways retrieved for the given internetGatewayId"
	errMultipleAttachments = "multiple Attachments retrieved for the given internetGatewayId"
	errCreate              = "failed to create the InternetGateway resource"
	errDetach              = "failed to detach the InternetGateway from VPC"
	errDelete              = "failed to delete the InternetGateway resource"
	errUpdate              = "failed to update the InternetGateway resource"
	errSpecUpdate          = "cannot update spec"
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
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewInternetGatewayClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ec2.InternetGatewayClient, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	p := &awsv1alpha3.Provider{}
	if err := conn.client.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		igClient, err := conn.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: igClient, kube: conn.client}, errors.Wrap(err, errClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := conn.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	igClient, err := conn.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: igClient, kube: conn.client}, errors.Wrap(err, errClient)
}

type external struct {
	kube   client.Client
	client ec2.InternetGatewayClient
	kube   client.Client
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
		return managed.ExternalObservation{}, errors.Wrapf(resource.Ignore(ec2.IsInternetGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.InternetGateways) != 1 {
		return managed.ExternalObservation{}, errors.Errorf(errNotSingleItem)
	}

	cr.SetConditions(runtimev1alpha1.Available())

	observed := response.InternetGateways[0]

	cr.Status.AtProvider = ec2.GenerateIGObservation(observed)

	isUptoDate := ec2.IsIgUpToDate(cr.Spec.ForProvider, observed)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isUptoDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.InternetGateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	ig, err := e.client.CreateInternetGatewayRequest(&awsec2.CreateInternetGatewayInput{}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(ig.InternetGateway.InternetGatewayId))

	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errSpecUpdate)
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

	// There can only be one attachmemt and if that is attached to
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

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

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
