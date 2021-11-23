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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	port80  int32 = 80
	port100 int32 = 100

	tcpProtocol = "tcp"
)

func sgPermissions(port int32, cidrs ...string) []ec2types.IpPermission {
	ranges := make([]ec2types.IpRange, 0, len(cidrs))
	for _, cidr := range cidrs {
		ranges = append(ranges, ec2types.IpRange{
			CidrIp: aws.String(cidr),
		})
	}
	return []ec2types.IpPermission{
		{
			FromPort:   aws.Int32(port),
			ToPort:     aws.Int32(port),
			IpProtocol: aws.String(tcpProtocol),
			IpRanges:   ranges,
		},
	}
}

func sgUserIDGroupPair(port int32, groupIDs ...string) []ec2types.IpPermission {
	groups := make([]ec2types.UserIdGroupPair, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		groups = append(groups, ec2types.UserIdGroupPair{
			GroupId: aws.String(groupID),
		})
	}
	return []ec2types.IpPermission{
		{
			FromPort:         aws.Int32(port),
			ToPort:           aws.Int32(port),
			IpProtocol:       aws.String(tcpProtocol),
			UserIdGroupPairs: groups,
		},
	}
}

// NOTE(muvaf): Sending -1 as FromPort or ToPort is valid but the returned
// object does not have that value. So, in case we have sent -1, we assume
// that the returned value is also -1 in case if it's nil.
// See the following about usage of -1
// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-security-group-egress.html

// mOne := int64(-1)

func TestDiffPermissions(t *testing.T) {
	type testCase struct {
		name string

		want, have  []ec2types.IpPermission
		add, remove []ec2types.IpPermission
	}

	cases := []testCase{
		{
			name:   "Same",
			want:   sgPermissions(port100, cidr),
			have:   sgPermissions(port100, cidr),
			add:    nil,
			remove: nil,
		},
		{
			name:   "Add",
			want:   sgPermissions(port100, cidr),
			have:   nil,
			add:    sgPermissions(port100, cidr),
			remove: nil,
		},
		{
			name:   "Remove",
			want:   nil,
			have:   sgPermissions(port100, cidr),
			add:    nil,
			remove: sgPermissions(port100, cidr),
		},
		{
			name:   "Replace",
			want:   sgPermissions(port80, cidr),
			have:   sgPermissions(port100, cidr),
			add:    sgPermissions(port80, cidr),
			remove: sgPermissions(port100, cidr),
		},
		{
			name:   "AddBlock",
			want:   sgPermissions(port100, cidr, "192.168.0.1/32"),
			have:   sgPermissions(port100, cidr),
			add:    sgPermissions(port100, "192.168.0.1/32"),
			remove: nil,
		},
		{
			name:   "RemoveBlock",
			want:   sgPermissions(port100, cidr),
			have:   sgPermissions(port100, cidr, "192.168.0.1/32"),
			add:    nil,
			remove: sgPermissions(port100, "192.168.0.1/32"),
		},
		{
			name:   "ReplaceBlock",
			want:   sgPermissions(port100, cidr, "172.240.1.1/32", "192.168.0.1/32"),
			have:   sgPermissions(port100, cidr, "172.240.2.2/32", "192.168.0.1/32"),
			add:    sgPermissions(port100, "172.240.1.1/32"),
			remove: sgPermissions(port100, "172.240.2.2/32"),
		},
		{
			name:   "DedupeWant",
			want:   append(sgPermissions(port100, cidr, "172.240.1.1/32", "172.240.1.1/32", "192.168.0.1/32"), sgPermissions(port100, cidr, "172.240.1.1/32", "172.240.1.1/32", "192.168.0.1/32")...),
			have:   sgPermissions(port100, cidr, "172.240.2.2/32", "192.168.0.1/32"),
			add:    sgPermissions(port100, "172.240.1.1/32"),
			remove: sgPermissions(port100, "172.240.2.2/32"),
		},
		{
			name:   "MergeWant",
			want:   append(sgPermissions(port100, "192.168.0.1/32"), sgPermissions(port100, "172.240.1.1/32")...),
			have:   nil,
			add:    sgPermissions(port100, "192.168.0.1/32", "172.240.1.1/32"),
			remove: nil,
		},
		{
			name:   "IgnoreOrder",
			want:   append(sgUserIDGroupPair(port100, "sg-2", "sg-1"), sgPermissions(port100, "172.240.1.1/32", "192.168.0.1/32", cidr)...),
			have:   append(sgUserIDGroupPair(port100, "sg-1", "sg-2"), sgPermissions(port100, "192.168.0.1/32", cidr, "172.240.1.1/32")...),
			add:    nil,
			remove: nil,
		},
		{
			name: "IgnoreProtocolCase",
			want: []ec2types.IpPermission{
				{
					IpProtocol: aws.String("TCP"),
					FromPort:   &port100,
					ToPort:     &port100,
					IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(cidr)}},
				},
			},
			have: []ec2types.IpPermission{
				{
					IpProtocol: aws.String("tcp"),
					FromPort:   &port100,
					ToPort:     &port100,
					IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(cidr)}},
				},
			},
			add:    nil,
			remove: nil,
		},
	}

	ipPermissionCmp := func(a, b ec2types.IpPermission) bool {
		return aws.ToString(a.IpProtocol) < aws.ToString(b.IpProtocol) && aws.ToInt32(a.FromPort) < aws.ToInt32(b.FromPort)
	}
	ipRangeCmp := func(a, b ec2types.IpRange) bool {
		return aws.ToString(a.CidrIp) < aws.ToString(b.CidrIp)
	}

	opts := cmp.Options{
		cmpopts.SortSlices(ipPermissionCmp),
		cmpopts.SortSlices(ipRangeCmp),
		cmpopts.IgnoreTypes(document.NoSerde{}),
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			add, remove := DiffPermissions(tc.want, tc.have)

			if diff := cmp.Diff(tc.add, add, opts); diff != "" {
				t.Errorf("r add: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.remove, remove, opts); diff != "" {
				t.Errorf("r remove: -want, +got:\n%s", diff)
			}
		})
	}
}
