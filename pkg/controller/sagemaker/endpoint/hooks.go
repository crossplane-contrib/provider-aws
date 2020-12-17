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

package endpoint

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// SetupEndpoint adds a controller that reconciles Endpoint.
func SetupEndpoint(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.EndpointGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Endpoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.EndpointGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Endpoint) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Endpoint, _ *svcsdk.DescribeEndpointOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.Endpoint) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.Endpoint, _ *svcsdk.CreateEndpointOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Endpoint) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Endpoint, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.EndpointParameters, *svcsdk.DescribeEndpointOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.Endpoint, *svcsdk.DescribeEndpointOutput) bool {
	return true
}

func preGenerateDescribeEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.DescribeEndpointInput) *svcsdk.DescribeEndpointInput {
	return obj
}

func postGenerateDescribeEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.DescribeEndpointInput) *svcsdk.DescribeEndpointInput {
	return obj
}

func preGenerateCreateEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.CreateEndpointInput) *svcsdk.CreateEndpointInput {
	return obj
}

func postGenerateCreateEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.CreateEndpointInput) *svcsdk.CreateEndpointInput {
	return obj
}
func preGenerateDeleteEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.DeleteEndpointInput) *svcsdk.DeleteEndpointInput {
	return obj
}

func postGenerateDeleteEndpointInput(_ *svcapitypes.Endpoint, obj *svcsdk.DeleteEndpointInput) *svcsdk.DeleteEndpointInput {
	return obj
}
