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

package stream

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/kinesis"
	svcsdkapi "github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kinesis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupStream adds a controller that reconciles Stream.
func SetupStream(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.StreamGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.preCreate = preCreate
			u := &updater{client: e.client}
			e.update = u.update
			e.isUpToDate = u.isUpToDate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.StreamGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Stream{}).
		Complete(r)
}

func preDelete(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.DeleteStreamInput) (bool, error) {
	obj.EnforceConsumerDeletion = cr.Spec.ForProvider.EnforceConsumerDeletion
	obj.StreamName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.DescribeStreamInput) error {
	obj.StreamName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.DescribeStreamOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(obj.StreamDescription.StreamStatus) {
	case string(svcapitypes.StreamStatus_SDK_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.StreamStatus_SDK_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.StreamStatus_SDK_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}

	cr.Status.AtProvider = GenerateObservation(obj.StreamDescription)

	obs.ConnectionDetails = managed.ConnectionDetails{
		"arn":  []byte(pointer.StringValue(obj.StreamDescription.StreamARN)),
		"name": []byte(meta.GetExternalName(cr)),
	}

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.CreateStreamInput) error {
	obj.StreamName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.CreateStreamOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	return managed.ExternalCreation{}, nil
}

type updater struct {
	client svcsdkapi.KinesisAPI
}

func (u *updater) isUpToDate(_ context.Context, cr *svcapitypes.Stream, obj *svcsdk.DescribeStreamOutput) (bool, string, error) { //nolint:gocyclo

	// ResourceInUseException: Stream example-stream not ACTIVE, instead in state CREATING
	if pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		// filter activeShards
		number, err := u.ActiveShards(cr)
		if err != nil {
			return false, "", err
		}

		if pointer.Int64Value(cr.Spec.ForProvider.ShardCount) != number {
			return false, "", nil
		}

		if pointer.Int64Value(cr.Spec.ForProvider.RetentionPeriodHours) != pointer.Int64Value(obj.StreamDescription.RetentionPeriodHours) {
			return false, "", nil
		}

		if pointer.StringValue(cr.Spec.ForProvider.KMSKeyARN) != pointer.StringValue(obj.StreamDescription.KeyId) {
			return false, "", nil
		}

		// Prevent an out of range panic when enhanced metrics
		// arent defined in the spec
		if len(cr.Spec.ForProvider.EnhancedMetrics) == 0 {
			cr.Spec.ForProvider.EnhancedMetrics = append(cr.Spec.ForProvider.EnhancedMetrics, &svcapitypes.EnhancedMetrics{})
		}

		createKey, deleteKey := DifferenceShardLevelMetrics(cr.Spec.ForProvider.EnhancedMetrics[0].ShardLevelMetrics, obj.StreamDescription.EnhancedMonitoring[0].ShardLevelMetrics)
		if len(createKey) != 0 || len(deleteKey) != 0 {
			return false, "", nil
		}

		objTags, err := u.ListTags(cr)
		if err != nil &&
			objTags == nil {
			return false, "", err
		}
		addTags, removeTags := DiffTags(cr.Spec.ForProvider.Tags, objTags.Tags)

		if len(addTags) != 0 || len(removeTags) != 0 {
			return false, "", nil
		}

	}

	return true, "", nil
}

