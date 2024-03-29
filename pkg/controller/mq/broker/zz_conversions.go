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

package broker

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeBrokerInput returns input for read
// operation.
func GenerateDescribeBrokerInput(cr *svcapitypes.Broker) *svcsdk.DescribeBrokerInput {
	res := &svcsdk.DescribeBrokerInput{}

	if cr.Status.AtProvider.BrokerID != nil {
		res.SetBrokerId(*cr.Status.AtProvider.BrokerID)
	}

	return res
}

// GenerateBroker returns the current state in the form of *svcapitypes.Broker.
func GenerateBroker(resp *svcsdk.DescribeBrokerResponse) *svcapitypes.Broker {
	cr := &svcapitypes.Broker{}

	if resp.AuthenticationStrategy != nil {
		cr.Spec.ForProvider.AuthenticationStrategy = resp.AuthenticationStrategy
	} else {
		cr.Spec.ForProvider.AuthenticationStrategy = nil
	}
	if resp.AutoMinorVersionUpgrade != nil {
		cr.Spec.ForProvider.AutoMinorVersionUpgrade = resp.AutoMinorVersionUpgrade
	} else {
		cr.Spec.ForProvider.AutoMinorVersionUpgrade = nil
	}
	if resp.BrokerArn != nil {
		cr.Status.AtProvider.BrokerARN = resp.BrokerArn
	} else {
		cr.Status.AtProvider.BrokerARN = nil
	}
	if resp.BrokerId != nil {
		cr.Status.AtProvider.BrokerID = resp.BrokerId
	} else {
		cr.Status.AtProvider.BrokerID = nil
	}
	if resp.BrokerInstances != nil {
		f5 := []*svcapitypes.BrokerInstance{}
		for _, f5iter := range resp.BrokerInstances {
			f5elem := &svcapitypes.BrokerInstance{}
			if f5iter.ConsoleURL != nil {
				f5elem.ConsoleURL = f5iter.ConsoleURL
			}
			if f5iter.Endpoints != nil {
				f5elemf1 := []*string{}
				for _, f5elemf1iter := range f5iter.Endpoints {
					var f5elemf1elem string
					f5elemf1elem = *f5elemf1iter
					f5elemf1 = append(f5elemf1, &f5elemf1elem)
				}
				f5elem.Endpoints = f5elemf1
			}
			if f5iter.IpAddress != nil {
				f5elem.IPAddress = f5iter.IpAddress
			}
			f5 = append(f5, f5elem)
		}
		cr.Status.AtProvider.BrokerInstances = f5
	} else {
		cr.Status.AtProvider.BrokerInstances = nil
	}
	if resp.BrokerState != nil {
		cr.Status.AtProvider.BrokerState = resp.BrokerState
	} else {
		cr.Status.AtProvider.BrokerState = nil
	}
	if resp.Configurations != nil {
		f8 := &svcapitypes.Configurations{}
		if resp.Configurations.Current != nil {
			f8f0 := &svcapitypes.ConfigurationID{}
			if resp.Configurations.Current.Id != nil {
				f8f0.ID = resp.Configurations.Current.Id
			}
			if resp.Configurations.Current.Revision != nil {
				f8f0.Revision = resp.Configurations.Current.Revision
			}
			f8.Current = f8f0
		}
		if resp.Configurations.History != nil {
			f8f1 := []*svcapitypes.ConfigurationID{}
			for _, f8f1iter := range resp.Configurations.History {
				f8f1elem := &svcapitypes.ConfigurationID{}
				if f8f1iter.Id != nil {
					f8f1elem.ID = f8f1iter.Id
				}
				if f8f1iter.Revision != nil {
					f8f1elem.Revision = f8f1iter.Revision
				}
				f8f1 = append(f8f1, f8f1elem)
			}
			f8.History = f8f1
		}
		if resp.Configurations.Pending != nil {
			f8f2 := &svcapitypes.ConfigurationID{}
			if resp.Configurations.Pending.Id != nil {
				f8f2.ID = resp.Configurations.Pending.Id
			}
			if resp.Configurations.Pending.Revision != nil {
				f8f2.Revision = resp.Configurations.Pending.Revision
			}
			f8.Pending = f8f2
		}
		cr.Status.AtProvider.Configurations = f8
	} else {
		cr.Status.AtProvider.Configurations = nil
	}
	if resp.Created != nil {
		cr.Status.AtProvider.Created = &metav1.Time{*resp.Created}
	} else {
		cr.Status.AtProvider.Created = nil
	}
	if resp.DataReplicationMode != nil {
		cr.Spec.ForProvider.DataReplicationMode = resp.DataReplicationMode
	} else {
		cr.Spec.ForProvider.DataReplicationMode = nil
	}
	if resp.DeploymentMode != nil {
		cr.Spec.ForProvider.DeploymentMode = resp.DeploymentMode
	} else {
		cr.Spec.ForProvider.DeploymentMode = nil
	}
	if resp.EncryptionOptions != nil {
		f13 := &svcapitypes.EncryptionOptions{}
		if resp.EncryptionOptions.KmsKeyId != nil {
			f13.KMSKeyID = resp.EncryptionOptions.KmsKeyId
		}
		if resp.EncryptionOptions.UseAwsOwnedKey != nil {
			f13.UseAWSOwnedKey = resp.EncryptionOptions.UseAwsOwnedKey
		}
		cr.Spec.ForProvider.EncryptionOptions = f13
	} else {
		cr.Spec.ForProvider.EncryptionOptions = nil
	}
	if resp.EngineType != nil {
		cr.Spec.ForProvider.EngineType = resp.EngineType
	} else {
		cr.Spec.ForProvider.EngineType = nil
	}
	if resp.EngineVersion != nil {
		cr.Spec.ForProvider.EngineVersion = resp.EngineVersion
	} else {
		cr.Spec.ForProvider.EngineVersion = nil
	}
	if resp.HostInstanceType != nil {
		cr.Spec.ForProvider.HostInstanceType = resp.HostInstanceType
	} else {
		cr.Spec.ForProvider.HostInstanceType = nil
	}
	if resp.LdapServerMetadata != nil {
		f17 := &svcapitypes.LDAPServerMetadataInput{}
		if resp.LdapServerMetadata.Hosts != nil {
			f17f0 := []*string{}
			for _, f17f0iter := range resp.LdapServerMetadata.Hosts {
				var f17f0elem string
				f17f0elem = *f17f0iter
				f17f0 = append(f17f0, &f17f0elem)
			}
			f17.Hosts = f17f0
		}
		if resp.LdapServerMetadata.RoleBase != nil {
			f17.RoleBase = resp.LdapServerMetadata.RoleBase
		}
		if resp.LdapServerMetadata.RoleName != nil {
			f17.RoleName = resp.LdapServerMetadata.RoleName
		}
		if resp.LdapServerMetadata.RoleSearchMatching != nil {
			f17.RoleSearchMatching = resp.LdapServerMetadata.RoleSearchMatching
		}
		if resp.LdapServerMetadata.RoleSearchSubtree != nil {
			f17.RoleSearchSubtree = resp.LdapServerMetadata.RoleSearchSubtree
		}
		if resp.LdapServerMetadata.ServiceAccountUsername != nil {
			f17.ServiceAccountUsername = resp.LdapServerMetadata.ServiceAccountUsername
		}
		if resp.LdapServerMetadata.UserBase != nil {
			f17.UserBase = resp.LdapServerMetadata.UserBase
		}
		if resp.LdapServerMetadata.UserRoleName != nil {
			f17.UserRoleName = resp.LdapServerMetadata.UserRoleName
		}
		if resp.LdapServerMetadata.UserSearchMatching != nil {
			f17.UserSearchMatching = resp.LdapServerMetadata.UserSearchMatching
		}
		if resp.LdapServerMetadata.UserSearchSubtree != nil {
			f17.UserSearchSubtree = resp.LdapServerMetadata.UserSearchSubtree
		}
		cr.Spec.ForProvider.LDAPServerMetadata = f17
	} else {
		cr.Spec.ForProvider.LDAPServerMetadata = nil
	}
	if resp.Logs != nil {
		f18 := &svcapitypes.Logs{}
		if resp.Logs.Audit != nil {
			f18.Audit = resp.Logs.Audit
		}
		if resp.Logs.General != nil {
			f18.General = resp.Logs.General
		}
		cr.Spec.ForProvider.Logs = f18
	} else {
		cr.Spec.ForProvider.Logs = nil
	}
	if resp.MaintenanceWindowStartTime != nil {
		f19 := &svcapitypes.WeeklyStartTime{}
		if resp.MaintenanceWindowStartTime.DayOfWeek != nil {
			f19.DayOfWeek = resp.MaintenanceWindowStartTime.DayOfWeek
		}
		if resp.MaintenanceWindowStartTime.TimeOfDay != nil {
			f19.TimeOfDay = resp.MaintenanceWindowStartTime.TimeOfDay
		}
		if resp.MaintenanceWindowStartTime.TimeZone != nil {
			f19.TimeZone = resp.MaintenanceWindowStartTime.TimeZone
		}
		cr.Spec.ForProvider.MaintenanceWindowStartTime = f19
	} else {
		cr.Spec.ForProvider.MaintenanceWindowStartTime = nil
	}
	if resp.PendingAuthenticationStrategy != nil {
		cr.Status.AtProvider.PendingAuthenticationStrategy = resp.PendingAuthenticationStrategy
	} else {
		cr.Status.AtProvider.PendingAuthenticationStrategy = nil
	}
	if resp.PendingEngineVersion != nil {
		cr.Status.AtProvider.PendingEngineVersion = resp.PendingEngineVersion
	} else {
		cr.Status.AtProvider.PendingEngineVersion = nil
	}
	if resp.PendingHostInstanceType != nil {
		cr.Status.AtProvider.PendingHostInstanceType = resp.PendingHostInstanceType
	} else {
		cr.Status.AtProvider.PendingHostInstanceType = nil
	}
	if resp.PendingLdapServerMetadata != nil {
		f25 := &svcapitypes.LDAPServerMetadataOutput{}
		if resp.PendingLdapServerMetadata.Hosts != nil {
			f25f0 := []*string{}
			for _, f25f0iter := range resp.PendingLdapServerMetadata.Hosts {
				var f25f0elem string
				f25f0elem = *f25f0iter
				f25f0 = append(f25f0, &f25f0elem)
			}
			f25.Hosts = f25f0
		}
		if resp.PendingLdapServerMetadata.RoleBase != nil {
			f25.RoleBase = resp.PendingLdapServerMetadata.RoleBase
		}
		if resp.PendingLdapServerMetadata.RoleName != nil {
			f25.RoleName = resp.PendingLdapServerMetadata.RoleName
		}
		if resp.PendingLdapServerMetadata.RoleSearchMatching != nil {
			f25.RoleSearchMatching = resp.PendingLdapServerMetadata.RoleSearchMatching
		}
		if resp.PendingLdapServerMetadata.RoleSearchSubtree != nil {
			f25.RoleSearchSubtree = resp.PendingLdapServerMetadata.RoleSearchSubtree
		}
		if resp.PendingLdapServerMetadata.ServiceAccountUsername != nil {
			f25.ServiceAccountUsername = resp.PendingLdapServerMetadata.ServiceAccountUsername
		}
		if resp.PendingLdapServerMetadata.UserBase != nil {
			f25.UserBase = resp.PendingLdapServerMetadata.UserBase
		}
		if resp.PendingLdapServerMetadata.UserRoleName != nil {
			f25.UserRoleName = resp.PendingLdapServerMetadata.UserRoleName
		}
		if resp.PendingLdapServerMetadata.UserSearchMatching != nil {
			f25.UserSearchMatching = resp.PendingLdapServerMetadata.UserSearchMatching
		}
		if resp.PendingLdapServerMetadata.UserSearchSubtree != nil {
			f25.UserSearchSubtree = resp.PendingLdapServerMetadata.UserSearchSubtree
		}
		cr.Status.AtProvider.PendingLDAPServerMetadata = f25
	} else {
		cr.Status.AtProvider.PendingLDAPServerMetadata = nil
	}
	if resp.PendingSecurityGroups != nil {
		f26 := []*string{}
		for _, f26iter := range resp.PendingSecurityGroups {
			var f26elem string
			f26elem = *f26iter
			f26 = append(f26, &f26elem)
		}
		cr.Status.AtProvider.PendingSecurityGroups = f26
	} else {
		cr.Status.AtProvider.PendingSecurityGroups = nil
	}
	if resp.PubliclyAccessible != nil {
		cr.Spec.ForProvider.PubliclyAccessible = resp.PubliclyAccessible
	} else {
		cr.Spec.ForProvider.PubliclyAccessible = nil
	}
	if resp.StorageType != nil {
		cr.Spec.ForProvider.StorageType = resp.StorageType
	} else {
		cr.Spec.ForProvider.StorageType = nil
	}
	if resp.Tags != nil {
		f31 := map[string]*string{}
		for f31key, f31valiter := range resp.Tags {
			var f31val string
			f31val = *f31valiter
			f31[f31key] = &f31val
		}
		cr.Spec.ForProvider.Tags = f31
	} else {
		cr.Spec.ForProvider.Tags = nil
	}
	if resp.Users != nil {
		f32 := []*svcapitypes.UserSummary{}
		for _, f32iter := range resp.Users {
			f32elem := &svcapitypes.UserSummary{}
			if f32iter.PendingChange != nil {
				f32elem.PendingChange = f32iter.PendingChange
			}
			if f32iter.Username != nil {
				f32elem.Username = f32iter.Username
			}
			f32 = append(f32, f32elem)
		}
		cr.Status.AtProvider.Users = f32
	} else {
		cr.Status.AtProvider.Users = nil
	}

	return cr
}

