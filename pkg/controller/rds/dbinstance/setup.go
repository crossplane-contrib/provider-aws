package dbinstance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	dbinstance "github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// error constants
const (
	errCreateReadReplica        = "cannot creat DB instance read replica"
	errS3RestoreFailed          = "cannot restore DB instance from S3 backup"
	errSnapshotRestoreFailed    = "cannot restore DB instance from snapshot"
	errPointInTimeRestoreFailed = "cannot restore DB instance from point in time"
	errUnknownRestoreSource     = "unknown DB Instance restore source"
	errAddTags                  = "cannot add tags"
	errRemoveTags               = "cannot remove tags"
)

// time formats
const (
	maintenanceWindowFormat = "Mon:15:04"
	backupWindowFormat      = "15:04"
)

// database roles
const (
	databaseRoleStandalone  = "Instance"
	databaseRolePrimary     = "Primary"
	databaseRoleReadReplica = "Replica"
)

// other
const (
	statusDeleting = "deleting"
)

// SetupDBInstance adds a controller that reconciles DBInstance
func SetupDBInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBInstanceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&customConnector{kube: mgr.GetClient()}),
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
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.DBInstance{}).
		Complete(r)
}

type customConnector struct {
	kube client.Client
}

type customExternal struct {
	external
	shared *shared
}

type shared struct {
	external
	cache *cache
}

type cache struct {
	addTags         []*svcsdk.Tag
	removeTags      []*string
	desiredPassword string
}

func newCustomExternal(kube client.Client, client svcsdkapi.RDSAPI) *customExternal {
	s := &shared{cache: &cache{}}
	e := external{
		kube:           kube,
		client:         client,
		preObserve:     preObserve,
		isUpToDate:     s.isUpToDate,
		postObserve:    s.postObserve,
		preUpdate:      s.preUpdate,
		postUpdate:     s.postUpdate,
		preCreate:      s.preCreate,
		postCreate:     nopPostCreate,
		preDelete:      s.preDelete,
		postDelete:     s.postDelete,
		filterList:     filterList,
		lateInitialize: lateInitialize,
	}
	s.external = e
	return &customExternal{
		external: e,
		shared:   s,
	}
}

func (c *customConnector) Connect(ctx context.Context, cr *svcapitypes.DBInstance) (managed.TypedExternalClient[*svcapitypes.DBInstance], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newCustomExternal(c.kube, svcsdk.New(sess)), nil
}

func (c *customExternal) Create(ctx context.Context, cr *svcapitypes.DBInstance) (managed.ExternalCreation, error) {
	if cr.Spec.ForProvider.SourceDBInstanceID != nil || cr.Spec.ForProvider.SourceDBInstanceIDRef != nil ||
		cr.Spec.ForProvider.SourceDBInstanceIDSelector != nil || cr.Spec.ForProvider.SourceDBClusterID != nil ||
		cr.Spec.ForProvider.SourceDBClusterIDRef != nil || cr.Spec.ForProvider.SourceDBClusterIDSelector != nil {
		cr.Status.SetConditions(xpv1.Creating())
		cr.Status.AtProvider.DatabaseRole = aws.String(databaseRoleReadReplica)

		createDBInstanceReadReplicaInput := dbinstance.GenerateCreateDBInstanceReadReplicaInput(cr)
		createDBInstanceReadReplicaInput.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

		_, err := c.client.CreateDBInstanceReadReplicaWithContext(ctx, createDBInstanceReadReplicaInput)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreateReadReplica)
		}
		return managed.ExternalCreation{}, nil
	}
	cr.Status.AtProvider.DatabaseRole = aws.String(databaseRoleStandalone)
	return c.external.Create(ctx, cr)
}

