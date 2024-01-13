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
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/labels"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// GenerateCreateNodeGroupInput from NodeGroupParameters.
func GenerateCreateNodeGroupInput(name string, p *manualv1alpha1.NodeGroupParameters) *eks.CreateNodegroupInput {
	c := &eks.CreateNodegroupInput{
		NodegroupName:  &name,
		AmiType:        ekstypes.AMITypes(pointer.StringValue(p.AMIType)),
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
	if p.CapacityType != nil {
		c.CapacityType = ekstypes.CapacityTypes(*p.CapacityType)
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
	if p.UpdateConfig != nil {
		c.UpdateConfig = &ekstypes.NodegroupUpdateConfig{
			MaxUnavailable:           p.UpdateConfig.MaxUnavailable,
			MaxUnavailablePercentage: p.UpdateConfig.MaxUnavailablePercentage,
		}
	}
	if p.LaunchTemplate != nil {
		c.LaunchTemplate = &ekstypes.LaunchTemplateSpecification{
			Id:      p.LaunchTemplate.ID,
			Name:    p.LaunchTemplate.Name,
			Version: p.LaunchTemplate.Version,
		}
	}
	if len(p.Taints) != 0 {
		c.Taints = make([]ekstypes.Taint, len(p.Taints))
		for i, t := range p.Taints {
			c.Taints[i] = ekstypes.Taint{
				Effect: ekstypes.TaintEffect(t.Effect),
				Key:    t.Key,
				Value:  t.Value,
			}
		}
	}
	return c
}

// GenerateUpdateNodeGroupVersionInput will check if version properties of the
// nodegroup have changed. If no version property has changed, it will return false
// and an incomplete update object. If true,
func GenerateUpdateNodeGroupVersionInput(name string, p *manualv1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) (bool, *eks.UpdateNodegroupVersionInput) {
	u := false
	i := &eks.UpdateNodegroupVersionInput{
		NodegroupName: &name,
		ClusterName:   &p.ClusterName,
	}
	if !notNilAndEquals(p.AMIType, "CUSTOM") {
		if !reflect.DeepEqual(p.Version, ng.Version) {
			u = true
			i.Version = p.Version
		}
		if !reflect.DeepEqual(p.ReleaseVersion, ng.ReleaseVersion) {
			u = true
			i.ReleaseVersion = p.ReleaseVersion
		}
	}
	if p.LaunchTemplate != nil && ng.LaunchTemplate != nil {
		// Since we late initialize the LaunchTemplate Version we can be sure that
		// there is a version set at this point
		if !reflect.DeepEqual(p.LaunchTemplate.Version, ng.LaunchTemplate.Version) {
			u = true
			i.LaunchTemplate = &ekstypes.LaunchTemplateSpecification{
				Version: p.LaunchTemplate.Version,
			}
			if p.LaunchTemplate.Name != nil {
				i.LaunchTemplate.Name = p.LaunchTemplate.Name
			} else {
				i.LaunchTemplate.Id = p.LaunchTemplate.ID
			}
		}
	}
	if p.UpdateConfig != nil && *p.UpdateConfig.Force {
		i.Force = true
	}

	return u, i
}

// GenerateUpdateNodeGroupConfigInput from NodeGroupParameters.
func GenerateUpdateNodeGroupConfigInput(name string, p *manualv1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) *eks.UpdateNodegroupConfigInput {
	u := &eks.UpdateNodegroupConfigInput{
		NodegroupName: &name,
		ClusterName:   &p.ClusterName,
	}

	if len(p.Labels) > 0 {
		addOrModify, remove := labels.DiffLabels(p.Labels, ng.Labels)
		// error: both or either addOrUpdateLabels or removeLabels must not be empty
		if len(addOrModify) > 0 || len(remove) > 0 {
			u.Labels = &ekstypes.UpdateLabelsPayload{
				AddOrUpdateLabels: addOrModify,
				RemoveLabels:      remove,
			}
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
	if p.UpdateConfig != nil {
		u.UpdateConfig = &ekstypes.NodegroupUpdateConfig{
			MaxUnavailable:           p.UpdateConfig.MaxUnavailable,
			MaxUnavailablePercentage: p.UpdateConfig.MaxUnavailablePercentage,
		}
	}

	if p.Taints != nil {
		u.Taints = generateUpdateTaintsPayload(p.Taints, ng.Taints)
	}
	return u
}

func generateUpdateTaintsPayload(spec []manualv1alpha1.Taint, current []ekstypes.Taint) *ekstypes.UpdateTaintsPayload {
	res := &ekstypes.UpdateTaintsPayload{}

	curMap := map[string]manualv1alpha1.Taint{}
	for _, t := range current {
		if t.Key == nil {
			continue
		}
		curMap[*t.Key] = manualv1alpha1.Taint{
			Effect: string(t.Effect),
			Key:    t.Key,
			Value:  t.Value,
		}
	}

	specMap := map[string]any{}
	for _, st := range spec {
		if st.Key == nil {
			continue
		}
		specMap[*st.Key] = nil

		ct, exists := curMap[*st.Key]
		if !exists || !cmp.Equal(st, ct) {
			res.AddOrUpdateTaints = append(res.AddOrUpdateTaints, ekstypes.Taint{
				Effect: ekstypes.TaintEffect(st.Effect),
				Key:    st.Key,
				Value:  st.Value,
			})
		}
	}
	for _, ct := range current {
		if ct.Key == nil {
			continue
		}
		if _, exists := specMap[*ct.Key]; !exists {
			res.RemoveTaints = append(res.RemoveTaints, ct)
		}
	}
	return res
}

// GenerateNodeGroupObservation is used to produce manualv1alpha1.NodeGroupObservation
// from eks.Nodegroup.
func GenerateNodeGroupObservation(ng *ekstypes.Nodegroup) manualv1alpha1.NodeGroupObservation { //nolint:gocyclo
	if ng == nil {
		return manualv1alpha1.NodeGroupObservation{}
	}
	o := manualv1alpha1.NodeGroupObservation{
		NodeGroupArn:   pointer.StringValue(ng.NodegroupArn),
		Version:        pointer.StringValue(ng.Version),
		ReleaseVersion: pointer.StringValue(ng.ReleaseVersion),
		Status:         manualv1alpha1.NodeGroupStatusType(ng.Status),
	}
	if ng.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *ng.CreatedAt}
	}
	if ng.Health != nil && len(ng.Health.Issues) > 0 {
		o.Health = manualv1alpha1.NodeGroupHealth{
			Issues: make([]manualv1alpha1.Issue, len(ng.Health.Issues)),
		}
		for c, i := range ng.Health.Issues {
			o.Health.Issues[c] = manualv1alpha1.Issue{
				Code:        string(i.Code),
				Message:     pointer.StringValue(i.Message),
				ResourceIDs: i.ResourceIds,
			}
		}
	}
	if ng.ModifiedAt != nil {
		o.ModifiedAt = &metav1.Time{Time: *ng.ModifiedAt}
	}
	if ng.Resources != nil {
		o.Resources = manualv1alpha1.NodeGroupResources{
			RemoteAccessSecurityGroup: pointer.StringValue(ng.Resources.RemoteAccessSecurityGroup),
		}
		if len(ng.Resources.AutoScalingGroups) > 0 {
			asg := make([]manualv1alpha1.AutoScalingGroup, len(ng.Resources.AutoScalingGroups))
			for c, a := range ng.Resources.AutoScalingGroups {
				asg[c] = manualv1alpha1.AutoScalingGroup{Name: pointer.StringValue(a.Name)}
			}
			o.Resources.AutoScalingGroups = asg
		}
	}

	if ng.ScalingConfig != nil {
		o.ScalingConfig = manualv1alpha1.NodeGroupScalingConfigStatus{
			DesiredSize: ng.ScalingConfig.DesiredSize,
		}
	}

	if ng.UpdateConfig != nil {
		o.UpdateConfig = manualv1alpha1.NodeGroupUpdateConfigStatus{
			MaxUnavailable:           ng.UpdateConfig.MaxUnavailable,
			MaxUnavailablePercentage: ng.UpdateConfig.MaxUnavailablePercentage,
		}
	}
	return o
}

// LateInitializeNodeGroup fills the empty fields in *manualv1alpha1.NodeGroupParameters with the
// values seen in eks.Nodegroup.
func LateInitializeNodeGroup(in *manualv1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) { //nolint:gocyclo
	if ng == nil {
		return
	}
	in.AMIType = pointer.LateInitialize(in.AMIType, pointer.ToOrNilIfZeroValue(string(ng.AmiType)))
	in.CapacityType = pointer.LateInitialize(in.CapacityType, pointer.ToOrNilIfZeroValue(string(ng.CapacityType)))
	in.DiskSize = pointer.LateInitialize(in.DiskSize, ng.DiskSize)
	if len(in.InstanceTypes) == 0 && len(ng.InstanceTypes) > 0 {
		in.InstanceTypes = ng.InstanceTypes
	}
	if len(in.Labels) == 0 && len(ng.Labels) > 0 {
		in.Labels = ng.Labels
	}
	if in.RemoteAccess == nil && ng.RemoteAccess != nil {
		in.RemoteAccess = &manualv1alpha1.RemoteAccessConfig{
			EC2SSHKey:            ng.RemoteAccess.Ec2SshKey,
			SourceSecurityGroups: ng.RemoteAccess.SourceSecurityGroups,
		}
	}
	if in.ScalingConfig == nil && ng.ScalingConfig != nil {
		in.ScalingConfig = &manualv1alpha1.NodeGroupScalingConfig{
			DesiredSize: ng.ScalingConfig.DesiredSize,
			MinSize:     ng.ScalingConfig.MinSize,
			MaxSize:     ng.ScalingConfig.MaxSize,
		}
	}
	// We will always have an UpdateConfig, because AWS will always create one
	// and even if not - will create one to add the default value for force.
	// We do not need to late init maxUnavailablePercentage. If it is set, it
	// must already be part of the Object, since there is no default for that
	// value.
	if in.UpdateConfig == nil {
		in.UpdateConfig = &manualv1alpha1.NodeGroupUpdateConfig{}
	}
	in.UpdateConfig.Force = pointer.LateInitialize(in.UpdateConfig.Force, aws.Bool(false))
	if ng.UpdateConfig != nil {
		in.UpdateConfig.MaxUnavailable = pointer.LateInitialize(in.UpdateConfig.MaxUnavailable, ng.UpdateConfig.MaxUnavailable)
	}
	if in.LaunchTemplate != nil && ng.LaunchTemplate != nil && ng.LaunchTemplate.Version != nil {
		in.LaunchTemplate.Version = pointer.LateInitialize(in.LaunchTemplate.Version, ng.LaunchTemplate.Version)
	}
	in.ReleaseVersion = pointer.LateInitialize(in.ReleaseVersion, ng.ReleaseVersion)
	in.Version = pointer.LateInitialize(in.Version, ng.Version)
	// NOTE(hasheddan): we always will set the default Crossplane tags in
	// practice during initialization in the controller, but we check if no tags
	// exist for consistency with expected late initialization behavior.
	if len(in.Tags) == 0 {
		in.Tags = ng.Tags
	}
	if len(in.Taints) == 0 && len(ng.Taints) != 0 {
		in.Taints = make([]manualv1alpha1.Taint, len(ng.Taints))
		for i, t := range ng.Taints {
			in.Taints[i] = manualv1alpha1.Taint{
				Effect: string(t.Effect),
				Key:    t.Key,
				Value:  t.Value,
			}
		}
	}
}

// IsNodeGroupUpToDate checks whether there is a change in any of the modifiable fields.
func IsNodeGroupUpToDate(p *manualv1alpha1.NodeGroupParameters, ng *ekstypes.Nodegroup) bool { //nolint:gocyclo
	if !cmp.Equal(p.Tags, ng.Tags, cmpopts.EquateEmpty()) {
		return false
	}
	if !notNilAndEquals(p.AMIType, "CUSTOM") {
		if !cmp.Equal(p.Version, ng.Version) {
			return false
		}
		if !cmp.Equal(p.ReleaseVersion, ng.ReleaseVersion) {
			return false
		}
	}
	if !cmp.Equal(p.Labels, ng.Labels, cmpopts.EquateEmpty()) {
		return false
	}
	if p.LaunchTemplate != nil && ng.LaunchTemplate != nil {
		if !cmp.Equal(p.LaunchTemplate.Version, ng.LaunchTemplate.Version) {
			return false
		}
	}
	if p.UpdateConfig != nil && ng.UpdateConfig != nil {
		if !cmp.Equal(p.UpdateConfig.MaxUnavailable, ng.UpdateConfig.MaxUnavailable) {
			return false
		}
		if !cmp.Equal(p.UpdateConfig.MaxUnavailablePercentage, ng.UpdateConfig.MaxUnavailablePercentage) {
			return false
		}
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
	taints := generateUpdateTaintsPayload(p.Taints, ng.Taints)
	if len(taints.AddOrUpdateTaints) > 0 || len(taints.RemoveTaints) > 0 {
		return false
	}
	return false
}

func notNilAndEquals(p *string, s string) bool {
	return p != nil && *p == s
}
