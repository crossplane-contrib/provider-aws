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

package v1alpha1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

// ResolveReferences of this CacheSubnetGroup
func (mg *CacheSubnetGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.subnetIDs
	mrsp, err := r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: mg.Spec.ForProvider.SubnetIDs,
		References:    mg.Spec.ForProvider.SubnetIDRefs,
		Selector:      mg.Spec.ForProvider.SubnetIDSelector,
		To:            reference.To{Managed: &v1alpha3.Subnet{}, List: &v1alpha3.SubnetList{}},
		Extract:       reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.SubnetIDs = mrsp.ResolvedValues
	mg.Spec.ForProvider.SubnetIDRefs = mrsp.ResolvedReferences

	return nil
}
