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

package identitypool

import (
	"context"
	"reflect"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentity"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentity/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupIdentityPool adds a controller that reconciles IdentityPool.
func SetupIdentityPool(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.IdentityPoolGroupKind)

	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.preCreate = preCreate
			e.isUpToDate = isUpToDate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
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
		resource.ManagedKind(svcapitypes.IdentityPoolGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.IdentityPool{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.DescribeIdentityPoolInput) error {
	obj.IdentityPoolId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.IdentityPool, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.CreateIdentityPoolInput) error {
	obj.OpenIdConnectProviderARNs = cr.Spec.ForProvider.OpenIDConnectProviderARNs
	if cr.Spec.ForProvider.CognitoIdentityProviders != nil {
		providers := make([]*svcsdk.Provider, len(cr.Spec.ForProvider.CognitoIdentityProviders))
		for i, p := range cr.Spec.ForProvider.CognitoIdentityProviders {
			providers[i] = &svcsdk.Provider{
				ClientId:             p.ClientID,
				ProviderName:         p.ProviderName,
				ServerSideTokenCheck: p.ServerSideTokenCheck,
			}
		}
		obj.CognitoIdentityProviders = providers
	}
	obj.AllowUnauthenticatedIdentities = cr.Spec.ForProvider.AllowUnauthenticatedIdentities
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.IdentityPool, obs managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(obj.IdentityPoolId))
	return managed.ExternalCreation{}, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.IdentityPool) error {
	obj.IdentityPoolId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.OpenIdConnectProviderARNs = cr.Spec.ForProvider.OpenIDConnectProviderARNs
	if cr.Spec.ForProvider.CognitoIdentityProviders != nil {
		providers := make([]*svcsdk.Provider, len(cr.Spec.ForProvider.CognitoIdentityProviders))
		for i, p := range cr.Spec.ForProvider.CognitoIdentityProviders {
			providers[i] = &svcsdk.Provider{
				ClientId:             p.ClientID,
				ProviderName:         p.ProviderName,
				ServerSideTokenCheck: p.ServerSideTokenCheck,
			}
		}
		obj.CognitoIdentityProviders = providers
	}
	obj.AllowUnauthenticatedIdentities = cr.Spec.ForProvider.AllowUnauthenticatedIdentities
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.IdentityPool, obj *svcsdk.DeleteIdentityPoolInput) (bool, error) {
	obj.IdentityPoolId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(cr *svcapitypes.IdentityPool, resp *svcsdk.IdentityPool) (bool, error) {
	switch {
	case !reflect.DeepEqual(cr.Spec.ForProvider.AllowClassicFlow, resp.AllowClassicFlow),
		!reflect.DeepEqual(cr.Spec.ForProvider.AllowUnauthenticatedIdentities, resp.AllowUnauthenticatedIdentities),
		!areCognitoIdentityProvidersEqual(cr.Spec.ForProvider.CognitoIdentityProviders, resp.CognitoIdentityProviders),
		!reflect.DeepEqual(cr.Spec.ForProvider.DeveloperProviderName, resp.DeveloperProviderName),
		cr.Spec.ForProvider.IdentityPoolTags != nil && !reflect.DeepEqual(cr.Spec.ForProvider.IdentityPoolTags, resp.IdentityPoolTags),
		!reflect.DeepEqual(cr.Spec.ForProvider.OpenIDConnectProviderARNs, resp.OpenIdConnectProviderARNs),
		!reflect.DeepEqual(cr.Spec.ForProvider.SamlProviderARNs, resp.SamlProviderARNs),
		!reflect.DeepEqual(cr.Spec.ForProvider.SupportedLoginProviders, resp.SupportedLoginProviders):
		return false, nil
	}
	return true, nil
}

func areCognitoIdentityProvidersEqual(spec []*svcapitypes.Provider, current []*svcsdk.Provider) bool {
	if spec == nil && current == nil {
		return true
	}
	if spec != nil && current != nil {
		if len(spec) != len(current) {
			return false
		}

		for i, provider := range spec {
			switch {
			case !reflect.DeepEqual(provider.ClientID, current[i].ClientId),
				!reflect.DeepEqual(provider.ProviderName, current[i].ProviderName),
				!reflect.DeepEqual(provider.ServerSideTokenCheck, current[i].ServerSideTokenCheck):
				return false
			}
		}
		return true
	}
	return false
}
