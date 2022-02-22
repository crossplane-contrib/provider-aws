package v1beta1

import (
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// QueueARN returns ARN of the Queue resource.
func QueueARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*Queue)
		if !ok {
			return ""
		}
		return cr.Status.AtProvider.ARN
	}
}
