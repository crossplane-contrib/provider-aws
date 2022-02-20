/*
Copyright 2020 The Crossplane Authors.

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

package instanceprofile

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	svcsdk "github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/iam/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupInstanceProfile adds a controller that reconciles InstanceProfile.
func SetupInstanceProfile(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.InstanceProfileGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.InstanceProfile{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.InstanceProfileGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.GetInstanceProfileInput) error {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.InstanceProfile, _ *svcsdk.GetInstanceProfileOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.CreateInstanceProfileInput) error {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.InstanceProfile, resp *svcsdk.CreateInstanceProfileOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	sess, err := session.NewSession()
	if err != nil {
		return cre, err
	}
	svc := iam.New(sess)
	input := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(cr.Name),
		RoleName:            cr.Spec.ForProvider.Role,
	}

	_, err = svc.AddRoleToInstanceProfile(input)
	return cre, err
}

func preDelete(_ context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.DeleteInstanceProfileInput) (bool, error) {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	sess, err := session.NewSession()
	if err != nil {
		return false, err
	}
	svc := iam.New(sess)
	input := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(cr.Name),
		RoleName:            cr.Spec.ForProvider.Role,
	}

	_, err = svc.RemoveRoleFromInstanceProfile(input)
	return false, err
}
