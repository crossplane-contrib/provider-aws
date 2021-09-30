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

package dbclusterparametergroup

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/docdb"
	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/docdb/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errUnexpectedObject = "managed resource is not an DBClusterParameterGroup resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create DBClusterParameterGroup in AWS"
	errUpdate        = "cannot update DBClusterParameterGroup in AWS"
	errDescribe      = "failed to describe DBClusterParameterGroup"
	errDelete        = "failed to delete DBClusterParameterGroup"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.DBClusterParameterGroup)
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
	cr, ok := mg.(*svcapitypes.DBClusterParameterGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateDescribeDBClusterParameterGroupsInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.DescribeDBClusterParameterGroupsWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	resp = e.filterList(cr, resp)
	if len(resp.DBClusterParameterGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateDBClusterParameterGroup(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)

	upToDate, err := e.isUpToDate(cr, resp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (e *external) Create(ctx context.Context, mg cpresource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.DBClusterParameterGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateDBClusterParameterGroupInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateDBClusterParameterGroupWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	if resp.DBClusterParameterGroup.DBClusterParameterGroupArn != nil {
		cr.Status.AtProvider.DBClusterParameterGroupARN = resp.DBClusterParameterGroup.DBClusterParameterGroupArn
	} else {
		cr.Status.AtProvider.DBClusterParameterGroupARN = nil
	}
	if resp.DBClusterParameterGroup.DBClusterParameterGroupName != nil {
		cr.Status.AtProvider.DBClusterParameterGroupName = resp.DBClusterParameterGroup.DBClusterParameterGroupName
	} else {
		cr.Status.AtProvider.DBClusterParameterGroupName = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.DBClusterParameterGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GenerateModifyDBClusterParameterGroupInput(cr)
	if err := e.preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.ModifyDBClusterParameterGroupWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate))
}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.DBClusterParameterGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteDBClusterParameterGroupInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return nil
	}
	resp, err := e.client.DeleteDBClusterParameterGroupWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.DocDBAPI, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		filterList:     nopFilterList,
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
	client         svcsdkapi.DocDBAPI
	preObserve     func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsInput) error
	postObserve    func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	filterList     func(*svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsOutput) *svcsdk.DescribeDBClusterParameterGroupsOutput
	lateInitialize func(*svcapitypes.DBClusterParameterGroupParameters, *svcsdk.DescribeDBClusterParameterGroupsOutput) error
	isUpToDate     func(*svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsOutput) (bool, error)
	preCreate      func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.CreateDBClusterParameterGroupInput) error
	postCreate     func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.CreateDBClusterParameterGroupOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DeleteDBClusterParameterGroupInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DeleteDBClusterParameterGroupOutput, error) error
	preUpdate      func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.ModifyDBClusterParameterGroupInput) error
	postUpdate     func(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.ModifyDBClusterParameterGroupOutput, managed.ExternalUpdate, error) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsInput) error {
	return nil
}
func nopPostObserve(_ context.Context, _ *svcapitypes.DBClusterParameterGroup, _ *svcsdk.DescribeDBClusterParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopFilterList(_ *svcapitypes.DBClusterParameterGroup, list *svcsdk.DescribeDBClusterParameterGroupsOutput) *svcsdk.DescribeDBClusterParameterGroupsOutput {
	return list
}

func nopLateInitialize(*svcapitypes.DBClusterParameterGroupParameters, *svcsdk.DescribeDBClusterParameterGroupsOutput) error {
	return nil
}
func alwaysUpToDate(*svcapitypes.DBClusterParameterGroup, *svcsdk.DescribeDBClusterParameterGroupsOutput) (bool, error) {
	return true, nil
}

func nopPreCreate(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.CreateDBClusterParameterGroupInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.DBClusterParameterGroup, _ *svcsdk.CreateDBClusterParameterGroupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.DeleteDBClusterParameterGroupInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.DBClusterParameterGroup, _ *svcsdk.DeleteDBClusterParameterGroupOutput, err error) error {
	return err
}
func nopPreUpdate(context.Context, *svcapitypes.DBClusterParameterGroup, *svcsdk.ModifyDBClusterParameterGroupInput) error {
	return nil
}
func nopPostUpdate(_ context.Context, _ *svcapitypes.DBClusterParameterGroup, _ *svcsdk.ModifyDBClusterParameterGroupOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
