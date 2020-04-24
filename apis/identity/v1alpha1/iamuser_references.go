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

package v1alpha1

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IAMUserNameReferencer is used to get the Name from a referenced IAM User object
type IAMUserNameReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus implements GetStatus method of AttributeReferencer interface
func (v *IAMUserNameReferencer) GetStatus(ctx context.Context, _ resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	return getUserStatus(ctx, v.Name, reader)
}

// Build retrieves and builds the IAM UserName
func (v *IAMUserNameReferencer) Build(ctx context.Context, _ resource.CanReference, reader client.Reader) (string, error) {
	user := IAMUser{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &user); err != nil {
		return "", err
	}

	return meta.GetExternalName(&user), nil
}

func getUserStatus(ctx context.Context, name string, reader client.Reader) ([]resource.ReferenceStatus, error) {
	user := IAMUser{}
	nn := types.NamespacedName{Name: name}
	if err := reader.Get(ctx, nn, &user); err != nil {
		if kerrors.IsNotFound(err) {
			return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotFound}}, nil
		}

		return nil, err
	}

	if !resource.IsConditionTrue(user.GetCondition(runtimev1alpha1.TypeReady)) {
		return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotReady}}, nil
	}

	return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceReady}}, nil
}
