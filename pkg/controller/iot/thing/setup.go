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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/iot"
	"github.com/google/go-cmp/cmp"
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

	iottypes "github.com/crossplane/provider-aws/apis/iot/v1alpha1"
	svcapitypes "github.com/crossplane/provider-aws/apis/iot/v1alpha1"
	aws2 "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupThing adds a controller that reconciles Thing.
func SetupThing(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&iottypes.Thing{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(iottypes.ThingGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func isUpToDate(cr *svcapitypes.Thing, resp *svcsdk.DescribeThingOutput) (bool, error) {
	patch, err := createPatch(resp, &cr.Spec.ForProvider)

	if err != nil {
		return false, err
	}

	return cmp.Equal(&svcapitypes.ThingParameters{}, patch), nil
}

func createPatch(in *svcsdk.DescribeThingOutput, target *svcapitypes.ThingParameters) (*svcapitypes.ThingParameters, error) {
	jsonPatch, err := aws2.CreateJSONPatch(in, target)
	if err != nil {
		return nil, err
	}

	patch := &svcapitypes.ThingParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}

	return patch, nil
}
