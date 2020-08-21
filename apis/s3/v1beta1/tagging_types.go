package v1beta1

// Tagging is the container for TagSet elements.
type Tagging struct {
	// A collection for a set of tags
	// TagSet is a required field
	TagSet []Tag `json:"tagSet"`
}

// Tag is a container for a key value name pair.
type Tag struct {
	// Name of the tag.
	// Key is a required field
	Key *string `json:"key,omitempty"`

	// Value of the tag.
	// Value is a required field
	Value *string `json:"value,omitempty"`
}
