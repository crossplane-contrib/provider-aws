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

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// GroupClient is the external client used for Group Custom Resource
type GroupClient interface {
	GetGroup(ctx context.Context, input *iam.GetGroupInput, opts ...func(*iam.Options)) (*iam.GetGroupOutput, error)
	CreateGroup(ctx context.Context, input *iam.CreateGroupInput, opts ...func(*iam.Options)) (*iam.CreateGroupOutput, error)
	DeleteGroup(ctx context.Context, input *iam.DeleteGroupInput, opts ...func(*iam.Options)) (*iam.DeleteGroupOutput, error)
	UpdateGroup(ctx context.Context, input *iam.UpdateGroupInput, opts ...func(*iam.Options)) (*iam.UpdateGroupOutput, error)
}

// NewGroupClient returns a new client using AWS credentials as JSON encoded data.
func NewGroupClient(cfg aws.Config) GroupClient {
	return iam.NewFromConfig(cfg)
}
