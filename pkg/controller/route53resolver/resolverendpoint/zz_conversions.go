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

// Code generated by ack-generate. DO NOT EDIT.

package resolverendpoint

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/route53resolver"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetResolverEndpointInput returns input for read
// operation.
func GenerateGetResolverEndpointInput(cr *svcapitypes.ResolverEndpoint) *svcsdk.GetResolverEndpointInput {
	res := &svcsdk.GetResolverEndpointInput{}

	return res
}

// GenerateResolverEndpoint returns the current state in the form of *svcapitypes.ResolverEndpoint.
func GenerateResolverEndpoint(resp *svcsdk.GetResolverEndpointOutput) *svcapitypes.ResolverEndpoint {
	cr := &svcapitypes.ResolverEndpoint{}

	if resp.ResolverEndpoint.Arn != nil {
		cr.Status.AtProvider.ARN = resp.ResolverEndpoint.Arn
	} else {
		cr.Status.AtProvider.ARN = nil
	}
	if resp.ResolverEndpoint.CreationTime != nil {
		cr.Status.AtProvider.CreationTime = resp.ResolverEndpoint.CreationTime
	} else {
		cr.Status.AtProvider.CreationTime = nil
	}
	if resp.ResolverEndpoint.CreatorRequestId != nil {
		cr.Status.AtProvider.CreatorRequestID = resp.ResolverEndpoint.CreatorRequestId
	} else {
		cr.Status.AtProvider.CreatorRequestID = nil
	}
	if resp.ResolverEndpoint.Direction != nil {
		cr.Spec.ForProvider.Direction = resp.ResolverEndpoint.Direction
	} else {
		cr.Spec.ForProvider.Direction = nil
	}
	if resp.ResolverEndpoint.HostVPCId != nil {
		cr.Status.AtProvider.HostVPCID = resp.ResolverEndpoint.HostVPCId
	} else {
		cr.Status.AtProvider.HostVPCID = nil
	}
	if resp.ResolverEndpoint.Id != nil {
		cr.Status.AtProvider.ID = resp.ResolverEndpoint.Id
	} else {
		cr.Status.AtProvider.ID = nil
	}
	if resp.ResolverEndpoint.IpAddressCount != nil {
		cr.Status.AtProvider.IPAddressCount = resp.ResolverEndpoint.IpAddressCount
	} else {
		cr.Status.AtProvider.IPAddressCount = nil
	}
	if resp.ResolverEndpoint.ModificationTime != nil {
		cr.Status.AtProvider.ModificationTime = resp.ResolverEndpoint.ModificationTime
	} else {
		cr.Status.AtProvider.ModificationTime = nil
	}
	if resp.ResolverEndpoint.Name != nil {
		cr.Spec.ForProvider.Name = resp.ResolverEndpoint.Name
	} else {
		cr.Spec.ForProvider.Name = nil
	}
	if resp.ResolverEndpoint.ResolverEndpointType != nil {
		cr.Spec.ForProvider.ResolverEndpointType = resp.ResolverEndpoint.ResolverEndpointType
	} else {
		cr.Spec.ForProvider.ResolverEndpointType = nil
	}
	if resp.ResolverEndpoint.SecurityGroupIds != nil {
		f10 := []*string{}
		for _, f10iter := range resp.ResolverEndpoint.SecurityGroupIds {
			var f10elem string
			f10elem = *f10iter
			f10 = append(f10, &f10elem)
		}
		cr.Status.AtProvider.SecurityGroupIDs = f10
	} else {
		cr.Status.AtProvider.SecurityGroupIDs = nil
	}
	if resp.ResolverEndpoint.Status != nil {
		cr.Status.AtProvider.Status = resp.ResolverEndpoint.Status
	} else {
		cr.Status.AtProvider.Status = nil
	}
	if resp.ResolverEndpoint.StatusMessage != nil {
		cr.Status.AtProvider.StatusMessage = resp.ResolverEndpoint.StatusMessage
	} else {
		cr.Status.AtProvider.StatusMessage = nil
	}

	return cr
}

// GenerateCreateResolverEndpointInput returns a create input.
func GenerateCreateResolverEndpointInput(cr *svcapitypes.ResolverEndpoint) *svcsdk.CreateResolverEndpointInput {
	res := &svcsdk.CreateResolverEndpointInput{}

	if cr.Spec.ForProvider.Direction != nil {
		res.SetDirection(*cr.Spec.ForProvider.Direction)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.ResolverEndpointType != nil {
		res.SetResolverEndpointType(*cr.Spec.ForProvider.ResolverEndpointType)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f3 := []*svcsdk.Tag{}
		for _, f3iter := range cr.Spec.ForProvider.Tags {
			f3elem := &svcsdk.Tag{}
			if f3iter.Key != nil {
				f3elem.SetKey(*f3iter.Key)
			}
			if f3iter.Value != nil {
				f3elem.SetValue(*f3iter.Value)
			}
			f3 = append(f3, f3elem)
		}
		res.SetTags(f3)
	}

	return res
}

// GenerateUpdateResolverEndpointInput returns an update input.
func GenerateUpdateResolverEndpointInput(cr *svcapitypes.ResolverEndpoint) *svcsdk.UpdateResolverEndpointInput {
	res := &svcsdk.UpdateResolverEndpointInput{}

	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.ResolverEndpointType != nil {
		res.SetResolverEndpointType(*cr.Spec.ForProvider.ResolverEndpointType)
	}

	return res
}

// GenerateDeleteResolverEndpointInput returns a deletion input.
func GenerateDeleteResolverEndpointInput(cr *svcapitypes.ResolverEndpoint) *svcsdk.DeleteResolverEndpointInput {
	res := &svcsdk.DeleteResolverEndpointInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "ResourceNotFoundException"
}
