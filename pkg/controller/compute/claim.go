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
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplaneio/stack-aws/apis/compute/v1alpha2"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	computev1alpha1 "github.com/crossplaneio/crossplane/apis/compute/v1alpha1"
)

// EKSClusterClaimController is responsible for adding the EKSCluster
// claim controller and its corresponding reconciler to the manager with any runtime configuration.
type EKSClusterClaimController struct{}

// SetupWithManager adds a controller that reconciles KubernetesCluster resource claims.
func (c *EKSClusterClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		computev1alpha1.KubernetesClusterKind,
		v1alpha2.EKSClusterKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(computev1alpha1.KubernetesClusterGroupVersionKind),
		resource.ClassKind(v1alpha2.EKSClusterClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.EKSClusterGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureEKSCluster),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.EKSClusterClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.EKSClusterGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.EKSClusterGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.EKSCluster{}}, &resource.EnqueueRequestForClaim{}).
		For(&computev1alpha1.KubernetesCluster{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureEKSCluster configures the supplied resource (presumed to be a
// EKSCluster) using the supplied resource claim (presumed to be a
// KubernetesCluster) and resource class.
func ConfigureEKSCluster(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	if _, cmok := cm.(*computev1alpha1.KubernetesCluster); !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), computev1alpha1.KubernetesClusterGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.EKSClusterClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.EKSClusterClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha2.EKSCluster)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.EKSClusterGroupVersionKind)
	}

	spec := &v1alpha2.EKSClusterSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		EKSClusterParameters: rs.SpecTemplate.EKSClusterParameters,
	}
	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}
