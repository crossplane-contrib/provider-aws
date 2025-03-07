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

package resolverrule

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/route53resolver"
	svcsdk "github.com/aws/aws-sdk-go/service/route53resolver"
	svcsdkapi "github.com/aws/aws-sdk-go/service/route53resolver/route53resolveriface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

const (
	errUnexpectedObject = "managed resource is not an ResolverRule resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create ResolverRule in AWS"
	errUpdate        = "cannot update ResolverRule in AWS"
	errDescribe      = "failed to describe ResolverRule"
	errDelete        = "failed to delete ResolverRule"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, cr *svcapitypes.ResolverRule) (managed.TypedExternalClient[*svcapitypes.ResolverRule], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, cr *svcapitypes.ResolverRule) (managed.ExternalObservation, error) {
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateGetResolverRuleInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.GetResolverRuleWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateResolverRule(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)
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

func (e *external) Create(ctx context.Context, cr *svcapitypes.ResolverRule) (managed.ExternalCreation, error) {
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateResolverRuleInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateResolverRuleWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	if resp.ResolverRule.Arn != nil {
		cr.Status.AtProvider.ARN = resp.ResolverRule.Arn
	} else {
		cr.Status.AtProvider.ARN = nil
	}
	if resp.ResolverRule.CreationTime != nil {
		cr.Status.AtProvider.CreationTime = resp.ResolverRule.CreationTime
	} else {
		cr.Status.AtProvider.CreationTime = nil
	}
	if resp.ResolverRule.CreatorRequestId != nil {
		cr.Status.AtProvider.CreatorRequestID = resp.ResolverRule.CreatorRequestId
	} else {
		cr.Status.AtProvider.CreatorRequestID = nil
	}
	if resp.ResolverRule.DomainName != nil {
		cr.Spec.ForProvider.DomainName = resp.ResolverRule.DomainName
	} else {
		cr.Spec.ForProvider.DomainName = nil
	}
	if resp.ResolverRule.Id != nil {
		cr.Status.AtProvider.ID = resp.ResolverRule.Id
	} else {
		cr.Status.AtProvider.ID = nil
	}
	if resp.ResolverRule.ModificationTime != nil {
		cr.Status.AtProvider.ModificationTime = resp.ResolverRule.ModificationTime
	} else {
		cr.Status.AtProvider.ModificationTime = nil
	}
	if resp.ResolverRule.Name != nil {
		cr.Spec.ForProvider.Name = resp.ResolverRule.Name
	} else {
		cr.Spec.ForProvider.Name = nil
	}
	if resp.ResolverRule.OwnerId != nil {
		cr.Status.AtProvider.OwnerID = resp.ResolverRule.OwnerId
	} else {
		cr.Status.AtProvider.OwnerID = nil
	}
	if resp.ResolverRule.ResolverEndpointId != nil {
		cr.Spec.ForProvider.ResolverEndpointID = resp.ResolverRule.ResolverEndpointId
	} else {
		cr.Spec.ForProvider.ResolverEndpointID = nil
	}
	if resp.ResolverRule.RuleType != nil {
		cr.Spec.ForProvider.RuleType = resp.ResolverRule.RuleType
	} else {
		cr.Spec.ForProvider.RuleType = nil
	}
	if resp.ResolverRule.ShareStatus != nil {
		cr.Status.AtProvider.ShareStatus = resp.ResolverRule.ShareStatus
	} else {
		cr.Status.AtProvider.ShareStatus = nil
	}
	if resp.ResolverRule.Status != nil {
		cr.Status.AtProvider.Status = resp.ResolverRule.Status
	} else {
		cr.Status.AtProvider.Status = nil
	}
	if resp.ResolverRule.StatusMessage != nil {
		cr.Status.AtProvider.StatusMessage = resp.ResolverRule.StatusMessage
	} else {
		cr.Status.AtProvider.StatusMessage = nil
	}
	if resp.ResolverRule.TargetIps != nil {
		f13 := []*svcapitypes.TargetAddress{}
		for _, f13iter := range resp.ResolverRule.TargetIps {
			f13elem := &svcapitypes.TargetAddress{}
			if f13iter.Ip != nil {
				f13elem.IP = f13iter.Ip
			}
			if f13iter.Ipv6 != nil {
				f13elem.IPv6 = f13iter.Ipv6
			}
			if f13iter.Port != nil {
				f13elem.Port = f13iter.Port
			}
			f13 = append(f13, f13elem)
		}
		cr.Spec.ForProvider.TargetIPs = f13
	} else {
		cr.Spec.ForProvider.TargetIPs = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, cr *svcapitypes.ResolverRule) (managed.ExternalUpdate, error) {
	input := GenerateUpdateResolverRuleInput(cr)
	if err := e.preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.UpdateResolverRuleWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}

func (e *external) Delete(ctx context.Context, cr *svcapitypes.ResolverRule) (managed.ExternalDelete, error) {
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteResolverRuleInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return managed.ExternalDelete{}, nil
	}
	resp, err := e.client.DeleteResolverRuleWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.Route53ResolverAPI, opts []option) *external {
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
	client         svcsdkapi.Route53ResolverAPI
	preObserve     func(context.Context, *svcapitypes.ResolverRule, *svcsdk.GetResolverRuleInput) error
	postObserve    func(context.Context, *svcapitypes.ResolverRule, *svcsdk.GetResolverRuleOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	lateInitialize func(*svcapitypes.ResolverRuleParameters, *svcsdk.GetResolverRuleOutput) error
	isUpToDate     func(context.Context, *svcapitypes.ResolverRule, *svcsdk.GetResolverRuleOutput) (bool, string, error)
	preCreate      func(context.Context, *svcapitypes.ResolverRule, *svcsdk.CreateResolverRuleInput) error
	postCreate     func(context.Context, *svcapitypes.ResolverRule, *svcsdk.CreateResolverRuleOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.ResolverRule, *svcsdk.DeleteResolverRuleInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.ResolverRule, *svcsdk.DeleteResolverRuleOutput, error) (managed.ExternalDelete, error)
	preUpdate      func(context.Context, *svcapitypes.ResolverRule, *svcsdk.UpdateResolverRuleInput) error
	postUpdate     func(context.Context, *svcapitypes.ResolverRule, *svcsdk.UpdateResolverRuleOutput, managed.ExternalUpdate, error) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.ResolverRule, *svcsdk.GetResolverRuleInput) error {
	return nil
}

func nopPostObserve(_ context.Context, _ *svcapitypes.ResolverRule, _ *svcsdk.GetResolverRuleOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopLateInitialize(*svcapitypes.ResolverRuleParameters, *svcsdk.GetResolverRuleOutput) error {
	return nil
}
func alwaysUpToDate(context.Context, *svcapitypes.ResolverRule, *svcsdk.GetResolverRuleOutput) (bool, string, error) {
	return true, "", nil
}

func nopPreCreate(context.Context, *svcapitypes.ResolverRule, *svcsdk.CreateResolverRuleInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.ResolverRule, _ *svcsdk.CreateResolverRuleOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.ResolverRule, *svcsdk.DeleteResolverRuleInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.ResolverRule, _ *svcsdk.DeleteResolverRuleOutput, err error) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, err
}
func nopPreUpdate(context.Context, *svcapitypes.ResolverRule, *svcsdk.UpdateResolverRuleInput) error {
	return nil
}
func nopPostUpdate(_ context.Context, _ *svcapitypes.ResolverRule, _ *svcsdk.UpdateResolverRuleOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
