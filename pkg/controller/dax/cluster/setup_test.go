/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS_IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dax"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dax/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/dax/fake"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	testClusterName          = "test-cluster"
	testSubnetGroupName      = "test-subnet-group"
	testOtherSubnetGroupName = "test-other-subnet-group"
	testDescription          = "test-description"
	testOtherDescription     = "some-other-description"

	testReplicationFactor         = 2
	testActiveNodes               = 3
	testClusterARN                = "test-cluster-ARN"
	testIamRoleARN                = "test-iam-role-ARN"
	testOtherIamRoleARN           = "test-other-iam-role-ARN"
	testNodeIDToRemove            = "test-node-id-to-remove"
	testOtherNodeIDToRemove       = "test-other-node-id-to-remove"
	testNodeID                    = "test-node-id"
	testOtherNodeID               = "test-other-node-id"
	testNodeStatus                = "test-node-status"
	testOtherNodeStatus           = "test-other-node-status"
	testParameterGroupStatus      = "test-parameter-group-status"
	testOtherParameterGroupStatus = "test-other-parameter-group-status"

	testEndpointAddress              = "test-endpoint-address"
	testEndpointPort                 = 1
	testEndpointURL                  = "test-endpoint-url"
	testAvailabilityZone             = "us-east-1a"
	testOtherAvailabilityZone        = "us-east-1b"
	testStatus                       = "test-status"
	testTopicARN                     = "test-topic-ARN"
	testOtherTopicARN                = "test-other-topic-ARN"
	testTopicStatus                  = "test-topic-status"
	testParameterGroupName           = "test-parameter-group-name"
	testOtherParameterGroupName      = "test-other-parameter-group-name"
	testParameterApplyStatus         = "test-parameter-apply-status"
	testNodeIDToReboot               = "test-node-id"
	testSubnetGroup                  = "test-subnet-group"
	testSecurityGroupIdentifier      = "test-security-group-identifier"
	testOtherSecurityGroupIdentifier = "test-other-security-group-identifier"
	testSecurityGroupStatus          = "test-security-group-status"

	testSSEDescriptionStatus = "test-sse-description-status"

	testClusterEndpointEncryptionType      = "test-cluster-endpoint-encryption-type"
	testOtherClusterEndpointEncryptionType = "test-other-cluster-endpoint-encryption-type"
	testNodeType                           = "test-node-type"
	testOtherNodeType                      = "test-other-node-type"
	testPreferredMaintenanceWindow         = "test-preferred-maintenance-window"
	testOtherPreferredMaintenanceWindow    = "test-other-preferred-maintenance-window"

	testSSESpecificationEnabled = true
	testTagKey                  = "test-tag-key"
	testTagValue                = "test-tag-value"

	testErrCreateClusterFailed    = "CreateCluster failed"
	testErrDeleteClusterFailed    = "DeleteSubnetGroup failed"
	testErrUpdateClusterFailed    = "UpdateCluster failed"
	testErrDescribeClustersFailed = "DescribeClusters failed"
)

type args struct {
	dax  *fake.MockDaxClient
	kube client.Client
	cr   *svcapitypes.Cluster
}

type daxModifier func(group *svcapitypes.Cluster)

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	e.preCreate = preCreate
	e.preUpdate = preUpdate
	e.preDelete = preDelete
	e.isUpToDate = isUpToDate
}

func instance(m ...daxModifier) *svcapitypes.Cluster {
	cr := &svcapitypes.Cluster{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		meta.SetExternalName(o, value)

	}
}

func withName(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Name = value
	}
}

func withStatusName(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Status.AtProvider.ClusterName = pointer.String(value)
	}
}

func withSpec(value svcapitypes.ClusterParameters) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider = value
	}
}

func withDescription(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.Description = pointer.String(value)
	}
}

func withParameterGroupName(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.ParameterGroupName = pointer.String(value)
	}
}

func withSubnetGroupName(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.SubnetGroupName = pointer.String(value)
	}
}

func withIAMRoleARN(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.IAMRoleARN = pointer.String(value)
	}
}

func withSecurityGroupIDs(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.SecurityGroupIDs = append(o.Spec.ForProvider.SecurityGroupIDs, pointer.String(value))
	}
}

