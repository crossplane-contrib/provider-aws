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

package connection

import (
	"context"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/glue/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupConnection adds a controller that reconciles Connection.
func SetupConnection(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.ConnectionGroupKind)
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
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(limiter),
		}).
		For(&svcapitypes.Connection{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ConnectionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preDelete(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.DeleteConnectionInput) (bool, error) {
	obj.ConnectionName = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.GetConnectionInput) error {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.GetConnectionOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.CreateConnectionOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, cr.Name)
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.CreateConnectionInput) error {

	if cr.Spec.ForProvider.CustomConnectionInput != nil && cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements != nil {
		obj.ConnectionInput = &svcsdk.ConnectionInput{
			Name:                 awsclients.String(meta.GetExternalName(cr)),
			ConnectionProperties: cr.Spec.ForProvider.CustomConnectionInput.ConnectionProperties,
			ConnectionType:       cr.Spec.ForProvider.CustomConnectionInput.ConnectionType,
			Description:          cr.Spec.ForProvider.CustomConnectionInput.Description,
			MatchCriteria:        cr.Spec.ForProvider.CustomConnectionInput.MatchCriteria,
			PhysicalConnectionRequirements: &svcsdk.PhysicalConnectionRequirements{
				AvailabilityZone: cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.AvailabilityZone,
				SubnetId:         cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetID,
			},
		}

		for i := range cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList {
			obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList = append(obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList, &cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList[i])
		}
	}

	return nil
}
