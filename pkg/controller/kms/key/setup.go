package key

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	svcsdkapi "github.com/aws/aws-sdk-go/service/kms/kmsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/policy"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupKey adds a controller that reconciles Key.
func SetupKey(mgr ctrl.Manager, o controller.Options) error {
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

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithInitializers(),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.KeyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Key{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.DescribeKeyInput) error {
	obj.KeyId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.DescribeKeyOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	// Set Condition
	switch pointer.StringValue(obj.KeyMetadata.KeyState) {
	case string(svcapitypes.KeyState_Enabled):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.KeyState_Disabled):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.KeyState_PendingDeletion):
		cr.SetConditions(xpv1.Deleting())
		return managed.ExternalObservation{ResourceExists: false}, nil
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
	meta.SetExternalName(cr, pointer.StringValue(obj.KeyMetadata.KeyId))
	return managed.ExternalCreation{}, nil
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
			KeyId:       pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Description: cr.Spec.ForProvider.Description,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	// Policy
	if _, err := u.client.PutKeyPolicyWithContext(ctx, &svcsdk.PutKeyPolicyInput{
		KeyId:      pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		PolicyName: pointer.ToOrNilIfZeroValue("default"),
		Policy:     cr.Spec.ForProvider.Policy,
	}); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	if pointer.BoolValue(cr.Spec.ForProvider.EnableKeyRotation) {
		// EnableKeyRotation
		if _, err := u.client.EnableKeyRotationWithContext(ctx, &svcsdk.EnableKeyRotationInput{
			KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	} else {
		// DisableKeyRotation
		if _, err := u.client.DisableKeyRotationWithContext(ctx, &svcsdk.DisableKeyRotationInput{
			KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
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
		KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return errorutils.Wrap(err, errUpdate)
	}

	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, tagsOutput.Tags)

	if len(addTags) != 0 {
		if _, err := u.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Tags:  addTags,
		}); err != nil {
			return errorutils.Wrap(err, "cannot tag Key")
		}
	}
	if len(removeTags) != 0 {
		if _, err := u.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			KeyId:   pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			TagKeys: removeTags,
		}); err != nil {
			return errorutils.Wrap(err, "cannot untag Key")
		}
	}
	return nil
}

func isUpToDateEnableDisable(cr *svcapitypes.Key) bool {
	return pointer.BoolValue(cr.Spec.ForProvider.Enabled) == pointer.BoolValue(cr.Status.AtProvider.Enabled)
}

func (u *updater) enableDisableKey(ctx context.Context, cr *svcapitypes.Key) error {
	if isUpToDateEnableDisable(cr) {
		return nil
	}

	if pointer.BoolValue(cr.Spec.ForProvider.Enabled) {
		if _, err := u.client.EnableKeyWithContext(ctx, &svcsdk.EnableKeyInput{
			KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return errorutils.Wrap(err, "cannot enable Key")
		}
	} else {
		if _, err := u.client.DisableKeyWithContext(ctx, &svcsdk.DisableKeyInput{
			KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		}); err != nil {
			return errorutils.Wrap(err, "cannot disable Key")
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
		KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}

	if cr.Spec.ForProvider.PendingWindowInDays != nil {
		req.PendingWindowInDays = cr.Spec.ForProvider.PendingWindowInDays
	}

	_, err := d.client.ScheduleKeyDeletionWithContext(ctx, req)

	return errorutils.Wrap(err, errDelete)
}

type observer struct {
	client svcsdkapi.KMSAPI
}

func (o *observer) lateInitialize(in *svcapitypes.KeyParameters, obj *svcsdk.DescribeKeyOutput) error {
	// Policy
	if in.Policy == nil {
		resPolicy, err := o.client.GetKeyPolicy(&svcsdk.GetKeyPolicyInput{
			KeyId:      obj.KeyMetadata.KeyId,
			PolicyName: pointer.ToOrNilIfZeroValue("default"),
		})

		if err != nil {
			return errorutils.Wrap(err, "cannot get key policy")
		}

		in.Policy = pointer.LateInitialize(in.Policy, resPolicy.Policy)
	}

	in.Enabled = pointer.LateInitialize(in.Enabled, obj.KeyMetadata.Enabled)

	if len(in.Tags) == 0 {
		resTags, err := o.client.ListResourceTags(&svcsdk.ListResourceTagsInput{
			KeyId: obj.KeyMetadata.KeyId,
		})

		if err != nil {
			return errorutils.Wrap(err, "cannot list tags")
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

func (o *observer) isUpToDate(_ context.Context, cr *svcapitypes.Key, obj *svcsdk.DescribeKeyOutput) (bool, string, error) { //nolint:gocyclo
	// Description
	if obj.KeyMetadata.Description != nil &&
		cr.Spec.ForProvider.Description != nil &&
		pointer.StringValue(obj.KeyMetadata.Description) != pointer.StringValue(cr.Spec.ForProvider.Description) {
		return false, "", nil
	}

	// Enabled
	if !isUpToDateEnableDisable(cr) {
		return false, "", nil
	}

	// KeyPolicy
	resPolicy, err := o.client.GetKeyPolicy(&svcsdk.GetKeyPolicyInput{
		KeyId:      pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		PolicyName: pointer.ToOrNilIfZeroValue("default"),
	})
	if err != nil {
		return false, "", errorutils.Wrap(err, "cannot get key policy")
	}
	specPolicy, err := policy.ParsePolicyStringPtr(cr.Spec.ForProvider.Policy)
	if err != nil {
		return false, "", errors.Wrap(err, "cannot parse spec policy")
	}
	currentPolicy, err := policy.ParsePolicyStringPtr(resPolicy.Policy)
	if err != nil {
		return false, "", errors.Wrap(err, "cannot parse current policy")
	}
	if equal, diff := policy.ArePoliciesEqal(specPolicy, currentPolicy); !equal {
		return false, "spec.forProvider.policy: " + diff, nil
	}

	// EnableKeyRotation
	resRotation, err := o.client.GetKeyRotationStatus(&svcsdk.GetKeyRotationStatusInput{
		KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, "", errorutils.Wrap(err, "cannot get key rotation status")
	}
	if pointer.BoolValue(cr.Spec.ForProvider.EnableKeyRotation) != pointer.BoolValue(resRotation.KeyRotationEnabled) {
		return false, "", nil
	}

	// Tags
	resTags, err := o.client.ListResourceTags(&svcsdk.ListResourceTagsInput{
		KeyId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, "", errorutils.Wrap(err, "cannot list tags")
	}
	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, resTags.Tags)
	return len(addTags) == 0 && len(removeTags) == 0, "", nil
}

// returns which AWS Tags exist in the resource tags and which are outdated and should be removed
func diffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[pointer.StringValue(t.TagKey)] = pointer.StringValue(t.TagValue)
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[pointer.StringValue(t.TagKey)] == pointer.StringValue(t.TagValue) {
			delete(addMap, pointer.StringValue(t.TagKey))
			continue
		}
		removeMap[pointer.StringValue(t.TagKey)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{TagKey: pointer.ToOrNilIfZeroValue(k), TagValue: pointer.ToOrNilIfZeroValue(v)})
	}
	for k := range removeMap {
		remove = append(remove, pointer.ToOrNilIfZeroValue(k))
	}
	return
}
