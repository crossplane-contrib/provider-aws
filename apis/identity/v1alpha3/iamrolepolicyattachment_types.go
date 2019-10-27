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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/pkg/errors"

	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

// Error strings
const (
	errResourceIsNotIAMRolePolicyAttachment = "the managed resource is not an IAMRolePolicyAttachment"
)

// IAMRoleNameReferencerForIAMRolePolicyAttachment is an attribute referencer that retrieves Name from a referenced IAMRole
type IAMRoleNameReferencerForIAMRolePolicyAttachment struct {
	IAMRoleNameReferencer `json:",inline"`
}

// Assign assigns the retrieved name to the managed resource
func (v *IAMRoleNameReferencerForIAMRolePolicyAttachment) Assign(res resource.CanReference, value string) error {
	p, ok := res.(*IAMRolePolicyAttachment)
	if !ok {
		return errors.New(errResourceIsNotIAMRolePolicyAttachment)
	}

	p.Spec.RoleName = value
	return nil
}

// IAMRolePolicyAttachmentParameters define the desired state of an AWS IAM
// Role policy attachment.
type IAMRolePolicyAttachmentParameters struct {

	// PolicyARN is the Amazon Resource Name (ARN) of the IAM policy you want to
	// attach.
	PolicyARN string `json:"policyArn"`

	// RoleName presents the name of the IAM role.
	RoleName string `json:"roleName,omitempty"`

	// RoleNameRef references to an IAMRole to retrieve its Name
	RoleNameRef *IAMRoleNameReferencerForIAMRolePolicyAttachment `json:"roleNameRef,omitempty" resource:"attributereferencer"`
}

// An IAMRolePolicyAttachmentSpec defines the desired state of an
// IAMRolePolicyAttachment.
type IAMRolePolicyAttachmentSpec struct {
	runtimev1alpha1.ResourceSpec      `json:",inline"`
	IAMRolePolicyAttachmentParameters `json:",inline"`
}

// IAMRolePolicyAttachmentExternalStatus keeps the state for the external resource
type IAMRolePolicyAttachmentExternalStatus struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// An IAMRolePolicyAttachmentStatus represents the observed state of an
// IAMRolePolicyAttachment.
type IAMRolePolicyAttachmentStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	IAMRolePolicyAttachmentExternalStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An IAMRolePolicyAttachment is a managed resource that represents an AWS IAM
// Role policy attachment.
// +kubebuilder:printcolumn:name="ROLENAME",type="string",JSONPath=".spec.roleName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type IAMRolePolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMRolePolicyAttachmentSpec   `json:"spec,omitempty"`
	Status IAMRolePolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMRolePolicyAttachmentList contains a list of IAMRolePolicyAttachments
type IAMRolePolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMRolePolicyAttachment `json:"items"`
}

// UpdateExternalStatus updates the external status object, given the observation
func (r *IAMRolePolicyAttachment) UpdateExternalStatus(observation iam.AttachedPolicy) {
	r.Status.IAMRolePolicyAttachmentExternalStatus = IAMRolePolicyAttachmentExternalStatus{
		AttachedPolicyARN: aws.StringValue(observation.PolicyArn),
	}
}
