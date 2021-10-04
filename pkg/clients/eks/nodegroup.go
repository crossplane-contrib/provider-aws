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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

// GenerateCreateNodeGroupInput from NodeGroupParameters.
func GenerateCreateNodeGroupInput(name string, p *v1alpha1.NodeGroupParameters) *eks.CreateNodegroupInput {
	c := &eks.CreateNodegroupInput{
		NodegroupName:  &name,
		AmiType:        ekstypes.AMITypes(awsclient.StringValue(p.AMIType)),
		ClusterName:    &p.ClusterName,
		DiskSize:       p.DiskSize,
		InstanceTypes:  p.InstanceTypes,
		Labels:         p.Labels,
		NodeRole:       &p.NodeRole,
		ReleaseVersion: p.ReleaseVersion,
		Subnets:        p.Subnets,
		Tags:           p.Tags,
		Version:        p.Version,
	}
	if p.RemoteAccess != nil {
		c.RemoteAccess = &ekstypes.RemoteAccessConfig{
			Ec2SshKey:            p.RemoteAccess.EC2SSHKey,
			SourceSecurityGroups: p.RemoteAccess.SourceSecurityGroups,
		}
	}
	if p.ScalingConfig != nil {
		c.ScalingConfig = &ekstypes.NodegroupScalingConfig{
			DesiredSize: p.ScalingConfig.DesiredSize,
			MinSize:     p.ScalingConfig.MinSize,
			MaxSize:     p.ScalingConfig.MaxSize,
		}

		// NOTE(mcavoyk): desizedSize is a required field for AWS, to support node scaling actions
		// outside of this provider, we allow desiredSize to be nil so the field can be ignored when
		// checking if the NodeGroup is up-to-date. If the field is nil, we set the desiredSize equal
		// to the minSize as an initial value.
		if p.ScalingConfig.DesiredSize == nil {
			c.ScalingConfig.DesiredSize = p.ScalingConfig.MinSize
		}
	}
	return c
}

// GenerateUpdateNodeGroupConfigInput from NodeGroupParameters.
func GenerateUpdateNodeGroupConfigInput(name string, p *v1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) *eks.UpdateNodegroupConfigInput {
	u := &eks.UpdateNodegroupConfigInput{
		NodegroupName: &name,
		ClusterName:   &p.ClusterName,
	}

	if len(p.Labels) > 0 {
		addOrModify, remove := awsclient.DiffLabels(p.Labels, ng.Labels)
		u.Labels = &ekstypes.UpdateLabelsPayload{
			AddOrUpdateLabels: addOrModify,
			RemoveLabels:      remove,
		}
	}
	if p.ScalingConfig != nil {
		u.ScalingConfig = &ekstypes.NodegroupScalingConfig{
			DesiredSize: p.ScalingConfig.DesiredSize,
			MinSize:     p.ScalingConfig.MinSize,
			MaxSize:     p.ScalingConfig.MaxSize,
		}

		// If desiredSize is not set, derive the value from either the
		// current observed desiredSize, or the min/max if observed is out of bounds.
		if p.ScalingConfig.DesiredSize == nil {
			// The min/max size set the floor/ceiling for the desiredSize
			switch desiredSizeVal := aws.ToInt32(ng.ScalingConfig.DesiredSize); {
			case desiredSizeVal < aws.ToInt32(p.ScalingConfig.MinSize):
				u.ScalingConfig.DesiredSize = p.ScalingConfig.MinSize
			case desiredSizeVal > aws.ToInt32(p.ScalingConfig.MaxSize):
				u.ScalingConfig.DesiredSize = p.ScalingConfig.MaxSize
			default:
				u.ScalingConfig.DesiredSize = ng.ScalingConfig.DesiredSize
			}
		}
	}
	return u
}

