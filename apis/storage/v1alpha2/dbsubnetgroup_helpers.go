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
	"github.com/aws/aws-sdk-go-v2/service/rds"

	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

// UpdateExternalStatus updates the external status object, given the observation
func (b *DBSubnetGroup) UpdateExternalStatus(observation rds.DBSubnetGroup) {

	subnets := make([]Subnet, len(observation.Subnets))
	for i, sn := range observation.Subnets {
		subnets[i] = Subnet{
			SubnetID:     aws.StringValue(sn.SubnetIdentifier),
			SubnetStatus: aws.StringValue(sn.SubnetStatus),
		}
	}

	b.Status.DBSubnetGroupExternalStatus = DBSubnetGroupExternalStatus{
		DBSubnetGroupARN:  aws.StringValue(observation.DBSubnetGroupArn),
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
