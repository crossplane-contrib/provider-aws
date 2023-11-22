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

package userpoolclient

import (
	"context"
	"reflect"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupUserPoolClient adds a controller that reconciles UserPoolClient.
func SetupUserPoolClient(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserPoolClientGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.postUpdate = postUpdate
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
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
		resource.ManagedKind(svcapitypes.UserPoolClientGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.UserPoolClient{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.DescribeUserPoolClientInput) error {
	if meta.GetExternalName(cr) != "" {
		obj.ClientId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	}
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.DeleteUserPoolClientInput) (bool, error) {
	obj.ClientId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.DescribeUserPoolClientOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.CreateUserPoolClientInput) error {
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.CreateUserPoolClientOutput, obs managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(obj.UserPoolClient.ClientId))
	conn := managed.ConnectionDetails{
		"clientID":     []byte(pointer.StringValue(cr.Status.AtProvider.ClientID)),
		"clientSecret": []byte(pointer.StringValue(cr.Status.AtProvider.ClientSecret)),
		"userPoolID":   []byte(pointer.StringValue(cr.Spec.ForProvider.UserPoolID)),
	}
	return managed.ExternalCreation{
		ConnectionDetails: conn,
	}, nil
}

func postUpdate(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.UpdateUserPoolClientOutput, obs managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	conn := managed.ConnectionDetails{
		"clientID":     []byte(pointer.StringValue(cr.Status.AtProvider.ClientID)),
		"clientSecret": []byte(pointer.StringValue(cr.Status.AtProvider.ClientSecret)),
		"userPoolID":   []byte(pointer.StringValue(cr.Spec.ForProvider.UserPoolID)),
	}
	return managed.ExternalUpdate{
		ConnectionDetails: conn,
	}, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.UserPoolClient, resp *svcsdk.DescribeUserPoolClientOutput) (bool, string, error) {
	client := resp.UserPoolClient

	switch {
	case pointer.Int64Value(cr.Spec.ForProvider.AccessTokenValidity) != pointer.Int64Value(client.AccessTokenValidity),
		!reflect.DeepEqual(cr.Spec.ForProvider.AllowedOAuthFlows, client.AllowedOAuthFlows),
		pointer.BoolValue(cr.Spec.ForProvider.AllowedOAuthFlowsUserPoolClient) != pointer.BoolValue(client.AllowedOAuthFlowsUserPoolClient),
		!reflect.DeepEqual(cr.Spec.ForProvider.AllowedOAuthScopes, client.AllowedOAuthScopes),
		!areAnalyticsConfigurationEqual(cr.Spec.ForProvider.AnalyticsConfiguration, client.AnalyticsConfiguration),
		!reflect.DeepEqual(cr.Spec.ForProvider.CallbackURLs, client.CallbackURLs),
		pointer.StringValue(cr.Spec.ForProvider.DefaultRedirectURI) != pointer.StringValue(client.DefaultRedirectURI),
		!reflect.DeepEqual(cr.Spec.ForProvider.ExplicitAuthFlows, client.ExplicitAuthFlows),
		pointer.Int64Value(cr.Spec.ForProvider.IDTokenValidity) != pointer.Int64Value(client.IdTokenValidity),
		!reflect.DeepEqual(cr.Spec.ForProvider.LogoutURLs, client.LogoutURLs),
		pointer.StringValue(cr.Spec.ForProvider.PreventUserExistenceErrors) != pointer.StringValue(client.PreventUserExistenceErrors),
		!reflect.DeepEqual(cr.Spec.ForProvider.ReadAttributes, client.ReadAttributes),
		pointer.Int64Value(cr.Spec.ForProvider.RefreshTokenValidity) != pointer.Int64Value(client.RefreshTokenValidity),
		!reflect.DeepEqual(cr.Spec.ForProvider.SupportedIdentityProviders, client.SupportedIdentityProviders),
		!areTokenValidityUnitsEqual(cr.Spec.ForProvider.TokenValidityUnits, client.TokenValidityUnits),
		!reflect.DeepEqual(cr.Spec.ForProvider.WriteAttributes, client.WriteAttributes):
		return false, "", nil
	}
	return true, "", nil
}

func areAnalyticsConfigurationEqual(spec *svcapitypes.AnalyticsConfigurationType, current *svcsdk.AnalyticsConfigurationType) bool {
	if spec != nil && current != nil {
		switch {
		case pointer.BoolValue(spec.UserDataShared) != pointer.BoolValue(current.UserDataShared),
			pointer.StringValue(spec.ApplicationARN) != pointer.StringValue(current.ApplicationArn),
			pointer.StringValue(spec.ApplicationID) != pointer.StringValue(current.ApplicationId),
			pointer.StringValue(spec.ExternalID) != pointer.StringValue(current.ExternalId),
			pointer.StringValue(spec.RoleARN) != pointer.StringValue(current.RoleArn):
			return false
		}
	}
	return true
}

func areTokenValidityUnitsEqual(spec *svcapitypes.TokenValidityUnitsType, current *svcsdk.TokenValidityUnitsType) bool {
	if spec != nil && current != nil {
		switch {
		case pointer.StringValue(spec.AccessToken) != pointer.StringValue(current.AccessToken),
			pointer.StringValue(spec.IDToken) != pointer.StringValue(current.IdToken),
			pointer.StringValue(spec.RefreshToken) != pointer.StringValue(current.RefreshToken):
			return false
		}
	}
	return true
}

func lateInitialize(cr *svcapitypes.UserPoolClientParameters, resp *svcsdk.DescribeUserPoolClientOutput) error {
	instance := resp.UserPoolClient

	cr.RefreshTokenValidity = pointer.LateInitialize(cr.RefreshTokenValidity, instance.RefreshTokenValidity)
	return nil
}
