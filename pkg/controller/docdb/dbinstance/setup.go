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

package dbinstance

import (
	"context"
	"strconv"

	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupDBInstance adds a controller that reconciles a DBInstance.
func SetupDBInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBInstanceGroupKind)
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
		resource.ManagedKind(svcapitypes.DBInstanceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBInstance{}).
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
	e.lateInitialize = lateInitialize
	e.postCreate = postCreate
	e.preDelete = preDelete
	e.filterList = filterList
}

type hooks struct {
	client docdbiface.DocDBAPI
	kube   client.Client
}

func preObserve(_ context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DescribeDBInstancesInput) error {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(cr.Status.AtProvider.DBInstanceStatus) {
	case svcapitypes.DocDBInstanceStateAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case svcapitypes.DocDBInstanceStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case svcapitypes.DocDBInstanceStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	obs.ConnectionDetails = getConnectionDetails(cr)
	return obs, nil
}

func (e *hooks) isUpToDate(_ context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput) (bool, string, error) {
	instance := resp.DBInstances[0]

	switch {
	case pointer.BoolValue(cr.Spec.ForProvider.AutoMinorVersionUpgrade) != pointer.BoolValue(instance.AutoMinorVersionUpgrade),
		pointer.StringValue(cr.Spec.ForProvider.CACertificateIdentifier) != pointer.StringValue(instance.CACertificateIdentifier),
		pointer.StringValue(cr.Spec.ForProvider.DBInstanceClass) != pointer.StringValue(instance.DBInstanceClass),
		pointer.StringValue(cr.Spec.ForProvider.PreferredMaintenanceWindow) != pointer.StringValue(instance.PreferredMaintenanceWindow),
		pointer.Int64Value(cr.Spec.ForProvider.PromotionTier) != pointer.Int64Value(instance.PromotionTier):
		return false, "", nil
	}

	areTagsUpToDate, err := svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, instance.DBInstanceArn)
	return areTagsUpToDate, "", err
}

func lateInitialize(cr *svcapitypes.DBInstanceParameters, resp *svcsdk.DescribeDBInstancesOutput) error {
	instance := resp.DBInstances[0]

	cr.AvailabilityZone = pointer.LateInitialize(cr.AvailabilityZone, instance.AvailabilityZone)
	cr.AutoMinorVersionUpgrade = pointer.LateInitialize(cr.AutoMinorVersionUpgrade, instance.AutoMinorVersionUpgrade)
	cr.CACertificateIdentifier = pointer.LateInitialize(cr.CACertificateIdentifier, instance.CACertificateIdentifier)
	cr.PreferredMaintenanceWindow = pointer.LateInitialize(cr.PreferredMaintenanceWindow, instance.PreferredMaintenanceWindow)
	cr.PromotionTier = pointer.LateInitialize(cr.PromotionTier, instance.PromotionTier)
	return nil
}

func preUpdate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.ModifyDBInstanceInput) error {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.CACertificateIdentifier = cr.Spec.ForProvider.CACertificateIdentifier
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately
	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.ModifyDBInstanceOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}
	return upd, svcutils.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.DBInstanceARN)
}

func preCreate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.CreateDBInstanceInput) error {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.DBClusterIdentifier = cr.Spec.ForProvider.DBClusterIdentifier
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.CreateDBInstanceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return cre, err
	}

	cre.ConnectionDetails = getConnectionDetails(cr)
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DeleteDBInstanceInput) (bool, error) {
	// Skip if cluster is already in deleting state
	if pointer.StringValue(cr.Status.AtProvider.DBInstanceStatus) == svcapitypes.DocDBInstanceStateDeleting {
		return true, nil
	}

	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func filterList(cr *svcapitypes.DBInstance, list *svcsdk.DescribeDBInstancesOutput) *svcsdk.DescribeDBInstancesOutput {
	id := meta.GetExternalName(cr)
	for _, instance := range list.DBInstances {
		if pointer.StringValue(instance.DBInstanceIdentifier) == id {
			return &svcsdk.DescribeDBInstancesOutput{
				Marker:      list.Marker,
				DBInstances: []*svcsdk.DBInstance{instance},
			}
		}
	}

	return &svcsdk.DescribeDBInstancesOutput{
		Marker:      list.Marker,
		DBInstances: []*svcsdk.DBInstance{},
	}
}

func getConnectionDetails(cr *svcapitypes.DBInstance) managed.ConnectionDetails {
	if cr.Status.AtProvider.Endpoint == nil {
		return nil
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(pointer.StringValue(cr.Status.AtProvider.Endpoint.Address)),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(pointer.Int64Value(cr.Status.AtProvider.Endpoint.Port)))),
	}
}
