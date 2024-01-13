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

//nolint:gocyclo,staticcheck,golint
package distribution

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// TODO: Aren't these defined as an API constant somewhere in aws-sdk-go?
// Generated zz_enums.go seems not to contain it either
const (
	stateDeployed = "Deployed"
)

// SetupDistribution adds a controller that reconciles Distribution.
func SetupDistribution(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DistributionGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			opts: []option{
				func(e *external) {
					e.preCreate = preCreate
					e.postCreate = postCreate
					e.lateInitialize = lateInitialize
					e.preObserve = preObserve
					e.postObserve = postObserve
					e.isUpToDate = isUpToDate
					e.preUpdate = preUpdate
					d := &deleter{external: e}
					e.preDelete = d.preDelete
					e.postUpdate = postUpdate
				},
			},
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.DistributionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Distribution{}).
		Complete(r)
}

func preCreate(_ context.Context, cr *svcapitypes.Distribution, cdi *svcsdk.CreateDistributionInput) error {
	if pointer.StringValue(cr.Spec.ForProvider.DistributionConfig.CallerReference) != "" {
		cdi.DistributionConfig.CallerReference = cr.Spec.ForProvider.DistributionConfig.CallerReference
	} else {
		cdi.DistributionConfig.CallerReference = pointer.ToOrNilIfZeroValue(string(cr.UID))
	}

	// if cr.Spec.ForProvider.DistributionConfig.Origins is not nil then cdi.DistributionConfig.Origins is not nil
	if cr.Spec.ForProvider.DistributionConfig.Origins != nil {
		cdi.DistributionConfig.Origins.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))
	}

	if cr.Spec.ForProvider.DistributionConfig.Aliases != nil {
		cdi.DistributionConfig.Aliases.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Aliases.Items))
	}

	if cr.Spec.ForProvider.DistributionConfig.CustomErrorResponses != nil {
		cdi.DistributionConfig.CustomErrorResponses.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.CustomErrorResponses.Items))
	}

	if cr.Spec.ForProvider.DistributionConfig.Restrictions != nil && cr.Spec.ForProvider.DistributionConfig.Restrictions.GeoRestriction != nil {
		cdi.DistributionConfig.Restrictions.GeoRestriction.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Restrictions.GeoRestriction.Items))
	}

	dcb := cr.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior
	if dcb != nil {
		if dcb.AllowedMethods != nil {
			cdi.DistributionConfig.DefaultCacheBehavior.AllowedMethods.Quantity =
				pointer.ToIntAsInt64(len(dcb.AllowedMethods.Items))

			if dcb.AllowedMethods != nil && dcb.AllowedMethods.CachedMethods != nil {
				cdi.DistributionConfig.DefaultCacheBehavior.AllowedMethods.CachedMethods.Quantity =
					pointer.ToIntAsInt64(len(dcb.AllowedMethods.CachedMethods.Items))
			}
		}

		if dcb.ForwardedValues != nil {
			if dcb.ForwardedValues.Cookies != nil && dcb.ForwardedValues.Cookies.WhitelistedNames != nil {
				cdi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.Cookies.WhitelistedNames.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.Cookies.WhitelistedNames.Items))
			}

			if dcb.ForwardedValues.Headers != nil {
				cdi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.Headers.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.Headers.Items))
			}

			if dcb.ForwardedValues.QueryStringCacheKeys != nil {
				cdi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.QueryStringCacheKeys.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.QueryStringCacheKeys.Items))
			}
		}

		if dcb.FunctionAssociations != nil {
			cdi.DistributionConfig.DefaultCacheBehavior.FunctionAssociations.Quantity =
				pointer.ToIntAsInt64(len(dcb.FunctionAssociations.Items))
		}

		if dcb.LambdaFunctionAssociations != nil {
			cdi.DistributionConfig.DefaultCacheBehavior.LambdaFunctionAssociations.Quantity =
				pointer.ToIntAsInt64(len(dcb.LambdaFunctionAssociations.Items))
		}

		if dcb.TrustedKeyGroups != nil {
			cdi.DistributionConfig.DefaultCacheBehavior.TrustedKeyGroups.Quantity =
				pointer.ToIntAsInt64(len(dcb.TrustedKeyGroups.Items))
		}

		if dcb.TrustedSigners != nil {
			cdi.DistributionConfig.DefaultCacheBehavior.TrustedSigners.Quantity =
				pointer.ToIntAsInt64(len(dcb.TrustedSigners.Items))
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.CacheBehaviors != nil {
		cdi.DistributionConfig.CacheBehaviors.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.CacheBehaviors.Items))

		for i, cbi := range cr.Spec.ForProvider.DistributionConfig.CacheBehaviors.Items {
			if cbi.AllowedMethods != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].AllowedMethods.Quantity =
					pointer.ToIntAsInt64(len(cbi.AllowedMethods.Items))
			}

			if cbi.AllowedMethods != nil && cbi.AllowedMethods.CachedMethods != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].AllowedMethods.CachedMethods.Quantity =
					pointer.ToIntAsInt64(len(cbi.AllowedMethods.CachedMethods.Items))
			}

			if cbi.ForwardedValues != nil {
				if cbi.ForwardedValues.Cookies != nil && cbi.ForwardedValues.Cookies.WhitelistedNames != nil {
					cdi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.Cookies.WhitelistedNames.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.Cookies.WhitelistedNames.Items))
				}

				if cbi.ForwardedValues.Headers != nil {
					cdi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.Headers.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.Headers.Items))
				}

				if cbi.ForwardedValues.QueryStringCacheKeys != nil {
					cdi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.QueryStringCacheKeys.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.QueryStringCacheKeys.Items))
				}
			}

			if cbi.FunctionAssociations != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].FunctionAssociations.Quantity =
					pointer.ToIntAsInt64(len(cbi.FunctionAssociations.Items))
			}

			if cbi.LambdaFunctionAssociations != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].LambdaFunctionAssociations.Quantity =
					pointer.ToIntAsInt64(len(cbi.LambdaFunctionAssociations.Items))
			}

			if cbi.TrustedKeyGroups != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].TrustedKeyGroups.Quantity =
					pointer.ToIntAsInt64(len(cbi.TrustedKeyGroups.Items))
			}

			if cbi.TrustedSigners != nil {
				cdi.DistributionConfig.CacheBehaviors.Items[i].TrustedSigners.Quantity =
					pointer.ToIntAsInt64(len(cbi.TrustedSigners.Items))
			}
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.OriginGroups != nil {
		cdi.DistributionConfig.OriginGroups.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.OriginGroups.Items))

		for i, ogi := range cr.Spec.ForProvider.DistributionConfig.OriginGroups.Items {
			if ogi.FailoverCriteria != nil && ogi.FailoverCriteria.StatusCodes != nil {
				cdi.DistributionConfig.OriginGroups.Items[i].FailoverCriteria.StatusCodes.Quantity =
					pointer.ToIntAsInt64(len(ogi.FailoverCriteria.StatusCodes.Items))
			}

			if ogi.Members != nil {
				cdi.DistributionConfig.OriginGroups.Items[i].Members.Quantity = pointer.ToIntAsInt64(len(ogi.Members.Items))
			}
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.Origins != nil {
		cdi.DistributionConfig.Origins.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))

		for i, io := range cr.Spec.ForProvider.DistributionConfig.Origins.Items {
			if io.CustomHeaders != nil {
				cdi.DistributionConfig.Origins.Items[i].CustomHeaders.Quantity =
					pointer.ToIntAsInt64(len(io.CustomHeaders.Items))
			}

			if io.CustomOriginConfig != nil && io.CustomOriginConfig.OriginSSLProtocols != nil {
				cdi.DistributionConfig.Origins.Items[i].CustomOriginConfig.OriginSslProtocols.Quantity =
					pointer.ToIntAsInt64(len(io.CustomOriginConfig.OriginSSLProtocols.Items))
			}
		}
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Distribution, cdo *svcsdk.CreateDistributionOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(cdo.Distribution.Id))
	return ec, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Distribution, gdi *svcsdk.GetDistributionInput) error {
	gdi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Distribution, gdo *svcsdk.GetDistributionOutput) (bool, string, error) {
	// We can only update a Distribution that's in state 'Deployed' so we
	// temporarily consider it 'up to date' until it is since updating it
	// wouldn't work.
	if pointer.StringValue(cr.Status.AtProvider.Distribution.Status) != stateDeployed {
		return true, "", nil
	}

	// NOTE(negz): As far as I can tell we can't use the typical CreatePatch
	// pattern, because this type has a bunch of nested, updatable fields.
	// It's not possible to cmpopts.IgnoreField a specific 'leaf' field
	// because cmp still considers the parent field being non-nil in the
	// patch to mean there's a diff, and we obviously don't want to ignore
	// the entire parent field because then we'd never be able to detect
	// when an update was needed.

	currentParams := &svcapitypes.DistributionParameters{}
	_ = lateInitialize(currentParams, gdo)

	diff := cmp.Diff(*currentParams, cr.Spec.ForProvider,
		// We don't late init region - it's not in the output.
		cmpopts.IgnoreFields(svcapitypes.DistributionParameters{}, "Region"),

		// This appears to always be nil in GetDistributionOutput, which
		// causes false positives for IsUpToDate.
		cmpopts.IgnoreFields(svcapitypes.ViewerCertificate{}, "CloudFrontDefaultCertificate"),

		// There's quite a few slices of *string and *int64 in this API
		// that we want to consider equal regardless of order.
		cmpopts.SortSlices(func(x, y *string) bool { return pointer.StringValue(x) > pointer.StringValue(y) }),
		cmpopts.SortSlices(func(x, y *int64) bool { return pointer.Int64Value(x) > pointer.Int64Value(y) }),

		// TODO(negz): Do we need to do something like this for all the
		// other 'Items' slices with struct elements in this API? I've
		// observed that the API doesn't return Origins.Items in the
		// same order it's supplied (at a glance it seems to be returned
		// ordered lexicographically by ID).
		cmpopts.SortSlices(func(x, y *svcapitypes.Origin) bool {
			return pointer.StringValue(x.ID) > pointer.StringValue(y.ID)
		}),
	)
	return diff == "", diff, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Distribution, gdo *svcsdk.GetDistributionOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if pointer.StringValue(gdo.Distribution.Status) == stateDeployed &&
		pointer.BoolValue(gdo.Distribution.DistributionConfig.Enabled) {
		cr.SetConditions(xpv1.Available())
	} else {
		cr.SetConditions(xpv1.Unavailable())
	}
	return eo, nil
}

