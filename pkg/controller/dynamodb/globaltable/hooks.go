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

package globaltable

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupGlobalTable adds a controller that reconciles GlobalTable.
func SetupGlobalTable(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.GlobalTableGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.postObserve = postObserve
			d := &deleter{client: e.client}
			e.delete = d.delete
			u := &updater{client: e.client}
			e.preUpdate = u.preUpdate
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
		resource.ManagedKind(svcapitypes.GlobalTableGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.GlobalTable{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.GlobalTable, obj *svcsdk.DescribeGlobalTableInput) error {
	obj.GlobalTableName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.GlobalTable, obj *svcsdk.CreateGlobalTableInput) error {
	obj.GlobalTableName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.GlobalTable, resp *svcsdk.DescribeGlobalTableOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.GlobalTableDescription.GlobalTableStatus) {
	case string(svcapitypes.GlobalTableStatus_SDK_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.GlobalTableStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.GlobalTableStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}
	return obs, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.GlobalTable, obj *svcsdk.DescribeGlobalTableOutput) (bool, string, error) {
	existing := make([]string, len(obj.GlobalTableDescription.ReplicationGroup))
	for i, r := range obj.GlobalTableDescription.ReplicationGroup {
		existing[i] = aws.StringValue(r.RegionName)
	}
	sort.Strings(existing)
	desired := make([]string, len(cr.Spec.ForProvider.ReplicationGroup))
	for i, r := range cr.Spec.ForProvider.ReplicationGroup {
		desired[i] = aws.StringValue(r.RegionName)
	}
	sort.Strings(desired)
	diff := cmp.Diff(existing, desired)
	return diff == "", diff, nil
}

type updater struct {
	client svcsdkapi.DynamoDBAPI
}

func (u *updater) preUpdate(ctx context.Context, cr *svcapitypes.GlobalTable, obj *svcsdk.UpdateGlobalTableInput) error {
	input := GenerateDescribeGlobalTableInput(cr)
	o, err := u.client.DescribeGlobalTableWithContext(ctx, input)
	if err != nil {
		return errorutils.Wrap(err, errDescribe)
	}
	desired := map[string]bool{}
	for _, r := range cr.Spec.ForProvider.ReplicationGroup {
		desired[aws.StringValue(r.RegionName)] = true
	}
	var del []string
	var add []string
	existing := map[string]bool{}
	for _, r := range o.GlobalTableDescription.ReplicationGroup {
		if _, ok := desired[aws.StringValue(r.RegionName)]; !ok {
			del = append(del, aws.StringValue(r.RegionName))
		}
		existing[aws.StringValue(r.RegionName)] = true
	}
	for regionName := range desired {
		if _, ok := existing[regionName]; !ok {
			add = append(add, regionName)
		}
	}
	obj.GlobalTableName = aws.String(meta.GetExternalName(cr))
	for _, rn := range add {
		obj.ReplicaUpdates = append(obj.ReplicaUpdates, &svcsdk.ReplicaUpdate{Create: &svcsdk.CreateReplicaAction{RegionName: aws.String(rn)}})
	}
	for _, rn := range del {
		obj.ReplicaUpdates = append(obj.ReplicaUpdates, &svcsdk.ReplicaUpdate{Delete: &svcsdk.DeleteReplicaAction{RegionName: aws.String(rn)}})
	}
	return nil
}

type deleter struct {
	client svcsdkapi.DynamoDBAPI
}

func (d *deleter) delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.GlobalTable)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	u := &svcsdk.UpdateGlobalTableInput{
		GlobalTableName: aws.String(meta.GetExternalName(mg)),
	}
	for _, region := range cr.Spec.ForProvider.ReplicationGroup {
		u.ReplicaUpdates = append(u.ReplicaUpdates, &svcsdk.ReplicaUpdate{Delete: &svcsdk.DeleteReplicaAction{RegionName: region.RegionName}})
	}
	if _, err := d.client.UpdateGlobalTableWithContext(ctx, u); err != nil {
		return errorutils.Wrap(err, "update call for deletion failed")
	}
	return nil
}
