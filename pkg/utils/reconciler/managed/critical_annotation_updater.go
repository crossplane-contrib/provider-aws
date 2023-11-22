/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package managed provides a custom implementation of RetryingCriticalAnnotationUpdater
// from the crossplane-runtime package managed (github.com/crossplane/crossplane-runtime/pkg/reconciler/managed/api.go)
// This custom implementation is currently used in all controllers to revert back to the behavior before
// this breaking change from crossplane-runtime:v1.14.0 (https://github.com/crossplane/crossplane-runtime/pull/526)
// See also https://github.com/crossplane-contrib/provider-aws/pull/1953 for more information
package managed

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Error strings.
const (
	errUpdateCriticalAnnotations = "cannot update critical annotations"
)

// A RetryingCriticalAnnotationUpdater is a CriticalAnnotationUpdater that
// retries annotation updates in the face of API server errors.
type RetryingCriticalAnnotationUpdater struct {
	client client.Client
}

// NewRetryingCriticalAnnotationUpdater returns a CriticalAnnotationUpdater that
// retries annotation updates in the face of API server errors.
func NewRetryingCriticalAnnotationUpdater(c client.Client) *RetryingCriticalAnnotationUpdater {
	return &RetryingCriticalAnnotationUpdater{client: c}
}

// UpdateCriticalAnnotations updates (i.e. persists) the annotations of the
// supplied Object. It retries in the face of any API server error several times
// in order to ensure annotations that contain critical state are persisted. Any
// pending changes to the supplied Object's spec, status, or other metadata are
// reset to their current state according to the API server.
func (u *RetryingCriticalAnnotationUpdater) UpdateCriticalAnnotations(ctx context.Context, o client.Object) error {
	a := o.GetAnnotations()
	err := retry.OnError(retry.DefaultRetry, resource.IsAPIError, func() error {
		nn := types.NamespacedName{Name: o.GetName()}
		if err := u.client.Get(ctx, nn, o); err != nil {
			return err
		}
		meta.AddAnnotations(o, a)
		return u.client.Update(ctx, o)
	})
	return errors.Wrap(err, errUpdateCriticalAnnotations)
}
