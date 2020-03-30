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

package routetable

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
	errUnexpectedObject = "The managed resource is not an RouteTable resource"

	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errClient             = "cannot create a new RouteTable client"
	errDescribe           = "failed to describe RouteTable"
	errMultipleItems      = "retrieved multiple RouteTables for the given routeTableId"
	errCreate             = "failed to create the RouteTable resource"
	errUpdate             = "failed to update the RouteTable"
	errUpdateNotFound     = "cannot update the RouteTable, since the RouteTableID is not present"
	errDeleteNotPresent   = "cannot delete the RouteTable, since the RouteTableID is not present"
	errDelete             = "failed to delete the RouteTable resource"
	errCreateRoute        = "failed to create a route in the RouteTable resource"
	errAssociateSubnet    = "failed to associate subnet %v to the RouteTable resource"
	errDisassociateSubnet = "failed to disassociate subnet %v from the RouteTable resource"
	errStatusUpdate       = "cannot update status"
	errSpecUpdate         = "cannot update spec"
)

// SetupRouteTable adds a controller that reconciles RouteTables.
func SetupRouteTable(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.RouteTableGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.RouteTable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.RouteTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewRouteTableClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ec2.RouteTableClient, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RouteTable)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		rtClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: rtClient, kube: c.client}, errors.Wrap(err, errUnexpectedObject)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: rdsClient, kube: c.client}, errors.Wrap(err, errClient)
}

type external struct {
	kube   client.Client
	client ec2.RouteTableClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// To find out whether a RouteTable exist:
	// - the object's ExternalState should have routeTableId populated
	// - a RouteTable with the given routeTableId should exist
	if cr.Status.AtProvider.RouteTableID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeRouteTablesRequest(&awsec2.DescribeRouteTablesInput{
		RouteTableIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrapf(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.RouteTables) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.RouteTables[0]

	stateAvailable := true
	for _, rt := range observed.Routes {
		if rt.State != awsec2.RouteStateActive {
			stateAvailable = false
			break
		}
	}
	if stateAvailable {
		cr.SetConditions(runtimev1alpha1.Available())
	}

	cr.Status.AtProvider = ec2.GenerateRTObservation(observed)

	upToDate, err := ec2.IsRtUpToDate(cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	result, err := e.client.CreateRouteTableRequest(&awsec2.CreateRouteTableInput{
		VpcId: aws.String(cr.Spec.ForProvider.VPCID),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	if result.RouteTable == nil {
		return managed.ExternalCreation{}, errors.New(errCreate)
	}

	cr.Status.AtProvider = ec2.GenerateRTObservation(*result.RouteTable)

	// We need to save status before spec update so that it's not lost.
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	meta.SetExternalName(cr, aws.StringValue(result.RouteTable.RouteTableId))

	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errSpecUpdate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeRouteTablesRequest(&awsec2.DescribeRouteTablesInput{
		RouteTableIds: []string{cr.Status.AtProvider.RouteTableID},
	}).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	if response.RouteTables == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNotFound)
	}

	table := response.RouteTables[0]

	patch, err := ec2.CreateRTPatch(&table, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	if patch.Routes != nil {
		// Attach the routes in Spec
		if err := e.createRoutes(ctx, cr.Status.AtProvider.RouteTableID, cr.Spec.ForProvider.Routes, cr.Status.AtProvider.Routes); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if patch.Associations != nil {
		// Associate route table to subnets in Spec.
		if err := e.createAssociations(ctx, cr.Status.AtProvider.RouteTableID, cr.Spec.ForProvider.Associations, cr.Status.AtProvider.Associations); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	if cr.Status.AtProvider.RouteTableID == "" {
		return errors.New(errDeleteNotPresent)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	// the subnet associations have to be deleted before deleting the route table.
	if err := e.deleteAssociations(ctx, cr.Status.AtProvider.Associations); err != nil {
		return err
	}

	_, err := e.client.DeleteRouteTableRequest(&awsec2.DeleteRouteTableInput{
		RouteTableId: aws.String(cr.Status.AtProvider.RouteTableID),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDelete)
}

func (e *external) createRoutes(ctx context.Context, tableID string, desired []v1beta1.Route, observed []v1beta1.RouteState) error {
	for _, rt := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.Route.GatewayID == rt.GatewayID && ob.Route.DestinationCIDRBlock == rt.DestinationCIDRBlock {
				isObserved = true
				break
			}
		}
		// if the route is already created, skip it
		if !isObserved {
			_, err := e.client.CreateRouteRequest(&awsec2.CreateRouteInput{
				RouteTableId:         aws.String(tableID),
				DestinationCidrBlock: aws.String(rt.DestinationCIDRBlock),
				GatewayId:            aws.String(rt.GatewayID),
			}).Send(ctx)

			if err != nil {
				return errors.Wrap(err, errCreateRoute)
			}
		}
	}

	return nil
}

func (e *external) createAssociations(ctx context.Context, tableID string, desired []v1beta1.Association, observed []v1beta1.AssociationState) error {
	for _, asc := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.Association.SubnetID == asc.SubnetID {
				isObserved = true
				break
			}
		}
		// if the association is already created, skip it
		if !isObserved {
			_, err := e.client.AssociateRouteTableRequest(&awsec2.AssociateRouteTableInput{
				RouteTableId: aws.String(tableID),
				SubnetId:     aws.String(asc.SubnetID),
			}).Send(ctx)

			if err != nil {
				return errors.Wrap(err, errAssociateSubnet)
			}
		}
	}

	return nil
}

func (e *external) deleteAssociations(ctx context.Context, observed []v1beta1.AssociationState) error {
	for _, asc := range observed {
		req := e.client.DisassociateRouteTableRequest(&awsec2.DisassociateRouteTableInput{
			AssociationId: aws.String(asc.AssociationID),
		})

		if _, err := req.Send(ctx); err != nil {
			if ec2.IsAssociationIDNotFoundErr(err) {
				continue
			}
			return errors.Wrap(err, errDisassociateSubnet)
		}
	}

	return nil
}
