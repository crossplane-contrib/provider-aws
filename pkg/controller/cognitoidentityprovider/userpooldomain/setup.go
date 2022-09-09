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

package userpooldomain

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupUserPoolDomain adds a controller that reconciles User.
func SetupUserPoolDomain(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserPoolDomainGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.UserPoolDomain{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserPoolDomainGroupVersionKind),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.UserPoolDomain, obj *svcsdk.DescribeUserPoolDomainInput) error {
	obj.Domain = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.UserPoolDomain, obj *svcsdk.DeleteUserPoolDomainInput) (bool, error) {
	obj.Domain = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.UserPoolDomain, obj *svcsdk.DescribeUserPoolDomainOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if obj.DomainDescription.UserPoolId == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.UserPoolDomain, obj *svcsdk.CreateUserPoolDomainInput) error {
	obj.Domain = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}
