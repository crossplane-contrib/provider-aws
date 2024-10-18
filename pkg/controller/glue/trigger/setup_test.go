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
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
)

type args struct {
	trigger          *svcapitypes.Trigger
	getTriggerOutput *svcsdk.GetTriggerOutput
}

func TestIsUpToDate(t *testing.T) {
	region := "eu-central-2"

	triggerName := "TriggerName"

	actionsJobName1 := "JobName1"
	actionsJobName2 := "jobName2"

	actionArgumentValue1 := "argumentValue"

	eventBatchingConditionBatchSize := int64(3)
	newEventBatchingConditionBatchSize := int64(4)

	eventBatchingConditionBatchWindow := int64(15)
	newEventBatchingConditionBatchWindow := int64(20)

	actionJobTimeout := int64(1)
	newActionJobTimeout := int64(2)

	description := "Trigger for Glue"
	newDescription := "New Description"

	triggerType := "SCHEDULED"
	newTriggerType := "ON_DEMAND"

	schedule := "cron(*/5 * * * ? *)"
	newSchedule := "cron(*/3 * * * ? *)"

	startOnCreation := true
	newStartOnCreation := false

	notificationPropertyNotifyDelayAfter := int64(1)
	newNotificationPropertyNotifyDelayAfter := int64(2)

	predicateConditionCrawlerStateReady := "READY"
	predicateConditionCrawlerStateRunning := "RUNNING"

	predicateConditionCrawlerName := "CrawlerName"
	newPredicateConditionCrawlerName := "NewCrawlerName"

	predicateConditionJobName := "JobName"
	newPredicateConditionJobName := "NewJobName"

	predicateConditionLogicalOperator := "EQUALS"

	predicateConditionStateSucceeded := "SUCCEEDED"
	predicateConditionStateStopped := "STOPPED"

	predicateLogical := "AND"
	newPredicateLogical := "ANY"

	workflowName := "WorkflowName"
	newWorkflowName := "NewWorkflowName"

	type want struct {
		result bool
		err    error
	}
	cases := map[string]struct {
		args
		want
	}{
		"NothingChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Region: region,
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName1,
									Timeout: &actionJobTimeout,
								},
							},
							Description:     &description,
							Schedule:        &schedule,
							StartOnCreation: &startOnCreation,
							TriggerType:     &triggerType,
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								Timeout: &actionJobTimeout,
							},
						},
						Description: &description,
						Schedule:    &schedule,
						Type:        &triggerType,
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ScheduleChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Schedule: &newSchedule,
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name:     &triggerName,
						Schedule: &schedule,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"DescriptionChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Description: &newDescription,
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name:        &triggerName,
						Description: &description,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ActionTimeoutChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName1,
									Timeout: &newActionJobTimeout,
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								Timeout: &actionJobTimeout,
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ActionArgumentsChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName1,
									Arguments: map[string]*string{
										"--foo": &actionArgumentValue1,
										"--bar": &actionArgumentValue1,
									},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								Arguments: map[string]*string{
									"--foo": &actionArgumentValue1,
								},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ActionNotificationPropertyNotifyDelayAfterChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName1,
									NotificationProperty: &svcapitypes.NotificationProperty{
										NotifyDelayAfter: &notificationPropertyNotifyDelayAfter,
									},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								NotificationProperty: &svcsdk.NotificationProperty{
									NotifyDelayAfter: &newNotificationPropertyNotifyDelayAfter,
								},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ActionAdded": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName1,
									Timeout: &actionJobTimeout,
								},
								{
									JobName: &actionsJobName2,
									Timeout: &newActionJobTimeout,
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								Timeout: &actionJobTimeout,
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ActionRemoved": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Actions: []*svcapitypes.Action{
								{
									JobName: &actionsJobName2,
									Timeout: &newActionJobTimeout,
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Actions: []*svcsdk.Action{
							{
								JobName: &actionsJobName1,
								Timeout: &actionJobTimeout,
							},
							{
								JobName: &actionsJobName2,
								Timeout: &newActionJobTimeout,
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"EventBatchingConditionBatchSizeChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							EventBatchingCondition: &svcapitypes.EventBatchingCondition{
								BatchWindow: &eventBatchingConditionBatchWindow,
								BatchSize:   &newEventBatchingConditionBatchSize,
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						EventBatchingCondition: &svcsdk.EventBatchingCondition{
							BatchWindow: &eventBatchingConditionBatchWindow,
							BatchSize:   &eventBatchingConditionBatchSize,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"EventBatchingConditionBatchWindowChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							EventBatchingCondition: &svcapitypes.EventBatchingCondition{
								BatchWindow: &newEventBatchingConditionBatchWindow,
								BatchSize:   &eventBatchingConditionBatchSize,
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						EventBatchingCondition: &svcsdk.EventBatchingCondition{
							BatchWindow: &eventBatchingConditionBatchWindow,
							BatchSize:   &eventBatchingConditionBatchSize,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PredicateConditionCrawlerStateChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Conditions: []*svcapitypes.Condition{
									{CrawlState: &predicateConditionCrawlerStateRunning},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Conditions: []*svcsdk.Condition{
								{CrawlState: &predicateConditionCrawlerStateReady},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PredicateConditionJobNameChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Conditions: []*svcapitypes.Condition{
									{JobName: &newPredicateConditionJobName},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Conditions: []*svcsdk.Condition{
								{JobName: &predicateConditionJobName},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PredicateConditionCrawlerNameChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Conditions: []*svcapitypes.Condition{
									{CrawlerName: &newPredicateConditionCrawlerName},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Conditions: []*svcsdk.Condition{
								{CrawlerName: &predicateConditionCrawlerName},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PredicateConditionLogicalOperatorAdded": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Conditions: []*svcapitypes.Condition{
									{LogicalOperator: &predicateConditionLogicalOperator},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Conditions: []*svcsdk.Condition{},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PredicateConditionStateChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Conditions: []*svcapitypes.Condition{
									{State: &predicateConditionStateSucceeded},
								},
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Conditions: []*svcsdk.Condition{
								{State: &predicateConditionStateStopped},
							},
						},
					},
				},
			},
		},
		"PredicateLogicalChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							Predicate: &svcapitypes.Predicate{
								Logical: &newPredicateLogical,
							},
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Predicate: &svcsdk.Predicate{
							Logical: &predicateLogical,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"WorkflowNameChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							WorkflowName: &newWorkflowName,
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						WorkflowName: &workflowName,
					},
				},
			},
		},
		"ImmutableParamsChanged": {
			args: args{
				trigger: &svcapitypes.Trigger{
					ObjectMeta: metav1.ObjectMeta{
						Name: triggerName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: triggerName,
						},
					},
					Spec: svcapitypes.TriggerSpec{
						ForProvider: svcapitypes.TriggerParameters{
							StartOnCreation: &newStartOnCreation,
							TriggerType:     &newTriggerType,
						},
					},
				},
				getTriggerOutput: &svcsdk.GetTriggerOutput{
					Trigger: &svcsdk.Trigger{
						Name: &triggerName,
						Type: &triggerType,
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _, _ := isUpToDate(context.TODO(), tc.args.trigger, tc.args.getTriggerOutput)
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
