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

package loggroup

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errListTags      = "cannot list tags"
	errTagResource   = "cannot tag resource"
	errUntagResource = "cannot untag resource"
)

// ControllerName of this controller.
var ControllerName = managed.ControllerName(svcapitypes.LogGroupGroupKind)

// SetupLogGroup adds a controller that reconciles LogGroup.
func SetupLogGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.LogGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.filterList = filterList
			u := &updater{client: e.client}
			e.isUpToDate = u.isUpToDate
			e.update = u.update
			e.preObserve = preObserve
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.LogGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LogGroupGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type updater struct {
	client svcsdkapi.CloudWatchLogsAPI
}

func filterList(cr *svcapitypes.LogGroup, obj *svcsdk.DescribeLogGroupsOutput) *svcsdk.DescribeLogGroupsOutput {
	logGroupIdentifier := awsclients.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeLogGroupsOutput{}
	for _, LogGroups := range obj.LogGroups {
		if awsclients.StringValue(LogGroups.LogGroupName) == awsclients.StringValue(logGroupIdentifier) {
			resp.LogGroups = append(resp.LogGroups, LogGroups)
			break
		}
	}
	return resp
}

func preObserve(ctx context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.DescribeLogGroupsInput) error {
	obj.SetLogGroupNamePrefix(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.DescribeLogGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider = generateObservation(obj)
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.CreateLogGroupInput) error {
	obj.KmsKeyId = cr.Spec.ForProvider.KMSKeyID
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.CreateLogGroupOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(cr.Spec.ForProvider.LogGroupName))
	return managed.ExternalCreation{}, nil
}

func (u *updater) isUpToDate(cr *svcapitypes.LogGroup, obj *svcsdk.DescribeLogGroupsOutput) (bool, error) {
	if awsclients.Int64Value(cr.Spec.ForProvider.RetentionInDays) != awsclients.Int64Value(obj.LogGroups[0].RetentionInDays) {
		return false, nil
	}

	tags, err := u.client.ListTagsLogGroup(&svcsdk.ListTagsLogGroupInput{
		LogGroupName: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, errors.Wrap(err, errListTags)
	}
	add, remove := awsclients.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, tags.Tags)

	return len(add) == 0 && len(remove) == 0, nil
}

func (u *updater) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*svcapitypes.LogGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	obj, err := u.client.DescribeLogGroupsWithContext(ctx, &svcsdk.DescribeLogGroupsInput{
		LogGroupNamePrefix: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclients.Wrap(err, errCreate)
	}

	tags, err := u.client.ListTagsLogGroup(&svcsdk.ListTagsLogGroupInput{
		LogGroupName: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errListTags)
	}
	add, remove := awsclients.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, tags.Tags)

	if len(add) > 0 {
		_, err := u.client.TagLogGroupWithContext(ctx, &svcsdk.TagLogGroupInput{
			LogGroupName: awsclients.String(meta.GetExternalName(cr)),
			Tags:         add,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTagResource)
		}
	}
	if len(remove) > 0 {
		_, err := u.client.UntagLogGroupWithContext(ctx, &svcsdk.UntagLogGroupInput{
			LogGroupName: awsclients.String(meta.GetExternalName(cr)),
			Tags:         remove,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUntagResource)
		}
	}

	var zero int64 = 0
	if cr.Spec.ForProvider.RetentionInDays == nil || awsclients.Int64Value(cr.Spec.ForProvider.RetentionInDays) == awsclients.Int64Value(&zero) {
		if _, err := u.client.DeleteRetentionPolicy(&svcsdk.DeleteRetentionPolicyInput{
			LogGroupName: awsclients.String(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdate)
		}
	}

	if cr.Spec.ForProvider.RetentionInDays != nil &&
		awsclients.Int64Value(cr.Spec.ForProvider.RetentionInDays) != awsclients.Int64Value(obj.LogGroups[0].RetentionInDays) {
		if _, err := u.client.PutRetentionPolicy(&svcsdk.PutRetentionPolicyInput{
			LogGroupName:    awsclients.String(meta.GetExternalName(cr)),
			RetentionInDays: cr.Spec.ForProvider.RetentionInDays,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func generateObservation(obj *svcsdk.DescribeLogGroupsOutput) svcapitypes.LogGroupObservation {
	if obj == nil || len(obj.LogGroups) == 0 {
		return svcapitypes.LogGroupObservation{}
	}

	o := svcapitypes.LogGroupObservation{
		ARN:               obj.LogGroups[0].Arn,
		CreationTime:      obj.LogGroups[0].CreationTime,
		KMSKeyID:          obj.LogGroups[0].KmsKeyId,
		LogGroupName:      obj.LogGroups[0].LogGroupName,
		MetricFilterCount: obj.LogGroups[0].MetricFilterCount,
		RetentionInDays:   obj.LogGroups[0].RetentionInDays,
		StoredBytes:       obj.LogGroups[0].StoredBytes,
	}

	return o
}
