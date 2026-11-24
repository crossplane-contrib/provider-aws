package v1alpha1

// TagIgnoreRule defines a single rule for ignoring a tag during diffing.
// The Key may be an exact tag key or a prefix pattern that ends with *.
type TagIgnoreRule struct {
	// Key of the ignore rule. Supports exact key match or prefix* glob.
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// CustomProvisionedProductParameters includes the custom fields.
type CustomProvisionedProductParameters struct {
	// TagsIgnore contains rules for tags that are never sent to AWS on update
	// (e.g. tags that cannot be changed after provisioning). A rule key supports
	// exact match (e.g. "created-by") or a prefix glob with a trailing * (e.g. "mdp:*").
	// +optional
	TagsIgnore []TagIgnoreRule `json:"tagsIgnore,omitempty"`
}

// CustomProvisionedProductObservation includes the custom status fields of ProvisionedProduct.
type CustomProvisionedProductObservation struct{}
