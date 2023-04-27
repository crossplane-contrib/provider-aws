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

// Code generated by angryjet. DO NOT EDIT.

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// GetCondition of this JobRun.
func (mg *JobRun) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this JobRun.
func (mg *JobRun) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicy of this JobRun.
func (mg *JobRun) GetManagementPolicy() xpv1.ManagementPolicy {
	return mg.Spec.ManagementPolicy
}

// GetProviderConfigReference of this JobRun.
func (mg *JobRun) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this JobRun.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *JobRun) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this JobRun.
func (mg *JobRun) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this JobRun.
func (mg *JobRun) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this JobRun.
func (mg *JobRun) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this JobRun.
func (mg *JobRun) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicy of this JobRun.
func (mg *JobRun) SetManagementPolicy(r xpv1.ManagementPolicy) {
	mg.Spec.ManagementPolicy = r
}

// SetProviderConfigReference of this JobRun.
func (mg *JobRun) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this JobRun.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *JobRun) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this JobRun.
func (mg *JobRun) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this JobRun.
func (mg *JobRun) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetCondition of this VirtualCluster.
func (mg *VirtualCluster) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this VirtualCluster.
func (mg *VirtualCluster) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicy of this VirtualCluster.
func (mg *VirtualCluster) GetManagementPolicy() xpv1.ManagementPolicy {
	return mg.Spec.ManagementPolicy
}

// GetProviderConfigReference of this VirtualCluster.
func (mg *VirtualCluster) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

/*
GetProviderReference of this VirtualCluster.
Deprecated: Use GetProviderConfigReference.
*/
func (mg *VirtualCluster) GetProviderReference() *xpv1.Reference {
	return mg.Spec.ProviderReference
}

// GetPublishConnectionDetailsTo of this VirtualCluster.
func (mg *VirtualCluster) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this VirtualCluster.
func (mg *VirtualCluster) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this VirtualCluster.
func (mg *VirtualCluster) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this VirtualCluster.
func (mg *VirtualCluster) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicy of this VirtualCluster.
func (mg *VirtualCluster) SetManagementPolicy(r xpv1.ManagementPolicy) {
	mg.Spec.ManagementPolicy = r
}

// SetProviderConfigReference of this VirtualCluster.
func (mg *VirtualCluster) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

/*
SetProviderReference of this VirtualCluster.
Deprecated: Use SetProviderConfigReference.
*/
func (mg *VirtualCluster) SetProviderReference(r *xpv1.Reference) {
	mg.Spec.ProviderReference = r
}

// SetPublishConnectionDetailsTo of this VirtualCluster.
func (mg *VirtualCluster) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this VirtualCluster.
func (mg *VirtualCluster) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}
