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

package distribution

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// TODO: isn't this defined as an API constant somewhere in aws-sdk-go? Generated zz_enums.go seems not to contain it either
const stateDeployed = "Deployed"

// SetupDistribution adds a controller that reconciles Distribution.
func SetupDistribution(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.DistributionGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Distribution{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DistributionGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
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
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preCreate(_ context.Context, cr *svcapitypes.Distribution, cdi *svcsdk.CreateDistributionInput) error {
	cdi.DistributionConfig.CallerReference = awsclients.String(string(cr.UID))
	// if cr.Spec.ForProvider.DistributionConfig.Origins is not nil then cdi.DistributionConfig.Origins is not nil
	if cr.Spec.ForProvider.DistributionConfig.Origins != nil {
		cdi.DistributionConfig.Origins.Quantity =
			awsclients.Int64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))
	}
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Distribution, cdo *svcsdk.CreateDistributionOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, awsclients.StringValue(cdo.Distribution.Id))
	ec.ExternalNameAssigned = true
	return ec, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Distribution, gdi *svcsdk.GetDistributionInput) error {
	gdi.Id = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func lateInitialize(in *svcapitypes.DistributionParameters, gdo *svcsdk.GetDistributionOutput) error {
	inConfig, respConfig := in.DistributionConfig, gdo.Distribution.DistributionConfig

	_, err := lateInitializeFromResponse("", inConfig, respConfig,
		replacer("ID", "Id"),
		replacer("ARN", "Arn"),
		mapReplacer(map[string]string{
			"HTTPVersion": "HttpVersion",
		}))
	return err
}

func isUpToDate(cr *svcapitypes.Distribution, gdo *svcsdk.GetDistributionOutput) (bool, error) {
	patch, err := createPatch(gdo, &cr.Spec.ForProvider)

	if err != nil {
		return false, err
	}

	return cmp.Equal(&svcapitypes.DistributionConfig{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{})), nil
}

func postObserve(_ context.Context, cr *svcapitypes.Distribution, gdo *svcsdk.GetDistributionOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if awsclients.StringValue(gdo.Distribution.Status) == stateDeployed &&
		awsclients.BoolValue(gdo.Distribution.DistributionConfig.Enabled) {
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
	udi.Id = awsclients.String(meta.GetExternalName(cr))
	udi.SetIfMatch(awsclients.StringValue(cr.Status.AtProvider.ETag))
	udi.DistributionConfig.CallerReference = awsclients.String(string(cr.UID))
	udi.DistributionConfig.Origins.Quantity =
		awsclients.Int64(len(cr.Spec.ForProvider.DistributionConfig.Origins.Items))
	return nil
}

type deleter struct {
	external *external
}

func (d *deleter) preDelete(ctx context.Context, cr *svcapitypes.Distribution, ddi *svcsdk.DeleteDistributionInput) (bool, error) {
	// In all cases, it needs to be "Deployed" to issue any update or delete requests.
	if awsclients.StringValue(cr.Status.AtProvider.Distribution.Status) != stateDeployed {
		return true, nil
	}
	// If the distribution is enabled, it needs to be disabled before deletion.
	if awsclients.BoolValue(cr.Status.AtProvider.Distribution.DistributionConfig.Enabled) {
		// We don't make the update call before user disables it because any update
		// (even no-op) takes ~5min.
		if awsclients.BoolValue(cr.Spec.ForProvider.DistributionConfig.Enabled) {
			return false, errors.New("distribution needs to be disabled before deletion")
		}
		if _, err := d.external.Update(ctx, cr); err != nil {
			return false, awsclients.Wrap(err, errUpdate)
		}
	}
	ddi.Id = awsclients.String(meta.GetExternalName(cr))
	ddi.SetIfMatch(awsclients.StringValue(cr.Status.AtProvider.ETag))
	return false, nil
}

func createPatch(actual *svcsdk.GetDistributionOutput,
	desired *svcapitypes.DistributionParameters) (*svcapitypes.DistributionConfig, error) {
	actualConfig := &svcapitypes.DistributionParameters{
		DistributionConfig: &svcapitypes.DistributionConfig{},
	}

	if err := lateInitialize(actualConfig, actual); err != nil {
		return nil, err
	}

	jsonPatch, err := awsclients.CreateJSONPatch(actualConfig.DistributionConfig, desired.DistributionConfig)

	if err != nil {
		return nil, err
	}

	patch := &svcapitypes.DistributionConfig{}

	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}
