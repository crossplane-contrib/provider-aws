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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"

	aws "github.com/crossplane/provider-aws/pkg/clients"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// Error strings
const (
	errResourceIsNotSecurityGroup = "the managed resource is not a SecurityGroup"
)

// VPCIDReferencerForSecurityGroup is an attribute referencer that resolves VPCID from a referenced VPC
type VPCIDReferencerForSecurityGroup struct {
	VPCIDReferencer `json:",inline"`
}

// Assign assigns the retrieved value to the managed resource
func (v *VPCIDReferencerForSecurityGroup) Assign(res resource.CanReference, value string) error {
	sg, ok := res.(*SecurityGroup)
	if !ok {
		return errors.New(errResourceIsNotSecurityGroup)
	}

	sg.Spec.VPCID = aws.String(value)
	return nil
}

// SecurityGroupParameters define the desired state of an AWS VPC Security
// Group.
type SecurityGroupParameters struct {
	// VPCID is the ID of the VPC.
	// +optional
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references to a VPC to and retrieves its vpcId
	// +optional
	VPCIDRef *VPCIDReferencerForSecurityGroup `json:"vpcIdRef,omitempty"`

	// A description of the security group.
	Description string `json:"description"`

	// The name of the security group.
	GroupName string `json:"groupName"`

	// One or more inbound rules associated with the security group.
	// +optional
	Ingress []IPPermission `json:"ingress,omitempty"`

	// [EC2-VPC] One or more outbound rules associated with the security group.
	// +optional
	Egress []IPPermission `json:"egress,omitempty"`
}

// IPRange describes an IPv4 range.
type IPRange struct {
	// The IPv4 CIDR range. You can either specify a CIDR range or a source security
	// group, not both. To specify a single IPv4 address, use the /32 prefix length.
	CIDRIP string `json:"cidrIp"`

	// A description for the security group rule that references this IPv4 address
	// range.
	//
	// Constraints: Up to 255 characters in length. Allowed characters are a-z,
	// A-Z, 0-9, spaces, and ._-:/()#,@[]+=&;{}!$*
	// +optional
	Description *string `json:"description,omitempty"`
}

// IPv6Range describes an IPv6 range.
type IPv6Range struct {
	// The IPv6 CIDR range. You can either specify a CIDR range or a source security
	// group, not both. To specify a single IPv6 address, use the /128 prefix length.
	CIDRIPv6 string `json:"cidrIPv6"`

	// A description for the security group rule that references this IPv6 address
	// range.
	//
	// Constraints: Up to 255 characters in length. Allowed characters are a-z,
	// A-Z, 0-9, spaces, and ._-:/()#,@[]+=&;{}!$*
	// +optional
	Description *string `json:"description,omitempty"`
}

// PrefixListID describes a prefix list ID.
type PrefixListID struct {
	// A description for the security group rule that references this prefix list
	// ID.
	//
	// Constraints: Up to 255 characters in length. Allowed characters are a-z,
	// A-Z, 0-9, spaces, and ._-:/()#,@[]+=;{}!$*
	// +optional
	Description *string `json:"description,omitempty"`

	// The ID of the prefix.
	PrefixListID string `json:"prefixListId"`
}

// UserIDGroupPair describes a security group and AWS account ID pair.
type UserIDGroupPair struct {
	// A description for the security group rule that references this user ID group
	// pair.
	//
	// Constraints: Up to 255 characters in length. Allowed characters are a-z,
	// A-Z, 0-9, spaces, and ._-:/()#,@[]+=;{}!$*
	// +optional
	Description *string `json:"description,omitempty"`

	// The ID of the security group.
	// +optional
	GroupID *string `json:"groupId,omitempty"`

	// The name of the security group. In a request, use this parameter for a security
	// group in EC2-Classic or a default VPC only. For a security group in a nondefault
	// VPC, use the security group ID.
	//
	// For a referenced security group in another VPC, this value is not returned
	// if the referenced security group is deleted.
	// +optional
	GroupName *string `json:"groupName,omitempty"`

	// The ID of an AWS account.
	//
	// For a referenced security group in another VPC, the account ID of the referenced
	// security group is returned in the response. If the referenced security group
	// is deleted, this value is not returned.
	//
	// [EC2-Classic] Required when adding or removing rules that reference a security
	// group in another AWS account.
	// +optional
	UserID *string `json:"userId,omitempty"`

	// The ID of the VPC for the referenced security group, if applicable.
	// +optional
	VPCID *string `json:"vpcId,omitempty"`

	// The ID of the VPC peering connection, if applicable.
	// +optional
	VPCPeeringConnectionID *string `json:"vpcPeeringConnectionId,omitempty"`
}

