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

	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	svcsdkapi "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
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

	svcapitypes "github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupGlobalTable adds a controller that reconciles GlobalTable.
func SetupGlobalTable(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.GlobalTable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.GlobalTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func isUpToDate(cr *svcapitypes.GlobalTable, obj *svcsdk.DescribeGlobalTableOutput) (bool, error) {
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
	return cmp.Equal(existing, desired), nil
}

type updater struct {
	client svcsdkapi.DynamoDBAPI
}

func (u *updater) preUpdate(ctx context.Context, cr *svcapitypes.GlobalTable, obj *svcsdk.UpdateGlobalTableInput) error {
	input := GenerateDescribeGlobalTableInput(cr)
	o, err := u.client.DescribeGlobalTableWithContext(ctx, input)
	if err != nil {
		return aws.Wrap(err, errDescribe)
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
		return aws.Wrap(err, "update call for deletion failed")
	}
	return nil
}
