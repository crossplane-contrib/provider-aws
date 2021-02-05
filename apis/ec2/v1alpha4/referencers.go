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

package v1alpha4

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	ec2v1alpha1 "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	ec2v1beta1 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// ResolveReferences of this RouteTable
func (mg *RouteTable) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.vpcId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: aws.StringValue(mg.Spec.ForProvider.VPCID),
		Reference:    mg.Spec.ForProvider.VPCIDRef,
		Selector:     mg.Spec.ForProvider.VPCIDSelector,
		To:           reference.To{Managed: &ec2v1beta1.VPC{}, List: &ec2v1beta1.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.vpcId")
	}
	mg.Spec.ForProvider.VPCID = aws.String(rsp.ResolvedValue)
	mg.Spec.ForProvider.VPCIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.routes[].gatewayId
	for i := range mg.Spec.ForProvider.Routes {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: aws.StringValue(mg.Spec.ForProvider.Routes[i].GatewayID),
			Reference:    mg.Spec.ForProvider.Routes[i].GatewayIDRef,
			Selector:     mg.Spec.ForProvider.Routes[i].GatewayIDSelector,
			To:           reference.To{Managed: &ec2v1beta1.InternetGateway{}, List: &ec2v1beta1.InternetGatewayList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrapf(err, "spec.forProvider.routes[%d].gatewayId", i)
		}
		mg.Spec.ForProvider.Routes[i].GatewayID = aws.String(rsp.ResolvedValue)
		mg.Spec.ForProvider.Routes[i].GatewayIDRef = rsp.ResolvedReference
	}

	// Resolve spec.forProvider.routes[].natGatewayId
	for i := range mg.Spec.ForProvider.Routes {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: aws.StringValue(mg.Spec.ForProvider.Routes[i].NatGatewayID),
			Reference:    mg.Spec.ForProvider.Routes[i].NatGatewayIDRef,
			Selector:     mg.Spec.ForProvider.Routes[i].NatGatewayIDSelector,
			To:           reference.To{Managed: &ec2v1alpha1.NATGateway{}, List: &ec2v1alpha1.NATGatewayList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrapf(err, "spec.forProvider.routes[%d].natGatewayId", i)
		}
		mg.Spec.ForProvider.Routes[i].NatGatewayID = aws.String(rsp.ResolvedValue)
		mg.Spec.ForProvider.Routes[i].NatGatewayIDRef = rsp.ResolvedReference
	}

	// Resolve spec.associations[].subnetId
	for i := range mg.Spec.ForProvider.Associations {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: aws.StringValue(mg.Spec.ForProvider.Associations[i].SubnetID),
			Reference:    mg.Spec.ForProvider.Associations[i].SubnetIDRef,
			Selector:     mg.Spec.ForProvider.Associations[i].SubnetIDSelector,
			To:           reference.To{Managed: &ec2v1beta1.Subnet{}, List: &ec2v1beta1.SubnetList{}},
			Extract:      reference.ExternalName(),
		})
		if err != nil {
			return errors.Wrapf(err, "spec.forProvider.associations[%d].subnetId", i)
		}
		mg.Spec.ForProvider.Associations[i].SubnetID = aws.String(rsp.ResolvedValue)
		mg.Spec.ForProvider.Associations[i].SubnetIDRef = rsp.ResolvedReference
	}

	return nil
}
