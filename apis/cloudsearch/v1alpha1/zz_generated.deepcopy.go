//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainObservation) DeepCopyInto(out *CustomDomainObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainObservation.
func (in *CustomDomainObservation) DeepCopy() *CustomDomainObservation {
	if in == nil {
		return nil
	}
	out := new(CustomDomainObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomDomainParameters) DeepCopyInto(out *CustomDomainParameters) {
	*out = *in
	if in.DesiredReplicationCount != nil {
		in, out := &in.DesiredReplicationCount, &out.DesiredReplicationCount
		*out = new(int64)
		**out = **in
	}
	if in.DesiredInstanceType != nil {
		in, out := &in.DesiredInstanceType, &out.DesiredInstanceType
		*out = new(string)
		**out = **in
	}
	if in.DesiredPartitionCount != nil {
		in, out := &in.DesiredPartitionCount, &out.DesiredPartitionCount
		*out = new(int64)
		**out = **in
	}
	if in.AccessPolicies != nil {
		in, out := &in.AccessPolicies, &out.AccessPolicies
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomDomainParameters.
func (in *CustomDomainParameters) DeepCopy() *CustomDomainParameters {
	if in == nil {
		return nil
	}
	out := new(CustomDomainParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DateArrayOptions) DeepCopyInto(out *DateArrayOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DateArrayOptions.
func (in *DateArrayOptions) DeepCopy() *DateArrayOptions {
	if in == nil {
		return nil
	}
	out := new(DateArrayOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DateOptions) DeepCopyInto(out *DateOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DateOptions.
func (in *DateOptions) DeepCopy() *DateOptions {
	if in == nil {
		return nil
	}
	out := new(DateOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Domain) DeepCopyInto(out *Domain) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Domain.
func (in *Domain) DeepCopy() *Domain {
	if in == nil {
		return nil
	}
	out := new(Domain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Domain) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainEndpointOptions) DeepCopyInto(out *DomainEndpointOptions) {
	*out = *in
	if in.EnforceHTTPS != nil {
		in, out := &in.EnforceHTTPS, &out.EnforceHTTPS
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainEndpointOptions.
func (in *DomainEndpointOptions) DeepCopy() *DomainEndpointOptions {
	if in == nil {
		return nil
	}
	out := new(DomainEndpointOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainList) DeepCopyInto(out *DomainList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Domain, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainList.
func (in *DomainList) DeepCopy() *DomainList {
	if in == nil {
		return nil
	}
	out := new(DomainList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DomainList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainObservation) DeepCopyInto(out *DomainObservation) {
	*out = *in
	if in.ARN != nil {
		in, out := &in.ARN, &out.ARN
		*out = new(string)
		**out = **in
	}
	if in.Created != nil {
		in, out := &in.Created, &out.Created
		*out = new(bool)
		**out = **in
	}
	if in.Deleted != nil {
		in, out := &in.Deleted, &out.Deleted
		*out = new(bool)
		**out = **in
	}
	if in.DocService != nil {
		in, out := &in.DocService, &out.DocService
		*out = new(ServiceEndpoint)
		(*in).DeepCopyInto(*out)
	}
	if in.DomainID != nil {
		in, out := &in.DomainID, &out.DomainID
		*out = new(string)
		**out = **in
	}
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = new(Limits)
		(*in).DeepCopyInto(*out)
	}
	if in.Processing != nil {
		in, out := &in.Processing, &out.Processing
		*out = new(bool)
		**out = **in
	}
	if in.RequiresIndexDocuments != nil {
		in, out := &in.RequiresIndexDocuments, &out.RequiresIndexDocuments
		*out = new(bool)
		**out = **in
	}
	if in.SearchInstanceCount != nil {
		in, out := &in.SearchInstanceCount, &out.SearchInstanceCount
		*out = new(int64)
		**out = **in
	}
	if in.SearchInstanceType != nil {
		in, out := &in.SearchInstanceType, &out.SearchInstanceType
		*out = new(string)
		**out = **in
	}
	if in.SearchPartitionCount != nil {
		in, out := &in.SearchPartitionCount, &out.SearchPartitionCount
		*out = new(int64)
		**out = **in
	}
	if in.SearchService != nil {
		in, out := &in.SearchService, &out.SearchService
		*out = new(ServiceEndpoint)
		(*in).DeepCopyInto(*out)
	}
	out.CustomDomainObservation = in.CustomDomainObservation
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainObservation.
func (in *DomainObservation) DeepCopy() *DomainObservation {
	if in == nil {
		return nil
	}
	out := new(DomainObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainParameters) DeepCopyInto(out *DomainParameters) {
	*out = *in
	if in.DomainName != nil {
		in, out := &in.DomainName, &out.DomainName
		*out = new(string)
		**out = **in
	}
	in.CustomDomainParameters.DeepCopyInto(&out.CustomDomainParameters)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainParameters.
func (in *DomainParameters) DeepCopy() *DomainParameters {
	if in == nil {
		return nil
	}
	out := new(DomainParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainSpec) DeepCopyInto(out *DomainSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainSpec.
func (in *DomainSpec) DeepCopy() *DomainSpec {
	if in == nil {
		return nil
	}
	out := new(DomainSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainStatus) DeepCopyInto(out *DomainStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainStatus.
func (in *DomainStatus) DeepCopy() *DomainStatus {
	if in == nil {
		return nil
	}
	out := new(DomainStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainStatus_SDK) DeepCopyInto(out *DomainStatus_SDK) {
	*out = *in
	if in.ARN != nil {
		in, out := &in.ARN, &out.ARN
		*out = new(string)
		**out = **in
	}
	if in.Created != nil {
		in, out := &in.Created, &out.Created
		*out = new(bool)
		**out = **in
	}
	if in.Deleted != nil {
		in, out := &in.Deleted, &out.Deleted
		*out = new(bool)
		**out = **in
	}
	if in.DocService != nil {
		in, out := &in.DocService, &out.DocService
		*out = new(ServiceEndpoint)
		(*in).DeepCopyInto(*out)
	}
	if in.DomainID != nil {
		in, out := &in.DomainID, &out.DomainID
		*out = new(string)
		**out = **in
	}
	if in.DomainName != nil {
		in, out := &in.DomainName, &out.DomainName
		*out = new(string)
		**out = **in
	}
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = new(Limits)
		(*in).DeepCopyInto(*out)
	}
	if in.Processing != nil {
		in, out := &in.Processing, &out.Processing
		*out = new(bool)
		**out = **in
	}
	if in.RequiresIndexDocuments != nil {
		in, out := &in.RequiresIndexDocuments, &out.RequiresIndexDocuments
		*out = new(bool)
		**out = **in
	}
	if in.SearchInstanceCount != nil {
		in, out := &in.SearchInstanceCount, &out.SearchInstanceCount
		*out = new(int64)
		**out = **in
	}
	if in.SearchInstanceType != nil {
		in, out := &in.SearchInstanceType, &out.SearchInstanceType
		*out = new(string)
		**out = **in
	}
	if in.SearchPartitionCount != nil {
		in, out := &in.SearchPartitionCount, &out.SearchPartitionCount
		*out = new(int64)
		**out = **in
	}
	if in.SearchService != nil {
		in, out := &in.SearchService, &out.SearchService
		*out = new(ServiceEndpoint)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainStatus_SDK.
func (in *DomainStatus_SDK) DeepCopy() *DomainStatus_SDK {
	if in == nil {
		return nil
	}
	out := new(DomainStatus_SDK)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DoubleArrayOptions) DeepCopyInto(out *DoubleArrayOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DoubleArrayOptions.
func (in *DoubleArrayOptions) DeepCopy() *DoubleArrayOptions {
	if in == nil {
		return nil
	}
	out := new(DoubleArrayOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DoubleOptions) DeepCopyInto(out *DoubleOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DoubleOptions.
func (in *DoubleOptions) DeepCopy() *DoubleOptions {
	if in == nil {
		return nil
	}
	out := new(DoubleOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntArrayOptions) DeepCopyInto(out *IntArrayOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntArrayOptions.
func (in *IntArrayOptions) DeepCopy() *IntArrayOptions {
	if in == nil {
		return nil
	}
	out := new(IntArrayOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntOptions) DeepCopyInto(out *IntOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntOptions.
func (in *IntOptions) DeepCopy() *IntOptions {
	if in == nil {
		return nil
	}
	out := new(IntOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LatLonOptions) DeepCopyInto(out *LatLonOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LatLonOptions.
func (in *LatLonOptions) DeepCopy() *LatLonOptions {
	if in == nil {
		return nil
	}
	out := new(LatLonOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Limits) DeepCopyInto(out *Limits) {
	*out = *in
	if in.MaximumPartitionCount != nil {
		in, out := &in.MaximumPartitionCount, &out.MaximumPartitionCount
		*out = new(int64)
		**out = **in
	}
	if in.MaximumReplicationCount != nil {
		in, out := &in.MaximumReplicationCount, &out.MaximumReplicationCount
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Limits.
func (in *Limits) DeepCopy() *Limits {
	if in == nil {
		return nil
	}
	out := new(Limits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LiteralArrayOptions) DeepCopyInto(out *LiteralArrayOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LiteralArrayOptions.
func (in *LiteralArrayOptions) DeepCopy() *LiteralArrayOptions {
	if in == nil {
		return nil
	}
	out := new(LiteralArrayOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LiteralOptions) DeepCopyInto(out *LiteralOptions) {
	*out = *in
	if in.FacetEnabled != nil {
		in, out := &in.FacetEnabled, &out.FacetEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SearchEnabled != nil {
		in, out := &in.SearchEnabled, &out.SearchEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LiteralOptions.
func (in *LiteralOptions) DeepCopy() *LiteralOptions {
	if in == nil {
		return nil
	}
	out := new(LiteralOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OptionStatus) DeepCopyInto(out *OptionStatus) {
	*out = *in
	if in.PendingDeletion != nil {
		in, out := &in.PendingDeletion, &out.PendingDeletion
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OptionStatus.
func (in *OptionStatus) DeepCopy() *OptionStatus {
	if in == nil {
		return nil
	}
	out := new(OptionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceEndpoint) DeepCopyInto(out *ServiceEndpoint) {
	*out = *in
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceEndpoint.
func (in *ServiceEndpoint) DeepCopy() *ServiceEndpoint {
	if in == nil {
		return nil
	}
	out := new(ServiceEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TextArrayOptions) DeepCopyInto(out *TextArrayOptions) {
	*out = *in
	if in.HighlightEnabled != nil {
		in, out := &in.HighlightEnabled, &out.HighlightEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TextArrayOptions.
func (in *TextArrayOptions) DeepCopy() *TextArrayOptions {
	if in == nil {
		return nil
	}
	out := new(TextArrayOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TextOptions) DeepCopyInto(out *TextOptions) {
	*out = *in
	if in.HighlightEnabled != nil {
		in, out := &in.HighlightEnabled, &out.HighlightEnabled
		*out = new(bool)
		**out = **in
	}
	if in.ReturnEnabled != nil {
		in, out := &in.ReturnEnabled, &out.ReturnEnabled
		*out = new(bool)
		**out = **in
	}
	if in.SortEnabled != nil {
		in, out := &in.SortEnabled, &out.SortEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TextOptions.
func (in *TextOptions) DeepCopy() *TextOptions {
	if in == nil {
		return nil
	}
	out := new(TextOptions)
	in.DeepCopyInto(out)
	return out
}
