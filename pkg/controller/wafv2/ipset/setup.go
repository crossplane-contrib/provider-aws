/*
   Copyright 2026 The Crossplane Authors.
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

package ipset

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/wafv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/wafv2/wafv2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/wafv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	tagutils "github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errCreateSession = "cannot create a new session"
	errDescribe      = "failed to describe IPSet"
	errCreate        = "failed to create IPSet"
	errUpdate        = "failed to update IPSet"
	errDelete        = "failed to delete IPSet"
)

// SetupIPSet adds a controller that reconciles IPSet.
func SetupIPSet(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.IPSetKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.IPSetGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.IPSet{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, cr *svcapitypes.IPSet) (managed.TypedExternalClient[*svcapitypes.IPSet], error) {
	sess, err := connectaws.GetConfigV1(ctx, c.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return &external{kube: c.kube, client: svcsdk.New(sess)}, nil
}

type external struct {
	kube   client.Client
	client svcsdkapi.WAFV2API
}

func (e *external) Disconnect(_ context.Context) error {
	return nil
}

func (e *external) Observe(ctx context.Context, cr *svcapitypes.IPSet) (managed.ExternalObservation, error) {
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// List IP sets to find our resource by name
	listInput := &svcsdk.ListIPSetsInput{Scope: cr.Spec.ForProvider.Scope}
	listOutput, err := e.client.ListIPSetsWithContext(ctx, listInput)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	var ipSetID *string
	for _, summary := range listOutput.IPSets {
		if aws.StringValue(summary.Name) == meta.GetExternalName(cr) {
			ipSetID = summary.Id
			break
		}
	}
	if ipSetID == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	getInput := &svcsdk.GetIPSetInput{
		Name:  aws.String(meta.GetExternalName(cr)),
		Scope: cr.Spec.ForProvider.Scope,
		Id:    ipSetID,
	}
	resp, err := e.client.GetIPSetWithContext(ctx, getInput)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(resource.Ignore(IsNotFound, err), errDescribe)
	}

	cr.Status.AtProvider.ARN = resp.IPSet.ARN
	cr.Status.AtProvider.ID = resp.IPSet.Id
	cr.Status.AtProvider.LockToken = resp.LockToken
	cr.SetConditions(xpv1.Available())

	upToDate, diff, err := e.isUpToDate(cr, resp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		Diff:             diff,
	}, nil
}

// error return kept for codebase convention consistency (see webacl isUpToDate)
func (e *external) isUpToDate(cr *svcapitypes.IPSet, resp *svcsdk.GetIPSetOutput) (bool, string, error) { //nolint:unparam
	if !addressesEqual(cr.Spec.ForProvider.Addresses, resp.IPSet.Addresses) {
		return false, "addresses differ", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Description) != aws.StringValue(resp.IPSet.Description) {
		return false, "description differs", nil
	}

	return true, "", nil
}

// addressesEqual compares two []*string slices as sets (order-independent).
func addressesEqual(a, b []*string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[aws.StringValue(v)] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[aws.StringValue(v)]; !ok {
			return false
		}
	}
	return true
}

func (e *external) Create(ctx context.Context, cr *svcapitypes.IPSet) (managed.ExternalCreation, error) {
	input := &svcsdk.CreateIPSetInput{
		Name:             aws.String(meta.GetExternalName(cr)),
		Scope:            cr.Spec.ForProvider.Scope,
		IPAddressVersion: cr.Spec.ForProvider.IPAddressVersion,
		Addresses:        cr.Spec.ForProvider.Addresses,
		Description:      cr.Spec.ForProvider.Description,
	}

	if len(cr.Spec.ForProvider.Tags) > 0 {
		var tags []*svcsdk.Tag
		for _, t := range cr.Spec.ForProvider.Tags {
			tags = append(tags, &svcsdk.Tag{Key: t.Key, Value: t.Value})
		}
		input.Tags = tags
	}

	resp, err := e.client.CreateIPSetWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	cr.Status.AtProvider.ARN = resp.Summary.ARN
	cr.Status.AtProvider.ID = resp.Summary.Id
	cr.Status.AtProvider.LockToken = resp.Summary.LockToken

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, cr *svcapitypes.IPSet) (managed.ExternalUpdate, error) {
	input := &svcsdk.UpdateIPSetInput{
		Name:        aws.String(meta.GetExternalName(cr)),
		Scope:       cr.Spec.ForProvider.Scope,
		Id:          cr.Status.AtProvider.ID,
		LockToken:   cr.Status.AtProvider.LockToken,
		Addresses:   cr.Spec.ForProvider.Addresses,
		Description: cr.Spec.ForProvider.Description,
	}

	resp, err := e.client.UpdateIPSetWithContext(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	cr.Status.AtProvider.LockToken = resp.NextLockToken

	// Handle tags
	if err := e.reconcileTags(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, cr *svcapitypes.IPSet) (managed.ExternalDelete, error) {
	input := &svcsdk.DeleteIPSetInput{
		Name:      aws.String(meta.GetExternalName(cr)),
		Scope:     cr.Spec.ForProvider.Scope,
		Id:        cr.Status.AtProvider.ID,
		LockToken: cr.Status.AtProvider.LockToken,
	}

	_, err := e.client.DeleteIPSetWithContext(ctx, input)
	return managed.ExternalDelete{}, errors.Wrap(err, errDelete)
}

func (e *external) reconcileTags(ctx context.Context, cr *svcapitypes.IPSet) error {
	listResp, err := e.client.ListTagsForResourceWithContext(ctx, &svcsdk.ListTagsForResourceInput{
		ResourceARN: cr.Status.AtProvider.ARN,
	})
	if err != nil {
		return err
	}

	desired := map[string]*string{}
	observed := map[string]*string{}

	for _, t := range cr.Spec.ForProvider.Tags {
		desired[aws.StringValue(t.Key)] = t.Value
	}
	for _, t := range listResp.TagInfoForResource.TagList {
		observed[aws.StringValue(t.Key)] = t.Value
	}

	toAdd, toRemove := tagutils.DiffTagsMapPtr(desired, observed)

	if len(toAdd) > 0 {
		var tags []*svcsdk.Tag
		for k, v := range toAdd {
			tags = append(tags, &svcsdk.Tag{Key: aws.String(k), Value: v})
		}
		_, err = e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			ResourceARN: cr.Status.AtProvider.ARN,
			Tags:        tags,
		})
		if err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		_, err = e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			ResourceARN: cr.Status.AtProvider.ARN,
			TagKeys:     toRemove,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// IsNotFound returns whether the error is a not-found error
func IsNotFound(err error) bool {
	return strings.Contains(err.Error(), svcsdk.ErrCodeWAFNonexistentItemException)
}
