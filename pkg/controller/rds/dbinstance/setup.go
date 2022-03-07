package dbinstance

import (
	"context"
	"encoding/json"
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
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/rds"
)

// error constants
const (
	errSaveSecretFailed = "failed to save generated password to Kubernetes secret"
)

// time formats
const (
	maintenanceWindowFormat = "Mon:15:04"
	backupWindowFormat      = "15:04"
)

// SetupDBInstance adds a controller that reconciles DBInstance
func SetupDBInstance(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.DBInstanceGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.lateInitialize = lateInitialize
			e.isUpToDate = c.isUpToDate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = c.preCreate
			e.postCreate = c.postCreate
			e.preDelete = c.preDelete
			e.filterList = filterList
			e.preUpdate = c.preUpdate
			e.postUpdate = c.postUpdate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.DBInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.CreateDBInstanceInput) error {
	pw, _, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
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
	return nil
}

func (e *custom) assembleConnectionDetails(ctx context.Context, cr *svcapitypes.DBInstance) (managed.ConnectionDetails, error) {
	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey: []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername)),
	}
	pw, _, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
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

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.DBInstance, out *svcsdk.CreateDBInstanceOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	conn, err := e.assembleConnectionDetails(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	return managed.ExternalCreation{
		ConnectionDetails: conn,
	}, nil
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.ModifyDBInstanceInput) error {
	obj.DBInstanceIdentifier = aws.String(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately
	pw, pwchanged, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return err
	}
	if pwchanged {
		obj.MasterUserPassword = aws.String(pw)
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

	_, _ = e.external.Update(ctx, cr)
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.DBInstances[0].DBInstanceStatus) {
	case "available", "modifying":
		cr.SetConditions(xpv1.Available())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	}
	return obs, nil
}

func lateInitialize(in *svcapitypes.DBInstanceParameters, out *svcsdk.DescribeDBInstancesOutput) error { // nolint:gocyclo
	// (PocketMobsters): The controller should already be checking if out is nil so we *should* have a dbinstance here, always
	db := out.DBInstances[0]
	in.DBInstanceClass = aws.LateInitializeStringPtr(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = aws.LateInitializeStringPtr(in.Engine, db.Engine)

	in.AllocatedStorage = aws.LateInitializeInt64Ptr(in.AllocatedStorage, db.AllocatedStorage)
	in.AutoMinorVersionUpgrade = aws.LateInitializeBoolPtr(in.AutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	in.AvailabilityZone = aws.LateInitializeStringPtr(in.AvailabilityZone, db.AvailabilityZone)
	in.BackupRetentionPeriod = aws.LateInitializeInt64Ptr(in.BackupRetentionPeriod, db.BackupRetentionPeriod)
	in.CharacterSetName = aws.LateInitializeStringPtr(in.CharacterSetName, db.CharacterSetName)
	in.CopyTagsToSnapshot = aws.LateInitializeBoolPtr(in.CopyTagsToSnapshot, db.CopyTagsToSnapshot)
	in.DBClusterIdentifier = aws.LateInitializeStringPtr(in.DBClusterIdentifier, db.DBClusterIdentifier)
	in.DBName = aws.LateInitializeStringPtr(in.DBName, db.DBName)
	in.DeletionProtection = aws.LateInitializeBoolPtr(in.DeletionProtection, db.DeletionProtection)
	in.EnableIAMDatabaseAuthentication = aws.LateInitializeBoolPtr(in.EnableIAMDatabaseAuthentication, db.IAMDatabaseAuthenticationEnabled)
	in.EnablePerformanceInsights = aws.LateInitializeBoolPtr(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = aws.LateInitializeInt64Ptr(in.IOPS, db.Iops)
	in.KMSKeyID = aws.LateInitializeStringPtr(in.KMSKeyID, db.KmsKeyId)
	in.LicenseModel = aws.LateInitializeStringPtr(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = aws.LateInitializeStringPtr(in.MasterUsername, db.MasterUsername)

	if aws.Int64Value(db.MonitoringInterval) > 0 {
		in.MonitoringInterval = aws.LateInitializeInt64Ptr(in.MonitoringInterval, db.MonitoringInterval)
	}

	in.MonitoringRoleARN = aws.LateInitializeStringPtr(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = aws.LateInitializeBoolPtr(in.MultiAZ, db.MultiAZ)
	in.PerformanceInsightsKMSKeyID = aws.LateInitializeStringPtr(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = aws.LateInitializeInt64Ptr(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.PreferredBackupWindow = aws.LateInitializeStringPtr(in.PreferredBackupWindow, db.PreferredBackupWindow)
	in.PreferredMaintenanceWindow = aws.LateInitializeStringPtr(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = aws.LateInitializeInt64Ptr(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = aws.LateInitializeBoolPtr(in.PubliclyAccessible, db.PubliclyAccessible)
	in.StorageEncrypted = aws.LateInitializeBoolPtr(in.StorageEncrypted, db.StorageEncrypted)
	in.StorageType = aws.LateInitializeStringPtr(in.StorageType, db.StorageType)
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
	if len(in.VPCSecurityGroupIDs) == 0 && len(db.VpcSecurityGroups) != 0 {
		in.VPCSecurityGroupIDs = make([]string, len(db.VpcSecurityGroups))
		for i, val := range db.VpcSecurityGroups {
			in.VPCSecurityGroupIDs[i] = aws.StringValue(val.VpcSecurityGroupId)
		}
	}
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

	return nil
}

func (e *custom) isUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DescribeDBInstancesOutput) (bool, error) {
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
	if status == "modifying" || status == "upgrading" {
		return true, nil
	}

	_, pwChanged, err := rds.GetPassword(ctx, e.kube, cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
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

	return cmp.Equal(&svcapitypes.DBInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "SkipFinalSnapshot"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "MasterUserPasswordSecretRef"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "AutogeneratePassword"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredMaintenanceWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredBackupWindow"),
	) && !maintenanceWindowChanged && !backupWindowChanged && !pwChanged, nil
}

func createPatch(out *svcsdk.DescribeDBInstancesOutput, target *svcapitypes.DBInstanceParameters) (*svcapitypes.DBInstanceParameters, error) {
	currentParams := &svcapitypes.DBInstanceParameters{}
	err := lateInitialize(currentParams, out)
	if err != nil {
		return nil, err
	}
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
