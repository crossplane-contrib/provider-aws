/*
Copyright 2019 The Crossplane Authors.

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

package rdsinstance

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	rds "github.com/crossplane-contrib/provider-aws/pkg/clients/database"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotRDSInstance                     = "managed resource is not an RDS instance custom resource"
	errKubeUpdateFailed                   = "cannot update RDS instance custom resource"
	errCreateFailed                       = "cannot create RDS instance"
	errS3RestoreFailed                    = "cannot restore RDS instance from S3 backup"
	errSnapshotRestoreFailed              = "cannot restore RDS instance from snapshot"
	errPointInTimeRestoreFailed           = "cannot restore RDS instance from point in time"
	errPointInTimeRestoreSourceNotDefined = "sourceDBInstanceAutomatedBackupsArn, sourceDBInstanceIdentifier or sourceDbiResourceId must be defined"
	errUnknownRestoreSource               = "unknown RDS restore source"
	errModifyFailed                       = "cannot modify RDS instance"
	errAddTagsFailed                      = "cannot add tags to RDS instance"
	errRemoveTagsFailed                   = "cannot remove tags from  RDS instance"
	errDeleteFailed                       = "cannot delete RDS instance"
	errDescribeFailed                     = "cannot describe RDS instance"
	errPatchCreationFailed                = "cannot create a patch object"
	errUpToDateFailed                     = "cannot check whether object is up-to-date"
	errGetPasswordSecretFailed            = "cannot get password secret"
)

// SetupRDSInstance adds a controller that reconciles RDSInstances.
func SetupRDSInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RDSInstanceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: rds.NewClient}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.RDSInstance{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config *aws.Config) rds.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(cfg), c.kube, Cache{}}, nil
}

type Cache struct {
	AddTags    []rdstypes.Tag
	RemoveTags []string
}

type external struct {
	client rds.Client
	kube   client.Client

	cache Cache
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}

	// TODO(muvaf): There are some parameters that require a specific call
	// for retrieval. For example, DescribeDBInstancesOutput does not expose
	// the tags map of the RDS instance, you have to make ListTagsForResourceRequest
	rsp, err := e.client.DescribeDBInstances(ctx, &awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	instance := rsp.DBInstances[0]
	current := cr.Spec.ForProvider.DeepCopy()
	rds.LateInitialize(&cr.Spec.ForProvider, &instance)
	cr.Status.AtProvider = rds.GenerateObservation(instance)

	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1beta1.RDSInstanceStateAvailable, v1beta1.RDSInstanceStateModifying, v1beta1.RDSInstanceStateBackingUp, v1beta1.RDSInstanceStateConfiguringEnhancedMonitoring, v1beta1.RDSInstanceStateStorageOptimization:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.RDSInstanceStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.RDSInstanceStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	var upToDate bool
	var diff string

	upToDate, diff, e.cache.AddTags, e.cache.RemoveTags, err = rds.IsUpToDate(ctx, e.kube, cr, instance)

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !reflect.DeepEqual(current, &cr.Spec.ForProvider),
		ConnectionDetails:       rds.GetConnectionDetails(*cr),
		Diff:                    diff,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRDSInstance)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateCreating {
		return managed.ExternalCreation{}, nil
	}
	pw, _, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if pw == "" {
		pw, err = password.Generate()
		if err != nil {
			return managed.ExternalCreation{}, err
		}
	}

	err = e.RestoreOrCreate(ctx, cr, pw)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
	}
	if cr.Spec.ForProvider.MasterUsername != nil {
		conn[xpv1.ResourceCredentialsSecretUserKey] = []byte(aws.ToString(cr.Spec.ForProvider.MasterUsername))
	}
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) RestoreOrCreate(ctx context.Context, cr *v1beta1.RDSInstance, pw string) error { //nolint:gocyclo
	if cr.Spec.ForProvider.RestoreFrom == nil {
		_, err := e.client.CreateDBInstance(ctx, rds.GenerateCreateRDSInstanceInput(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
		if err != nil {
			return errorutils.Wrap(err, errCreateFailed)
		}
		return nil
	}

	switch *cr.Spec.ForProvider.RestoreFrom.Source {
	case "S3":
		_, err := e.client.RestoreDBInstanceFromS3(ctx, rds.GenerateRestoreRDSInstanceFromS3Input(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
		if err != nil {
			return errorutils.Wrap(err, errS3RestoreFailed)
		}
	case "Snapshot":
		_, err := e.client.RestoreDBInstanceFromDBSnapshot(ctx, rds.GenerateRestoreRDSInstanceFromSnapshotInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
		if err != nil {
			return errorutils.Wrap(err, errSnapshotRestoreFailed)
		}
	case "PointInTime":
		if cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDBInstanceIdentifier == nil && cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDbiResourceID == nil && cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDBInstanceAutomatedBackupsArn == nil {
			return errors.New(errPointInTimeRestoreSourceNotDefined)
		}
		_, err := e.client.RestoreDBInstanceToPointInTime(ctx, rds.GenerateRestoreRDSInstanceToPointInTimeInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
		if err != nil {
			return errorutils.Wrap(err, errPointInTimeRestoreFailed)
		}
	default:
		return errors.New(errUnknownRestoreSource)
	}
	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRDSInstance)
	}
	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1beta1.RDSInstanceStateModifying, v1beta1.RDSInstanceStateCreating:
		return managed.ExternalUpdate{}, nil
	}
	// AWS rejects modification requests if you send fields whose value is same
	// as the current one. So, we have to create a patch out of the desired state
	// and the current state. Since the DBInstance is not fully mirrored in status,
	// we lose the current state after a change is made to spec, which forces us
	// to make a DescribeDBInstancesRequest to get the current state.
	rsp, err := e.client.DescribeDBInstances(ctx, &awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errDescribeFailed)
	}
	patch, err := rds.CreatePatch(&rsp.DBInstances[0], &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errPatchCreationFailed)
	}
	modify := rds.GenerateModifyDBInstanceInput(meta.GetExternalName(cr), patch, &rsp.DBInstances[0])
	var conn managed.ConnectionDetails

	pwd, changed, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	if changed {
		conn = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretPasswordKey: []byte(pwd),
		}
		modify.MasterUserPassword = aws.String(pwd)
	}

	if _, err = e.client.ModifyDBInstance(ctx, modify); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyFailed)
	}

	// Update tags if necessary
	if len(e.cache.RemoveTags) > 0 {
		_, err := e.client.RemoveTagsFromResource(ctx, &awsrds.RemoveTagsFromResourceInput{
			ResourceName: aws.String(cr.Status.AtProvider.DBInstanceArn),
			TagKeys:      e.cache.RemoveTags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errRemoveTagsFailed)
		}
	}
	// remove before add for case where we just simply update a tag value
	if len(e.cache.AddTags) > 0 {
		_, err := e.client.AddTagsToResource(ctx, &awsrds.AddTagsToResourceInput{
			ResourceName: aws.String(cr.Status.AtProvider.DBInstanceArn),
			Tags:         e.cache.AddTags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{ConnectionDetails: conn}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateDeleting {
		return nil
	}
	// TODO(muvaf): There are cases where deletion results in an error that can
	// be solved only by a config change. But to do that, reconciler has to call
	// Update before Delete, which is not the case currently. In RDS, deletion
	// protection is an example for that and it's pretty common to use it. So,
	// until managed reconciler does Update before Delete, we do it here manually.
	// Update here is a best effort and deletion should not stop if it fails since
	// user may want to delete a resource whose fields are causing error.
	_, err := e.Update(ctx, cr)
	if rds.IsErrorNotFound(err) {
		return nil
	}
	input := awsrds.DeleteDBInstanceInput{
		DBInstanceIdentifier:      aws.String(meta.GetExternalName(cr)),
		DeleteAutomatedBackups:    cr.Spec.ForProvider.DeleteAutomatedBackups,
		SkipFinalSnapshot:         aws.ToBool(cr.Spec.ForProvider.SkipFinalSnapshotBeforeDeletion),
		FinalDBSnapshotIdentifier: cr.Spec.ForProvider.FinalDBSnapshotIdentifier,
	}
	_, err = e.client.DeleteDBInstance(ctx, &input)
	return errorutils.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}
