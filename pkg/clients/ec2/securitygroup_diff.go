/*
Copyright 2021 The Crossplane Authors.

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

package ec2

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// ruleKey represents the unique tuple (protocol, from-to port) in a
// format supported as a map key
type ruleKey struct {
	protocol string // lower case
	fromPort int32  // -1 for nil
	toPort   int32  // -1 for nil
}

func getInt32Key(port *int32) int32 {
	if port == nil {
		return -1
	}
	return *port
}

func getKey(perm ec2types.IpPermission) ruleKey {
	return ruleKey{
		protocol: strings.ToLower(aws.ToString(perm.IpProtocol)),
		fromPort: getInt32Key(perm.FromPort),
		toPort:   getInt32Key(perm.ToPort),
	}
}

type ipPermissionMap struct { //nolint:recvcheck
	FromPort   *int32
	ToPort     *int32
	IPProtocol *string

	ipRanges      map[string]*string
	ipv6Ranges    map[string]*string
	prefixListIDs map[string]*string
	groups        map[string]ec2types.UserIdGroupPair
}

// merge adds rules from the permission set m into this permission
// set. The caller must ensure that the permission set is for the same
// protocol and port range.
func (i *ipPermissionMap) merge(m ec2types.IpPermission) {
	i.FromPort = m.FromPort
	i.ToPort = m.ToPort
	i.IPProtocol = m.IpProtocol

	for _, r := range m.IpRanges {
		i.ipRanges[aws.ToString(r.CidrIp)] = r.Description
	}

	for _, r := range m.Ipv6Ranges {
		i.ipv6Ranges[aws.ToString(r.CidrIpv6)] = r.Description
	}

	for _, r := range m.PrefixListIds {
		i.prefixListIDs[aws.ToString(r.PrefixListId)] = r.Description
	}

	for _, r := range m.UserIdGroupPairs {
		// a UserIdGroupPair must have a group id or group name, and
		// they are unique so we can use them for keys
		key := aws.ToString(r.GroupId)
		if key == "" {
			key = aws.ToString(r.GroupName)
		}
		i.groups[key] = r
	}
}

// diff returns rules that should be added or removed.
func (i ipPermissionMap) diff(other ipPermissionMap) (add ec2types.IpPermission, remove ec2types.IpPermission) {
	add.IpProtocol = i.IPProtocol
	add.FromPort = i.FromPort
	add.ToPort = i.ToPort
	remove = add

	add.IpRanges = i.diffRanges(other)
	remove.IpRanges = other.diffRanges(i)

	add.Ipv6Ranges = i.diffIPv6Ranges(other)
	remove.Ipv6Ranges = other.diffIPv6Ranges(i)

	add.PrefixListIds = i.diffPrefixListIDs(other)
	remove.PrefixListIds = other.diffPrefixListIDs(i)

	add.UserIdGroupPairs = i.diffUserIDGroupPair(other)
	remove.UserIdGroupPairs = other.diffUserIDGroupPair(i)

	return add, remove
}

func (i ipPermissionMap) diffRanges(other ipPermissionMap) []ec2types.IpRange {
	var ret []ec2types.IpRange
	for cidr, description := range i.ipRanges {
		cidr := cidr
		description2, ok := other.ipRanges[cidr]
		if !ok || aws.ToString(description) != aws.ToString(description2) {
			ret = append(ret, ec2types.IpRange{CidrIp: &cidr, Description: description})
		}
	}
	return ret
}

func (i ipPermissionMap) diffIPv6Ranges(other ipPermissionMap) []ec2types.Ipv6Range {
	var ret []ec2types.Ipv6Range
	for cidr, description := range i.ipv6Ranges {
		cidr := cidr
		description2, ok := other.ipv6Ranges[cidr]
		if !ok || aws.ToString(description) != aws.ToString(description2) {
			ret = append(ret, ec2types.Ipv6Range{CidrIpv6: &cidr, Description: description})
		}
	}
	return ret
}

func (i ipPermissionMap) diffPrefixListIDs(other ipPermissionMap) []ec2types.PrefixListId {
	var ret []ec2types.PrefixListId
	for id, description := range i.prefixListIDs {
		id := id
		description2, ok := other.prefixListIDs[id]
		if !ok || aws.ToString(description) != aws.ToString(description2) {
			ret = append(ret, ec2types.PrefixListId{PrefixListId: &id, Description: description})
		}
	}
	return ret
}

func (i *ipPermissionMap) diffUserIDGroupPair(other ipPermissionMap) []ec2types.UserIdGroupPair {
	var ret []ec2types.UserIdGroupPair
	for key, r := range i.groups {
		r2, ok := other.groups[key]
		if !ok || aws.ToString(r.Description) != aws.ToString(r2.Description) {
			// there are other fields here:
			// * peering connection id
			// * peering status
			// * VPC ID
			// * User ID
			// but all are immutable per `key` (group name/id), or one-way status fields
			ret = append(ret, r)
		}
	}
	return ret
}

func convertToMaps(rules []ec2types.IpPermission) map[ruleKey]*ipPermissionMap {
	ret := make(map[ruleKey]*ipPermissionMap)

	for _, rule := range rules {
		k := getKey(rule)
		normalized, ok := ret[k]
		if !ok {
			normalized = &ipPermissionMap{}
			normalized.ipRanges = make(map[string]*string)
			normalized.ipv6Ranges = make(map[string]*string)
			normalized.prefixListIDs = make(map[string]*string)
			normalized.groups = make(map[string]ec2types.UserIdGroupPair)
			ret[k] = normalized
		}

		normalized.merge(rule)
	}

	return ret
}

func hasRules(perm ec2types.IpPermission) bool {
	return perm.IpRanges != nil || perm.Ipv6Ranges != nil || perm.UserIdGroupPairs != nil || perm.PrefixListIds != nil
}

// DiffPermissions compares two permission sets, and returns the rules
// to add and remove to make them identical
func DiffPermissions(want, have []ec2types.IpPermission) (add, remove []ec2types.IpPermission) {
	// Convert the rule matrix to a map of arrays.

	// We do this to avoid O(n^2) lookup if the rule sets are large,
	// and also because the user might represent two rules
	//
	//   [(proto,port,[iprange1,iprange2])]
	// as
	//   [(proto,port,[iprange1]), (proto,port,[iprange2])]
	//
	// By converting to maps and merging rules we can get the compact
	// first form and easily check if rules are present or not.
	wantMap := convertToMaps(want)
	haveMap := convertToMaps(have)

	for key, have := range haveMap {
		want, ok := wantMap[key]
		if !ok {
			want = &ipPermissionMap{}
		}

		removeRules, addRules := have.diff(*want)

		if hasRules(addRules) {
			add = append(add, addRules)
		}

		if hasRules(removeRules) {
			remove = append(remove, removeRules)
		}
	}

	for key, want := range wantMap {
		if _, ok := haveMap[key]; !ok {
			addRules, _ := want.diff(ipPermissionMap{})
			add = append(add, addRules)
		}
	}

	return add, remove
}
