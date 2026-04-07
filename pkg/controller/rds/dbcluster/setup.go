package dbcluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

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
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	dbinstance "github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// error constants
const (
	errUpdateTags               = "cannot update tags"
	errRestore                  = "cannot restore DBCluster in AWS"
	errUnknownRestoreFromSource = "unknown restoreFrom source"
)

type shared struct {
	kube   client.Client
	client svcsdkapi.RDSAPI
	cache  *cache
}

type cache struct {
	addTags                            []*svcsdk.Tag
	removeTags                         []*string
	desiredPassword                    string
	passwordChanged                    bool
	engineVersionChanged               bool
	dbClusterParameterGroupNameChanged bool
	backupRetentionPeriodChanged       bool
	backupWindowChanged                bool
}

func setupExternal(e *external) {
	s := &shared{client: e.client, kube: e.kube, cache: &cache{}}
	e.preObserve = preObserve
	e.postObserve = s.postObserve
	e.isUpToDate = s.isUpToDate
	e.preUpdate = s.preUpdate
	e.postUpdate = s.postUpdate
	e.preCreate = s.preCreate
	e.postCreate = s.postCreate
	e.preDelete = preDelete
	e.filterList = filterList
}

// SetupDBCluster adds a controller that reconciles DbCluster.
func SetupDBCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBClusterGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
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
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.DBCluster{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

// This probably requires shared Conditions to be defined for handling all statuses
// described here https://docs.pointer.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Status.html
// Need to get help from community on how to deal with this. Ideally the status should reflect
// the true status value as described by the provider.
func (s *shared) postObserve(ctx context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider.KMSKeyID = resp.DBClusters[0].KmsKeyId
	cr.Status.AtProvider.Port = resp.DBClusters[0].Port

	switch pointer.StringValue(resp.DBClusters[0].Status) {
	case "available", "storage-optimization", "backing-up":
		cr.SetConditions(xpv1.Available())
	case "modifying":
		cr.SetConditions(xpv1.Available().WithMessage("DB Cluster is " + pointer.StringValue(resp.DBClusters[0].Status) + ", availability may vary"))
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	default:
		cr.SetConditions(xpv1.Unavailable().WithMessage("DB Cluster is " + pointer.StringValue(resp.DBClusters[0].Status)))
	}
	obs.ConnectionDetails, err = s.updateConnectionDetails(ctx, cr, obs.ConnectionDetails)

	return obs, err
}

func (s *shared) preCreate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) (err error) { //nolint:gocyclo
	restoreFrom := cr.Spec.ForProvider.RestoreFrom
	autogenerate := cr.Spec.ForProvider.AutogeneratePassword
	masterUserPasswordSecretRef := cr.Spec.ForProvider.MasterUserPasswordSecretRef

	var pw string
	switch {
	case masterUserPasswordSecretRef == nil && !autogenerate && restoreFrom == nil:
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
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
	for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
		obj.VpcSecurityGroupIds[i] = pointer.ToOrNilIfZeroValue(v)
	}

	passwordRestoreInfo := map[string]string{dbinstance.PasswordCacheKey: pw}
	if restoreFrom != nil {
		passwordRestoreInfo[dbinstance.RestoreFlagCacheKay] = string(dbinstance.RestoreStateRestored)

		switch *restoreFrom.Source {
		case "S3":
			input := generateRestoreDBClusterFromS3Input(cr)
			input.MasterUserPassword = obj.MasterUserPassword
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = s.client.RestoreDBClusterFromS3WithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		case "Snapshot":
			input := generateRestoreDBClusterFromSnapshotInput(cr)
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = s.client.RestoreDBClusterFromSnapshotWithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		case "PointInTime":
			input := generateRestoreDBClusterToPointInTimeInput(cr)
			input.DBClusterIdentifier = obj.DBClusterIdentifier
			input.VpcSecurityGroupIds = obj.VpcSecurityGroupIds

			if _, err = s.client.RestoreDBClusterToPointInTimeWithContext(ctx, input); err != nil {
				return errors.Wrap(err, errRestore)
			}
		default:
			return errors.New(errUnknownRestoreFromSource)
		}
	}

	obj.EngineVersion = cr.Spec.ForProvider.EngineVersion

	if _, err = dbinstance.Cache(ctx, s.kube, cr, passwordRestoreInfo); err != nil {
		return errors.Wrap(err, dbinstance.ErrCachePassword)
	}
	return nil
}

