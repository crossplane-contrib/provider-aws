/*
Copyright 2022 The Crossplane Authors.

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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Defines the possible state of a target.
const (
	TargetStatusHealthy     = "healthy"
	TargetStatusInitial     = "initial"
	TargetStatusUnhealthy   = "unhealthy"
	TargetStatusUnused      = "unused"
	TargetStatusDraining    = "draining"
	TargetStatusUnavailable = "unavailable"
)

// TargetParameters defines the desired state of a
// Target
type TargetParameters struct {
	// The AWS region the target resides in.
	Region string `json:"region"`

	// The Amazon Resource Name (ARN) of the target group.
	//
	// One of TargetGroupARN, TargetGroupARNRef or TargetGroupARNSelector is
	// required.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1.TargetGroup
	// +immutable
	TargetGroupARN *string `json:"targetGroupArn,omitempty"`

	// TargetGroupARNRef selects a ELBv2 TargetGroupARN with the given name.
	TargetGroupARNRef *xpv1.Reference `json:"targetGroupArnRef,omitempty"`

	// TargetGroupARNSelector selects a ELBv2 TargetGroupARN with the given
	// labels.
	TargetGroupARNSelector *xpv1.Selector `json:"targetGroupArnSelector,omitempty"`

	// The LambdaARN that should be used as target.
	//
	// Note: If you want to reference anything else than Lambda you currently
	// have to specify the crossplane.io/external-name annotation directly.
	// If the target type of the target group is instance,
	// specify an instance ID. If the target type is ip, specify an IP address.
	// If the target type is lambda, specify the ARN of the Lambda function. If
	// the target type is alb, specify the ARN of the Application Load Balancer
	// target.
	//
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1.Function
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1.FunctionARN()
	LambdaARN *string `json:"lambdaArn,omitempty"`

	// LambdaARNRef references a Lambda Function to set LambdaARN.
	LambdaARNRef *xpv1.Reference `json:"lambdaArnRef,omitempty"`

	// LambdaARNSelector references a Lambda Function to set LambdaARN.
	LambdaARNSelector *xpv1.Selector `json:"lambdaArnSelector,omitempty"`

	// The port on which the target is listening. If the target group protocol is
	// GENEVE, the supported port is 6081. If the target type is alb, the targeted
	// Application Load Balancer must have at least one listener whose port matches
	// the target group port. Not used if the target is a Lambda function.
	// +immutable
	Port *int32 `json:"port,omitempty"`

	// An Availability Zone or all. This determines whether the target receives
	// traffic from the load balancer nodes in the specified Availability Zone or
	// from all enabled Availability Zones for the load balancer.
	//
	// This parameter is not supported if the target type of the target group is
	// instance or alb.
	//
	// If the target type is ip and the IP address is in a subnet of the VPC for
	// the target group, the Availability Zone is automatically detected and this
	// parameter is optional. If the IP address is outside the VPC, this parameter
	// is required.
	//
	// With an Application Load Balancer, if the target type is ip and the IP address
	// is outside the VPC for the target group, the only supported value is all.
	//
	// If the target type is lambda, this parameter is optional and the only supported
	// value is all.
	// +immutable
	AvailabilityZone *string `json:"availabilityZone,omitempty"`
}

// TargetSpec defines the desired state of a Target
type TargetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TargetParameters `json:"forProvider"`
}

// TargetHealth describes the health state of a Target.
type TargetHealth struct {
	// A description of the target health that provides additional details. If the
	// state is healthy, a description is not provided.
	Description *string `json:"description,omitempty"`

	// The reason code.
	//
	// If the target state is healthy, a reason code is not provided.
	//
	// If the target state is initial, the reason code can be one of the following
	// values:
	//
	//    * Elb.RegistrationInProgress - The target is in the process of being registered
	//    with the load balancer.
	//
	//    * Elb.InitialHealthChecking - The load balancer is still sending the target
	//    the minimum number of health checks required to determine its health status.
	//
	// If the target state is unhealthy, the reason code can be one of the following
	// values:
	//
	//    * Target.ResponseCodeMismatch - The health checks did not return an expected
	//    HTTP code. Applies only to Application Load Balancers and Gateway Load
	//    Balancers.
	//
	//    * Target.Timeout - The health check requests timed out. Applies only to
	//    Application Load Balancers and Gateway Load Balancers.
	//
	//    * Target.FailedHealthChecks - The load balancer received an error while
	//    establishing a connection to the target or the target response was malformed.
	//
	//    * Elb.InternalError - The health checks failed due to an internal error.
	//    Applies only to Application Load Balancers.
	//
	// If the target state is unused, the reason code can be one of the following
	// values:
	//
	//    * Target.NotRegistered - The target is not registered with the target
	//    group.
	//
	//    * Target.NotInUse - The target group is not used by any load balancer
	//    or the target is in an Availability Zone that is not enabled for its load
	//    balancer.
	//
	//    * Target.InvalidState - The target is in the stopped or terminated state.
	//
	//    * Target.IpUnusable - The target IP address is reserved for use by a load
	//    balancer.
	//
	// If the target state is draining, the reason code can be the following value:
	//
	//    * Target.DeregistrationInProgress - The target is in the process of being
	//    deregistered and the deregistration delay period has not expired.
	//
	// If the target state is unavailable, the reason code can be the following
	// value:
	//
	//    * Target.HealthCheckDisabled - Health checks are disabled for the target
	//    group. Applies only to Application Load Balancers.
	//
	//    * Elb.InternalError - Target health is unavailable due to an internal
	//    error. Applies only to Network Load Balancers.
	Reason *string `json:"reason,omitempty"`

	// The state of the target.
	State *string `json:"state,omitempty"`
}

// TargetObservation defines the observed state of a
// Target
type TargetObservation struct {
	// The port to use to connect with the target.
	HealthCheckPort *string `json:"healthCheckPort,omitempty"`

	// The health information for the target.
	TargetHealth *TargetHealth `json:"targetHealth,omitempty"`
}

// GetState returns s.TargetHealth.State if it is not nil, otherwise an empty
// string.
func (s *TargetObservation) GetState() string {
	if s.TargetHealth != nil && s.TargetHealth.State != nil {
		return *s.TargetHealth.State
	}
	return ""
}

// GetReason returns s.TargetHealth.Reason if it is not nil, otherwise an empty
// string.
func (s *TargetObservation) GetReason() string {
	if s.TargetHealth != nil && s.TargetHealth.Reason != nil {
		return *s.TargetHealth.Reason
	}
	return ""
}

// TargetStatus defines the observed state of a
// Target
type TargetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TargetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Target is the Schema for registering a target to an
// ELBV2 TargetGroup.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.targetHealth.state"
// +kubebuilder:printcolumn:name="TARGET",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="GROUP",type="string",JSONPath=".spec.forProvider.targetGroupArn"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TargetSpec   `json:"spec"`
	Status            TargetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetList contains a list of Targets
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Target `json:"items"`
}
