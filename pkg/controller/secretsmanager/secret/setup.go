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

	svcsdk "github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/secretsmanager/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	policyutils "github.com/crossplane-contrib/provider-aws/pkg/utils/policy"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotSecret            = "managed resource is not a secret custom resource"
	errKubeUpdateFailed     = "failed to update Secret custom resource"
	errCreateTags           = "failed to create tags for the secret"
	errRemoveTags           = "failed to remove tags for the secret"
	errFmtKeyNotFound       = "key %s is not found in referenced Kubernetes secret"
	errGetSecretValue       = "cannot get the value of secret from AWS"
	errGetResourcePolicy    = "cannot get resource policy"
	errPutResourcePolicy    = "cannot put resource policy"
	errDeleteResourcePolicy = "cannot delete resource policy"
	errParseSecretValue     = "cannot parse AWS secret value"
	errGetAWSSecretValue    = "cannot get AWS secret value"
	errCreateK8sSecret      = "canoot create secret in K8s"
	errNoAWSValue           = "neither SecretString nor SecretBinary field is filled in the returned object"
	errNoSecretRef          = "neither binarySecretRef nor stringSecretRef is given"
	errOnlyOneSecretRef     = "only one of binarySecretRef or stringSecretRef must be set"
	errParseSpecPolicy      = "cannot parse spec policy"
	errParseExternalPolicy  = "cannot parse external policy"
)

// SetupSecret adds a controller that reconciles a Secret.
func SetupSecret(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.SecretGroupKind)
	opts := []option{setupExternal}

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
		resource.ManagedKind(svcapitypes.SecretGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Secret{}).
		Complete(r)
}

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	h := &hooks{client: e.client, kube: e.kube}
	e.lateInitialize = h.lateInitialize
	e.isUpToDate = h.isUpToDate
	e.preUpdate = h.preUpdate
	e.preCreate = h.preCreate
	e.preDelete = preDelete
}

func preObserve(_ context.Context, cr *svcapitypes.Secret, obj *svcsdk.DescribeSecretInput) error {
	obj.SecretId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
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

func (e *hooks) lateInitialize(spec *svcapitypes.SecretParameters, resp *svcsdk.DescribeSecretOutput) error {
	_, err := e.getPayload(context.TODO(), spec)
	if err := client.IgnoreNotFound(err); err != nil {
		return err
	}
	// Proceed only if the secret does not exist because empty value might be
	// valid content.
	if !kerrors.IsNotFound(err) {
		// Set secretref.Type with referenced k8s SecretType, if not provided by user
		// or not lateInitialized yet
		ref, err := getSecretRef(spec)
		if err != nil {
			return err
		}
		if ref.Type == nil {
			nn := types.NamespacedName{
				Name:      ref.Name,
				Namespace: ref.Namespace,
			}
			sc := &corev1.Secret{}
			if err := e.kube.Get(context.TODO(), nn, sc); err != nil {
				return err
			}
			ref.Type = pointer.ToOrNilIfZeroValue(string(sc.Type))
		}

		return nil
	}

	// If the K8s does not exist, create it with the data from AWS
	req := &svcsdk.GetSecretValueInput{
		SecretId: resp.ARN,
	}
	valueResp, err := e.client.GetSecretValueWithContext(context.TODO(), req)
	if err != nil {
		return errors.Wrap(err, errGetSecretValue)
	}
	ref, err := getSecretRef(spec)
	if err != nil {
		return err
	}

	data, err := getAWSSecretData(ref, valueResp)
	if err != nil {
		return errors.Wrap(err, errGetAWSSecretValue)
	}
	sc := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
		Data: data,
	}
	if ref.Type != nil {
		sc.Type = corev1.SecretType(pointer.StringValue(ref.Type))
	}
	return errors.Wrap(e.kube.Create(context.TODO(), sc), errCreateK8sSecret)
}

