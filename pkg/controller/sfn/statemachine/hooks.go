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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sfn/v1alpha1"
)

// SetupStateMachine adds a controller that reconciles StateMachine.
func SetupStateMachine(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.StateMachineGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.StateMachine{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StateMachineGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.StateMachine) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.StateMachine, _ *svcsdk.DescribeStateMachineOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.StateMachine) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.StateMachine, _ *svcsdk.CreateStateMachineOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.StateMachine) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.StateMachine, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.StateMachineParameters, *svcsdk.DescribeStateMachineOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.StateMachine, *svcsdk.DescribeStateMachineOutput) bool {
	return true
}

func preGenerateDescribeStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.DescribeStateMachineInput) *svcsdk.DescribeStateMachineInput {
	return obj
}

func postGenerateDescribeStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.DescribeStateMachineInput) *svcsdk.DescribeStateMachineInput {
	return obj
}

func preGenerateCreateStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.CreateStateMachineInput) *svcsdk.CreateStateMachineInput {
	return obj
}

func postGenerateCreateStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.CreateStateMachineInput) *svcsdk.CreateStateMachineInput {
	return obj
}
func preGenerateDeleteStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.DeleteStateMachineInput) *svcsdk.DeleteStateMachineInput {
	return obj
}

func postGenerateDeleteStateMachineInput(_ *svcapitypes.StateMachine, obj *svcsdk.DeleteStateMachineInput) *svcsdk.DeleteStateMachineInput {
	return obj
}
