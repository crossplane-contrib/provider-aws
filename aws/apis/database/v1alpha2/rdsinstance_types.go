/*
Copyright 2019 The Crossplane Authors.

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

package v1alpha2

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SQL database engines.
const (
	MysqlEngine      = "mysql"
	PostgresqlEngine = "postgres"
)

// RDSInstanceParameters define the desired state of an AWS Relational Database
// Service instance.
type RDSInstanceParameters struct {
	// MasterUsername for this RDSInstance.
	MasterUsername string `json:"masterUsername"`

	// Engine for this RDSInstance - either mysql or postgres.
	// +kubebuilder:validation:Enum=mysql;postgres
	Engine string `json:"engine"`

	// EngineVersion for this RDS instance, for example "5.6".
	// +optional
	EngineVersion string `json:"engineVersion,omitempty"`

	// Class of this RDS instance, for example "db.t2.micro".
	Class string `json:"class"`

	// Size in GB of this RDS instance.
	Size int64 `json:"size"`

	// SubnetGroupName specifies a database subnet group for the RDS instance.
	// The new instance is created in the VPC associated with the DB subnet
	// group. If no DB subnet group is specified, then the instance is not
	// created in a VPC.
	// +optional
	SubnetGroupName string `json:"subnetGroupName,omitempty"`

	// SecurityGroups that will allow the RDS instance to be accessed over the network.
	// +optional
	SecurityGroups []string `json:"securityGroups,omitempty"`
}

// An RDSInstanceSpec defines the desired state of an RDSInstance.
type RDSInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	RDSInstanceParameters        `json:",inline"`
}

// RDSInstanceState represents the state of an RDS instance.
type RDSInstanceState string

// RDS instance states.
const (
	// The instance is healthy and available
	RDSInstanceStateAvailable RDSInstanceState = "available"
	// The instance is being created. The instance is inaccessible while it is being created.
	RDSInstanceStateCreating RDSInstanceState = "creating"
	// The instance is being deleted.
	RDSInstanceStateDeleting RDSInstanceState = "deleting"
	// The instance has failed and Amazon RDS can't recover it. Perform a point-in-time restore to the latest restorable time of the instance to recover the data.
	RDSInstanceStateFailed RDSInstanceState = "failed"
)

// An RDSInstanceStatus represents the observed state of an RDSInstance.
type RDSInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this RDS instance.
	State string `json:"state,omitempty"`

	// ProviderID is the AWS identifier for this RDS instance.
	ProviderID string `json:"providerID,omitempty"`

	// InstanceName of this RDS instance.
	InstanceName string `json:"instanceName,omitempty"`

	// Endpoint of this RDS instance.
	Endpoint string `json:"endpoint,omitempty"`
}

// +kubebuilder:object:root=true

// An RDSInstance is a managed resource that represents an AWS Relational
// Database Service instance.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type RDSInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDSInstanceSpec   `json:"spec,omitempty"`
	Status RDSInstanceStatus `json:"status,omitempty"`
}

// SetBindingPhase of this RDSInstance.
func (i *RDSInstance) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	i.Status.SetBindingPhase(p)
}

// GetBindingPhase of this RDSInstance.
func (i *RDSInstance) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return i.Status.GetBindingPhase()
}

// SetConditions of this RDSInstance.
func (i *RDSInstance) SetConditions(c ...runtimev1alpha1.Condition) {
	i.Status.SetConditions(c...)
}

// SetClaimReference of this RDSInstance.
func (i *RDSInstance) SetClaimReference(r *corev1.ObjectReference) {
	i.Spec.ClaimReference = r
}

// GetClaimReference of this RDSInstance.
func (i *RDSInstance) GetClaimReference() *corev1.ObjectReference {
	return i.Spec.ClaimReference
}

// SetNonPortableClassReference of this RDSInstance.
func (i *RDSInstance) SetNonPortableClassReference(r *corev1.ObjectReference) {
	i.Spec.NonPortableClassReference = r
}

// GetNonPortableClassReference of this RDSInstance.
func (i *RDSInstance) GetNonPortableClassReference() *corev1.ObjectReference {
	return i.Spec.NonPortableClassReference
}

// SetWriteConnectionSecretToReference of this RDSInstance.
func (i *RDSInstance) SetWriteConnectionSecretToReference(r corev1.LocalObjectReference) {
	i.Spec.WriteConnectionSecretToReference = r
}

// GetWriteConnectionSecretToReference of this RDSInstance.
func (i *RDSInstance) GetWriteConnectionSecretToReference() corev1.LocalObjectReference {
	return i.Spec.WriteConnectionSecretToReference
}

// GetReclaimPolicy of this RDSInstance.
func (i *RDSInstance) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return i.Spec.ReclaimPolicy
}

// SetReclaimPolicy of this RDSInstance.
func (i *RDSInstance) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	i.Spec.ReclaimPolicy = p
}

// +kubebuilder:object:root=true

// RDSInstanceList contains a list of RDSInstance
type RDSInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDSInstance `json:"items"`
}

// An RDSInstanceClassSpecTemplate is a template for the spec of a dynamically
// provisioned RDSInstance.
type RDSInstanceClassSpecTemplate struct {
	runtimev1alpha1.NonPortableClassSpecTemplate `json:",inline"`
	RDSInstanceParameters                        `json:",inline"`
}

// All non-portable classes must implement the NonPortableClass interface.
var _ resource.NonPortableClass = &RDSInstanceClass{}

// +kubebuilder:object:root=true

// An RDSInstanceClass is a non-portable resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type RDSInstanceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// RDSInstance.
	SpecTemplate RDSInstanceClassSpecTemplate `json:"specTemplate"`
}

// GetReclaimPolicy of this RDSInstanceClass.
func (i *RDSInstanceClass) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return i.SpecTemplate.ReclaimPolicy
}

// SetReclaimPolicy of this RDSInstanceClass.
func (i *RDSInstanceClass) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	i.SpecTemplate.ReclaimPolicy = p
}

// +kubebuilder:object:root=true

// RDSInstanceClassList contains a list of cloud memorystore resource classes.
type RDSInstanceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDSInstanceClass `json:"items"`
}
