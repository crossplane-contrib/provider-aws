package dbinstance

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	dbinstance "github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// error constants
const (
	errSaveSecretFailed = "failed to save generated password to Kubernetes secret"
)

// time formats
const (
	maintenanceWindowFormat     = "Mon:15:04"
	backupWindowFormat          = "15:04"
	errS3RestoreFailed          = "cannot restore DB instance from S3 backup"
	errSnapshotRestoreFailed    = "cannot restore DB instance from snapshot"
	errPointInTimeRestoreFailed = "cannot restore DB instance from point in time"
	errUnknownRestoreSource     = "unknown DB Instance restore source"
	statusDeleting              = "deleting"
)

// SetupDBInstance adds a controller that reconciles DBInstance
func SetupDBInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBInstanceGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.lateInitialize = lateInitialize
			e.isUpToDate = c.isUpToDate
			e.preObserve = preObserve
			e.postObserve = c.postObserve
			e.preCreate = c.preCreate
			e.preDelete = c.preDelete
			e.filterList = filterList
			e.preUpdate = c.preUpdate
			e.postUpdate = c.postUpdate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.DBInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.RDSAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DescribeDBInstancesInput) error {
	obj.DBInstanceIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.CreateDBInstanceInput) error { // nolint:gocyclo
	pw, _, err := dbinstance.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "cannot get password from the given secret")
	}
	if pw == "" && cr.Spec.ForProvider.AutogeneratePassword {
		pw, err = password.Generate()
		if err != nil {
			return errors.Wrap(err, "unable to generate a password")
		}
		if err := e.savePasswordSecret(ctx, cr, pw); err != nil {
			return errors.Wrap(err, errSaveSecretFailed)
		}
	}
	obj.MasterUserPassword = aws.String(pw)
	obj.DBInstanceIdentifier = aws.String(meta.GetExternalName(cr))
	if len(cr.Spec.ForProvider.VPCSecurityGroupIDs) > 0 {
		obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
		for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
			obj.VpcSecurityGroupIds[i] = aws.String(v)
		}
	}
	if len(cr.Spec.ForProvider.DBSecurityGroups) > 0 {
		obj.DBSecurityGroups = make([]*string, len(cr.Spec.ForProvider.DBSecurityGroups))
		for i, v := range cr.Spec.ForProvider.DBSecurityGroups {
			obj.DBSecurityGroups[i] = aws.String(v)
		}
	}
	if cr.Spec.ForProvider.RestoreFrom != nil {
		switch *cr.Spec.ForProvider.RestoreFrom.Source {
		case "S3":
			_, err := e.client.RestoreDBInstanceFromS3WithContext(ctx, dbinstance.GenerateRestoreDBInstanceFromS3Input(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
			if err != nil {
				return aws.Wrap(err, errS3RestoreFailed)
			}

		case "Snapshot":
			_, err := e.client.RestoreDBInstanceFromDBSnapshotWithContext(ctx, dbinstance.GenerateRestoreDBInstanceFromSnapshotInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
			if err != nil {
				return aws.Wrap(err, errSnapshotRestoreFailed)
			}
		case "PointInTime":
			_, err := e.client.RestoreDBInstanceToPointInTimeWithContext(ctx, dbinstance.GenerateRestoreDBInstanceToPointInTimeInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
			if err != nil {
				return aws.Wrap(err, errPointInTimeRestoreFailed)
			}
		default:
			return errors.New(errUnknownRestoreSource)

		}
	}
	return nil
}

func (e *custom) assembleConnectionDetails(ctx context.Context, cr *svcapitypes.DBInstance) (managed.ConnectionDetails, error) {
	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey: []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername)),
	}
	pw, _, err := dbinstance.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return managed.ConnectionDetails{}, errors.Wrap(err, "cannot get password from the given secret")
	}
	if pw != "" {
		conn[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(pw)
	}
	if cr.Status.AtProvider.Endpoint != nil {
		if aws.StringValue(cr.Status.AtProvider.Endpoint.Address) != "" {
			conn[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(aws.StringValue(cr.Status.AtProvider.Endpoint.Address))
		}
		if aws.Int64Value(cr.Status.AtProvider.Endpoint.Port) > 0 {
			conn[xpv1.ResourceCredentialsSecretPortKey] = []byte(strconv.FormatInt(*cr.Status.AtProvider.Endpoint.Port, 10))
		}
	}
	return conn, nil
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.ModifyDBInstanceInput) error {
	obj.DBInstanceIdentifier = aws.String(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately
	pw, pwchanged, err := dbinstance.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return err
	}
	if pwchanged {
		obj.MasterUserPassword = aws.String(pw)
	}

	if cr.Spec.ForProvider.VPCSecurityGroupIDs != nil {
		obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
		for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
			obj.VpcSecurityGroupIds[i] = aws.String(v)
		}
	}

	return nil
}

