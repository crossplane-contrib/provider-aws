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

package manualv1alpha1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eksv1beta1 "github.com/crossplane/provider-aws/apis/eks/v1beta1"
)

// ResolveReferences of this IdentityProviderConfig
func (mg *IdentityProviderConfig) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.clusterName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ClusterName,
		Reference:    mg.Spec.ForProvider.ClusterNameRef,
		Selector:     mg.Spec.ForProvider.ClusterNameSelector,
		To:           reference.To{Managed: &eksv1beta1.Cluster{}, List: &eksv1beta1.ClusterList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.clusterName")
	}
	mg.Spec.ForProvider.ClusterName = rsp.ResolvedValue
	mg.Spec.ForProvider.ClusterNameRef = rsp.ResolvedReference

	return nil
}
