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

package activity

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sfn"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sfn/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupActivity adds a controller that reconciles Activity.
func SetupActivity(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.ActivityGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Activity{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ActivityGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Activity) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Activity, _ *svcsdk.DescribeActivityOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.Activity) error {
	return nil
}

func (*external) postCreate(_ context.Context, cr *svcapitypes.Activity, resp *svcsdk.CreateActivityOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.ActivityArn))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.Activity) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Activity, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.ActivityParameters, *svcsdk.DescribeActivityOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.Activity, *svcsdk.DescribeActivityOutput) bool {
	return true
}

func preGenerateDescribeActivityInput(_ *svcapitypes.Activity, obj *svcsdk.DescribeActivityInput) *svcsdk.DescribeActivityInput {
	return obj
}

func postGenerateDescribeActivityInput(cr *svcapitypes.Activity, obj *svcsdk.DescribeActivityInput) *svcsdk.DescribeActivityInput {
	obj.ActivityArn = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateActivityInput(_ *svcapitypes.Activity, obj *svcsdk.CreateActivityInput) *svcsdk.CreateActivityInput {
	return obj
}

func postGenerateCreateActivityInput(_ *svcapitypes.Activity, obj *svcsdk.CreateActivityInput) *svcsdk.CreateActivityInput {
	return obj
}
func preGenerateDeleteActivityInput(_ *svcapitypes.Activity, obj *svcsdk.DeleteActivityInput) *svcsdk.DeleteActivityInput {
	return obj
}

func postGenerateDeleteActivityInput(cr *svcapitypes.Activity, obj *svcsdk.DeleteActivityInput) *svcsdk.DeleteActivityInput {
	obj.ActivityArn = aws.String(meta.GetExternalName(cr))
	return obj
}
