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

package user

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	svcsdkapi "github.com/aws/aws-sdk-go/service/transfer/transferiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/transfer/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errImportSSHKeys = "cannot import SSH keys"
	errDeleteSSHKeys = "cannot delete SSH keys"
	errTag           = "cannot add tags"
	errUntag         = "cannot remove tags"
)

// SetupUser adds a controller that reconciles User.
func SetupUser(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserGroupKind)

	opts := []option{setupHooks()}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.UserGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.User{}).
		Complete(r)
}

func setupHooks() option {
	return func(e *external) {
		h := hooks{client: e.client}
		e.postObserve = postObserve
		e.preObserve = preObserve
		e.preDelete = preDelete
		e.preCreate = preCreate
		e.lateInitialize = lateInitialize
		e.isUpToDate = h.isUpToDate
		e.postUpdate = h.postUpdate
	}
}

type hooks struct {
	client svcsdkapi.TransferAPI

	cache struct {
		keyBodiesToImport []string
		keyIDsToDelete    []string

		tagsToAdd    []*svcsdk.Tag
		tagsToDelete []*string
	}
}

func preObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserInput) error {
	obj.UserName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.ServerId = cr.Spec.ForProvider.ServerID
	return nil
}

func lateInitialize(spec *svcapitypes.UserParameters, obj *svcsdk.DescribeUserOutput) error {
	// LateInitialize SSHPublicKeys with the deprecated SshPublicKeyBody
	// to ensure backward compatibility with previous versions
	if spec.SSHPublicKeys == nil && spec.SshPublicKeyBody != nil { //nolint:staticcheck
		spec.SSHPublicKeys = []svcapitypes.SSHPublicKeySpec{
			{
				Body: *spec.SshPublicKeyBody, //nolint:staticcheck
			},
		}
	}

	if obj.User != nil && spec.SSHPublicKeys == nil {
		spec.SSHPublicKeys = generateAPISSHPublicKeys(obj.User.SshPublicKeys)
	}

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DeleteUserInput) (bool, error) {
	obj.UserName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.ServerId = cr.Spec.ForProvider.ServerID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if obj.User != nil {
		cr.Status.AtProvider.SshPublicKeys = generateAPIStatusSSHPublicKeys(obj.User.SshPublicKeys)
		cr.Status.AtProvider.UserName = obj.User.UserName
		cr.Status.AtProvider.ARN = obj.User.Arn
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func (h *hooks) isUpToDate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserOutput) (bool, string, error) {
	if obj.User == nil {
		return true, "", nil
	}

	var areKeysUpToDate bool
	areKeysUpToDate, h.cache.keyBodiesToImport, h.cache.keyIDsToDelete = isSSHPublicKeysUpToDate(cr, obj)

	var areTagsUpToDate bool
	areTagsUpToDate, h.cache.tagsToAdd, h.cache.tagsToDelete = utils.DiffTags(cr.Spec.ForProvider.Tags, obj.User.Tags)

	return areKeysUpToDate && areTagsUpToDate, "", nil
}

func isSSHPublicKeysUpToDate(cr *svcapitypes.User, obj *svcsdk.DescribeUserOutput) (isUpToDate bool, toAdd, toRemove []string) {
	specMap := make(map[string]any, len(cr.Spec.ForProvider.SSHPublicKeys))
	for _, k := range cr.Spec.ForProvider.SSHPublicKeys {
		specMap[k.Body] = nil
	}

	curMap := make(map[string]any, len(obj.User.SshPublicKeys))

	toRemove = []string{}
	toAdd = []string{}

	for _, k := range obj.User.SshPublicKeys {
		body := ptr.Deref(k.SshPublicKeyBody, "")
		curMap[body] = nil

		if _, exists := specMap[body]; !exists {
			toRemove = append(toRemove, ptr.Deref(k.SshPublicKeyId, ""))
		}
	}

	for _, k := range cr.Spec.ForProvider.SSHPublicKeys {
		if _, exists := curMap[k.Body]; !exists {
			toAdd = append(toAdd, k.Body)
		}
	}

	return len(toRemove) == 0 && len(toAdd) == 0, toAdd, toRemove
}

func preCreate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserInput) error {
	obj.ServerId = cr.Spec.ForProvider.ServerID
	obj.Role = cr.Spec.ForProvider.Role
	obj.UserName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// NOTE: SSH public keys are added during postUpdate.
	return nil
}

func (h *hooks) postUpdate(ctx context.Context, cr *svcapitypes.User, resp *svcsdk.UpdateUserOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	if len(h.cache.keyBodiesToImport) > 0 {
		if err := h.importSSHKeys(ctx, cr, h.cache.keyBodiesToImport); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errImportSSHKeys)
		}
	}
	if len(h.cache.keyIDsToDelete) > 0 {
		if err := h.deleteSSHKeys(ctx, cr, h.cache.keyIDsToDelete); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errDeleteSSHKeys)
		}
	}

	if len(h.cache.tagsToAdd) > 0 {
		_, err := h.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			Arn:  cr.Status.AtProvider.ARN,
			Tags: h.cache.tagsToAdd,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTag)
		}
	}
	if len(h.cache.tagsToDelete) > 0 {
		_, err := h.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			Arn:     cr.Status.AtProvider.ARN,
			TagKeys: h.cache.tagsToDelete,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUntag)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (h *hooks) importSSHKeys(ctx context.Context, cr *svcapitypes.User, keyBodies []string) error {
	errs := []error{}
	for _, k := range keyBodies {
		_, err := h.client.ImportSshPublicKeyWithContext(ctx, &svcsdk.ImportSshPublicKeyInput{
			ServerId:         cr.Spec.ForProvider.ServerID,
			UserName:         pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			SshPublicKeyBody: ptr.To(k),
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errorutils.Combine(errs)
}

func (h *hooks) deleteSSHKeys(ctx context.Context, cr *svcapitypes.User, keyIDs []string) error {
	errs := []error{}
	for _, k := range keyIDs {
		_, err := h.client.DeleteSshPublicKeyWithContext(ctx, &svcsdk.DeleteSshPublicKeyInput{
			ServerId:       cr.Spec.ForProvider.ServerID,
			UserName:       pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			SshPublicKeyId: ptr.To(k),
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errorutils.Combine(errs)
}
