package dbcluster

import (
	"context"
	"strconv"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"

	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"

	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/features"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

// error constants
const (
	errSaveSecretFailed         = "failed to save generated password to Kubernetes secret"
	errUpdateTags               = "cannot update tags"
	errRestore                  = "cannot restore DBCluster in AWS"
	errUnknownRestoreFromSource = "unknown restoreFrom source"
)

type updater struct {
	client svcsdkapi.RDSAPI
}

// SetupDBCluster adds a controller that reconciles DbCluster.
func SetupDBCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = c.postObserve
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
			u := &updater{client: e.client}
			e.postUpdate = u.postUpdate
			e.preCreate = c.preCreate
			e.preDelete = preDelete
			e.filterList = filterList
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.DBCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

// This probably requires custom Conditions to be defined for handling all statuses
// described here https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Status.html
// Need to get help from community on how to deal with this. Ideally the status should reflect
// the true status value as described by the provider.
func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.DBClusters[0].Status) {
	case "available", "modifying":
		cr.SetConditions(xpv1.Available())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(cr.Status.AtProvider.Endpoint)),
		xpv1.ResourceCredentialsSecretUserKey:     []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername)),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.FormatInt(aws.Int64Value(cr.Spec.ForProvider.Port), 10)),
	}
	pw, _, _ := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if pw != "" {
		obs.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(pw)
	}

	return obs, nil
}

type custom struct {
	kube   client.Client
	client svcsdkapi.RDSAPI
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) error { // nolint:gocyclo
	pw, _, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "cannot get password from the given secret")
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

	obj.MasterUserPassword = aws.String(pw)
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
	for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
		obj.VpcSecurityGroupIds[i] = aws.String(v)
	}

	if cr.Spec.ForProvider.RestoreFrom != nil {
		switch *cr.Spec.ForProvider.RestoreFrom.Source {
		case "S3":
			input := generateRestoreDBClusterFromS3Input(cr)
			input.MasterUserPassword = obj.MasterUserPassword
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = e.client.RestoreDBClusterFromS3WithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		default:
			return errors.New(errUnknownRestoreFromSource)
		}
	}

	return nil
}

