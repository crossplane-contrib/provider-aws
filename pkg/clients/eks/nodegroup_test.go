/*
Copyright 2020 The Crossplane Authors.

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

package eks

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	ngName      = "my-cool-ng"
	amiType     = "cool-ami"
	diskSize    = int32(20)
	size        = int32(2)
	currentSize = int32(5)
	maxSize     = int32(8)
	nodeRole    = "cool-role"
)

func TestGenerateCreateNodeGroupInput(t *testing.T) {
	type args struct {
		name string
		p    *v1alpha1.NodeGroupParameters
	}

	cases := map[string]struct {
		args args
		want *eks.CreateNodegroupInput
	}{
		"AllFields": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &v1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Subnets: []string{"cool-subnet"},
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
				},
			},
			want: &eks.CreateNodegroupInput{
				AmiType:        ekstypes.AMITypes(amiType),
				ClusterName:    &clusterName,
				DiskSize:       &diskSize,
				InstanceTypes:  []string{"cool-type"},
				Labels:         map[string]string{"cool": "label"},
				NodeRole:       &nodeRole,
				NodegroupName:  &ngName,
				ReleaseVersion: &version,
				RemoteAccess: &ekstypes.RemoteAccessConfig{
					Ec2SshKey:            &keyArn,
					SourceSecurityGroups: []string{"cool-group"},
				},
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
				Subnets: []string{"cool-subnet"},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
		"SomeFields": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:       &amiType,
					ClusterName:   clusterName,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					NodeRole:      nodeRole,
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Subnets: []string{"cool-subnet"},
				},
			},
			want: &eks.CreateNodegroupInput{
				AmiType:       ekstypes.AMITypes(amiType),
				ClusterName:   &clusterName,
				DiskSize:      &diskSize,
				InstanceTypes: []string{"cool-type"},
				NodeRole:      &nodeRole,
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
				Subnets: []string{"cool-subnet"},
			},
		},
		"DefaultDesiredSize": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:       &amiType,
					ClusterName:   clusterName,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					NodeRole:      nodeRole,
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						MaxSize: &maxSize,
						MinSize: &size,
					},
					Subnets: []string{"cool-subnet"},
				},
			},
			want: &eks.CreateNodegroupInput{
				AmiType:       ekstypes.AMITypes(amiType),
				ClusterName:   &clusterName,
				DiskSize:      &diskSize,
				InstanceTypes: []string{"cool-type"},
				NodeRole:      &nodeRole,
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &maxSize,
					MinSize:     &size,
				},
				Subnets: []string{"cool-subnet"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateNodeGroupInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateNodeGroupInput(t *testing.T) {
	type args struct {
		name string
		p    *v1alpha1.NodeGroupParameters
		n    *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want *eks.UpdateNodegroupConfigInput
	}{
		"AllFieldsEmpty": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &v1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Subnets: []string{"cool-subnet"},
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
				},
				n: &ekstypes.Nodegroup{},
			},
			want: &eks.UpdateNodegroupConfigInput{
				ClusterName: &clusterName,
				Labels: &ekstypes.UpdateLabelsPayload{
					AddOrUpdateLabels: map[string]string{"cool": "label"},
					RemoveLabels:      []string{},
				},
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
			},
		},
		"LabelsOnly": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label", "key": "val"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &v1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Subnets: []string{"cool-subnet"},
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"remove": "label", "key": "badval"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				ClusterName: &clusterName,
				Labels: &ekstypes.UpdateLabelsPayload{
					AddOrUpdateLabels: map[string]string{"cool": "label", "key": "val"},
					RemoveLabels:      []string{"remove"},
				},
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
			},
		},
		"IgnoreDesiredSize": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &v1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						MaxSize: &maxSize,
						MinSize: &size,
					},
					Subnets: []string{"cool-subnet"},
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
				},
				n: &ekstypes.Nodegroup{
					ClusterName:   &clusterName,
					NodegroupName: &ngName,
					Labels:        map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &currentSize,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				ClusterName: &clusterName,
				Labels: &ekstypes.UpdateLabelsPayload{
					AddOrUpdateLabels: map[string]string{},
					RemoveLabels:      []string{},
				},
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: &currentSize,
					MaxSize:     &maxSize,
					MinSize:     &size,
				},
			},
		},
		"BoundDesiredSize": {
			args: args{
				name: ngName,
				p: &v1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &v1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						MaxSize: awsclients.Int32(10),
						MinSize: awsclients.Int32(6),
					},
					Subnets: []string{"cool-subnet"},
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
				},
				n: &ekstypes.Nodegroup{
					ClusterName:   &clusterName,
					NodegroupName: &ngName,
					Labels:        map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: awsclients.Int32(5),
						MaxSize:     awsclients.Int32(10),
						MinSize:     awsclients.Int32(3),
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				ClusterName: &clusterName,
				Labels: &ekstypes.UpdateLabelsPayload{
					AddOrUpdateLabels: map[string]string{},
					RemoveLabels:      []string{},
				},
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: awsclients.Int32(6),
					MaxSize:     awsclients.Int32(10),
					MinSize:     awsclients.Int32(6),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateNodeGroupConfigInput(tc.args.name, tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateNodeObservation(t *testing.T) {
	ngArn := "cool:arn"
	now := time.Now()
	message := "cool message"
	rasg := "cool-sg"
	asg := "my-asg"

	type args struct {
		n *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want v1alpha1.NodeGroupObservation
	}{
		"Full": {
			args: args{
				n: &ekstypes.Nodegroup{
					NodegroupArn: &ngArn,
					Status:       ekstypes.NodegroupStatusActive,
					CreatedAt:    &now,
					Health: &ekstypes.NodegroupHealth{
						Issues: []ekstypes.Issue{
							{
								Code:        ekstypes.NodegroupIssueCodeAccessDenied,
								Message:     &message,
								ResourceIds: []string{"my-resource"},
							},
						},
					},
					ModifiedAt: &now,
					Resources: &ekstypes.NodegroupResources{
						RemoteAccessSecurityGroup: &rasg,
						AutoScalingGroups: []ekstypes.AutoScalingGroup{
							{
								Name: &asg,
							},
						},
					},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
				},
			},
			want: v1alpha1.NodeGroupObservation{
				NodeGroupArn: ngArn,
				Status:       v1alpha1.NodeGroupStatusActive,
				CreatedAt:    &v1.Time{Time: now},
				Health: v1alpha1.NodeGroupHealth{
					Issues: []v1alpha1.Issue{
						{
							Code:        "AccessDenied",
							Message:     message,
							ResourceIDs: []string{"my-resource"},
						},
					},
				},
				ModifiedAt: &v1.Time{Time: now},
				Resources: v1alpha1.NodeGroupResources{
					RemoteAccessSecurityGroup: rasg,
					AutoScalingGroups: []v1alpha1.AutoScalingGroup{
						{
							Name: asg,
						},
					},
				},
				ScalingConfig: v1alpha1.NodeGroupScalingConfigStatus{
					DesiredSize: &size,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateNodeGroupObservation(tc.args.n)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeNodeGroup(t *testing.T) {
	ami := "AL2_x86_64"
	type args struct {
		p *v1alpha1.NodeGroupParameters
		n *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want *v1alpha1.NodeGroupParameters
	}{
		"AllFieldsEmpty": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{},
				n: &ekstypes.Nodegroup{
					AmiType:       ekstypes.AMITypesAl2X8664,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					Labels:        map[string]string{"cool": "label"},
					RemoteAccess: &ekstypes.RemoteAccessConfig{
						Ec2SshKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: &v1alpha1.NodeGroupParameters{
				AMIType:        &ami,
				DiskSize:       &diskSize,
				InstanceTypes:  []string{"cool-type"},
				Labels:         map[string]string{"cool": "label"},
				ReleaseVersion: &version,
				RemoteAccess: &v1alpha1.RemoteAccessConfig{
					EC2SSHKey:            &keyArn,
					SourceSecurityGroups: []string{"cool-group"},
				},
				ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
		"IgnoreDesiredSize": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: nil,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					AmiType:       ekstypes.AMITypesAl2X8664,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					Labels:        map[string]string{"cool": "label"},
					RemoteAccess: &ekstypes.RemoteAccessConfig{
						Ec2SshKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: &v1alpha1.NodeGroupParameters{
				AMIType:        &ami,
				DiskSize:       &diskSize,
				InstanceTypes:  []string{"cool-type"},
				Labels:         map[string]string{"cool": "label"},
				ReleaseVersion: &version,
				RemoteAccess: &v1alpha1.RemoteAccessConfig{
					EC2SSHKey:            &keyArn,
					SourceSecurityGroups: []string{"cool-group"},
				},
				ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
					MaxSize: &maxSize,
					MinSize: &size,
				},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeNodeGroup(tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, tc.args.p); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsNodeGroupUpToDate(t *testing.T) {
	otherVersion := "1.17"
	otherSize := int32(100)

	type args struct {
		p *v1alpha1.NodeGroupParameters
		n *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Version: &version,
					Tags:    map[string]string{"cool": "tag"},
				},
			},
			want: true,
		},
		"UpdateTags": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag", "another": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					Version: &version,
					Tags:    map[string]string{"cool": "tag"},
				},
			},
			want: false,
		},
		"UpdateVersion": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &otherVersion,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: false,
		},
		"UpdateScaling": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &otherSize,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: false,
		},
		"IgnoreDesiredSize": {
			args: args{
				p: &v1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &v1alpha1.NodeGroupScalingConfig{
						DesiredSize: nil,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &currentSize,
						MaxSize:     &maxSize,
						MinSize:     &size,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			upToDate := IsNodeGroupUpToDate(tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, upToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
