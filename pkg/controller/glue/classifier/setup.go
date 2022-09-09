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

package classifier

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupClassifier adds a controller that reconciles Classifier.
func SetupClassifier(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ClassifierGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Classifier{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ClassifierGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preDelete(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.DeleteClassifierInput) (bool, error) {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.GetClassifierInput) error {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.GetClassifierOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if obj.Classifier.CsvClassifier != nil {
		cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Classifier.CsvClassifier.CreationTime)
		cr.Status.AtProvider.Version = obj.Classifier.CsvClassifier.Version
		cr.Status.AtProvider.LastUpdated = fromTimePtr(obj.Classifier.CsvClassifier.LastUpdated)
	}
	if obj.Classifier.XMLClassifier != nil {
		cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Classifier.XMLClassifier.CreationTime)
		cr.Status.AtProvider.Version = obj.Classifier.XMLClassifier.Version
		cr.Status.AtProvider.LastUpdated = fromTimePtr(obj.Classifier.XMLClassifier.LastUpdated)
	}
	if obj.Classifier.GrokClassifier != nil {
		cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Classifier.GrokClassifier.CreationTime)
		cr.Status.AtProvider.Version = obj.Classifier.GrokClassifier.Version
		cr.Status.AtProvider.LastUpdated = fromTimePtr(obj.Classifier.GrokClassifier.LastUpdated)
	}
	if obj.Classifier.JsonClassifier != nil {
		cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Classifier.JsonClassifier.CreationTime)
		cr.Status.AtProvider.Version = obj.Classifier.JsonClassifier.Version
		cr.Status.AtProvider.LastUpdated = fromTimePtr(obj.Classifier.JsonClassifier.LastUpdated)
	}

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

// no lateInitialize bc AWS seems to not provide any default values through the API
// (even if the docs and the AWS Console hint so)
// potential defaults e.g.  csvClassifier.DisableValueTrimming <-> true
// csvClassifier.QuoteSymbol <-> Double-Quote (") | csvClassifier.Delimiter <-> Comma (,)

func isUpToDate(cr *svcapitypes.Classifier, resp *svcsdk.GetClassifierOutput) (bool, error) {

	currentParams := customGenerateClassifier(resp).Spec.ForProvider

	if cr.Spec.ForProvider.CustomGrokClassifier != nil {

		if awsclients.StringValue(cr.Spec.ForProvider.CustomGrokClassifier.CustomPatterns) !=
			awsclients.StringValue(resp.Classifier.GrokClassifier.CustomPatterns) {

			return false, nil
		}
	}

	return cmp.Equal(cr.Spec.ForProvider, currentParams, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(svcapitypes.ClassifierParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.CustomCreateGrokClassifierRequest{}, "CustomPatterns")), nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.UpdateClassifierInput) error {

	if cr.Spec.ForProvider.CustomCSVClassifier != nil {
		obj.CsvClassifier = &svcsdk.UpdateCsvClassifierRequest{
			Name:                 awsclients.String(meta.GetExternalName(cr)),
			AllowSingleColumn:    cr.Spec.ForProvider.CustomCSVClassifier.AllowSingleColumn,
			ContainsHeader:       cr.Spec.ForProvider.CustomCSVClassifier.ContainsHeader,
			Delimiter:            cr.Spec.ForProvider.CustomCSVClassifier.Delimiter,
			DisableValueTrimming: cr.Spec.ForProvider.CustomCSVClassifier.DisableValueTrimming,
			Header:               cr.Spec.ForProvider.CustomCSVClassifier.Header,
			QuoteSymbol:          cr.Spec.ForProvider.CustomCSVClassifier.QuoteSymbol,
		}
	}

	if cr.Spec.ForProvider.CustomXMLClassifier != nil {

		obj.XMLClassifier = &svcsdk.UpdateXMLClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: awsclients.String(cr.Spec.ForProvider.CustomXMLClassifier.Classification),
			RowTag:         cr.Spec.ForProvider.CustomXMLClassifier.RowTag,
		}
	}

	if cr.Spec.ForProvider.CustomGrokClassifier != nil {

		obj.GrokClassifier = &svcsdk.UpdateGrokClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: awsclients.String(cr.Spec.ForProvider.CustomGrokClassifier.Classification),
			CustomPatterns: cr.Spec.ForProvider.CustomGrokClassifier.CustomPatterns,
			GrokPattern:    awsclients.String(cr.Spec.ForProvider.CustomGrokClassifier.GrokPattern),
		}
		// if CustomPatterns was not nil before but is changed to nil through update, AWS just keeps the old value... (see on AWS Console)
		// however if we fill the spec field with "", AWS sets it to nil/empty
		if cr.Spec.ForProvider.CustomGrokClassifier.CustomPatterns == nil {
			s := ""
			obj.GrokClassifier.CustomPatterns = &s
		}
	}

	if cr.Spec.ForProvider.CustomJSONClassifier != nil {
		obj.JsonClassifier = &svcsdk.UpdateJsonClassifierRequest{
			Name:     awsclients.String(meta.GetExternalName(cr)),
			JsonPath: cr.Spec.ForProvider.CustomJSONClassifier.JSONPath,
		}
	}

	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.CreateClassifierInput) error {

	if cr.Spec.ForProvider.CustomCSVClassifier != nil {
		obj.CsvClassifier = &svcsdk.CreateCsvClassifierRequest{
			Name:                 awsclients.String(meta.GetExternalName(cr)),
			AllowSingleColumn:    cr.Spec.ForProvider.CustomCSVClassifier.AllowSingleColumn,
			ContainsHeader:       cr.Spec.ForProvider.CustomCSVClassifier.ContainsHeader,
			Delimiter:            cr.Spec.ForProvider.CustomCSVClassifier.Delimiter,
			DisableValueTrimming: cr.Spec.ForProvider.CustomCSVClassifier.DisableValueTrimming,
			Header:               cr.Spec.ForProvider.CustomCSVClassifier.Header,
			QuoteSymbol:          cr.Spec.ForProvider.CustomCSVClassifier.QuoteSymbol,
		}
	}

	if cr.Spec.ForProvider.CustomXMLClassifier != nil {
		obj.XMLClassifier = &svcsdk.CreateXMLClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: awsclients.String(cr.Spec.ForProvider.CustomXMLClassifier.Classification),
			RowTag:         cr.Spec.ForProvider.CustomXMLClassifier.RowTag,
		}
	}

	if cr.Spec.ForProvider.CustomGrokClassifier != nil {
		obj.GrokClassifier = &svcsdk.CreateGrokClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: awsclients.String(cr.Spec.ForProvider.CustomGrokClassifier.Classification),
			CustomPatterns: cr.Spec.ForProvider.CustomGrokClassifier.CustomPatterns,
			GrokPattern:    awsclients.String(cr.Spec.ForProvider.CustomGrokClassifier.GrokPattern),
		}
	}

	if cr.Spec.ForProvider.CustomJSONClassifier != nil {
		obj.JsonClassifier = &svcsdk.CreateJsonClassifierRequest{
			Name:     awsclients.String(meta.GetExternalName(cr)),
			JsonPath: cr.Spec.ForProvider.CustomJSONClassifier.JSONPath,
		}
	}

	return nil
}

