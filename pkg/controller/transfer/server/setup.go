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
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/transfer/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupServer adds a controller that reconciles Server.
func SetupServer(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.Server{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ServerGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerInput) error {
	if meta.GetExternalName(cr) != "" {
		obj.ServerId = awsclients.String(meta.GetExternalName(cr))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DeleteServerInput) (bool, error) {
	obj.ServerId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.DescribeServerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.Server.State) {
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
		"HostKeyFingerprint": []byte(awsclients.StringValue(obj.Server.HostKeyFingerprint)),
	}

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.ServerId))
	return managed.ExternalCreation{}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Server, obj *svcsdk.CreateServerInput) error {
	obj.Certificate = cr.Spec.ForProvider.Certificate
	obj.LoggingRole = cr.Spec.ForProvider.LoggingRole
	obj.EndpointDetails = &svcsdk.EndpointDetails{}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs) > 0 {
		obj.EndpointDetails.SecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs))
		for i, v := range cr.Spec.ForProvider.CustomEndpointDetails.SecurityGroupIDs {
			obj.EndpointDetails.SecurityGroupIds[i] = v
		}
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs) > 0 {
		obj.EndpointDetails.SubnetIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs))
		for i, v := range cr.Spec.ForProvider.CustomEndpointDetails.SubnetIDs {
			obj.EndpointDetails.SubnetIds[i] = v
		}
	}

	if len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs) > 0 {
		obj.EndpointDetails.AddressAllocationIds = make([]*string, len(cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs))
		for i, v := range cr.Spec.ForProvider.CustomEndpointDetails.AddressAllocationIDs {
			obj.EndpointDetails.AddressAllocationIds[i] = v
		}
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID != nil {
		obj.EndpointDetails.VpcEndpointId = cr.Spec.ForProvider.CustomEndpointDetails.VPCEndpointID
	}

	if cr.Spec.ForProvider.CustomEndpointDetails.VPCID != nil {
		obj.EndpointDetails.VpcId = cr.Spec.ForProvider.CustomEndpointDetails.VPCID
	}

	return nil
}
