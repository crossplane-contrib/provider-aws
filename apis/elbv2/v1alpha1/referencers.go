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
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	acm "github.com/crossplane-contrib/provider-aws/apis/acm/v1beta1"
	ec2 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences resolves references for Listeners
func (mg *Listener) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve certificate ARN reference
	for i := range mg.Spec.ForProvider.Certificates {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Certificates[i].CertificateARN),
			Reference:    mg.Spec.ForProvider.Certificates[i].CertificateARNRef,
			Selector:     mg.Spec.ForProvider.Certificates[i].CertificateARNSelector,
			To:           reference.To{Managed: &acm.Certificate{}, List: &acm.CertificateList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.forProvider.certificateArn")
		}
		mg.Spec.ForProvider.Certificates[i].CertificateARN = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.Certificates[i].CertificateARNRef = rsp.ResolvedReference
	}

	// resolve loadbalancer ARN reference
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.LoadBalancerARN),
		Reference:    mg.Spec.ForProvider.LoadBalancerARNRef,
		Selector:     mg.Spec.ForProvider.LoadBalancerARNSelector,
		To:           reference.To{Managed: &LoadBalancer{}, List: &LoadBalancerList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.loadBalancerArn")
	}
	mg.Spec.ForProvider.LoadBalancerARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.LoadBalancerARNRef = rsp.ResolvedReference

	for i, a := range mg.Spec.ForProvider.DefaultActions {
		// resolve single target group ARN references for each default action
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(a.TargetGroupARN),
			Reference:    a.TargetGroupARNRef,
			Selector:     a.TargetGroupARNSelector,
			To:           reference.To{Managed: &TargetGroup{}, List: &TargetGroupList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("spec.forProvider.DefaultActions[%d].targetGroupArn", i))
		}

		a.TargetGroupARN = reference.ToPtrValue(rsp.ResolvedValue)
		a.TargetGroupARNRef = rsp.ResolvedReference

		// resolve target group ARN references in forwardconfig if there are any
		if a.ForwardConfig != nil {
			for j, tg := range a.ForwardConfig.TargetGroups {
				rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
					CurrentValue: reference.FromPtrValue(tg.TargetGroupARN),
					Reference:    tg.TargetGroupARNRef,
					Selector:     tg.TargetGroupARNSelector,
					To:           reference.To{Managed: &TargetGroup{}, List: &TargetGroupList{}},
					Extract:      reference.ExternalName(),
				})
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("spec.forProvider.DefaultActions[%d].forwardConfig.targetGroups[%d]", i, j))
				}

				tg.TargetGroupARN = reference.ToPtrValue(rsp.ResolvedValue)
				tg.TargetGroupARNRef = rsp.ResolvedReference
			}
		}
	}

	return nil
}

// ResolveReferences resolves references for LoadBalancers
func (mg *LoadBalancer) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve subnets references
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.Subnets),
		References:    mg.Spec.ForProvider.SubnetRefs,
		Selector:      mg.Spec.ForProvider.SubnetSelector,
		To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.ForProvider.subnets")
	}

	mg.Spec.ForProvider.Subnets = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.SubnetRefs = mrsp.ResolvedReferences

	// resolve SecurityGroups references
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.SecurityGroups),
		References:    mg.Spec.ForProvider.SecurityGroupRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.ForProvider.SecurityGroups")
	}

	mg.Spec.ForProvider.SecurityGroups = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.SecurityGroupRefs = mrsp.ResolvedReferences

	return nil
}

// ResolveReferences resolves references for TargetGroups
func (mg *TargetGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// resolve vpc ID reference
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.VPCID),
		Reference:    mg.Spec.ForProvider.VPCIDRef,
		Selector:     mg.Spec.ForProvider.VPCIDSelector,
		To:           reference.To{Managed: &ec2.VPC{}, List: &ec2.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.loadBalancerArn")
	}
	mg.Spec.ForProvider.VPCID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.VPCIDRef = rsp.ResolvedReference
	return nil
}
