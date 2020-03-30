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
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Client is the external client used for DBSubnetGroup Custom Resource
type Client interface { // nolint:gocyclo
	CreateDBSubnetGroupRequest(input *rds.CreateDBSubnetGroupInput) rds.CreateDBSubnetGroupRequest
	DeleteDBSubnetGroupRequest(input *rds.DeleteDBSubnetGroupInput) rds.DeleteDBSubnetGroupRequest
	DescribeDBSubnetGroupsRequest(input *rds.DescribeDBSubnetGroupsInput) rds.DescribeDBSubnetGroupsRequest
	ModifyDBSubnetGroupRequest(input *rds.ModifyDBSubnetGroupInput) rds.ModifyDBSubnetGroupRequest
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
func IsDBSubnetGroupUpToDate(p v1beta1.DBSubnetGroupParameters, sg rds.DBSubnetGroup) (bool, error) {
	patch, err := CreatePatch(&sg, &p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.DBSubnetGroupParameters{}, patch, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{})), nil
}

// CreatePatch creates a *v1beta1.DBSubnetGroupParameters that has only the changed
// values between the target *v1beta1.DBSubnetGroupParameters and the current *rds.DBSubnetGroup
func CreatePatch(in *rds.DBSubnetGroup, target *v1beta1.DBSubnetGroupParameters) (*v1beta1.DBSubnetGroupParameters, error) {
	currentParams := &v1beta1.DBSubnetGroupParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.DBSubnetGroupParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBSubnetGroup
func GenerateObservation(sg rds.DBSubnetGroup) v1beta1.DBSubnetGroupObservation {
	o := v1beta1.DBSubnetGroupObservation{
		SubnetGroupStatus: aws.StringValue(sg.SubnetGroupStatus),
		DBSubnetGroupArn:  aws.StringValue(sg.DBSubnetGroupArn),
		VPCID:             aws.StringValue(sg.VpcId),
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

	in.DBSubnetGroupDescription = awsclients.LateInitializeString(in.DBSubnetGroupDescription, sg.DBSubnetGroupDescription)
	in.DBSubnetGroupName = awsclients.LateInitializeString(in.DBSubnetGroupName, sg.DBSubnetGroupName)

	if len(in.SubnetIDs) == 0 && len(sg.Subnets) != 0 {
		in.SubnetIDs = make([]string, len(sg.Subnets))
		for i, val := range sg.Subnets {
			in.SubnetIDs[i] = aws.StringValue(val.SubnetIdentifier)
		}
	}
}

// IsErrorAlreadyExists returns true if the supplied error indicates a DB subnet
// group already exists.
func IsErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBSubnetGroupAlreadyExistsFault)
}

// IsErrorNotFound helper function to test for ErrCodeDBSubnetGroupNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBSubnetGroupNotFoundFault)
}
