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

// Code generated by ack-generate. DO NOT EDIT.

package dbinstance

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane/provider-aws/apis/docdb/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeDBInstancesInput returns input for read
// operation.
func GenerateDescribeDBInstancesInput(cr *svcapitypes.DBInstance) *svcsdk.DescribeDBInstancesInput {
	res := &svcsdk.DescribeDBInstancesInput{}

	if cr.Status.AtProvider.DBInstanceIdentifier != nil {
		res.SetDBInstanceIdentifier(*cr.Status.AtProvider.DBInstanceIdentifier)
	}

	return res
}

// GenerateDBInstance returns the current state in the form of *svcapitypes.DBInstance.
func GenerateDBInstance(resp *svcsdk.DescribeDBInstancesOutput) *svcapitypes.DBInstance {
	cr := &svcapitypes.DBInstance{}

	found := false
	for _, elem := range resp.DBInstances {
		if elem.AutoMinorVersionUpgrade != nil {
			cr.Spec.ForProvider.AutoMinorVersionUpgrade = elem.AutoMinorVersionUpgrade
		}
		if elem.AvailabilityZone != nil {
			cr.Spec.ForProvider.AvailabilityZone = elem.AvailabilityZone
		}
		if elem.BackupRetentionPeriod != nil {
			cr.Status.AtProvider.BackupRetentionPeriod = elem.BackupRetentionPeriod
		}
		if elem.CACertificateIdentifier != nil {
			cr.Status.AtProvider.CACertificateIdentifier = elem.CACertificateIdentifier
		}
		if elem.DBClusterIdentifier != nil {
			cr.Spec.ForProvider.DBClusterIdentifier = elem.DBClusterIdentifier
		}
		if elem.DBInstanceArn != nil {
			cr.Status.AtProvider.DBInstanceARN = elem.DBInstanceArn
		}
		if elem.DBInstanceClass != nil {
			cr.Spec.ForProvider.DBInstanceClass = elem.DBInstanceClass
		}
		if elem.DBInstanceIdentifier != nil {
			cr.Status.AtProvider.DBInstanceIdentifier = elem.DBInstanceIdentifier
		}
		if elem.DBInstanceStatus != nil {
			cr.Status.AtProvider.DBInstanceStatus = elem.DBInstanceStatus
		}
		if elem.DBSubnetGroup != nil {
			f9 := &svcapitypes.DBSubnetGroup_SDK{}
			if elem.DBSubnetGroup.DBSubnetGroupArn != nil {
				f9.DBSubnetGroupARN = elem.DBSubnetGroup.DBSubnetGroupArn
			}
			if elem.DBSubnetGroup.DBSubnetGroupDescription != nil {
				f9.DBSubnetGroupDescription = elem.DBSubnetGroup.DBSubnetGroupDescription
			}
			if elem.DBSubnetGroup.DBSubnetGroupName != nil {
				f9.DBSubnetGroupName = elem.DBSubnetGroup.DBSubnetGroupName
			}
			if elem.DBSubnetGroup.SubnetGroupStatus != nil {
				f9.SubnetGroupStatus = elem.DBSubnetGroup.SubnetGroupStatus
			}
			if elem.DBSubnetGroup.Subnets != nil {
				f9f4 := []*svcapitypes.Subnet{}
				for _, f9f4iter := range elem.DBSubnetGroup.Subnets {
					f9f4elem := &svcapitypes.Subnet{}
					if f9f4iter.SubnetAvailabilityZone != nil {
						f9f4elemf0 := &svcapitypes.AvailabilityZone{}
						if f9f4iter.SubnetAvailabilityZone.Name != nil {
							f9f4elemf0.Name = f9f4iter.SubnetAvailabilityZone.Name
						}
						f9f4elem.SubnetAvailabilityZone = f9f4elemf0
					}
					if f9f4iter.SubnetIdentifier != nil {
						f9f4elem.SubnetIdentifier = f9f4iter.SubnetIdentifier
					}
					if f9f4iter.SubnetStatus != nil {
						f9f4elem.SubnetStatus = f9f4iter.SubnetStatus
					}
					f9f4 = append(f9f4, f9f4elem)
				}
				f9.Subnets = f9f4
			}
			if elem.DBSubnetGroup.VpcId != nil {
				f9.VPCID = elem.DBSubnetGroup.VpcId
			}
			cr.Status.AtProvider.DBSubnetGroup = f9
		}
		if elem.DbiResourceId != nil {
			cr.Status.AtProvider.DBIResourceID = elem.DbiResourceId
		}
		if elem.EnabledCloudwatchLogsExports != nil {
			f11 := []*string{}
			for _, f11iter := range elem.EnabledCloudwatchLogsExports {
				var f11elem string
				f11elem = *f11iter
				f11 = append(f11, &f11elem)
			}
			cr.Status.AtProvider.EnabledCloudwatchLogsExports = f11
		}
		if elem.Endpoint != nil {
			f12 := &svcapitypes.Endpoint{}
			if elem.Endpoint.Address != nil {
				f12.Address = elem.Endpoint.Address
			}
			if elem.Endpoint.HostedZoneId != nil {
				f12.HostedZoneID = elem.Endpoint.HostedZoneId
			}
			if elem.Endpoint.Port != nil {
				f12.Port = elem.Endpoint.Port
			}
			cr.Status.AtProvider.Endpoint = f12
		}
		if elem.Engine != nil {
			cr.Spec.ForProvider.Engine = elem.Engine
		}
		if elem.EngineVersion != nil {
			cr.Status.AtProvider.EngineVersion = elem.EngineVersion
		}
		if elem.InstanceCreateTime != nil {
			cr.Status.AtProvider.InstanceCreateTime = &metav1.Time{*elem.InstanceCreateTime}
		}
		if elem.KmsKeyId != nil {
			cr.Status.AtProvider.KMSKeyID = elem.KmsKeyId
		}
		if elem.LatestRestorableTime != nil {
			cr.Status.AtProvider.LatestRestorableTime = &metav1.Time{*elem.LatestRestorableTime}
		}
		if elem.PendingModifiedValues != nil {
			f18 := &svcapitypes.PendingModifiedValues{}
			if elem.PendingModifiedValues.AllocatedStorage != nil {
				f18.AllocatedStorage = elem.PendingModifiedValues.AllocatedStorage
			}
			if elem.PendingModifiedValues.BackupRetentionPeriod != nil {
				f18.BackupRetentionPeriod = elem.PendingModifiedValues.BackupRetentionPeriod
			}
			if elem.PendingModifiedValues.CACertificateIdentifier != nil {
				f18.CACertificateIdentifier = elem.PendingModifiedValues.CACertificateIdentifier
			}
			if elem.PendingModifiedValues.DBInstanceClass != nil {
				f18.DBInstanceClass = elem.PendingModifiedValues.DBInstanceClass
			}
			if elem.PendingModifiedValues.DBInstanceIdentifier != nil {
				f18.DBInstanceIdentifier = elem.PendingModifiedValues.DBInstanceIdentifier
			}
			if elem.PendingModifiedValues.DBSubnetGroupName != nil {
				f18.DBSubnetGroupName = elem.PendingModifiedValues.DBSubnetGroupName
			}
			if elem.PendingModifiedValues.EngineVersion != nil {
				f18.EngineVersion = elem.PendingModifiedValues.EngineVersion
			}
			if elem.PendingModifiedValues.Iops != nil {
				f18.IOPS = elem.PendingModifiedValues.Iops
			}
			if elem.PendingModifiedValues.LicenseModel != nil {
				f18.LicenseModel = elem.PendingModifiedValues.LicenseModel
			}
			if elem.PendingModifiedValues.MasterUserPassword != nil {
				f18.MasterUserPassword = elem.PendingModifiedValues.MasterUserPassword
			}
			if elem.PendingModifiedValues.MultiAZ != nil {
				f18.MultiAZ = elem.PendingModifiedValues.MultiAZ
			}
			if elem.PendingModifiedValues.PendingCloudwatchLogsExports != nil {
				f18f11 := &svcapitypes.PendingCloudwatchLogsExports{}
				if elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable != nil {
					f18f11f0 := []*string{}
					for _, f18f11f0iter := range elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable {
						var f18f11f0elem string
						f18f11f0elem = *f18f11f0iter
						f18f11f0 = append(f18f11f0, &f18f11f0elem)
					}
					f18f11.LogTypesToDisable = f18f11f0
				}
				if elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable != nil {
					f18f11f1 := []*string{}
					for _, f18f11f1iter := range elem.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable {
						var f18f11f1elem string
						f18f11f1elem = *f18f11f1iter
						f18f11f1 = append(f18f11f1, &f18f11f1elem)
					}
					f18f11.LogTypesToEnable = f18f11f1
				}
				f18.PendingCloudwatchLogsExports = f18f11
			}
			if elem.PendingModifiedValues.Port != nil {
				f18.Port = elem.PendingModifiedValues.Port
			}
			if elem.PendingModifiedValues.StorageType != nil {
				f18.StorageType = elem.PendingModifiedValues.StorageType
			}
			cr.Status.AtProvider.PendingModifiedValues = f18
		}
		if elem.PreferredBackupWindow != nil {
			cr.Status.AtProvider.PreferredBackupWindow = elem.PreferredBackupWindow
		}
		if elem.PreferredMaintenanceWindow != nil {
			cr.Spec.ForProvider.PreferredMaintenanceWindow = elem.PreferredMaintenanceWindow
		}
		if elem.PromotionTier != nil {
			cr.Spec.ForProvider.PromotionTier = elem.PromotionTier
		}
		if elem.PubliclyAccessible != nil {
			cr.Status.AtProvider.PubliclyAccessible = elem.PubliclyAccessible
		}
		if elem.StatusInfos != nil {
			f23 := []*svcapitypes.DBInstanceStatusInfo{}
			for _, f23iter := range elem.StatusInfos {
				f23elem := &svcapitypes.DBInstanceStatusInfo{}
				if f23iter.Message != nil {
					f23elem.Message = f23iter.Message
				}
				if f23iter.Normal != nil {
					f23elem.Normal = f23iter.Normal
				}
				if f23iter.Status != nil {
					f23elem.Status = f23iter.Status
				}
				if f23iter.StatusType != nil {
					f23elem.StatusType = f23iter.StatusType
				}
				f23 = append(f23, f23elem)
			}
			cr.Status.AtProvider.StatusInfos = f23
		}
		if elem.StorageEncrypted != nil {
			cr.Status.AtProvider.StorageEncrypted = elem.StorageEncrypted
		}
		if elem.VpcSecurityGroups != nil {
			f25 := []*svcapitypes.VPCSecurityGroupMembership{}
			for _, f25iter := range elem.VpcSecurityGroups {
				f25elem := &svcapitypes.VPCSecurityGroupMembership{}
				if f25iter.Status != nil {
					f25elem.Status = f25iter.Status
				}
				if f25iter.VpcSecurityGroupId != nil {
					f25elem.VPCSecurityGroupID = f25iter.VpcSecurityGroupId
				}
				f25 = append(f25, f25elem)
			}
			cr.Status.AtProvider.VPCSecurityGroups = f25
		}
		found = true
		break
	}
	if !found {
		return cr
	}

	return cr
}

