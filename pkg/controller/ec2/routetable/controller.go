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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an RouteTable resource"
	errInvalidRoutes    = "RouteTable routes are invalid"

	errDescribe           = "failed to describe RouteTable"
	errMultipleItems      = "retrieved multiple RouteTables for the given routeTableId"
	errCreate             = "failed to create the RouteTable resource"
	errUpdate             = "failed to update the RouteTable"
	errUpdateNotFound     = "cannot update the RouteTable, since the RouteTableID is not present"
	errDelete             = "failed to delete the RouteTable resource"
	errCreateRoute        = "failed to create a route in the RouteTable resource"
	errDeleteRoute        = "failed to delete a route in the RouteTable resource"
	errAssociateSubnet    = "failed to associate subnet to the RouteTable resource"
	errDisassociateSubnet = "failed to disassociate subnet from the RouteTable resource"
	errCreateTags         = "failed to create tags for the RouteTable resource"
	errDeleteTags         = "failed to delete tags for the RouteTable resource"
)

// SetupRouteTable adds a controller that reconciles RouteTables.
func SetupRouteTable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RouteTableGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewRouteTableClient}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RouteTableGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.RouteTable{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.RouteTableClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RouteTable)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg)}, nil
}

type external struct {
	client ec2.RouteTableClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if err := ec2.ValidateRoutes(cr.Spec.ForProvider.Routes); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errInvalidRoutes)
	}
	// To find out whether a RouteTable exist:
	// - the object's ExternalName should have routeTableId populated
	// - a RouteTable with the given routeTableId should exist
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeRouteTables(ctx, &awsec2.DescribeRouteTablesInput{
		RouteTableIds: []string{meta.GetExternalName(cr)},
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.RouteTables) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.RouteTables[0]
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeRT(&cr.Spec.ForProvider, &response.RouteTables[0])

	stateAvailable := true
	for _, rt := range observed.Routes {
		if rt.State != awsec2types.RouteStateActive {
			stateAvailable = false
			break
		}
	}
	if stateAvailable {
		cr.SetConditions(xpv1.Available())
	}

	cr.Status.AtProvider = ec2.GenerateRTObservation(observed)

	upToDate, err := ec2.IsRtUpToDate(cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errDescribe)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	if err := ec2.ValidateRoutes(cr.Spec.ForProvider.Routes); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errInvalidRoutes)
	}
	result, err := e.client.CreateRouteTable(ctx, &awsec2.CreateRouteTableInput{
		VpcId: cr.Spec.ForProvider.VPCID,
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.ToString(result.RouteTable.RouteTableId))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.RouteTable)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	if err := ec2.ValidateRoutes(cr.Spec.ForProvider.Routes); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errInvalidRoutes)
	}
	response, err := e.client.DescribeRouteTables(ctx, &awsec2.DescribeRouteTablesInput{
		RouteTableIds: []string{meta.GetExternalName(cr)},
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDescribe)
	}

	if response.RouteTables == nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdateNotFound)
	}

	table := response.RouteTables[0]

	patch, err := ec2.CreateRTPatch(table, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	if len(patch.Tags) != 0 {
		// tagging the RouteTable
		addTags, removeTags := ec2.DiffEC2Tags(ec2.GenerateEC2TagsV1Beta1(cr.Spec.ForProvider.Tags), table.Tags)
		if len(addTags) > 0 {
			if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
				Resources: []string{meta.GetExternalName(cr)},
				Tags:      addTags,
			}); err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreateTags)
			}
		}
		if len(removeTags) > 0 {
			if _, err := e.client.DeleteTags(ctx, &awsec2.DeleteTagsInput{
				Resources: []string{meta.GetExternalName(cr)},
				Tags:      removeTags,
			}); err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errDeleteTags)
			}
		}
	}

	if patch.Routes != nil {
		// Attach the routes in Spec
		if err := e.reconcileRoutes(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider.Routes, cr.Status.AtProvider.Routes); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if patch.Associations != nil {
		// Associate route table to subnets in Spec.
		if err := e.reconcileAssociations(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider.Associations, cr.Status.AtProvider.Associations); err != nil {
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

	if err := ec2.ValidateRoutes(cr.Spec.ForProvider.Routes); err != nil {
		return errors.Wrap(err, errInvalidRoutes)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// the subnet associations have to be deleted before deleting the route table.
	if err := e.deleteAssociations(ctx, cr.Status.AtProvider.Associations); err != nil {
		return err
	}

	_, err := e.client.DeleteRouteTable(ctx, &awsec2.DeleteRouteTableInput{
		RouteTableId: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(ec2.IsRouteTableNotFoundErr, err), errDelete)
}

func (e *external) deleteRoutes(ctx context.Context, tableID string, desired []v1beta1.RouteBeta, observed []v1beta1.RouteState) error { //nolint:gocyclo
	for _, rt := range observed {
		found := false
		for _, ds := range desired {
			if aws.ToString(ds.DestinationCIDRBlock) == rt.DestinationCIDRBlock && (aws.ToString(ds.GatewayID) == rt.GatewayID &&
				aws.ToString(ds.InstanceID) == rt.InstanceID &&
				aws.ToString(ds.LocalGatewayID) == rt.LocalGatewayID &&
				aws.ToString(ds.NatGatewayID) == rt.NatGatewayID &&
				aws.ToString(ds.NetworkInterfaceID) == rt.NetworkInterfaceID &&
				aws.ToString(ds.TransitGatewayID) == rt.TransitGatewayID &&
				aws.ToString(ds.VpcPeeringConnectionID) == rt.VpcPeeringConnectionID) {

				found = true
				break
			}
		}
		if !found && rt.GatewayID != ec2.DefaultLocalGatewayID {
			if rt.DestinationCIDRBlock != "" {
				_, err := e.client.DeleteRoute(ctx, &awsec2.DeleteRouteInput{
					RouteTableId:         aws.String(tableID),
					DestinationCidrBlock: aws.String(rt.DestinationCIDRBlock),
				})

				if err != nil {
					return err
				}
			} else {
				_, err := e.client.DeleteRoute(ctx, &awsec2.DeleteRouteInput{
					RouteTableId:             aws.String(tableID),
					DestinationIpv6CidrBlock: aws.String(rt.DestinationIPV6CIDRBlock),
				})

				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *external) createRoutes(ctx context.Context, tableID string, desired []v1beta1.RouteBeta, observed []v1beta1.RouteState) error { //nolint:gocyclo
	for _, rt := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.DestinationCIDRBlock == aws.ToString(rt.DestinationCIDRBlock) && (ob.GatewayID == aws.ToString(rt.GatewayID) &&
				ob.InstanceID == aws.ToString(rt.InstanceID) &&
				ob.LocalGatewayID == aws.ToString(rt.LocalGatewayID) &&
				ob.NatGatewayID == aws.ToString(rt.NatGatewayID) &&
				ob.NetworkInterfaceID == aws.ToString(rt.NetworkInterfaceID) &&
				ob.TransitGatewayID == aws.ToString(rt.TransitGatewayID) &&
				ob.VpcPeeringConnectionID == aws.ToString(rt.VpcPeeringConnectionID)) {
				isObserved = true
				break
			}
		}
		// if the route is already created, skip it
		if !isObserved {
			_, err := e.client.CreateRoute(ctx, &awsec2.CreateRouteInput{
				RouteTableId:             aws.String(tableID),
				DestinationCidrBlock:     rt.DestinationCIDRBlock,
				GatewayId:                rt.GatewayID,
				DestinationIpv6CidrBlock: rt.DestinationIPV6CIDRBlock,
				InstanceId:               rt.InstanceID,
				LocalGatewayId:           rt.LocalGatewayID,
				NatGatewayId:             rt.NatGatewayID,
				NetworkInterfaceId:       rt.NetworkInterfaceID,
				TransitGatewayId:         rt.TransitGatewayID,
				VpcPeeringConnectionId:   rt.VpcPeeringConnectionID,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *external) reconcileRoutes(ctx context.Context, tableID string, desired []v1beta1.RouteBeta, observed []v1beta1.RouteState) error {

	if err := e.deleteRoutes(ctx, tableID, desired, observed); err != nil {
		return errorutils.Wrap(err, errDeleteRoute)
	}
	if err := e.createRoutes(ctx, tableID, desired, observed); err != nil {
		return errorutils.Wrap(err, errCreateRoute)
	}

	return nil
}

func (e *external) removeAssociations(ctx context.Context, desired []v1beta1.Association, observed []v1beta1.AssociationState) error {
	var toDelete []v1beta1.AssociationState
	for _, asc := range observed {
		found := false
		for _, ds := range desired {
			if asc.SubnetID == aws.ToString(ds.SubnetID) {
				found = true
				break
			}
		}
		if !found {
			// No longer needed add to delete list
			toDelete = append(toDelete, asc)
		}
	}
	return e.deleteAssociations(ctx, toDelete)
}

func (e *external) createAssociations(ctx context.Context, tableID string, desired []v1beta1.Association, observed []v1beta1.AssociationState) error {
	for _, asc := range desired {
		isObserved := false
		for _, ob := range observed {
			if ob.SubnetID == aws.ToString(asc.SubnetID) {
				isObserved = true
				break
			}
		}
		// if the association is already created, skip it
		if !isObserved {
			_, err := e.client.AssociateRouteTable(ctx, &awsec2.AssociateRouteTableInput{
				RouteTableId: aws.String(tableID),
				SubnetId:     asc.SubnetID,
			})

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *external) reconcileAssociations(ctx context.Context, tableID string, desired []v1beta1.Association, observed []v1beta1.AssociationState) error {
	if err := e.removeAssociations(ctx, desired, observed); err != nil {
		// underlying deleteAssociations already wraps the error
		return err
	}

	if err := e.createAssociations(ctx, tableID, desired, observed); err != nil {
		return errorutils.Wrap(err, errAssociateSubnet)
	}
	return nil
}

func (e *external) deleteAssociations(ctx context.Context, observed []v1beta1.AssociationState) error {
	for _, asc := range observed {
		_, err := e.client.DisassociateRouteTable(ctx, &awsec2.DisassociateRouteTableInput{
			AssociationId: aws.String(asc.AssociationID),
		})

		if err != nil {
			if ec2.IsAssociationIDNotFoundErr(err) {
				continue
			}
			return errorutils.Wrap(err, errDisassociateSubnet)
		}
	}

	return nil
}