// GenerateNodeGroupObservation is used to produce v1alpha1.NodeGroupObservation
// from eks.Nodegroup.
func GenerateNodeGroupObservation(ng *ekstypes.Nodegroup) v1alpha1.NodeGroupObservation { // nolint:gocyclo
	if ng == nil {
		return v1alpha1.NodeGroupObservation{}
	}
	o := v1alpha1.NodeGroupObservation{
		NodeGroupArn: awsclient.StringValue(ng.NodegroupArn),
		Status:       v1alpha1.NodeGroupStatusType(ng.Status),
	}
	if ng.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *ng.CreatedAt}
	}
	if ng.Health != nil && len(ng.Health.Issues) > 0 {
		o.Health = v1alpha1.NodeGroupHealth{
			Issues: make([]v1alpha1.Issue, len(ng.Health.Issues)),
		}
		for c, i := range ng.Health.Issues {
			o.Health.Issues[c] = v1alpha1.Issue{
				Code:        string(i.Code),
				Message:     awsclient.StringValue(i.Message),
				ResourceIDs: i.ResourceIds,
			}
		}
	}
	if ng.ModifiedAt != nil {
		o.ModifiedAt = &metav1.Time{Time: *ng.ModifiedAt}
	}
	if ng.Resources != nil {
		o.Resources = v1alpha1.NodeGroupResources{
			RemoteAccessSecurityGroup: awsclient.StringValue(ng.Resources.RemoteAccessSecurityGroup),
		}
		if len(ng.Resources.AutoScalingGroups) > 0 {
			asg := make([]v1alpha1.AutoScalingGroup, len(ng.Resources.AutoScalingGroups))
			for c, a := range ng.Resources.AutoScalingGroups {
				asg[c] = v1alpha1.AutoScalingGroup{Name: awsclient.StringValue(a.Name)}
			}
			o.Resources.AutoScalingGroups = asg
		}
	}

	if ng.ScalingConfig != nil {
		o.ScalingConfig = v1alpha1.NodeGroupScalingConfigStatus{
			DesiredSize: ng.ScalingConfig.DesiredSize,
		}
	}
	return o
}

// LateInitializeNodeGroup fills the empty fields in *v1alpha1.NodeGroupParameters with the
// values seen in eks.Nodegroup.
func LateInitializeNodeGroup(in *v1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) { // nolint:gocyclo
	if ng == nil {
		return
	}
	in.AMIType = awsclient.LateInitializeStringPtr(in.AMIType, awsclient.String(string(ng.AmiType)))
	in.DiskSize = awsclient.LateInitializeInt32Ptr(in.DiskSize, ng.DiskSize)
	if len(in.InstanceTypes) == 0 && len(ng.InstanceTypes) > 0 {
		in.InstanceTypes = ng.InstanceTypes
	}
	if len(in.Labels) == 0 && len(ng.Labels) > 0 {
		in.Labels = ng.Labels
	}
	if in.RemoteAccess == nil && ng.RemoteAccess != nil {
		in.RemoteAccess = &v1alpha1.RemoteAccessConfig{
			EC2SSHKey:            ng.RemoteAccess.Ec2SshKey,
			SourceSecurityGroups: ng.RemoteAccess.SourceSecurityGroups,
		}
	}
	if in.ScalingConfig == nil && ng.ScalingConfig != nil {
		in.ScalingConfig = &v1alpha1.NodeGroupScalingConfig{
			DesiredSize: ng.ScalingConfig.DesiredSize,
			MinSize:     ng.ScalingConfig.MinSize,
			MaxSize:     ng.ScalingConfig.MaxSize,
		}
	}
	in.ReleaseVersion = awsclient.LateInitializeStringPtr(in.ReleaseVersion, ng.ReleaseVersion)
	in.Version = awsclient.LateInitializeStringPtr(in.Version, ng.Version)
	// NOTE(hasheddan): we always will set the default Crossplane tags in
	// practice during initialization in the controller, but we check if no tags
	// exist for consistency with expected late initialization behavior.
	if len(in.Tags) == 0 {
		in.Tags = ng.Tags
	}
}

// IsNodeGroupUpToDate checks whether there is a change in any of the modifiable fields.
func IsNodeGroupUpToDate(p *v1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) bool { // nolint:gocyclo
	if !cmp.Equal(p.Tags, ng.Tags, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(p.Version, ng.Version) {
		return false
	}
	if !cmp.Equal(p.Labels, ng.Labels, cmpopts.EquateEmpty()) {
		return false
	}
	if p.ScalingConfig == nil && ng.ScalingConfig == nil {
		return true
	}
	if p.ScalingConfig != nil && ng.ScalingConfig != nil {
		if p.ScalingConfig.DesiredSize != nil &&
			aws.ToInt32(p.ScalingConfig.DesiredSize) != aws.ToInt32(ng.ScalingConfig.DesiredSize) {
			return false
		}
		if !cmp.Equal(p.ScalingConfig.MaxSize, ng.ScalingConfig.MaxSize) {
			return false
		}
		if !cmp.Equal(p.ScalingConfig.MinSize, ng.ScalingConfig.MinSize) {
			return false
		}
		return true
	}
	return false
}