func getAWSSecretData(ref *svcapitypes.SecretReference, s *svcsdk.GetSecretValueOutput) (map[string][]byte, error) {
	if ref.Key != nil {
		switch {
		case pointer.StringValue(s.SecretString) != "":
			return map[string][]byte{
				*ref.Key: []byte(pointer.StringValue(s.SecretString)),
			}, nil
		case len(s.SecretBinary) != 0:
			return map[string][]byte{
				*ref.Key: s.SecretBinary,
			}, nil
		default:
			return nil, errors.New(errNoAWSValue)
		}
	}

	var raw []byte

	switch {
	case pointer.StringValue(s.SecretString) != "":
		raw = []byte(pointer.StringValue(s.SecretString))
	case len(s.SecretBinary) != 0:
		raw = s.SecretBinary
	default:
		return nil, errors.New(errNoAWSValue)
	}

	parsed := map[string]string{}
	err := json.Unmarshal(raw, &parsed)
	if err != nil {
		return nil, errors.Wrap(err, errParseSecretValue)
	}

	payload := map[string][]byte{}
	for k, v := range parsed {
		payload[k] = []byte(v)
	}
	return payload, nil
}

func (e *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.Secret, resp *svcsdk.DescribeSecretOutput) (bool, string, error) {
	// NOTE(muvaf): No operation can be done on secrets that are marked for deletion.
	if resp.DeletedDate != nil {
		return true, "", nil
	}
	if pointer.StringValue(cr.Spec.ForProvider.Description) != pointer.StringValue(resp.Description) {
		return false, "", nil
	}
	if pointer.StringValue(cr.Spec.ForProvider.KMSKeyID) != pointer.StringValue(resp.KmsKeyId) {
		return false, "", nil
	}
	add, remove := DiffTags(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(add) != 0 && len(remove) != 0 {
		return false, "", nil
	}

	isPolicyUpToDate, err := e.isPolicyUpToDate(ctx, cr)
	if err != nil {
		return false, "", err
	}
	if !isPolicyUpToDate {
		return false, "", nil
	}

	isPayloadUpToDate, err := e.isPayloadUpToDate(ctx, cr)
	return isPayloadUpToDate, "", err
}

func (e *hooks) isPolicyUpToDate(ctx context.Context, cr *svcapitypes.Secret) (bool, error) {
	res, err := e.client.GetResourcePolicyWithContext(ctx, &svcsdk.GetResourcePolicyInput{
		SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, errors.Wrap(err, errGetResourcePolicy)
	}

	if res.ResourcePolicy == nil || cr.Spec.ForProvider.ResourcePolicy == nil {
		return res.ResourcePolicy == cr.Spec.ForProvider.ResourcePolicy, nil
	}

	specPol, err := policyutils.ParsePolicyString(*cr.Spec.ForProvider.ResourcePolicy)
	if err != nil {
		return false, errors.Wrap(err, errParseSpecPolicy)
	}
	curPol, err := policyutils.ParsePolicyString(*res.ResourcePolicy)
	if err != nil {
		return false, errors.Wrap(err, errParseExternalPolicy)
	}
	areEqal, _ := policyutils.ArePoliciesEqal(&specPol, &curPol)
	return areEqal, nil
}

func (e *hooks) isPayloadUpToDate(ctx context.Context, cr *svcapitypes.Secret) (bool, error) {
	s, err := e.client.GetSecretValueWithContext(ctx, &svcsdk.GetSecretValueInput{
		SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return false, errorutils.Wrap(err, errGetSecretValue)
	}
	payload, err := e.getPayload(ctx, &cr.Spec.ForProvider)
	if err != nil {
		return false, err
	}
	switch {
	case pointer.StringValue(s.SecretString) != "":
		return string(payload) == pointer.StringValue(s.SecretString), nil
	case len(s.SecretBinary) != 0:
		return bytes.Equal(payload, s.SecretBinary), nil
	}
	return false, errors.New(errNoAWSValue)
}

func (e *hooks) getPayload(ctx context.Context, params *svcapitypes.SecretParameters) ([]byte, error) {
	ref, err := getSecretRef(params)
	if err != nil {
		return nil, err
	}
	nn := types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	sc := &corev1.Secret{}
	if err := e.kube.Get(ctx, nn, sc); err != nil {
		return nil, err
	}

	if ref.Key != nil {
		val, ok := sc.Data[pointer.StringValue(ref.Key)]
		if !ok {
			return nil, errors.New(fmt.Sprintf(errFmtKeyNotFound, pointer.StringValue(ref.Key)))
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

// getSecretRef returns either params.StringSecretRef, params.BinarySecretRef or an error if none or both of them are set
func getSecretRef(params *svcapitypes.SecretParameters) (*svcapitypes.SecretReference, error) {
	if params.StringSecretRef != nil {
		if params.BinarySecretRef != nil {
			return nil, errors.New(errOnlyOneSecretRef)
		}
		return params.StringSecretRef, nil
	} else if params.BinarySecretRef != nil {
		return params.BinarySecretRef, nil
	}
	return nil, errors.New(errNoSecretRef)
}

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.Secret, obj *svcsdk.UpdateSecretInput) error { //nolint:gocyclo
	resp, err := e.client.DescribeSecretWithContext(ctx, &svcsdk.DescribeSecretInput{
		SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil {
		return errorutils.Wrap(err, errDescribe)
	}
	add, remove := DiffTags(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			TagKeys:  remove,
		}); err != nil {
			return errorutils.Wrap(err, errRemoveTags)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Tags:     add,
		}); err != nil {
			return errorutils.Wrap(err, errCreateTags)
		}
	}

	// Update resource policy
	if cr.Spec.ForProvider.ResourcePolicy != nil {
		_, err := e.client.PutResourcePolicyWithContext(ctx, &svcsdk.PutResourcePolicyInput{
			SecretId:       pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			ResourcePolicy: cr.Spec.ForProvider.ResourcePolicy,
		})
		if err != nil {
			return errors.Wrap(err, errPutResourcePolicy)
		}
	} else {
		_, err := e.client.DeleteResourcePolicyWithContext(ctx, &svcsdk.DeleteResourcePolicyInput{
			SecretId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		})
		if err != nil {
			return errors.Wrap(err, errDeleteResourcePolicy)
		}
	}

	payload, err := e.getPayload(ctx, &cr.Spec.ForProvider)
	if err != nil {
		return err
	}
	switch {
	case cr.Spec.ForProvider.StringSecretRef != nil:
		obj.SecretString = pointer.ToOrNilIfZeroValue(string(payload))
	case cr.Spec.ForProvider.BinarySecretRef != nil:
		obj.SecretBinary = payload
	}
	obj.SecretId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.Description = cr.Spec.ForProvider.Description
	obj.KmsKeyId = cr.Spec.ForProvider.KMSKeyID
	return nil
}

func (e *hooks) preCreate(ctx context.Context, cr *svcapitypes.Secret, obj *svcsdk.CreateSecretInput) error {
	payload, err := e.getPayload(ctx, &cr.Spec.ForProvider)
	if err != nil {
		return err
	}
	switch {
	case cr.Spec.ForProvider.StringSecretRef != nil:
		obj.SecretString = pointer.ToOrNilIfZeroValue(string(payload))
	case cr.Spec.ForProvider.BinarySecretRef != nil:
		obj.SecretBinary = payload
	}
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Secret, obj *svcsdk.DeleteSecretInput) (bool, error) {
	obj.ForceDeleteWithoutRecovery = cr.Spec.ForProvider.ForceDeleteWithoutRecovery
	obj.RecoveryWindowInDays = cr.Spec.ForProvider.RecoveryWindowInDays
	obj.SecretId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, remove []*string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[pointer.StringValue(t.Key)] == pointer.StringValue(t.Value) {
			delete(addMap, pointer.StringValue(t.Key))
			continue
		}
		removeMap[pointer.StringValue(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, &svcsdk.Tag{Key: pointer.ToOrNilIfZeroValue(k), Value: pointer.ToOrNilIfZeroValue(v)})
	}
	for k := range removeMap {
		remove = append(remove, pointer.ToOrNilIfZeroValue(k))
	}
	return
}
