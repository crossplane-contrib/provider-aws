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

package v1alpha2

import (
	"context"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IAMRoleARNReferencer is used to get the ARN from a referenced IAMRole object
type IAMRoleARNReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus implements GetStatus method of AttributeReferencer interface
func (v *IAMRoleARNReferencer) GetStatus(ctx context.Context, res resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	return getRoleStatus(ctx, v.Name, res, reader)
}

// Build retrieves and builds the IAMRoleARN
func (v *IAMRoleARNReferencer) Build(ctx context.Context, res resource.CanReference, reader client.Reader) (string, error) {
	role := IAMRole{}
	nn := types.NamespacedName{Name: v.Name, Namespace: res.GetNamespace()}
	if err := reader.Get(ctx, nn, &role); err != nil {
		return "", err
	}

	return role.Status.ARN, nil
}

// IAMRoleNameReferencer is used to get the Name from a referenced IAMRole object
type IAMRoleNameReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus implements GetStatus method of AttributeReferencer interface
func (v *IAMRoleNameReferencer) GetStatus(ctx context.Context, res resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	return getRoleStatus(ctx, v.Name, res, reader)
}

// Build retrieves and builds the IAMRoleName
func (v *IAMRoleNameReferencer) Build(ctx context.Context, res resource.CanReference, reader client.Reader) (string, error) {
	role := IAMRole{}
	nn := types.NamespacedName{Name: v.Name, Namespace: res.GetNamespace()}
	if err := reader.Get(ctx, nn, &role); err != nil {
		return "", err
	}

	return role.Spec.RoleName, nil
}

func getRoleStatus(ctx context.Context, name string, res resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	role := IAMRole{}
	nn := types.NamespacedName{Name: name, Namespace: res.GetNamespace()}
	if err := reader.Get(ctx, nn, &role); err != nil {
		if kerrors.IsNotFound(err) {
			return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotFound}}, nil
		}

		return nil, err
	}

	if !resource.IsConditionTrue(role.GetCondition(runtimev1alpha1.TypeReady)) {
		return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceNotReady}}, nil
	}

	return []resource.ReferenceStatus{{Name: name, Status: resource.ReferenceReady}}, nil
}
