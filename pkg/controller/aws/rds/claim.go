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
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplaneio/crossplane/pkg/apis/aws/database/v1alpha1"
	corev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/core/v1alpha1"
	databasev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/database/v1alpha1"
	"github.com/crossplaneio/crossplane/pkg/resource"
)

// AddPostgreSQLClaim adds a controller that reconciles PostgreSQLInstance resource claims by
// managing RDSInstance resources to the supplied Manager.
func AddPostgreSQLClaim(mgr manager.Manager) error {
	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(corev1alpha1.ResourceClassGroupVersionKind),
		resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigurePostgreRDSInstance),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	name := strings.ToLower(fmt.Sprintf("%s.%s", databasev1alpha1.PostgreSQLInstanceKind, controllerName))
	c, err := controller.New(name, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrapf(err, "cannot create %s controller", name)
	}

	if err := c.Watch(&source.Kind{Type: &v1alpha1.RDSInstance{}}, &resource.EnqueueRequestForClaim{}); err != nil {
		return errors.Wrapf(err, "cannot watch for %s", v1alpha1.RDSInstanceGroupVersionKind)
	}

	p := v1alpha1.RDSInstanceKindAPIVersion
	return errors.Wrapf(c.Watch(
		&source.Kind{Type: &databasev1alpha1.PostgreSQLInstance{}},
		&handler.EnqueueRequestForObject{},
		resource.NewPredicates(resource.ObjectHasProvisioner(mgr.GetClient(), p)),
	), "cannot watch for %s", databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
}

// ConfigurePostgreRDSInstance configures the supplied resource (presumed
// to be a RDSInstance) using the supplied resource claim (presumed to be a
// PostgreSQLInstance) and resource class.
func ConfigurePostgreRDSInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	pg, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*corev1alpha1.ResourceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), corev1alpha1.ResourceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha1.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha1.RDSInstanceGroupVersionKind)
	}

	spec := v1alpha1.NewRDSInstanceSpec(rs.Parameters)
	spec.Engine = v1alpha1.PostgresqlEngine
	v, err := validateEngineVersion(spec.EngineVersion, pg.Spec.EngineVersion)
	if err != nil {
		return err
	}
	spec.EngineVersion = v

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.ProviderReference
	spec.ReclaimPolicy = rs.ReclaimPolicy

	i.Spec = *spec

	return nil
}

// AddMySQLClaim adds a controller that reconciles MySQLInstance resource claims by
// managing RDSInstance resources to the supplied Manager.
func AddMySQLClaim(mgr manager.Manager) error {
	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
		resource.ClassKind(corev1alpha1.ResourceClassGroupVersionKind),
		resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureMyRDSInstance),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	name := strings.ToLower(fmt.Sprintf("%s.%s", databasev1alpha1.MySQLInstanceKind, controllerName))
	c, err := controller.New(name, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrapf(err, "cannot create %s controller", name)
	}

	if err := c.Watch(
		&source.Kind{Type: &v1alpha1.RDSInstance{}},
		&resource.EnqueueRequestForClaim{},
	); err != nil {
		return errors.Wrapf(err, "cannot watch for %s", v1alpha1.RDSInstanceGroupVersionKind)
	}

	p := v1alpha1.RDSInstanceKindAPIVersion
	return errors.Wrapf(c.Watch(
		&source.Kind{Type: &databasev1alpha1.MySQLInstance{}},
		&handler.EnqueueRequestForObject{},
		resource.NewPredicates(resource.ObjectHasProvisioner(mgr.GetClient(), p)),
	), "cannot watch for %s", databasev1alpha1.MySQLInstanceGroupVersionKind)
}

// ConfigureMyRDSInstance configures the supplied resource (presumed to be
// a RDSInstance) using the supplied resource claim (presumed to be a
// MySQLInstance) and resource class.
func ConfigureMyRDSInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	my, cmok := cm.(*databasev1alpha1.MySQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.MySQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*corev1alpha1.ResourceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), corev1alpha1.ResourceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha1.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha1.RDSInstanceGroupVersionKind)
	}

	spec := v1alpha1.NewRDSInstanceSpec(rs.Parameters)
	spec.Engine = v1alpha1.MysqlEngine
	v, err := validateEngineVersion(spec.EngineVersion, my.Spec.EngineVersion)
	if err != nil {
		return err
	}
	spec.EngineVersion = v

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.ProviderReference
	spec.ReclaimPolicy = rs.ReclaimPolicy

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
