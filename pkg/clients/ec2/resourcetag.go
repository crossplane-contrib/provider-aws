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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// ResourceTagClient is the external client used for ResourceTag Custom Resource
type ResourceTagClient interface {
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	DescribeTagsRequest(*ec2.DescribeTagsInput) ec2.DescribeTagsRequest
	DeleteTagsRequest(*ec2.DeleteTagsInput) ec2.DeleteTagsRequest
}

// NewResourceTagClient returns a new client using AWS credentials as JSON encoded data.
func NewResourceTagClient(cfg aws.Config) ResourceTagClient {
	return ec2.New(cfg)
}
