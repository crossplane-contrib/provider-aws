package v1alpha1

import (
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	elbv2 "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
)

// LoadBalancerName returns the Name of a LoadBalancer
func LoadBalancerName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		lb, ok := mg.(*elbv2.LoadBalancer)
		if !ok {
			return ""
		}
		if len(lb.Status.AtProvider.LoadBalancers) == 0 {
			return ""
		}
		return *lb.Status.AtProvider.LoadBalancers[0].LoadBalancerName
	}
}
