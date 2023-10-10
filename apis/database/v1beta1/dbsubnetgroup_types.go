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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DBSubnetGroupStateAvailable states that a DBSubnet Group is healthy and available
const DBSubnetGroupStateAvailable = "Complete"

// Subnet represents a aws subnet
type Subnet struct {
	// Specifies the identifier of the subnet.
	SubnetID string `json:"subnetID"`

	// Specifies the status of the subnet.
	SubnetStatus string `json:"subnetStatus"`
}

// DBSubnetGroupParameters define the desired state of an AWS VPC Database
// Subnet Group.
type DBSubnetGroupParameters struct {
	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your DBSubnetGroup to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// The description for the DB subnet group.
	Description string `json:"description"`

	// The EC2 Subnet IDs for the DB subnet group.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a set of references that each retrieve the subnetID from the referenced Subnet
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects a set of references that each retrieve the subnetID from the referenced Subnet
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// A list of tags. For more information, see Tagging Amazon RDS Resources (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// A DBSubnetGroupSpec defines the desired state of a DBSubnetGroup.
type DBSubnetGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DBSubnetGroupParameters `json:"forProvider,omitempty"`
}

// DBSubnetGroupObservation is the representation of the current state that is observed
type DBSubnetGroupObservation struct {
	// State specifies the current state of this DB subnet group.
	State string `json:"state,omitempty"`

	// ARN is the Amazon Resource Name (ARN) for this DB subnet group.
	ARN string `json:"arn,omitempty"`

	// Subnets contains a list of Subnet elements.
	Subnets []Subnet `json:"subnets,omitempty"`

	// VPCID provides the VPCID of the DB subnet group.
	VPCID string `json:"vpcId,omitempty"`
}

// A DBSubnetGroupStatus represents the observed state of a DBSubnetGroup.
type DBSubnetGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DBSubnetGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DBSubnetGroup is a managed resource that represents an AWS VPC Database
// Subnet Group.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type DBSubnetGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBSubnetGroupSpec   `json:"spec"`
	Status DBSubnetGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DBSubnetGroupList contains a list of DBSubnetGroups
type DBSubnetGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBSubnetGroup `json:"items"`
}
