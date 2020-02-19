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

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/target"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane/apis/workload/v1alpha1"

	"github.com/crossplane/stack-aws/apis/compute/v1alpha3"
)

// SetupEKSClusterTarget adds a controller that propagates EKSCluster connection
// secrets to the connection secrets of their targets.
func SetupEKSClusterTarget(mgr ctrl.Manager, l logging.Logger) error {
	name := target.ControllerName(v1alpha3.EKSClusterGroupKind)
	p := resource.NewPredicates(resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.EKSClusterGroupVersionKind)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(strings.ToLower(fmt.Sprintf("kubernetestarget.%s.%s", v1alpha3.EKSClusterKind, v1alpha3.Group))).
		For(&v1alpha1.KubernetesTarget{}).
		WithEventFilter(p).
		Complete(target.NewReconciler(mgr,
			resource.TargetKind(v1alpha1.KubernetesTargetGroupVersionKind),
			resource.ManagedKind(v1alpha3.EKSClusterGroupVersionKind),
			target.WithLogger(l.WithValues("controller", name)),
			target.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}
