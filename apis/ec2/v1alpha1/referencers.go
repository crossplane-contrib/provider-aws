package v1alpha1

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
)

// ResolveReferences of this NatGateway
func (mg *NATGateway) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// // Resolve spec.subnetId
	subnetIDResponse, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: aws.StringValue(mg.Spec.ForProvider.SubnetID),
		Reference:    mg.Spec.ForProvider.SubnetIDRef,
		Selector:     mg.Spec.ForProvider.SubnetIDSelector,
		To:           reference.To{Managed: &v1beta1.Subnet{}, List: &v1beta1.SubnetList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.SubnetID = aws.String(subnetIDResponse.ResolvedValue)
	mg.Spec.ForProvider.SubnetIDRef = subnetIDResponse.ResolvedReference

	// // Resolve spec.elasticIp
	AllocationIDRespone, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: aws.StringValue(mg.Spec.ForProvider.AllocationID),
		Reference:    mg.Spec.ForProvider.AllocationIDRef,
		Selector:     mg.Spec.ForProvider.AllocationIDSelector,
		To:           reference.To{Managed: &ElasticIP{}, List: &ElasticIPList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return err
	}
	mg.Spec.ForProvider.AllocationID = aws.String(AllocationIDRespone.ResolvedValue)
	mg.Spec.ForProvider.AllocationIDRef = AllocationIDRespone.ResolvedReference

	return nil
}
