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
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// RoleARN returns the status.atProvider.ARN of a Role.
func RoleARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Role)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN

	}
}

// PolicyARN returns a function that returns the ARN of the given policy.
func PolicyARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Policy)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}

// UserARN returns a function that returns the ARN of the given policy.
func UserARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*User)
		if !ok {
			return ""
		}
		return r.Status.AtProvider.ARN
	}
}
