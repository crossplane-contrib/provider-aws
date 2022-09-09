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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/commonnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupPublicDNSNamespace adds a controller that reconciles PublicDNSNamespaces.
func SetupPublicDNSNamespace(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.PublicDNSNamespaceGroupKind)
	opts := []option{
		func(e *external) {
			h := commonnamespace.NewHooks(e.kube, e.client)
			e.preCreate = preCreate
			e.delete = h.Delete
			e.observe = h.Observe
			e.postCreate = postCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.PublicDNSNamespace{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.PublicDNSNamespaceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preCreate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, obj *svcsdk.CreatePublicDnsNamespaceInput) error {
	obj.CreatorRequestId = awsclient.String(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.PublicDNSNamespace, resp *svcsdk.CreatePublicDnsNamespaceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	cr.SetOperationID(resp.OperationId)
	return cre, err
}