func generateRestoreDBClusterFromS3Input(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterFromS3Input { // nolint:gocyclo
	res := &svcsdk.RestoreDBClusterFromS3Input{}

	if cr.Spec.ForProvider.AvailabilityZones != nil {
		res.SetAvailabilityZones(cr.Spec.ForProvider.AvailabilityZones)
	}

	if cr.Spec.ForProvider.BacktrackWindow != nil {
		res.SetBacktrackWindow(*cr.Spec.ForProvider.BacktrackWindow)
	}

	if cr.Spec.ForProvider.BackupRetentionPeriod != nil {
		res.SetBackupRetentionPeriod(*cr.Spec.ForProvider.BackupRetentionPeriod)
	}

	if cr.Spec.ForProvider.CharacterSetName != nil {
		res.SetCharacterSetName(*cr.Spec.ForProvider.CharacterSetName)
	}

	if cr.Spec.ForProvider.CopyTagsToSnapshot != nil {
		res.SetCopyTagsToSnapshot(*cr.Spec.ForProvider.CopyTagsToSnapshot)
	}

	if cr.Spec.ForProvider.DBClusterParameterGroupName != nil {
		res.SetDBClusterParameterGroupName(*cr.Spec.ForProvider.DBClusterParameterGroupName)
	}

	if cr.Spec.ForProvider.DBSubnetGroupName != nil {
		res.SetDBSubnetGroupName(*cr.Spec.ForProvider.DBSubnetGroupName)
	}

	if cr.Spec.ForProvider.DatabaseName != nil {
		res.SetDatabaseName(*cr.Spec.ForProvider.DatabaseName)
	}

	if cr.Spec.ForProvider.DeletionProtection != nil {
		res.SetDeletionProtection(*cr.Spec.ForProvider.DeletionProtection)
	}

	if cr.Spec.ForProvider.Domain != nil {
		res.SetDomain(*cr.Spec.ForProvider.Domain)
	}

	if cr.Spec.ForProvider.DomainIAMRoleName != nil {
		res.SetDomainIAMRoleName(*cr.Spec.ForProvider.DomainIAMRoleName)
	}

	if cr.Spec.ForProvider.EnableCloudwatchLogsExports != nil {
		res.SetEnableCloudwatchLogsExports(cr.Spec.ForProvider.EnableCloudwatchLogsExports)
	}

	if cr.Spec.ForProvider.EnableIAMDatabaseAuthentication != nil {
		res.SetEnableIAMDatabaseAuthentication(*cr.Spec.ForProvider.EnableIAMDatabaseAuthentication)
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

	if cr.Spec.ForProvider.MasterUsername != nil {
		res.SetMasterUsername(*cr.Spec.ForProvider.MasterUsername)
	}

	if cr.Spec.ForProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Spec.ForProvider.OptionGroupName)
	}

	if cr.Spec.ForProvider.Port != nil {
		res.SetPort(*cr.Spec.ForProvider.Port)
	}

	if cr.Spec.ForProvider.PreferredBackupWindow != nil {
		res.SetPreferredBackupWindow(*cr.Spec.ForProvider.PreferredBackupWindow)
	}

	if cr.Spec.ForProvider.PreferredMaintenanceWindow != nil {
		res.SetPreferredMaintenanceWindow(*cr.Spec.ForProvider.PreferredMaintenanceWindow)
	}

	if cr.Spec.ForProvider.StorageEncrypted != nil {
		res.SetStorageEncrypted(*cr.Spec.ForProvider.StorageEncrypted)
	}

	if cr.Spec.ForProvider.RestoreFrom != nil && cr.Spec.ForProvider.RestoreFrom.S3 != nil {
		if cr.Spec.ForProvider.RestoreFrom.S3.BucketName != nil {
			res.SetS3BucketName(*cr.Spec.ForProvider.RestoreFrom.S3.BucketName)
		}

		if cr.Spec.ForProvider.RestoreFrom.S3.IngestionRoleARN != nil {
			res.SetS3IngestionRoleArn(*cr.Spec.ForProvider.RestoreFrom.S3.IngestionRoleARN)
		}

		if cr.Spec.ForProvider.RestoreFrom.S3.Prefix != nil {
			res.SetS3Prefix(*cr.Spec.ForProvider.RestoreFrom.S3.Prefix)
		}

		if cr.Spec.ForProvider.RestoreFrom.S3.SourceEngine != nil {
			res.SetSourceEngine(*cr.Spec.ForProvider.RestoreFrom.S3.SourceEngine)
		}

		if cr.Spec.ForProvider.RestoreFrom.S3.SourceEngineVersion != nil {
			res.SetSourceEngineVersion(*cr.Spec.ForProvider.RestoreFrom.S3.SourceEngineVersion)
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

func isUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) (bool, error) { // nolint:gocyclo
	status := aws.StringValue(out.DBClusters[0].Status)
	if status == "modifying" || status == "upgrading" || status == "configuring-iam-database-auth" || status == "migrating" || status == "prepairing-data-migration" {
		return true, nil
	}

	if aws.BoolValue(cr.Spec.ForProvider.EnableIAMDatabaseAuthentication) != aws.BoolValue(out.DBClusters[0].IAMDatabaseAuthenticationEnabled) {
		return false, nil
	}

	if !isPreferredMaintenanceWindowUpToDate(cr, out) {
		return false, nil
	}

	if !isPreferredBackupWindowUpToDate(cr, out) {
		return false, nil
	}

	if aws.Int64Value(cr.Spec.ForProvider.BacktrackWindow) != aws.Int64Value(out.DBClusters[0].BacktrackWindow) {
		return false, nil
	}

	if !isBackupRetentionPeriodUpToDate(cr, out) {
		return false, nil
	}

	if aws.BoolValue(cr.Spec.ForProvider.CopyTagsToSnapshot) != aws.BoolValue(out.DBClusters[0].CopyTagsToSnapshot) {
		return false, nil
	}

	if aws.BoolValue(cr.Spec.ForProvider.DeletionProtection) != aws.BoolValue(out.DBClusters[0].DeletionProtection) {
		return false, nil
	}

	if !isEngineVersionUpToDate(cr, out) {
		return false, nil
	}

	if !isPortUpToDate(cr, out) {
		return false, nil
	}

	isScalingConfigurationUpToDate, err := isScalingConfigurationUpToDate(cr.Spec.ForProvider.ScalingConfiguration, out.DBClusters[0].ScalingConfigurationInfo)
	if !isScalingConfigurationUpToDate {
		return false, err
	}

	add, remove := DiffTags(cr.Spec.ForProvider.Tags, out.DBClusters[0].TagList)
	if len(add) > 0 || len(remove) > 0 {
		return false, nil
	}
	return true, nil
}

func isPreferredMaintenanceWindowUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If PreferredMaintenanceWindow is not set, aws sets a random window
	// so we do not try to update in this case
	if cr.Spec.ForProvider.PreferredMaintenanceWindow != nil {

		// AWS accepts uppercase weekdays, but returns lowercase values,
		// therfore we compare usinf equalFold
		if !strings.EqualFold(aws.StringValue(cr.Spec.ForProvider.PreferredMaintenanceWindow), aws.StringValue(out.DBClusters[0].PreferredMaintenanceWindow)) {
			return false
		}
	}
	return true
}

func isPreferredBackupWindowUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If PreferredBackupWindow is not set, aws sets a random window
	// so we do not try to update in this case
	if cr.Spec.ForProvider.PreferredBackupWindow != nil {
		if aws.StringValue(cr.Spec.ForProvider.PreferredBackupWindow) != aws.StringValue(out.DBClusters[0].PreferredBackupWindow) {
			return false
		}
	}
	return true
}

func isBackupRetentionPeriodUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If BackupRetentionPeriod is not set, aws sets a default value
	// so we do not try to update in this case
	if cr.Spec.ForProvider.BackupRetentionPeriod != nil {
		if aws.Int64Value(cr.Spec.ForProvider.BackupRetentionPeriod) != aws.Int64Value(out.DBClusters[0].BackupRetentionPeriod) {
			return false
		}
	}
	return true
}

func isScalingConfigurationUpToDate(sc *svcapitypes.ScalingConfiguration, obj *svcsdk.ScalingConfigurationInfo) (bool, error) {
	jsonPatch, err := aws.CreateJSONPatch(sc, obj)
	if err != nil {
		return false, err
	}
	// if there is no difference, jsonPatch is {}
	if len(jsonPatch) > 2 {
		return false, nil
	}
	return true, nil
}

func isEngineVersionUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If EngineVersion is not set, aws sets a default value
	// so we do not try to update in this case
	if cr.Spec.ForProvider.EngineVersion != nil {
		if aws.StringValue(cr.Spec.ForProvider.EngineVersion) != aws.StringValue(out.DBClusters[0].EngineVersion) {
			return false
		}
	}
	return true
}

func isPortUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If Port is not set, aws sets a default value
	// so we do not try to update in this case
	if cr.Spec.ForProvider.Port != nil {
		if aws.Int64Value(cr.Spec.ForProvider.Port) != aws.Int64Value(out.DBClusters[0].Port) {
			return false
		}
	}
	return true
}

func preUpdate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately

	return nil
}

func (u *updater) postUpdate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err == nil {

		input := GenerateDescribeDBClustersInput(cr)
		resp, err := u.client.DescribeDBClustersWithContext(ctx, input)

		tags := resp.DBClusters[0].TagList

		add, remove := DiffTags(cr.Spec.ForProvider.Tags, tags)

		if len(add) > 0 || len(remove) > 0 {
			err := u.updateTags(ctx, cr, add, remove)
			if err != nil {
				return managed.ExternalUpdate{}, err
			}
		}
		if err != nil {
			if err != nil {
				return managed.ExternalUpdate{}, aws.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
			}
		}
		if !isPreferredMaintenanceWindowUpToDate(cr, resp) {
			return upd, errors.New("PreferredMaintenanceWindow not matching aws data")
		}

		if !isPreferredBackupWindowUpToDate(cr, resp) {
			return upd, errors.New("PreferredBackupWindow not matching aws data")
		}
	}

	return upd, err
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) (bool, error) {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))

	obj.SkipFinalSnapshot = aws.Bool(cr.Spec.ForProvider.SkipFinalSnapshot)

	if !cr.Spec.ForProvider.SkipFinalSnapshot {
		obj.FinalDBSnapshotIdentifier = aws.String(cr.Spec.ForProvider.FinalDBSnapshotIdentifier)
	}
	return false, nil
}

func filterList(cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersOutput) *svcsdk.DescribeDBClustersOutput {
	clusterIdentifier := aws.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeDBClustersOutput{}
	for _, dbCluster := range obj.DBClusters {
		if aws.StringValue(dbCluster.DBClusterIdentifier) == aws.StringValue(clusterIdentifier) {
			resp.DBClusters = append(resp.DBClusters, dbCluster)
			break
		}
	}
	return resp
}

func (e *custom) savePasswordSecret(ctx context.Context, cr *svcapitypes.DBCluster, pw string) error {
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

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	removeMap := make(map[string]string, len(spec))
	for _, t := range current {
		if addMap[aws.StringValue(t.Key)] == aws.StringValue(t.Value) {
			delete(addMap, aws.StringValue(t.Key))
			continue
		}
		removeMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, aws.String(k))
	}
	return
}

func (u *updater) updateTags(ctx context.Context, cr *svcapitypes.DBCluster, addTags []*svcsdk.Tag, removeTags []*string) error {

	arn := cr.Status.AtProvider.DBClusterARN
	if arn != nil {
		if len(removeTags) > 0 {
			inputR := &svcsdk.RemoveTagsFromResourceInput{
				ResourceName: arn,
				TagKeys:      removeTags,
			}

			_, err := u.client.RemoveTagsFromResourceWithContext(ctx, inputR)
			if err != nil {
				return errors.New(errUpdateTags)
			}
		}
		if len(addTags) > 0 {
			inputC := &svcsdk.AddTagsToResourceInput{
				ResourceName: arn,
				Tags:         addTags,
			}

			_, err := u.client.AddTagsToResourceWithContext(ctx, inputC)
			if err != nil {
				return errors.New(errUpdateTags)
			}

		}
	}
	return nil

}
