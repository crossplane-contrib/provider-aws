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

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dynamodb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupBackup adds a controller that reconciles Backup.
func SetupBackup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.BackupGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.BackupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Backup{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Backup, obj *svcsdk.DescribeBackupInput) error {
	obj.BackupArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.DescribeBackupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.BackupDescription.BackupDetails.BackupStatus) {
	case string(svcapitypes.BackupStatus_SDK_AVAILABLE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.BackupStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.BackupStatus_SDK_DELETED):
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Backup, obj *svcsdk.CreateBackupInput) error {
	obj.TableName = aws.String(cr.Spec.ForProvider.TableName)
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.CreateBackupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.BackupDetails.BackupArn))
	return cre, err
}

func preDelete(_ context.Context, cr *svcapitypes.Backup, obj *svcsdk.DeleteBackupInput) (bool, error) {
	obj.BackupArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
