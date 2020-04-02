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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client is the external client used for DBSubnetGroup Custom Resource
type Client interface {
	CreateDBSubnetGroupRequest(input *rds.CreateDBSubnetGroupInput) rds.CreateDBSubnetGroupRequest
	DeleteDBSubnetGroupRequest(input *rds.DeleteDBSubnetGroupInput) rds.DeleteDBSubnetGroupRequest
	DescribeDBSubnetGroupsRequest(input *rds.DescribeDBSubnetGroupsInput) rds.DescribeDBSubnetGroupsRequest
	ModifyDBSubnetGroupRequest(input *rds.ModifyDBSubnetGroupInput) rds.ModifyDBSubnetGroupRequest
	AddTagsToResourceRequest(input *rds.AddTagsToResourceInput) rds.AddTagsToResourceRequest
	ListTagsForResourceRequest(input *rds.ListTagsForResourceInput) rds.ListTagsForResourceRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return rds.New(*cfg), nil
}

// IsDBSubnetGroupNotFoundErr returns true if the error is because the item doesn't exist
func IsDBSubnetGroupNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBSubnetGroupNotFoundFault)
}

// IsDBSubnetGroupUpToDate checks whether there is a change in any of the modifiable fields.
func IsDBSubnetGroupUpToDate(p v1beta1.DBSubnetGroupParameters, sg rds.DBSubnetGroup, tags []rds.Tag) bool { // nolint:gocyclo
	if p.Description != aws.StringValue(sg.DBSubnetGroupDescription) {
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
		if _, ok := pSubnets[aws.StringValue(id.SubnetIdentifier)]; !ok {
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
		val, ok := pTags[aws.StringValue(tag.Key)]
		if !ok || !strings.EqualFold(val, aws.StringValue(tag.Value)) {
			return false
		}
	}
	return true
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBSubnetGroup
func GenerateObservation(sg rds.DBSubnetGroup) v1beta1.DBSubnetGroupObservation {
	o := v1beta1.DBSubnetGroupObservation{
		State: aws.StringValue(sg.SubnetGroupStatus),
		ARN:   aws.StringValue(sg.DBSubnetGroupArn),
		VPCID: aws.StringValue(sg.VpcId),
	}

	if len(sg.Subnets) != 0 {
		o.Subnets = make([]v1beta1.Subnet, len(sg.Subnets))
		for i, val := range sg.Subnets {
			o.Subnets[i] = v1beta1.Subnet{
				SubnetID:     aws.StringValue(val.SubnetIdentifier),
				SubnetStatus: aws.StringValue(val.SubnetStatus),
			}
		}
	}
	return o
}

// LateInitialize fills the empty fields in *v1beta1.DBSubnetGroupParameters with
func LateInitialize(in *v1beta1.DBSubnetGroupParameters, sg *rds.DBSubnetGroup) {
	if sg == nil {
		return
	}

	in.Description = awsclients.LateInitializeString(in.Description, sg.DBSubnetGroupDescription)
	if len(in.SubnetIDs) == 0 && len(sg.Subnets) != 0 {
		in.SubnetIDs = make([]string, len(sg.Subnets))
		for i, val := range sg.Subnets {
			in.SubnetIDs[i] = aws.StringValue(val.SubnetIdentifier)
		}
	}
}

// IsErrorNotFound helper function to test for ErrCodeDBSubnetGroupNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBSubnetGroupNotFoundFault)
}
