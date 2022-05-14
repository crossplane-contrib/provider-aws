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

package job

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/features"
)

// SetupJob adds a controller that reconciles Job.
func SetupJob(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.preObserve = preObserve
			e.postObserve = postObserve
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
		For(&svcapitypes.Job{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.JobGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preDelete(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.DeleteJobInput) (bool, error) {
	obj.JobName = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.GetJobInput) error {
	obj.JobName = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.GetJobOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.CreateJobOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.Name))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.CreateJobInput) error {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	obj.Role = awsclients.String(cr.Spec.ForProvider.RoleArn)
	return nil
}
