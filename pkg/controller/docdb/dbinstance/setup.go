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
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errNotDBInstance    = "managed resource is not a DocDB instance custom resource"
	errKubeUpdateFailed = "cannot update DocDB instance custom resource"
)

// SetupDBInstance adds a controller that reconciles a DBInstance.
func SetupDBInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBInstanceGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBInstance{}).
		WithOptions(o.ForControllerRuntime()).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
	obj.DBInstanceIdentifier = awsclient.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclient.StringValue(cr.Status.AtProvider.DBInstanceStatus) {
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

func (e *hooks) isUpToDate(cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput) (bool, error) { // nolint:gocyclo
	instance := resp.DBInstances[0]

	switch {
	case awsclient.BoolValue(cr.Spec.ForProvider.AutoMinorVersionUpgrade) != awsclient.BoolValue(instance.AutoMinorVersionUpgrade),
		awsclient.StringValue(cr.Spec.ForProvider.CACertificateIdentifier) != awsclient.StringValue(instance.CACertificateIdentifier),
		awsclient.StringValue(cr.Spec.ForProvider.DBInstanceClass) != awsclient.StringValue(instance.DBInstanceClass),
		awsclient.StringValue(cr.Spec.ForProvider.PreferredMaintenanceWindow) != awsclient.StringValue(instance.PreferredMaintenanceWindow),
		awsclient.Int64Value(cr.Spec.ForProvider.PromotionTier) != awsclient.Int64Value(instance.PromotionTier):
		return false, nil
	}

	return svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, instance.DBInstanceArn)
}

func lateInitialize(cr *svcapitypes.DBInstanceParameters, resp *svcsdk.DescribeDBInstancesOutput) error {
	instance := resp.DBInstances[0]

	cr.AvailabilityZone = awsclient.LateInitializeStringPtr(cr.AvailabilityZone, instance.AvailabilityZone)
	cr.AutoMinorVersionUpgrade = awsclient.LateInitializeBoolPtr(cr.AutoMinorVersionUpgrade, instance.AutoMinorVersionUpgrade)
	cr.CACertificateIdentifier = awsclient.LateInitializeStringPtr(cr.CACertificateIdentifier, instance.CACertificateIdentifier)
	cr.PreferredMaintenanceWindow = awsclient.LateInitializeStringPtr(cr.PreferredMaintenanceWindow, instance.PreferredMaintenanceWindow)
	cr.PromotionTier = awsclient.LateInitializeInt64Ptr(cr.PromotionTier, instance.PromotionTier)
	return nil
}

func preUpdate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.ModifyDBInstanceInput) error {
	obj.DBInstanceIdentifier = awsclient.String(meta.GetExternalName(cr))
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
	obj.DBInstanceIdentifier = awsclient.String(meta.GetExternalName(cr))
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
	if awsclient.StringValue(cr.Status.AtProvider.DBInstanceStatus) == svcapitypes.DocDBInstanceStateDeleting {
		return true, nil
	}

	obj.DBInstanceIdentifier = awsclient.String(meta.GetExternalName(cr))
	return false, nil
}

func filterList(cr *svcapitypes.DBInstance, list *svcsdk.DescribeDBInstancesOutput) *svcsdk.DescribeDBInstancesOutput {
	id := meta.GetExternalName(cr)
	for _, instance := range list.DBInstances {
		if awsclient.StringValue(instance.DBInstanceIdentifier) == id {
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
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(awsclient.StringValue(cr.Status.AtProvider.Endpoint.Address)),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(awsclient.Int64Value(cr.Status.AtProvider.Endpoint.Port)))),
	}
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.DBInstance)
	if !ok {
		return errors.New(errNotDBInstance)
	}

	cr.Spec.ForProvider.Tags = svcutils.AddExternalTags(mg, cr.Spec.ForProvider.Tags)
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
