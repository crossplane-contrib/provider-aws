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

package publicdnsnamespace

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	clientsvcdk "github.com/crossplane-contrib/provider-aws/pkg/clients/servicediscovery"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/commonnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupPublicDNSNamespace adds a controller that reconciles PublicDNSNamespaces.
func SetupPublicDNSNamespace(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.PublicDNSNamespaceGroupKind)
	opts := []option{
		func(e *external) {
			h := commonnamespace.NewHooks(e.kube, e.client)
			hL := &hooks{client: e.client}
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preUpdate = preUpdate
			e.postUpdate = hL.postUpdate
			e.delete = h.Delete
			e.observe = h.Observe

		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.PublicDNSNamespaceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.PublicDNSNamespace{}).
		Complete(r)
}

type hooks struct {
	client clientsvcdk.Client
}

func preCreate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, obj *svcsdk.CreatePublicDnsNamespaceInput) error {
	obj.CreatorRequestId = pointer.ToOrNilIfZeroValue(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, resp *svcsdk.CreatePublicDnsNamespaceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	cr.SetOperationID(resp.OperationId)
	return cre, err
}

func preUpdate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, obj *svcsdk.UpdatePublicDnsNamespaceInput) error {
	obj.UpdaterRequestId = pointer.ToOrNilIfZeroValue(string(cr.UID))
	obj.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// Description and TTL are required
	obj.Namespace = &svcsdk.PublicDnsNamespaceChange{
		Description: cr.GetDescription(),
	}

	// Namespace.Description are required for update publicdnsnamespace
	// Set an empty string if Description are nil
	if cr.GetDescription() == nil {
		var tmpEmpty = ""
		obj.Namespace.Description = &tmpEmpty
	}

	if cr.GetTTL() != nil {
		obj.Namespace.Properties = &svcsdk.PublicDnsNamespacePropertiesChange{
			DnsProperties: &svcsdk.PublicDnsPropertiesMutableChange{
				SOA: &svcsdk.SOAChange{
					TTL: cr.GetTTL(),
				},
			},
		}
	}

	return nil
}

func (e *hooks) postUpdate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, resp *svcsdk.UpdatePublicDnsNamespaceOutput, cre managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return cre, err
	}
	cr.Status.SetConditions(v1.Available())

	// Update Tags
	return cre, commonnamespace.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, cr)
}
