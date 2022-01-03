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

package v1beta1

import (
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

// EKSOIDCIsser returns a function that returns the OIDC Issuer URL of the given EKS-controlplane.
func EKSOIDCIsser() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Cluster)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.Identity.OIDC.Issuer
	}
}
