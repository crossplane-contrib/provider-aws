/*
Copyright 2021 The Crossplane Authors.
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

package securityconfiguration

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupSecurityConfiguration adds a controller that reconciles SecurityConfiguration.
func SetupSecurityConfiguration(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.SecurityConfigurationGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.SecurityConfigurationGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.SecurityConfiguration{}).
		Complete(r)
}

func preDelete(_ context.Context, cr *svcapitypes.SecurityConfiguration, obj *svcsdk.DeleteSecurityConfigurationInput) (bool, error) {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.SecurityConfiguration, obj *svcsdk.GetSecurityConfigurationInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.SecurityConfiguration, obj *svcsdk.GetSecurityConfigurationOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// field not set as expected in zz_conversions.go GenerateSecurityConfiguration()
	// the setting of the field in zz_controller.go Create() seems to not work correctly (but Name works fine there)
	cr.Status.AtProvider.CreatedTimestamp = fromTimePtr(obj.SecurityConfiguration.CreatedTimeStamp)

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.SecurityConfiguration, obj *svcsdk.CreateSecurityConfigurationOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.Name))
	return managed.ExternalCreation{}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.SecurityConfiguration, obj *svcsdk.CreateSecurityConfigurationInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	if cr.Spec.ForProvider.CustomEncryptionConfiguration != nil {

		obj.EncryptionConfiguration = &svcsdk.EncryptionConfiguration{}
		if cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption != nil {

			obj.EncryptionConfiguration.CloudWatchEncryption = &svcsdk.CloudWatchEncryption{
				CloudWatchEncryptionMode: cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.CloudWatchEncryptionMode,
				KmsKeyArn:                cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomCloudWatchEncryption.KMSKeyARN,
			}
		}

		if cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption != nil {

			obj.EncryptionConfiguration.JobBookmarksEncryption = &svcsdk.JobBookmarksEncryption{
				JobBookmarksEncryptionMode: cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.JobBookmarksEncryptionMode,
				KmsKeyArn:                  cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomJobBookmarksEncryption.KMSKeyARN,
			}
		}

		obj.EncryptionConfiguration.S3Encryption = []*svcsdk.S3Encryption{}
		for _, s3Encryption := range cr.Spec.ForProvider.CustomEncryptionConfiguration.CustomS3Encryption {
			obj.EncryptionConfiguration.S3Encryption = append(obj.EncryptionConfiguration.S3Encryption, &svcsdk.S3Encryption{
				S3EncryptionMode: s3Encryption.S3EncryptionMode,
				KmsKeyArn:        s3Encryption.KMSKeyARN,
			})
		}
	}

	return nil
}

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func fromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}
