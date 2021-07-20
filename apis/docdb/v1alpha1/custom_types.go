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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// RDS instance states.
const (
	// The instance is healthy and available
	DocDBInstanceStateAvailable = "available"
	// The instance is being created. The instance is inaccessible while it is being created.
	DocDBInstanceStateCreating = "creating"
	// The instance is being deleted.
	DocDBInstanceStateDeleting = "deleting"
	// The instance is being modified.
	DocDBInstanceStateModifying = "modifying"
	// The instance has failed and Amazon RDS can't recover it. Perform a point-in-time restore to the latest restorable time of the instance to recover the data.
	DocDBInstanceStateFailed = "failed"
)

// CustomDBInstanceParameters for DBInstance
type CustomDBInstanceParameters struct {
	// Specifies whether the modifications in this request and any pending modifications
	// are asynchronously applied as soon as possible, regardless of the PreferredMaintenanceWindow
	// setting for the instance.
	//
	// If this parameter is set to false, changes to the instance are applied during
	// the next maintenance window. Some parameter changes can cause an outage and
	// are applied on the next reboot.
	//
	// Default: false
	// +optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// The identifier of the CA certificate for this DB instance.
	// +optional
	CACertificateIdentifier *string `json:"caCertificateIdentifier,omitempty"`
}

// CustomDBSubnetGroupParameters for DBSubnetGroupParameters
type CustomDBSubnetGroupParameters struct{}

// CustomDBClusterParameterGroupParameters for DBClusterParameterGroup
type CustomDBClusterParameterGroupParameters struct {
	Parameters []*Parameter `json:"parameters,omitempty"`
}

// CustomDBClusterParameters for DBCluster
type CustomDBClusterParameters struct {
	// A value that specifies whether the changes in this request and any pending
	// changes are asynchronously applied as soon as possible, regardless of the
	// PreferredMaintenanceWindow setting for the cluster. If this parameter is
	// set to false, changes to the cluster are applied during the next maintenance
	// window.
	//
	// The ApplyImmediately parameter affects only the NewDBClusterIdentifier and
	// MasterUserPassword values. If you set this parameter value to false, the
	// changes to the NewDBClusterIdentifier and MasterUserPassword values are applied
	// during the next maintenance window. All other changes are applied immediately,
	// regardless of the value of the ApplyImmediately parameter.
	//
	// Default: false
	// +optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Determines whether a final cluster snapshot is created before the cluster
	// is deleted. If true is specified, no cluster snapshot is created. If false
	// is specified, a cluster snapshot is created before the DB cluster is deleted.
	//
	// If SkipFinalSnapshot is false, you must specify a FinalDBSnapshotIdentifier
	// parameter.
	//
	// Default: false
	// +optional
	SkipFinalSnapshot *bool `json:"skipFinalSnapshot,omitempty"`

	// The cluster snapshot identifier of the new cluster snapshot created when
	// SkipFinalSnapshot is set to false.
	//
	// Specifying this parameter and also setting the SkipFinalShapshot parameter
	// to true results in an error.
	//
	// Constraints:
	//
	//    * Must be from 1 to 255 letters, numbers, or hyphens.
	//
	//    * The first character must be a letter.
	//
	//    * Cannot end with a hyphen or contain two consecutive hyphens.
	// +optional
	FinalDBSnapshotIdentifier *string `json:"finalDBSnapshotIdentifier,omitempty"`

	// MasterUserPasswordSecretRef references the secret that contains the password for the master database user. This password can contain any
	// printable ASCII character except forward slash (/), double quote ("), or
	// the "at" symbol (@).
	//
	// Constraints: Must contain from 8 to 100 characters.
	MasterUserPasswordSecretRef *xpv1.SecretKeySelector `json:"masterUserPasswordSecretRef,omitempty"`
}
