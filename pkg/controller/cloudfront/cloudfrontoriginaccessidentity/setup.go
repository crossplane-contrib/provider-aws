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
	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupCloudFrontOriginAccessIdentity adds a controller that reconciles CloudFrontOriginAccessIdentity .
func SetupCloudFrontOriginAccessIdentity(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.CloudFrontOriginAccessIdentityGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.CloudFrontOriginAccessIdentity{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.CloudFrontOriginAccessIdentityGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube: mgr.GetClient(),
				opts: []option{
					func(e *external) {
						e.preObserve = preObserve
						e.postObserve = postObserve
						e.preCreate = preCreate
						e.postCreate = postCreate
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

func preCreate(_ context.Context, cr *svcapitypes.CloudFrontOriginAccessIdentity, cdi *svcsdk.CreateCloudFrontOriginAccessIdentityInput) error {
	cdi.CloudFrontOriginAccessIdentityConfig.CallerReference = awsclients.String(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIdentity, cpo *svcsdk.CreateCloudFrontOriginAccessIdentityOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cp, awsclients.StringValue(cpo.CloudFrontOriginAccessIdentity.Id))
	return ec, nil
}

func preObserve(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIdentity, gpi *svcsdk.GetCloudFrontOriginAccessIdentityInput) error {
	gpi.Id = awsclients.String(meta.GetExternalName(cp))
	return nil
}

func postObserve(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIdentity, _ *svcsdk.GetCloudFrontOriginAccessIdentityOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cp.SetConditions(xpv1.Available())
	return eo, nil
}

func preUpdate(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIdentity, upi *svcsdk.UpdateCloudFrontOriginAccessIdentityInput) error {
	upi.CloudFrontOriginAccessIdentityConfig.CallerReference = awsclients.String(string(cp.UID))
	upi.Id = awsclients.String(meta.GetExternalName(cp))
	upi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return nil
}

func preDelete(_ context.Context, cp *svcapitypes.CloudFrontOriginAccessIdentity, dpi *svcsdk.DeleteCloudFrontOriginAccessIdentityInput) (bool, error) {
	dpi.Id = awsclients.String(meta.GetExternalName(cp))
	dpi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return false, nil
}

func isUpToDate(cp *svcapitypes.CloudFrontOriginAccessIdentity, gpo *svcsdk.GetCloudFrontOriginAccessIdentityOutput) (bool, error) {
	return cmp.Equal(cp.Spec.ForProvider.CloudFrontOriginAccessIdentityConfig.Comment, gpo.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig.Comment), nil
}
