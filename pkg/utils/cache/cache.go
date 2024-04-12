/*
Copyright 2024 The Crossplane Authors.

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

package cache

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const EXTERNAL_NAME_FIELD = "externalName"

// IndexByExternalName setups index by extrenal name in the cache for a given resource type.
func IndexByExternalName(ctx context.Context, cache cache.Cache, object client.Object) error {
	return cache.IndexField(ctx, object, EXTERNAL_NAME_FIELD, func(o client.Object) []string {
		return []string{meta.GetExternalName(o)}
	})
}

// ListByExternalName lists objects in the cache by external name.
func ListByExternalName(ctx context.Context, cache cache.Cache, objects client.ObjectList, value string) error {
	return cache.List(ctx, objects, client.MatchingFields{EXTERNAL_NAME_FIELD: value})
}