func withAvailabilityZones(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.AvailabilityZones = append(o.Spec.ForProvider.AvailabilityZones, pointer.String(value))
	}
}

func withPreferredMaintenanceWindow(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.PreferredMaintenanceWindow = pointer.String(value)
	}
}

func withClusterEndpointEncryptionType(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.ClusterEndpointEncryptionType = pointer.String(value)
	}
}

func withNodeType(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.NodeType = pointer.String(value)
	}
}

func withNotificationTopicARN(value string) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Spec.ForProvider.NotificationTopicARN = pointer.String(value)
	}
}

func withStatus(value svcapitypes.ClusterObservation) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Status.AtProvider = value
	}
}

func withConditions(value ...xpv1.Condition) daxModifier {
	return func(o *svcapitypes.Cluster) {
		o.Status.SetConditions(value...)
	}
}

func baseCluster() *dax.Cluster {
	return &dax.Cluster{
		ActiveNodes: pointer.Int64(testActiveNodes),
		ClusterArn:  pointer.String(testClusterARN),
		ClusterDiscoveryEndpoint: &dax.Endpoint{
			Address: pointer.String(testEndpointAddress),
			Port:    pointer.Int64(testEndpointPort),
			URL:     pointer.String(testEndpointURL),
		},
		ClusterEndpointEncryptionType: pointer.String(testClusterEndpointEncryptionType),
		ClusterName:                   pointer.String(testClusterName),
		Description:                   pointer.String(testDescription),
		IamRoleArn:                    pointer.String(testIamRoleARN),
		NodeIdsToRemove:               []*string{pointer.String(testNodeIDToRemove), pointer.String(testOtherNodeIDToRemove)},
		NodeType:                      pointer.String(testNodeType),
		Nodes: []*dax.Node{
			{
				AvailabilityZone:     pointer.String(testAvailabilityZone),
				NodeId:               pointer.String(testNodeID),
				NodeStatus:           pointer.String(testNodeStatus),
				ParameterGroupStatus: pointer.String(testParameterGroupStatus),
			},
			{
				AvailabilityZone:     pointer.String(testOtherAvailabilityZone),
				NodeId:               pointer.String(testOtherNodeID),
				NodeStatus:           pointer.String(testOtherNodeStatus),
				ParameterGroupStatus: pointer.String(testOtherParameterGroupStatus),
			},
		},
		NotificationConfiguration: &dax.NotificationConfiguration{
			TopicArn:    pointer.String(testTopicARN),
			TopicStatus: pointer.String(testTopicStatus),
		},
		ParameterGroup: &dax.ParameterGroupStatus{
			NodeIdsToReboot:      []*string{pointer.String(testNodeIDToReboot)},
			ParameterApplyStatus: pointer.String(testParameterApplyStatus),
			ParameterGroupName:   pointer.String(testParameterGroupName),
		},
		PreferredMaintenanceWindow: pointer.String(testPreferredMaintenanceWindow),
		SSEDescription:             &dax.SSEDescription{Status: pointer.String(testSSEDescriptionStatus)},
		SecurityGroups: []*dax.SecurityGroupMembership{
			{
				SecurityGroupIdentifier: pointer.String(testSecurityGroupIdentifier),
				Status:                  pointer.String(testSecurityGroupStatus),
			},
		},
		Status:      pointer.String(testStatus),
		SubnetGroup: pointer.String(testSubnetGroup),
		TotalNodes:  pointer.Int64(2),
	}
}

func baseClusterParameters() svcapitypes.ClusterParameters {
	return svcapitypes.ClusterParameters{
		Region: "",
		AvailabilityZones: []*string{
			pointer.String(testAvailabilityZone),
			pointer.String(testOtherAvailabilityZone),
		},
		ClusterEndpointEncryptionType: pointer.String(testClusterEndpointEncryptionType),
		Description:                   pointer.String(testDescription),
		NodeType:                      pointer.String(testNodeType),
		PreferredMaintenanceWindow:    pointer.String(testPreferredMaintenanceWindow),
		ReplicationFactor:             pointer.Int64(testReplicationFactor),
		SSESpecification:              &svcapitypes.SSESpecification{Enabled: pointer.Bool(testSSESpecificationEnabled)},
		Tags:                          []*svcapitypes.Tag{{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)}},
		CustomClusterParameters: svcapitypes.CustomClusterParameters{
			NotificationTopicARN: pointer.String(testTopicARN),
			IAMRoleARN:           pointer.String(testIamRoleARN),
			ParameterGroupName:   pointer.String(testParameterGroupName),
			SubnetGroupName:      pointer.String(testSubnetGroupName),
			SecurityGroupIDs:     []*string{pointer.String(testSecurityGroupIdentifier)},
		},
	}
}

