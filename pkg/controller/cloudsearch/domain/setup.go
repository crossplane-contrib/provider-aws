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

package domain

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/aws/aws-sdk-go/service/cloudsearch/cloudsearchiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudsearch/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	legacypolicy "github.com/crossplane-contrib/provider-aws/pkg/utils/policy/old"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errDescribeServiceAccessPolicies = "cannot retrieve service access policies for Domain in AWS"
	errDescribeScalingParameters     = "cannot retrieve scaling parameters for Domain in AWS"
	errUpdateServiceAccessPolicies   = "cannot update service access policies for Domain in AWS"
	errUpdateScalingParameters       = "cannot update scaling parameters for Domain in AWS"
	errUpdateIndexing                = "cannot initiate indexing for Domain in AWS"

	infoConditionProcessing = "currently processing"
)

// SetupDomain adds a controller that reconciles CloudSearch domains.
func SetupDomain(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DomainGroupKind)
	opts := []option{setupHooks}

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
		resource.ManagedKind(svcapitypes.DomainGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Domain{}).
		Complete(r)
}

func setupHooks(e *external) {
	h := &hooks{client: e.client}
	e.postObserve = h.postObserve
	e.lateInitialize = h.lateInitialize
	e.isUpToDate = h.isUpToDate
	e.update = h.update
	e.preDelete = preDelete
}

type hooks struct {
	client cloudsearchiface.CloudSearchAPI
}

func (h *hooks) lateInitialize(forProvider *svcapitypes.DomainParameters, _ *svcsdk.DescribeDomainsOutput) error {
	spec := &forProvider.CustomDomainParameters

	resp, err := h.client.DescribeScalingParameters(&svcsdk.DescribeScalingParametersInput{
		DomainName: forProvider.DomainName,
	})
	if err != nil {
		return errors.Wrap(err, errDescribeScalingParameters)
	}

	current := resp.ScalingParameters.Options

	spec.DesiredReplicationCount = pointer.LateInitialize(spec.DesiredReplicationCount, current.DesiredReplicationCount)
	spec.DesiredInstanceType = pointer.LateInitialize(spec.DesiredInstanceType, current.DesiredInstanceType)
	spec.DesiredPartitionCount = pointer.LateInitialize(spec.DesiredPartitionCount, current.DesiredPartitionCount)

	respAccessPolicies, err := h.client.DescribeServiceAccessPolicies(&svcsdk.DescribeServiceAccessPoliciesInput{
		DomainName: forProvider.DomainName,
		Deployed:   pointer.ToOrNilIfZeroValue(false),
	})
	if err != nil {
		return errors.Wrap(err, errDescribeServiceAccessPolicies)
	}

	spec.AccessPolicies = pointer.LateInitialize(spec.AccessPolicies, respAccessPolicies.AccessPolicies.Options)

	return nil
}

func (h *hooks) isUpToDateScalingParameters(ctx context.Context, cr *svcapitypes.Domain, domainName *string) (bool, error) {
	in := svcsdk.DescribeScalingParametersInput{
		DomainName: domainName,
	}

	resp, err := h.client.DescribeScalingParametersWithContext(ctx, &in)

	if err != nil {
		return false, errors.Wrap(err, errDescribeScalingParameters)
	}

	spec := cr.Spec.ForProvider.CustomDomainParameters
	current := resp.ScalingParameters.Options

	isUpToDate := pointer.Int64Value(spec.DesiredReplicationCount) == pointer.Int64Value(current.DesiredReplicationCount) &&
		pointer.StringValue(spec.DesiredInstanceType) == pointer.StringValue(current.DesiredInstanceType) &&
		pointer.Int64Value(spec.DesiredPartitionCount) == pointer.Int64Value(current.DesiredPartitionCount)

	return isUpToDate, nil
}

