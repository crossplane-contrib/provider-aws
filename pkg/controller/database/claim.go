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

package database

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	databasev1alpha1 "github.com/crossplane/crossplane/apis/database/v1alpha1"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupPostgreSQLInstanceClaimScheduling adds a controller that reconciles
// PostgreSQLInstance claims that include a class selector but omit their class
// and resource references by picking a random matching RDSInstanceClass, if
// any.
func SetupPostgreSQLInstanceClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupPostgreSQLInstanceClaimDefaulting adds a controller that reconciles
// PostgreSQLInstance claims that omit their resource ref, class ref, and class
// selector by choosing a default RDSInstanceClass if one exists.
func SetupPostgreSQLInstanceClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupPostgreSQLInstanceClaimBinding adds a controller that reconciles
// PostgreSQLInstance claims with RDSInstances, dynamically provisioning them if
// needed.
func SetupPostgreSQLInstanceClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
		resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigurePostgreRDSInstance),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1beta1.RDSInstance{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigurePostgreRDSInstance configures the supplied resource (presumed
// to be a RDSInstance) using the supplied resource claim (presumed to be a
// PostgreSQLInstance) and resource class.
func ConfigurePostgreRDSInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	pg, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1beta1.RDSInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1beta1.RDSInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1beta1.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1beta1.RDSInstanceGroupVersionKind)
	}

	spec := &v1beta1.RDSInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	spec.ForProvider.Engine = v1beta1.PostgresqlEngine
	v, err := validateEngineVersion(aws.StringValue(spec.ForProvider.EngineVersion), pg.Spec.EngineVersion)
	if err != nil {
		return err
	}
	spec.ForProvider.EngineVersion = v

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}

// SetupMySQLInstanceClaimScheduling adds a controller that reconciles
// MySQLInstance claims that include a class selector but omit their class and
// resource references by picking a random matching RDSInstanceClass, if any.
func SetupMySQLInstanceClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(databasev1alpha1.MySQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupMySQLInstanceClaimDefaulting adds a controller that reconciles
// MySQLInstance claims that omit their resource ref, class ref, and class
// selector by choosing a default RDSInstanceClass if one exists.
func SetupMySQLInstanceClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(databasev1alpha1.MySQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupMySQLInstanceClaimBinding adds a controller that reconciles
// MySQLInstance claims with RDSInstances, dynamically provisioning them if
// needed
func SetupMySQLInstanceClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(databasev1alpha1.MySQLInstanceGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
		resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind),
		resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureMyRDSInstance),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1beta1.RDSInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1beta1.RDSInstance{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureMyRDSInstance configures the supplied resource (presumed to be
// a RDSInstance) using the supplied resource claim (presumed to be a
// MySQLInstance) and resource class.
func ConfigureMyRDSInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	my, cmok := cm.(*databasev1alpha1.MySQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.MySQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1beta1.RDSInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1beta1.RDSInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1beta1.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1beta1.RDSInstanceGroupVersionKind)
	}

	spec := &v1beta1.RDSInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	spec.ForProvider.Engine = v1beta1.MysqlEngine
	v, err := validateEngineVersion(aws.StringValue(spec.ForProvider.EngineVersion), my.Spec.EngineVersion)
	if err != nil {
		return err
	}
	spec.ForProvider.EngineVersion = v

	// TODO(muvaf): When ApplyModificationsImmediately is true, all up-to-date
	// checks return false.
	if spec.ForProvider.ApplyModificationsImmediately == nil {
		spec.ForProvider.ApplyModificationsImmediately = aws.Bool(true)
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

// validateEngineVersion compares class and claim engine values and returns an engine value or error
// if class values is empty - claim value returned (could be an empty string),
// otherwise if claim value is not a prefix of the class value - return an error
// else return class value
// Examples:
// class: "", claim: "" - result: "" (let the provider decide)
// class: 5.6, claim: "" - result: 5.6
// class: "", claim: 5.7 - result: 5.7
// class: 5.6.45, claim 5.6 - result: 5.6.45
// class: 5.6, claim 5.7 - result error
func validateEngineVersion(class, claim string) (*string, error) {
	if class == "" && claim == "" {
		return nil, nil
	}
	if class == "" && claim != "" {
		return &claim, nil
	}
	// class is definitely not empty string at this point.
	if strings.HasPrefix(class, claim) {
		return &class, nil
	}
	return nil, errors.Errorf("claim value [%s] does not match class value [%s]", claim, class)
}
