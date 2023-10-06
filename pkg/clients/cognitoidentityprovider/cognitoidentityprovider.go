/*
Copyright 2019 The Crossplane Authors.

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

package cognitoidentityprovider

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errGetProviderDetailsSecretFailed  = "cannot get ProviderDetails secret"
	errMissingProviderDetailsSecretRef = "expected providerDetailsSecretRef but was empty"
)

// NewResolver creates a new ResolverService.
func NewResolver() ResolverService {
	return &Resolver{}
}

// Resolver struct
type Resolver struct {
}

// ResolverService is the interface for Client methods
type ResolverService interface {
	GetProviderDetails(ctx context.Context, kube client.Client, ref *xpv1.SecretReference) (map[string]*string, error)
}

// GetProviderDetails fetches the referenced input secret containing the ProviderDetails for an IdentityProvider
func (r *Resolver) GetProviderDetails(ctx context.Context, kube client.Client, ref *xpv1.SecretReference) (map[string]*string, error) {

	if ref.Name == "" {
		return make(map[string]*string), errors.New(errMissingProviderDetailsSecretRef)
	}

	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	s := &corev1.Secret{}
	if err := kube.Get(ctx, nn, s); err != nil {
		return make(map[string]*string), errors.Wrap(err, errGetProviderDetailsSecretFailed)
	}

	providerDetails := make(map[string]*string)
	for key, value := range s.Data {
		str := string(value)
		providerDetails[key] = &str
	}

	return providerDetails, nil
}
