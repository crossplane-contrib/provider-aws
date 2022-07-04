package common

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis"
)

const (
	// ErrNotTagged error message
	ErrNotTagged = "Resource does not implement the Tagged interface"
	// ErrUpdateTags error message
	ErrUpdateTags = "Failed to update tags for Tagged resource"
)

// Tagger is a controller initializer that adds all default tags to a managed
// resource if it implements the Tagged interface.
type Tagger struct {
	kube client.Client
}

// NewTagger creates a new Tagger instance.
func NewTagger(kube client.Client, _ apis.Tagged) *Tagger {
	return &Tagger{
		kube: kube,
	}
}

// Initialize adds the default tags to the given managed resource if it
// implements the Tagged interface.
func (t *Tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	tagged, ok := mg.(apis.Tagged)
	if !ok {
		return errors.New(ErrNotTagged)
	}
	shouldUpdate := false
	for k, v := range resource.GetExternalTags(mg) {
		if added := tagged.AddTag(k, v); added {
			shouldUpdate = true
		}
	}
	if shouldUpdate {
		return errors.Wrap(t.kube.Update(ctx, mg), ErrUpdateTags)
	}
	return nil
}
