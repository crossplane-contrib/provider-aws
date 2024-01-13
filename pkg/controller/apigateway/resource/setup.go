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

package resource

import (
	"context"
	"encoding/json"

	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
	apigwclient "github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupResource adds a controller that reconciles Resource.
func SetupResource(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ResourceGroupKind)
	opts := []option{
		func(e *external) {
			e.lateInitialize = lateInitialize
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate
			c := &custom{
				Client: &apigwclient.GatewayClient{Client: e.client},
			}
			e.preUpdate = c.preUpdate
			e.preCreate = c.preCreate
		},
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ResourceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Resource{}).
		Complete(r)
}

type custom struct {
	Client apigwclient.Client
}

func (c *custom) preCreate(ctx context.Context, cr *svcapitypes.Resource, obj *svcsdk.CreateResourceInput) error {
	obj.RestApiId = cr.Spec.ForProvider.RestAPIID

	if cr.Spec.ForProvider.ParentResourceID != nil {
		obj.ParentId = cr.Spec.ForProvider.ParentResourceID
	} else if resourceID, err := c.Client.GetRestAPIRootResource(ctx, obj.RestApiId); err != nil {
		return errors.Wrap(err, "could not get root resource for api")
	} else {
		obj.ParentId = resourceID
	}

	return nil
}

func (c *custom) preUpdate(ctx context.Context, cr *svcapitypes.Resource, obj *svcsdk.UpdateResourceInput) error {
	in := &svcsdk.GetResourceInput{
		RestApiId:  cr.Spec.ForProvider.RestAPIID,
		ResourceId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}

	cur := &svcapitypes.ResourceParameters{
		Region: cr.Spec.ForProvider.Region,
		CustomResourceParameters: svcapitypes.CustomResourceParameters{
			RestAPIID: cr.Spec.ForProvider.RestAPIID,
		},
	}
	if r, err := c.Client.GetResource(ctx, in); err != nil {
		return errors.Wrap(err, "cannot retrieve resource")
	} else if err := lateInitialize(cur, r); err != nil {
		return errors.Wrap(err, "cannot late init")
	}

	pOps, err := apigwclient.GetPatchOperations(cur, &cr.Spec.ForProvider)
	if err != nil {
		return errors.Wrap(err, "cannot compute patch")
	}
	obj.PatchOperations = pOps

	return nil
}

func preObserve(_ context.Context, cr *svcapitypes.Resource, obj *svcsdk.GetResourceInput) error {
	obj.RestApiId = cr.Spec.ForProvider.RestAPIID
	obj.ResourceId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Resource, _ *svcsdk.Resource, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Resource, resp *svcsdk.Resource, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(resp.Id))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Resource, obj *svcsdk.DeleteResourceInput) (bool, error) {
	obj.ResourceId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.RestApiId = cr.Spec.ForProvider.RestAPIID

	return false, nil
}

func lateInitialize(cr *svcapitypes.ResourceParameters, cur *svcsdk.Resource) error {
	cr.PathPart = pointer.LateInitialize(cr.PathPart, cur.PathPart)
	cr.ParentResourceID = pointer.LateInitialize(cr.ParentResourceID, cur.ParentId)
	return nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Resource, cur *svcsdk.Resource) (bool, string, error) {
	patchJSON, err := jsonpatch.CreateJSONPatch(cr.Spec.ForProvider, cur)
	if err != nil {
		return true, "", errors.Wrap(err, "error checking up to date")
	}

	patch := &svcapitypes.ResourceParameters{}
	if err := json.Unmarshal(patchJSON, &patch); err != nil {
		return true, "", errors.Wrap(err, "error checking up to date")
	}

	diff := cmp.Diff(&svcapitypes.ResourceParameters{}, patch,
		cmpopts.IgnoreTypes([]xpv1.Reference{}, []xpv1.Selector{}),
		cmpopts.IgnoreFields(svcapitypes.ResourceParameters{}, "Region"),
	)
	return diff == "", diff, nil
}
