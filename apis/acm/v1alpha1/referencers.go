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

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	acmpcav1alpha1 "github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
)

// ResolveReferences of this Certificate
func (mg *Certificate) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.CertificateAuthorityARN
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CertificateAuthorityARN),
		Reference:    mg.Spec.ForProvider.CertificateAuthorityARNRef,
		Selector:     mg.Spec.ForProvider.CertificateAuthorityARNSelector,
		To:           reference.To{Managed: &acmpcav1alpha1.CertificateAuthority{}, List: &acmpcav1alpha1.CertificateAuthorityList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.CertificateAuthorityARN")
	}
	mg.Spec.ForProvider.CertificateAuthorityARN = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.CertificateAuthorityARNRef = rsp.ResolvedReference

	return nil
}