func (u *updater) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*svcapitypes.Stream)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// we need information from stream for decisions
	obj, err := u.client.DescribeStreamWithContext(ctx, &svcsdk.DescribeStreamInput{
		StreamName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreate)
	}

	// we need information about activeShards for decision
	number, err := u.ActiveShards(cr)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}
	if pointer.Int64Value(cr.Spec.ForProvider.ShardCount) != number &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		scalingType := svcsdk.ScalingTypeUniformScaling
		if _, err := u.client.UpdateShardCountWithContext(ctx, &svcsdk.UpdateShardCountInput{
			StreamName:       pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			TargetShardCount: cr.Spec.ForProvider.ShardCount,
			ScalingType:      &scalingType,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if pointer.Int64Value(cr.Spec.ForProvider.RetentionPeriodHours) > pointer.Int64Value(obj.StreamDescription.RetentionPeriodHours) &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		if _, err := u.client.IncreaseStreamRetentionPeriodWithContext(ctx, &svcsdk.IncreaseStreamRetentionPeriodInput{
			StreamName:           pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			RetentionPeriodHours: cr.Spec.ForProvider.RetentionPeriodHours,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if pointer.Int64Value(cr.Spec.ForProvider.RetentionPeriodHours) < pointer.Int64Value(obj.StreamDescription.RetentionPeriodHours) &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		if _, err := u.client.DecreaseStreamRetentionPeriodWithContext(ctx, &svcsdk.DecreaseStreamRetentionPeriodInput{
			StreamName:           pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			RetentionPeriodHours: cr.Spec.ForProvider.RetentionPeriodHours,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if cr.Spec.ForProvider.KMSKeyARN != nil &&
		pointer.StringValue(obj.StreamDescription.EncryptionType) == svcsdk.EncryptionTypeNone &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		kmsType := svcsdk.EncryptionTypeKms
		if _, err := u.client.StartStreamEncryptionWithContext(ctx, &svcsdk.StartStreamEncryptionInput{
			EncryptionType: &kmsType,
			KeyId:          cr.Spec.ForProvider.KMSKeyARN,
			StreamName:     pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if cr.Spec.ForProvider.KMSKeyARN == nil &&
		pointer.StringValue(obj.StreamDescription.EncryptionType) == svcsdk.EncryptionTypeKms &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		// The quirk about this API is that, when we are disabling the StreamEncryption
		// We need to pass in that old KMS Key Id that was being used for Encryption and
		// We also need to pass in the type of Encryption we were using - i.e. KMS as that
		// Is the only supported Encryption method right now
		// If we don't get this and pass in the actual EncryptionType we want to move to i.e. NONE
		// We get the following error
		//
		//        InvalidArgumentException: Encryption type cannot be NONE.
		if _, err := u.client.StopStreamEncryptionWithContext(ctx, &svcsdk.StopStreamEncryptionInput{
			EncryptionType: obj.StreamDescription.EncryptionType,
			KeyId:          obj.StreamDescription.KeyId,
			StreamName:     pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	// Prevent an out of range panic when enhanced metrics
	// arent defined in the spec
	if len(cr.Spec.ForProvider.EnhancedMetrics) == 0 {
		cr.Spec.ForProvider.EnhancedMetrics = append(cr.Spec.ForProvider.EnhancedMetrics, &svcapitypes.EnhancedMetrics{})
	}

	enableMetrics, disableMetrics := DifferenceShardLevelMetrics(cr.Spec.ForProvider.EnhancedMetrics[0].ShardLevelMetrics, obj.StreamDescription.EnhancedMonitoring[0].ShardLevelMetrics)
	if len(enableMetrics) != 0 &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {

		if _, err := u.client.EnableEnhancedMonitoringWithContext(ctx, &svcsdk.EnableEnhancedMonitoringInput{
			ShardLevelMetrics: cr.Spec.ForProvider.EnhancedMetrics[0].ShardLevelMetrics,
			StreamName:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if len(disableMetrics) != 0 &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {

		if _, err := u.client.DisableEnhancedMonitoringWithContext(ctx, &svcsdk.DisableEnhancedMonitoringInput{
			ShardLevelMetrics: obj.StreamDescription.EnhancedMonitoring[0].ShardLevelMetrics,
			StreamName:        pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	objTags, err := u.ListTags(cr)
	if err != nil &&
		objTags == nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}
	addTags, removeTags := DiffTags(cr.Spec.ForProvider.Tags, objTags.Tags)

	if len(addTags) != 0 &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		if _, err := u.client.AddTagsToStreamWithContext(ctx, &svcsdk.AddTagsToStreamInput{
			StreamName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Tags:       addTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	if len(removeTags) != 0 &&
		pointer.StringValue(obj.StreamDescription.StreamStatus) == svcsdk.StreamStatusActive {
		if _, err := u.client.RemoveTagsFromStreamWithContext(ctx, &svcsdk.RemoveTagsFromStreamInput{
			StreamName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			TagKeys:    removeTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
		// You can't make other updates to the data stream while it is being updated.
		return managed.ExternalUpdate{}, nil
	}

	return managed.ExternalUpdate{}, nil
}

// DifferenceShardLevelMetrics returns the lists of ShardLevelMetrics that need to be removed and added according
// to current and desired states.
func DifferenceShardLevelMetrics(local []*string, remote []*string) ([]*string, []*string) {
	createKey := []*string{}
	removeKey := []*string{}
	m := map[string]int{}

	for _, value := range local {
		m[*value] = 1
	}

	for _, value := range remote {
		m[*value] += 2
	}

	for mKey, mVal := range m {
		// need for scopelint
		mKey2 := mKey
		if mVal == 1 {
			createKey = append(createKey, &mKey2)
		}

		if mVal == 2 {
			removeKey = append(removeKey, &mKey2)
		}
	}
	return createKey, removeKey
}

// ActiveShards count open shards without EndingSequenceNumber
func (u *updater) ActiveShards(cr *svcapitypes.Stream) (int64, error) {
	var count int64

	shards, err := u.client.ListShards(&svcsdk.ListShardsInput{
		StreamName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return count, err
	}

	for _, shard := range shards.Shards {
		if shard.SequenceNumberRange.EndingSequenceNumber == nil {
			count++
		}
	}

	return count, nil
}

// ListTags return the current tags
func (u *updater) ListTags(cr *svcapitypes.Stream) (*svcsdk.ListTagsForStreamOutput, error) {

	tags, err := u.client.ListTagsForStream(&svcsdk.ListTagsForStreamInput{
		StreamName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// DiffTags returns the lists of tags that need to be removed and added according
// to current and desired states.
func DiffTags(local []svcapitypes.CustomTag, remote []*svcsdk.Tag) (add map[string]*string, remove []*string) {
	addMap := make(map[string]string, len(local))
	for _, t := range local {
		addMap[t.Key] = t.Value
	}

	removeMap := map[string]struct{}{}
	for _, t := range remote {
		if addMap[pointer.StringValue(t.Key)] == pointer.StringValue(t.Value) {
			delete(addMap, pointer.StringValue(t.Key))
			continue
		}
		removeMap[pointer.StringValue(t.Key)] = struct{}{}
	}

	addTag := make([]svcsdk.Tag, 0)
	for k, v := range addMap {
		k := k
		v := v
		addTag = append(addTag, svcsdk.Tag{
			Key:   &k,
			Value: &v,
		})
	}

	for k := range removeMap {
		k := k
		remove = append(remove, &k)
	}

	add = make(map[string]*string)
	for _, t := range addTag {
		add[*t.Key] = t.Value
	}
	return add, remove
}

// GenerateObservation is used to produce v1beta1.ClusterObservation from
// ekstypes.Cluster.
func GenerateObservation(obj *svcsdk.StreamDescription) svcapitypes.StreamObservation { //nolint:gocyclo
	if obj == nil {
		return svcapitypes.StreamObservation{}
	}

	o := svcapitypes.StreamObservation{
		EncryptionType:       obj.EncryptionType,
		HasMoreShards:        obj.HasMoreShards,
		KeyID:                obj.KeyId,
		RetentionPeriodHours: obj.RetentionPeriodHours,
		StreamARN:            obj.StreamARN,
		StreamStatus:         obj.StreamStatus,
	}

	if obj.EnhancedMonitoring != nil {
		f1 := []*svcapitypes.EnhancedMetrics{}
		for _, f1iter := range obj.EnhancedMonitoring {
			f1elem := &svcapitypes.EnhancedMetrics{}
			if f1iter.ShardLevelMetrics != nil {
				f1elemf0 := []*string{}
				for _, f1elemf0iter := range f1iter.ShardLevelMetrics {
					var f1elemf0elem = *f1elemf0iter
					f1elemf0 = append(f1elemf0, &f1elemf0elem)
				}
				f1elem.ShardLevelMetrics = f1elemf0
			}
			f1 = append(f1, f1elem)
		}
		o.EnhancedMonitoring = f1
	}

	if obj.Shards != nil {
		f5 := []*svcapitypes.Shard{}
		for _, f5iter := range obj.Shards {
			f5elem := &svcapitypes.Shard{}
			if f5iter.AdjacentParentShardId != nil {
				f5elem.AdjacentParentShardID = f5iter.AdjacentParentShardId
			}
			if f5iter.HashKeyRange != nil {
				f5elemf1 := &svcapitypes.HashKeyRange{}
				if f5iter.HashKeyRange.EndingHashKey != nil {
					f5elemf1.EndingHashKey = f5iter.HashKeyRange.EndingHashKey
				}
				if f5iter.HashKeyRange.StartingHashKey != nil {
					f5elemf1.StartingHashKey = f5iter.HashKeyRange.StartingHashKey
				}
				f5elem.HashKeyRange = f5elemf1
			}
			if f5iter.ParentShardId != nil {
				f5elem.ParentShardID = f5iter.ParentShardId
			}
			if f5iter.SequenceNumberRange != nil {
				f5elemf3 := &svcapitypes.SequenceNumberRange{}
				if f5iter.SequenceNumberRange.EndingSequenceNumber != nil {
					f5elemf3.EndingSequenceNumber = f5iter.SequenceNumberRange.EndingSequenceNumber
				}
				if f5iter.SequenceNumberRange.StartingSequenceNumber != nil {
					f5elemf3.StartingSequenceNumber = f5iter.SequenceNumberRange.StartingSequenceNumber
				}
				f5elem.SequenceNumberRange = f5elemf3
			}
			if f5iter.ShardId != nil {
				f5elem.ShardID = f5iter.ShardId
			}
			f5 = append(f5, f5elem)
		}
		o.Shards = f5
	}

	return o
}
