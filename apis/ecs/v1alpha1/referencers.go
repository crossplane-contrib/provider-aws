package v1alpha1

import (
	"github.com/aws/aws-sdk-go/aws"
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
		return aws.StringValue(lb.Status.AtProvider.LoadBalancerName)
	}
}
