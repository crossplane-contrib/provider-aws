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

package v1alpha1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"

	network "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	iamv1beta1 "github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	kms "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// ResolveReferences of this Job
func (mg *Job) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.roleArn
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.RoleArn,
		Reference:    mg.Spec.ForProvider.RoleArnRef,
		Selector:     mg.Spec.ForProvider.RoleArnSelector,
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      iamv1beta1.RoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.roleArn")
	}
	mg.Spec.ForProvider.RoleArn = rsp.ResolvedValue
	mg.Spec.ForProvider.RoleArnRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Crawler
func (mg *Crawler) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.roleArn
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.RoleArn,
		Reference:    mg.Spec.ForProvider.RoleArnRef,
		Selector:     mg.Spec.ForProvider.RoleArnSelector,
		To:           reference.To{Managed: &iamv1beta1.Role{}, List: &iamv1beta1.RoleList{}},
		Extract:      iamv1beta1.RoleARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.roleArn")
	}
	mg.Spec.ForProvider.RoleArn = rsp.ResolvedValue
	mg.Spec.ForProvider.RoleArnRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Connection
func (mg *Connection) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.connectionInput.physicalConnectionRequirements.securityGroupIDList
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList,
		References:    mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDSelector,
		To:            reference.To{Managed: &network.SecurityGroup{}, List: &network.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.connectionInput.physicalConnectionRequirements.securityGroupIDList")
	}
	mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList = mrsp.ResolvedValues
	mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.connectionInput.physicalConnectionRequirements.subnetID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetID),
		Reference:    mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetIDRef,
		Selector:     mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetIDSelector,
		To:           reference.To{Managed: &network.Subnet{}, List: &network.SubnetList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.connectionInput.physicalConnectionRequirements.subnetID")
	}
	mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this SecurityConfiguration
func (mg *SecurityConfiguration) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.encryptionConfiguration.cloudWatchEncryption.kmsKeyARN
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARN),
		Reference:    mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARNRef,
		Selector:     mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARNSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      KMSKeyARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.encryptionConfiguration.cloudWatchEncryption.kmsKeyARN")
	}
	mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARNRef = rsp.ResolvedReference

	// Resolve spec.forProvider.encryptionConfiguration.jobBookmarksEncryption.kmsKeyARN
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARN),
		Reference:    mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARNRef,
		Selector:     mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARNSelector,
		To:           reference.To{Managed: &kms.Key{}, List: &kms.KeyList{}},
		Extract:      KMSKeyARN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.encryptionConfiguration.jobBookmarksEncryption.kmsKeyARN")
	}
	mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARNRef = rsp.ResolvedReference

	return nil
}

// KMSKeyARN returns the status.atProvider.ARN of an KMSKey.
func KMSKeyARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*kms.Key)
		if !ok {
			return ""
		}
		if r.Status.AtProvider.ARN == nil {
			return ""
		}
		return *r.Status.AtProvider.ARN
	}
}