func preObserve(_ context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DescribeDBInstancesInput) error {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (s *shared) preCreate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.CreateDBInstanceInput) (err error) { //nolint:gocyclo
	// If the DBInstance is going to be created as a read replica, we do not need to set the MasterUserPassword and the others
	if cr.Spec.ForProvider.SourceDBInstanceID != nil || cr.Spec.ForProvider.SourceDBInstanceIDRef != nil ||
		cr.Spec.ForProvider.SourceDBInstanceIDSelector != nil || cr.Spec.ForProvider.SourceDBClusterID != nil ||
		cr.Spec.ForProvider.SourceDBClusterIDRef != nil || cr.Spec.ForProvider.SourceDBClusterIDSelector != nil {
		return nil
	}
	restoreFrom := cr.Spec.ForProvider.RestoreFrom
	autogenerate := cr.Spec.ForProvider.AutogeneratePassword
	masterUserPasswordSecretRef := cr.Spec.ForProvider.MasterUserPasswordSecretRef
	clusterIdentifier := cr.Spec.ForProvider.DBClusterIdentifier

	var pw string
	switch {
	case clusterIdentifier != nil:
		break
	case masterUserPasswordSecretRef == nil && restoreFrom == nil && !autogenerate:
		return errors.New(dbinstance.ErrNoMasterUserPasswordSecretRefNorAutogenerateNoRestore)
	case masterUserPasswordSecretRef == nil && autogenerate:
		pw, err = password.Generate()
	case masterUserPasswordSecretRef != nil && autogenerate,
		masterUserPasswordSecretRef != nil && !autogenerate:
		pw, err = dbinstance.GetSecretValue(ctx, s.kube, masterUserPasswordSecretRef)
	}
	if err != nil {
		return errors.Wrap(err, dbinstance.ErrNoRetrievePasswordOrGenerate)
	}

	obj.MasterUserPassword = pointer.ToOrNilIfZeroValue(pw)
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// VpcSecurityGroupIds cannot be set on an instance that belongs to a DBCluster
	// NOTE: Unlike in preUpdate we are using spec here because status is not yet available.
	if cr.Spec.ForProvider.DBClusterIdentifier == nil {
		if len(cr.Spec.ForProvider.VPCSecurityGroupIDs) > 0 {
			obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
			for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
				obj.VpcSecurityGroupIds[i] = pointer.ToOrNilIfZeroValue(v)
			}
		}
	} else {
		obj.VpcSecurityGroupIds = nil
	}

	if len(cr.Spec.ForProvider.DBSecurityGroups) > 0 {
		obj.DBSecurityGroups = make([]*string, len(cr.Spec.ForProvider.DBSecurityGroups))
		for i, v := range cr.Spec.ForProvider.DBSecurityGroups {
			obj.DBSecurityGroups[i] = pointer.ToOrNilIfZeroValue(v)
		}
	}

	passwordRestoreInfo := map[string]string{dbinstance.PasswordCacheKey: pw}
	if restoreFrom != nil {
		passwordRestoreInfo[dbinstance.RestoreFlagCacheKay] = string(dbinstance.RestoreStateRestored)

		switch *restoreFrom.Source {
		case "S3":
			_, err := s.client.RestoreDBInstanceFromS3WithContext(ctx, dbinstance.GenerateRestoreDBInstanceFromS3Input(meta.GetExternalName(cr), pw, &cr.Spec.ForProvider))
			if err != nil {
				return errorutils.Wrap(err, errS3RestoreFailed)
			}

		case "Snapshot":
			_, err := s.client.RestoreDBInstanceFromDBSnapshotWithContext(ctx, dbinstance.GenerateRestoreDBInstanceFromSnapshotInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
			if err != nil {
				return errorutils.Wrap(err, errSnapshotRestoreFailed)
			}
		case "PointInTime":
			_, err := s.client.RestoreDBInstanceToPointInTimeWithContext(ctx, dbinstance.GenerateRestoreDBInstanceToPointInTimeInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
			if err != nil {
				return errorutils.Wrap(err, errPointInTimeRestoreFailed)
			}
		default:
			return errors.New(errUnknownRestoreSource)
		}
	}

	if cr.Spec.ForProvider.EngineVersion != nil {
		obj.EngineVersion = cr.Spec.ForProvider.EngineVersion
	}

	if _, err = dbinstance.Cache(ctx, s.kube, cr, passwordRestoreInfo); err != nil {
		return errors.Wrap(err, dbinstance.ErrCachePassword)
	}

	// for storageType gp3 below engine specific allocatedStorage threshold, do not send iops and storageThroughput
	// to avoid errors like "You can't specify IOPS or storage throughput for engine postgres and a storage size less than 400."
	// This allows users to set iops/storageThroughput to the default values themselves.
	if isStorageTypeGP3BelowAllocatedStorageThreshold(cr) {
		obj.Iops = nil
		obj.StorageThroughput = nil
	}

	return nil
}

func (s *shared) updateConnectionDetails(ctx context.Context, cr *svcapitypes.DBInstance, details managed.ConnectionDetails) (managed.ConnectionDetails, error) {
	if details == nil {
		details = managed.ConnectionDetails{}
	}

	details[xpv1.ResourceCredentialsSecretUserKey] = []byte(pointer.StringValue(cr.Spec.ForProvider.MasterUsername))
	if s.cache.desiredPassword == "" {
		pw, err := dbinstance.GetDesiredPassword(ctx, s.kube, cr)
		if err != nil && pointer.StringValue(cr.Status.AtProvider.DatabaseRole) != databaseRoleReadReplica {
			return details, errors.Wrap(err, dbinstance.ErrGetCachedPassword)
		}
		s.cache.desiredPassword = pw
	}
	details[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(s.cache.desiredPassword)

	if cr.Status.AtProvider.Endpoint == nil {
		return details, nil
	}
	if pointer.StringValue(cr.Status.AtProvider.Endpoint.Address) != "" {
		details[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(pointer.StringValue(cr.Status.AtProvider.Endpoint.Address))
	}
	if pointer.Int64Value(cr.Status.AtProvider.Endpoint.Port) > 0 {
		details[xpv1.ResourceCredentialsSecretPortKey] = []byte(strconv.FormatInt(*cr.Status.AtProvider.Endpoint.Port, 10))
	}

	return details, nil
}

func (s *shared) preUpdate(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.ModifyDBInstanceInput) (err error) {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately
	obj.MasterUserPassword = pointer.ToOrNilIfZeroValue(s.cache.desiredPassword)

	// VpcSecurityGroupIds cannot be set on an instance that belongs to a DBCluster
	if cr.Status.AtProvider.DBClusterIdentifier == nil {
		if cr.Spec.ForProvider.VPCSecurityGroupIDs != nil {
			obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
			for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
				obj.VpcSecurityGroupIds[i] = pointer.ToOrNilIfZeroValue(v)
			}
		}
	} else {
		obj.VpcSecurityGroupIds = nil
	}

	// for storageType gp3 below engine specific allocatedStorage threshold, do not send iops and storageThroughput
	// to avoid errors like "You can't specify IOPS or storage throughput for engine postgres and a storage size less than 400."
	// This allows users to set iops/storageThroughput to the default values themselves.
	if isStorageTypeGP3BelowAllocatedStorageThreshold(cr) {
		obj.Iops = nil
		obj.StorageThroughput = nil
	}

	input := GenerateDescribeDBInstancesInput(cr)

	out, err := s.client.DescribeDBInstancesWithContext(ctx, input)
	if err != nil {
		return errors.Wrap(err, dbinstance.ErrDescribe)
	}
	if !isEngineVersionUpToDate(cr, out) && cr.Spec.ForProvider.EngineVersion != nil {
		obj.EngineVersion = cr.Spec.ForProvider.EngineVersion // add EngineVersion if changed and no downgrade
	}

	return nil
}

func (s *shared) postUpdate(ctx context.Context, cr *svcapitypes.DBInstance, out *svcsdk.ModifyDBInstanceOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	upd.ConnectionDetails, err = s.updateConnectionDetails(ctx, cr, upd.ConnectionDetails)

	if s.cache.desiredPassword != "" {
		_, err = dbinstance.Cache(ctx, s.kube, cr, map[string]string{
			dbinstance.PasswordCacheKey:    s.cache.desiredPassword,
			dbinstance.RestoreFlagCacheKay: "", // reset restore flag
		})
		if err != nil {
			return upd, errors.Wrap(err, dbinstance.ErrCachePassword)
		}
	}

	// Update tags if necessary
	if len(s.cache.addTags) > 0 {
		_, err := s.client.AddTagsToResourceWithContext(ctx, &svcsdk.AddTagsToResourceInput{
			ResourceName: out.DBInstance.DBInstanceArn,
			Tags:         s.cache.addTags,
		})
		if err != nil {
			return upd, errors.Wrap(err, errAddTags)
		}
	}
	if len(s.cache.removeTags) > 0 {
		_, err := s.client.RemoveTagsFromResourceWithContext(ctx, &svcsdk.RemoveTagsFromResourceInput{
			ResourceName: out.DBInstance.DBInstanceArn,
			TagKeys:      s.cache.removeTags,
		})
		if err != nil {
			return upd, errors.Wrap(err, errRemoveTags)
		}
	}

	return upd, err
}

func (s *shared) preDelete(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DeleteDBInstanceInput) (bool, error) {
	obj.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.FinalDBSnapshotIdentifier = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.FinalDBSnapshotIdentifier)
	obj.SkipFinalSnapshot = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.SkipFinalSnapshot)
	obj.DeleteAutomatedBackups = cr.Spec.ForProvider.DeleteAutomatedBackups

	_, _ = s.external.Update(ctx, cr)
	if *cr.Status.AtProvider.DBInstanceStatus == statusDeleting {
		return true, nil
	}
	return false, nil
}

