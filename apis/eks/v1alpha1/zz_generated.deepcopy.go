// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoScalingGroup) DeepCopyInto(out *AutoScalingGroup) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoScalingGroup.
func (in *AutoScalingGroup) DeepCopy() *AutoScalingGroup {
	if in == nil {
		return nil
	}
	out := new(AutoScalingGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Issue) DeepCopyInto(out *Issue) {
	*out = *in
	if in.ResourceIDs != nil {
		in, out := &in.ResourceIDs, &out.ResourceIDs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Issue.
func (in *Issue) DeepCopy() *Issue {
	if in == nil {
		return nil
	}
	out := new(Issue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroup) DeepCopyInto(out *NodeGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroup.
func (in *NodeGroup) DeepCopy() *NodeGroup {
	if in == nil {
		return nil
	}
	out := new(NodeGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupHealth) DeepCopyInto(out *NodeGroupHealth) {
	*out = *in
	if in.Issues != nil {
		in, out := &in.Issues, &out.Issues
		*out = make([]Issue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupHealth.
func (in *NodeGroupHealth) DeepCopy() *NodeGroupHealth {
	if in == nil {
		return nil
	}
	out := new(NodeGroupHealth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupList) DeepCopyInto(out *NodeGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupList.
func (in *NodeGroupList) DeepCopy() *NodeGroupList {
	if in == nil {
		return nil
	}
	out := new(NodeGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupObservation) DeepCopyInto(out *NodeGroupObservation) {
	*out = *in
	if in.CreatedAt != nil {
		in, out := &in.CreatedAt, &out.CreatedAt
		*out = (*in).DeepCopy()
	}
	in.Health.DeepCopyInto(&out.Health)
	if in.ModifiedAt != nil {
		in, out := &in.ModifiedAt, &out.ModifiedAt
		*out = (*in).DeepCopy()
	}
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupObservation.
func (in *NodeGroupObservation) DeepCopy() *NodeGroupObservation {
	if in == nil {
		return nil
	}
	out := new(NodeGroupObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupParameters) DeepCopyInto(out *NodeGroupParameters) {
	*out = *in
	if in.AMIType != nil {
		in, out := &in.AMIType, &out.AMIType
		*out = new(string)
		**out = **in
	}
	if in.ClusterNameRef != nil {
		in, out := &in.ClusterNameRef, &out.ClusterNameRef
		*out = new(corev1alpha1.Reference)
		**out = **in
	}
	if in.ClusterNameSelector != nil {
		in, out := &in.ClusterNameSelector, &out.ClusterNameSelector
		*out = new(corev1alpha1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.DiskSize != nil {
		in, out := &in.DiskSize, &out.DiskSize
		*out = new(int64)
		**out = **in
	}
	if in.InstanceTypes != nil {
		in, out := &in.InstanceTypes, &out.InstanceTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.NodeRoleRef != nil {
		in, out := &in.NodeRoleRef, &out.NodeRoleRef
		*out = new(corev1alpha1.Reference)
		**out = **in
	}
	if in.NodeRoleSelector != nil {
		in, out := &in.NodeRoleSelector, &out.NodeRoleSelector
		*out = new(corev1alpha1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.ReleaseVersion != nil {
		in, out := &in.ReleaseVersion, &out.ReleaseVersion
		*out = new(string)
		**out = **in
	}
	if in.RemoteAccess != nil {
		in, out := &in.RemoteAccess, &out.RemoteAccess
		*out = new(RemoteAccessConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ScalingConfig != nil {
		in, out := &in.ScalingConfig, &out.ScalingConfig
		*out = new(NodeGroupScalingConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Subnets != nil {
		in, out := &in.Subnets, &out.Subnets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SubnetRefs != nil {
		in, out := &in.SubnetRefs, &out.SubnetRefs
		*out = make([]corev1alpha1.Reference, len(*in))
		copy(*out, *in)
	}
	if in.SubnetSelector != nil {
		in, out := &in.SubnetSelector, &out.SubnetSelector
		*out = new(corev1alpha1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupParameters.
func (in *NodeGroupParameters) DeepCopy() *NodeGroupParameters {
	if in == nil {
		return nil
	}
	out := new(NodeGroupParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupResources) DeepCopyInto(out *NodeGroupResources) {
	*out = *in
	if in.AutoScalingGroups != nil {
		in, out := &in.AutoScalingGroups, &out.AutoScalingGroups
		*out = make([]AutoScalingGroup, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupResources.
func (in *NodeGroupResources) DeepCopy() *NodeGroupResources {
	if in == nil {
		return nil
	}
	out := new(NodeGroupResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupScalingConfig) DeepCopyInto(out *NodeGroupScalingConfig) {
	*out = *in
	if in.DesiredSize != nil {
		in, out := &in.DesiredSize, &out.DesiredSize
		*out = new(int64)
		**out = **in
	}
	if in.MaxSize != nil {
		in, out := &in.MaxSize, &out.MaxSize
		*out = new(int64)
		**out = **in
	}
	if in.MinSize != nil {
		in, out := &in.MinSize, &out.MinSize
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupScalingConfig.
func (in *NodeGroupScalingConfig) DeepCopy() *NodeGroupScalingConfig {
	if in == nil {
		return nil
	}
	out := new(NodeGroupScalingConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupSpec) DeepCopyInto(out *NodeGroupSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupSpec.
func (in *NodeGroupSpec) DeepCopy() *NodeGroupSpec {
	if in == nil {
		return nil
	}
	out := new(NodeGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupStatus) DeepCopyInto(out *NodeGroupStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupStatus.
func (in *NodeGroupStatus) DeepCopy() *NodeGroupStatus {
	if in == nil {
		return nil
	}
	out := new(NodeGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteAccessConfig) DeepCopyInto(out *RemoteAccessConfig) {
	*out = *in
	if in.EC2SSHKey != nil {
		in, out := &in.EC2SSHKey, &out.EC2SSHKey
		*out = new(string)
		**out = **in
	}
	if in.SourceSecurityGroups != nil {
		in, out := &in.SourceSecurityGroups, &out.SourceSecurityGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SourceSecurityGroupRefs != nil {
		in, out := &in.SourceSecurityGroupRefs, &out.SourceSecurityGroupRefs
		*out = make([]corev1alpha1.Reference, len(*in))
		copy(*out, *in)
	}
	if in.SourceSecurityGroupSelector != nil {
		in, out := &in.SourceSecurityGroupSelector, &out.SourceSecurityGroupSelector
		*out = new(corev1alpha1.Selector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteAccessConfig.
func (in *RemoteAccessConfig) DeepCopy() *RemoteAccessConfig {
	if in == nil {
		return nil
	}
	out := new(RemoteAccessConfig)
	in.DeepCopyInto(out)
	return out
}
