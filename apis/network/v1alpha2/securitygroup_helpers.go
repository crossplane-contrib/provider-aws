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

package v1alpha2

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

// BuildEC2Permissions converts object Permissions to ec2 format
func BuildEC2Permissions(objectPerms []IPPermission) []ec2.IpPermission {
	permissions := make([]ec2.IpPermission, len(objectPerms))
	for i, p := range objectPerms {

		ipPerm := ec2.IpPermission{
			FromPort:   aws.Int64(int(p.FromPort)),
			ToPort:     aws.Int64(int(p.ToPort)),
			IpProtocol: aws.String(p.IPProtocol),
		}

		ipPerm.IpRanges = make([]ec2.IpRange, len(p.CIDRBlocks))
		for j, c := range p.CIDRBlocks {
			ipPerm.IpRanges[j] = ec2.IpRange{
				CidrIp:      aws.String(c.CIDRIP),
				Description: aws.String(c.Description),
			}
		}

		permissions[i] = ipPerm
	}

	return permissions
}
