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

// Code generated by ack-generate. DO NOT EDIT.

package vpclink

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.
// TODO(muvaf): We can generate one-time boilerplate for these hooks but currently
// ACK doesn't support not generating if file exists.

// GenerateGetVpcLinkInput returns input for read
// operation.
func GenerateGetVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.GetVpcLinkInput {
	res := preGenerateGetVpcLinkInput(cr, &svcsdk.GetVpcLinkInput{})

	if cr.Status.AtProvider.VPCLinkID != nil {
		res.SetVpcLinkId(*cr.Status.AtProvider.VPCLinkID)
	}

	return postGenerateGetVpcLinkInput(cr, res)
}

// GenerateVPCLink returns the current state in the form of *svcapitypes.VPCLink.
func GenerateVPCLink(resp *svcsdk.GetVpcLinkOutput) *svcapitypes.VPCLink {
	cr := &svcapitypes.VPCLink{}

	if resp.CreatedDate != nil {
		cr.Status.AtProvider.CreatedDate = &metav1.Time{*resp.CreatedDate}
	}
	if resp.Name != nil {
		cr.Status.AtProvider.Name = resp.Name
	}
	if resp.SecurityGroupIds != nil {
		f2 := []*string{}
		for _, f2iter := range resp.SecurityGroupIds {
			var f2elem string
			f2elem = *f2iter
			f2 = append(f2, &f2elem)
		}
		cr.Status.AtProvider.SecurityGroupIDs = f2
	}
	if resp.SubnetIds != nil {
		f3 := []*string{}
		for _, f3iter := range resp.SubnetIds {
			var f3elem string
			f3elem = *f3iter
			f3 = append(f3, &f3elem)
		}
		cr.Status.AtProvider.SubnetIDs = f3
	}
	if resp.VpcLinkId != nil {
		cr.Status.AtProvider.VPCLinkID = resp.VpcLinkId
	}
	if resp.VpcLinkStatus != nil {
		cr.Status.AtProvider.VPCLinkStatus = resp.VpcLinkStatus
	}
	if resp.VpcLinkStatusMessage != nil {
		cr.Status.AtProvider.VPCLinkStatusMessage = resp.VpcLinkStatusMessage
	}
	if resp.VpcLinkVersion != nil {
		cr.Status.AtProvider.VPCLinkVersion = resp.VpcLinkVersion
	}

	return cr
}

// GenerateCreateVpcLinkInput returns a create input.
func GenerateCreateVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.CreateVpcLinkInput {
	res := preGenerateCreateVpcLinkInput(cr, &svcsdk.CreateVpcLinkInput{})

	if cr.Spec.ForProvider.Tags != nil {
		f0 := map[string]*string{}
		for f0key, f0valiter := range cr.Spec.ForProvider.Tags {
			var f0val string
			f0val = *f0valiter
			f0[f0key] = &f0val
		}
		res.SetTags(f0)
	}

	return postGenerateCreateVpcLinkInput(cr, res)
}

// GenerateDeleteVpcLinkInput returns a deletion input.
func GenerateDeleteVpcLinkInput(cr *svcapitypes.VPCLink) *svcsdk.DeleteVpcLinkInput {
	res := preGenerateDeleteVpcLinkInput(cr, &svcsdk.DeleteVpcLinkInput{})

	if cr.Status.AtProvider.VPCLinkID != nil {
		res.SetVpcLinkId(*cr.Status.AtProvider.VPCLinkID)
	}

	return postGenerateDeleteVpcLinkInput(cr, res)
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}