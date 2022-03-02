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

package configuration

import (
	"context"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/kafka"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/kafka/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupConfiguration adds a controller that reconciles Configuration.
func SetupConfiguration(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ConfigurationGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.postDelete = postDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Configuration{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ConfigurationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preCreate(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.CreateConfigurationInput) error {
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	serverProperties := strings.Join(cr.Spec.ForProvider.Properties, "\n")
	obj.ServerProperties = []byte(serverProperties)
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.CreateConfigurationOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.Arn))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DescribeConfigurationInput) error {
	obj.Arn = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DescribeConfigurationOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.State) {
	case string(svcapitypes.ConfigurationState_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.ConfigurationState_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}

	return obs, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DeleteConfigurationInput) (bool, error) {
	obj.Arn = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postDelete(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DeleteConfigurationOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeBadRequestException) {
			// skip: failed to delete Configuration: BadRequestException:
			// This operation is only valid for resources that are in one of
			// the following states :[ACTIVE, DELETE_FAILED]
			return nil
		}
		return err
	}
	return err
}
