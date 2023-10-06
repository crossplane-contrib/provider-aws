/*
Copyright 2019 The Crossplane Authors.

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

package v1beta1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	kms "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	s3v1beta1 "github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
)

// FunctionARN returns the status.atProvider.ARN of a Function.
func FunctionARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Function)
		if !ok || r.Status.AtProvider.FunctionARN == nil {
			return ""
		}
		return *r.Status.AtProvider.FunctionARN
	}
}

// ResolveReferences of this Function
func (mg *Function) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.role
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Role),
		Reference:    mg.Spec.ForProvider.RoleRef,
		Selector:     mg.Spec.ForProvider.RoleSelector,
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      iamv1beta1.RoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.role")
	}
	mg.Spec.ForProvider.Role = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RoleRef = rsp.ResolvedReference

	// Resolve spec.forProvider.VPCConfig
	if mg.Spec.ForProvider.CustomFunctionVPCConfigParameters != nil {
		// Resolve spec.forProvider.VPCConfig.subnetIds
		mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
			CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.CustomFunctionVPCConfigParameters.SubnetIDs),
			References:    mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SubnetIDRefs,
			Selector:      mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SubnetIDSelector,
			To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
			Extract:       reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.vpcConfig.subnetIds")
		}
		mg.Spec.ForProvider.CustomFunctionVPCConfigParameters.SubnetIDs = reference.ToPtrValues(mrsp.ResolvedValues)
		mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SubnetIDRefs = mrsp.ResolvedReferences

		// Resolve spec.forProvider.VPCConfig.securityGroupIds
		mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
			CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs),
			References:    mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SecurityGroupIDRefs,
			Selector:      mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SecurityGroupIDSelector,
			To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
			Extract:       reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.vpcConfig.securityGroupIds")
		}
		mg.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs = reference.ToPtrValues(mrsp.ResolvedValues)
		mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionVPCConfigParameters.SecurityGroupIDRefs = mrsp.ResolvedReferences
	}

	// Resolve spec.forProvider.code.s3Bucket
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomFunctionCodeParameters.S3Bucket),
		Reference:    mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionCodeParameters.S3BucketRef,
		Selector:     mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionCodeParameters.S3BucketSelector,
		To:           reference.To{Managed: &s3v1beta1.Bucket{}, List: &s3v1beta1.BucketList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.code.s3Bucket")
	}
	mg.Spec.ForProvider.CustomFunctionCodeParameters.S3Bucket = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomFunctionParameters.CustomFunctionCodeParameters.S3BucketRef = rsp.ResolvedReference

	// Resolve spec.forProvider.kmsKeyARN
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KMSKeyARN),
		Reference:    mg.Spec.ForProvider.KMSKeyARNRef,
		Selector:     mg.Spec.ForProvider.KMSKeyARNSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.kmsKeyARN")
	}
	mg.Spec.ForProvider.KMSKeyARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.KMSKeyARNRef = rsp.ResolvedReference

	return nil
}
