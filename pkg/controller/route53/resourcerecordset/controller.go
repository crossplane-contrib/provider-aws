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

package resourcerecordset

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/resourcerecordset"
)

const (
	errUnexpectedObject = "The managed resource is not an ResourceRecordSet resource"
	errKubeUpdate       = "failed to update the ResourceRecordSet custom resource"
	errList             = "failed to list the ResourceRecordSet resource"
	errCreate           = "failed to create the ResourceRecordSet resource"
	errUpdate           = "failed to update the ResourceRecordSet resource"
	errDelete           = "failed to delete the ResourceRecordSet resource"
	errState            = "failed to determine resource state"
)

// SetupResourceRecordSet adds a controller that reconciles ResourceRecordSets.
func SetupResourceRecordSet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ResourceRecordSetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ResourceRecordSet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ResourceRecordSetGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: resourcerecordset.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) resourcerecordset.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client resourcerecordset.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rrs, err := resourcerecordset.GetResourceRecordSet(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider, e.client)
	if err != nil {
		// Either there is err and retry. Or Resource does not exist.
		return managed.ExternalObservation{
			ResourceExists: false,
		}, awsclient.Wrap(resource.Ignore(resourcerecordset.IsNotFound, err), errList)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	resourcerecordset.LateInitialize(&cr.Spec.ForProvider, rrs)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdate)
		}
	}

	cr.Status.SetConditions(xpv1.Available())
	upToDate, err := resourcerecordset.IsUpToDate(cr.Spec.ForProvider, *rrs)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errState)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)
	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.client.ChangeResourceRecordSetsRequest(
		resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53.ChangeActionDelete),
	).Send(ctx)

	// There is no way to confirm 404 (from response) when deleting a recordset
	// which isn't present using ChangeResourceRecordSetRequest.
	return awsclient.Wrap(err, errDelete)
}
