// /*
// Copyright 2021 The Crossplane Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package server

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
)

type vpcEndpointClient interface {
	DescribeVpcEndpoints(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error)
}

// newVPCClient generates an ec2 client for describing the vpc endpoint
func newVPCClient(sess *session.Session) vpcEndpointClient {
	return ec2.New(sess)
}

type customConnector struct {
	*connector
	newClientFn func(config *session.Session) vpcEndpointClient
}

func (c *customConnector) Connect(ctx context.Context, cr *svcapitypes.Server) (managed.TypedExternalClient[*svcapitypes.Server], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}

	external := newExternal(c.kube, svcsdk.New(sess), c.opts)
	vpcClient := c.newClientFn(sess)
	if vpcClient == nil {
		return nil, errors.New("failed to initialize VPC client")
	}
	custom := &custom{
		client:            external.client,
		kube:              external.kube,
		external:          external,
		vpcEndpointClient: vpcClient,
	}

	external.postObserve = custom.postObserve
	external.postCreate = postCreate
	external.preObserve = preObserve
	external.preDelete = preDelete
	external.preCreate = preCreate
	external.isUpToDate = isUpToDate
	external.lateInitialize = lateInitialize
	external.update = external.UpdateServer
	return external, nil
}
