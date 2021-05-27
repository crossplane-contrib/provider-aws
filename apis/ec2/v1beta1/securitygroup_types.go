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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SecurityGroupParameters define the desired state of an AWS VPC Security
// Group.
type SecurityGroupParameters struct {
	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your SecurityGroup to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// A description of the security group.
	// +immutable
	Description string `json:"description"`

	// The name of the security group.
	// +immutable
	GroupName string `json:"groupName"`

	// One or more inbound rules associated with the security group.
	// +optional
	Ingress []IPPermission `json:"ingress,omitempty"`

	// [EC2-VPC] One or more outbound rules associated with the security group.
	// +optional
	Egress []IPPermission `json:"egress,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// VPCID is the ID of the VPC.
	// +optional
	// +immutable
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to and retrieves its vpcId
	// +optional
	// +immutable
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to and retrieves its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
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

	// VPCIDRef reference a VPC to retrieve its vpcId
	// +optional
	// +immutable
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects reference to a VPC to retrieve its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

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

	// The IP protocol name (tcp, udp, icmp, icmpv6) or number (see Protocol Numbers
	// (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)).
	//
	// [VPC only] Use -1 to specify all protocols. When authorizing security group
	// rules, specifying -1 or a protocol number other than tcp, udp, icmp, or icmpv6
	// allows traffic on all ports, regardless of any port range you specify. For
	// tcp, udp, and icmp, you must specify a port range. For icmpv6, the port range
	// is optional; if you omit the port range, traffic for all types and codes
	// is allowed.
	IPProtocol string `json:"ipProtocol"`

	// The IPv4 ranges.
	// +optional
	IPRanges []IPRange `json:"ipRanges,omitempty"`

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
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecurityGroupParameters `json:"forProvider"`
}

// SecurityGroupObservation keeps the state for the external resource
type SecurityGroupObservation struct {
	// The AWS account ID of the owner of the security group.
	OwnerID string `json:"ownerId"`

	// SecurityGroupID is the ID of the SecurityGroup.
	SecurityGroupID string `json:"securityGroupID"`
}

// A SecurityGroupStatus represents the observed state of a SecurityGroup.
type SecurityGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecurityGroupObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A SecurityGroup is a managed resource that represents an AWS VPC Security
// Group.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.forProvider.vpcId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:storageversion
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