func (s *shared) postDelete(ctx context.Context, cr *svcapitypes.DBInstance, obj *svcsdk.DeleteDBInstanceOutput, err error) (managed.ExternalDelete, error) {
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	return managed.ExternalDelete{}, dbinstance.DeleteCache(ctx, s.kube, cr)
}

func (s *shared) postObserve(ctx context.Context, cr *svcapitypes.DBInstance, resp *svcsdk.DescribeDBInstancesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	cr.Status.AtProvider.KMSKeyID = resp.DBInstances[0].KmsKeyId

	cr.Spec.ForProvider.DBClusterIdentifier = resp.DBInstances[0].DBClusterIdentifier

	switch pointer.StringValue(resp.DBInstances[0].DBInstanceStatus) {
	case "available", "configuring-enhanced-monitoring", "storage-optimization", "backing-up":
		cr.SetConditions(xpv1.Available())
	case "modifying":
		cr.SetConditions(xpv1.Available().WithMessage("DB Instance is " + pointer.StringValue(resp.DBInstances[0].DBInstanceStatus) + ", availability may vary"))
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	default:
		cr.SetConditions(xpv1.Unavailable().WithMessage("DB Instance is " + pointer.StringValue(resp.DBInstances[0].DBInstanceStatus)))
	}

	obs.ConnectionDetails, err = s.updateConnectionDetails(ctx, cr, obs.ConnectionDetails)
	return obs, err
}

