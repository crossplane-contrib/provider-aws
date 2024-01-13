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

package thing

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/iot"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	ctrl "sigs.k8s.io/controller-runtime"

	iottypes "github.com/crossplane-contrib/provider-aws/apis/iot/v1alpha1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/iot/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupThing adds a controller that reconciles Thing.
func SetupThing(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(iottypes.ThingGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
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
		resource.ManagedKind(iottypes.ThingGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&iottypes.Thing{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Thing, obj *svcsdk.DescribeThingInput) error {
	obj.ThingName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Thing, _ *svcsdk.DescribeThingOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Thing, obj *svcsdk.CreateThingInput) error {
	obj.ThingName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Thing, resp *svcsdk.CreateThingOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, aws.StringValue(resp.ThingName))
	return cre, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Thing, obj *svcsdk.UpdateThingInput) error {
	obj.ThingName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Thing, obj *svcsdk.DeleteThingInput) (bool, error) {
	obj.ThingName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Thing, resp *svcsdk.DescribeThingOutput) (bool, string, error) {
	patch, err := createPatch(resp, &cr.Spec.ForProvider)

	if err != nil {
		return false, "", err
	}

	diff := cmp.Diff(&svcapitypes.ThingParameters{}, patch)
	return diff == "", diff, nil
}

func createPatch(in *svcsdk.DescribeThingOutput, target *svcapitypes.ThingParameters) (*svcapitypes.ThingParameters, error) {
	jsonPatch, err := jsonpatch.CreateJSONPatch(in, target)
	if err != nil {
		return nil, err
	}

	patch := &svcapitypes.ThingParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}

	return patch, nil
}
