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

package cacheparametergroup

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elasticache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupCacheParameterGroup adds a controller that reconciles a CacheParameterGroup.
func SetupCacheParameterGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CacheParameterGroupKind)
	opts := []option{setupExternal}

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
		resource.ManagedKind(svcapitypes.CacheParameterGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.CacheParameterGroup{}).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(r)
}

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	h := &hooks{client: e.client, kube: e.kube}
	e.isUpToDate = h.isUpToDate
	e.preUpdate = preUpdate
	e.postUpdate = h.postUpdate
	e.preCreate = preCreate
	e.preDelete = preDelete
}

type hooks struct {
	client elasticacheiface.ElastiCacheAPI
	kube   client.Client
}

func preObserve(_ context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.DescribeCacheParameterGroupsInput) error {
	obj.CacheParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.CacheParameterGroup, resp *svcsdk.DescribeCacheParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(v1.Available())
	return obs, nil
}

func (e *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.CacheParameterGroup, resp *svcsdk.DescribeCacheParameterGroupsOutput) (bool, string, error) {
	input := &svcsdk.DescribeCacheParametersInput{
		CacheParameterGroupName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}
	var results []*svcsdk.Parameter
	err := e.client.DescribeCacheParametersPagesWithContext(ctx, input, func(page *svcsdk.DescribeCacheParametersOutput, lastPage bool) bool {
		results = append(results, page.Parameters...)
		return !lastPage
	})
	if err != nil {
		return false, "", err
	}

	var observed []svcapitypes.ParameterNameValue
	for _, v := range results {
		if svcapitypes.SourceType(pointer.StringValue(v.Source)) == svcapitypes.SourceType_user {
			observed = append(observed, svcapitypes.ParameterNameValue{
				ParameterName:  v.ParameterName,
				ParameterValue: v.ParameterValue,
			})
		}
	}

	diff := cmp.Diff(observed, cr.Spec.ForProvider.ParameterNameValues, cmpopts.SortSlices(func(a, b svcapitypes.ParameterNameValue) bool {
		return pointer.StringValue(a.ParameterName) < pointer.StringValue(b.ParameterName)
	}))

	return diff == "", diff, nil
}

func preUpdate(ctx context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.ModifyCacheParameterGroupInput) error {
	obj.CacheParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.ParameterNameValues = make([]*svcsdk.ParameterNameValue, len(cr.Spec.ForProvider.ParameterNameValues))

	for i, v := range cr.Spec.ForProvider.ParameterNameValues {
		obj.ParameterNameValues[i] = &svcsdk.ParameterNameValue{
			ParameterName:  v.ParameterName,
			ParameterValue: v.ParameterValue,
		}
	}

	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.CacheParameterGroup, resp *svcsdk.CacheParameterGroupNameMessage, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	cr.Status.SetConditions(v1.Available())
	return upd, nil
}

func preCreate(ctx context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.CreateCacheParameterGroupInput) error {
	obj.CacheParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.DeleteCacheParameterGroupInput) (bool, error) {
	obj.CacheParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}
