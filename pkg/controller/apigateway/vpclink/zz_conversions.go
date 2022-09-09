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

package vpclink

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetVpcLinkInput returns input for read
// operation.
func GenerateGetVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.GetVpcLinkInput {
	res := &svcsdk.GetVpcLinkInput{}

	return res
}

// GenerateVPCLink returns the current state in the form of *svcapitypes.VPCLink.
func GenerateVPCLink(resp *svcsdk.UpdateVpcLinkOutput) *svcapitypes.VPCLink {
	cr := &svcapitypes.VPCLink{}

	if resp.Description != nil {
		cr.Spec.ForProvider.Description = resp.Description
	} else {
		cr.Spec.ForProvider.Description = nil
	}
	if resp.Id != nil {
		cr.Status.AtProvider.ID = resp.Id
	} else {
		cr.Status.AtProvider.ID = nil
	}
	if resp.Name != nil {
		cr.Spec.ForProvider.Name = resp.Name
	} else {
		cr.Spec.ForProvider.Name = nil
	}
	if resp.Status != nil {
		cr.Status.AtProvider.Status = resp.Status
	} else {
		cr.Status.AtProvider.Status = nil
	}
	if resp.StatusMessage != nil {
		cr.Status.AtProvider.StatusMessage = resp.StatusMessage
	} else {
		cr.Status.AtProvider.StatusMessage = nil
	}
	if resp.Tags != nil {
		f5 := map[string]*string{}
		for f5key, f5valiter := range resp.Tags {
			var f5val string
			f5val = *f5valiter
			f5[f5key] = &f5val
		}
		cr.Spec.ForProvider.Tags = f5
	} else {
		cr.Spec.ForProvider.Tags = nil
	}
	if resp.TargetArns != nil {
		f6 := []*string{}
		for _, f6iter := range resp.TargetArns {
			var f6elem string
			f6elem = *f6iter
			f6 = append(f6, &f6elem)
		}
		cr.Spec.ForProvider.TargetARNs = f6
	} else {
		cr.Spec.ForProvider.TargetARNs = nil
	}

	return cr
}

// GenerateCreateVpcLinkInput returns a create input.
func GenerateCreateVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.CreateVpcLinkInput {
	res := &svcsdk.CreateVpcLinkInput{}

	if cr.Spec.ForProvider.Description != nil {
		res.SetDescription(*cr.Spec.ForProvider.Description)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f2 := map[string]*string{}
		for f2key, f2valiter := range cr.Spec.ForProvider.Tags {
			var f2val string
			f2val = *f2valiter
			f2[f2key] = &f2val
		}
		res.SetTags(f2)
	}
	if cr.Spec.ForProvider.TargetARNs != nil {
		f3 := []*string{}
		for _, f3iter := range cr.Spec.ForProvider.TargetARNs {
			var f3elem string
			f3elem = *f3iter
			f3 = append(f3, &f3elem)
		}
		res.SetTargetArns(f3)
	}

	return res
}

// GenerateUpdateVpcLinkInput returns an update input.
func GenerateUpdateVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.UpdateVpcLinkInput {
	res := &svcsdk.UpdateVpcLinkInput{}

	return res
}

// GenerateDeleteVpcLinkInput returns a deletion input.
func GenerateDeleteVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.DeleteVpcLinkInput {
	res := &svcsdk.DeleteVpcLinkInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
