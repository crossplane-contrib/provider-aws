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

// Code generated by angryjet. DO NOT EDIT.

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GetBindingPhase of this SNSSubscription.
func (mg *SNSSubscription) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return mg.Status.GetBindingPhase()
}

// GetClaimReference of this SNSSubscription.
func (mg *SNSSubscription) GetClaimReference() *corev1.ObjectReference {
	return mg.Spec.ClaimReference
}

// GetClassReference of this SNSSubscription.
func (mg *SNSSubscription) GetClassReference() *corev1.ObjectReference {
	return mg.Spec.ClassReference
}

// GetCondition of this SNSSubscription.
func (mg *SNSSubscription) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetProviderReference of this SNSSubscription.
func (mg *SNSSubscription) GetProviderReference() *corev1.ObjectReference {
	return mg.Spec.ProviderReference
}

// GetReclaimPolicy of this SNSSubscription.
func (mg *SNSSubscription) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return mg.Spec.ReclaimPolicy
}

// GetWriteConnectionSecretToReference of this SNSSubscription.
func (mg *SNSSubscription) GetWriteConnectionSecretToReference() *runtimev1alpha1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetBindingPhase of this SNSSubscription.
func (mg *SNSSubscription) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	mg.Status.SetBindingPhase(p)
}

// SetClaimReference of this SNSSubscription.
func (mg *SNSSubscription) SetClaimReference(r *corev1.ObjectReference) {
	mg.Spec.ClaimReference = r
}

// SetClassReference of this SNSSubscription.
func (mg *SNSSubscription) SetClassReference(r *corev1.ObjectReference) {
	mg.Spec.ClassReference = r
}

// SetConditions of this SNSSubscription.
func (mg *SNSSubscription) SetConditions(c ...runtimev1alpha1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetProviderReference of this SNSSubscription.
func (mg *SNSSubscription) SetProviderReference(r *corev1.ObjectReference) {
	mg.Spec.ProviderReference = r
}

// SetReclaimPolicy of this SNSSubscription.
func (mg *SNSSubscription) SetReclaimPolicy(r runtimev1alpha1.ReclaimPolicy) {
	mg.Spec.ReclaimPolicy = r
}

// SetWriteConnectionSecretToReference of this SNSSubscription.
func (mg *SNSSubscription) SetWriteConnectionSecretToReference(r *runtimev1alpha1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetBindingPhase of this SNSTopic.
func (mg *SNSTopic) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return mg.Status.GetBindingPhase()
}

// GetClaimReference of this SNSTopic.
func (mg *SNSTopic) GetClaimReference() *corev1.ObjectReference {
	return mg.Spec.ClaimReference
}

// GetClassReference of this SNSTopic.
func (mg *SNSTopic) GetClassReference() *corev1.ObjectReference {
	return mg.Spec.ClassReference
}

// GetCondition of this SNSTopic.
func (mg *SNSTopic) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetProviderReference of this SNSTopic.
func (mg *SNSTopic) GetProviderReference() *corev1.ObjectReference {
	return mg.Spec.ProviderReference
}

// GetReclaimPolicy of this SNSTopic.
func (mg *SNSTopic) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return mg.Spec.ReclaimPolicy
}

// GetWriteConnectionSecretToReference of this SNSTopic.
func (mg *SNSTopic) GetWriteConnectionSecretToReference() *runtimev1alpha1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetBindingPhase of this SNSTopic.
func (mg *SNSTopic) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	mg.Status.SetBindingPhase(p)
}

// SetClaimReference of this SNSTopic.
func (mg *SNSTopic) SetClaimReference(r *corev1.ObjectReference) {
	mg.Spec.ClaimReference = r
}

// SetClassReference of this SNSTopic.
func (mg *SNSTopic) SetClassReference(r *corev1.ObjectReference) {
	mg.Spec.ClassReference = r
}

// SetConditions of this SNSTopic.
func (mg *SNSTopic) SetConditions(c ...runtimev1alpha1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetProviderReference of this SNSTopic.
func (mg *SNSTopic) SetProviderReference(r *corev1.ObjectReference) {
	mg.Spec.ProviderReference = r
}

// SetReclaimPolicy of this SNSTopic.
func (mg *SNSTopic) SetReclaimPolicy(r runtimev1alpha1.ReclaimPolicy) {
	mg.Spec.ReclaimPolicy = r
}

// SetWriteConnectionSecretToReference of this SNSTopic.
func (mg *SNSTopic) SetWriteConnectionSecretToReference(r *runtimev1alpha1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}
