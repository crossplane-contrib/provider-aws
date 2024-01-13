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

package restapi

import (
	"context"
	"encoding/json"

	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
	apigwclient "github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupRestAPI adds a controller that reconciles RestAPI.
func SetupRestAPI(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RestAPIGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.postCreate = postCreate
			c := &custom{
				Client: &apigwclient.GatewayClient{Client: e.client},
			}
			e.preUpdate = c.preUpdate
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
		resource.ManagedKind(svcapitypes.RestAPIGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.RestAPI{}).
		Complete(r)
}

type custom struct {
	Client apigwclient.Client
}

func preObserve(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.GetRestApiInput) error {
	obj.RestApiId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.RestApi, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.RestAPI, resp *svcsdk.RestApi, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(resp.Id))
	return cre, nil
}

func (c *custom) preUpdate(ctx context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.UpdateRestApiInput) error {

	rapi, err := c.Client.GetRestAPIByID(ctx, pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)))
	if err != nil {
		return errors.Wrap(err, "cannot get rest api")
	}

	cur := &svcapitypes.RestAPIParameters{
		Name:   rapi.Name,
		Region: cr.Spec.ForProvider.Region,
	}

	if err := lateInitialize(cur, rapi); err != nil {
		return errors.Wrap(err, "cannot late init current restApi")
	}

	err = lateInitializePolicies(&cr.Spec.ForProvider, rapi)
	if err != nil {
		return errors.Wrap(err, "comparing spec and current policies post late init")
	}

	pOps, err := apigwclient.GetPatchOperations(&cur, cr.Spec.ForProvider)
	if err != nil {
		return errors.Wrap(err, "cannot compute patch preUpdate")
	}

	obj.PatchOperations = pOps
	obj.RestApiId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.DeleteRestApiInput) (bool, error) {
	obj.RestApiId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.RestAPI, cur *svcsdk.RestApi) (bool, string, error) {
	s := &svcapitypes.RestAPIParameters{
		Name:   cur.Name,
		Region: cr.Spec.ForProvider.Region,
	}

	var err error

	if err = lateInitialize(s, cur); err != nil {
		return false, "", errors.Wrap(err, "cannot lateinit")
	}

	patchJSON, err := jsonpatch.CreateJSONPatch(cr.Spec.ForProvider, &s)
	if err != nil {
		return false, "", errors.Wrap(err, "error checking up to date")
	}

	patch := &svcapitypes.RestAPIParameters{}
	if err := json.Unmarshal(patchJSON, patch); err != nil {
		return false, "", errors.Wrap(err, "error checking up to date")
	}

	diff := cmp.Diff(&svcapitypes.RestAPIParameters{}, patch,
		cmpopts.IgnoreTypes([]xpv1.Reference{}, []xpv1.Selector{}),
		cmpopts.IgnoreFields(svcapitypes.RestAPIParameters{}, "Region"),
	)
	return diff == "", diff, nil
}

func lateInitialize(in *svcapitypes.RestAPIParameters, cur *svcsdk.RestApi) error {
	in.APIKeySource = pointer.LateInitialize(in.APIKeySource, cur.ApiKeySource)
	in.BinaryMediaTypes = pointer.LateInitializeSlice(in.BinaryMediaTypes, cur.BinaryMediaTypes)
	in.Description = pointer.LateInitialize(in.Description, cur.Description)
	in.DisableExecuteAPIEndpoint = pointer.LateInitialize(in.DisableExecuteAPIEndpoint, cur.DisableExecuteApiEndpoint)
	in.MinimumCompressionSize = pointer.LateInitialize(in.MinimumCompressionSize, cur.MinimumCompressionSize)

	if cur.EndpointConfiguration != nil {
		if in.EndpointConfiguration == nil {
			in.EndpointConfiguration = &svcapitypes.EndpointConfiguration{}
		}
		in.EndpointConfiguration.Types = pointer.LateInitializeSlice(in.EndpointConfiguration.Types, cur.EndpointConfiguration.Types)
		in.EndpointConfiguration.VPCEndpointIDs = pointer.LateInitializeSlice(in.EndpointConfiguration.VPCEndpointIDs, cur.EndpointConfiguration.VpcEndpointIds)
	}

	return lateInitializePolicies(in, cur)
}

func lateInitializePolicies(in *svcapitypes.RestAPIParameters, cur *svcsdk.RestApi) error {
	inPol, err := policyStringToMap(in.Policy)
	if err != nil {
		return errors.Wrap(err, "converting spec policy to map")
	}

	curPol, err := policyEscapedStringToMap(cur.Policy)
	if err != nil {
		curPol, err = policyStringToMap(cur.Policy)
		if err != nil {
			return errors.Wrap(err, "converting current policy to map")
		}
	}

	// this is a hack since AWS does minor adaptions to the policy after creation. We want to treat just that case
	// and avoid copying from the AWS status of a resources as the source of truth
	res, err := policiesAreKindOfTheSame(inPol, curPol)
	if err != nil {
		return errors.Wrap(err, "policies could not be compared")
	}

	cur.Policy, err = policyMapToString(curPol)

	if err != nil {
		return err
	}
	if res {
		in.Policy = cur.Policy
	} else {
		in.Policy, err = policyMapToString(inPol)
		if err != nil {
			return err
		}
	}
	in.Policy = pointer.LateInitialize(in.Policy, cur.Policy)

	return err
}
