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

package topic

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sns"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sns/v1alpha1"
)

// SetupTopic adds a controller that reconciles Topic.
func SetupTopic(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.TopicGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Topic{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TopicGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Topic) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.Topic, _ *svcsdk.GetTopicAttributesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.Topic) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.Topic, _ *svcsdk.CreateTopicOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Topic) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Topic, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.TopicParameters, *svcsdk.GetTopicAttributesOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.Topic, *svcsdk.GetTopicAttributesOutput) bool {
	return true
}

func preGenerateGetTopicAttributesInput(_ *svcapitypes.Topic, obj *svcsdk.GetTopicAttributesInput) *svcsdk.GetTopicAttributesInput {
	return obj
}

func postGenerateGetTopicAttributesInput(_ *svcapitypes.Topic, obj *svcsdk.GetTopicAttributesInput) *svcsdk.GetTopicAttributesInput {
	return obj
}

func preGenerateCreateTopicInput(_ *svcapitypes.Topic, obj *svcsdk.CreateTopicInput) *svcsdk.CreateTopicInput {
	return obj
}

func postGenerateCreateTopicInput(_ *svcapitypes.Topic, obj *svcsdk.CreateTopicInput) *svcsdk.CreateTopicInput {
	return obj
}
func preGenerateDeleteTopicInput(_ *svcapitypes.Topic, obj *svcsdk.DeleteTopicInput) *svcsdk.DeleteTopicInput {
	return obj
}

func postGenerateDeleteTopicInput(_ *svcapitypes.Topic, obj *svcsdk.DeleteTopicInput) *svcsdk.DeleteTopicInput {
	return obj
}
