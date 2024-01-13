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

package dbcluster

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotDBCluster             = "managed resource is not a DB Cluster custom resource"
	errKubeUpdateFailed         = "cannot update DBCluster instance custom resource"
	errGetPasswordSecretFailed  = "cannot get password secret"
	errSaveSecretFailed         = "failed to save generated password to Kubernetes secret"
	errRestore                  = "cannot restore DBCluster in AWS"
	errUnknownRestoreFromSource = "unknown restoreFrom source"
)

// SetupDBCluster adds a controller that reconciles a DBCluster.
func SetupDBCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBClusterKind)
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
		resource.ManagedKind(svcapitypes.DBClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBCluster{}).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(r)
}

func setupExternal(e *external) {
	h := &hooks{client: e.client, kube: e.kube}
	e.preObserve = preObserve
	e.postObserve = h.postObserve
	e.isUpToDate = h.isUpToDate
	e.preUpdate = h.preUpdate
	e.postUpdate = h.postUpdate
	e.preCreate = h.preCreate
	e.postCreate = h.postCreate
	e.preDelete = preDelete
	e.filterList = filterList
	e.lateInitialize = lateInitialize
}

type hooks struct {
	client docdbiface.DocDBAPI
	kube   client.Client
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (e *hooks) postObserve(ctx context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	obs.ConnectionDetails = getConnectionDetails(cr)

	if !meta.WasDeleted(cr) {
		pw, _, _ := e.getPasswordFromRef(ctx, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
		if pw != "" {
			obs.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(pw)
		}
	}

	switch pointer.StringValue(cr.Status.AtProvider.Status) {
	case svcapitypes.DocDBInstanceStateAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case svcapitypes.DocDBInstanceStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case svcapitypes.DocDBInstanceStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func lateInitialize(cr *svcapitypes.DBClusterParameters, resp *svcsdk.DescribeDBClustersOutput) error {
	cluster := resp.DBClusters[0]

	if cr.AvailabilityZones == nil {
		cr.AvailabilityZones = cluster.AvailabilityZones
	}

	cr.BackupRetentionPeriod = pointer.LateInitialize(cr.BackupRetentionPeriod, cluster.BackupRetentionPeriod)
	cr.DBClusterParameterGroupName = pointer.LateInitialize(cr.DBClusterParameterGroupName, cluster.DBClusterParameterGroup)
	cr.DBSubnetGroupName = pointer.LateInitialize(cr.DBSubnetGroupName, cluster.DBSubnetGroup)
	cr.DeletionProtection = pointer.LateInitialize(cr.DeletionProtection, cluster.DeletionProtection)
	cr.EngineVersion = pointer.LateInitialize(cr.EngineVersion, cluster.EngineVersion)
	cr.KMSKeyID = pointer.LateInitialize(cr.KMSKeyID, cluster.KmsKeyId)
	cr.Port = pointer.LateInitialize(cr.Port, cluster.Port)
	cr.PreferredBackupWindow = pointer.LateInitialize(cr.PreferredBackupWindow, cluster.PreferredBackupWindow)
	cr.PreferredMaintenanceWindow = pointer.LateInitialize(cr.PreferredMaintenanceWindow, cluster.PreferredMaintenanceWindow)
	cr.StorageEncrypted = pointer.LateInitialize(cr.StorageEncrypted, cluster.StorageEncrypted)

	if cr.EnableCloudwatchLogsExports == nil {
		cr.EnableCloudwatchLogsExports = cluster.EnabledCloudwatchLogsExports
	}
	if cr.VPCSecurityGroupIDs == nil {
		cr.VPCSecurityGroupIDs = make([]*string, len(cluster.VpcSecurityGroups))
		for i, group := range cluster.VpcSecurityGroups {
			cr.VPCSecurityGroupIDs[i] = group.VpcSecurityGroupId
		}
	}

	return nil
}

func (e *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput) (bool, string, error) {
	cluster := resp.DBClusters[0]

	_, pwChanged, err := e.getPasswordFromRef(ctx, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil || pwChanged {
		return false, "", err
	}

	switch {
	case pointer.Int64Value(cr.Spec.ForProvider.BackupRetentionPeriod) != pointer.Int64Value(cluster.BackupRetentionPeriod),
		pointer.StringValue(cr.Spec.ForProvider.DBClusterParameterGroupName) != pointer.StringValue(cluster.DBClusterParameterGroup),
		pointer.BoolValue(cr.Spec.ForProvider.DeletionProtection) != pointer.BoolValue(cluster.DeletionProtection),
		!areSameElements(cr.Spec.ForProvider.EnableCloudwatchLogsExports, cluster.EnabledCloudwatchLogsExports),
		pointer.Int64Value(cr.Spec.ForProvider.Port) != pointer.Int64Value(cluster.Port),
		pointer.StringValue(cr.Spec.ForProvider.PreferredBackupWindow) != pointer.StringValue(cluster.PreferredBackupWindow),
		pointer.StringValue(cr.Spec.ForProvider.PreferredMaintenanceWindow) != pointer.StringValue(cluster.PreferredMaintenanceWindow):
		return false, "", nil
	}

	areTagsUpToDate, err := svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, cluster.DBClusterArn)
	return areTagsUpToDate, "", err
}

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterInput) error {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.CloudwatchLogsExportConfiguration = generateCloudWatchExportConfiguration(
		cr.Spec.ForProvider.EnableCloudwatchLogsExports,
		cr.Status.AtProvider.EnabledCloudwatchLogsExports)
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately

	pw, pwchanged, err := e.getPasswordFromRef(ctx, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return err
	}
	if pwchanged {
		obj.MasterUserPassword = aws.String(pw)
	}
	return nil
}

func (e *hooks) postUpdate(_ context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.ModifyDBClusterOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	return upd, svcutils.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, resp.DBCluster.DBClusterArn)
}

func (e *hooks) preCreate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) error { //nolint:gocyclo
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	pw, _, err := e.getPasswordFromRef(ctx, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil {
		return err
	}
	if pw == "" && aws.BoolValue(&cr.Spec.ForProvider.AutogeneratePassword) {
		pw, err = password.Generate()
		if err != nil {
			return errors.Wrap(err, "unable to generate a password")
		}
		if err := e.savePasswordSecret(ctx, cr, pw); err != nil {
			return errors.Wrap(err, errSaveSecretFailed)
		}
	}

	obj.MasterUserPassword = pointer.ToOrNilIfZeroValue(pw)
	if cr.Spec.ForProvider.RestoreFrom != nil {
		switch cr.Spec.ForProvider.RestoreFrom.Source {
		case svcapitypes.RestoreSourceSnapshot:
			input := generateRestoreDBClusterFromSnapshotInput(cr)
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = e.client.RestoreDBClusterFromSnapshotWithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		case svcapitypes.RestoreSourcePointInTime:
			input := generateRestoreDBClusterToPointInTimeInput(cr)
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = e.client.RestoreDBClusterToPointInTimeWithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		default:
			return errors.New(errUnknownRestoreFromSource)
		}
	}
	return nil
}

func (e *hooks) postCreate(ctx context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.CreateDBClusterOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	cre.ConnectionDetails = getConnectionDetails(cr)
	pw, _, err := e.getPasswordFromRef(ctx, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if pw != "" {
		cre.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(pw)
	}

	cre.ConnectionDetails[xpv1.ResourceCredentialsSecretUserKey] = []byte(pointer.StringValue(cr.Spec.ForProvider.MasterUsername))
	// Tags are added during update
	return cre, nil
}

func generateRestoreDBClusterFromSnapshotInput(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterFromSnapshotInput { //nolint:gocyclo
	res := &svcsdk.RestoreDBClusterFromSnapshotInput{}

	if cr.Spec.ForProvider.AvailabilityZones != nil {
		res.SetAvailabilityZones(cr.Spec.ForProvider.AvailabilityZones)
	}

	if cr.Spec.ForProvider.DBSubnetGroupName != nil {
		res.SetDBSubnetGroupName(*cr.Spec.ForProvider.DBSubnetGroupName)
	}

	if cr.Spec.ForProvider.DeletionProtection != nil {
		res.SetDeletionProtection(*cr.Spec.ForProvider.DeletionProtection)
	}

	if cr.Spec.ForProvider.EnableCloudwatchLogsExports != nil {
		res.SetEnableCloudwatchLogsExports(cr.Spec.ForProvider.EnableCloudwatchLogsExports)
	}

	if cr.Spec.ForProvider.Engine != nil {
		res.SetEngine(*cr.Spec.ForProvider.Engine)
	}

	if cr.Spec.ForProvider.EngineVersion != nil {
		res.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}

	if cr.Spec.ForProvider.KMSKeyID != nil {
		res.SetKmsKeyId(*cr.Spec.ForProvider.KMSKeyID)
	}

	if cr.Spec.ForProvider.Port != nil {
		res.SetPort(*cr.Spec.ForProvider.Port)
	}

	if cr.Spec.ForProvider.RestoreFrom != nil && cr.Spec.ForProvider.RestoreFrom.Snapshot != nil {
		res.SetSnapshotIdentifier(cr.Spec.ForProvider.RestoreFrom.Snapshot.SnapshotIdentifier)
	}

	if cr.Spec.ForProvider.Tags != nil {
		var tags []*svcsdk.Tag
		for _, tag := range cr.Spec.ForProvider.Tags {
			tags = append(tags, &svcsdk.Tag{Key: tag.Key, Value: tag.Value})
		}

		res.SetTags(tags)
	}

	return res
}

func generateRestoreDBClusterToPointInTimeInput(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterToPointInTimeInput {
	p := cr.Spec.ForProvider
	res := &svcsdk.RestoreDBClusterToPointInTimeInput{
		DBSubnetGroupName:           p.DBSubnetGroupName,
		DeletionProtection:          p.DeletionProtection,
		EnableCloudwatchLogsExports: p.EnableCloudwatchLogsExports,
		KmsKeyId:                    p.KMSKeyID,
		Port:                        p.Port,
		UseLatestRestorableTime:     p.RestoreFrom.PointInTime.UseLatestRestorableTime,
		VpcSecurityGroupIds:         p.VPCSecurityGroupIDs,
	}
	if p.RestoreFrom != nil {
		if p.RestoreFrom.PointInTime != nil {
			if p.RestoreFrom.PointInTime.RestoreTime != nil {
				res.SetRestoreToTime(p.RestoreFrom.PointInTime.RestoreTime.Time)
			}
			res.RestoreType = p.RestoreFrom.PointInTime.RestoreType
			res.SourceDBClusterIdentifier = &p.RestoreFrom.PointInTime.SourceDBClusterIdentifier
		}
	}
	if cr.Spec.ForProvider.Tags != nil {
		var tags []*svcsdk.Tag
		for _, tag := range cr.Spec.ForProvider.Tags {
			tags = append(tags, &svcsdk.Tag{Key: tag.Key, Value: tag.Value})
		}

		res.SetTags(tags)
	}

	return res
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) (bool, error) {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.FinalDBSnapshotIdentifier = cr.Spec.ForProvider.FinalDBSnapshotIdentifier
	obj.SkipFinalSnapshot = cr.Spec.ForProvider.SkipFinalSnapshot
	return false, nil
}

func filterList(cr *svcapitypes.DBCluster, list *svcsdk.DescribeDBClustersOutput) *svcsdk.DescribeDBClustersOutput {
	id := meta.GetExternalName(cr)
	for _, instance := range list.DBClusters {
		if pointer.StringValue(instance.DBClusterIdentifier) == id {
			return &svcsdk.DescribeDBClustersOutput{
				Marker:     list.Marker,
				DBClusters: []*svcsdk.DBCluster{instance},
			}
		}
	}

	return &svcsdk.DescribeDBClustersOutput{
		Marker:     list.Marker,
		DBClusters: []*svcsdk.DBCluster{},
	}
}

func generateCloudWatchExportConfiguration(spec, current []*string) *svcsdk.CloudwatchLogsExportConfiguration {
	toEnable := []*string{}
	toDisable := []*string{}

	currentMap := make(map[string]struct{}, len(current))
	for _, currentID := range current {
		currentMap[pointer.StringValue(currentID)] = struct{}{}
	}

	specMap := make(map[string]struct{}, len(spec))
	for _, specID := range spec {
		key := pointer.StringValue(specID)
		specMap[key] = struct{}{}

		if _, exists := currentMap[key]; !exists {
			toEnable = append(toEnable, specID)
		}
	}

	for _, currentID := range current {
		if _, exists := specMap[pointer.StringValue(currentID)]; !exists {
			toDisable = append(toDisable, currentID)
		}
	}

	return &svcsdk.CloudwatchLogsExportConfiguration{
		EnableLogTypes:  toEnable,
		DisableLogTypes: toDisable,
	}
}

func areSameElements(a1, a2 []*string) bool {
	if len(a1) != len(a2) {
		return false
	}

	m2 := make(map[string]struct{}, len(a2))
	for _, s2 := range a2 {
		m2[pointer.StringValue(s2)] = struct{}{}
	}

	for _, s1 := range a1 {
		v1 := pointer.StringValue(s1)
		if _, exists := m2[v1]; !exists {
			return false
		}
	}

	return true
}

func (e *hooks) getPasswordFromRef(ctx context.Context, in *xpv1.SecretKeySelector, out *xpv1.SecretReference) (newPwd string, changed bool, err error) {
	if in == nil {
		return "", false, nil
	}
	nn := types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
	s := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, s); err != nil {
		return "", false, errors.Wrap(err, errGetPasswordSecretFailed)
	}
	newPwd = string(s.Data[in.Key])

	if out != nil {
		nn = types.NamespacedName{
			Name:      out.Name,
			Namespace: out.Namespace,
		}
		s = &corev1.Secret{}
		// the output secret may not exist yet, so we can skip returning an
		// error if the error is NotFound
		if err := e.kube.Get(ctx, nn, s); resource.IgnoreNotFound(err) != nil {
			return "", false, err
		}
		// if newPwd was set to some value, compare value in output secret with
		// newPwd
		changed = newPwd != "" && newPwd != string(s.Data[xpv1.ResourceCredentialsSecretPasswordKey])
	}
	return newPwd, changed, nil
}

func getConnectionDetails(cr *svcapitypes.DBCluster) managed.ConnectionDetails {
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey:     []byte(pointer.StringValue(cr.Spec.ForProvider.MasterUsername)),
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(pointer.StringValue(cr.Status.AtProvider.Endpoint)),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(pointer.Int64Value(cr.Spec.ForProvider.Port)))),
		"readerEndpoint":                          []byte(pointer.StringValue(cr.Status.AtProvider.ReaderEndpoint)),
	}
}

func (e *hooks) savePasswordSecret(ctx context.Context, cr *svcapitypes.DBCluster, pw string) error {
	if cr.Spec.ForProvider.MasterUserPasswordSecretRef == nil {
		return errors.New("no MasterUserPasswordSecretRef given, unable to store password")
	}
	patcher := resource.NewAPIPatchingApplicator(e.kube)
	ref := cr.Spec.ForProvider.MasterUserPasswordSecretRef
	sc := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
		Data: map[string][]byte{
			ref.Key: []byte(pw),
		},
	}
	return patcher.Apply(ctx, sc)
}