func (s *shared) postCreate(ctx context.Context, cr *svcapitypes.DBCluster, _ *svcsdk.CreateDBClusterOutput, extCreation managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	cd, err := s.updateConnectionDetails(ctx, cr, managed.ConnectionDetails{})
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	extCreation.ConnectionDetails = cd
	return extCreation, nil
}

func generateRestoreDBClusterFromS3Input(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterFromS3Input { //nolint:gocyclo
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

	if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration != nil {
		serverlessScalingConfiguration := &svcsdk.ServerlessV2ScalingConfiguration{}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity != nil {
			serverlessScalingConfiguration.SetMaxCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity)
		}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity != nil {
			serverlessScalingConfiguration.SetMinCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity)
		}
		res.SetServerlessV2ScalingConfiguration(serverlessScalingConfiguration)
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

func generateRestoreDBClusterFromSnapshotInput(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterFromSnapshotInput { //nolint:gocyclo
	res := &svcsdk.RestoreDBClusterFromSnapshotInput{}

	if cr.Spec.ForProvider.AvailabilityZones != nil {
		res.SetAvailabilityZones(cr.Spec.ForProvider.AvailabilityZones)
	}

	if cr.Spec.ForProvider.BacktrackWindow != nil {
		res.SetBacktrackWindow(*cr.Spec.ForProvider.BacktrackWindow)
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

	if cr.Spec.ForProvider.EngineMode != nil {
		res.SetEngineMode(*cr.Spec.ForProvider.EngineMode)
	}

	if cr.Spec.ForProvider.EngineVersion != nil {
		res.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}

	if cr.Spec.ForProvider.IOPS != nil {
		res.SetIops(*cr.Spec.ForProvider.IOPS)
	}

	if cr.Spec.ForProvider.KMSKeyID != nil {
		res.SetKmsKeyId(*cr.Spec.ForProvider.KMSKeyID)
	}

	if cr.Spec.ForProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Spec.ForProvider.OptionGroupName)
	}

	if cr.Spec.ForProvider.Port != nil {
		res.SetPort(*cr.Spec.ForProvider.Port)
	}

	if cr.Spec.ForProvider.PubliclyAccessible != nil {
		res.SetPubliclyAccessible(*cr.Spec.ForProvider.PubliclyAccessible)
	}

	if cr.Spec.ForProvider.ScalingConfiguration != nil {
		scalingConfiguration := &svcsdk.ScalingConfiguration{}
		if cr.Spec.ForProvider.ScalingConfiguration.AutoPause != nil {
			scalingConfiguration.SetAutoPause(*cr.Spec.ForProvider.ScalingConfiguration.AutoPause)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.MaxCapacity != nil {
			scalingConfiguration.SetMaxCapacity(*cr.Spec.ForProvider.ScalingConfiguration.MaxCapacity)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.MinCapacity != nil {
			scalingConfiguration.SetMinCapacity(*cr.Spec.ForProvider.ScalingConfiguration.MinCapacity)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.SecondsBeforeTimeout != nil {
			scalingConfiguration.SetSecondsBeforeTimeout(*cr.Spec.ForProvider.ScalingConfiguration.SecondsBeforeTimeout)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.SecondsUntilAutoPause != nil {
			scalingConfiguration.SetSecondsUntilAutoPause(*cr.Spec.ForProvider.ScalingConfiguration.SecondsUntilAutoPause)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.TimeoutAction != nil {
			scalingConfiguration.SetTimeoutAction(*cr.Spec.ForProvider.ScalingConfiguration.TimeoutAction)
		}
		res.SetScalingConfiguration(scalingConfiguration)
	}

	if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration != nil {
		serverlessScalingConfiguration := &svcsdk.ServerlessV2ScalingConfiguration{}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity != nil {
			serverlessScalingConfiguration.SetMaxCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity)
		}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity != nil {
			serverlessScalingConfiguration.SetMinCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity)
		}
		res.SetServerlessV2ScalingConfiguration(serverlessScalingConfiguration)
	}

	if cr.Spec.ForProvider.RestoreFrom != nil && cr.Spec.ForProvider.RestoreFrom.Snapshot != nil {
		res.SetSnapshotIdentifier(*cr.Spec.ForProvider.RestoreFrom.Snapshot.SnapshotIdentifier)
	}

	if cr.Spec.ForProvider.StorageType != nil {
		res.SetStorageType(*cr.Spec.ForProvider.StorageType)
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

func generateRestoreDBClusterToPointInTimeInput(cr *svcapitypes.DBCluster) *svcsdk.RestoreDBClusterToPointInTimeInput { //nolint:gocyclo

	p := cr.Spec.ForProvider
	res := &svcsdk.RestoreDBClusterToPointInTimeInput{
		BacktrackWindow:                 p.BacktrackWindow,
		CopyTagsToSnapshot:              p.CopyTagsToSnapshot,
		DBClusterInstanceClass:          p.DBClusterInstanceClass,
		DBClusterParameterGroupName:     p.DBClusterParameterGroupName,
		DBSubnetGroupName:               p.DBSubnetGroupName,
		DeletionProtection:              p.DeletionProtection,
		Domain:                          p.Domain,
		DomainIAMRoleName:               p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:     p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication: p.EnableIAMDatabaseAuthentication,
		EngineMode:                      p.EngineMode,
		Iops:                            p.IOPS,
		KmsKeyId:                        p.KMSKeyID,
		OptionGroupName:                 p.OptionGroupName,
		Port:                            p.Port,
		PubliclyAccessible:              p.PubliclyAccessible,
		StorageType:                     p.StorageType,
		UseLatestRestorableTime:         &p.RestoreFrom.PointInTime.UseLatestRestorableTime,
	}
	if p.RestoreFrom.PointInTime != nil && p.RestoreFrom.PointInTime.RestoreTime != nil {
		res.RestoreToTime = &p.RestoreFrom.PointInTime.RestoreTime.Time
	}
	if p.RestoreFrom.PointInTime != nil && p.RestoreFrom.PointInTime.RestoreType != nil {
		res.RestoreType = p.RestoreFrom.PointInTime.RestoreType
	}
	if p.RestoreFrom.PointInTime != nil && p.RestoreFrom.PointInTime.SourceDBClusterIdentifier != nil {
		res.SourceDBClusterIdentifier = p.RestoreFrom.PointInTime.SourceDBClusterIdentifier
	}
	if p.RestoreFrom.PointInTime != nil && p.RestoreFrom.PointInTime.SourceDBClusterResourceID != nil {
		res.SourceDbClusterResourceId = p.RestoreFrom.PointInTime.SourceDBClusterResourceID
	}
	if cr.Spec.ForProvider.Tags != nil {
		var tags []*svcsdk.Tag
		for _, tag := range cr.Spec.ForProvider.Tags {
			tags = append(tags, &svcsdk.Tag{Key: tag.Key, Value: tag.Value})
		}

		res.SetTags(tags)
	}

	if cr.Spec.ForProvider.ScalingConfiguration != nil {
		scalingConfiguration := &svcsdk.ScalingConfiguration{}
		if cr.Spec.ForProvider.ScalingConfiguration.AutoPause != nil {
			scalingConfiguration.SetAutoPause(*cr.Spec.ForProvider.ScalingConfiguration.AutoPause)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.MaxCapacity != nil {
			scalingConfiguration.SetMaxCapacity(*cr.Spec.ForProvider.ScalingConfiguration.MaxCapacity)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.MinCapacity != nil {
			scalingConfiguration.SetMinCapacity(*cr.Spec.ForProvider.ScalingConfiguration.MinCapacity)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.SecondsBeforeTimeout != nil {
			scalingConfiguration.SetSecondsBeforeTimeout(*cr.Spec.ForProvider.ScalingConfiguration.SecondsBeforeTimeout)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.SecondsUntilAutoPause != nil {
			scalingConfiguration.SetSecondsUntilAutoPause(*cr.Spec.ForProvider.ScalingConfiguration.SecondsUntilAutoPause)
		}
		if cr.Spec.ForProvider.ScalingConfiguration.TimeoutAction != nil {
			scalingConfiguration.SetTimeoutAction(*cr.Spec.ForProvider.ScalingConfiguration.TimeoutAction)
		}
		res.SetScalingConfiguration(scalingConfiguration)
	}

	if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration != nil {
		serverlessScalingConfiguration := &svcsdk.ServerlessV2ScalingConfiguration{}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity != nil {
			serverlessScalingConfiguration.SetMaxCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MaxCapacity)
		}
		if cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity != nil {
			serverlessScalingConfiguration.SetMinCapacity(*cr.Spec.ForProvider.ServerlessV2ScalingConfiguration.MinCapacity)
		}
		res.SetServerlessV2ScalingConfiguration(serverlessScalingConfiguration)
	}

	return res
}

func (s *shared) isUpToDate(ctx context.Context, cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) (bool, string, error) { //nolint:gocyclo
	// If ApplyImmediately is not true we update observed state of db cluster with pending modified values to prevent redundant updates
	if !ptr.Deref(cr.Spec.ForProvider.ApplyImmediately, false) {
		utils.SetPmvDBCluster(out)
	}
	observed := GenerateDBCluster(out)

	status := pointer.StringValue(out.DBClusters[0].Status)
	if status == "modifying" || status == "upgrading" || status == "configuring-iam-database-auth" || status == "migrating" || status == "prepairing-data-migration" || status == "creating" {
		return true, "", nil
	}

	passwordUpToDate, desiredPassword, err := dbinstance.PasswordUpToDate(ctx, s.kube, cr)
	if err != nil {
		return false, "", errors.Wrap(err, dbinstance.ErrNoPasswordUpToDate)
	}
	s.cache.desiredPassword = desiredPassword
	s.cache.passwordChanged = !passwordUpToDate

	patch, err := createPatch(&observed.Spec.ForProvider, &cr.Spec.ForProvider)
	if err != nil {
		return false, "", err
	}
	diff := cmp.Diff(&svcapitypes.DBClusterParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "AllowMajorVersionUpgrade"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "BacktrackWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "BackupRetentionPeriod"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "DBSubnetGroupName"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "DBClusterParameterGroupName"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "EnableCloudwatchLogsExports"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "EngineVersion"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "MasterUserPasswordSecretRef"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "ScalingConfiguration"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "SkipFinalSnapshot"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "StorageType"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "OptionGroupName"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "PreferredBackupWindow"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.DBClusterParameters{}, "Tags"),
		cmpopts.IgnoreTypes(svcapitypes.CustomDBClusterParameters{}),
	)

	if s.cache.passwordChanged {
		diff += "\nmaster user password changed"
	}

	s.cache.backupWindowChanged = !isPreferredBackupWindowUpToDate(cr, out)
	if s.cache.backupWindowChanged {
		diff += "\ndesired preferredBackupWindow: " + pointer.StringValue(cr.Spec.ForProvider.PreferredBackupWindow) +
			", observed preferredBackupWindow: " + pointer.StringValue(out.DBClusters[0].PreferredBackupWindow)
	}

	if pointer.Int64Value(cr.Spec.ForProvider.BacktrackWindow) != pointer.Int64Value(out.DBClusters[0].BacktrackWindow) {
		return false, "", nil
	}

	if !isStorageTypeUpToDate(cr, out) {
		diff += "\ndesired storageType: " + pointer.StringValue(cr.Spec.ForProvider.StorageType) +
			", observed storageType: " + pointer.StringValue(out.DBClusters[0].StorageType)
	}

	s.cache.backupRetentionPeriodChanged = !isBackupRetentionPeriodUpToDate(cr, out)
	if s.cache.backupRetentionPeriodChanged {
		diff += "\ndesired backupRetentionPeriod: " + strconv.FormatInt(pointer.Int64Value(cr.Spec.ForProvider.BackupRetentionPeriod), 10) +
			", observed backupRetentionPeriod: " + strconv.FormatInt(pointer.Int64Value(out.DBClusters[0].BackupRetentionPeriod), 10)
	}

	if !isPortUpToDate(cr, out) {
		diff += "\ndesired port: " + strconv.FormatInt(pointer.Int64Value(cr.Spec.ForProvider.Port), 10) +
			", observed port: " + strconv.FormatInt(pointer.Int64Value(out.DBClusters[0].Port), 10)
	}

	s.cache.engineVersionChanged = !isEngineVersionUpToDate(cr, out)
	if s.cache.engineVersionChanged {
		if ptr.Deref(cr.Spec.ForProvider.EngineVersion, "") == ptr.Deref(out.DBClusters[0].EngineVersion, "") && out.DBClusters[0].PendingModifiedValues != nil && ptr.Deref(out.DBClusters[0].PendingModifiedValues.EngineVersion, "") != "" {
			diff += fmt.Sprintf("\ndesired engineVersion: %s \npending modified engineVersion: %s ", pointer.StringValue(cr.Spec.ForProvider.EngineVersion), pointer.StringValue(out.DBClusters[0].PendingModifiedValues.EngineVersion))
		} else {
			diff += fmt.Sprintf("\ndesired engineVersion: %s \nobserved engineVersion: %s ", pointer.StringValue(cr.Spec.ForProvider.EngineVersion), pointer.StringValue(out.DBClusters[0].EngineVersion))
		}
	}

	if !areVPCSecurityGroupIDsUpToDate(cr, out) {
		observedIDs := make([]string, 0, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
		for _, grp := range out.DBClusters[0].VpcSecurityGroups {
			observedIDs = append(observedIDs, *grp.VpcSecurityGroupId)
		}
		diff += "\ndesired vpcSecurityGroupIDs: " + strings.Join(cr.Spec.ForProvider.VPCSecurityGroupIDs, ",") +
			", observed vpcSecurityGroupIDs: " + strings.Join(observedIDs, ",")
	}

	if !areSameElements(cr.Spec.ForProvider.EnableCloudwatchLogsExports, out.DBClusters[0].EnabledCloudwatchLogsExports) {
		diff += "\n enabledCloudwatchLogsExports changed"
	}

	s.cache.dbClusterParameterGroupNameChanged = !isDBClusterParameterGroupNameUpToDate(cr, out)
	if s.cache.dbClusterParameterGroupNameChanged {
		diff += "\ndesired dbClusterParameterGroupName: " + pointer.StringValue(cr.Spec.ForProvider.DBClusterParameterGroupName) +
			", observed dbClusterParameterGroupName: " + pointer.StringValue(out.DBClusters[0].DBClusterParameterGroup)
	}

	isScalingConfigurationUpToDate, err := isScalingConfigurationUpToDate(cr.Spec.ForProvider.ScalingConfiguration, out.DBClusters[0].ScalingConfigurationInfo)
	if err != nil {
		return false, "", errors.Wrap(err, "failed to compare scaling configuration")
	}
	if !isScalingConfigurationUpToDate {
		diff += "\nscalingConfiguration changed"
	}

	s.cache.addTags, s.cache.removeTags = utils.DiffTags(cr.Spec.ForProvider.Tags, out.DBClusters[0].TagList)
	tagsChanged := len(s.cache.addTags) != 0 || len(s.cache.removeTags) != 0
	if tagsChanged {
		diff += fmt.Sprintf("\nadd %d tag(s) and remove %d tag(s)", len(s.cache.addTags), len(s.cache.removeTags))
	}

	if diff != "" {
		diff = "Found observed differences in dbCluster\n" + diff
		log.Println(diff)
		return false, diff, nil
	}

	return true, "", nil
}

func isStorageTypeUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If StorageType is not set by user, we do not need to check and accept AWS set default (or the current value)
	if cr.Spec.ForProvider.StorageType == nil {
		return true
	}

	// AWS returns "" when StorageType is explicitly "aurora" (default for DBClusters with aurora engines)
	// see also note in AWS docs: https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBCluster.html
	if pointer.StringValue(cr.Spec.ForProvider.StorageType) == "aurora" &&
		pointer.StringValue(out.DBClusters[0].StorageType) == "" {
		return true
	}

	return pointer.StringValue(cr.Spec.ForProvider.StorageType) == pointer.StringValue(out.DBClusters[0].StorageType)
}

func isPreferredMaintenanceWindowUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If PreferredMaintenanceWindow is not set, aws sets a random window
	// so we do not try to update in this case
	if cr.Spec.ForProvider.PreferredMaintenanceWindow != nil {

		// AWS accepts uppercase weekdays, but returns lowercase values,
		// therfore we compare usinf equalFold
		if !strings.EqualFold(pointer.StringValue(cr.Spec.ForProvider.PreferredMaintenanceWindow), pointer.StringValue(out.DBClusters[0].PreferredMaintenanceWindow)) {
			return false
		}
	}
	return true
}

func isPreferredBackupWindowUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// 1. If PreferredBackupWindow is not set, AWS sets a random window
	// so we do not try to update in this case
	// 2. AWS Backup takes ownership of PreferredBackupWindow if it is in use.
	// So we only check PreferredBackupWindow, if there is no associated RecoveryPoint.
	if cr.Spec.ForProvider.PreferredBackupWindow != nil &&
		out.DBClusters[0].AwsBackupRecoveryPointArn == nil {
		if pointer.StringValue(cr.Spec.ForProvider.PreferredBackupWindow) != pointer.StringValue(out.DBClusters[0].PreferredBackupWindow) {
			return false
		}
	}
	return true
}

func isBackupRetentionPeriodUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// 1. If BackupRetentionPeriod is not set, AWS sets a default value
	// so we do not try to update in this case
	// 2. AWS Backup takes ownership of BackupRetentionPeriod if it is in use.
	// So we only check BackupRetentionPeriod, if there is no associated RecoveryPoint.
	if cr.Spec.ForProvider.BackupRetentionPeriod != nil &&
		out.DBClusters[0].AwsBackupRecoveryPointArn == nil {
		if pointer.Int64Value(cr.Spec.ForProvider.BackupRetentionPeriod) != pointer.Int64Value(out.DBClusters[0].BackupRetentionPeriod) {
			return false
		}
	}
	return true
}

func isScalingConfigurationUpToDate(sc *svcapitypes.ScalingConfiguration, obj *svcsdk.ScalingConfigurationInfo) (bool, error) {
	jsonPatch, err := jsonpatch.CreateJSONPatch(sc, obj)
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
	// If EngineVersion is not set, AWS sets a default value,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.EngineVersion != nil {
		if out.DBClusters[0].EngineVersion == nil {
			return false
		}

		desiredEngineVersion := ptr.Deref(cr.Spec.ForProvider.EngineVersion, "")
		observedEngineVersion := ptr.Deref(out.DBClusters[0].EngineVersion, "")
		pendingEngineVersion := ""
		if out.DBClusters[0].PendingModifiedValues != nil {
			pendingEngineVersion = ptr.Deref(out.DBClusters[0].PendingModifiedValues.EngineVersion, "")
		}
		// If the desired version matches the current version and pending version is set means an upgrade should be reverted
		// so controller should send an update request with the current version again to cancel the pending upgrade
		if desiredEngineVersion == observedEngineVersion && pendingEngineVersion != "" {
			return false
		}
		// If ApplyImmediately is false and there is a pending version, consider that as the current version for comparison
		if !ptr.Deref(cr.Spec.ForProvider.ApplyImmediately, false) && pendingEngineVersion != "" {
			observedEngineVersion = pendingEngineVersion
		}
		// Upgrade is only necessary if the spec version is higher.
		// Downgrades are not possible in pointer.
		c := utils.CompareEngineVersions(*cr.Spec.ForProvider.EngineVersion, observedEngineVersion)
		return c <= 0
	}
	return true
}

func isDBClusterParameterGroupNameUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If DBClusterParameterGroupName is not set, AWS sets a default value,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.DBClusterParameterGroupName != nil {
		return pointer.StringValue(cr.Spec.ForProvider.DBClusterParameterGroupName) == pointer.StringValue(out.DBClusters[0].DBClusterParameterGroup)
	}
	return true
}

func isPortUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// If Port is not set, aws sets a default value
	// so we do not try to update in this case
	if cr.Spec.ForProvider.Port != nil {
		if pointer.Int64Value(cr.Spec.ForProvider.Port) != pointer.Int64Value(out.DBClusters[0].Port) {
			return false
		}
	}
	return true
}

func areVPCSecurityGroupIDsUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) bool {
	// AWS uses the default SG which is really restrictive, and it seems to use it even when it is
	// patched (with "required") - might be race condition. Anyway with checking if there is a diff
	// we can rectify and even make it configurable after creation.

	desiredIDs := cr.Spec.ForProvider.VPCSecurityGroupIDs

	// if user is fine with default SG (removing all SGs is not possible, AWS will keep last set SGs)
	if len(desiredIDs) == 0 {
		return true
	}

	actualGroups := out.DBClusters[0].VpcSecurityGroups

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

func (s *shared) preUpdate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterInput) error {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately

	obj.CloudwatchLogsExportConfiguration = generateCloudWatchExportConfiguration(
		cr.Spec.ForProvider.EnableCloudwatchLogsExports,
		cr.Status.AtProvider.EnabledCloudwatchLogsExports)
	// Only set MasterUserPassword if it has changed, otherwise it triggers "Resetting master password" on aws side,
	// it happens because aws doesn't know current password, so any set causes a change.
	if s.cache.passwordChanged {
		obj.MasterUserPassword = pointer.ToOrNilIfZeroValue(s.cache.desiredPassword)
	}

	if cr.Spec.ForProvider.VPCSecurityGroupIDs != nil {
		obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
		for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
			obj.VpcSecurityGroupIds[i] = pointer.ToOrNilIfZeroValue(v)
		}
	}

	// ModifyDBCluster() returns error, when trying to upgrade major (minor is fine) EngineVersion:
	// "Cannot change VPC security group while doing a major version upgrade."
	// even when the provided VPCSecurityGroupIDs are upToDate...
	// therefore EngineVersion update is entirely done separately in postUpdate
	// Note: strangely ModifyDBInstance does not seem to behave this way
	obj.EngineVersion = nil
	// In case of a shared DBClusterParameterGroup, AWS requires for a major version update that
	// EngineVersion and DBClusterParameterGroupName are in the same ModifyDBCluster()-call
	obj.DBClusterParameterGroupName = nil

	// AWS Backup takes ownership of BackupRetentionPeriod and PreferredBackupWindow if it is in use.
	// So we need to check there is no associated RecoveryPoint, before trying to update these fields.
	// Therefore we only set these fields in the ModifyDBInstanceInput if they are changed and not in use by AWS Backup.
	obj.PreferredBackupWindow = nil
	obj.BackupRetentionPeriod = nil

	return nil
}

//nolint:gocyclo
func (s *shared) postUpdate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	if s.cache.passwordChanged {
		_, err = dbinstance.Cache(ctx, s.kube, cr, map[string]string{
			dbinstance.PasswordCacheKey:    s.cache.desiredPassword,
			dbinstance.RestoreFlagCacheKay: "", // reset restore flag
		})
		if err != nil {
			return upd, errors.Wrap(err, dbinstance.ErrCachePassword)
		}
	}

	input := GenerateDescribeDBClustersInput(cr)
	// GenerateDescribeDBClustersInput returns an empty DescribeDBClustersInput
	// and the function is generated by ack-generate, so we manually need to set the
	// DBClusterIdentifier
	input.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	resp, err := s.client.DescribeDBClustersWithContext(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}

	needsPostUpdate := s.cache.engineVersionChanged || s.cache.dbClusterParameterGroupNameChanged || s.cache.backupWindowChanged || s.cache.backupRetentionPeriodChanged

	if needsPostUpdate {
		modifyInput := &svcsdk.ModifyDBClusterInput{
			DBClusterIdentifier:         pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			ApplyImmediately:            cr.Spec.ForProvider.ApplyImmediately,
			DBClusterParameterGroupName: cr.Spec.ForProvider.DBClusterParameterGroupName,
		}
		if s.cache.engineVersionChanged {
			modifyInput.EngineVersion = cr.Spec.ForProvider.EngineVersion
			modifyInput.AllowMajorVersionUpgrade = cr.Spec.ForProvider.AllowMajorVersionUpgrade
		}
		if s.cache.backupWindowChanged {
			modifyInput.PreferredBackupWindow = cr.Spec.ForProvider.PreferredBackupWindow
		}
		if s.cache.backupRetentionPeriodChanged {
			modifyInput.BackupRetentionPeriod = cr.Spec.ForProvider.BackupRetentionPeriod
		}

		if _, err = s.client.ModifyDBClusterWithContext(ctx, modifyInput); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if len(s.cache.addTags) > 0 || len(s.cache.removeTags) > 0 {
		err := s.updateTags(ctx, cr, s.cache.addTags, s.cache.removeTags)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if !isPreferredMaintenanceWindowUpToDate(cr, resp) {
		return upd, errors.New("PreferredMaintenanceWindow not matching aws data")
	}

	return upd, err
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) (bool, error) {
	obj.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.SkipFinalSnapshot = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.SkipFinalSnapshot)
	obj.DeleteAutomatedBackups = cr.Spec.ForProvider.DeleteAutomatedBackups

	if !cr.Spec.ForProvider.SkipFinalSnapshot {
		obj.FinalDBSnapshotIdentifier = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.FinalDBSnapshotIdentifier)
	}

	return false, nil
}

