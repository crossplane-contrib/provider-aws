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

package rds

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// DBSubnetGroupClient is the external client used for DBSubnetGroup Custom Resource
type DBSubnetGroupClient interface {
	CreateDBSubnetGroupRequest(input *rds.CreateDBSubnetGroupInput) rds.CreateDBSubnetGroupRequest
	DeleteDBSubnetGroupRequest(input *rds.DeleteDBSubnetGroupInput) rds.DeleteDBSubnetGroupRequest
	DescribeDBSubnetGroupsRequest(input *rds.DescribeDBSubnetGroupsInput) rds.DescribeDBSubnetGroupsRequest
}

// NewDBSubnetGroupClient returns a new client using AWS credentials as JSON encoded data.
func NewDBSubnetGroupClient(cfg *aws.Config) (DBSubnetGroupClient, error) {
	return rds.New(*cfg), nil
}

// IsDBSubnetGroupNotFoundErr returns true if the error is because the item doesn't exist
func IsDBSubnetGroupNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == rds.ErrCodeDBSubnetGroupNotFoundFault {
			return true
		}
	}

	return false
}
