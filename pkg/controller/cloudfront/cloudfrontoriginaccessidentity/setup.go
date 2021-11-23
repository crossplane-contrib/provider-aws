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

package cloudfrontoriginaccessidentity

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/controller/cloudfront"
)

// SetupCloudFrontOriginAccessIDentity adds a controller that reconciles CloudFrontOriginAccessIDentity .
func SetupCloudFrontOriginAccessIDentity(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.CloudFrontOriginAccessIDentityGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.CloudFrontOriginAccessIDentity{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.CloudFrontOriginAccessIDentityGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube: mgr.GetClient(),
				opts: []option{
					func(e *external) {
						e.preObserve = preObserve
						e.postObserve = postObserve
						e.postCreate = postCreate
						e.lateInitialize = lateInitialize
						e.preUpdate = preUpdate
						e.isUpToDate = isUpToDate
						e.preDelete = preDelete
					},
				},
			}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func postCreate(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIDentity, cpo *svcsdk.CreateCloudFrontOriginAccessIdentityOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cp, awsclients.StringValue(cpo.CloudFrontOriginAccessIdentity.Id))
	return ec, nil
}

func preObserve(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIDentity, gpi *svcsdk.GetCloudFrontOriginAccessIdentityInput) error {
	gpi.Id = awsclients.String(meta.GetExternalName(cp))
	return nil
}

func postObserve(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIDentity, _ *svcsdk.GetCloudFrontOriginAccessIdentityOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cp.SetConditions(xpv1.Available())
	return eo, nil
}

func preUpdate(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIDentity, upi *svcsdk.UpdateCloudFrontOriginAccessIdentityInput) error {
	upi.Id = awsclients.String(meta.GetExternalName(cp))
	upi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return nil
}

func preDelete(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIDentity, dpi *svcsdk.DeleteCloudFrontOriginAccessIdentityInput) (bool, error) {
	dpi.Id = awsclients.String(meta.GetExternalName(cp))
	dpi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return false, nil
}

var mappingOptions = []cloudfront.LateInitOption{cloudfront.Replacer("ID", "Id")}

func lateInitialize(in *svcapitypes.CloudFrontOriginAccessIDentityParameters, gpo *svcsdk.GetCloudFrontOriginAccessIdentityOutput) error {
	_, err := cloudfront.LateInitializeFromResponse("",
		in.CloudFrontOriginAccessIDentityConfig, gpo.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig, mappingOptions...)
	return err
}

func isUpToDate(cp *svcapitypes.CloudFrontOriginAccessIDentity, gpo *svcsdk.GetCloudFrontOriginAccessIdentityOutput) (bool, error) {
	return cloudfront.IsUpToDate(gpo.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig, cp.Spec.ForProvider.CloudFrontOriginAccessIDentityConfig,
		mappingOptions...)
}