// IPPermission Describes a set of permissions for a security group rule.
type IPPermission struct {
	// The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6
	// type number. A value of -1 indicates all ICMP/ICMPv6 types. If you specify
	// all ICMP/ICMPv6 types, you must specify all codes.
	// +optional
	FromPort *int64 `json:"fromPort,omitempty"`

	// TODO(muvaf): jsontag of this should be ipProtocol when we make this resource
	// v1beta1

	// The IP protocol name (tcp, udp, icmp, icmpv6) or number (see Protocol Numbers
	// (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)).
	//
	// [VPC only] Use -1 to specify all protocols. When authorizing security group
	// rules, specifying -1 or a protocol number other than tcp, udp, icmp, or icmpv6
	// allows traffic on all ports, regardless of any port range you specify. For
	// tcp, udp, and icmp, you must specify a port range. For icmpv6, the port range
	// is optional; if you omit the port range, traffic for all types and codes
	// is allowed.
	IPProtocol string `json:"protocol"`

	// TODO(muvaf): The jsontag should be ipRanges instead of cidrBlocks, do that
	// when we bump the version.

	// The IPv4 ranges.
	// +optional
	IPRanges []IPRange `json:"cidrBlocks,omitempty"`

	// The IPv6 ranges.
	//
	// [VPC only]
	// +optional
	IPv6Ranges []IPv6Range `json:"ipv6Ranges,omitempty"`

	// PrefixListIDs for an AWS service. With outbound rules, this
	// is the AWS service to access through a VPC endpoint from instances associated
	// with the security group.
	//
	// [VPC only]
	// +optional
	PrefixListIDs []PrefixListID `json:"prefixListIds,omitempty"`

	// The end of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code.
	// A value of -1 indicates all ICMP/ICMPv6 codes. If you specify all ICMP/ICMPv6
	// types, you must specify all codes.
	// +optional
	ToPort *int64 `json:"toPort,omitempty"`

	// UserIDGroupPairs are the source security group and AWS account ID pairs.
	// It contains one or more accounts and security groups to allow flows from
	// security groups of other accounts.
	// +optional
	UserIDGroupPairs []UserIDGroupPair `json:"userIdGroupPairs,omitempty"`
}

// A SecurityGroupSpec defines the desired state of a SecurityGroup.
type SecurityGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	SecurityGroupParameters      `json:",inline"`
}

// SecurityGroupExternalStatus keeps the state for the external resource
type SecurityGroupExternalStatus struct {
	// SecurityGroupID is the ID of the SecurityGroup.
	SecurityGroupID string `json:"securityGroupID"`

	// Tags represents to current ec2 tags.
	Tags []Tag `json:"tags,omitempty"`
}

// A SecurityGroupStatus represents the observed state of a SecurityGroup.
type SecurityGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	SecurityGroupExternalStatus    `json:",inline"`
}

// +kubebuilder:object:root=true

// A SecurityGroup is a managed resource that represents an AWS VPC Security
// Group.
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.groupName"
// +kubebuilder:printcolumn:name="VPCID",type="string",JSONPath=".spec.vpcId"
// +kubebuilder:printcolumn:name="DESCRIPTION",type="string",JSONPath=".spec.description"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type SecurityGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityGroupSpec   `json:"spec"`
	Status SecurityGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityGroupList contains a list of SecurityGroups
type SecurityGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityGroup `json:"items"`
}

// UpdateExternalStatus updates the external status object, given the observation
func (s *SecurityGroup) UpdateExternalStatus(observation ec2.SecurityGroup) {
	s.Status.SecurityGroupExternalStatus = SecurityGroupExternalStatus{
		SecurityGroupID: aws.StringValue(observation.GroupId),
		Tags:            BuildFromEC2Tags(observation.Tags),
	}
}
