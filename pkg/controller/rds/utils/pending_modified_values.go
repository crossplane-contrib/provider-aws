package utils

import (
	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// SetPmvDBInstance updates the DescribeDBInstancesOutput with any
// PendingModifiedValues so that they are considered during isUpToDate checks. Exception is Engine version, which is handled separately.
func SetPmvDBInstance(obs *svcsdk.DescribeDBInstancesOutput) { //nolint:gocyclo
	if len(obs.DBInstances) > 0 {
		if obs.DBInstances[0].PendingModifiedValues != nil {
			if obs.DBInstances[0].PendingModifiedValues.AllocatedStorage != nil {
				obs.DBInstances[0].AllocatedStorage = obs.DBInstances[0].PendingModifiedValues.AllocatedStorage
			}
			if obs.DBInstances[0].PendingModifiedValues.AutomationMode != nil {
				obs.DBInstances[0].AutomationMode = obs.DBInstances[0].PendingModifiedValues.AutomationMode
			}
			if obs.DBInstances[0].PendingModifiedValues.BackupRetentionPeriod != nil {
				obs.DBInstances[0].BackupRetentionPeriod = obs.DBInstances[0].PendingModifiedValues.BackupRetentionPeriod
			}
			if obs.DBInstances[0].PendingModifiedValues.CACertificateIdentifier != nil {
				obs.DBInstances[0].CACertificateIdentifier = obs.DBInstances[0].PendingModifiedValues.CACertificateIdentifier
			}
			if obs.DBInstances[0].PendingModifiedValues.DBInstanceClass != nil {
				obs.DBInstances[0].DBInstanceClass = obs.DBInstances[0].PendingModifiedValues.DBInstanceClass
			}
			if obs.DBInstances[0].PendingModifiedValues.DBSubnetGroupName != nil {
				obs.DBInstances[0].DBSubnetGroup = &svcsdk.DBSubnetGroup{
					DBSubnetGroupName: obs.DBInstances[0].PendingModifiedValues.DBSubnetGroupName,
				}
			}
			if obs.DBInstances[0].PendingModifiedValues.DedicatedLogVolume != nil {
				obs.DBInstances[0].DedicatedLogVolume = obs.DBInstances[0].PendingModifiedValues.DedicatedLogVolume
			}
			if obs.DBInstances[0].PendingModifiedValues.Iops != nil {
				obs.DBInstances[0].Iops = obs.DBInstances[0].PendingModifiedValues.Iops
			}
			if obs.DBInstances[0].PendingModifiedValues.LicenseModel != nil {
				obs.DBInstances[0].LicenseModel = obs.DBInstances[0].PendingModifiedValues.LicenseModel
			}
			if obs.DBInstances[0].PendingModifiedValues.MultiAZ != nil {
				obs.DBInstances[0].MultiAZ = obs.DBInstances[0].PendingModifiedValues.MultiAZ
			}
			if obs.DBInstances[0].PendingModifiedValues.PendingCloudwatchLogsExports != nil {
				applyInstancePendingCloudwatchLogsExports(obs)
			}
			if obs.DBInstances[0].PendingModifiedValues.Port != nil {
				if obs.DBInstances[0].Endpoint == nil {
					obs.DBInstances[0].Endpoint = &svcsdk.Endpoint{}
				}
				obs.DBInstances[0].Endpoint.Port = obs.DBInstances[0].PendingModifiedValues.Port
			}
			if obs.DBInstances[0].PendingModifiedValues.ProcessorFeatures != nil {
				obs.DBInstances[0].ProcessorFeatures = obs.DBInstances[0].PendingModifiedValues.ProcessorFeatures
			}
			if obs.DBInstances[0].PendingModifiedValues.StorageThroughput != nil {
				obs.DBInstances[0].StorageThroughput = obs.DBInstances[0].PendingModifiedValues.StorageThroughput
			}
			if obs.DBInstances[0].PendingModifiedValues.StorageType != nil {
				obs.DBInstances[0].StorageType = obs.DBInstances[0].PendingModifiedValues.StorageType
			}
		}
	}
}

// SetPmvDBCluster updates the DescribeDBClustersOutput with any
// PendingModifiedValues so that they are considered during isUpToDate checks. Exception is Engine version, which is handled separately.
func SetPmvDBCluster(obs *svcsdk.DescribeDBClustersOutput) {
	if len(obs.DBClusters) > 0 {
		if obs.DBClusters[0].PendingModifiedValues != nil {
			if obs.DBClusters[0].PendingModifiedValues.AllocatedStorage != nil {
				obs.DBClusters[0].AllocatedStorage = obs.DBClusters[0].PendingModifiedValues.AllocatedStorage
			}
			if obs.DBClusters[0].PendingModifiedValues.BackupRetentionPeriod != nil {
				obs.DBClusters[0].BackupRetentionPeriod = obs.DBClusters[0].PendingModifiedValues.BackupRetentionPeriod
			}
			if obs.DBClusters[0].PendingModifiedValues.IAMDatabaseAuthenticationEnabled != nil {
				obs.DBClusters[0].IAMDatabaseAuthenticationEnabled = obs.DBClusters[0].PendingModifiedValues.IAMDatabaseAuthenticationEnabled
			}
			if obs.DBClusters[0].PendingModifiedValues.Iops != nil {
				obs.DBClusters[0].Iops = obs.DBClusters[0].PendingModifiedValues.Iops
			}
			if obs.DBClusters[0].PendingModifiedValues.PendingCloudwatchLogsExports != nil {
				applyClusterPendingCloudwatchLogsExports(obs)
			}
			if obs.DBClusters[0].PendingModifiedValues.RdsCustomClusterConfiguration != nil {
				obs.DBClusters[0].RdsCustomClusterConfiguration = obs.DBClusters[0].PendingModifiedValues.RdsCustomClusterConfiguration
			}
			if obs.DBClusters[0].PendingModifiedValues.StorageType != nil {
				obs.DBClusters[0].StorageType = obs.DBClusters[0].PendingModifiedValues.StorageType
			}
		}
	}
}

func applyInstancePendingCloudwatchLogsExports(cr *svcsdk.DescribeDBInstancesOutput) { //nolint:gocyclo
	if len(cr.DBInstances) == 0 || cr.DBInstances[0].PendingModifiedValues == nil {
		return
	}

	pending := cr.DBInstances[0].PendingModifiedValues.PendingCloudwatchLogsExports
	if pending == nil {
		return
	}

	// Handle LogTypesToDisable
	if pending.LogTypesToDisable != nil {
		disableSet := make(map[string]struct{}, len(pending.LogTypesToDisable))
		for _, logType := range pending.LogTypesToDisable {
			disableSet[pointer.StringValue(logType)] = struct{}{}
		}

		filtered := make([]*string, 0)
		for _, enabled := range cr.DBInstances[0].EnabledCloudwatchLogsExports {
			if _, shouldDisable := disableSet[pointer.StringValue(enabled)]; !shouldDisable {
				filtered = append(filtered, enabled)
			}
		}
		cr.DBInstances[0].EnabledCloudwatchLogsExports = filtered
	}

	// Handle LogTypesToEnable
	if pending.LogTypesToEnable != nil {
		for _, toEnable := range pending.LogTypesToEnable {
			enableStr := pointer.StringValue(toEnable)
			found := false
			for _, enabled := range cr.DBInstances[0].EnabledCloudwatchLogsExports {
				if pointer.StringValue(enabled) == enableStr {
					found = true
					break
				}
			}
			if !found {
				cr.DBInstances[0].EnabledCloudwatchLogsExports = append(cr.DBInstances[0].EnabledCloudwatchLogsExports, toEnable)
			}
		}
	}
}

func applyClusterPendingCloudwatchLogsExports(cr *svcsdk.DescribeDBClustersOutput) { //nolint:gocyclo
	if len(cr.DBClusters) == 0 || cr.DBClusters[0].PendingModifiedValues == nil {
		return
	}

	pending := cr.DBClusters[0].PendingModifiedValues.PendingCloudwatchLogsExports
	if pending == nil {
		return
	}

	// Handle LogTypesToDisable
	if pending.LogTypesToDisable != nil {
		disableSet := make(map[string]struct{}, len(pending.LogTypesToDisable))
		for _, logType := range pending.LogTypesToDisable {
			disableSet[pointer.StringValue(logType)] = struct{}{}
		}

		filtered := make([]*string, 0)
		for _, enabled := range cr.DBClusters[0].EnabledCloudwatchLogsExports {
			if _, shouldDisable := disableSet[pointer.StringValue(enabled)]; !shouldDisable {
				filtered = append(filtered, enabled)
			}
		}
		cr.DBClusters[0].EnabledCloudwatchLogsExports = filtered
	}

	// Handle LogTypesToEnable
	if pending.LogTypesToEnable != nil {
		for _, toEnable := range pending.LogTypesToEnable {
			enableStr := pointer.StringValue(toEnable)
			found := false
			for _, enabled := range cr.DBClusters[0].EnabledCloudwatchLogsExports {
				if pointer.StringValue(enabled) == enableStr {
					found = true
					break
				}
			}
			if !found {
				cr.DBClusters[0].EnabledCloudwatchLogsExports = append(cr.DBClusters[0].EnabledCloudwatchLogsExports, toEnable)
			}
		}
	}
}
