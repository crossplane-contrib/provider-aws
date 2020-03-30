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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"

	network "github.com/crossplane/provider-aws/apis/network/v1alpha3"
	aws "github.com/crossplane/provider-aws/pkg/clients"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Error strings
const (
	errResourceIsNotDBSubnetGroup = "the managed resource is not a DBSubnetGroup"
)

// DBSubnetGroup states.
const (
	// The DB subnet group is healthy and available
	DBSubnetGroupStateAvailable = "available"
	// The DB subnet group is being created. The DB subnet group is inaccessible while it is being created.
	DBSubnetGroupStateCreating = "creating"
	// The DB subnet group is being deleted.
	DBSubnetGroupStateDeleting = "deleting"
	// The DB subnet group is being modified.
	DBSubnetGroupStateModifying = "modifying"
	// The DB subnet group has failed and Amazon RDS can't recover it. Perform a point-in-time restore to the latest restorable time of the instance to recover the data.
	DBSubnetGroupStateFailed = "failed"
)

// Subnet represents a aws subnet
type Subnet struct {
	// Specifies the identifier of the subnet.
	SubnetID string `json:"subnetID"`

	// Specifies the status of the subnet.
	SubnetStatus string `json:"subnetStatus"`
}

// SubnetIDReferencerForDBSubnetGroup is an attribute referencer that resolves SubnetID from a referenced Subnet
type SubnetIDReferencerForDBSubnetGroup struct {
	network.SubnetIDReferencer `json:",inline"`
}

// Assign assigns the retrieved value to the managed resource
func (v *SubnetIDReferencerForDBSubnetGroup) Assign(res resource.CanReference, value string) error {
	sg, ok := res.(*DBSubnetGroup)
	if !ok {
		return errors.New(errResourceIsNotDBSubnetGroup)
	}

	for _, id := range sg.Spec.ForProvider.SubnetIDs {
		if id == value {
			return nil
		}
	}

	sg.Spec.ForProvider.SubnetIDs = append(sg.Spec.ForProvider.SubnetIDs, value)
	return nil
}

var _ resource.AttributeReferencer = (*SubnetIDReferencerForDBSubnetGroup)(nil)

// DBSubnetGroupParameters define the desired state of an AWS VPC Database
// Subnet Group.
type DBSubnetGroupParameters struct {
	// The description for the DB subnet group.
	DBSubnetGroupDescription string `json:"description"`

	// The name for the DB subnet group. This value is stored as a lowercase string.
	// +immutable
	DBSubnetGroupName string `json:"groupName"`

	// The EC2 Subnet IDs for the DB subnet group.
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a set of referencers that each retrieve the subnetID from the referenced Subnet
	SubnetIDRefs []*SubnetIDReferencerForDBSubnetGroup `json:"subnetIdRefs,omitempty" resource:"attributereferencer"`

	// A list of tags. For more information, see Tagging Amazon RDS Resources (http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Tagging.html)
	// in the Amazon RDS User Guide.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// A DBSubnetGroupSpec defines the desired state of a DBSubnetGroup.
type DBSubnetGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  DBSubnetGroupParameters `json:"forProvider,omitempty"`
}

// DBSubnetGroupExternalStatus keeps the state for the external resource
type DBSubnetGroupExternalStatus struct {
	// The Amazon Resource Name (ARN) for the DB subnet group.
	DBSubnetGroupARN string `json:"groupArn,omitempty"`

	// Provides the status of the DB subnet group.
	SubnetGroupStatus string `json:"groupStatus,omitempty"`

	// Contains a list of Subnet elements.
	Subnets []Subnet `json:"subnets,omitempty"`

	// Provides the VpcId of the DB subnet group.
	VPCID string `json:"vpcId,omitempty"`
}

// A DBSubnetGroupStatus represents the observed state of a DBSubnetGroup.
type DBSubnetGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     DBSubnetGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DBSubnetGroup is a managed resource that represents an AWS VPC Database
// Subnet Group.
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupName"
// +kubebuilder:printcolumn:name="DESCRIPTION",type="string",JSONPath=".spec.forProvider.description"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.groupStatus"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
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

// DBSubnetGroupObservation is the representation of the current state that is observed
type DBSubnetGroupObservation struct {
	// SubnetGroupStatus specifies the current state of this DB subnet group.
	SubnetGroupStatus string `json:"subnetGroupStatus,omitempty"`

	// DBSubnetGroupArn is the Amazon Resource Name (ARN) for this DB subnet group.
	DBSubnetGroupArn string `json:"dbSubnetGroupArn,omitempty"`

	// Subnets contains a list of Subnet elements.
	Subnets []Subnet `json:"subnets,omitempty"`

	// VPCID provides the VPCID of the DB subnet group.
	VPCID string `json:"vpcId,omitempty"`
}

// UpdateExternalStatus updates the external status object, given the observation
func (b *DBSubnetGroup) UpdateExternalStatus(observation rds.DBSubnetGroup) {

	subnets := make([]Subnet, len(observation.Subnets))
	for i, sn := range observation.Subnets {
		subnets[i] = Subnet{
			SubnetID:     aws.StringValue(sn.SubnetIdentifier),
			SubnetStatus: aws.StringValue(sn.SubnetStatus),
		}
	}

	b.Status.AtProvider = DBSubnetGroupObservation{
		DBSubnetGroupArn:  aws.StringValue(observation.DBSubnetGroupArn),
		SubnetGroupStatus: aws.StringValue(observation.SubnetGroupStatus),
		Subnets:           subnets,
		VPCID:             aws.StringValue(observation.VpcId),
	}
}

// BuildFromRDSTags returns a list of tags, off of the given RDS tags
func BuildFromRDSTags(tags []rds.Tag) []Tag {
	res := make([]Tag, len(tags))
	for i, t := range tags {
		res[i] = Tag{aws.StringValue(t.Key), aws.StringValue(t.Value)}
	}

	return res
}