// GenerateCreateDBInstanceInput returns a create input.
func GenerateCreateDBInstanceInput(cr *svcapitypes.DBInstance) *svcsdk.CreateDBInstanceInput {
	res := &svcsdk.CreateDBInstanceInput{}

	if cr.Spec.ForProvider.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*cr.Spec.ForProvider.AutoMinorVersionUpgrade)
	}
	if cr.Spec.ForProvider.AvailabilityZone != nil {
		res.SetAvailabilityZone(*cr.Spec.ForProvider.AvailabilityZone)
	}
	if cr.Spec.ForProvider.DBClusterIdentifier != nil {
		res.SetDBClusterIdentifier(*cr.Spec.ForProvider.DBClusterIdentifier)
	}
	if cr.Spec.ForProvider.DBInstanceClass != nil {
		res.SetDBInstanceClass(*cr.Spec.ForProvider.DBInstanceClass)
	}
	if cr.Spec.ForProvider.Engine != nil {
		res.SetEngine(*cr.Spec.ForProvider.Engine)
	}
	if cr.Spec.ForProvider.PreferredMaintenanceWindow != nil {
		res.SetPreferredMaintenanceWindow(*cr.Spec.ForProvider.PreferredMaintenanceWindow)
	}
	if cr.Spec.ForProvider.PromotionTier != nil {
		res.SetPromotionTier(*cr.Spec.ForProvider.PromotionTier)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f7 := []*svcsdk.Tag{}
		for _, f7iter := range cr.Spec.ForProvider.Tags {
			f7elem := &svcsdk.Tag{}
			if f7iter.Key != nil {
				f7elem.SetKey(*f7iter.Key)
			}
			if f7iter.Value != nil {
				f7elem.SetValue(*f7iter.Value)
			}
			f7 = append(f7, f7elem)
		}
		res.SetTags(f7)
	}

	return res
}

