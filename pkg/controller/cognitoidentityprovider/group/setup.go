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

package group

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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

// SetupGroup adds a controller that reconciles Group.
func SetupGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.GroupGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.isUpToDate = isUpToDate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Group{},
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.LabelChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			))).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.GroupGroupVersionKind),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.Group, obj *svcsdk.GetGroupInput) error {
	obj.GroupName = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Group, obj *svcsdk.DeleteGroupInput) (bool, error) {
	obj.GroupName = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Group, obj *svcsdk.GetGroupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Group, obj *svcsdk.CreateGroupInput) error {
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	obj.RoleArn = cr.Spec.ForProvider.RoleARN
	obj.GroupName = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Group, obj *svcsdk.UpdateGroupInput) error {
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	obj.RoleArn = cr.Spec.ForProvider.RoleARN
	obj.GroupName = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func isUpToDate(cr *svcapitypes.Group, resp *svcsdk.GetGroupOutput) (bool, error) {
	switch {
	case awsclients.StringValue(cr.Spec.ForProvider.Description) != awsclients.StringValue(resp.Group.Description),
		awsclients.Int64Value(cr.Spec.ForProvider.Precedence) != awsclients.Int64Value(resp.Group.Precedence):
		return false, nil
	}
	return true, nil
}
