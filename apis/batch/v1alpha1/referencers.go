package v1alpha1

import (
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// ComputeEnvironmentARN returns ARN of the ComputeEnvironment resource.
func ComputeEnvironmentARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*ComputeEnvironment)
		if !ok {
			return ""
		}
		if cr.Status.AtProvider.ComputeEnvironmentARN == nil {
			return ""
		}
		return *cr.Status.AtProvider.ComputeEnvironmentARN
	}
}