func baseClusterObservation() svcapitypes.ClusterObservation {
	return svcapitypes.ClusterObservation{
		ActiveNodes: pointer.Int64(testActiveNodes),
		ClusterARN:  pointer.String(testClusterARN),
		ClusterDiscoveryEndpoint: &svcapitypes.Endpoint{
			Address: pointer.String(testEndpointAddress),
			Port:    pointer.Int64(testEndpointPort),
			URL:     pointer.String(testEndpointURL),
		},
		ClusterName:     pointer.String(testClusterName),
		IAMRoleARN:      pointer.String(testIamRoleARN),
		NodeIDsToRemove: []*string{pointer.String(testNodeIDToRemove), pointer.String(testOtherNodeIDToRemove)},
		Nodes: []*svcapitypes.Node{
			{
				AvailabilityZone:     pointer.String(testAvailabilityZone),
				NodeID:               pointer.String(testNodeID),
				NodeStatus:           pointer.String(testNodeStatus),
				ParameterGroupStatus: pointer.String(testParameterGroupStatus),
			},
			{
				AvailabilityZone:     pointer.String(testOtherAvailabilityZone),
				NodeID:               pointer.String(testOtherNodeID),
				NodeStatus:           pointer.String(testOtherNodeStatus),
				ParameterGroupStatus: pointer.String(testOtherParameterGroupStatus),
			},
		},
		NotificationConfiguration: &svcapitypes.NotificationConfiguration{
			TopicARN:    pointer.String(testTopicARN),
			TopicStatus: pointer.String(testTopicStatus),
		},
		ParameterGroup: &svcapitypes.ParameterGroupStatus_SDK{
			NodeIDsToReboot:      []*string{pointer.String(testNodeIDToReboot)},
			ParameterApplyStatus: pointer.String(testParameterApplyStatus),
			ParameterGroupName:   pointer.String(testParameterGroupName),
		},
		SSEDescription: &svcapitypes.SSEDescription{Status: pointer.String(testSSEDescriptionStatus)},
		SecurityGroups: []*svcapitypes.SecurityGroupMembership{
			{
				SecurityGroupIdentifier: pointer.String(testSecurityGroupIdentifier),
				Status:                  pointer.String(testSecurityGroupStatus),
			},
		},
		Status:      pointer.String(testStatus),
		SubnetGroup: pointer.String(testSubnetGroup),
		TotalNodes:  pointer.Int64(2),
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.Cluster
		result managed.ExternalObservation
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableStateAndUpToDate": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedDescription": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withDescription(testOtherDescription),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withDescription(testOtherDescription),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedIAMRoleARN": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withIAMRoleARN(testOtherIamRoleARN),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withIAMRoleARN(testOtherIamRoleARN),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedNodeType": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withNodeType(testOtherNodeType),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withNodeType(testOtherNodeType),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedClusterEndpointEncryptionType": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withClusterEndpointEncryptionType(testOtherClusterEndpointEncryptionType),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withClusterEndpointEncryptionType(testOtherClusterEndpointEncryptionType),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedSubnetGroupName": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withSubnetGroupName(testOtherSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withSubnetGroupName(testOtherSubnetGroupName),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedPreferredMaintenanceWindow": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedParameterGroupName": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withParameterGroupName(testOtherParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withParameterGroupName(testOtherParameterGroupName),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedNotificationTopicARN": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withNotificationTopicARN(testOtherTopicARN),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withNotificationTopicARN(testOtherTopicARN),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedSecurityIds": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withSecurityGroupIDs(testOtherSecurityGroupIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withSecurityGroupIDs(testOtherSecurityGroupIdentifier),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedAvailabilityZones": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{baseCluster()}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withAvailabilityZones(testOtherAvailabilityZone),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
					withAvailabilityZones(testOtherAvailabilityZone),
					withStatus(baseClusterObservation()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"Empty_DescribeClustersOutput": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{Clusters: []*dax.Cluster{}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(instance(
								withName(testClusterName),
								withStatus(baseClusterObservation()),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"ErrDescribeClustersWithContext": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeClustersWithContext: func(c context.Context, dci *dax.DescribeClustersInput, o []request.Option) (*dax.DescribeClustersOutput, error) {
						return &dax.DescribeClustersOutput{}, errors.New(testErrDescribeClustersFailed)
					},
				},
				cr: instance(
					withExternalName(testClusterName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
				),
				err: errors.Wrap(errors.New(testErrDescribeClustersFailed), errDescribe),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeClustersWithContext: []*fake.CallDescribeClustersWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeClustersInput(
								instance(
									withExternalName(testClusterName),
									withName(testClusterName),
									withStatusName(testClusterName))),
							Opts: nil,
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.dax, opts)
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.Cluster
		result managed.ExternalUpdate
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdateOneParameter": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateClusterWithContext: func(c context.Context, uci *dax.UpdateClusterInput, o []request.Option) (*dax.UpdateClusterOutput, error) {
						return &dax.UpdateClusterOutput{Cluster: &dax.Cluster{
							ClusterName: pointer.String(testClusterName),
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withDescription(testDescription),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withDescription(testDescription),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateClusterWithContext: []*fake.CallUpdateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateClusterInput{
								ClusterName: pointer.String(testClusterName),
								Description: pointer.String(testDescription),
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdateSomeParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateClusterWithContext: func(c context.Context, uci *dax.UpdateClusterInput, o []request.Option) (*dax.UpdateClusterOutput, error) {
						return &dax.UpdateClusterOutput{Cluster: &dax.Cluster{
							ClusterName: pointer.String(testClusterName),
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSpec(baseClusterParameters()),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateClusterWithContext: []*fake.CallUpdateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateClusterInput{
								ClusterName:                pointer.String(testClusterName),
								Description:                pointer.String(testDescription),
								NotificationTopicArn:       pointer.String(testTopicARN),
								ParameterGroupName:         pointer.String(testParameterGroupName),
								PreferredMaintenanceWindow: pointer.String(testPreferredMaintenanceWindow),
								SecurityGroupIds:           []*string{pointer.String(testSecurityGroupIdentifier)},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdateSecurityGroupId": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateClusterWithContext: func(c context.Context, uci *dax.UpdateClusterInput, o []request.Option) (*dax.UpdateClusterOutput, error) {
						return &dax.UpdateClusterOutput{Cluster: &dax.Cluster{
							ClusterName: pointer.String(testClusterName),
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
					withSecurityGroupIDs(testOtherSecurityGroupIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withSecurityGroupIDs(testOtherSecurityGroupIdentifier),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateClusterWithContext: []*fake.CallUpdateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateClusterInput{
								ClusterName:      pointer.String(testClusterName),
								SecurityGroupIds: []*string{pointer.String(testOtherSecurityGroupIdentifier)},
							},
						},
					},
				},
			},
		},
		"testErrUpdateClusterFailed": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateClusterWithContext: func(c context.Context, uci *dax.UpdateClusterInput, o []request.Option) (*dax.UpdateClusterOutput, error) {
						return nil, errors.New(testErrUpdateClusterFailed)
					},
				},
				cr: instance(
					withExternalName(testClusterName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
				),
				err:    errors.Wrap(errors.New(testErrUpdateClusterFailed), errUpdate),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateClusterWithContext: []*fake.CallUpdateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateClusterInput{
								ClusterName: pointer.String(testClusterName),
							},
						},
					},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.dax, opts)
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.Cluster
		result managed.ExternalCreation
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreateWithParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateClusterWithContext: func(c context.Context, cci *dax.CreateClusterInput, o []request.Option) (*dax.CreateClusterOutput, error) {
						return &dax.CreateClusterOutput{Cluster: baseCluster()}, nil
					},
				},
				cr: instance(
					withName(testClusterName),
					withSpec(baseClusterParameters()),
				),
			},
			want: want{
				cr: instance(
					withName(testClusterName),
					withSpec(baseClusterParameters()),
					withConditions(xpv1.Creating()),
					withStatus(baseClusterObservation()),
					withExternalName(testClusterName),
				),
				result: managed.ExternalCreation{},
				dax: fake.MockDaxClientCall{
					CreateClusterWithContext: []*fake.CallCreateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateClusterInput{
								AvailabilityZones:             []*string{pointer.String(testAvailabilityZone), pointer.String(testOtherAvailabilityZone)},
								ClusterEndpointEncryptionType: pointer.String(testClusterEndpointEncryptionType),
								ClusterName:                   pointer.String(testClusterName),
								Description:                   pointer.String(testDescription),
								IamRoleArn:                    pointer.String(testIamRoleARN),
								NodeType:                      pointer.String(testNodeType),
								NotificationTopicArn:          pointer.String(testTopicARN),
								ParameterGroupName:            pointer.String(testParameterGroupName),
								PreferredMaintenanceWindow:    pointer.String(testPreferredMaintenanceWindow),
								ReplicationFactor:             pointer.Int64(testReplicationFactor),
								SSESpecification:              &dax.SSESpecification{Enabled: pointer.Bool(testSSESpecificationEnabled)},
								SecurityGroupIds:              []*string{pointer.String(testSecurityGroupIdentifier)},
								SubnetGroupName:               pointer.String(testSubnetGroupName),
								Tags: []*dax.Tag{{
									Key:   pointer.String(testTagKey),
									Value: pointer.String(testTagValue),
								}},
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateClusterWithContext: func(c context.Context, cci *dax.CreateClusterInput, o []request.Option) (*dax.CreateClusterOutput, error) {
						return &dax.CreateClusterOutput{}, errors.New(testErrCreateClusterFailed)
					},
				},
				cr: instance(
					withName(testClusterName),
					withSpec(baseClusterParameters()),
				),
			},
			want: want{
				cr: instance(
					withName(testClusterName),
					withSpec(baseClusterParameters()),
					withExternalName(testClusterName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateClusterFailed), errCreate),
				dax: fake.MockDaxClientCall{
					CreateClusterWithContext: []*fake.CallCreateClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateClusterInput{
								AvailabilityZones:             []*string{pointer.String(testAvailabilityZone), pointer.String(testOtherAvailabilityZone)},
								ClusterEndpointEncryptionType: pointer.String(testClusterEndpointEncryptionType),
								ClusterName:                   pointer.String(testClusterName),
								Description:                   pointer.String(testDescription),
								IamRoleArn:                    pointer.String(testIamRoleARN),
								NodeType:                      pointer.String(testNodeType),
								NotificationTopicArn:          pointer.String(testTopicARN),
								ParameterGroupName:            pointer.String(testParameterGroupName),
								PreferredMaintenanceWindow:    pointer.String(testPreferredMaintenanceWindow),
								ReplicationFactor:             pointer.Int64(testReplicationFactor),
								SSESpecification:              &dax.SSESpecification{Enabled: pointer.Bool(testSSESpecificationEnabled)},
								SecurityGroupIds:              []*string{pointer.String(testSecurityGroupIdentifier)},
								SubnetGroupName:               pointer.String(testSubnetGroupName),
								Tags: []*dax.Tag{{
									Key:   pointer.String(testTagKey),
									Value: pointer.String(testTagValue),
								}},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.dax, opts)
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *svcapitypes.Cluster
		err error
		dax fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDeleteClusterWithContext: func(c context.Context, dci *dax.DeleteClusterInput, o []request.Option) (*dax.DeleteClusterOutput, error) {
						return &dax.DeleteClusterOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testClusterName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withConditions(xpv1.Deleting()),
				),
				dax: fake.MockDaxClientCall{
					DeleteClusterWithContext: []*fake.CallDeleteClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteClusterInput{
								ClusterName: pointer.String(testClusterName),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDeleteClusterWithContext: func(c context.Context, dci *dax.DeleteClusterInput, o []request.Option) (*dax.DeleteClusterOutput, error) {
						return &dax.DeleteClusterOutput{}, errors.New(testErrDeleteClusterFailed)
					},
				},
				cr: instance(
					withExternalName(testClusterName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testClusterName),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteClusterFailed), errDelete),
				dax: fake.MockDaxClientCall{
					DeleteClusterWithContext: []*fake.CallDeleteClusterWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteClusterInput{
								ClusterName: pointer.String(testClusterName),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.dax, opts)
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