func (s *shared) postDelete(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterOutput, err error) (managed.ExternalDelete, error) {
	if err != nil {
		return managed.ExternalDelete{}, err
	}

	return managed.ExternalDelete{}, dbinstance.DeleteCache(ctx, s.kube, cr)
}

func filterList(cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersOutput) *svcsdk.DescribeDBClustersOutput {
	clusterIdentifier := pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeDBClustersOutput{}
	for _, dbCluster := range obj.DBClusters {
		if pointer.StringValue(dbCluster.DBClusterIdentifier) == pointer.StringValue(clusterIdentifier) {
			resp.DBClusters = append(resp.DBClusters, dbCluster)
			break
		}
	}
	return resp
}

func (s *shared) updateTags(ctx context.Context, cr *svcapitypes.DBCluster, addTags []*svcsdk.Tag, removeTags []*string) error {

	arn := cr.Status.AtProvider.DBClusterARN
	if arn != nil {
		if len(removeTags) > 0 {
			inputR := &svcsdk.RemoveTagsFromResourceInput{
				ResourceName: arn,
				TagKeys:      removeTags,
			}

			_, err := s.client.RemoveTagsFromResourceWithContext(ctx, inputR)
			if err != nil {
				return errors.New(errUpdateTags)
			}
		}
		if len(addTags) > 0 {
			inputC := &svcsdk.AddTagsToResourceInput{
				ResourceName: arn,
				Tags:         addTags,
			}

			_, err := s.client.AddTagsToResourceWithContext(ctx, inputC)
			if err != nil {
				return errors.New(errUpdateTags)
			}

		}
	}
	return nil

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
func (s *shared) updateConnectionDetails(ctx context.Context, cr *svcapitypes.DBCluster, details managed.ConnectionDetails) (managed.ConnectionDetails, error) {
	if details == nil {
		details = managed.ConnectionDetails{}
	}

	details[xpv1.ResourceCredentialsSecretUserKey] = []byte(pointer.StringValue(cr.Spec.ForProvider.MasterUsername))
	password := s.cache.desiredPassword
	if password == "" {
		pw, err := dbinstance.GetDesiredPassword(ctx, s.kube, cr)
		if err != nil {
			return details, errors.Wrap(err, dbinstance.ErrGetCachedPassword)
		}
		password = pw
	}
	details[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(password)

	if cr.Status.AtProvider.Endpoint == nil {
		return details, nil
	}
	if pointer.StringValue(cr.Status.AtProvider.Endpoint) != "" {
		details[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(pointer.StringValue(cr.Status.AtProvider.Endpoint))
	}
	if pointer.Int64Value(cr.Status.AtProvider.Port) > 0 {
		details[xpv1.ResourceCredentialsSecretPortKey] = []byte(strconv.FormatInt(*cr.Status.AtProvider.Port, 10))
	}
	if pointer.StringValue(cr.Status.AtProvider.ReaderEndpoint) != "" {
		details["readerEndpoint"] = []byte(pointer.StringValue(cr.Status.AtProvider.ReaderEndpoint))
	}

	return details, nil
}

func createPatch(observed *svcapitypes.DBClusterParameters, desired *svcapitypes.DBClusterParameters) (*svcapitypes.DBClusterParameters, error) {
	jsonPatch, err := jsonpatch.CreateJSONPatch(observed, desired)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.DBClusterParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}
