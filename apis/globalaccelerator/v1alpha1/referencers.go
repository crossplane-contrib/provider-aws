package v1alpha1

import (
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/utils/ptr"
)

// AcceleratorARN returns the status.atProvider.ARN of an Accelerator
func AcceleratorARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Accelerator)
		if !ok {
			return ""
		}

		return ptr.Deref(r.Status.AtProvider.AcceleratorARN, "")
	}
}

// ListenerARN returns the status.atProvider.ARN of an Listener
func ListenerARN() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Listener)
		if !ok {
			return ""
		}

		return ptr.Deref(r.Status.AtProvider.ListenerARN, "")
	}
}
