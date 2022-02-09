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

	svcapitypes "github.com/crossplane/provider-aws/apis/mq/v1alpha1"
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
	if resp.DeploymentMode != nil {
		cr.Spec.ForProvider.DeploymentMode = resp.DeploymentMode
	} else {
		cr.Spec.ForProvider.DeploymentMode = nil
	}
	if resp.EncryptionOptions != nil {
		f10 := &svcapitypes.EncryptionOptions{}
		if resp.EncryptionOptions.KmsKeyId != nil {
			f10.KMSKeyID = resp.EncryptionOptions.KmsKeyId
		}
		if resp.EncryptionOptions.UseAwsOwnedKey != nil {
			f10.UseAWSOwnedKey = resp.EncryptionOptions.UseAwsOwnedKey
		}
		cr.Spec.ForProvider.EncryptionOptions = f10
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
		f14 := &svcapitypes.LDAPServerMetadataInput{}
		if resp.LdapServerMetadata.Hosts != nil {
			f14f0 := []*string{}
			for _, f14f0iter := range resp.LdapServerMetadata.Hosts {
				var f14f0elem string
				f14f0elem = *f14f0iter
				f14f0 = append(f14f0, &f14f0elem)
			}
			f14.Hosts = f14f0
		}
		if resp.LdapServerMetadata.RoleBase != nil {
			f14.RoleBase = resp.LdapServerMetadata.RoleBase
		}
		if resp.LdapServerMetadata.RoleName != nil {
			f14.RoleName = resp.LdapServerMetadata.RoleName
		}
		if resp.LdapServerMetadata.RoleSearchMatching != nil {
			f14.RoleSearchMatching = resp.LdapServerMetadata.RoleSearchMatching
		}
		if resp.LdapServerMetadata.RoleSearchSubtree != nil {
			f14.RoleSearchSubtree = resp.LdapServerMetadata.RoleSearchSubtree
		}
		if resp.LdapServerMetadata.ServiceAccountUsername != nil {
			f14.ServiceAccountUsername = resp.LdapServerMetadata.ServiceAccountUsername
		}
		if resp.LdapServerMetadata.UserBase != nil {
			f14.UserBase = resp.LdapServerMetadata.UserBase
		}
		if resp.LdapServerMetadata.UserRoleName != nil {
			f14.UserRoleName = resp.LdapServerMetadata.UserRoleName
		}
		if resp.LdapServerMetadata.UserSearchMatching != nil {
			f14.UserSearchMatching = resp.LdapServerMetadata.UserSearchMatching
		}
		if resp.LdapServerMetadata.UserSearchSubtree != nil {
			f14.UserSearchSubtree = resp.LdapServerMetadata.UserSearchSubtree
		}
		cr.Spec.ForProvider.LDAPServerMetadata = f14
	} else {
		cr.Spec.ForProvider.LDAPServerMetadata = nil
	}
	if resp.Logs != nil {
		f15 := &svcapitypes.Logs{}
		if resp.Logs.Audit != nil {
			f15.Audit = resp.Logs.Audit
		}
		if resp.Logs.General != nil {
			f15.General = resp.Logs.General
		}
		cr.Spec.ForProvider.Logs = f15
	} else {
		cr.Spec.ForProvider.Logs = nil
	}
	if resp.MaintenanceWindowStartTime != nil {
		f16 := &svcapitypes.WeeklyStartTime{}
		if resp.MaintenanceWindowStartTime.DayOfWeek != nil {
			f16.DayOfWeek = resp.MaintenanceWindowStartTime.DayOfWeek
		}
		if resp.MaintenanceWindowStartTime.TimeOfDay != nil {
			f16.TimeOfDay = resp.MaintenanceWindowStartTime.TimeOfDay
		}
		if resp.MaintenanceWindowStartTime.TimeZone != nil {
			f16.TimeZone = resp.MaintenanceWindowStartTime.TimeZone
		}
		cr.Spec.ForProvider.MaintenanceWindowStartTime = f16
	} else {
		cr.Spec.ForProvider.MaintenanceWindowStartTime = nil
	}
	if resp.PubliclyAccessible != nil {
		cr.Spec.ForProvider.PubliclyAccessible = resp.PubliclyAccessible
	} else {
		cr.Spec.ForProvider.PubliclyAccessible = nil
	}
	if resp.SecurityGroups != nil {
		f23 := []*string{}
		for _, f23iter := range resp.SecurityGroups {
			var f23elem string
			f23elem = *f23iter
			f23 = append(f23, &f23elem)
		}
		cr.Spec.ForProvider.SecurityGroups = f23
	} else {
		cr.Spec.ForProvider.SecurityGroups = nil
	}
	if resp.StorageType != nil {
		cr.Spec.ForProvider.StorageType = resp.StorageType
	} else {
		cr.Spec.ForProvider.StorageType = nil
	}
	if resp.SubnetIds != nil {
		f25 := []*string{}
		for _, f25iter := range resp.SubnetIds {
			var f25elem string
			f25elem = *f25iter
			f25 = append(f25, &f25elem)
		}
		cr.Spec.ForProvider.SubnetIDs = f25
	} else {
		cr.Spec.ForProvider.SubnetIDs = nil
	}
	if resp.Tags != nil {
		f26 := map[string]*string{}
		for f26key, f26valiter := range resp.Tags {
			var f26val string
			f26val = *f26valiter
			f26[f26key] = &f26val
		}
		cr.Spec.ForProvider.Tags = f26
	} else {
		cr.Spec.ForProvider.Tags = nil
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
	if cr.Spec.ForProvider.DeploymentMode != nil {
		res.SetDeploymentMode(*cr.Spec.ForProvider.DeploymentMode)
	}
	if cr.Spec.ForProvider.EncryptionOptions != nil {
		f5 := &svcsdk.EncryptionOptions{}
		if cr.Spec.ForProvider.EncryptionOptions.KMSKeyID != nil {
			f5.SetKmsKeyId(*cr.Spec.ForProvider.EncryptionOptions.KMSKeyID)
		}
		if cr.Spec.ForProvider.EncryptionOptions.UseAWSOwnedKey != nil {
			f5.SetUseAwsOwnedKey(*cr.Spec.ForProvider.EncryptionOptions.UseAWSOwnedKey)
		}
		res.SetEncryptionOptions(f5)
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
		f9 := &svcsdk.LdapServerMetadataInput{}
		if cr.Spec.ForProvider.LDAPServerMetadata.Hosts != nil {
			f9f0 := []*string{}
			for _, f9f0iter := range cr.Spec.ForProvider.LDAPServerMetadata.Hosts {
				var f9f0elem string
				f9f0elem = *f9f0iter
				f9f0 = append(f9f0, &f9f0elem)
			}
			f9.SetHosts(f9f0)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleBase != nil {
			f9.SetRoleBase(*cr.Spec.ForProvider.LDAPServerMetadata.RoleBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleName != nil {
			f9.SetRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.RoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching != nil {
			f9.SetRoleSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree != nil {
			f9.SetRoleSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword != nil {
			f9.SetServiceAccountPassword(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername != nil {
			f9.SetServiceAccountUsername(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserBase != nil {
			f9.SetUserBase(*cr.Spec.ForProvider.LDAPServerMetadata.UserBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName != nil {
			f9.SetUserRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching != nil {
			f9.SetUserSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree != nil {
			f9.SetUserSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree)
		}
		res.SetLdapServerMetadata(f9)
	}
	if cr.Spec.ForProvider.Logs != nil {
		f10 := &svcsdk.Logs{}
		if cr.Spec.ForProvider.Logs.Audit != nil {
			f10.SetAudit(*cr.Spec.ForProvider.Logs.Audit)
		}
		if cr.Spec.ForProvider.Logs.General != nil {
			f10.SetGeneral(*cr.Spec.ForProvider.Logs.General)
		}
		res.SetLogs(f10)
	}
	if cr.Spec.ForProvider.MaintenanceWindowStartTime != nil {
		f11 := &svcsdk.WeeklyStartTime{}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek != nil {
			f11.SetDayOfWeek(*cr.Spec.ForProvider.MaintenanceWindowStartTime.DayOfWeek)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay != nil {
			f11.SetTimeOfDay(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeOfDay)
		}
		if cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone != nil {
			f11.SetTimeZone(*cr.Spec.ForProvider.MaintenanceWindowStartTime.TimeZone)
		}
		res.SetMaintenanceWindowStartTime(f11)
	}
	if cr.Spec.ForProvider.PubliclyAccessible != nil {
		res.SetPubliclyAccessible(*cr.Spec.ForProvider.PubliclyAccessible)
	}
	if cr.Spec.ForProvider.SecurityGroups != nil {
		f13 := []*string{}
		for _, f13iter := range cr.Spec.ForProvider.SecurityGroups {
			var f13elem string
			f13elem = *f13iter
			f13 = append(f13, &f13elem)
		}
		res.SetSecurityGroups(f13)
	}
	if cr.Spec.ForProvider.StorageType != nil {
		res.SetStorageType(*cr.Spec.ForProvider.StorageType)
	}
	if cr.Spec.ForProvider.SubnetIDs != nil {
		f15 := []*string{}
		for _, f15iter := range cr.Spec.ForProvider.SubnetIDs {
			var f15elem string
			f15elem = *f15iter
			f15 = append(f15, &f15elem)
		}
		res.SetSubnetIds(f15)
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
	if cr.Spec.ForProvider.EngineVersion != nil {
		res.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}
	if cr.Spec.ForProvider.HostInstanceType != nil {
		res.SetHostInstanceType(*cr.Spec.ForProvider.HostInstanceType)
	}
	if cr.Spec.ForProvider.LDAPServerMetadata != nil {
		f6 := &svcsdk.LdapServerMetadataInput{}
		if cr.Spec.ForProvider.LDAPServerMetadata.Hosts != nil {
			f6f0 := []*string{}
			for _, f6f0iter := range cr.Spec.ForProvider.LDAPServerMetadata.Hosts {
				var f6f0elem string
				f6f0elem = *f6f0iter
				f6f0 = append(f6f0, &f6f0elem)
			}
			f6.SetHosts(f6f0)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleBase != nil {
			f6.SetRoleBase(*cr.Spec.ForProvider.LDAPServerMetadata.RoleBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleName != nil {
			f6.SetRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.RoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching != nil {
			f6.SetRoleSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree != nil {
			f6.SetRoleSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.RoleSearchSubtree)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword != nil {
			f6.SetServiceAccountPassword(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountPassword)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername != nil {
			f6.SetServiceAccountUsername(*cr.Spec.ForProvider.LDAPServerMetadata.ServiceAccountUsername)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserBase != nil {
			f6.SetUserBase(*cr.Spec.ForProvider.LDAPServerMetadata.UserBase)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName != nil {
			f6.SetUserRoleName(*cr.Spec.ForProvider.LDAPServerMetadata.UserRoleName)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching != nil {
			f6.SetUserSearchMatching(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchMatching)
		}
		if cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree != nil {
			f6.SetUserSearchSubtree(*cr.Spec.ForProvider.LDAPServerMetadata.UserSearchSubtree)
		}
		res.SetLdapServerMetadata(f6)
	}
	if cr.Spec.ForProvider.Logs != nil {
		f7 := &svcsdk.Logs{}
		if cr.Spec.ForProvider.Logs.Audit != nil {
			f7.SetAudit(*cr.Spec.ForProvider.Logs.Audit)
		}
		if cr.Spec.ForProvider.Logs.General != nil {
			f7.SetGeneral(*cr.Spec.ForProvider.Logs.General)
		}
		res.SetLogs(f7)
	}
	if cr.Spec.ForProvider.SecurityGroups != nil {
		f8 := []*string{}
		for _, f8iter := range cr.Spec.ForProvider.SecurityGroups {
			var f8elem string
			f8elem = *f8iter
			f8 = append(f8, &f8elem)
		}
		res.SetSecurityGroups(f8)
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
