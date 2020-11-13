/*
Copyright 2020 The Crossplane Authors.

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

package vpclink

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupVPCLink adds a controller that reconciles VPCLink.
func SetupVPCLink(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.VPCLinkGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.VPCLink{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.VPCLinkGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.VPCLink) error {
	return nil
}
func (e *external) postObserve(_ context.Context, cr *svcapitypes.VPCLink, resp *svcsdk.GetVpcLinksOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	vl := e.filterList(cr, resp).Items
	if len(vl) != 1 {
		return managed.ExternalObservation{}, errors.New("there needs to be one element in the filtered response")
	}
	if aws.StringValue(vl[0].VpcLinkStatus) == "AVAILABLE" {
		cr.SetConditions(v1alpha1.Available())
	}
	return obs, nil
}

func (*external) filterList(cr *svcapitypes.VPCLink, list *svcsdk.GetVpcLinksOutput) *svcsdk.GetVpcLinksOutput {
	res := &svcsdk.GetVpcLinksOutput{}
	for _, vl := range list.Items {
		if meta.GetExternalName(cr) == aws.StringValue(vl.Name) {
			res.Items = append(res.Items, vl)
			break
		}
	}
	return res
}

func (*external) preCreate(context.Context, *svcapitypes.VPCLink) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.VPCLink, _ *svcsdk.CreateVpcLinkOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.VPCLink) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.VPCLink, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.VPCLinkParameters, *svcsdk.GetVpcLinksOutput) error {
	return nil
}

func preGenerateGetVpcLinksInput(_ *svcapitypes.VPCLink, obj *svcsdk.GetVpcLinksInput) *svcsdk.GetVpcLinksInput {
	return obj
}

func postGenerateGetVpcLinksInput(_ *svcapitypes.VPCLink, obj *svcsdk.GetVpcLinksInput) *svcsdk.GetVpcLinksInput {
	return obj
}

func preGenerateCreateVpcLinkInput(_ *svcapitypes.VPCLink, obj *svcsdk.CreateVpcLinkInput) *svcsdk.CreateVpcLinkInput {
	return obj
}

func postGenerateCreateVpcLinkInput(cr *svcapitypes.VPCLink, obj *svcsdk.CreateVpcLinkInput) *svcsdk.CreateVpcLinkInput {
	obj.Name = aws.String(meta.GetExternalName(cr))
	for _, sg := range cr.Spec.ForProvider.SecurityGroupIDs {
		obj.SecurityGroupIds = append(obj.SecurityGroupIds, aws.String(sg))
	}
	for _, s := range cr.Spec.ForProvider.SubnetIDs {
		obj.SubnetIds = append(obj.SubnetIds, aws.String(s))
	}
	return obj
}

func preGenerateDeleteVpcLinkInput(_ *svcapitypes.VPCLink, obj *svcsdk.DeleteVpcLinkInput) *svcsdk.DeleteVpcLinkInput {
	return obj
}

func postGenerateDeleteVpcLinkInput(_ *svcapitypes.VPCLink, obj *svcsdk.DeleteVpcLinkInput) *svcsdk.DeleteVpcLinkInput {
	return obj
}
