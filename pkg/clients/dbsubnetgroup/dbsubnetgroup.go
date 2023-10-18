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

package dbsubnetgroup

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// Client is the external client used for DBSubnetGroup Custom Resource
type Client interface {
	CreateDBSubnetGroup(context.Context, *rds.CreateDBSubnetGroupInput, ...func(*rds.Options)) (*rds.CreateDBSubnetGroupOutput, error)
	DeleteDBSubnetGroup(context.Context, *rds.DeleteDBSubnetGroupInput, ...func(*rds.Options)) (*rds.DeleteDBSubnetGroupOutput, error)
	DescribeDBSubnetGroups(context.Context, *rds.DescribeDBSubnetGroupsInput, ...func(*rds.Options)) (*rds.DescribeDBSubnetGroupsOutput, error)
	ModifyDBSubnetGroup(context.Context, *rds.ModifyDBSubnetGroupInput, ...func(*rds.Options)) (*rds.ModifyDBSubnetGroupOutput, error)
	AddTagsToResource(context.Context, *rds.AddTagsToResourceInput, ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
	ListTagsForResource(context.Context, *rds.ListTagsForResourceInput, ...func(*rds.Options)) (*rds.ListTagsForResourceOutput, error)
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(cfg aws.Config) Client {
	return rds.NewFromConfig(cfg)
}

// IsDBSubnetGroupNotFoundErr returns true if the error is because the item doesn't exist
func IsDBSubnetGroupNotFoundErr(err error) bool {
	var nff *rdstypes.DBSubnetGroupNotFoundFault
	return errors.As(err, &nff)
}

// IsDBSubnetGroupUpToDate checks whether there is a change in any of the modifiable fields.
func IsDBSubnetGroupUpToDate(p v1beta1.DBSubnetGroupParameters, sg rdstypes.DBSubnetGroup, tags []rdstypes.Tag) bool { //nolint:gocyclo
	if p.Description != pointer.StringValue(sg.DBSubnetGroupDescription) {
		return false
	}

	if len(p.SubnetIDs) != len(sg.Subnets) {
		return false
	}

	pSubnets := make(map[string]struct{}, len(p.SubnetIDs))
	for _, id := range p.SubnetIDs {
		pSubnets[id] = struct{}{}
	}
	for _, id := range sg.Subnets {
		if _, ok := pSubnets[aws.ToString(id.SubnetIdentifier)]; !ok {
			return false
		}
	}

	if len(p.Tags) != len(tags) {
		return false
	}

	pTags := make(map[string]string, len(p.Tags))
	for _, tag := range p.Tags {
		pTags[tag.Key] = tag.Value
	}
	for _, tag := range tags {
		val, ok := pTags[aws.ToString(tag.Key)]
		if !ok || !strings.EqualFold(val, aws.ToString(tag.Value)) {
			return false
		}
	}
	return true
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBSubnetGroup
func GenerateObservation(sg rdstypes.DBSubnetGroup) v1beta1.DBSubnetGroupObservation {
	o := v1beta1.DBSubnetGroupObservation{
		State: aws.ToString(sg.SubnetGroupStatus),
		ARN:   aws.ToString(sg.DBSubnetGroupArn),
		VPCID: aws.ToString(sg.VpcId),
	}

	if len(sg.Subnets) != 0 {
		o.Subnets = make([]v1beta1.Subnet, len(sg.Subnets))
		for i, val := range sg.Subnets {
			o.Subnets[i] = v1beta1.Subnet{
				SubnetID:     aws.ToString(val.SubnetIdentifier),
				SubnetStatus: aws.ToString(val.SubnetStatus),
			}
		}
	}
	return o
}

// LateInitialize fills the empty fields in *v1beta1.DBSubnetGroupParameters with
func LateInitialize(in *v1beta1.DBSubnetGroupParameters, sg *rdstypes.DBSubnetGroup) {
	if sg == nil {
		return
	}

	in.Description = pointer.LateInitializeValueFromPtr(in.Description, sg.DBSubnetGroupDescription)
	if len(in.SubnetIDs) == 0 && len(sg.Subnets) != 0 {
		in.SubnetIDs = make([]string, len(sg.Subnets))
		for i, val := range sg.Subnets {
			in.SubnetIDs[i] = aws.ToString(val.SubnetIdentifier)
		}
	}
}
