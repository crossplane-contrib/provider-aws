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

package identityprovider

import (
	"context"
	"reflect"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupIdentityProvider adds a controller that reconciles IdentityProvider.
func SetupIdentityProvider(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.IdentityProviderGroupKind)

	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e, resolver: cognitoidentityprovider.NewResolver()}
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = c.preCreate
			e.preUpdate = c.preUpdate
			e.isUpToDate = c.isUpToDate
			e.lateInitialize = lateInitialize
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
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
		resource.ManagedKind(svcapitypes.IdentityProviderGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.IdentityProvider{}).
		Complete(r)
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.CognitoIdentityProviderAPI
	external *external
	resolver cognitoidentityprovider.ResolverService
}

func preObserve(_ context.Context, cr *svcapitypes.IdentityProvider, obj *svcsdk.DescribeIdentityProviderInput) error {
	obj.ProviderName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.IdentityProvider, obj *svcsdk.DeleteIdentityProviderInput) (bool, error) {
	obj.ProviderName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.IdentityProvider, obj *svcsdk.DescribeIdentityProviderOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.IdentityProvider, obj *svcsdk.CreateIdentityProviderInput) error {
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	obj.ProviderName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	providerDetails, err := e.resolver.GetProviderDetails(ctx, e.kube, &cr.Spec.ForProvider.ProviderDetailsSecretRef)
	if err != nil {
		return err
	}
	obj.SetProviderDetails(providerDetails)

	return nil
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.IdentityProvider, obj *svcsdk.UpdateIdentityProviderInput) error {
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	obj.ProviderName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	providerDetails, err := e.resolver.GetProviderDetails(ctx, e.kube, &cr.Spec.ForProvider.ProviderDetailsSecretRef)
	if err != nil {
		return err
	}
	obj.SetProviderDetails(providerDetails)

	return nil
}

func (e *custom) isUpToDate(ctx context.Context, cr *svcapitypes.IdentityProvider, resp *svcsdk.DescribeIdentityProviderOutput) (bool, string, error) {
	provider := resp.IdentityProvider
	p, err := e.resolver.GetProviderDetails(ctx, e.kube, &cr.Spec.ForProvider.ProviderDetailsSecretRef)
	if err != nil {
		return false, "", err
	}
	providerDetails := p

	switch {
	case !reflect.DeepEqual(cr.Spec.ForProvider.AttributeMapping, provider.AttributeMapping),
		cr.Spec.ForProvider.IDpIdentifiers != nil && !reflect.DeepEqual(cr.Spec.ForProvider.IDpIdentifiers, provider.IdpIdentifiers),
		!reflect.DeepEqual(providerDetails, provider.ProviderDetails):
		return false, "", nil
	}
	return true, "", nil
}

func lateInitialize(cr *svcapitypes.IdentityProviderParameters, current *svcsdk.DescribeIdentityProviderOutput) error {
	instance := current.IdentityProvider

	if cr.AttributeMapping == nil {
		cr.AttributeMapping = instance.AttributeMapping
	}
	return nil
}
