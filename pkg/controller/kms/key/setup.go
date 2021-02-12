package key

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	svcsdkapi "github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/kms/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupKey adds a controller that reconciles Key.
func SetupKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.KeyGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.postCreate = postCreate
			u := &updater{client: e.client}
			e.update = u.update
			d := &deleter{client: e.client}
			e.delete = d.delete
			o := &observer{client: e.client}
			e.isUpToDate = o.isUpToDate
			e.lateInitialize = o.lateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Key{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.KeyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.DescribeKeyInput) error {
	obj.KeyId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.DescribeKeyOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	// Set Condition
	switch awsclients.StringValue(obj.KeyMetadata.KeyState) {
	case string(svcapitypes.KeyState_Enabled):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.KeyState_Disabled):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.KeyState_PendingDeletion):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.KeyState_PendingImport):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.KeyState_Unavailable):
		cr.SetConditions(xpv1.Unavailable())
	}

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.CreateKeyOutput, creation managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return creation, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.KeyMetadata.KeyId))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

type updater struct {
	client svcsdkapi.KMSAPI
}

func (u *updater) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Key)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	if cr.Spec.ForProvider.Description != nil {
		if _, err := u.client.UpdateKeyDescriptionWithContext(ctx, &svcsdk.UpdateKeyDescriptionInput{
			KeyId:       awsclients.String(meta.GetExternalName(cr)),
			Description: cr.Spec.ForProvider.Description,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdate)
		}
	}

	// Policy
	if _, err := u.client.PutKeyPolicyWithContext(ctx, &svcsdk.PutKeyPolicyInput{
		KeyId:      awsclients.String(meta.GetExternalName(cr)),
		PolicyName: awsclients.String("default"),
		Policy:     cr.Spec.ForProvider.Policy,
	}); err != nil {
		return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdate)
	}

	// Tags
	if err := u.updateTags(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Enable / Disable
	if err := u.enableDisableKey(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

func (u *updater) updateTags(ctx context.Context, cr *svcapitypes.Key) error {
	tagsOutput, err := u.client.ListResourceTagsWithContext(ctx, &svcsdk.ListResourceTagsInput{
		KeyId: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return awsclients.Wrap(err, errUpdate)
	}

	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, tagsOutput.Tags)

	if len(addTags) != 0 {
		if _, err := u.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			KeyId: awsclients.String(meta.GetExternalName(cr)),
			Tags:  addTags,
		}); err != nil {
			return awsclients.Wrap(err, "cannot tag Key")
		}
	}
	if len(removeTags) != 0 {
		if _, err := u.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			KeyId:   awsclients.String(meta.GetExternalName(cr)),
			TagKeys: removeTags,
		}); err != nil {
			return awsclients.Wrap(err, "cannot untag Key")
		}
	}
	return nil
}

func isUpToDateEnableDisable(cr *svcapitypes.Key) bool {
	return awsclients.BoolValue(cr.Spec.ForProvider.Enabled) == awsclients.BoolValue(cr.Status.AtProvider.Enabled)
}

func (u *updater) enableDisableKey(ctx context.Context, cr *svcapitypes.Key) error {
	if isUpToDateEnableDisable(cr) {
		return nil
	}

	if awsclients.BoolValue(cr.Spec.ForProvider.Enabled) {
		if _, err := u.client.EnableKeyWithContext(ctx, &svcsdk.EnableKeyInput{
			KeyId: awsclients.String(meta.GetExternalName(cr)),
		}); err != nil {
			return awsclients.Wrap(err, "cannot enable Key")
		}
	} else {
		if _, err := u.client.DisableKeyWithContext(ctx, &svcsdk.DisableKeyInput{
			KeyId: awsclients.String(meta.GetExternalName(cr)),
		}); err != nil {
			return awsclients.Wrap(err, "cannot disable Key")
		}
	}
	return nil
}

type deleter struct {
	client svcsdkapi.KMSAPI
}