// Custom GenerateClassifier for isuptodate (the generated one in zz_conversion.go is empty)
func customGenerateClassifier(resp *svcsdk.GetClassifierOutput) *svcapitypes.Classifier {

	cr := &svcapitypes.Classifier{}

	if resp.Classifier.CsvClassifier != nil {
		cr.Spec.ForProvider.CustomCSVClassifier = &svcapitypes.CustomCreateCSVClassifierRequest{
			AllowSingleColumn:    resp.Classifier.CsvClassifier.AllowSingleColumn,
			ContainsHeader:       resp.Classifier.CsvClassifier.ContainsHeader,
			Delimiter:            resp.Classifier.CsvClassifier.Delimiter,
			DisableValueTrimming: resp.Classifier.CsvClassifier.DisableValueTrimming,
			Header:               resp.Classifier.CsvClassifier.Header,
			QuoteSymbol:          resp.Classifier.CsvClassifier.QuoteSymbol,
		}
	}

	if resp.Classifier.XMLClassifier != nil {
		cr.Spec.ForProvider.CustomXMLClassifier = &svcapitypes.CustomCreateXMLClassifierRequest{
			Classification: awsclients.StringValue(resp.Classifier.XMLClassifier.Classification),
			RowTag:         resp.Classifier.XMLClassifier.RowTag,
		}
	}

	if resp.Classifier.GrokClassifier != nil {
		cr.Spec.ForProvider.CustomGrokClassifier = &svcapitypes.CustomCreateGrokClassifierRequest{
			Classification: awsclients.StringValue(resp.Classifier.GrokClassifier.Classification),
			CustomPatterns: resp.Classifier.GrokClassifier.CustomPatterns,
			GrokPattern:    awsclients.StringValue(resp.Classifier.GrokClassifier.GrokPattern),
		}
	}

	if resp.Classifier.JsonClassifier != nil {
		cr.Spec.ForProvider.CustomJSONClassifier = &svcapitypes.CustomCreateJSONClassifierRequest{
			JSONPath: resp.Classifier.JsonClassifier.JsonPath,
		}
	}

	return cr
}

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func fromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}
