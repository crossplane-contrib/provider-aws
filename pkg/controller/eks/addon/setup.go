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

package addon

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awseks "github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	eksv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/eks/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errNotEKSCluster                  = "managed resource is not an EKS cluster custom resource"
	errKubeUpdateFailed               = "cannot update EKS cluster custom resource"
	errTagResource                    = "cannot tag resource"
	errUntagResource                  = "cannot untag resource"
	errDescribePodIdentityAssociation = "failed to describe Pod Identity Association"
	errEmptyPodIdentityAssociation    = "describe Pod Identity Association returned an unexpected empty result"
)

// SetupAddon adds a controller that reconciles Clusters.
func SetupAddon(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(eksv1alpha1.AddonGroupKind)
	opts := []option{
		setupHooks,
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(eksv1alpha1.AddonGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&eksv1alpha1.Addon{}).
		Complete(r)
}

func setupHooks(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	e.lateInitialize = lateInitialize
	h := &hooks{client: e.client, kube: e.kube, cache: &cache{}}
	e.isUpToDate = h.isUpToDate
	e.preUpdate = h.preUpdate
	e.postUpdate = h.postUpdate
	e.preCreate = preCreate
	e.postCreate = postCreate
	e.preDelete = preDelete
}

type hooks struct {
	client eksiface.EKSAPI
	kube   client.Client

	cache *cache
}

type cache struct {
	removePIAs bool
}

func preObserve(_ context.Context, cr *eksv1alpha1.Addon, obj *awseks.DescribeAddonInput) error {
	obj.ClusterName = cr.Spec.ForProvider.ClusterName
	return nil
}

func postObserve(_ context.Context, cr *eksv1alpha1.Addon, _ *awseks.DescribeAddonOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(cr.Status.AtProvider.Status) {
	case awseks.AddonStatusCreating:
		cr.SetConditions(xpv1.Creating())
	case awseks.AddonStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	case awseks.AddonStatusActive:
		cr.SetConditions(xpv1.Available())
	case awseks.AddonStatusUpdating:
		cr.SetConditions(xpv1.Available().WithMessage(awseks.AddonStatusUpdating))
		// Prevent Update() call during update state - which will fail.
		obs.ResourceUpToDate = true
	default:
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func lateInitialize(spec *eksv1alpha1.AddonParameters, resp *awseks.DescribeAddonOutput) error {
	if resp.Addon != nil {
		spec.ServiceAccountRoleARN = pointer.LateInitialize(spec.ServiceAccountRoleARN, resp.Addon.ServiceAccountRoleArn)
	}
	return nil
}

func (h *hooks) isUpToDate(ctx context.Context, cr *eksv1alpha1.Addon, resp *awseks.DescribeAddonOutput) (bool, string, error) {
	switch {
	case resp.Addon == nil,
		cr.Spec.ForProvider.AddonVersion != nil && pointer.StringValue(cr.Spec.ForProvider.AddonVersion) != pointer.StringValue(resp.Addon.AddonVersion),
		cr.Spec.ForProvider.ServiceAccountRoleARN != nil && pointer.StringValue(cr.Spec.ForProvider.ServiceAccountRoleARN) != pointer.StringValue(resp.Addon.ServiceAccountRoleArn):
		return false, "", nil
	}

	// PodIdentityAssociations
	piasUpToDate, piasUpToDateErr := h.isUpToDatePodIdentityAssociations(ctx, cr, resp)
	if piasUpToDateErr != nil {
		return false, "", piasUpToDateErr
	}
	if !piasUpToDate {
		return false, "", nil
	}

	// AddOn Configuration Values
	configUpToDate, configUpToDateDiff, configUpToDateErr := isUpToDateConfigurationValues(cr, resp)
	if configUpToDateErr != nil {
		return false, "", configUpToDateErr
	}
	if !configUpToDate {
		return false, configUpToDateDiff, nil
	}

	add, remove := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, resp.Addon.Tags)
	return len(add) == 0 && len(remove) == 0, "", nil
}

func (h *hooks) isUpToDatePodIdentityAssociations(ctx context.Context, cr *eksv1alpha1.Addon, resp *awseks.DescribeAddonOutput) (bool, error) { //nolint:gocyclo
	if len(cr.Spec.ForProvider.PodIdentityAssociations) != len(resp.Addon.PodIdentityAssociations) {
		if len(cr.Spec.ForProvider.PodIdentityAssociations) == 0 {
			// Edge Case for update request
			h.cache.removePIAs = true
		}
		return false, nil
	}
	// From the Addon we only get the ARNs of the PodIdentityAssociations.
	// To be able to compare/match them with the user provided roleArn and namespace,
	// we need to Describe the PodIdentityAssociations individually.
	// Copied/inspired from Terraform: https://github.com/hashicorp/terraform-provider-aws/blob/62fba4b53b0a584b4f30cc73284d5315974289af/internal/service/eks/addon.go#L461
	describedPIAs := []*eksv1alpha1.AddonPodIdentityAssociations{}
	for _, observedPIAArn := range resp.Addon.PodIdentityAssociations {
		// schema "arn:aws:eks:<region>:<account-id>:podidentityassociation/<cluster-name>/a-andother17characs"
		parts := strings.Split(aws.StringValue(observedPIAArn), "/")
		if len(parts) != 3 {
			return false, fmt.Errorf(`unable to extract Pod Identity Association ID from observed ARN %s`, *observedPIAArn)
		}
		associationID := parts[2]
		pia, err := h.client.DescribePodIdentityAssociationWithContext(ctx, &awseks.DescribePodIdentityAssociationInput{
			AssociationId: &associationID,
			ClusterName:   cr.Spec.ForProvider.ClusterName,
		})
		if err != nil {
			return false, errors.Wrap(err, errDescribePodIdentityAssociation)
		}
		if pia == nil || pia.Association == nil {
			return false, errors.New(errEmptyPodIdentityAssociation + "; id: " + associationID)
		}
		describedPIAs = append(describedPIAs, &eksv1alpha1.AddonPodIdentityAssociations{
			RoleARN:        pia.Association.RoleArn,
			ServiceAccount: pia.Association.ServiceAccount,
		})
		found := slices.ContainsFunc(cr.Spec.ForProvider.PodIdentityAssociations, func(a *eksv1alpha1.AddonPodIdentityAssociations) bool {
			return a != nil &&
				pointerValueEqual(a.RoleARN, pia.Association.RoleArn) &&
				pointerValueEqual(a.ServiceAccount, pia.Association.ServiceAccount)
		})
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func pointerValueEqual[T comparable](a, b *T) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}

func isUpToDateConfigurationValues(cr *eksv1alpha1.Addon, obj *awseks.DescribeAddonOutput) (bool, string, error) {
	// Handle nil pointer refs
	crConfigurationValues := aws.StringValue(cr.Spec.ForProvider.ConfigurationValues)
	objConfigurationValues := aws.StringValue(obj.Addon.ConfigurationValues)
	if crConfigurationValues == "" && objConfigurationValues == "" {
		return true, "", nil
	}

	if crConfigurationValues == "" && objConfigurationValues != "" {
		return false, "", nil
	}
	if crConfigurationValues != "" && objConfigurationValues == "" {
		return false, "", nil
	}

	// Normalize the data
	objConfigurationValuesCmp, objErr := convertConfigurationValuesToObj(objConfigurationValues)
	if objErr != nil {
		return false, "", objErr
	}
	crConfigurationValuesCmp, crErr := convertConfigurationValuesToObj(crConfigurationValues)
	if crErr != nil {
		return false, "", crErr
	}
	// Compare objects and return a diff
	diff := cmp.Diff(objConfigurationValuesCmp, crConfigurationValuesCmp)
	return diff == "", diff, nil
}

// convertConfigurationValuesToObj will deserialize in order to normalize and compare
func convertConfigurationValuesToObj(configurationValues string) (map[string]interface{}, error) {
	var objConfigurationValues map[string]interface{}
	// Yaml parser is able to handle Unmarshal for both YAML and JSON
	err := yaml.Unmarshal([]byte(configurationValues), &objConfigurationValues)
	return objConfigurationValues, err
}

func (h *hooks) preUpdate(_ context.Context, cr *eksv1alpha1.Addon, obj *awseks.UpdateAddonInput) error {
	obj.ClusterName = cr.Spec.ForProvider.ClusterName

	// Need to set empty array [] in next update-call,
	// if we want to remove all PodIdentityAssociations
	if h.cache.removePIAs {
		obj.PodIdentityAssociations = []*awseks.AddonPodIdentityAssociations{}
	}
	return nil
}

func (h *hooks) postUpdate(ctx context.Context, cr *eksv1alpha1.Addon, resp *awseks.UpdateAddonOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Tag update needs to separate because UpdateAddon does not include tags (for unknown reason).

	desc, err := h.client.DescribeAddonWithContext(ctx, &awseks.DescribeAddonInput{
		AddonName:   cr.Spec.ForProvider.AddonName,
		ClusterName: cr.Spec.ForProvider.ClusterName,
	})
	if err != nil || desc.Addon == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribe)
	}

	add, remove := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, desc.Addon.Tags)
	if len(add) > 0 {
		_, err := h.client.TagResourceWithContext(ctx, &awseks.TagResourceInput{
			ResourceArn: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Tags:        add,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTagResource)
		}
	}
	if len(remove) > 0 {
		_, err := h.client.UntagResourceWithContext(ctx, &awseks.UntagResourceInput{
			ResourceArn: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			TagKeys:     remove,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUntagResource)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func preCreate(_ context.Context, cr *eksv1alpha1.Addon, obj *awseks.CreateAddonInput) error {
	obj.ClusterName = cr.Spec.ForProvider.ClusterName
	return nil
}

func postCreate(_ context.Context, cr *eksv1alpha1.Addon, res *awseks.CreateAddonOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	if res.Addon != nil && meta.GetExternalName(cr) != pointer.StringValue(res.Addon.AddonArn) {
		meta.SetExternalName(cr, pointer.StringValue(res.Addon.AddonArn))
	}
	return cre, nil
}

func preDelete(_ context.Context, cr *eksv1alpha1.Addon, obj *awseks.DeleteAddonInput) (bool, error) {
	obj.ClusterName = cr.Spec.ForProvider.ClusterName
	return false, nil
}