// GenerateModifyDBInstanceInput returns an update input.
func GenerateModifyDBInstanceInput(cr *svcapitypes.DBInstance) *svcsdk.ModifyDBInstanceInput {
	res := &svcsdk.ModifyDBInstanceInput{}

	if cr.Spec.ForProvider.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*cr.Spec.ForProvider.AutoMinorVersionUpgrade)
	}
	if cr.Status.AtProvider.CACertificateIdentifier != nil {
		res.SetCACertificateIdentifier(*cr.Status.AtProvider.CACertificateIdentifier)
	}
	if cr.Spec.ForProvider.DBInstanceClass != nil {
		res.SetDBInstanceClass(*cr.Spec.ForProvider.DBInstanceClass)
	}
	if cr.Spec.ForProvider.PreferredMaintenanceWindow != nil {
		res.SetPreferredMaintenanceWindow(*cr.Spec.ForProvider.PreferredMaintenanceWindow)
	}
	if cr.Spec.ForProvider.PromotionTier != nil {
		res.SetPromotionTier(*cr.Spec.ForProvider.PromotionTier)
	}

	return res
}

// GenerateDeleteDBInstanceInput returns a deletion input.
func GenerateDeleteDBInstanceInput(cr *svcapitypes.DBInstance) *svcsdk.DeleteDBInstanceInput {
	res := &svcsdk.DeleteDBInstanceInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "DBInstanceNotFound"
}
