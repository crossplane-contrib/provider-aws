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
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudwatchlogs/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errListTags      = "cannot list tags"
	errTagResource   = "cannot tag resource"
	errUntagResource = "cannot untag resource"
)

// SetupLogGroup adds a controller that reconciles LogGroup.
func SetupLogGroup(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter, poll time.Duration) error {
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
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(limiter),
		}).
		For(&svcapitypes.LogGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LogGroupGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func postObserve(_ context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.DescribeLogGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.CreateLogGroupInput) error {
	obj.KmsKeyId = cr.Spec.ForProvider.KMSKeyID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.LogGroup, obj *svcsdk.CreateLogGroupOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
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
