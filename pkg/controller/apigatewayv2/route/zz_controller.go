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

	svcapi "github.com/aws/aws-sdk-go/service/apigatewayv2"
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
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

func (c *connector) Connect(ctx context.Context, cr *svcapitypes.Route) (managed.TypedExternalClient[*svcapitypes.Route], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, cr *svcapitypes.Route) (managed.ExternalObservation, error) {
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateGetRouteInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.GetRouteWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateRoute(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)
	upToDate := true
	diff := ""
	if !meta.WasDeleted(cr) { // There is no need to run isUpToDate if the resource is deleted
		upToDate, diff, err = e.isUpToDate(ctx, cr, resp)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
		}
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		Diff:                    diff,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (e *external) Create(ctx context.Context, cr *svcapitypes.Route) (managed.ExternalCreation, error) {
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateRouteInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateRouteWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	if resp.ApiGatewayManaged != nil {
		cr.Status.AtProvider.APIGatewayManaged = resp.ApiGatewayManaged
	} else {
		cr.Status.AtProvider.APIGatewayManaged = nil
	}
	if resp.ApiKeyRequired != nil {
		cr.Spec.ForProvider.APIKeyRequired = resp.ApiKeyRequired
	} else {
		cr.Spec.ForProvider.APIKeyRequired = nil
	}
	if resp.AuthorizationScopes != nil {
		f2 := []*string{}
		for _, f2iter := range resp.AuthorizationScopes {
			var f2elem string
			f2elem = *f2iter
			f2 = append(f2, &f2elem)
		}
		cr.Spec.ForProvider.AuthorizationScopes = f2
	} else {
		cr.Spec.ForProvider.AuthorizationScopes = nil
	}
	if resp.AuthorizationType != nil {
		cr.Spec.ForProvider.AuthorizationType = resp.AuthorizationType
	} else {
		cr.Spec.ForProvider.AuthorizationType = nil
	}
	if resp.AuthorizerId != nil {
		cr.Spec.ForProvider.AuthorizerID = resp.AuthorizerId
	} else {
		cr.Spec.ForProvider.AuthorizerID = nil
	}
	if resp.ModelSelectionExpression != nil {
		cr.Spec.ForProvider.ModelSelectionExpression = resp.ModelSelectionExpression
	} else {
		cr.Spec.ForProvider.ModelSelectionExpression = nil
	}
	if resp.OperationName != nil {
		cr.Spec.ForProvider.OperationName = resp.OperationName
	} else {
		cr.Spec.ForProvider.OperationName = nil
	}
	if resp.RequestModels != nil {
		f7 := map[string]*string{}
		for f7key, f7valiter := range resp.RequestModels {
			var f7val string
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		cr.Spec.ForProvider.RequestModels = f7
	} else {
		cr.Spec.ForProvider.RequestModels = nil
	}
	if resp.RequestParameters != nil {
		f8 := map[string]*svcapitypes.ParameterConstraints{}
		for f8key, f8valiter := range resp.RequestParameters {
			f8val := &svcapitypes.ParameterConstraints{}
			if f8valiter.Required != nil {
				f8val.Required = f8valiter.Required
			}
			f8[f8key] = f8val
		}
		cr.Spec.ForProvider.RequestParameters = f8
	} else {
		cr.Spec.ForProvider.RequestParameters = nil
	}
	if resp.RouteId != nil {
		cr.Status.AtProvider.RouteID = resp.RouteId
	} else {
		cr.Status.AtProvider.RouteID = nil
	}
	if resp.RouteKey != nil {
		cr.Spec.ForProvider.RouteKey = resp.RouteKey
	} else {
		cr.Spec.ForProvider.RouteKey = nil
	}
	if resp.RouteResponseSelectionExpression != nil {
		cr.Spec.ForProvider.RouteResponseSelectionExpression = resp.RouteResponseSelectionExpression
	} else {
		cr.Spec.ForProvider.RouteResponseSelectionExpression = nil
	}
	if resp.Target != nil {
		cr.Status.AtProvider.Target = resp.Target
	} else {
		cr.Status.AtProvider.Target = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, cr *svcapitypes.Route) (managed.ExternalUpdate, error) {
	input := GenerateUpdateRouteInput(cr)
	if err := e.preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.UpdateRouteWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}

func (e *external) Delete(ctx context.Context, cr *svcapitypes.Route) (managed.ExternalDelete, error) {
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteRouteInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return managed.ExternalDelete{}, nil
	}
	resp, err := e.client.DeleteRouteWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.ApiGatewayV2API, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		preCreate:      nopPreCreate,
		postCreate:     nopPostCreate,
		preDelete:      nopPreDelete,
		postDelete:     nopPostDelete,
		preUpdate:      nopPreUpdate,
		postUpdate:     nopPostUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube           client.Client
	client         svcsdkapi.ApiGatewayV2API
	preObserve     func(context.Context, *svcapitypes.Route, *svcsdk.GetRouteInput) error
	postObserve    func(context.Context, *svcapitypes.Route, *svcsdk.GetRouteOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	lateInitialize func(*svcapitypes.RouteParameters, *svcsdk.GetRouteOutput) error
	isUpToDate     func(context.Context, *svcapitypes.Route, *svcsdk.GetRouteOutput) (bool, string, error)
	preCreate      func(context.Context, *svcapitypes.Route, *svcsdk.CreateRouteInput) error
	postCreate     func(context.Context, *svcapitypes.Route, *svcsdk.CreateRouteOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.Route, *svcsdk.DeleteRouteInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.Route, *svcsdk.DeleteRouteOutput, error) (managed.ExternalDelete, error)
	preUpdate      func(context.Context, *svcapitypes.Route, *svcsdk.UpdateRouteInput) error
	postUpdate     func(context.Context, *svcapitypes.Route, *svcsdk.UpdateRouteOutput, managed.ExternalUpdate, error) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.Route, *svcsdk.GetRouteInput) error {
	return nil
}

func nopPostObserve(_ context.Context, _ *svcapitypes.Route, _ *svcsdk.GetRouteOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopLateInitialize(*svcapitypes.RouteParameters, *svcsdk.GetRouteOutput) error {
	return nil
}
func alwaysUpToDate(context.Context, *svcapitypes.Route, *svcsdk.GetRouteOutput) (bool, string, error) {
	return true, "", nil
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
func nopPostDelete(_ context.Context, _ *svcapitypes.Route, _ *svcsdk.DeleteRouteOutput, err error) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, err
}
func nopPreUpdate(context.Context, *svcapitypes.Route, *svcsdk.UpdateRouteInput) error {
	return nil
}
func nopPostUpdate(_ context.Context, _ *svcapitypes.Route, _ *svcsdk.UpdateRouteOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
