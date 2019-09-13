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

package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplaneio/stack-aws/aws/apis/storage/v1alpha2"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"
)

var s3ACL = map[storagev1alpha1.PredefinedACL]s3.BucketCannedACL{
	storagev1alpha1.ACLPrivate:           s3.BucketCannedACLPrivate,
	storagev1alpha1.ACLPublicRead:        s3.BucketCannedACLPublicRead,
	storagev1alpha1.ACLPublicReadWrite:   s3.BucketCannedACLPublicReadWrite,
	storagev1alpha1.ACLAuthenticatedRead: s3.BucketCannedACLAuthenticatedRead,
}

// BucketClaimController is responsible for adding the Bucket claim controller and its
// corresponding reconciler to the manager with any runtime configuration.
type BucketClaimController struct{}

// SetupWithManager adds a controller that reconciles Bucket resource claims.
func (c *BucketClaimController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKinds{Portable: storagev1alpha1.BucketClassGroupVersionKind, NonPortable: v1alpha2.S3BucketClassGroupVersionKind},
		resource.ManagedKind(v1alpha2.S3BucketGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureS3Bucket),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	name := strings.ToLower(fmt.Sprintf("%s.%s", storagev1alpha1.BucketKind, controllerName))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.S3Bucket{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.HasClassReferenceKinds(mgr.GetClient(), mgr.GetScheme(), resource.ClassKinds{Portable: storagev1alpha1.BucketClassGroupVersionKind, NonPortable: v1alpha2.S3BucketClassGroupVersionKind}))).
		Complete(r)
}

// ConfigureS3Bucket configures the supplied resource (presumed
// to be a S3Bucket) using the supplied resource claim (presumed
// to be a Bucket) and resource class.
func ConfigureS3Bucket(_ context.Context, cm resource.Claim, cs resource.NonPortableClass, mg resource.Managed) error {
	b, cmok := cm.(*storagev1alpha1.Bucket)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.S3BucketClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.S3BucketClassGroupVersionKind)
	}

	s3b, mgok := mg.(*v1alpha2.S3Bucket)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.S3BucketGroupVersionKind)
	}

	spec := &v1alpha2.S3BucketSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		S3BucketParameters: rs.SpecTemplate.S3BucketParameters,
	}

	if b.Spec.Name != "" {
		spec.NameFormat = b.Spec.Name
	}

	if b.Spec.PredefinedACL != nil {
		spec.CannedACL = translateACL(b.Spec.PredefinedACL)
	}

	if b.Spec.LocalPermission != nil {
		spec.LocalPermission = b.Spec.LocalPermission
	}

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	s3b.Spec = *spec

	return nil
}

func translateACL(acl *storagev1alpha1.PredefinedACL) *s3.BucketCannedACL {
	if acl == nil {
		return nil
	}
	s3acl, found := s3ACL[*acl]
	if !found {
		return nil
	}
	return &s3acl
}
