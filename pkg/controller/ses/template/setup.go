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

package template

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/ses"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ses/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupTemplate adds a controller that reconciles Template.
func SetupTemplate(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.TemplateGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate

		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Template{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TemplateGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Template, obj *svcsdk.GetTemplateInput) error {
	obj.TemplateName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Template, obj *svcsdk.GetTemplateOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func preDelete(context context.Context, cr *svcapitypes.Template, obj *svcsdk.DeleteTemplateInput) (bool, error) {
	obj.TemplateName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postCreate(context context.Context, cr *svcapitypes.Template, obj *svcsdk.CreateTemplateOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(cr.Spec.ForProvider.Template.TemplateName))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func isUpToDate(cr *svcapitypes.Template, obj *svcsdk.GetTemplateOutput) (bool, error) {
	if !cmp.Equal(cr.Spec.ForProvider.Template.TemplateName, obj.Template.TemplateName) {
		return false, nil
	}
	if !cmp.Equal(cr.Spec.ForProvider.Template.HTMLPart, obj.Template.HtmlPart) {
		return false, nil
	}
	if !cmp.Equal(cr.Spec.ForProvider.Template.SubjectPart, obj.Template.SubjectPart) {
		return false, nil
	}
	if !cmp.Equal(cr.Spec.ForProvider.Template.TextPart, obj.Template.TextPart) {
		return false, nil
	}
	return true, nil
}
