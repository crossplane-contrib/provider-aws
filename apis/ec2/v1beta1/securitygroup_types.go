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

// SecurityGroupParameters define the desired state of an AWS VPC Security
// Group.
type SecurityGroupParameters struct {

	// Region is the region you'd like your SecurityGroup to be created in.
	Region string `json:"region"`

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
	// +crossplane:generate:reference:type=VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to and retrieves its vpcId
	// +optional
	// +immutable
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to and retrieves its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

	// Dont manage the ingress settings for the created resource
	IgnoreIngress *bool `json:"ignoreIngress,omitempty"`

	// Dont manage the egress settings for the created resource
	IgnoreEgress *bool `json:"ignoreEgress,omitempty"`
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
	// +crossplane:generate:reference:type=SecurityGroup
	GroupID *string `json:"groupId,omitempty"`

	// GroupIDRef reference a security group to retrieve its GroupID
	// +optional
	// +immutable
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects reference to a security group to retrieve its GroupID
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`

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
	// +crossplane:generate:reference:type=VPC
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

// ClearRefSelectors nils out ref and selectors
func (u *UserIDGroupPair) ClearRefSelectors() {
	u.VPCIDRef = nil
	u.VPCIDSelector = nil
	u.GroupIDRef = nil
	u.GroupIDSelector = nil
}

// IPPermission Describes a set of permissions for a security group rule.
type IPPermission struct {
	// The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6
	// type number. A value of -1 indicates all ICMP/ICMPv6 types. If you specify
	// all ICMP/ICMPv6 types, you must specify all codes.
	// +optional
	FromPort *int32 `json:"fromPort,omitempty"`

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
	ToPort *int32 `json:"toPort,omitempty"`

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

	// IngressRules of the observed SecurityGroup.
	IngressRules []SecurityGroupRuleObservation `json:"ingressRules,omitempty"`

	// EgressRules of the observed SecurityGroup.
	EgressRules []SecurityGroupRuleObservation `json:"egressRules,omitempty"`
}

type SecurityGroupRuleObservation struct {
	// ID of the security group rule.
	ID *string `json:"id,omitempty"`

	// CidrIpv4 range.
	CidrIpv4 *string `json:"cidrIpv4,omitempty"`

	// CidrIpv6 range.
	CidrIpv6 *string `json:"cidrIpv6,omitempty"`

	// The IP protocol name (tcp, udp, icmp, icmpv6) or number (see Protocol Numbers
	// (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)). Use
	// -1 to specify all protocols.
	IpProtocol *string `json:"ipProtocol,omitempty"`

	// Description of this rule.
	Description *string `json:"description,omitempty"`

	// The start of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 type. A
	// value of -1 indicates all ICMP/ICMPv6 types. If you specify all ICMP/ICMPv6
	// types, you must specify all codes.
	FromPort *int32 `json:"fromPort,omitempty"`

	// The ID of the prefix list.
	PrefixListId *string `json:"prefixListId,omitempty"`

	// Describes the security group that is referenced in the rule.
	ReferencedGroupInfo *ReferencedSecurityGroup `json:"referencedGroupInfo,omitempty"`

	// The end of port range for the TCP and UDP protocols, or an ICMP/ICMPv6 code. A
	// value of -1 indicates all ICMP/ICMPv6 codes. If you specify all ICMP/ICMPv6
	// types, you must specify all codes.
	ToPort *int32 `json:"toPort,omitempty"`
}

// A ReferencedSecurityGroup describes the security group that is referenced in the security group rule.
type ReferencedSecurityGroup struct {

	// The ID of the security group.
	GroupId *string `json:"groupId,omitempty"`

	// The status of a VPC peering connection, if applicable.
	PeeringStatus *string `json:"peeringStatus,omitempty"`

	// The Amazon Web Services account ID.
	UserId *string `json:"userId,omitempty"`

	// The ID of the VPC.
	VpcId *string `json:"vpcId,omitempty"`

	// The ID of the VPC peering connection.
	VpcPeeringConnectionId *string `json:"vpcPeeringConnectionId,omitempty"`
}

// A SecurityGroupStatus represents the observed state of a SecurityGroup.
type SecurityGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecurityGroupObservation `json:"atProvider,omitempty"`
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
