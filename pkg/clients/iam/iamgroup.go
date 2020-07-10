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

	"github.com/aws/aws-sdk-go-v2/service/iam"

	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// GroupClient is the external client used for IAMGroup Custom Resource
type GroupClient interface {
	CreateGroupRequest(*iam.CreateGroupInput) iam.CreateGroupRequest
	GetGroupRequest(*iam.GetGroupInput) iam.GetGroupRequest
	UpdateGroupRequest(*iam.UpdateGroupInput) iam.UpdateGroupRequest
	DeleteGroupRequest(*iam.DeleteGroupInput) iam.DeleteGroupRequest
}

// NewGroupClient returns a new client using AWS credentials as JSON encoded data.
func NewGroupClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (GroupClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), nil
}
