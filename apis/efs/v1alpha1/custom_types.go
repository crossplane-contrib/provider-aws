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

const (
	// ResourceCredentialsSecretIDKey is the name of the key in the connection
	// secret for FileSystem ID.
	ResourceCredentialsSecretIDKey = "id"
)

// CustomAccessPointParameters contains the additional fields for AccessPointParameters.
type CustomAccessPointParameters struct {
	// The ID of the file system for which to create the mount target.
	// +immutable
	// +optional
	FileSystemID *string `json:"fileSystemID,omitempty"`

	// FileSystemIDRef are references to Filesystem used to set
	// the FileSystemID.
	// +immutable
	// +optional
	FileSystemIDRef *xpv1.Reference `json:"fileSystemIDRef,omitempty"`

	// FileSystemIDSelector selects references to Filesystem used
	// to set the FileSystemID.
	// +immutable
	// +optional
	FileSystemIDSelector *xpv1.Selector `json:"fileSystemIDSelector,omitempty"`
}

// CustomAccessPointObservation includes the custom status fields of AccessPoint.
type CustomAccessPointObservation struct{}

// CustomFileSystemParameters contains the additional fields for FileSystemParameters.
type CustomFileSystemParameters struct {

	// The throughput, measured in MiB/s, that you want to provision for a file
	// system that you're creating. Valid values are 1-1024. Required if ThroughputMode
	// is set to provisioned. The upper limit for throughput is 1024 MiB/s. You
	// can get this limit increased by contacting AWS Support. For more information,
	// see Amazon EFS Limits That You Can Increase (https://docs.aws.amazon.com/efs/latest/ug/limits.html#soft-limits)
	// in the Amazon EFS User Guide.
	// +optional
	ProvisionedThroughputInMibps *int64 `json:"provisionedThroughputInMibps,omitempty"`

	// KMSKeyIDRef is a reference to an Key used to set
	// the KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

	// KMSKeyIDSelector selects references to Key used
	// to set the KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`
}

// CustomFileSystemObservation includes the custom status fields of FileSystem.
type CustomFileSystemObservation struct{}

// CustomMountTargetParameters contains the additional fields for MountTargetParameters.
type CustomMountTargetParameters struct {

	// Up to five VPC security group IDs, of the form sg-xxxxxxxx. These must be
	// for the same VPC as subnet specified.
	// +immutable
	// +optional
	SecurityGroups []string `json:"securityGroups,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupsRefs []xpv1.Reference `json:"securityGroupsRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used
	// to set the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupsSelector *xpv1.Selector `json:"securityGroupsSelector,omitempty"`

	// The ID of the file system for which to create the mount target.
	// +immutable
	// +optional
	FileSystemID *string `json:"fileSystemID,omitempty"`

	// FileSystemIDRef are references to Filesystem used to set
	// the FileSystemID.
	// +immutable
	// +optional
	FileSystemIDRef *xpv1.Reference `json:"fileSystemIDRef,omitempty"`

	// FileSystemIDSelector selects references to Filesystem used
	// to set the FileSystemID.
	// +immutable
	// +optional
	FileSystemIDSelector *xpv1.Selector `json:"fileSystemIDSelector,omitempty"`

	// The ID of the subnet to add the mount target in.
	// +immutable
	// +optional
	SubnetID *string `json:"subnetID"`

	// SubnetIDRef are references to Subnet used to set
	// the SubnetID.
	// +immutable
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIDRef,omitempty"`

	// SubnetIDSelector selects references to Subnet used
	// to set the SubnetID.
	// +immutable
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`
}

// CustomMountTargetObservation includes the custom status fields of MountTarget.
type CustomMountTargetObservation struct{}
