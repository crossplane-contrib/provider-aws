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

package user

import (
	"context"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/transfer/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupUser adds a controller that reconciles User.
func SetupUser(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.UserGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.User{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserInput) error {
	obj.UserName = awsclients.String(meta.GetExternalName(cr))
	obj.ServerId = cr.Spec.ForProvider.ServerID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DeleteUserInput) (bool, error) {
	obj.UserName = awsclients.String(meta.GetExternalName(cr))
	obj.ServerId = cr.Spec.ForProvider.ServerID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.UserName))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserInput) error {
	obj.ServerId = cr.Spec.ForProvider.ServerID
	obj.Role = cr.Spec.ForProvider.Role
	obj.UserName = &cr.Name
	return nil
}
