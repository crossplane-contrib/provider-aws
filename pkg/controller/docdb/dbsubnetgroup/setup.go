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

package dbsubnetgroup

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errNotDBSubnetGroup = "managed resource is not a DBSubnetGroup custom resource"
	errKubeUpdateFailed = "cannot update DocDBSubnetGroup custom resource"
)

// SetupDBSubnetGroup adds a controller that reconciles a DBSubnetGroup.
func SetupDBSubnetGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBSubnetGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.DBSubnetGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBSubnetGroup{}).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(r)
}

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	h := &hooks{client: e.client, kube: e.kube}
	e.isUpToDate = h.isUpToDate
	e.preUpdate = preUpdate
	e.postUpdate = h.postUpdate
	e.preCreate = preCreate
	e.preDelete = preDelete
	e.filterList = filterList
}

type hooks struct {
	client docdbiface.DocDBAPI
	kube   client.Client
}

func preObserve(_ context.Context, cr *svcapitypes.DBSubnetGroup, obj *svcsdk.DescribeDBSubnetGroupsInput) error {
	obj.DBSubnetGroupName = awsclient.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBSubnetGroup, resp *svcsdk.DescribeDBSubnetGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	cr.SetConditions(v1.Available())
	return obs, nil
}

func (e *hooks) isUpToDate(cr *svcapitypes.DBSubnetGroup, resp *svcsdk.DescribeDBSubnetGroupsOutput) (bool, error) {
	group := resp.DBSubnetGroups[0]

	switch {
	case awsclient.StringValue(cr.Spec.ForProvider.DBSubnetGroupDescription) != awsclient.StringValue(group.DBSubnetGroupDescription),
		!areSubnetsEqual(cr.Spec.ForProvider.SubnetIDs, group.Subnets):
		return false, nil
	}

	return svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, group.DBSubnetGroupArn)
}

func areSubnetsEqual(specSubnetIds []*string, current []*svcsdk.Subnet) bool {
	if len(specSubnetIds) != len(current) {
		return false
	}

	currentMap := make(map[string]*svcsdk.Subnet, len(current))
	for _, s := range current {
		currentMap[awsclient.StringValue(s.SubnetIdentifier)] = s
	}

	for _, spec := range specSubnetIds {
		if _, exists := currentMap[awsclient.StringValue(spec)]; !exists {
			return false
		}
	}

	return true
}

func preUpdate(ctx context.Context, cr *svcapitypes.DBSubnetGroup, obj *svcsdk.ModifyDBSubnetGroupInput) error {
	obj.DBSubnetGroupName = awsclient.String(meta.GetExternalName(cr))
	obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs
	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.DBSubnetGroup, resp *svcsdk.ModifyDBSubnetGroupOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	cr.Status.SetConditions(v1.Available())
	return upd, svcutils.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, resp.DBSubnetGroup.DBSubnetGroupArn)
}

func preCreate(ctx context.Context, cr *svcapitypes.DBSubnetGroup, obj *svcsdk.CreateDBSubnetGroupInput) error {
	obj.DBSubnetGroupName = awsclient.String(meta.GetExternalName(cr))
	obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBSubnetGroup, obj *svcsdk.DeleteDBSubnetGroupInput) (bool, error) {
	obj.DBSubnetGroupName = awsclient.String(meta.GetExternalName(cr))
	return false, nil
}

func filterList(cr *svcapitypes.DBSubnetGroup, list *svcsdk.DescribeDBSubnetGroupsOutput) *svcsdk.DescribeDBSubnetGroupsOutput {
	name := meta.GetExternalName(cr)
	for _, group := range list.DBSubnetGroups {
		if awsclient.StringValue(group.DBSubnetGroupName) == name {
			return &svcsdk.DescribeDBSubnetGroupsOutput{
				Marker:         list.Marker,
				DBSubnetGroups: []*svcsdk.DBSubnetGroup{group},
			}
		}
	}

	return &svcsdk.DescribeDBSubnetGroupsOutput{
		Marker:         list.Marker,
		DBSubnetGroups: []*svcsdk.DBSubnetGroup{},
	}
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.DBSubnetGroup)
	if !ok {
		return errors.New(errNotDBSubnetGroup)
	}

	cr.Spec.ForProvider.Tags = svcutils.AddExternalTags(mg, cr.Spec.ForProvider.Tags)
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
