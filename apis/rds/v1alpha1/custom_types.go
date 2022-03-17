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

// CustomDBClusterParameterGroupParameters are custom parameters for DBClusterParameterGroup
type CustomDBClusterParameterGroupParameters struct {
	// A list of parameters to associate with this DB cluster parameter group
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

	// DBClusterParameterGroupNameRef is a reference to a DBClusterParameterGroup used to set
	// DBClusterParameterGroupName.
	// +optional
	DBClusterParameterGroupNameRef *xpv1.Reference `json:"dbClusterParameterGroupNameRef,omitempty"`

	// DBClusterParameterGroupNameSelector selects a reference to a DBClusterParameterGroup used to
	// set DBClusterParameterGroupName.
	// +optional
	DBClusterParameterGroupNameSelector *xpv1.Selector `json:"dbClusterParameterGroupNameSelector,omitempty"`

	// A value that indicates whether the modifications in this request and any
	// pending modifications are asynchronously applied as soon as possible, regardless
	// of the PreferredMaintenanceWindow setting for the DB cluster. If this parameter
	// is disabled, changes to the DB cluster are applied during the next maintenance
	// window.
	//
	// The ApplyImmediately parameter only affects the EnableIAMDatabaseAuthentication,
	// MasterUserPassword values. If the ApplyImmediately
	// parameter is disabled, then changes to the EnableIAMDatabaseAuthentication,
	// MasterUserPassword values are applied during
	// the next maintenance window. All other changes are applied immediately, regardless
	// of the value of the ApplyImmediately parameter.
	//
	// By default, this parameter is disabled.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`
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

	// DBClusterIdentifierRef is a reference to a DBCluster used to set
	// DBClusterIdentifier.
	// +immutable
	// +optional
	DBClusterIdentifierRef *xpv1.Reference `json:"dbClusterIdentifierRef,omitempty"`

	// DBClusterIdentifierSelector selects a reference to a DBCluster used to
	// set DBClusterIdentifier.
	// +immutable
	// +optional
	DBClusterIdentifierSelector *xpv1.Selector `json:"dbClusterIdentifierSelector,omitempty"`

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

	// DBParameterGroupNameRef is a reference to a DBParameterGroup used to set
	// DBParameterGroupName.
	// +optional
	DBParameterGroupNameRef *xpv1.Reference `json:"dbParameterGroupNameRef,omitempty"`

	// DBParameterGroupNameSelector selects a reference to a DBParameterGroup used to
	// set DBParameterGroupName.
	// +optional
	DBParameterGroupNameSelector *xpv1.Selector `json:"dbParameterGroupNameSelector,omitempty"`

	// A value that indicates whether the modifications in this request and any
	// pending modifications are asynchronously applied as soon as possible, regardless
	// of the PreferredMaintenanceWindow setting for the DB instance. By default,
	// this parameter is disabled.
	//
	// If this parameter is disabled, changes to the DB instance are applied during
	// the next maintenance window. Some parameter changes can cause an outage and
	// are applied on the next call to RebootDBInstance, or the next failure reboot.
	// Review the table of parameters in Modifying a DB Instance (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
	// in the Amazon RDS User Guide. to see the impact of enabling or disabling
	// ApplyImmediately for each modified parameter and to determine when the changes
	// are applied.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`
}

// CustomDBInstanceRoleAssociationParameters are custom parameters for the DBInstanceRoleAssociation
type CustomDBInstanceRoleAssociationParameters struct {
	// The name of the DB instance to associate the IAM role with.
	// +crossplane:generate:reference:type=DBInstance
	// +optional
	DBInstanceIdentifier *string `json:"dbInstanceIdentifier,omitempty"`

	// DBInstanceIdentifierRef is a reference to a DBInstance used to set
	// the DBInstanceIdentifier.
	// +optional
	DBInstanceIdentifierRef *xpv1.Reference `json:"dbInstanceIdentifierRef,omitempty"`

	// DBInstanceIdentifierSelector selects references to a DBInstance used
	// to set the DBInstanceIdentifier.
	// +optional
	DBInstanceIdentifierSelector *xpv1.Selector `json:"dbInstanceIdentifierSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role to associate with the DB instance,
	// for example arn:aws:iam::123456789012:role/AccessRole.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane/provider-aws/apis/iam/v1beta1.RoleARN()
	// +optional
	RoleARN *string `json:"roleArn,omitempty"`

	// RoleARNRef is a reference to a IAM Role used to set
	// RoleARN.
	// +optional
	RoleARNRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleARNSelector selects a reference to a IAM Role used to
	// set RoleARN.
	// +optional
	RoleARNSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`
}
