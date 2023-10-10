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
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
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
		obj.ClientId = awsclients.String(meta.GetExternalName(cr))
	}
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.DeleteUserPoolClientInput) (bool, error) {
	obj.ClientId = awsclients.String(meta.GetExternalName(cr))
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

	meta.SetExternalName(cr, awsclients.StringValue(obj.UserPoolClient.ClientId))
	conn := managed.ConnectionDetails{
		"clientID":     []byte(awsclients.StringValue(cr.Status.AtProvider.ClientID)),
		"clientSecret": []byte(awsclients.StringValue(cr.Status.AtProvider.ClientSecret)),
		"userPoolID":   []byte(awsclients.StringValue(cr.Spec.ForProvider.UserPoolID)),
	}
	return managed.ExternalCreation{
		ExternalNameAssigned: true,
		ConnectionDetails:    conn,
	}, nil
}

func postUpdate(_ context.Context, cr *svcapitypes.UserPoolClient, obj *svcsdk.UpdateUserPoolClientOutput, obs managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	conn := managed.ConnectionDetails{
		"clientID":     []byte(awsclients.StringValue(cr.Status.AtProvider.ClientID)),
		"clientSecret": []byte(awsclients.StringValue(cr.Status.AtProvider.ClientSecret)),
		"userPoolID":   []byte(awsclients.StringValue(cr.Spec.ForProvider.UserPoolID)),
	}
	return managed.ExternalUpdate{
		ConnectionDetails: conn,
	}, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.UserPoolClient, resp *svcsdk.DescribeUserPoolClientOutput) (bool, string, error) {
	client := resp.UserPoolClient

	switch {
	case awsclients.Int64Value(cr.Spec.ForProvider.AccessTokenValidity) != awsclients.Int64Value(client.AccessTokenValidity),
		!reflect.DeepEqual(cr.Spec.ForProvider.AllowedOAuthFlows, client.AllowedOAuthFlows),
		awsclients.BoolValue(cr.Spec.ForProvider.AllowedOAuthFlowsUserPoolClient) != awsclients.BoolValue(client.AllowedOAuthFlowsUserPoolClient),
		!reflect.DeepEqual(cr.Spec.ForProvider.AllowedOAuthScopes, client.AllowedOAuthScopes),
		!areAnalyticsConfigurationEqual(cr.Spec.ForProvider.AnalyticsConfiguration, client.AnalyticsConfiguration),
		!reflect.DeepEqual(cr.Spec.ForProvider.CallbackURLs, client.CallbackURLs),
		awsclients.StringValue(cr.Spec.ForProvider.DefaultRedirectURI) != awsclients.StringValue(client.DefaultRedirectURI),
		!reflect.DeepEqual(cr.Spec.ForProvider.ExplicitAuthFlows, client.ExplicitAuthFlows),
		awsclients.Int64Value(cr.Spec.ForProvider.IDTokenValidity) != awsclients.Int64Value(client.IdTokenValidity),
		!reflect.DeepEqual(cr.Spec.ForProvider.LogoutURLs, client.LogoutURLs),
		awsclients.StringValue(cr.Spec.ForProvider.PreventUserExistenceErrors) != awsclients.StringValue(client.PreventUserExistenceErrors),
		!reflect.DeepEqual(cr.Spec.ForProvider.ReadAttributes, client.ReadAttributes),
		awsclients.Int64Value(cr.Spec.ForProvider.RefreshTokenValidity) != awsclients.Int64Value(client.RefreshTokenValidity),
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
		case awsclients.BoolValue(spec.UserDataShared) != awsclients.BoolValue(current.UserDataShared),
			awsclients.StringValue(spec.ApplicationARN) != awsclients.StringValue(current.ApplicationArn),
			awsclients.StringValue(spec.ApplicationID) != awsclients.StringValue(current.ApplicationId),
			awsclients.StringValue(spec.ExternalID) != awsclients.StringValue(current.ExternalId),
			awsclients.StringValue(spec.RoleARN) != awsclients.StringValue(current.RoleArn):
			return false
		}
	}
	return true
}

func areTokenValidityUnitsEqual(spec *svcapitypes.TokenValidityUnitsType, current *svcsdk.TokenValidityUnitsType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.AccessToken) != awsclients.StringValue(current.AccessToken),
			awsclients.StringValue(spec.IDToken) != awsclients.StringValue(current.IdToken),
			awsclients.StringValue(spec.RefreshToken) != awsclients.StringValue(current.RefreshToken):
			return false
		}
	}
	return true
}

func lateInitialize(cr *svcapitypes.UserPoolClientParameters, resp *svcsdk.DescribeUserPoolClientOutput) error {
	instance := resp.UserPoolClient

	cr.RefreshTokenValidity = awsclients.LateInitializeInt64Ptr(cr.RefreshTokenValidity, instance.RefreshTokenValidity)
	return nil
}
