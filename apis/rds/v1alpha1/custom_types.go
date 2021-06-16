/*
Copyright 2020 The Crossplane Authors.

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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomDBParameterGroupParameters are custom parameters for DBParameterGroup
type CustomDBParameterGroupParameters struct {
	// A list of parameters to associate with this DB parameter group
	// +optional
	Parameters []Parameter `json:"parameters,omitempty"`
}

// CustomDBClusterParameters are custom parameters for DBCluster
type CustomDBClusterParameters struct {

	// DomainIAMRoleNameRef is a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	DomainIAMRoleNameRef *xpv1.Reference `json:"domainIAMRoleNameRef,omitempty"`

	// DomainIAMRoleNameSelector selects a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	DomainIAMRoleNameSelector *xpv1.Selector `json:"domainIAMRoleNameSelector,omitempty"`

	// KMSKeyIDRef is a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIDRef,omitempty"`

	// KMSKeyIDSelector selects a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIDSelector,omitempty"`

	// The password for the master database user. This password can contain any
	// printable ASCII character except "/", """, or "@".
	//
	// Constraints: Must contain from 8 to 41 characters. Required.
	MasterUserPasswordSecretRef xpv1.SecretKeySelector `json:"masterUserPasswordSecretRef"`

	// A list of EC2 VPC security groups to associate with this DB cluster.
	VPCSecurityGroupIDs []string `json:"vpcSecurityGroupIDs,omitempty"`

	// VPCSecurityGroupIDRefs are references to VPCSecurityGroups used to set
	// the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDRefs []xpv1.Reference `json:"vpcSecurityGroupIDRefs,omitempty"`

	// VPCSecurityGroupIDSelector selects references to VPCSecurityGroups used
	// to set the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDSelector *xpv1.Selector `json:"vpcSecurityGroupIDSelector,omitempty"`

	// DBSubnetGroupNameRef is a reference to a DBSubnetGroup used to set
	// DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameRef *xpv1.Reference `json:"dbSubnetGroupNameRef,omitempty"`

	// DBSubnetGroupNameSelector selects a reference to a DBSubnetGroup used to
	// set DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameSelector *xpv1.Selector `json:"dbSubnetGroupNameSelector,omitempty"`

	// The DB cluster snapshot identifier of the new DB cluster snapshot created
	// when SkipFinalSnapshot is disabled.
	//
	// Specifying this parameter and also skipping the creation of a final DB cluster
	// snapshot with the SkipFinalShapshot parameter results in an error.
	//
	// Constraints:
	//
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//
	//    * First character must be a letter
	//
	//    * Can't end with a hyphen or contain two consecutive hyphens
	// +immutable
	// +optional
	FinalDBSnapshotIdentifier string `json:"finalDBSnapshotIdentifier,omitempty"`

	// A value that indicates whether to skip the creation of a final DB cluster
	// snapshot before the DB cluster is deleted. If skip is specified, no DB cluster
	// snapshot is created. If skip isn't specified, a DB cluster snapshot is created
	// before the DB cluster is deleted. By default, skip isn't specified, and the
	// DB cluster snapshot is created. By default, this parameter is disabled.
	//
	// You must specify a FinalDBSnapshotIdentifier parameter if SkipFinalSnapshot
	// is disabled.
	// +immutable
	// +optional
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,omitempty"`
}

// CustomGlobalClusterParameters are custom parameters for a GlobalCluster
type CustomGlobalClusterParameters struct {
	// SourceDBClusterIdentifierRef is a reference to a DBCluster used to set
	// SourceDBClusterIdentifier.
	// +immutable
	// +optional
	SourceDBClusterIdentifierRef *xpv1.Reference `json:"sourceDBClusterIdentifierRef,omitempty"`

	// SourceDBClusterIdentifierSelector selects a reference to a DBCluster used to
	// set SourceDBClusterIdentifier.
	// +immutable
	// +optional
	SourceDBClusterIdentifierSelector *xpv1.Selector `json:"sourceDBClusterIdentifierSelector,omitempty"`
}

// CustomDBInstanceParameters are custom parameters for the DBInstance
type CustomDBInstanceParameters struct {
	// AutogeneratePassword indicates whether the controller should generate
	// a random password for the master user if one is not provided via
	// MasterUserPasswordSecretRef.
	//
	// If a password is generated, it will
	// be stored as a secret at the location specified by MasterUserPasswordSecretRef.
	// +optional
	AutogeneratePassword bool `json:"autogeneratePassword,omitempty"`

	// A list of database security groups to associate with this DB instance
	DBSecurityGroups []string `json:"dbSecurityGroups,omitempty"`

	// DBSubnetGroupNameRef is a reference to a DBSubnetGroup used to set
	// DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameRef *xpv1.Reference `json:"dbSubnetGroupNameRef,omitempty"`

	// DBSubnetGroupNameSelector selects a reference to a DBSubnetGroup used to
	// set DBSubnetGroupName.
	// +immutable
	// +optional
	DBSubnetGroupNameSelector *xpv1.Selector `json:"dbSubnetGroupNameSelector,omitempty"`

	// DomainIAMRoleNameRef is a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	// +immutable
	DomainIAMRoleNameRef *xpv1.Reference `json:"domainIAMRoleNameRef,omitempty"`

	// DomainIAMRoleNameSelector selects a reference to an IAMRole used to set
	// DomainIAMRoleName.
	// +optional
	// +immutable
	DomainIAMRoleNameSelector *xpv1.Selector `json:"domainIAMRoleNameSelector,omitempty"`

	// The DB instance snapshot identifier of the new DB instance snapshot created
	// when SkipFinalSnapshot is disabled.
	//
	// Specifying this parameter and also skipping the creation of a final DB instance
	// snapshot with the SkipFinalShapshot parameter results in an error.
	//
	// Constraints:
	//
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//
	//    * First character must be a letter
	//
	//    * Can't end with a hyphen or contain two consecutive hyphens
	// +immutable
	// +optional
	FinalDBSnapshotIdentifier string `json:"finalDBSnapshotIdentifier,omitempty"`

	// The password for the master database user. This password can contain any
	// printable ASCII character except "/", """, or "@".
	//
	// Constraints: Must contain from 8 to 41 characters.
	// +optional
	MasterUserPasswordSecretRef *xpv1.SecretKeySelector `json:"masterUserPasswordSecretRef,omitempty"`

	// MonitoringRoleARNRef is a reference to an IAMRole used to set
	// MonitoringRoleARN.
	// +optional
	// +immutable
	MonitoringRoleARNRef *xpv1.Reference `json:"monitoringRoleArnRef,omitempty"`

	// MonitoringRoleARNSelector selects a reference to an IAMRole used to set
	// MonitoringRoleARN.
	// +optional
	// +immutable
	MonitoringRoleARNSelector *xpv1.Selector `json:"monitoringRoleArnSelector,omitempty"`

	// A value that indicates whether to skip the creation of a final DB instance
	// snapshot before the DB instance is deleted. If skip is specified, no DB instance
	// snapshot is created. If skip isn't specified, a DB instance snapshot is created
	// before the DB instance is deleted. By default, skip isn't specified, and the
	// DB instance snapshot is created. By default, this parameter is disabled.
	//
	// You must specify a FinalDBSnapshotIdentifier parameter if SkipFinalSnapshot
	// is disabled.
	// +immutable
	// +optional
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,omitempty"`

	// A list of EC2 VPC security groups to associate with this DB instance.
	VPCSecurityGroupIDs []string `json:"vpcSecurityGroupIDs,omitempty"`

	// VPCSecurityGroupIDRefs are references to VPCSecurityGroups used to set
	// the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDRefs []xpv1.Reference `json:"vpcSecurityGroupIDRefs,omitempty"`

	// VPCSecurityGroupIDSelector selects references to VPCSecurityGroups used
	// to set the VPCSecurityGroupIDs.
	// +optional
	VPCSecurityGroupIDSelector *xpv1.Selector `json:"vpcSecurityGroupIDSelector,omitempty"`
}
