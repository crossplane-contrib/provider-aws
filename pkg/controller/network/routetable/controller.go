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
	errUnexpectedObject    = "The managed resource is not an RouteTable resource"
	errClient              = "cannot create a new RouteTable client"
	errDescribe            = "failed to describe RouteTable"
	errMultipleItems       = "retrieved multiple RouteTables for the given routeTableId"
	errCreate              = "failed to create the RouteTable resource"
	errDelete              = "failed to delete the RouteTable resource"
	errCreateRoute         = "failed to create a route in the RouteTable resource"
	errPersistExternalName = "failed to persist RouteTable"
	errDeleteRoute         = "failed to delete a route in the RouteTable resource"
	errAssociateSubnet     = "failed to associate subnet with the RouteTable resource"
	errDisassociateSubnet  = "failed to disassociate subnet from the RouteTable resource"
)

// SetupRouteTable adds a controller that reconciles RouteTables.
func SetupRouteTable(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.RouteTableGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.RouteTable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.RouteTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewRouteTableClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (ec2.RouteTableClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.RouteTable)
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
	client ec2.RouteTableClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.RouteTable)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// AWS network resources are uniquely identified by an ID that is returned
	// on create time; we can't tell whether they exist unless we have recorded
	// their ID.
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	req := e.client.DescribeRouteTablesRequest(&awsec2.DescribeRouteTablesInput{
		RouteTableIds: []string{meta.GetExternalName(cr)},
	})

	response, err := req.Send(ctx)
	if ec2.IsRouteTableNotFoundErr(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
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

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.RouteTable)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.CreateRouteTableRequest(&awsec2.CreateRouteTableInput{
		VpcId: aws.String(cr.Spec.VPCID),
	})
	rsp, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(rsp.RouteTable.RouteTableId))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	cr.UpdateExternalStatus(*rsp.RouteTable)

	if err := e.createRoutes(ctx, meta.GetExternalName(cr), cr.Spec.Routes, cr.Status.Routes); err != nil {
		return managed.ExternalCreation{}, err
	}

	if err := e.createAssociations(ctx, meta.GetExternalName(cr), cr.Spec.Associations, cr.Status.Associations); err != nil {
		return managed.ExternalCreation{}, err
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.RouteTable)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	// in order to delete a route table, all of its dependencies need to be
	// deleted first

	if err := e.deleteRoutes(ctx, meta.GetExternalName(cr), cr.Status.Routes); err != nil {
		return err
	}

	if err := e.deleteAssociations(ctx, cr.Status.Associations); err != nil {
		return err
	}

	req := e.client.DeleteRouteTableRequest(&awsec2.DeleteRouteTableInput{
		RouteTableId: aws.String(meta.GetExternalName(cr)),
	})

	_, err := req.Send(ctx)
	return errors.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDelete)
}

func (e *external) createRoutes(ctx context.Context, tableID string, desired []v1alpha3.Route, observed []v1alpha3.RouteState) error {
	for _, rt := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.GatewayID == rt.GatewayID && ob.DestinationCIDRBlock == rt.DestinationCIDRBlock {
				isObserved = true
				break
			}
		}
		// if the route is already created (e.g. is observed), skip it
		if !isObserved {
			req := e.client.CreateRouteRequest(&awsec2.CreateRouteInput{
				RouteTableId:         aws.String(tableID),
				DestinationCidrBlock: aws.String(rt.DestinationCIDRBlock),
				GatewayId:            aws.String(rt.GatewayID),
			})

			if _, err := req.Send(ctx); err != nil {
				return errors.Wrap(err, errCreateRoute)
			}
		}
	}

	return nil
}

func (e *external) deleteRoutes(ctx context.Context, tableID string, observed []v1alpha3.RouteState) error {
	for _, rt := range observed {
		// "local" routes cannot be deleted
		// https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
		if rt.GatewayID == ec2.LocalGatewayID {
			continue
		}
		req := e.client.DeleteRouteRequest(&awsec2.DeleteRouteInput{
			RouteTableId:         aws.String(tableID),
			DestinationCidrBlock: aws.String(rt.DestinationCIDRBlock),
		})

		if _, err := req.Send(ctx); err != nil {
			if ec2.IsRouteNotFoundErr(err) {
				continue
			}
			return errors.Wrap(err, errDeleteRoute)
		}
	}

	return nil
}

func (e *external) createAssociations(ctx context.Context, tableID string, desired []v1alpha3.Association, observed []v1alpha3.AssociationState) error {
	for _, asc := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.SubnetID == asc.SubnetID {
				isObserved = true
				break
			}
		}
		// if the association is already created (e.g. is observed), skip it
		if !isObserved {
			req := e.client.AssociateRouteTableRequest(&awsec2.AssociateRouteTableInput{
				RouteTableId: aws.String(tableID),
				SubnetId:     aws.String(asc.SubnetID),
			})

			if _, err := req.Send(ctx); err != nil {
				return errors.Wrap(err, errAssociateSubnet)
			}
		}
	}

	return nil
}

func (e *external) deleteAssociations(ctx context.Context, observed []v1alpha3.AssociationState) error {
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
