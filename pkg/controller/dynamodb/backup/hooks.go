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

	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupBackup adds a controller that reconciles Backup.
func SetupBackup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.BackupGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Backup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.BackupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Backup) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.DescribeBackupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.BackupDescription.BackupDetails.BackupStatus) {
	case string(svcapitypes.BackupStatus_SDK_AVAILABLE):
		cr.SetConditions(v1alpha1.Available())
	case string(svcapitypes.BackupStatus_SDK_CREATING):
		cr.SetConditions(v1alpha1.Creating())
	case string(svcapitypes.BackupStatus_SDK_DELETED):
		cr.SetConditions(v1alpha1.Unavailable())
	}
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.Backup) error {
	return nil
}

func (*external) postCreate(_ context.Context, cr *svcapitypes.Backup, resp *svcsdk.CreateBackupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.BackupDetails.BackupArn))
	cre.ExternalNameAssigned = true
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Backup) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Backup, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.BackupParameters, *svcsdk.DescribeBackupOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.Backup, *svcsdk.DescribeBackupOutput) bool {
	return true
}

func preGenerateDescribeBackupInput(_ *svcapitypes.Backup, obj *svcsdk.DescribeBackupInput) *svcsdk.DescribeBackupInput {
	return obj
}

func postGenerateDescribeBackupInput(cr *svcapitypes.Backup, obj *svcsdk.DescribeBackupInput) *svcsdk.DescribeBackupInput {
	obj.BackupArn = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateBackupInput(_ *svcapitypes.Backup, obj *svcsdk.CreateBackupInput) *svcsdk.CreateBackupInput {
	return obj
}

func postGenerateCreateBackupInput(cr *svcapitypes.Backup, obj *svcsdk.CreateBackupInput) *svcsdk.CreateBackupInput {
	obj.TableName = aws.String(cr.Spec.ForProvider.TableName)
	return obj
}

func preGenerateDeleteBackupInput(_ *svcapitypes.Backup, obj *svcsdk.DeleteBackupInput) *svcsdk.DeleteBackupInput {
	return obj
}

func postGenerateDeleteBackupInput(cr *svcapitypes.Backup, obj *svcsdk.DeleteBackupInput) *svcsdk.DeleteBackupInput {
	obj.BackupArn = aws.String(meta.GetExternalName(cr))
	return obj
}
