/*
Copyright 2019 The Crossplane Authors.
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

package secret

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	svcsdk "github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/secretsmanager/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errNotSecret        = "managed resource is not a secret custom resource"
	errKubeUpdateFailed = "failed to update Secret custom resource"
	errCreateTags       = "failed to create tags for the secret"
	errRemoveTags       = "failed to remove tags for the secret"
	errFmtKeyNotFound   = "key %s is not found in referenced Kubernetes secret"
	errGetSecretFailed  = "failed to get Kubernetes secret"
	errGetSecretValue   = "cannot get the value of secret from AWS"
)

// SetupSecret adds a controller that reconciles a Secret.
func SetupSecret(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.SecretGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			h := &hooks{client: e.client, kube: e.kube}
			e.isUpToDate = h.isUpToDate
			e.preUpdate = h.preUpdate
			e.preCreate = h.preCreate
			e.preDelete = preDelete
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Secret{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.SecretGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Secret, obj *svcsdk.DescribeSecretInput) error {
	obj.SecretId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Secret, resp *svcsdk.DescribeSecretOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	// NOTE(muvaf): No operation can be done for secrets that are marked for deletion,
	// including fetching the content.
	if resp.DeletedDate != nil {
		obs.ResourceExists = false
		return obs, nil
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

type hooks struct {
	client secretsmanageriface.SecretsManagerAPI
	kube   client.Client
}

func (e *hooks) isUpToDate(cr *svcapitypes.Secret, resp *svcsdk.DescribeSecretOutput) (bool, error) { // nolint:gocyclo
	// NOTE(muvaf): No operation can be done on secrets that are marked for deletion.
	if resp.DeletedDate != nil {
		return true, nil
	}
	if awsclients.StringValue(cr.Spec.ForProvider.Description) != awsclients.StringValue(resp.Description) {
		return false, nil
	}
	if awsclients.StringValue(cr.Spec.ForProvider.KMSKeyID) != awsclients.StringValue(resp.KmsKeyId) {
		return false, nil
	}
	add, remove := DiffTags(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(add) != 0 && len(remove) != 0 {
		return false, nil
	}
	// TODO(muvaf): We need isUpToDate to have context.
	ctx := context.TODO()
	s, err := e.client.GetSecretValueWithContext(ctx, &svcsdk.GetSecretValueInput{
		SecretId: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, awsclients.Wrap(err, errGetSecretValue)
	}
	payload, err := e.getPayload(ctx, cr)
	if err != nil {
		return false, err
	}
	switch {
	case awsclients.StringValue(s.SecretString) != "":
		return string(payload) == awsclients.StringValue(s.SecretString), nil
	case len(s.SecretBinary) != 0:
		return bytes.Equal(payload, s.SecretBinary), nil
	}
	return false, errors.New("neither SecretString nor SecretBinary field is filled in the returned object")
}

func (e *hooks) getPayload(ctx context.Context, cr *svcapitypes.Secret) ([]byte, error) {
	var ref *svcapitypes.SecretReference
	switch {
	case cr.Spec.ForProvider.StringSecretRef != nil:
		ref = cr.Spec.ForProvider.StringSecretRef
	case cr.Spec.ForProvider.BinarySecretRef != nil:
		ref = cr.Spec.ForProvider.BinarySecretRef
	default:
		return nil, errors.New("neither binarySecretRef nor stringSecretRef is given")
	}
	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	sc := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, sc); err != nil {
		return nil, errors.Wrap(err, errGetSecretFailed)
	}
	if ref.Key != nil {
		val, ok := sc.Data[awsclients.StringValue(ref.Key)]
		if !ok {
			return nil, errors.New(fmt.Sprintf(errFmtKeyNotFound, awsclients.StringValue(ref.Key)))
		}
		return val, nil
	}
	d := map[string]string{}
	for k, v := range sc.Data {
		d[k] = string(v)
	}
	payload, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.Secret, obj *svcsdk.UpdateSecretInput) error {
	resp, err := e.client.DescribeSecretWithContext(ctx, &svcsdk.DescribeSecretInput{
		SecretId: awsclients.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return awsclients.Wrap(err, errDescribe)
	}
	add, remove := DiffTags(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			SecretId: awsclients.String(meta.GetExternalName(cr)),
			TagKeys:  remove,
		}); err != nil {
			return awsclients.Wrap(err, errRemoveTags)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			SecretId: awsclients.String(meta.GetExternalName(cr)),
			Tags:     add,
		}); err != nil {
			return awsclients.Wrap(err, errCreateTags)
		}
	}
	payload, err := e.getPayload(ctx, cr)
	if err != nil {
		return err
	}
	switch {
	case cr.Spec.ForProvider.StringSecretRef != nil:
		obj.SecretString = awsclients.String(string(payload))
	case cr.Spec.ForProvider.BinarySecretRef != nil:
		obj.SecretBinary = payload
	}
	obj.SecretId = awsclients.String(meta.GetExternalName(cr))
	obj.Description = cr.Spec.ForProvider.Description
	obj.KmsKeyId = cr.Spec.ForProvider.KMSKeyID
	return nil
}

func (e *hooks) preCreate(ctx context.Context, cr *svcapitypes.Secret, obj *svcsdk.CreateSecretInput) error {
	payload, err := e.getPayload(ctx, cr)
	if err != nil {
		return err
	}
	switch {
	case cr.Spec.ForProvider.StringSecretRef != nil:
		obj.SecretString = awsclients.String(string(payload))
	case cr.Spec.ForProvider.BinarySecretRef != nil:
		obj.SecretBinary = payload
	}
	obj.Name = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Secret, obj *svcsdk.DeleteSecretInput) (bool, error) {
	obj.ForceDeleteWithoutRecovery = cr.Spec.ForProvider.ForceDeleteWithoutRecovery
	obj.RecoveryWindowInDays = cr.Spec.ForProvider.RecoveryWindowInDays
	obj.SecretId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

type tagger struct {
	kube client.Client
}

// TODO(knappek): split this out as it is used in several controllers
func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Secret)
	if !ok {
		return errors.New(errNotSecret)
	}
	tagMap := map[string]string{}
	for _, tags := range cr.Spec.ForProvider.Tags {
		tagMap[awsclients.StringValue(tags.Key)] = awsclients.StringValue(tags.Value)
	}
	for k, v := range resource.GetExternalTags(mg) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]*svcapitypes.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = &svcapitypes.Tag{Key: awsclients.String(k), Value: awsclients.String(v)}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return awsclients.StringValue(cr.Spec.ForProvider.Tags[i].Key) < awsclients.StringValue(cr.Spec.ForProvider.Tags[j].Key)
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[awsclients.StringValue(t.Key)] = awsclients.StringValue(t.Value)
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[awsclients.StringValue(t.Key)] == awsclients.StringValue(t.Value) {
			delete(addMap, awsclients.StringValue(t.Key))
			continue
		}
		removeMap[awsclients.StringValue(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{Key: awsclients.String(k), Value: awsclients.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, awsclients.String(k))
	}
	return
}
