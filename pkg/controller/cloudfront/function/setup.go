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

package function

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
)

// SetupCloudFrontFunction adds a controller that reconciles cloudfrontfunction .
func SetupCloudFrontFunction(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FunctionGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Function{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.FunctionGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube: mgr.GetClient(),
				opts: []option{
					func(e *external) {
						e.preObserve = preObserve
						e.postObserve = postObserve
						e.preCreate = preCreate
						e.postCreate = postCreate
						e.preUpdate = preUpdate
						e.isUpToDate = isUpToDate
						e.preDelete = preDelete
					},
				},
			}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preCreate(_ context.Context, cr *svcapitypes.Function, cdi *svcsdk.CreateFunctionInput) error {
	return nil
}

func postCreate(_ context.Context, cp *svcapitypes.Function, cpo *svcsdk.CreateFunctionOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return ec, nil
}

func preObserve(_ context.Context, cp *svcapitypes.Function, gpi *svcsdk.GetFunctionInput) error {
	return nil
}

func postObserve(_ context.Context, cp *svcapitypes.Function, _ *svcsdk.GetFunctionOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return eo, nil
}

func preUpdate(_ context.Context, cp *svcapitypes.Function, upi *svcsdk.UpdateFunctionInput) error {
	return nil
}

func preDelete(_ context.Context, cp *svcapitypes.Function, dpi *svcsdk.DeleteFunctionInput) (bool, error) {
	return false, nil
}

func isUpToDate(cp *svcapitypes.Function, gpo *svcsdk.GetFunctionOutput) (bool, error) {
	return true, nil
}
