package provisionedproduct

import (
	"fmt"

	cfsdkv2types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/utils/pointer"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicecatalog/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

func provisioningParamsAreChanged(cfStackParams []cfsdkv2types.Parameter, currentParams []*svcapitypes.ProvisioningParameter) bool {
	if len(cfStackParams) != len(currentParams) {
		return true
	}

	cfStackKeyValue := make(map[string]string)
	for _, v := range cfStackParams {
		cfStackKeyValue[*v.ParameterKey] = pointer.StringDeref(v.ParameterValue, "")
	}

	for _, v := range currentParams {
		if cfv, ok := cfStackKeyValue[*v.Key]; ok && pointer.StringEqual(&cfv, v.Value) {
			continue
		} else {
			return true
		}
	}

	return false
}

// productOrArtifactAreChanged will attempt to determine whether or not the currently requested SC product and provisioning artifact IDs have changed when using names instead of IDs
func (c *custom) productOrArtifactAreChanged(ds *svcapitypes.ProvisionedProductParameters, resp *svcsdk.ProvisionedProductDetail) (bool, error) { // nolint:gocyclo
	var productID, artifactID string

	if ds.ProductID != nil {
		productID = pointer.StringDeref(ds.ProductID, "")
	}
	if ds.ProvisioningArtifactID != nil {
		artifactID = pointer.StringDeref(ds.ProvisioningArtifactID, "")
	}

	if productID == "" || artifactID == "" {
		// If we are using names for either the product or artifact, we should
		// do a lookup using the DescribeProduct API to find the matching
		// product (either by name or ID). In the returned output, we will have
		// access to both the product ID and all of the available provisioning
		// artifact IDs so we will be able to discover both IDs with a single
		// API call.
		dpInput := &svcsdk.DescribeProductInput{}

		if productID != "" {
			dpInput.Id = aws.String(productID)
		} else {
			dpInput.Name = ds.ProductName
		}

		describeProduct, err := c.client.DescribeProduct(dpInput)
		if err != nil {
			return true, errors.Wrap(err, "could not lookup product")
		}
		if describeProduct == nil || describeProduct.ProductViewSummary == nil {
			return true, errors.New("could not find product")
		}

		if productID == "" {
			// fill in the product ID if we do not yet have it
			if describeProduct == nil || describeProduct.ProductViewSummary == nil {
				return true, errors.New("could not find product")
			}
			productID = pointer.StringDeref(describeProduct.ProductViewSummary.ProductId, "")
		}

		if artifactID == "" {
			// find the matching artifact
			for _, artifact := range describeProduct.ProvisioningArtifacts {
				if pointer.StringEqual(ds.ProvisioningArtifactName, artifact.Name) {
					artifactID = pointer.StringDeref(artifact.Id, "")
					break
				}
			}
		}

		if artifactID == "" {
			// we were unable to find an artifact that matches
			return true, errors.New("could not find provisioning artifact")
		}
	}

	return pointer.StringDeref(resp.ProductId, "") != productID ||
		pointer.StringDeref(resp.ProvisioningArtifactId, "") != artifactID, nil
}

func genIdempotencyToken() string {
	return fmt.Sprintf("provider-aws-%s", uuid.New())
}
