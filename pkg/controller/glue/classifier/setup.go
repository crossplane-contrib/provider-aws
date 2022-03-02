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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/glue/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupClassifier adds a controller that reconciles Classifier.
func SetupClassifier(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ClassifierGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.preCreate = preCreate
		},
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
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.CreateClassifierOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, cr.Name)
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Classifier, obj *svcsdk.CreateClassifierInput) error {

	if cr.Spec.ForProvider.CustomCsvClassifier != nil {
		obj.CsvClassifier = &svcsdk.CreateCsvClassifierRequest{
			Name:                 awsclients.String(meta.GetExternalName(cr)),
			AllowSingleColumn:    cr.Spec.ForProvider.CustomCsvClassifier.AllowSingleColumn,
			ContainsHeader:       cr.Spec.ForProvider.CustomCsvClassifier.ContainsHeader,
			Delimiter:            cr.Spec.ForProvider.CustomCsvClassifier.Delimiter,
			DisableValueTrimming: cr.Spec.ForProvider.CustomCsvClassifier.DisableValueTrimming,
			Header:               cr.Spec.ForProvider.CustomCsvClassifier.Header,
			QuoteSymbol:          cr.Spec.ForProvider.CustomCsvClassifier.QuoteSymbol,
		}
	}

	if cr.Spec.ForProvider.CustomXMLClassifier != nil {
		obj.XMLClassifier = &svcsdk.CreateXMLClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: cr.Spec.ForProvider.CustomXMLClassifier.Classification,
			RowTag:         cr.Spec.ForProvider.CustomXMLClassifier.RowTag,
		}
	}

	if cr.Spec.ForProvider.CustomGrokClassifier != nil {
		obj.GrokClassifier = &svcsdk.CreateGrokClassifierRequest{
			Name:           awsclients.String(meta.GetExternalName(cr)),
			Classification: cr.Spec.ForProvider.CustomGrokClassifier.Classification,
			CustomPatterns: cr.Spec.ForProvider.CustomGrokClassifier.CustomPatterns,
			GrokPattern:    cr.Spec.ForProvider.CustomGrokClassifier.GrokPattern,
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
