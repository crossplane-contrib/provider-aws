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

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tag defines a key value pair that can be attached to an ELB
type Tag struct {

	// The key of the tag.
	Key string `json:"key"`

	// The value of the tag.
	// +optional
	Value *string `json:"value,omitempty"`
}

// Listener represents the port binding(s) between the ELB and EC2 instances.
type Listener struct {

	// The port on which the instance is listening.
	InstancePort int32 `json:"instancePort"`

	// The protocol to use for routing traffic to instances: HTTP, HTTPS, TCP, or
	// SSL. If not specified, the value is same as for Protocol.
	// +optional
	InstanceProtocol *string `json:"instanceProtocol,omitempty"`

	// The port on which the load balancer is listening.
	LoadBalancerPort int32 `json:"loadBalancerPort"`

	// The load balancer transport protocol to use for routing: HTTP, HTTPS, TCP,
	// or SSL.
	Protocol string `json:"protocol"`

	// The Amazon Resource Name (ARN) of the server certificate.
	// +optional
	SSLCertificateID *string `json:"sslCertificateId,omitempty"`
}

// BackendServerDescription provides information about the instances attached to the ELB.
type BackendServerDescription struct {

	// The port on which the EC2 instance is listening.
	InstancePort int32 `json:"instancePort,omitempty"`

	// The names of the policies enabled for the EC2 instance.
	PolicyNames []string `json:"policyNames,omitempty"`
}

// HealthCheck defines the rules that the ELB uses to decide if an attached instance is healthy.
type HealthCheck struct {

	// The number of consecutive health checks successes required before moving
	// the instance to the Healthy state.
	HealthyThreshold int32 `json:"healthyThreshold"`

	// The approximate interval, in seconds, between health checks of an individual
	// instance.
	Interval int32 `json:"interval"`

	// The instance being checked.
	Target string `json:"target"`

	// The amount of time, in seconds, during which no response means a failed health
	// check.
	Timeout int32 `json:"timeout"`

	// The number of consecutive health check failures required before moving the
	// instance to the Unhealthy state.
	UnhealthyThreshold int32 `json:"unhealthyThreshold"`
}

// ELBParameters define the desired state of an AWS ELB.
type ELBParameters struct {
	// Region is the region you'd like your ELB to be created in.
	Region string `json:"region"`

	// One or more Availability Zones from the same region as the load balancer.
	// +optional
	AvailabilityZones []string `json:"availabilityZones,omitempty"`

	// Information about the health checks conducted on the load balancer.
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`

	// The listeners for this ELB.
	Listeners []Listener `json:"listeners"`

	// The type of a load balancer. Valid only for load balancers in a VPC.
	// +optional
	// +immutable
	Scheme *string `json:"scheme,omitempty"`

	// The IDs of the security groups to assign to the load balancer.
	// +optional
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs references to a SecurityGroup and retrieves its SecurityGroupID
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDSelector selects a set of references that each retrieve the SecurityGroupID from the referenced SecurityGroup
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// The IDs of the subnets in your VPC to attach to the load balancer. Specify
	// one subnet per Availability Zone specified in AvailabilityZones.
	// +optional
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetRefs references to a Subnet to and retrieves its SubnetID
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetSelector selects a set of references that each retrieve the subnetID from the referenced Subnet
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// A list of tags to assign to the load balancer.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// An ELBSpec defines the desired state of an ELB.
type ELBSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ELBParameters `json:"forProvider"`
}

// ELBObservation keeps the state for the external resource
type ELBObservation struct {
	// Information about the EC2 instances for this ELB.
	BackendServerDescriptions []BackendServerDescription `json:"backendServerDescriptions,omitempty"`

	// The DNS name of the load balancer.
	CanonicalHostedZoneName string `json:"canonicalHostedZoneName,omitempty"`

	// The ID of the Amazon Route 53 hosted zone for the load balancer.
	CanonicalHostedZoneNameID string `json:"canonicalHostedZoneNameId,omitempty"`

	// The DNS name of the load balancer.
	DNSName string `json:"dnsName,omitempty"`

	// The ID of the VPC for the load balancer.
	VPCID string `json:"vpcId,omitempty"`
}

// An ELBStatus represents the observed state of an ELB.
type ELBStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ELBObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An ELB is a managed resource that represents an AWS Classic Load Balancer.
// +kubebuilder:printcolumn:name="ELBNAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="DNSNAME",type="string",JSONPath=".status.atProvider.dnsName"
// +kubebuilder:printcolumn:name="VPCID",type="string",JSONPath=".status.atProvider.vpcId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ELB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ELBSpec   `json:"spec"`
	Status ELBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ELBList contains a list of ELBs
type ELBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ELB `json:"items"`
}
