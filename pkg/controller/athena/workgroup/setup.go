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

package workgroup

import (
	"context"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/athena"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/athena/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupWorkGroup adds a controller that reconciles WorkGroup.
func SetupWorkGroup(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.WorkGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.lateInitialize = LateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(limiter),
		}).
		For(&svcapitypes.WorkGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.WorkGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preDelete(_ context.Context, cr *svcapitypes.WorkGroup, obj *svcsdk.DeleteWorkGroupInput) (bool, error) {
	obj.WorkGroup = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.WorkGroup, obj *svcsdk.GetWorkGroupInput) error {
	obj.WorkGroup = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.WorkGroup, obj *svcsdk.GetWorkGroupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.WorkGroup.State) {
	case string(svcapitypes.WorkGroupState_ENABLED):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.WorkGroupState_DISABLED):
		cr.SetConditions(xpv1.Unavailable())
	}

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.WorkGroup, obj *svcsdk.CreateWorkGroupInput) error {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	return nil
}

// LateInitialize fills the empty fields in *svcapitypes.WorkGroupParameters with
// the values seen in svcsdk.GetWorkGroupOutput.
// nolint:gocyclo
func LateInitialize(cr *svcapitypes.WorkGroupParameters, obj *svcsdk.GetWorkGroupOutput) error {

	if cr.Configuration == nil && obj.WorkGroup.Configuration != nil {
		cr.Configuration = &svcapitypes.WorkGroupConfiguration{
			EnforceWorkGroupConfiguration:   obj.WorkGroup.Configuration.EnforceWorkGroupConfiguration,
			PublishCloudWatchMetricsEnabled: obj.WorkGroup.Configuration.PublishCloudWatchMetricsEnabled,
			RequesterPaysEnabled:            obj.WorkGroup.Configuration.RequesterPaysEnabled,
			EngineVersion: &svcapitypes.EngineVersion{
				SelectedEngineVersion:  obj.WorkGroup.Configuration.EngineVersion.SelectedEngineVersion,
				EffectiveEngineVersion: obj.WorkGroup.Configuration.EngineVersion.EffectiveEngineVersion,
			},
		}
	}

	return nil
}
