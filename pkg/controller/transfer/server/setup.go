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

package server

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupServer adds a controller that reconciles Server.
func SetupServer(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServerGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ServerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Server{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerInput) error {
	if meta.GetExternalName(cr) != "" {
		obj.ServerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DeleteServerInput) (bool, error) {
	obj.ServerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(obj.Server.State) {
	case string(svcapitypes.State_OFFLINE):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.State_ONLINE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.State_STARTING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.State_STOPPING):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.State_START_FAILED):
		cr.SetConditions(xpv1.ReconcileError(err))
	case string(svcapitypes.State_STOP_FAILED):
		cr.SetConditions(xpv1.ReconcileError(err))
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		"HostKeyFingerprint": []byte(pointer.StringValue(obj.Server.HostKeyFingerprint)),
	}

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.ServerId))
	return managed.ExternalCreation{}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerInput) error {
	obj.Certificate = cr.Spec.ForProvider.Certificate
	obj.LoggingRole = cr.Spec.ForProvider.LoggingRole
	obj.EndpointDetails = &svcsdk.EndpointDetails{}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs) > 0 {
		obj.EndpointDetails.SecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs))
		copy(obj.EndpointDetails.SecurityGroupIds, cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs)
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs) > 0 {
		obj.EndpointDetails.SubnetIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs))
		copy(obj.EndpointDetails.SubnetIds, cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs)
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs) > 0 {
		obj.EndpointDetails.AddressAllocationIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs))
		copy(obj.EndpointDetails.AddressAllocationIds, cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs)
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID != nil {
		obj.EndpointDetails.VpcEndpointId = cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCID != nil {
		obj.EndpointDetails.VpcId = cr.Spec.ForProvider.CustomEndpointDetails.VPCID
	}

	return nil
}
