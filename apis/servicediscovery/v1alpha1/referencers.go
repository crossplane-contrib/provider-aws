package v1alpha1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2v1beta1 "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences of this PrivateDNSNamespace.
func (mg *PrivateDNSNamespace) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.vpc
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.VPC),
		Reference:    mg.Spec.ForProvider.VPCRef,
		Selector:     mg.Spec.ForProvider.VPCSelector,
		To:           reference.To{Managed: &ec2v1beta1.VPC{}, List: &ec2v1beta1.VPCList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.vpc")
	}
	mg.Spec.ForProvider.VPC = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.VPCRef = rsp.ResolvedReference
	return nil
}