func postUpdate(_ context.Context, cr *svcapitypes.Distribution, resp *svcsdk.UpdateDistributionOutput,
	upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	// We need etag of update operation for the next operations.
	cr.Status.AtProvider.ETag = resp.ETag
	return upd, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Distribution, udi *svcsdk.UpdateDistributionInput) error {
	udi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	udi.SetIfMatch(pointer.StringValue(cr.Status.AtProvider.ETag))
	udi.DistributionConfig.Origins.Quantity =
		pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))

	if cr.Spec.ForProvider.DistributionConfig.Aliases != nil {
		udi.DistributionConfig.Aliases.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Aliases.Items))
	}

	if cr.Spec.ForProvider.DistributionConfig.CustomErrorResponses != nil {
		udi.DistributionConfig.CustomErrorResponses.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.CustomErrorResponses.Items))
	}

	if cr.Spec.ForProvider.DistributionConfig.Restrictions != nil && cr.Spec.ForProvider.DistributionConfig.Restrictions.GeoRestriction != nil {
		udi.DistributionConfig.Restrictions.GeoRestriction.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Restrictions.GeoRestriction.Items))
	}

	dcb := cr.Spec.ForProvider.DistributionConfig.DefaultCacheBehavior
	if dcb != nil {
		if dcb.AllowedMethods != nil {
			udi.DistributionConfig.DefaultCacheBehavior.AllowedMethods.Quantity =
				pointer.ToIntAsInt64(len(dcb.AllowedMethods.Items))

			if dcb.AllowedMethods != nil && dcb.AllowedMethods.CachedMethods != nil {
				udi.DistributionConfig.DefaultCacheBehavior.AllowedMethods.CachedMethods.Quantity =
					pointer.ToIntAsInt64(len(dcb.AllowedMethods.CachedMethods.Items))
			}
		}

		if dcb.ForwardedValues != nil {
			if dcb.ForwardedValues.Cookies != nil && dcb.ForwardedValues.Cookies.WhitelistedNames != nil {
				udi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.Cookies.WhitelistedNames.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.Cookies.WhitelistedNames.Items))
			}

			if dcb.ForwardedValues.Headers != nil {
				udi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.Headers.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.Headers.Items))
			}

			if dcb.ForwardedValues.QueryStringCacheKeys != nil {
				udi.DistributionConfig.DefaultCacheBehavior.ForwardedValues.QueryStringCacheKeys.Quantity =
					pointer.ToIntAsInt64(len(dcb.ForwardedValues.QueryStringCacheKeys.Items))
			}
		}
		if dcb.FunctionAssociations != nil {
			udi.DistributionConfig.DefaultCacheBehavior.FunctionAssociations.Quantity =
				pointer.ToIntAsInt64(len(dcb.FunctionAssociations.Items))
		}

		if dcb.LambdaFunctionAssociations != nil {
			udi.DistributionConfig.DefaultCacheBehavior.LambdaFunctionAssociations.Quantity =
				pointer.ToIntAsInt64(len(dcb.LambdaFunctionAssociations.Items))
		}

		if dcb.TrustedKeyGroups != nil {
			udi.DistributionConfig.DefaultCacheBehavior.TrustedKeyGroups.Quantity =
				pointer.ToIntAsInt64(len(dcb.TrustedKeyGroups.Items))
		}

		if dcb.TrustedSigners != nil {
			udi.DistributionConfig.DefaultCacheBehavior.TrustedSigners.Quantity =
				pointer.ToIntAsInt64(len(dcb.TrustedSigners.Items))
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.CacheBehaviors != nil {
		udi.DistributionConfig.CacheBehaviors.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.CacheBehaviors.Items))

		for i, cbi := range cr.Spec.ForProvider.DistributionConfig.CacheBehaviors.Items {
			if cbi.AllowedMethods != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].AllowedMethods.Quantity =
					pointer.ToIntAsInt64(len(cbi.AllowedMethods.Items))
			}

			if cbi.AllowedMethods != nil && cbi.AllowedMethods.CachedMethods != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].AllowedMethods.CachedMethods.Quantity =
					pointer.ToIntAsInt64(len(cbi.AllowedMethods.CachedMethods.Items))
			}

			if cbi.ForwardedValues != nil {
				if cbi.ForwardedValues.Cookies != nil && cbi.ForwardedValues.Cookies.WhitelistedNames != nil {
					udi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.Cookies.WhitelistedNames.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.Cookies.WhitelistedNames.Items))
				}

				if cbi.ForwardedValues.Headers != nil {
					udi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.Headers.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.Headers.Items))
				}

				if cbi.ForwardedValues.QueryStringCacheKeys != nil {
					udi.DistributionConfig.CacheBehaviors.Items[i].ForwardedValues.QueryStringCacheKeys.Quantity =
						pointer.ToIntAsInt64(len(cbi.ForwardedValues.QueryStringCacheKeys.Items))
				}
			}

			if cbi.FunctionAssociations != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].FunctionAssociations.Quantity =
					pointer.ToIntAsInt64(len(cbi.FunctionAssociations.Items))
			}

			if cbi.LambdaFunctionAssociations != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].LambdaFunctionAssociations.Quantity =
					pointer.ToIntAsInt64(len(cbi.LambdaFunctionAssociations.Items))
			}

			if cbi.TrustedKeyGroups != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].TrustedKeyGroups.Quantity =
					pointer.ToIntAsInt64(len(cbi.TrustedKeyGroups.Items))
			}

			if cbi.TrustedSigners != nil {
				udi.DistributionConfig.CacheBehaviors.Items[i].TrustedSigners.Quantity =
					pointer.ToIntAsInt64(len(cbi.TrustedSigners.Items))
			}
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.OriginGroups != nil {
		udi.DistributionConfig.OriginGroups.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.OriginGroups.Items))

		for i, ogi := range cr.Spec.ForProvider.DistributionConfig.OriginGroups.Items {
			if ogi.FailoverCriteria != nil && ogi.FailoverCriteria.StatusCodes != nil {
				udi.DistributionConfig.OriginGroups.Items[i].FailoverCriteria.StatusCodes.Quantity =
					pointer.ToIntAsInt64(len(ogi.FailoverCriteria.StatusCodes.Items))
			}

			if ogi.Members != nil {
				udi.DistributionConfig.OriginGroups.Items[i].Members.Quantity = pointer.ToIntAsInt64(len(ogi.Members.Items))
			}
		}
	}

	if cr.Spec.ForProvider.DistributionConfig.Origins != nil {
		udi.DistributionConfig.Origins.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))

		for i, io := range cr.Spec.ForProvider.DistributionConfig.Origins.Items {
			if io.CustomHeaders != nil {
				udi.DistributionConfig.Origins.Items[i].CustomHeaders.Quantity =
					pointer.ToIntAsInt64(len(io.CustomHeaders.Items))
			}
			if io.CustomOriginConfig != nil && io.CustomOriginConfig.OriginSSLProtocols != nil {
				udi.DistributionConfig.Origins.Items[i].CustomOriginConfig.OriginSslProtocols.Quantity =
					pointer.ToIntAsInt64(len(io.CustomOriginConfig.OriginSSLProtocols.Items))
			}
		}
	}

	return nil
}

type deleter struct {
	external *external
}

func (d *deleter) preDelete(ctx context.Context, cr *svcapitypes.Distribution, ddi *svcsdk.DeleteDistributionInput) (bool, error) {
	// In all cases, it needs to be "Deployed" to issue any update or delete requests.
	if pointer.StringValue(cr.Status.AtProvider.Distribution.Status) != stateDeployed {
		return true, nil
	}
	// If the distribution is enabled, it needs to be disabled before deletion.
	if pointer.BoolValue(cr.Status.AtProvider.Distribution.DistributionConfig.Enabled) {
		// We don't make the update call before user disables it because any update
		// (even no-op) takes ~5min.
		if pointer.BoolValue(cr.Spec.ForProvider.DistributionConfig.Enabled) {
			return false, errors.New("distribution needs to be disabled before deletion")
		}
		if _, err := d.external.Update(ctx, cr); err != nil {
			return false, errorutils.Wrap(err, errUpdate)
		}
	}
	ddi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	ddi.SetIfMatch(pointer.StringValue(cr.Status.AtProvider.ETag))
	return false, nil
}
