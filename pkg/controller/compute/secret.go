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
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/secret"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane/apis/compute/v1alpha1"
	workloadv1alpha1 "github.com/crossplane/crossplane/apis/workload/v1alpha1"

	"github.com/crossplane/provider-aws/apis/compute/v1alpha3"
)

// SetupEKSClusterSecret adds a controller that propagates EKSCluster connection
// secrets to the connection secrets of their resource claims.
func SetupEKSClusterSecret(mgr ctrl.Manager, l logging.Logger) error {
	name := secret.ControllerName(v1alpha3.EKSClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &corev1.Secret{}}, &resource.EnqueueRequestForPropagated{}).
		For(&corev1.Secret{}).
		WithEventFilter(resource.NewPredicates(resource.AnyOf(
			resource.AllOf(resource.IsControlledByKind(v1alpha1.KubernetesClusterGroupVersionKind), resource.IsPropagated()),
			resource.AllOf(resource.IsControlledByKind(workloadv1alpha1.KubernetesTargetGroupVersionKind), resource.IsPropagated()),
			resource.AllOf(resource.IsControlledByKind(v1alpha3.EKSClusterGroupVersionKind), resource.IsPropagator())))).
		Complete(secret.NewReconciler(mgr,
			secret.WithLogger(l.WithValues("controller", name)),
			secret.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}
