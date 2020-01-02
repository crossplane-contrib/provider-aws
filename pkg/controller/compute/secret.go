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

package compute

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane/apis/compute/v1alpha1"

	"github.com/crossplaneio/stack-aws/apis/compute/v1alpha3"
)

// EKSClusterSecretController is responsible for adding the EKSCluster secret
// controller and its corresponding reconciler to the manager with any runtime configuration.
type EKSClusterSecretController struct{}

// SetupWithManager adds a controller that propagates EKSCluster connection
// secrets to the connection secrets of their resource claims.
func (c *EKSClusterSecretController) SetupWithManager(mgr ctrl.Manager) error {
	p := resource.NewPredicates(resource.AnyOf(
		resource.AllOf(resource.IsControlledByKind(v1alpha1.KubernetesClusterGroupVersionKind), resource.IsPropagated()),
		resource.AllOf(resource.IsControlledByKind(v1alpha3.EKSClusterGroupVersionKind), resource.IsPropagator()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(strings.ToLower(fmt.Sprintf("connectionsecret.%s.%s", v1alpha3.EKSClusterKind, v1alpha3.Group))).
		Watches(&source.Kind{Type: &corev1.Secret{}}, &resource.EnqueueRequestForPropagated{}).
		For(&corev1.Secret{}).
		WithEventFilter(p).
		Complete(resource.NewSecretPropagatingReconciler(mgr))
}
