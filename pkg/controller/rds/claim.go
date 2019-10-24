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

package rds

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplaneio/stack-aws/apis/database/v1alpha2"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	databasev1alpha1 "github.com/crossplaneio/crossplane/apis/database/v1alpha1"
)

// A PostgreSQLInstanceClaimSchedulingController reconciles PostgreSQLInstance
// claims that include a class selector but omit their class and resource
// references by picking a random matching RDSInstanceClass, if any.
type PostgreSQLInstanceClaimSchedulingController struct{}

// SetupWithManager sets up the
// PostgreSQLInstanceClaimSchedulingController using the supplied manager.
func (c *PostgreSQLInstanceClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		))
}

// A PostgreSQLInstanceClaimDefaultingController reconciles PostgreSQLInstance
// claims that omit their resource ref, class ref, and class selector by
// choosing a default RDSInstanceClass if one exists.
type PostgreSQLInstanceClaimDefaultingController struct{}

// SetupWithManager sets up the PostgreSQLInstanceClaimDefaultingController
// using the supplied manager.
func (c *PostgreSQLInstanceClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		))
}

// A PostgreSQLInstanceClaimController reconciles PostgreSQLInstance claims with
// RDSInstances, dynamically provisioning them if needed
type PostgreSQLInstanceClaimController struct{}

// SetupWithManager adds a controller that reconciles PostgreSQLInstance instance claims.
func (c *PostgreSQLInstanceClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigurePostgreRDSInstance),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.RDSInstance{}}, &resource.EnqueueRequestForClaim{}).
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

	rs, csok := cs.(*v1alpha2.RDSInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.RDSInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha2.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.RDSInstanceGroupVersionKind)
	}

	spec := &v1alpha2.RDSInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	spec.ForProvider.Engine = v1alpha2.PostgresqlEngine
	v, err := validateEngineVersion(spec.ForProvider.EngineVersion, pg.Spec.EngineVersion)
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

// A MySQLInstanceClaimSchedulingController reconciles MySQLInstance claims that
// include a class selector but omit their class and resource references by
// picking a random matching RDSInstanceClass, if any.
type MySQLInstanceClaimSchedulingController struct{}

// SetupWithManager sets up the MySQLInstanceClaimSchedulingController using the
// supplied manager.
func (c *MySQLInstanceClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		))
}

// A MySQLInstanceClaimDefaultingController reconciles MySQLInstance claims that
// omit their resource ref, class ref, and class selector by choosing a default
// RDSInstanceClass if one exists.
type MySQLInstanceClaimDefaultingController struct{}

// SetupWithManager sets up the MySQLInstanceClaimDefaultingController
// using the supplied manager.
func (c *MySQLInstanceClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		))
}

// A MySQLInstanceClaimController reconciles MySQLInstance claims with
// RDSInstances, dynamically provisioning them if needed
type MySQLInstanceClaimController struct{}

// SetupWithManager adds a controller that reconciles MySQLInstance instance claims.
func (c *MySQLInstanceClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.RDSInstanceKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
		resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureMyRDSInstance),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.RDSInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.RDSInstanceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.RDSInstance{}}, &resource.EnqueueRequestForClaim{}).
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

	rs, csok := cs.(*v1alpha2.RDSInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.RDSInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha2.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.RDSInstanceGroupVersionKind)
	}

	spec := &v1alpha2.RDSInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	spec.ForProvider.Engine = v1alpha2.MysqlEngine
	v, err := validateEngineVersion(spec.ForProvider.EngineVersion, my.Spec.EngineVersion)
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

// validateEngineVersion compares class and claim engine values and returns an engine value or error
// if class values is empty - claim value returned (could be an empty string),
// otherwise if claim value is not a prefix of the class value - return an error
// else return class value
// Examples:
// class: "", claim: "" - result: ""
// class: 5.6, claim: "" - result: 5.6
// class: "", claim: 5.7 - result: 5.7
// class: 5.6.45, claim 5.6 - result: 5.6.45
// class: 5.6, claim 5.7 - result error
func validateEngineVersion(class, claim string) (string, error) {
	if class == "" {
		return claim, nil
	}
	if strings.HasPrefix(class, claim) {
		return class, nil
	}
	return "", errors.Errorf("claim value [%s] does not match class value [%s]", claim, class)
}
