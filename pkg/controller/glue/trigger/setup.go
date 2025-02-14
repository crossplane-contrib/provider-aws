/*
Copyright 2024 The Crossplane Authors.

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

package trigger

import (
	"context"
	"encoding/json"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	svcsdkapi "github.com/aws/aws-sdk-go/service/glue/glueiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

type customConnector struct {
	kube client.Client
}

// customExternal is external connector with overridden Update method due to ACK doesn't correctly generate it.
type customExternal struct {
	external
}

func (c *customConnector) Connect(ctx context.Context, cr *svcapitypes.Trigger) (managed.TypedExternalClient[*svcapitypes.Trigger], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newCustomExternal(c.kube, svcsdk.New(sess)), nil
}

func newCustomExternal(kube client.Client, client svcsdkapi.GlueAPI) *customExternal {
	return &customExternal{
		external{
			kube:           kube,
			client:         client,
			preObserve:     preObserve,
			postObserve:    postObserve,
			isUpToDate:     isUpToDate,
			preCreate:      preCreate,
			preDelete:      preDelete,
			preUpdate:      preUpdate,
			lateInitialize: nopLateInitialize,
			postCreate:     nopPostCreate,
			postDelete:     nopPostDelete,
			postUpdate:     nopPostUpdate,
		},
	}
}

func (e *customExternal) Update(ctx context.Context, cr *svcapitypes.Trigger) (managed.ExternalUpdate, error) { //nolint:gocyclo
	triggerUpdate := &svcsdk.TriggerUpdate{}

	if cr.Spec.ForProvider.Predicate != nil {
		if cr.Spec.ForProvider.Predicate.Conditions != nil {
			triggerUpdate.Predicate.Conditions = make([]*svcsdk.Condition, 0, len(cr.Spec.ForProvider.Predicate.Conditions))

			for _, crCondition := range cr.Spec.ForProvider.Predicate.Conditions {
				inputCondition := &svcsdk.Condition{}
				if crCondition.CrawlState != nil {
					inputCondition.CrawlState = crCondition.CrawlState
				}
				if crCondition.CrawlerName != nil {
					inputCondition.CrawlerName = crCondition.CrawlerName
				}
				if crCondition.JobName != nil {
					inputCondition.JobName = crCondition.JobName
				}
				if crCondition.LogicalOperator != nil {
					inputCondition.LogicalOperator = crCondition.LogicalOperator
				}
				if crCondition.State != nil {
					inputCondition.State = crCondition.State
				}
				triggerUpdate.Predicate.Conditions = append(triggerUpdate.Predicate.Conditions, inputCondition)
			}
		}
		if cr.Spec.ForProvider.Predicate.Logical != nil {
			triggerUpdate.Predicate.Logical = cr.Spec.ForProvider.Predicate.Logical
		}
	}
	if cr.Spec.ForProvider.EventBatchingCondition != nil {
		if cr.Spec.ForProvider.EventBatchingCondition.BatchSize != nil {
			triggerUpdate.EventBatchingCondition.BatchSize = cr.Spec.ForProvider.EventBatchingCondition.BatchSize
		}
		if cr.Spec.ForProvider.EventBatchingCondition.BatchWindow != nil {
			triggerUpdate.EventBatchingCondition.BatchWindow = cr.Spec.ForProvider.EventBatchingCondition.BatchWindow
		}
	}

	if cr.Spec.ForProvider.Actions != nil {
		triggerUpdate.Actions = make([]*svcsdk.Action, 0, len(cr.Spec.ForProvider.Actions))
		for _, crAction := range cr.Spec.ForProvider.Actions {
			inputAction := &svcsdk.Action{}
			if crAction.NotificationProperty != nil && crAction.NotificationProperty.NotifyDelayAfter != nil {
				inputAction.NotificationProperty = &svcsdk.NotificationProperty{NotifyDelayAfter: crAction.NotificationProperty.NotifyDelayAfter}
			}
			if crAction.Arguments != nil {
				inputAction.Arguments = crAction.Arguments
			}
			if crAction.CrawlerName != nil {
				inputAction.CrawlerName = crAction.CrawlerName
			}
			if crAction.JobName != nil {
				inputAction.JobName = crAction.JobName
			}
			if crAction.SecurityConfiguration != nil {
				inputAction.SecurityConfiguration = crAction.SecurityConfiguration
			}
			if crAction.Timeout != nil {
				inputAction.Timeout = crAction.Timeout
			}
			triggerUpdate.Actions = append(triggerUpdate.Actions, inputAction)
		}
	}

	if cr.Spec.ForProvider.Schedule != nil {
		triggerUpdate.Schedule = cr.Spec.ForProvider.Schedule
	}

	if cr.Spec.ForProvider.Description != nil {
		triggerUpdate.Description = cr.Spec.ForProvider.Description
	}
	input := &svcsdk.UpdateTriggerInput{TriggerUpdate: triggerUpdate}
	if err := preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.UpdateTriggerWithContext(ctx, input)
	return nopPostUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}

// SetupTrigger adds a controller that reconciles Trigger.
func SetupTrigger(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.TriggerGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Trigger{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TriggerGroupVersionKind),
			managed.WithTypedExternalConnector(&customConnector{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preCreate(_ context.Context, cr *svcapitypes.Trigger, input *svcsdk.CreateTriggerInput) error {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Trigger, input *svcsdk.DeleteTriggerInput) (bool, error) {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	state := ptr.Deref(cr.Status.AtProvider.State, "")
	if state == svcsdk.TriggerStateActivating || state == svcsdk.TriggerStateDeactivating ||
		state == svcsdk.TriggerStateCreating || state == svcsdk.TriggerStateDeleting {
		return false, nil
	}
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Trigger, input *svcsdk.GetTriggerInput) error {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Trigger, resp *svcsdk.GetTriggerOutput) (upToDate bool, diff string, err error) {
	state := ptr.Deref(cr.Status.AtProvider.State, "")
	if state == svcsdk.TriggerStateActivating || state == svcsdk.TriggerStateDeactivating ||
		state == svcsdk.TriggerStateCreating || state == svcsdk.TriggerStateDeleting {
		return true, "", nil
	}
	patch, err := createPatch(&cr.Spec.ForProvider, resp)
	if err != nil {
		return false, "", err
	}
	diff = cmp.Diff(&svcapitypes.TriggerParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(svcapitypes.TriggerParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.TriggerParameters{}, "Tags"),
		cmpopts.IgnoreFields(svcapitypes.TriggerParameters{}, "StartOnCreation"),
		cmpopts.IgnoreFields(svcapitypes.TriggerParameters{}, "TriggerType"), // TriggerType is immutable
	)
	if diff != "" {
		return false, "Found observed difference in glue trigger " + diff, nil
	}
	return true, "", nil
}

func postObserve(_ context.Context, cr *svcapitypes.Trigger, resp *svcsdk.GetTriggerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.Status.AtProvider.ID = resp.Trigger.Id
	cr.Status.AtProvider.State = resp.Trigger.State
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Trigger, input *svcsdk.UpdateTriggerInput) error {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func createPatch(currentParams *svcapitypes.TriggerParameters, resp *svcsdk.GetTriggerOutput) (*svcapitypes.TriggerParameters, error) { //nolint:gocyclo
	targetConfig := currentParams.DeepCopy()
	externalConfig := &svcapitypes.TriggerParameters{}
	externalConfig.Schedule = resp.Trigger.Schedule
	if resp.Trigger.Predicate != nil {
		externalConfig.Predicate = &svcapitypes.Predicate{}
		if resp.Trigger.Predicate.Conditions != nil {
			externalConfig.Predicate = &svcapitypes.Predicate{Conditions: make([]*svcapitypes.Condition, 0, len(resp.Trigger.Predicate.Conditions))}
			for _, respCondition := range resp.Trigger.Predicate.Conditions {
				curCondition := &svcapitypes.Condition{}
				if respCondition.CrawlState != nil {
					curCondition.CrawlState = respCondition.CrawlState
				}
				if respCondition.CrawlerName != nil {
					curCondition.CrawlerName = respCondition.CrawlerName
				}
				if respCondition.JobName != nil {
					curCondition.JobName = respCondition.JobName
				}
				if respCondition.LogicalOperator != nil {
					curCondition.LogicalOperator = respCondition.LogicalOperator
				}
				if respCondition.State != nil {
					curCondition.State = respCondition.State
				}
				externalConfig.Predicate.Conditions = append(externalConfig.Predicate.Conditions, curCondition)
			}
		}
		if resp.Trigger.Predicate != nil && resp.Trigger.Predicate.Logical != nil {
			externalConfig.Predicate = &svcapitypes.Predicate{Logical: resp.Trigger.Predicate.Logical}
		}
	}
	if resp.Trigger.Actions != nil {
		externalConfig.Actions = make([]*svcapitypes.Action, 0, len(resp.Trigger.Actions))
		for _, respAction := range resp.Trigger.Actions {
			crAction := &svcapitypes.Action{}
			if respAction.NotificationProperty != nil && respAction.NotificationProperty.NotifyDelayAfter != nil {
				crAction.NotificationProperty = &svcapitypes.NotificationProperty{NotifyDelayAfter: respAction.NotificationProperty.NotifyDelayAfter}
			}
			if respAction.Arguments != nil {
				crAction.Arguments = respAction.Arguments
			}
			if respAction.CrawlerName != nil {
				crAction.CrawlerName = respAction.CrawlerName
			}
			if respAction.JobName != nil {
				crAction.JobName = respAction.JobName
			}
			if respAction.SecurityConfiguration != nil {
				crAction.SecurityConfiguration = respAction.SecurityConfiguration
			}
			if respAction.Timeout != nil {
				crAction.Timeout = respAction.Timeout
			}
			externalConfig.Actions = append(externalConfig.Actions, crAction)
		}
	}
	externalConfig.Description = resp.Trigger.Description
	eventBatchingCondition := &svcapitypes.EventBatchingCondition{}
	if resp.Trigger.EventBatchingCondition != nil {
		eventBatchingCondition.BatchSize = resp.Trigger.EventBatchingCondition.BatchSize
		eventBatchingCondition.BatchWindow = resp.Trigger.EventBatchingCondition.BatchWindow
	}
	externalConfig.EventBatchingCondition = eventBatchingCondition
	if resp.Trigger.Type != nil {
		externalConfig.TriggerType = resp.Trigger.Type
	}
	if resp.Trigger.WorkflowName != nil {
		externalConfig.WorkflowName = resp.Trigger.WorkflowName
	}

	jsonPatch, err := jsonpatch.CreateJSONPatch(externalConfig, targetConfig)
	if err != nil {
		return nil, err
	}
	patch := &svcapitypes.TriggerParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}