// schedule for deletion instead of delete
func (d *deleter) delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Key)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Deleting())
	// special case: if key is scheduled for deletion, abort early and do not schedule for deletion again
	if cr.Status.AtProvider.DeletionDate != nil {
		return nil
	}

	req := &svcsdk.ScheduleKeyDeletionInput{
		KeyId: awsclients.String(meta.GetExternalName(cr)),
	}

	if cr.Spec.ForProvider.PendingWindowInDays != nil {
		req.PendingWindowInDays = cr.Spec.ForProvider.PendingWindowInDays
	}

	_, err := d.client.ScheduleKeyDeletionWithContext(ctx, req)

	return awsclients.Wrap(err, errDelete)
}

type observer struct {
	client svcsdkapi.KMSAPI
}

func (o *observer) lateInitialize(in *svcapitypes.KeyParameters, obj *svcsdk.DescribeKeyOutput) error {
	// Policy
	if in.Policy == nil {
		resPolicy, err := o.client.GetKeyPolicy(&svcsdk.GetKeyPolicyInput{
			KeyId:      obj.KeyMetadata.KeyId,
			PolicyName: awsclients.String("default"),
		})

		if err != nil {
			return awsclients.Wrap(err, "cannot get key policy")
		}

		in.Policy = awsclients.LateInitializeStringPtr(in.Policy, resPolicy.Policy)
	}

	in.Enabled = awsclients.LateInitializeBoolPtr(in.Enabled, obj.KeyMetadata.Enabled)

	if len(in.Tags) == 0 {
		resTags, err := o.client.ListResourceTags(&svcsdk.ListResourceTagsInput{
			KeyId: obj.KeyMetadata.KeyId,
		})

		if err != nil {
			return awsclients.Wrap(err, "cannot list tags")
		}

		if len(resTags.Tags) > 0 {
			for _, t := range resTags.Tags {
				in.Tags = append(in.Tags, &svcapitypes.Tag{
					TagKey:   t.TagKey,
					TagValue: t.TagValue,
				})
			}
		}
	}

	return nil
}

func (o *observer) isUpToDate(cr *svcapitypes.Key, obj *svcsdk.DescribeKeyOutput) (bool, error) {
	// Description
	if obj.KeyMetadata.Description != nil &&
		cr.Spec.ForProvider.Description != nil &&
		awsclients.StringValue(obj.KeyMetadata.Description) != awsclients.StringValue(cr.Spec.ForProvider.Description) {
		return false, nil
	}

	// Enabled
	if !isUpToDateEnableDisable(cr) {
		return false, nil
	}

	// KeyPolicy
	resPolicy, err := o.client.GetKeyPolicy(&svcsdk.GetKeyPolicyInput{
		KeyId:      awsclients.String(meta.GetExternalName(cr)),
		PolicyName: awsclients.String("default"),
	})
	if err != nil {
		return false, awsclients.Wrap(err, "cannot get key policy")
	}
	if awsclients.StringValue(cr.Spec.ForProvider.Policy) != awsclients.StringValue(resPolicy.Policy) {
		return false, nil
	}

	// Tags
	resTags, err := o.client.ListResourceTags(&svcsdk.ListResourceTagsInput{
		KeyId: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, awsclients.Wrap(err, "cannot list tags")
	}
	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, resTags.Tags)
	return len(addTags) == 0 && len(removeTags) == 0, nil
}

// returns which AWS Tags exist in the resource tags and which are outdated and should be removed
func diffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[awsclients.StringValue(t.TagKey)] = awsclients.StringValue(t.TagValue)
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[awsclients.StringValue(t.TagKey)] == awsclients.StringValue(t.TagValue) {
			delete(addMap, awsclients.StringValue(t.TagKey))
			continue
		}
		removeMap[awsclients.StringValue(t.TagKey)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{TagKey: awsclients.String(k), TagValue: awsclients.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, awsclients.String(k))
	}
	return
}
