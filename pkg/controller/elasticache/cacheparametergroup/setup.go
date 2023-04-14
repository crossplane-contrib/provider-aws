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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elasticache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// ManagesKind returns the kind this controller manages
func ManagesKind() string {
	return svcapitypes.CacheParameterGroupGroupKind
}

// SetupCacheParameterGroup adds a controller that reconciles a CacheParameterGroup.
func SetupCacheParameterGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CacheParameterGroupGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.CacheParameterGroup{}).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.CacheParameterGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
	obj.CacheParameterGroupName = awsclient.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.CacheParameterGroup, resp *svcsdk.DescribeCacheParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(v1.Available())
	return obs, nil
}

func (e *hooks) isUpToDate(cr *svcapitypes.CacheParameterGroup, resp *svcsdk.DescribeCacheParameterGroupsOutput) (bool, error) {
	ctx := context.TODO()

	input := &svcsdk.DescribeCacheParametersInput{
		CacheParameterGroupName: awsclient.String(meta.GetExternalName(cr)),
	}
	var results []*svcsdk.Parameter
	err := e.client.DescribeCacheParametersPagesWithContext(ctx, input, func(page *svcsdk.DescribeCacheParametersOutput, lastPage bool) bool {
		results = append(results, page.Parameters...)
		return !lastPage
	})
	if err != nil {
		return false, err
	}

	var observed []svcapitypes.ParameterNameValue
	for _, v := range results {
		if svcapitypes.SourceType(awsclient.StringValue(v.Source)) == svcapitypes.SourceType_user {
			observed = append(observed, svcapitypes.ParameterNameValue{
				ParameterName:  v.ParameterName,
				ParameterValue: v.ParameterValue,
			})
		}
	}

	diff := cmp.Diff(observed, cr.Spec.ForProvider.ParameterNameValues, cmpopts.SortSlices(func(a, b svcapitypes.ParameterNameValue) bool {
		return awsclient.StringValue(a.ParameterName) < awsclient.StringValue(b.ParameterName)
	}))

	// TODO: We should be able to return the diff to crossplane-runtime here
	return diff == "", nil
}

func preUpdate(ctx context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.ModifyCacheParameterGroupInput) error {
	obj.CacheParameterGroupName = awsclient.String(meta.GetExternalName(cr))
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
	obj.CacheParameterGroupName = awsclient.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.CacheParameterGroup, obj *svcsdk.DeleteCacheParameterGroupInput) (bool, error) {
	obj.CacheParameterGroupName = awsclient.String(meta.GetExternalName(cr))
	return false, nil
}