func lateInitialize(in *svcapitypes.DBInstanceParameters, out *svcsdk.DescribeDBInstancesOutput) error { //nolint:gocyclo
	// (PocketMobsters): The controller should already be checking if out is nil so we *should* have a dbinstance here, always
	db := out.DBInstances[0]
	in.DBInstanceClass = pointer.LateInitialize(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = pointer.LateInitialize(in.Engine, db.Engine)

	in.DBClusterIdentifier = pointer.LateInitialize(in.DBClusterIdentifier, db.DBClusterIdentifier)
	// if the instance belongs to a cluster, these fields should not be lateinit,
	// to allow the user to manage these via the cluster
	if in.DBClusterIdentifier == nil {
		in.AllocatedStorage = pointer.LateInitialize(in.AllocatedStorage, db.AllocatedStorage)
		in.BackupRetentionPeriod = pointer.LateInitialize(in.BackupRetentionPeriod, db.BackupRetentionPeriod)
		in.CopyTagsToSnapshot = pointer.LateInitialize(in.CopyTagsToSnapshot, db.CopyTagsToSnapshot)
		in.DeletionProtection = pointer.LateInitialize(in.DeletionProtection, db.DeletionProtection)
		in.EnableIAMDatabaseAuthentication = pointer.LateInitialize(in.EnableIAMDatabaseAuthentication, db.IAMDatabaseAuthenticationEnabled)
		in.PreferredBackupWindow = pointer.LateInitialize(in.PreferredBackupWindow, db.PreferredBackupWindow)
		in.StorageEncrypted = pointer.LateInitialize(in.StorageEncrypted, db.StorageEncrypted)
		in.StorageType = pointer.LateInitialize(in.StorageType, db.StorageType)
		in.EngineVersion = pointer.LateInitialize(in.EngineVersion, db.EngineVersion)
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
				in.VPCSecurityGroupIDs[i] = pointer.StringValue(val.VpcSecurityGroupId)
			}
		}
	}
	in.AutoMinorVersionUpgrade = pointer.LateInitialize(in.AutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	in.AvailabilityZone = pointer.LateInitialize(in.AvailabilityZone, db.AvailabilityZone)
	in.CACertificateIdentifier = pointer.LateInitialize(in.CACertificateIdentifier, db.CACertificateIdentifier)
	in.CharacterSetName = pointer.LateInitialize(in.CharacterSetName, db.CharacterSetName)
	in.DBName = pointer.LateInitialize(in.DBName, db.DBName)
	in.EnablePerformanceInsights = pointer.LateInitialize(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = pointer.LateInitialize(in.IOPS, db.Iops)
	kmsKey := handleKmsKey(in.KMSKeyID, db.KmsKeyId)
	in.KMSKeyID = pointer.LateInitialize(in.KMSKeyID, kmsKey)
	in.LicenseModel = pointer.LateInitialize(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = pointer.LateInitialize(in.MasterUsername, db.MasterUsername)
	in.MaxAllocatedStorage = pointer.LateInitialize(in.MaxAllocatedStorage, db.MaxAllocatedStorage)
	in.StorageThroughput = pointer.LateInitialize(in.StorageThroughput, db.StorageThroughput)

	if pointer.Int64Value(db.MonitoringInterval) > 0 {
		in.MonitoringInterval = pointer.LateInitialize(in.MonitoringInterval, db.MonitoringInterval)
	}

	in.MonitoringRoleARN = pointer.LateInitialize(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = pointer.LateInitialize(in.MultiAZ, db.MultiAZ)
	in.PerformanceInsightsKMSKeyID = pointer.LateInitialize(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = pointer.LateInitialize(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.PreferredMaintenanceWindow = pointer.LateInitialize(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = pointer.LateInitialize(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = pointer.LateInitialize(in.PubliclyAccessible, db.PubliclyAccessible)
	in.Timezone = pointer.LateInitialize(in.Timezone, db.Timezone)

	if db.Endpoint != nil {
		in.Port = pointer.LateInitialize(in.Port, db.Endpoint.Port)
	}

	if len(in.DBSecurityGroups) == 0 && len(db.DBSecurityGroups) != 0 {
		in.DBSecurityGroups = make([]string, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			in.DBSecurityGroups[i] = pointer.StringValue(val.DBSecurityGroupName)
		}
	}
	if pointer.StringValue(in.DBSubnetGroupName) == "" && db.DBSubnetGroup != nil {
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

func (s *shared) isUpToDate(ctx context.Context, cr *svcapitypes.DBInstance, out *svcsdk.DescribeDBInstancesOutput) (upToDate bool, diff string, err error) { //nolint:gocyclo
	db := out.DBInstances[0]

	patch, err := createPatch(out, &cr.Spec.ForProvider)
	if err != nil {
		return false, "", err
	}
	// (PocketMobsters): Certain statuses can cause us to send excessive updates because the
	// expected state of the kubernetes resource differs from the actual state of the remote
	// AWS resource temporarily. Once modifications are done, we can begin sending update requests
	// again.
	// This could be matured a bit more for specific statuses, such as not allowing storage changes
	// when the status is "storage-optimization"
	status := pointer.StringValue(out.DBInstances[0].DBInstanceStatus)
	if status == "modifying" || status == "upgrading" || status == "rebooting" || status == "creating" || status == "deleting" {
		return true, "", nil
	}
	switch {
	case db.ReadReplicaDBClusterIdentifiers != nil || db.ReadReplicaDBInstanceIdentifiers != nil:
		cr.Status.AtProvider.DatabaseRole = aws.String(databaseRolePrimary)
	case db.ReadReplicaSourceDBClusterIdentifier != nil || db.ReadReplicaSourceDBInstanceIdentifier != nil:
		cr.Status.AtProvider.DatabaseRole = aws.String(databaseRoleReadReplica)
	default:
		cr.Status.AtProvider.DatabaseRole = aws.String(databaseRoleStandalone)
	}

	autogenerate := cr.Spec.ForProvider.AutogeneratePassword
	masterUserPasswordSecretRef := cr.Spec.ForProvider.MasterUserPasswordSecretRef
	cachedMasterPasswordExist := true
	_, err = dbinstance.GetCachedPassword(ctx, s.kube, cr)
	if err != nil {
		cachedMasterPasswordExist = false
	}

	// If the instance is a read replica and the password was not created before, and it is not assumed to be
	// generated/created(by autogenerate or masterUserPasswordSecretRef), we don't check the password. By
	// default, a read replica has the same credentials as the primary instance.
	if !(pointer.StringValue(cr.Status.AtProvider.DatabaseRole) == databaseRoleReadReplica &&
		!autogenerate && masterUserPasswordSecretRef == nil && !cachedMasterPasswordExist) {

		passwordUpToDate, err := dbinstance.PasswordUpToDate(ctx, s.kube, cr)
		if err != nil {
			return false, "", errors.Wrap(err, dbinstance.ErrNoPasswordUpToDate)
		}
		if !passwordUpToDate {
			return false, "", nil
		}
	}

	// (PocketMobsters): AWS reformats our preferred time windows for backups and maintenance,
	// so we can't rely on automatic equality checks for them
	maintenanceWindowChanged, err := compareTimeRanges(maintenanceWindowFormat, cr.Spec.ForProvider.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	if err != nil {
		return false, "", err
	}
	backupWindowChanged, err := compareTimeRanges(backupWindowFormat, cr.Spec.ForProvider.PreferredBackupWindow, db.PreferredBackupWindow)
	if err != nil {
		return false, "", err
	}

	// Depending on whether the instance was created as gp2 or modified from another type (s.g. gp3) to gp2,
	// AWS provides different responses for IOPS/StorageThroughput (either 0 or nil).
	// Therefore, we consider both 0 and nil to be equivalent.
	iopsChanged := !(pointer.Int64Value(cr.Spec.ForProvider.IOPS) == pointer.Int64Value(db.Iops))
	storageThroughputChanged := !(pointer.Int64Value(cr.Spec.ForProvider.StorageThroughput) == pointer.Int64Value(db.StorageThroughput))

	versionChanged := !isEngineVersionUpToDate(cr, out)

	vpcSGsChanged := !areVPCSecurityGroupIDsUpToDate(cr, db)
	dbParameterGroupChanged := !isDBParameterGroupNameUpToDate(cr, db)
	optionGroupChanged := !isOptionGroupUpToDate(cr, db)

	diff = cmp.Diff(&svcapitypes.DBInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "AllowMajorVersionUpgrade"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "DBParameterGroupName"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "EngineVersion"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "IOPS"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "StorageThroughput"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "SkipFinalSnapshot"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "MasterUserPasswordSecretRef"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "AutogeneratePassword"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredMaintenanceWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "PreferredBackupWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBInstanceParameters{}, "OptionGroupName"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "ApplyImmediately"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "RestoreFrom"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "VPCSecurityGroupIDs"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "DeleteAutomatedBackups"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{},
			"SourceDBClusterID", "SourceDBClusterIDRef", "SourceDBClusterIDSelector",
			"SourceDBInstanceID", "SourceDBInstanceIDRef", "SourceDBInstanceIDSelector"),
		cmpopts.IgnoreFields(svcapitypes.CustomDBInstanceParameters{}, "TagIgnorePrefixes"),
	)

	ignore := append([]string{"aws:"}, cr.Spec.ForProvider.TagIgnorePrefixes...)
	var observedTags []*svcsdk.Tag
	if db.TagList != nil {
		for _, tag := range db.TagList { // index discarded with _
			if utils.ShouldIgnore(pointer.StringValue(tag.Key), ignore) {
				continue
			}
			observedTags = append(observedTags, &svcsdk.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}
	s.cache.addTags, s.cache.removeTags = utils.DiffTags(cr.Spec.ForProvider.Tags, observedTags)
	tagsChanged := len(s.cache.addTags) != 0 || len(s.cache.removeTags) != 0

	if diff == "" && !maintenanceWindowChanged && !backupWindowChanged && !iopsChanged && !storageThroughputChanged && !versionChanged && !vpcSGsChanged && !dbParameterGroupChanged && !optionGroupChanged && !tagsChanged {
		return true, diff, nil
	}

	diff = "Found observed difference in dbinstance\n" + diff
	if maintenanceWindowChanged {
		diff += "\ndesired maintenanceWindow: "
		diff += *cr.Spec.ForProvider.PreferredMaintenanceWindow
		diff += "\nobserved maintenanceWindow: "
		diff += *db.PreferredMaintenanceWindow
	}
	if backupWindowChanged {
		diff += "\ndesired backupWindow: "
		diff += *cr.Spec.ForProvider.PreferredBackupWindow
		diff += "\nobserved backupWindow: "
		diff += *db.PreferredBackupWindow
	}
	if iopsChanged {
		diff += fmt.Sprintf("\ndesired iops: %d \nobserved iops: %d ", pointer.Int64Value(cr.Spec.ForProvider.IOPS), pointer.Int64Value(db.Iops))
	}
	if storageThroughputChanged {
		diff += fmt.Sprintf("\ndesired storageThroughput: %d \nobserved storageThroughput: %d ", pointer.Int64Value(cr.Spec.ForProvider.StorageThroughput), pointer.Int64Value(db.StorageThroughput))
	}
	if versionChanged {
		diff += fmt.Sprintf("\ndesired engineVersion: %s \nobserved engineVersion: %s ", pointer.StringValue(cr.Spec.ForProvider.EngineVersion), pointer.StringValue(db.EngineVersion))
	}
	if vpcSGsChanged {
		diff += fmt.Sprintf("\ndesired vpcSecurityGroupIDs: %v \nobserved vpcSecurityGroupIDs: ", cr.Spec.ForProvider.VPCSecurityGroupIDs)
		for _, grp := range db.VpcSecurityGroups {
			diff += fmt.Sprintf("\n - %s ", pointer.StringValue(grp.VpcSecurityGroupId))
		}
	}
	if dbParameterGroupChanged {
		diff += fmt.Sprintf("\ndesired dbParameterGroupName: %s \nobserved dbParameterGroupName: %s ", pointer.StringValue(cr.Spec.ForProvider.DBParameterGroupName), pointer.StringValue(db.DBParameterGroups[0].DBParameterGroupName))
	}
	if optionGroupChanged {
		diff += fmt.Sprintf("\ndesired optionGroupName: %s \nobserved optionGroupName: %s ", pointer.StringValue(cr.Spec.ForProvider.OptionGroupName), pointer.StringValue(db.OptionGroupMemberships[0].OptionGroupName))
	}
	if tagsChanged {
		diff += fmt.Sprintf("\nadd %d tag(s) and remove %d tag(s)", len(s.cache.addTags), len(s.cache.removeTags))
	}

	log.Println(diff)

	return false, diff, nil
}

func isEngineVersionUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DescribeDBInstancesOutput) bool {
	// If EngineVersion is not set, AWS sets a default value,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.EngineVersion != nil {
		if out.DBInstances[0].EngineVersion == nil {
			return false
		}

		// Upgrade is only necessary if the spec version is higher.
		// Downgrades are not possible in pointer.
		c := utils.CompareEngineVersions(*cr.Spec.ForProvider.EngineVersion, *out.DBInstances[0].EngineVersion)
		return c <= 0
	}
	return true
}

func isOptionGroupUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DBInstance) bool {
	// If OptionGroupName is not set, AWS sets a default OptionGroup,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.OptionGroupName != nil {
		for _, group := range out.OptionGroupMemberships {
			if group.OptionGroupName != nil && (pointer.StringValue(group.OptionGroupName) == pointer.StringValue(cr.Spec.ForProvider.OptionGroupName)) {

				switch pointer.StringValue(group.Status) {
				case "pending-maintenance-apply":
					// If ApplyImmediately was turned on after the OptionGroup change was requested,
					// we can make a new Modify request
					if pointer.BoolValue(cr.Spec.ForProvider.ApplyImmediately) {
						return false
					}
					return true
				case "pending-maintenance-removal":
					return false
				default: // "in-sync", "applying", "pending-apply", "pending-removal", "removing", "failed"
					return true
				}
			}
		}
		return false
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
	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.DBInstanceParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

func parseTimeWindowSpan(span, format string) (string, time.Time, error) {
	// Regex: optional day, always HH:MM
	re := regexp.MustCompile(`^\s*([A-Za-z]+:)?(\d{2}:\d{2})\s*$`)
	matches := re.FindStringSubmatch(span)
	if len(matches) != 3 {
		return "", time.Time{}, errors.New("invalid time format, expected '[day:]HH:MM'")
	}
	day := ""
	if matches[1] != "" {
		day = strings.ToLower(strings.TrimSuffix(matches[1], ":"))
	}
	t, err := time.Parse(format, strings.TrimSpace(span))
	if err != nil {
		return day, time.Time{}, err
	}
	return day, t, nil
}

func compareTimeRanges(format string, expectedWindow *string, actualWindow *string) (bool, error) {
	if pointer.StringValue(expectedWindow) == "" {
		return false, nil
	}
	if pointer.StringValue(actualWindow) == "" {
		return true, nil
	}
	leftSpans := strings.Split(*expectedWindow, "-")
	rightSpans := strings.Split(*actualWindow, "-")

	if len(leftSpans) != 2 || len(rightSpans) != 2 {
		return false, errors.New("invalid time window format")
	}

	for i := range leftSpans {
		dayA, timeA, err := parseTimeWindowSpan(leftSpans[i], format)
		if err != nil {
			return false, err
		}
		dayB, timeB, err := parseTimeWindowSpan(rightSpans[i], format)
		if err != nil {
			return false, err
		}
		if dayA != dayB || !timeA.Equal(timeB) {
			return true, nil
		}
	}
	return false, nil
}

func areVPCSecurityGroupIDsUpToDate(cr *svcapitypes.DBInstance, out *svcsdk.DBInstance) bool {
	// VPCSecurityGroupIDs is ignored for instances that belong to a cluster.
	if out.DBClusterIdentifier != nil {
		return true
	}

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

		if pointer.StringValue(grp.DBParameterGroupName) == pointer.StringValue(desiredGroup) {
			return true
		}

	}

	return false
}

func filterList(cr *svcapitypes.DBInstance, obj *svcsdk.DescribeDBInstancesOutput) *svcsdk.DescribeDBInstancesOutput {
	resp := &svcsdk.DescribeDBInstancesOutput{}
	for _, dbInstance := range obj.DBInstances {
		if pointer.StringValue(dbInstance.DBInstanceIdentifier) == meta.GetExternalName(cr) {
			resp.DBInstances = append(resp.DBInstances, dbInstance)
			break
		}
	}
	return resp
}

func handleKmsKey(inKey *string, dbKey *string) *string {
	if inKey != nil && dbKey != nil && !strings.Contains(*inKey, "/") {
		lastInd := strings.LastIndex(*dbKey, "/")
		keyID := (*dbKey)[lastInd+1:]
		return &keyID
	}
	return dbKey
}

// isStorageTypeGP3BelowAllocatedStorageThreshold returns true if storageType is gp3 and allocatedStorage is below engine specific threshold
// See also https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#gp3-storage.
func isStorageTypeGP3BelowAllocatedStorageThreshold(cr *svcapitypes.DBInstance) bool {
	if pointer.StringValue(cr.Spec.ForProvider.StorageType) != "gp3" {
		return false
	}

	switch allocatedStorage, engine := pointer.Int64Value(cr.Spec.ForProvider.AllocatedStorage), pointer.StringValue(cr.Spec.ForProvider.Engine); engine {
	case "mariadb", "mysql", "postgres":
		return allocatedStorage < 400
	case "oracle-ee", "oracle-ee-cdb", "oracle-se2", "oracle-se2-cdb":
		return allocatedStorage < 200
	}

	return false
}
