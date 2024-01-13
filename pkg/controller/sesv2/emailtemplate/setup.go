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

package emailtemplate

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sesv2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/sesv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupEmailTemplate adds a controller that reconciles SES EmailTemplate.
func SetupEmailTemplate(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.EmailTemplateGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.EmailTemplate{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.EmailTemplateGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func isUpToDate(_ context.Context, cr *svcapitypes.EmailTemplate, resp *svcsdk.GetEmailTemplateOutput) (bool, string, error) {
	if cr.Spec.ForProvider.TemplateContent != nil && resp.TemplateContent != nil {
		if pointer.StringValue(cr.Spec.ForProvider.TemplateContent.HTML) != pointer.StringValue(resp.TemplateContent.Html) {
			return false, "", nil
		}
		if pointer.StringValue(cr.Spec.ForProvider.TemplateContent.Subject) != pointer.StringValue(resp.TemplateContent.Subject) {
			return false, "", nil
		}
		if pointer.StringValue(cr.Spec.ForProvider.TemplateContent.Text) != pointer.StringValue(resp.TemplateContent.Text) {
			return false, "", nil
		}
	}
	return true, "", nil
}

func preObserve(_ context.Context, cr *svcapitypes.EmailTemplate, obj *svcsdk.GetEmailTemplateInput) error {
	obj.TemplateName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.EmailTemplate, resp *svcsdk.GetEmailTemplateOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.EmailTemplate, obj *svcsdk.CreateEmailTemplateInput) error {
	obj.TemplateName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.EmailTemplate, obj *svcsdk.UpdateEmailTemplateInput) error {
	obj.TemplateName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.EmailTemplate, obj *svcsdk.DeleteEmailTemplateInput) (bool, error) {
	obj.TemplateName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}
