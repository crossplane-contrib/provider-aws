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

package v1alpha1

import (
	"context"

	ec2 "github.com/crossplane/provider-aws/apis/ec2/v1beta1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIID extracts the resolved API's ID.
func APIID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*API)
		if !ok {
			return ""
		}
		return reference.FromPtrValue(cr.Status.AtProvider.APIID)
	}
}

// AuthorizerID extracts the resolved Authorizer's ID.
func AuthorizerID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*Authorizer)
		if !ok {
			return ""
		}
		return reference.FromPtrValue(cr.Status.AtProvider.AuthorizerID)
	}
}

// ResolveReferences of this Stage
func (mg *Stage) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Route
func (mg *Route) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.authorizerId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.AuthorizerID),
		Reference:    mg.Spec.ForProvider.AuthorizerIDRef,
		Selector:     mg.Spec.ForProvider.AuthorizerIDSelector,
		To:           reference.To{Managed: &Authorizer{}, List: &AuthorizerList{}},
		Extract:      AuthorizerID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.authorizerId")
	}
	mg.Spec.ForProvider.AuthorizerID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.AuthorizerIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this RouteResponse
func (mg *RouteResponse) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.routeId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.RouteID),
		Reference:    mg.Spec.ForProvider.RouteIDRef,
		Selector:     mg.Spec.ForProvider.RouteIDSelector,
		To:           reference.To{Managed: &Route{}, List: &RouteList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.routeId")
	}
	mg.Spec.ForProvider.RouteID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.RouteIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Authorizer
func (mg *Authorizer) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Integration
func (mg *Integration) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this IntegrationResponse
func (mg *IntegrationResponse) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.integrationId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IntegrationID),
		Reference:    mg.Spec.ForProvider.IntegrationIDRef,
		Selector:     mg.Spec.ForProvider.IntegrationIDSelector,
		To:           reference.To{Managed: &Integration{}, List: &IntegrationList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.integrationId")
	}
	mg.Spec.ForProvider.IntegrationID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IntegrationIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Deployment
func (mg *Deployment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.stageName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.StageName),
		Reference:    mg.Spec.ForProvider.StageNameRef,
		Selector:     mg.Spec.ForProvider.StageNameSelector,
		To:           reference.To{Managed: &Stage{}, List: &StageList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.stageName")
	}
	mg.Spec.ForProvider.StageName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.StageNameRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Model
func (mg *Model) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this APIMapping
func (mg *APIMapping) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.apiId
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.APIID),
		Reference:    mg.Spec.ForProvider.APIIDRef,
		Selector:     mg.Spec.ForProvider.APIIDSelector,
		To:           reference.To{Managed: &API{}, List: &APIList{}},
		Extract:      APIID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.apiId")
	}
	mg.Spec.ForProvider.APIID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.APIIDRef = rsp.ResolvedReference

	// Resolve spec.forProvider.stage
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Stage),
		Reference:    mg.Spec.ForProvider.StageRef,
		Selector:     mg.Spec.ForProvider.StageSelector,
		To:           reference.To{Managed: &Stage{}, List: &StageList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.stage")
	}
	mg.Spec.ForProvider.Stage = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.StageRef = rsp.ResolvedReference

	// Resolve spec.forProvider.domainName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.DomainName),
		Reference:    mg.Spec.ForProvider.DomainNameRef,
		Selector:     mg.Spec.ForProvider.DomainNameSelector,
		To:           reference.To{Managed: &DomainName{}, List: &DomainNameList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.domainName")
	}
	mg.Spec.ForProvider.DomainName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.DomainNameRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this VPCLink
func (mg *VPCLink) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.subnetIds
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SubnetIDs,
		References:    mg.Spec.ForProvider.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.SubnetIDSelector,
		To:            reference.To{Managed: &ec2.Subnet{}, List: &ec2.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.subnetIds")
	}
	mg.Spec.ForProvider.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetIDRefs = mrsp.ResolvedReferences

	// Resolve spec.forProvider.securityGroupIds
	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SecurityGroupIDs,
		References:    mg.Spec.ForProvider.SecurityGroupIDRefs,
		Selector:      mg.Spec.ForProvider.SecurityGroupIDSelector,
		To:            reference.To{Managed: &ec2.SecurityGroup{}, List: &ec2.SecurityGroupList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.securityGroupIds")
	}
	mg.Spec.ForProvider.SecurityGroupIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SecurityGroupIDRefs = mrsp.ResolvedReferences
	return nil
}