func (h *hooks) isUpToDateAccessPolicies(ctx context.Context, cr *svcapitypes.Domain, domainName *string) (bool, error) {
	in := svcsdk.DescribeServiceAccessPoliciesInput{
		DomainName: domainName,
		Deployed:   pointer.ToOrNilIfZeroValue(false), // include pending policies as well
	}

	resp, err := h.client.DescribeServiceAccessPoliciesWithContext(ctx, &in)

	if err != nil {
		return false, errors.Wrap(err, errDescribeServiceAccessPolicies)
	}

	spec := cr.Spec.ForProvider.CustomDomainParameters
	current := resp.AccessPolicies

	isUpToDate := legacypolicy.IsPolicyUpToDate(spec.AccessPolicies, current.Options) && !pointer.BoolValue(current.Status.PendingDeletion)

	return isUpToDate, nil
}

func (h *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.Domain, obj *svcsdk.DescribeDomainsOutput) (bool, string, error) {
	ds := obj.DomainStatusList[0]

	scalingParametersUpToDate, err := h.isUpToDateScalingParameters(ctx, cr, ds.DomainName)
	if !scalingParametersUpToDate || err != nil {
		return false, "", err
	}
	accessPoliciesUpToDate, err := h.isUpToDateAccessPolicies(ctx, cr, ds.DomainName)
	if !accessPoliciesUpToDate || err != nil {
		return false, "", err
	}

	return !pointer.BoolValue(ds.RequiresIndexDocuments), "", nil
}

func updateConditions(cr *svcapitypes.Domain, ds *svcsdk.DomainStatus) {
	switch {
	case pointer.BoolValue(ds.Deleted):
		cr.SetConditions(xpv1.Deleting())
	case pointer.BoolValue(ds.Created) && ds.SearchService.Endpoint != nil && ds.DocService.Endpoint != nil:
		if pointer.BoolValue(ds.Processing) {
			cr.SetConditions(xpv1.Available().WithMessage(infoConditionProcessing))
		} else {
			cr.SetConditions(xpv1.Available())
		}
	default:
		cr.SetConditions(xpv1.Creating())
	}
}

func (h *hooks) postObserve(ctx context.Context, cr *svcapitypes.Domain, obj *svcsdk.DescribeDomainsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	ds := obj.DomainStatusList[0]

	updateConditions(cr, ds)

	obs.ConnectionDetails = managed.ConnectionDetails{
		"docServiceEndpoint":    []byte(pointer.StringValue(ds.DocService.Endpoint)),
		"searchServiceEndpoint": []byte(pointer.StringValue(ds.SearchService.Endpoint)),
	}

	return obs, nil
}

func (h *hooks) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	isUpToDateAccessPolicies, err := h.isUpToDateAccessPolicies(ctx, cr, cr.Spec.ForProvider.DomainName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateServiceAccessPolicies)
	}
	if !isUpToDateAccessPolicies {
		_, err := h.client.UpdateServiceAccessPoliciesWithContext(ctx, &svcsdk.UpdateServiceAccessPoliciesInput{
			DomainName:     cr.Spec.ForProvider.DomainName,
			AccessPolicies: cr.Spec.ForProvider.AccessPolicies,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateServiceAccessPolicies)
		}
		return managed.ExternalUpdate{}, nil
	}

	isUpToDateScalingParameters, err := h.isUpToDateScalingParameters(ctx, cr, cr.Spec.ForProvider.DomainName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateScalingParameters)
	}
	if !isUpToDateScalingParameters {
		_, err = h.client.UpdateScalingParametersWithContext(ctx, &svcsdk.UpdateScalingParametersInput{
			DomainName: cr.Spec.ForProvider.DomainName,
			ScalingParameters: &svcsdk.ScalingParameters{
				DesiredReplicationCount: cr.Spec.ForProvider.DesiredReplicationCount,
				DesiredInstanceType:     cr.Spec.ForProvider.DesiredInstanceType,
				DesiredPartitionCount:   cr.Spec.ForProvider.DesiredPartitionCount,
			},
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateScalingParameters)
		}
		return managed.ExternalUpdate{}, nil
	}

	if pointer.BoolValue(cr.Status.AtProvider.RequiresIndexDocuments) {
		_, err = h.client.IndexDocumentsWithContext(ctx, &svcsdk.IndexDocumentsInput{
			DomainName: cr.Spec.ForProvider.DomainName,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateIndexing)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Domain, _ *svcsdk.DeleteDomainInput) (bool, error) {
	return pointer.BoolValue(cr.Status.AtProvider.Deleted), nil
}
