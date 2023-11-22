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

package dbclusterparametergroup

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotDBClusterParameterGroup = "managed resource is not a DocDBClusterParameterGroup custom resource"
	errKubeUpdateFailed           = "cannot update DocDB DBClusterParameterGroup custom resource"
	errModifyFamily               = "cannot modify DBParameterGroupFamily of an existing DBClusterParameterGroup"
	errModifyDescription          = "cannot modify Description of an existing DBClusterParameterGroup"
	errDescribeParameters         = "cannot describe parameters for DBClusterParameterGroup"
)

// SetupDBClusterParameterGroup adds a controller that reconciles a DBClusterParameterGroup.
func SetupDBClusterParameterGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBClusterParameterGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.DBClusterParameterGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBClusterParameterGroup{}).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(r)
}

func setupExternal(e *external) {
	h := &hooks{client: e.client, kube: e.kube}
	e.preObserve = preObserve
	e.postObserve = postObserve
	e.isUpToDate = h.isUpToDate
	e.preUpdate = preUpdate
	e.postUpdate = h.postUpdate
	e.preCreate = preCreate
	e.preDelete = preDelete
	e.filterList = filterList
	e.lateInitialize = h.lateInitialize
}

type hooks struct {
	client docdbiface.DocDBAPI
	kube   client.Client
}

func preObserve(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, obj *svcsdk.DescribeDBClusterParameterGroupsInput) error {
	obj.DBClusterParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (e *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.DBClusterParameterGroup, resp *svcsdk.DescribeDBClusterParameterGroupsOutput) (bool, string, error) {
	group := resp.DBClusterParameterGroups[0]

	if pointer.StringValue(cr.Spec.ForProvider.DBParameterGroupFamily) != pointer.StringValue(group.DBParameterGroupFamily) {
		return false, "", errors.New(errModifyFamily)
	}

	if pointer.StringValue(cr.Spec.ForProvider.Description) != pointer.StringValue(group.Description) {
		return false, "", errors.New(errModifyDescription)
	}

	paramsReq := &svcsdk.DescribeDBClusterParametersInput{DBClusterParameterGroupName: group.DBClusterParameterGroupName}
	paramsResp, err := e.client.DescribeDBClusterParametersWithContext(ctx, paramsReq)
	if err != nil {
		return false, "", err
	}

	if !areParametersEqual(cr.Spec.ForProvider.Parameters, paramsResp.Parameters) {
		return false, "", nil
	}

	areTagsUpToDate, err := svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, group.DBClusterParameterGroupArn)
	return areTagsUpToDate, "", err
}

func postObserve(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, _ *svcsdk.DescribeDBClusterParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	cr.SetConditions(v1.Available())
	return obs, err
}

func (e *hooks) lateInitialize(cr *svcapitypes.DBClusterParameterGroupParameters, resp *svcsdk.DescribeDBClusterParameterGroupsOutput) error {
	group := resp.DBClusterParameterGroups[0]

	// TODO: Add context to lateInitialize
	ctx := context.TODO()
	paramsReq := &svcsdk.DescribeDBClusterParametersInput{DBClusterParameterGroupName: group.DBClusterParameterGroupName}
	paramsResp, err := e.client.DescribeDBClusterParametersWithContext(ctx, paramsReq)
	if err != nil {
		return errors.Wrap(err, errDescribeParameters)
	}

	cr.CustomDBClusterParameterGroupParameters.Parameters = lateInitializeParameters(cr.Parameters, paramsResp.Parameters)
	return nil
}

func lateInitializeParameters(in []*svcapitypes.CustomParameter, from []*svcsdk.Parameter) []*svcapitypes.CustomParameter {
	out := in
	if out == nil {
		out = []*svcapitypes.CustomParameter{}
	}

	apiParams := make(map[string]*svcapitypes.CustomParameter, len(out))
	for _, p := range out {
		apiParams[pointer.StringValue(p.ParameterName)] = p
	}

	for _, sdkP := range from {
		if _, exists := apiParams[pointer.StringValue(sdkP.ParameterName)]; !exists {
			newP := &svcapitypes.CustomParameter{}
			generateAPIParameter(sdkP, newP)
			out = append(out, newP)
		}
	}

	return out
}

func preUpdate(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, obj *svcsdk.ModifyDBClusterParameterGroupInput) error {
	obj.DBClusterParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.Parameters = generateSdkParameters(cr.Spec.ForProvider.Parameters)
	return nil
}

func (e *hooks) postUpdate(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, resp *svcsdk.ModifyDBClusterParameterGroupOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}
	cr.Status.SetConditions(v1.Available())
	return upd, svcutils.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.DBClusterParameterGroupARN)
}

func preCreate(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, obj *svcsdk.CreateDBClusterParameterGroupInput) error {
	obj.DBClusterParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	// CreateDBClusterParameterGroup does not create the parameters themselves. Parameters are added during update.
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBClusterParameterGroup, obj *svcsdk.DeleteDBClusterParameterGroupInput) (bool, error) {
	obj.DBClusterParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func filterList(cr *svcapitypes.DBClusterParameterGroup, list *svcsdk.DescribeDBClusterParameterGroupsOutput) *svcsdk.DescribeDBClusterParameterGroupsOutput {
	id := meta.GetExternalName(cr)
	for _, instance := range list.DBClusterParameterGroups {
		if pointer.StringValue(instance.DBClusterParameterGroupName) == id {
			return &svcsdk.DescribeDBClusterParameterGroupsOutput{
				Marker:                   list.Marker,
				DBClusterParameterGroups: []*svcsdk.DBClusterParameterGroup{instance},
			}
		}
	}

	return &svcsdk.DescribeDBClusterParameterGroupsOutput{
		Marker:                   list.Marker,
		DBClusterParameterGroups: []*svcsdk.DBClusterParameterGroup{},
	}
}

func areParametersEqual(spec []*svcapitypes.CustomParameter, current []*svcsdk.Parameter) bool {
	currentMap := make(map[string]*svcsdk.Parameter, len(current))
	for _, currentParam := range current {
		currentMap[pointer.StringValue(currentParam.ParameterName)] = currentParam
	}

	for _, specParam := range spec {
		currentParam, exists := currentMap[pointer.StringValue(specParam.ParameterName)]
		if !exists || !cmp.Equal(
			specParam.ParameterValue,
			generateAPIParameter(currentParam, &svcapitypes.CustomParameter{}).ParameterValue,
		) {
			return false
		}
	}

	return true
}

func generateSdkParameters(params []*svcapitypes.CustomParameter) []*svcsdk.Parameter {
	sdkParams := make([]*svcsdk.Parameter, len(params))
	for i, p := range params {
		sdkParams[i] = &svcsdk.Parameter{
			ApplyMethod:    p.ApplyMethod,
			ParameterName:  p.ParameterName,
			ParameterValue: p.ParameterValue,
		}
	}

	return sdkParams
}

func generateAPIParameter(p *svcsdk.Parameter, o *svcapitypes.CustomParameter) *svcapitypes.CustomParameter {
	o.ApplyMethod = p.ApplyMethod
	o.ParameterName = p.ParameterName
	o.ParameterValue = p.ParameterValue

	return o
}
