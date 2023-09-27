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
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsrdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	xperrors "github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	corev1 "k8s.io/api/core/v1"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	rds "github.com/crossplane-contrib/provider-aws/pkg/clients/database"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
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
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: rds.NewClient}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
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
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(cfg), c.kube}, nil
}

type external struct {
	client rds.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}

	// TODO(muvaf): There are some parameters that require a specific call
	// for retrieval. For example, DescribeDBInstancesOutput does not expose
	// the tags map of the RDS instance, you have to make ListTagsForResourceRequest
	rsp, err := e.client.DescribeDBInstances(ctx, &awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
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
	upToDate, diff, err := rds.IsUpToDate(ctx, e.kube, cr, instance)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errUpToDateFailed)
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
	cr.Status.SetConditions(xpv1.Creating(), rds.PasswordSetPending("Creating"))
	if err := e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
		return managed.ExternalCreation{}, err
	}
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

func (e *external) RestoreOrCreate(ctx context.Context, cr *v1beta1.RDSInstance, pw string) error { // nolint:gocyclo
	if cr.Spec.ForProvider.RestoreFrom == nil {
		_, err := e.client.CreateDBInstance(ctx, rds.GenerateCreateRDSInstanceInput(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
		if err != nil {
			return awsclient.Wrap(err, errCreateFailed)
		}
		cr.Status.SetConditions(rds.PasswordSet("Set by CreateDBInstance"))
		if err = e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
			return err
		}
		return nil
	}

	switch *cr.Spec.ForProvider.RestoreFrom.Source {
	case "S3":
		_, err := e.client.RestoreDBInstanceFromS3(ctx, rds.GenerateRestoreRDSInstanceFromS3Input(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
		if err != nil {
			return awsclient.Wrap(err, errS3RestoreFailed)
		}
	case "Snapshot":
		_, err := e.client.RestoreDBInstanceFromDBSnapshot(ctx, rds.GenerateRestoreRDSInstanceFromSnapshotInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
		if err != nil {
			return awsclient.Wrap(err, errSnapshotRestoreFailed)
		}
	case "PointInTime":
		if cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDBInstanceIdentifier == nil && cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDbiResourceID == nil && cr.Spec.ForProvider.RestoreFrom.PointInTime.SourceDBInstanceAutomatedBackupsArn == nil {
			return errors.New(errPointInTimeRestoreSourceNotDefined)
		}
		_, err := e.client.RestoreDBInstanceToPointInTime(ctx, rds.GenerateRestoreRDSInstanceToPointInTimeInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
		if err != nil {
			return awsclient.Wrap(err, errPointInTimeRestoreFailed)
		}
	default:
		return errors.New(errUnknownRestoreSource)
	}

	_, ret := e.client.ModifyDBInstance(ctx, &awsrds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(meta.GetExternalName(cr)),
		MasterUserPassword:   aws.String(pw),
	})

	if ret == nil {
		cr.Status.SetConditions(rds.PasswordSet("Password set via ModifyDBInstance in Create"))
	} else {
		cr.Status.SetConditions(rds.PasswordSetFail(ret))
	}

	if err := e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
		return xperrors.Join(ret, err)
	}

	return ret
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
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
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errDescribeFailed)
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

	// In this additional check we test for != ConditionFalse because we don't want to disturb RDS
	// steady state RDS created before this fix for https://github.com/crossplane-contrib/provider-aws/issues/1121
	// was added.
	// False will only be set on those created after the fix went in or that had a password change attempt after
	// the fix went in. ( And somehow the password setting didn't work of course )
	if cr.Status.GetCondition(rds.CndtnPaswordSet).Status == corev1.ConditionFalse {
		changed = true
	}
	if changed {
		cr.Status.SetConditions(rds.PasswordSetPending("Updating"))
		if err = e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
			return managed.ExternalUpdate{}, err
		}
		// In case of restore from snapshot we still might not have a set password when we get here.
		// So like in CreateOrRestore we look for "" and generate a password if we see it.
		if pwd == "" {
			pwd, err = password.Generate()
			if err != nil {
				return managed.ExternalUpdate{}, err
			}
		}
	}

	if changed {
		conn = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretPasswordKey: []byte(pwd),
			// If we're here after after restore from Snapshot then user name
			// might not be set.
			xpv1.ResourceCredentialsSecretUserKey: []byte(aws.ToString(cr.Spec.ForProvider.MasterUsername)),
		}
		modify.MasterUserPassword = aws.String(pwd)
	}

	if _, err = e.client.ModifyDBInstance(ctx, modify); err != nil {
		if changed {
			cr.Status.SetConditions(rds.PasswordSetFail(err))
			if errSts := e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
				return managed.ExternalUpdate{}, xperrors.Join(err, errSts)
			}
		}
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyFailed)
	}
	if changed {
		cr.Status.SetConditions(rds.PasswordSet("Password set via ModifyDBInstance in Update"))
		if err = e.kube.Status().Update(ctx, cr); err != nil { // Why can't we just let controller-runtime take care of this?
			return managed.ExternalUpdate{}, err
		}
	}
	if len(patch.Tags) > 0 {
		tags := make([]awsrdstypes.Tag, len(patch.Tags))
		for i, t := range patch.Tags {
			tags[i] = awsrdstypes.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}
		_, err = e.client.AddTagsToResource(ctx, &awsrds.AddTagsToResourceInput{
			ResourceName: aws.String(cr.Status.AtProvider.DBInstanceArn),
			Tags:         tags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errAddTagsFailed)
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
	return awsclient.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mg) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]v1beta1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = v1beta1.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
