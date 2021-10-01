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

// Code generated by ack-generate. DO NOT EDIT.

package route

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/ec2"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errUnexpectedObject = "managed resource is not an Route resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create Route in AWS"
	errUpdate        = "cannot update Route in AWS"
	errDescribe      = "failed to describe Route"
	errDelete        = "failed to delete Route"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Route)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := awsclient.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.Route)

	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	input := svcsdk.DescribeRouteTablesInput{
		RouteTableIds: []*string{cr.Spec.ForProvider.RouteTableID},
	}

	resp, errr := e.client.DescribeRouteTables(&input)
	if errr != nil {
		return managed.ExternalObservation{}, errors.New(errDescribe)
	}

	if len(resp.RouteTables) != 1 {
		return managed.ExternalObservation{}, errors.New(errDescribe)
	}

	observed := resp.RouteTables[0]

	obsRoutes := observed.Routes
	current := cr.Spec.ForProvider.DeepCopy()
	for _, route := range obsRoutes {
		if derefString(route.DestinationCidrBlock) == derefString(cr.Spec.ForProvider.DestinationCIDRBlock) &&
			derefString(route.VpcPeeringConnectionId) == derefString(cr.Spec.ForProvider.VPCPeeringConnectionID) &&
			derefString(route.GatewayId) == derefString(cr.Spec.ForProvider.GatewayID) &&
			derefString(route.LocalGatewayId) == derefString(cr.Spec.ForProvider.LocalGatewayID) &&
			derefString(route.NatGatewayId) == derefString(cr.Spec.ForProvider.NatGatewayID) &&
			derefString(route.NetworkInterfaceId) == derefString(cr.Spec.ForProvider.NetworkInterfaceID) &&
			derefString(route.TransitGatewayId) == derefString(cr.Spec.ForProvider.TransitGatewayID) &&
			derefString(route.InstanceId) == derefString(cr.Spec.ForProvider.InstanceID) {
			cr.SetConditions(xpv1.Available())
			return managed.ExternalObservation{
				ResourceExists:          true,
				ResourceUpToDate:        true,
				ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
			}, nil
		}
	}
	return e.observe(ctx, mg)
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

func (e *external) Create(ctx context.Context, mg cpresource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.Route)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateRouteInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateRouteWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	if resp.Return != nil {
		cr.Status.AtProvider.Return = resp.Return
	} else {
		cr.Status.AtProvider.Return = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	return e.update(ctx, mg)

}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.Route)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteRouteInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return nil
	}
	resp, err := e.client.DeleteRouteWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.EC2API, opts []option) *external {
	e := &external{
		kube:       kube,
		client:     client,
		observe:    nopObserve,
		preCreate:  nopPreCreate,
		postCreate: nopPostCreate,
		preDelete:  nopPreDelete,
		postDelete: nopPostDelete,
		update:     nopUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube       client.Client
	client     svcsdkapi.EC2API
	observe    func(context.Context, cpresource.Managed) (managed.ExternalObservation, error)
	preCreate  func(context.Context, *svcapitypes.Route, *svcsdk.CreateRouteInput) error
	postCreate func(context.Context, *svcapitypes.Route, *svcsdk.CreateRouteOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete  func(context.Context, *svcapitypes.Route, *svcsdk.DeleteRouteInput) (bool, error)
	postDelete func(context.Context, *svcapitypes.Route, *svcsdk.DeleteRouteOutput, error) error
	update     func(context.Context, cpresource.Managed) (managed.ExternalUpdate, error)
}

func nopObserve(context.Context, cpresource.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}

func nopPreCreate(context.Context, *svcapitypes.Route, *svcsdk.CreateRouteInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.Route, _ *svcsdk.CreateRouteOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.Route, *svcsdk.DeleteRouteInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.Route, _ *svcsdk.DeleteRouteOutput, err error) error {
	return err
}
func nopUpdate(context.Context, cpresource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