// GenerateCreateBrokerRequest returns a create input.
func GenerateCreateBrokerRequest(cr *svcapitypes.Broker) *svcsdk.CreateBrokerRequest {
	res := &svcsdk.CreateBrokerRequest{}

	if cr.Spec.ForProvider.AuthenticationStrategy != nil {
		res.SetAuthenticationStrategy(*cr.Spec.ForProvider.AuthenticationStrategy)
	}
	if cr.Spec.ForProvider.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*cr.Spec.ForProvider.AutoMinorVersionUpgrade)
	}
	if cr.Spec.ForProvider.Configuration != nil {
		f2 := &svcsdk.ConfigurationId{}
		if cr.Spec.ForProvider.Configuration.ID != nil {
			f2.SetId(*cr.Spec.ForProvider.Configuration.ID)
		}
		if cr.Spec.ForProvider.Configuration.Revision != nil {
			f2.SetRevision(*cr.Spec.ForProvider.Configuration.Revision)
		}
		res.SetConfiguration(f2)
	}
	if cr.Spec.ForProvider.CreatorRequestID != nil {
		res.SetCreatorRequestId(*cr.Spec.ForProvider.CreatorRequestID)
	}
	if cr.Spec.ForProvider.DataReplicationMode != nil {
		res.SetDataReplicationMode(*cr.Spec.ForProvider.DataReplicationMode)
	}
	if cr.Spec.ForProvider.DataReplicationPrimaryBrokerARN != nil {
		res.SetDataReplicationPrimaryBrokerArn(*cr.Spec.ForProvider.DataReplicationPrimaryBrokerARN)
	}
	if cr.Spec.ForProvider.DeploymentMode != nil {
		res.SetDeploymentMode(*cr.Spec.ForProvider.DeploymentMode)
	}
	if cr.Spec.ForProvider.EncryptionOptions != nil {
		f7 := &svcsdk.EncryptionOptions{}
		if cr.Spec.ForProvider.EncryptionOptions.KMSKeyID != nil {
			f7.SetKmsKeyId(*cr.Spec.ForProvider.EncryptionOptions.KMSKeyID)
		}
		if cr.Spec.ForProvider.EncryptionOptions.UseAWSOwnedKey != nil {
			f7.SetUseAwsOwnedKey(*cr.Spec.ForProvider.EncryptionOptions.UseAWSOwnedKey)
		}
		res.SetEncryptionOptions(f7)
	}
	if cr.Spec.ForProvider.EngineType != nil {
		res.SetEngineType(*cr.Spec.ForProvider.EngineType)
	}
	if cr.Spec.ForProvider.EngineVersion != nil {
		res.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}
	if cr.Spec.ForProvider.HostInstanceType != nil {
		res.SetHostInstanceType(*cr.Spec.ForProvider.HostInstanceType)
	}
	if cr.Spec.ForProvider.LDAPServerMetadata != nil {
		f11 := &svcsdk.LdapServerMetadataInput{}
		if cr.Spec.ForProvider.LDAPServerMetadata.Hosts != nil {
			f11f0 := []*string{}
			for _, f11f0iter := range cr.Spec.ForProvider.LDAPServerMetadata.Hosts {
				var f11f0elem string
				f11f0elem = *f11f0iter
				f11f0 = append(f11f0, &f11f0elem)
			}
			f11.SetHosts(f11f0)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleBase != nil {
			f11.SetRoleBase(*cr.Spec.ForProvider.LDAPServerMetadata.RoleBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleName != nil {
			f11.SetRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.RoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching != nil {
			f11.SetRoleSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree != nil {
			f11.SetRoleSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword != nil {
			f11.SetServiceAccountPassword(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername != nil {
			f11.SetServiceAccountUsername(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserBase != nil {
			f11.SetUserBase(*cr.Spec.ForProvider.LDAPServerMetadata.UserBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName != nil {
			f11.SetUserRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching != nil {
			f11.SetUserSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree != nil {
			f11.SetUserSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree)
		}
		res.SetLdapServerMetadata(f11)
	}
	if cr.Spec.ForProvider.Logs != nil {
		f12 := &svcsdk.Logs{}
		if cr.Spec.ForProvider.Logs.Audit != nil {
			f12.SetAudit(*cr.Spec.ForProvider.Logs.Audit)
		}
		if cr.Spec.ForProvider.Logs.General != nil {
			f12.SetGeneral(*cr.Spec.ForProvider.Logs.General)
		}
		res.SetLogs(f12)
	}
	if cr.Spec.ForProvider.MaintenanceWindowStartTime != nil {
		f13 := &svcsdk.WeeklyStartTime{}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek != nil {
			f13.SetDayOfWeek(*cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay != nil {
			f13.SetTimeOfDay(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone != nil {
			f13.SetTimeZone(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone)
		}
		res.SetMaintenanceWindowStartTime(f13)
	}
	if cr.Spec.ForProvider.PubliclyAccessible != nil {
		res.SetPubliclyAccessible(*cr.Spec.ForProvider.PubliclyAccessible)
	}
	if cr.Spec.ForProvider.StorageType != nil {
		res.SetStorageType(*cr.Spec.ForProvider.StorageType)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f16 := map[string]*string{}
		for f16key, f16valiter := range cr.Spec.ForProvider.Tags {
			var f16val string
			f16val = *f16valiter
			f16[f16key] = &f16val
		}
		res.SetTags(f16)
	}

	return res
}

// GenerateUpdateBrokerRequest returns an update input.
func GenerateUpdateBrokerRequest(cr *svcapitypes.Broker) *svcsdk.UpdateBrokerRequest {
	res := &svcsdk.UpdateBrokerRequest{}

	if cr.Spec.ForProvider.AuthenticationStrategy != nil {
		res.SetAuthenticationStrategy(*cr.Spec.ForProvider.AuthenticationStrategy)
	}
	if cr.Spec.ForProvider.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*cr.Spec.ForProvider.AutoMinorVersionUpgrade)
	}
	if cr.Status.AtProvider.BrokerID != nil {
		res.SetBrokerId(*cr.Status.AtProvider.BrokerID)
	}
	if cr.Spec.ForProvider.Configuration != nil {
		f3 := &svcsdk.ConfigurationId{}
		if cr.Spec.ForProvider.Configuration.ID != nil {
			f3.SetId(*cr.Spec.ForProvider.Configuration.ID)
		}
		if cr.Spec.ForProvider.Configuration.Revision != nil {
			f3.SetRevision(*cr.Spec.ForProvider.Configuration.Revision)
		}
		res.SetConfiguration(f3)
	}
	if cr.Spec.ForProvider.DataReplicationMode != nil {
		res.SetDataReplicationMode(*cr.Spec.ForProvider.DataReplicationMode)
	}
	if cr.Spec.ForProvider.EngineVersion != nil {
		res.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}
	if cr.Spec.ForProvider.HostInstanceType != nil {
		res.SetHostInstanceType(*cr.Spec.ForProvider.HostInstanceType)
	}
	if cr.Spec.ForProvider.LDAPServerMetadata != nil {
		f7 := &svcsdk.LdapServerMetadataInput{}
		if cr.Spec.ForProvider.LDAPServerMetadata.Hosts != nil {
			f7f0 := []*string{}
			for _, f7f0iter := range cr.Spec.ForProvider.LDAPServerMetadata.Hosts {
				var f7f0elem string
				f7f0elem = *f7f0iter
				f7f0 = append(f7f0, &f7f0elem)
			}
			f7.SetHosts(f7f0)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleBase != nil {
			f7.SetRoleBase(*cr.Spec.ForProvider.LDAPServerMetadata.RoleBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleName != nil {
			f7.SetRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.RoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching != nil {
			f7.SetRoleSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree != nil {
			f7.SetRoleSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword != nil {
			f7.SetServiceAccountPassword(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername != nil {
			f7.SetServiceAccountUsername(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserBase != nil {
			f7.SetUserBase(*cr.Spec.ForProvider.LDAPServerMetadata.UserBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName != nil {
			f7.SetUserRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching != nil {
			f7.SetUserSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree != nil {
			f7.SetUserSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree)
		}
		res.SetLdapServerMetadata(f7)
	}
	if cr.Spec.ForProvider.Logs != nil {
		f8 := &svcsdk.Logs{}
		if cr.Spec.ForProvider.Logs.Audit != nil {
			f8.SetAudit(*cr.Spec.ForProvider.Logs.Audit)
		}
		if cr.Spec.ForProvider.Logs.General != nil {
			f8.SetGeneral(*cr.Spec.ForProvider.Logs.General)
		}
		res.SetLogs(f8)
	}
	if cr.Spec.ForProvider.MaintenanceWindowStartTime != nil {
		f9 := &svcsdk.WeeklyStartTime{}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek != nil {
			f9.SetDayOfWeek(*cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay != nil {
			f9.SetTimeOfDay(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone != nil {
			f9.SetTimeZone(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone)
		}
		res.SetMaintenanceWindowStartTime(f9)
	}

	return res
}

// GenerateDeleteBrokerInput returns a deletion input.
func GenerateDeleteBrokerInput(cr *svcapitypes.Broker) *svcsdk.DeleteBrokerInput {
	res := &svcsdk.DeleteBrokerInput{}

	if cr.Status.AtProvider.BrokerID != nil {
		res.SetBrokerId(*cr.Status.AtProvider.BrokerID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