func (e *custom) postUpdate(ctx context.Context, cr *svcapitypes.DBInstance, out *svcsdk.ModifyDBInstanceOutput, _ managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	conn, err := e.assembleConnectionDetails(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	return managed.ExternalUpdate{
		ConnectionDetails: conn,
	}, nil
}

func (e *custom) preDelete(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DeleteDBInstanceInput) (bool, error) {
	obj.DBInstanceIdentifier = aws.String(meta.GetExternalName(cr))
	obj.FinalDBSnapshotIdentifier = aws.String(cr.Spec.ForProvider.FinalDBSnapshotIdentifier)
	obj.SkipFinalSnapshot = aws.Bool(cr.Spec.ForProvider.SkipFinalSnapshot)
	obj.DeleteAutomatedBackups = cr.Spec.ForProvider.DeleteAutomatedBackups

	_, _ = e.external.Update(ctx, cr)
	if *cr.Status.AtProvider.DBInstanceStatus == statusDeleting {
		return true, nil
	}
	return false, nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(resp.DBInstances[0].DBInstanceStatus) {
	case "available", "configuring-enhanced-monitoring", "storage-optimization", "backing-up":
		cr.SetConditions(xpv1.Available())
	case "modifying":
		cr.SetConditions(xpv1.Available().WithMessage("DB Instance is " + aws.StringValue(resp.DBInstances[0].DBInstanceStatus) + ", availability may vary"))
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	default:
		cr.SetConditions(xpv1.Unavailable().WithMessage("DB Instance is " + aws.StringValue(resp.DBInstances[0].DBInstanceStatus)))
	}

	obs.ConnectionDetails, _ = e.assembleConnectionDetails(ctx, cr)
	return obs, nil
}

func lateInitialize(in *svcapitypes.DBInstanceParameters, out *svcsdk.DescribeDBInstancesOutput) error { // nolint:gocyclo
	// (PocketMobsters): The controller should already be checking if out is nil so we *should* have a dbinstance here, always
	db := out.DBInstances[0]
	in.DBInstanceClass = aws.LateInitializeStringPtr(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = aws.LateInitializeStringPtr(in.Engine, db.Engine)

	in.DBClusterIdentifier = aws.LateInitializeStringPtr(in.DBClusterIdentifier, db.DBClusterIdentifier)
	// if the instance belongs to a cluster, these fields should not be lateinit,
	// to allow the user to manage these via the cluster
	if in.DBClusterIdentifier == nil {
		in.AllocatedStorage = aws.LateInitializeInt64Ptr(in.AllocatedStorage, db.AllocatedStorage)
		in.BackupRetentionPeriod = aws.LateInitializeInt64Ptr(in.BackupRetentionPeriod, db.BackupRetentionPeriod)
		in.CopyTagsToSnapshot = aws.LateInitializeBoolPtr(in.CopyTagsToSnapshot, db.CopyTagsToSnapshot)
		in.DeletionProtection = aws.LateInitializeBoolPtr(in.DeletionProtection, db.DeletionProtection)
		in.EnableIAMDatabaseAuthentication = aws.LateInitializeBoolPtr(in.EnableIAMDatabaseAuthentication, db.IAMDatabaseAuthenticationEnabled)
		in.PreferredBackupWindow = aws.LateInitializeStringPtr(in.PreferredBackupWindow, db.PreferredBackupWindow)
		in.StorageEncrypted = aws.LateInitializeBoolPtr(in.StorageEncrypted, db.StorageEncrypted)
		in.StorageType = aws.LateInitializeStringPtr(in.StorageType, db.StorageType)
		in.EngineVersion = aws.LateInitializeStringPtr(in.EngineVersion, db.EngineVersion)
		// When version 5.6 is chosen, AWS creates 5.6.41 and that's totally valid.
		// But we detect as if we need to update it all the time. Here, we assign
		// the actual full version to our spec to avoid unnecessary update signals.
		if strings.HasPrefix(aws.StringValue(db.EngineVersion), aws.StringValue(in.EngineVersion)) {
			in.EngineVersion = db.EngineVersion
		}
		if in.DBParameterGroupName == nil {
			for i := range db.DBParameterGroups {
				if db.DBParameterGroups[i].DBParameterGroupName != nil {
					in.DBParameterGroupName = db.DBParameterGroups[i].DBParameterGroupName
					break
				}
			}
		}
		if len(in.VPCSecurityGroupIDs) == 0 && len(db.VpcSecurityGroups) != 0 {
			in.VPCSecurityGroupIDs = make([]string, len(db.VpcSecurityGroups))
			for i, val := range db.VpcSecurityGroups {
				in.VPCSecurityGroupIDs[i] = aws.StringValue(val.VpcSecurityGroupId)
			}
		}
	}
	in.AutoMinorVersionUpgrade = aws.LateInitializeBoolPtr(in.AutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	in.AvailabilityZone = aws.LateInitializeStringPtr(in.AvailabilityZone, db.AvailabilityZone)
	in.CACertificateIdentifier = aws.LateInitializeStringPtr(in.CACertificateIdentifier, db.CACertificateIdentifier)
	in.CharacterSetName = aws.LateInitializeStringPtr(in.CharacterSetName, db.CharacterSetName)
	in.DBName = aws.LateInitializeStringPtr(in.DBName, db.DBName)
	in.EnablePerformanceInsights = aws.LateInitializeBoolPtr(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = aws.LateInitializeInt64Ptr(in.IOPS, db.Iops)
	kmsKey := handleKmsKey(in.KMSKeyID, db.KmsKeyId)
	in.KMSKeyID = aws.LateInitializeStringPtr(in.KMSKeyID, kmsKey)
	in.LicenseModel = aws.LateInitializeStringPtr(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = aws.LateInitializeStringPtr(in.MasterUsername, db.MasterUsername)
	in.MaxAllocatedStorage = aws.LateInitializeInt64Ptr(in.MaxAllocatedStorage, db.MaxAllocatedStorage)
	in.StorageThroughput = aws.LateInitializeInt64Ptr(in.StorageThroughput, db.StorageThroughput)

	if aws.Int64Value(db.MonitoringInterval) > 0 {
		in.MonitoringInterval = aws.LateInitializeInt64Ptr(in.MonitoringInterval, db.MonitoringInterval)
	}

	in.MonitoringRoleARN = aws.LateInitializeStringPtr(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = aws.LateInitializeBoolPtr(in.MultiAZ, db.MultiAZ)
	in.PerformanceInsightsKMSKeyID = aws.LateInitializeStringPtr(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = aws.LateInitializeInt64Ptr(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.PreferredMaintenanceWindow = aws.LateInitializeStringPtr(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = aws.LateInitializeInt64Ptr(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = aws.LateInitializeBoolPtr(in.PubliclyAccessible, db.PubliclyAccessible)
	in.Timezone = aws.LateInitializeStringPtr(in.Timezone, db.Timezone)

	if db.Endpoint != nil {
		in.Port = aws.LateInitializeInt64Ptr(in.Port, db.Endpoint.Port)
	}

	if len(in.DBSecurityGroups) == 0 && len(db.DBSecurityGroups) != 0 {
		in.DBSecurityGroups = make([]string, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			in.DBSecurityGroups[i] = aws.StringValue(val.DBSecurityGroupName)
		}
	}
	if aws.StringValue(in.DBSubnetGroupName) == "" && db.DBSubnetGroup != nil {
		in.DBSubnetGroupName = db.DBSubnetGroup.DBSubnetGroupName
	}
	if len(in.EnableCloudwatchLogsExports) == 0 && len(db.EnabledCloudwatchLogsExports) != 0 {
		in.EnableCloudwatchLogsExports = db.EnabledCloudwatchLogsExports
	}
	if len(in.ProcessorFeatures) == 0 && len(db.ProcessorFeatures) != 0 {
		in.ProcessorFeatures = make([]*svcapitypes.ProcessorFeature, len(db.ProcessorFeatures))
		for i, val := range db.ProcessorFeatures {
			in.ProcessorFeatures[i] = &svcapitypes.ProcessorFeature{
				Name:  val.Name,
				Value: val.Value,
			}
		}
	}

	return nil
}

func (e *custom) isUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DescribeDBInstancesOutput) (bool, error) { // nolint:gocyclo
	// (PocketMobsters): Creating a context here is a temporary thing until a future
	// update drops for aws-controllers-k8s/code-generator
	ctx := context.Background()

	db := out.DBInstances[0]
	patch, err := createPatch(out, &cr.Spec.ForProvider)
	if err != nil {
		return false, err
	}
	// (PocketMobsters): Certain statuses can cause us to send excessive updates because the
	// expected state of the kubernetes resource differs from the actual state of the remote
	// AWS resource temporarily. Once modifications are done, we can begin sending update requests
	// again.
	// This could be matured a bit more for specific statuses, such as not allowing storage changes
	// when the status is "storage-optimization"
	status := aws.StringValue(out.DBInstances[0].DBInstanceStatus)
	if status == "modifying" || status == "upgrading" || status == "rebooting" || status == "creating" || status == "deleting" {
		return true, nil
	}

	_, pwChanged, err := dbinstance.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, err
	}

	// (PocketMobsters): AWS reformats our preferred time windows for backups and maintenance
	// so we can't rely on automatic equality checks for them
	maintenanceWindowChanged, err := compareTimeRanges(maintenanceWindowFormat, cr.Spec.ForProvider.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	if err != nil {
		return false, err
	}
	backupWindowChanged, err := compareTimeRanges(backupWindowFormat, cr.Spec.ForProvider.PreferredBackupWindow, db.PreferredBackupWindow)
	if err != nil {
		return false, err
	}

	versionChanged := !isEngineVersionUpToDate(cr, out)

	vpcSGsChanged := !areVPCSecurityGroupIDsUpToDate(cr, db)

	dbParameterGroupChanged := !isDBParameterGroupNameUpToDate(cr, db)

	diff := cmp.Diff(&svcapitypes.DBInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "AllowMajorVersionUpgrade"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "DBParameterGroupName"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "EngineVersion"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "SkipFinalSnapshot"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "MasterUserPasswordSecretRef"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "AutogeneratePassword"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredMaintenanceWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredBackupWindow"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "ApplyImmediately"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "RestoreFrom"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "VPCSecurityGroupIDs"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "DeleteAutomatedBackups"),
	)

	if diff == "" && !maintenanceWindowChanged && !backupWindowChanged && !pwChanged && !versionChanged && !vpcSGsChanged && !dbParameterGroupChanged {
		return true, nil
	}

	diff = "Found observed difference in dbinstance\n" + diff
	if maintenanceWindowChanged {
		diff += "\ndesired maintanaceWindow: "
		diff += *cr.Spec.ForProvider.PreferredMaintenanceWindow
		diff += "\nobserved maintanaceWindow: "
		diff += *db.PreferredMaintenanceWindow
	}
	if backupWindowChanged {
		diff += "\ndesired backupWindow: "
		diff += *cr.Spec.ForProvider.PreferredBackupWindow
		diff += "\nobserved backupWindow: "
		diff += *db.PreferredBackupWindow
	}

	log.Println(diff)

	return false, nil
}

func isEngineVersionUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DescribeDBInstancesOutput) bool {
	// If EngineVersion is not set, AWS sets a default value,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.EngineVersion != nil {
		if out.DBInstances[0].EngineVersion == nil {
			return false
		}

		// Upgrade is only necessary if the spec version is higher.
		// Downgrades are not possible in AWS.
		c := utils.CompareEngineVersions(*cr.Spec.ForProvider.EngineVersion, *out.DBInstances[0].EngineVersion)
		return c <= 0
	}
	return true
}

func createPatch(out *svcsdk.DescribeDBInstancesOutput, target *svcapitypes.DBInstanceParameters) (*svcapitypes.DBInstanceParameters, error) {
	currentParams := &svcapitypes.DBInstanceParameters{}
	err := lateInitialize(currentParams, out)
	if err != nil {
		return nil, err
	}
	currentParams.KMSKeyID = handleKmsKey(target.KMSKeyID, currentParams.KMSKeyID)
	jsonPatch, err := aws.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.DBInstanceParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

func compareTimeRanges(format string, expectedWindow *string, actualWindow *string) (bool, error) {
	if aws.StringValue(expectedWindow) == "" {
		// no window to set, don't bother
		return false, nil
	}
	if aws.StringValue(actualWindow) == "" {
		// expected is set but actual is not, so we should set it
		return true, nil
	}
	// all windows here have a "-" in between two values in the expected format, so just split
	leftSpans := strings.Split(*expectedWindow, "-")
	rightSpans := strings.Split(*actualWindow, "-")
	for i := range leftSpans {
		left, err := time.Parse(format, leftSpans[i])
		if err != nil {
			return false, err
		}
		right, err := time.Parse(format, rightSpans[i])
		if err != nil {
			return false, err
		}
		if left != right {
			return true, nil
		}
	}
	return false, nil
}

func areVPCSecurityGroupIDsUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DBInstance) bool {
	desiredIDs := cr.Spec.ForProvider.VPCSecurityGroupIDs

	// if user is fine with default SG or lets DBCluster manage it
	// (removing all SGs is not possible, AWS will keep last set SGs)
	if len(desiredIDs) == 0 {
		return true
	}

	actualGroups := out.VpcSecurityGroups

	if len(desiredIDs) != len(actualGroups) {
		return false
	}

	actualIDs := make([]string, 0, len(actualGroups))
	for _, grp := range actualGroups {
		actualIDs = append(actualIDs, *grp.VpcSecurityGroupId)
	}

	sort.Strings(desiredIDs)
	sort.Strings(actualIDs)

	return cmp.Equal(desiredIDs, actualIDs)
}

func isDBParameterGroupNameUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DBInstance) bool {
	desiredGroup := cr.Spec.ForProvider.DBParameterGroupName

	// if user is fine with default DBParameterGroup or lets DBCluster manage it
	if desiredGroup == nil {
		return true
	}

	actualGroups := out.DBParameterGroups

	for _, grp := range actualGroups {

		if aws.StringValue(grp.DBParameterGroupName) == aws.StringValue(desiredGroup) {
			return true
		}

	}

	return false
}

func filterList(cr *svcapitypes.DBInstance, obj *svcsdk.DescribeDBInstancesOutput) *svcsdk.DescribeDBInstancesOutput {
	resp := &svcsdk.DescribeDBInstancesOutput{}
	for _, dbInstance := range obj.DBInstances {
		if aws.StringValue(dbInstance.DBInstanceIdentifier) == meta.GetExternalName(cr) {
			resp.DBInstances = append(resp.DBInstances, dbInstance)
			break
		}
	}
	return resp
}

func (e *custom) savePasswordSecret(ctx context.Context, cr *svcapitypes.DBInstance, pw string) error {
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

func handleKmsKey(inKey *string, dbKey *string) *string {
	if inKey != nil && dbKey != nil && !strings.Contains(*inKey, "/") {
		lastInd := strings.LastIndex(*dbKey, "/")
		keyID := (*dbKey)[lastInd+1:]
		return &keyID
	}
	return dbKey
}
