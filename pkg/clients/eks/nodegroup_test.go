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
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	ngName                   = "my-cool-ng"
	amiType                  = "cool-ami"
	diskSize                 = int32(20)
	otherSize                = int32(100)
	size                     = int32(2)
	currentSize              = int32(5)
	maxSize                  = int32(8)
	nodeRole                 = "cool-role"
	otherVersion             = "1.17"
	releaseVersion           = "1.16.3-20220523"
	otherReleaseVersion      = "1.17.4-20220523"
	ltVersion                = "1"
	ltID                     = "lt-id"
	ltName                   = "my-cool-lt"
	otherLtVersion           = "2"
	maxUnavailablePercentage = int32(80)
	maxUnavailable           = int32(2)
	force                    = true
	dontForce                = false
)

func TestGenerateCreateNodeGroupInput(t *testing.T) {
	type args struct {
		name string
		p    *manualv1alpha1.NodeGroupParameters
	}

	cases := map[string]struct {
		args args
		want *eks.CreateNodegroupInput
	}{
		"AllFields": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailablePercentage,
						Force:                    &force,
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
				UpdateConfig: &ekstypes.NodegroupUpdateConfig{
					MaxUnavailablePercentage: &maxUnavailablePercentage,
				},
				Subnets: []string{"cool-subnet"},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
		"SomeFields": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:       &amiType,
					ClusterName:   clusterName,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					NodeRole:      nodeRole,
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
						Force:          &dontForce,
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
				UpdateConfig: &ekstypes.NodegroupUpdateConfig{
					MaxUnavailable: &maxUnavailable,
				},
				Subnets: []string{"cool-subnet"},
			},
		},
		"DefaultDesiredSize": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:       &amiType,
					ClusterName:   clusterName,
					DiskSize:      &diskSize,
					InstanceTypes: []string{"cool-type"},
					NodeRole:      nodeRole,
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateNodeGroupVersionInput(t *testing.T) {
	type args struct {
		name string
		p    *manualv1alpha1.NodeGroupParameters
		n    *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args
		wantUpdate bool
		wantInput  *eks.UpdateNodegroupVersionInput
	}{
		"NothingChanged": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
						Version: &ltVersion,
					},
				},
				n: &ekstypes.Nodegroup{
					ClusterName:    &clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						Version: &ltVersion,
					},
				},
			},
			wantUpdate: false,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:   &clusterName,
				NodegroupName: &ngName,
			}},
		"VersionChanged": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &otherVersion,
					ReleaseVersion: &releaseVersion,
				},
				n: &ekstypes.Nodegroup{
					NodegroupName:  &ngName,
					ClusterName:    &clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:   &clusterName,
				NodegroupName: &ngName,
				Version:       &otherVersion,
			},
		},
		"ReleaseVersionChanged": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &version,
					ReleaseVersion: &otherReleaseVersion,
				},
				n: &ekstypes.Nodegroup{
					NodegroupName:  &ngName,
					ClusterName:    &clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:    &clusterName,
				NodegroupName:  &ngName,
				ReleaseVersion: &otherReleaseVersion,
			},
		},
		"LaunchTemplateVersionChanged": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
						Name:    &ltName,
						Version: &otherLtVersion,
					},
				},
				n: &ekstypes.Nodegroup{
					ClusterName:    &clusterName,
					NodegroupName:  &ngName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						Version: &ltVersion,
					},
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:   &clusterName,
				NodegroupName: &ngName,
				LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
					Name:    &ltName,
					Version: &otherLtVersion,
				},
			},
		},
		"LaunchTemplateVersionChangedWithId": {
			// The launch template can be referenced via ID or via Name. If there is no Name we should use the ID.
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
						ID:      &ltID,
						Version: &otherLtVersion,
					},
				},
				n: &ekstypes.Nodegroup{
					ClusterName:    &clusterName,
					NodegroupName:  &ngName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						// Having a name here and not an ID is on purpose to
						// make sure that the k8s resource is "leading"
						Name:    &ltName,
						Version: &ltVersion,
					},
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:   &clusterName,
				NodegroupName: &ngName,
				LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
					Id:      &ltID,
					Version: &otherLtVersion,
				},
			},
		},
		"EverythingChanged": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName:    clusterName,
					Version:        &otherVersion,
					ReleaseVersion: &otherReleaseVersion,
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
						Name:    &ltName,
						Version: &otherLtVersion,
					},
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						Force: &dontForce,
					},
				},
				n: &ekstypes.Nodegroup{
					ClusterName:    &clusterName,
					NodegroupName:  &ngName,
					Version:        &version,
					ReleaseVersion: &releaseVersion,
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						Name:    &ltName,
						Version: &ltVersion,
					},
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:    &clusterName,
				NodegroupName:  &ngName,
				Version:        &otherVersion,
				ReleaseVersion: &otherReleaseVersion,
				LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
					Name:    &ltName,
					Version: &otherLtVersion,
				},
			},
		},
		"ForceUpdate": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName: clusterName,
					Version:     &otherVersion,
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						Force: &force,
					},
				},
				n: &ekstypes.Nodegroup{
					ClusterName:   &clusterName,
					NodegroupName: &ngName,
					Version:       &version,
				},
			},
			wantUpdate: true,
			wantInput: &eks.UpdateNodegroupVersionInput{
				ClusterName:   &clusterName,
				NodegroupName: &ngName,
				Version:       &otherVersion,
				Force:         true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gotUpdate, gotInput := GenerateUpdateNodeGroupVersionInput(tc.args.name, tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.wantUpdate, gotUpdate); diff != "" {
				t.Errorf("r: -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantInput, gotInput, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateNodeGroupInput(t *testing.T) {
	type args struct {
		name string
		p    *manualv1alpha1.NodeGroupParameters
		n    *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want *eks.UpdateNodegroupConfigInput
	}{
		"AllFieldsEmpty": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label", "key": "val"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
				ClusterName:   &clusterName,
				Labels:        nil,
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
				p: &manualv1alpha1.NodeGroupParameters{
					AMIType:        &amiType,
					ClusterName:    clusterName,
					DiskSize:       &diskSize,
					InstanceTypes:  []string{"cool-type"},
					Labels:         map[string]string{"cool": "label"},
					NodeRole:       nodeRole,
					ReleaseVersion: &version,
					RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
						EC2SSHKey:            &keyArn,
						SourceSecurityGroups: []string{"cool-group"},
					},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
						MaxSize: pointer.ToIntAsInt32(10),
						MinSize: pointer.ToIntAsInt32(6),
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
						DesiredSize: pointer.ToIntAsInt32(5),
						MaxSize:     pointer.ToIntAsInt32(10),
						MinSize:     pointer.ToIntAsInt32(3),
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				ClusterName:   &clusterName,
				Labels:        nil,
				NodegroupName: &ngName,
				ScalingConfig: &ekstypes.NodegroupScalingConfig{
					DesiredSize: pointer.ToIntAsInt32(6),
					MaxSize:     pointer.ToIntAsInt32(10),
					MinSize:     pointer.ToIntAsInt32(6),
				},
			},
		},
		"UpdateConfig": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName: clusterName,
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailable:           &maxUnavailablePercentage,
						MaxUnavailablePercentage: &maxUnavailable,
						Force:                    &force,
					},
				},
				n: &ekstypes.Nodegroup{
					NodegroupName: &ngName,
					ClusterName:   &clusterName,
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable:           &maxUnavailable,
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				NodegroupName: &ngName,
				ClusterName:   &clusterName,
				UpdateConfig: &ekstypes.NodegroupUpdateConfig{
					MaxUnavailable:           &maxUnavailablePercentage,
					MaxUnavailablePercentage: &maxUnavailable,
				},
			},
		},
		"DiffTaints": {
			args: args{
				name: ngName,
				p: &manualv1alpha1.NodeGroupParameters{
					ClusterName: clusterName,
					Taints: []manualv1alpha1.Taint{
						{
							Effect: "effect",
							Key:    ptr.To("toAdd"),
							Value:  ptr.To("value"),
						},
						{
							Effect: "effect",
							Key:    ptr.To("toChange"),
							Value:  ptr.To("newValue"),
						},
						{
							Effect: "effect",
							Key:    ptr.To("toKeep"),
							Value:  ptr.To("value"),
						},
					},
				},
				n: &ekstypes.Nodegroup{
					NodegroupName: &ngName,
					ClusterName:   &clusterName,
					Taints: []ekstypes.Taint{
						{
							Effect: "effect",
							Key:    ptr.To("toKeep"),
							Value:  ptr.To("value"),
						},
						{
							Effect: "effect",
							Key:    ptr.To("toRemove"),
							Value:  ptr.To("value"),
						},
						{
							Effect: "effect",
							Key:    ptr.To("toChange"),
							Value:  ptr.To("oldValue"),
						},
					},
				},
			},
			want: &eks.UpdateNodegroupConfigInput{
				NodegroupName: &ngName,
				ClusterName:   &clusterName,
				Taints: &ekstypes.UpdateTaintsPayload{
					AddOrUpdateTaints: []ekstypes.Taint{
						{
							Effect: "effect",
							Key:    ptr.To("toAdd"),
							Value:  ptr.To("value"),
						},
						{
							Effect: "effect",
							Key:    ptr.To("toChange"),
							Value:  ptr.To("newValue"),
						},
					},
					RemoveTaints: []ekstypes.Taint{
						{
							Effect: "effect",
							Key:    ptr.To("toRemove"),
							Value:  ptr.To("value"),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateNodeGroupConfigInput(tc.args.name, tc.args.p, tc.args.n)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
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
		want manualv1alpha1.NodeGroupObservation
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
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable:           &maxUnavailable,
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
			},
			want: manualv1alpha1.NodeGroupObservation{
				NodeGroupArn: ngArn,
				Status:       manualv1alpha1.NodeGroupStatusActive,
				CreatedAt:    &v1.Time{Time: now},
				Health: manualv1alpha1.NodeGroupHealth{
					Issues: []manualv1alpha1.Issue{
						{
							Code:        "AccessDenied",
							Message:     message,
							ResourceIDs: []string{"my-resource"},
						},
					},
				},
				ModifiedAt: &v1.Time{Time: now},
				Resources: manualv1alpha1.NodeGroupResources{
					RemoteAccessSecurityGroup: rasg,
					AutoScalingGroups: []manualv1alpha1.AutoScalingGroup{
						{
							Name: asg,
						},
					},
				},
				ScalingConfig: manualv1alpha1.NodeGroupScalingConfigStatus{
					DesiredSize: &size,
				},
				UpdateConfig: manualv1alpha1.NodeGroupUpdateConfigStatus{
					MaxUnavailable:           &maxUnavailable,
					MaxUnavailablePercentage: &maxUnavailablePercentage,
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
		p *manualv1alpha1.NodeGroupParameters
		n *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want *manualv1alpha1.NodeGroupParameters
	}{
		"AllFieldsEmpty": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{},
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
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: &manualv1alpha1.NodeGroupParameters{
				AMIType:        &ami,
				DiskSize:       &diskSize,
				InstanceTypes:  []string{"cool-type"},
				Labels:         map[string]string{"cool": "label"},
				ReleaseVersion: &version,
				RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
					EC2SSHKey:            &keyArn,
					SourceSecurityGroups: []string{"cool-group"},
				},
				ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
					DesiredSize: &size,
					MaxSize:     &size,
					MinSize:     &size,
				},
				UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
					MaxUnavailable: &maxUnavailable,
					Force:          &dontForce,
				},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
		"IgnoreDesiredSize": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
					ReleaseVersion: &version,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: &manualv1alpha1.NodeGroupParameters{
				AMIType:        &ami,
				DiskSize:       &diskSize,
				InstanceTypes:  []string{"cool-type"},
				Labels:         map[string]string{"cool": "label"},
				ReleaseVersion: &version,
				RemoteAccess: &manualv1alpha1.RemoteAccessConfig{
					EC2SSHKey:            &keyArn,
					SourceSecurityGroups: []string{"cool-group"},
				},
				ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
					MaxSize: &maxSize,
					MinSize: &size,
				},
				UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
					MaxUnavailable: &maxUnavailable,
					Force:          &dontForce,
				},
				Tags:    map[string]string{"cool": "tag"},
				Version: &version,
			},
		},
		"SetLaunchTemplateVersionIfEmpty": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{},
				},
				n: &ekstypes.Nodegroup{
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						Version: &ltVersion,
					},
				},
			},
			want: &manualv1alpha1.NodeGroupParameters{
				LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
					Version: &ltVersion,
				},
				UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
					Force: &dontForce,
				},
			},
		},
		"UpdateConfigDefaultDontForce": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
				n: &ekstypes.Nodegroup{
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
			},
			want: &manualv1alpha1.NodeGroupParameters{
				UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
					MaxUnavailablePercentage: &maxUnavailablePercentage,
					Force:                    &dontForce,
				},
			},
		},
		"UpdateConfigForceIfWanted": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						Force: &force,
					},
				},
				n: &ekstypes.Nodegroup{
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
				},
			},
			want: &manualv1alpha1.NodeGroupParameters{
				UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
					Force:          &force,
					MaxUnavailable: &maxUnavailable,
				},
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
	type args struct {
		p *manualv1alpha1.NodeGroupParameters
		n *ekstypes.Nodegroup
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
						Force:          &dontForce,
					},
				},
				n: &ekstypes.Nodegroup{
					Labels: map[string]string{"cool": "label"},
					ScalingConfig: &ekstypes.NodegroupScalingConfig{
						DesiredSize: &size,
						MaxSize:     &size,
						MinSize:     &size,
					},
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
					Version: &version,
					Tags:    map[string]string{"cool": "tag"},
				},
			},
			want: true,
		},
		"UpdateTags": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag", "another": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:           map[string]string{"cool": "tag"},
					ReleaseVersion: &releaseVersion,
					Version:        &otherVersion,
					Labels:         map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
					ReleaseVersion: &releaseVersion,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: false,
		},
		"UpdateReleaseVersion": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:           map[string]string{"cool": "tag"},
					Version:        &version,
					ReleaseVersion: &otherReleaseVersion,
					Labels:         map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
					Version:        &releaseVersion,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: false,
		},
		"UpdateLaunchTemplate": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					LaunchTemplate: &manualv1alpha1.LaunchTemplateSpecification{
						Version: &otherLtVersion,
					},
				},
				n: &ekstypes.Nodegroup{
					LaunchTemplate: &ekstypes.LaunchTemplateSpecification{
						Version: &ltVersion,
					},
				},
			},
			want: false,
		},
		"UpdateScaling": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:    map[string]string{"cool": "tag"},
					Version: &version,
					Labels:  map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
				p: &manualv1alpha1.NodeGroupParameters{
					Tags:           map[string]string{"cool": "tag"},
					ReleaseVersion: &releaseVersion,
					Version:        &version,
					Labels:         map[string]string{"cool": "label"},
					ScalingConfig: &manualv1alpha1.NodeGroupScalingConfig{
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
					ReleaseVersion: &releaseVersion,
					Version:        &version,
					Tags:           map[string]string{"cool": "tag"},
				},
			},
			want: true,
		},
		"UpdateConfigMaxUnavailable": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
				},
				n: &ekstypes.Nodegroup{
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailablePercentage,
					},
				},
			},
			want: false,
		},
		"UpdateConfigMaxUnavailablePercentage": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailable,
					},
				},
				n: &ekstypes.Nodegroup{
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
			},
			want: false,
		},
		"UpdateConfigAbsoluteToPercentage": {
			args: args{
				p: &manualv1alpha1.NodeGroupParameters{
					UpdateConfig: &manualv1alpha1.NodeGroupUpdateConfig{
						MaxUnavailablePercentage: &maxUnavailablePercentage,
					},
				},
				n: &ekstypes.Nodegroup{
					UpdateConfig: &ekstypes.NodegroupUpdateConfig{
						MaxUnavailable: &maxUnavailable,
					},
				},
			},
			want: false,
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
