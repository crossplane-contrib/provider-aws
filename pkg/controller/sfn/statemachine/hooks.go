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

package statemachine

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sfn"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sfn/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupStateMachine adds a controller that reconciles StateMachine.
func SetupStateMachine(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.StateMachineGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.StateMachine{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StateMachineGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.StateMachine, obj *svcsdk.DescribeStateMachineInput) error {
	obj.StateMachineArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.StateMachine, resp *svcsdk.DescribeStateMachineOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.Status) {
	case string(svcapitypes.StateMachineStatus_SDK_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.StateMachineStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.StateMachine, obj *svcsdk.CreateStateMachineInput) error {
	obj.Type = aws.String(string(cr.Spec.ForProvider.Type))
	obj.RoleArn = cr.Spec.ForProvider.RoleARN
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.StateMachine, resp *svcsdk.CreateStateMachineOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.StateMachineArn))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.StateMachine, obj *svcsdk.DeleteStateMachineInput) (bool, error) {
	obj.StateMachineArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
