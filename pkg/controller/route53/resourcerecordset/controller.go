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
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/hostedzone"
	"github.com/crossplane/provider-aws/pkg/clients/resourcerecordset"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an ResourceRecordSet resource"
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
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: resourcerecordset.NewClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) resourcerecordset.Client
	awsConfigFn func(context.Context, client.Reader, runtimev1alpha1.Reference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c := conn.newClientFn(awsconfig)

	return &external{kube: conn.client, client: c}, nil
}

type external struct {
	kube   client.Client
	client resourcerecordset.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.ResourceRecordSet)

	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rrset, err := resourcerecordset.GetResourceRecordSet(ctx, e.client, cr.Spec.ForProvider.Name, cr.Spec.ForProvider.ZoneID, cr.Spec.ForProvider.Type, cr.Spec.ForProvider.SetIdentifier)
	if err != nil {
		// Either there is err and retry. Or Resource does not exist.
		return managed.ExternalObservation{
			ResourceExists: false,
		}, errors.Wrap(resource.Ignore(resourcerecordset.IsErrorRRsetNotFound, err), errList)
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	upToDate, err := resourcerecordset.IsUpToDate(cr.Spec.ForProvider, rrset)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errState)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionCreate)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	return managed.ExternalUpdate{}, errors.Wrap(nil, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ResourceRecordSet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionDelete)

	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	// There is no way to confirm 404 (from response) when deleting a recordset
	// which isn't present using ChangeResourceRecordSetRequest
	//
	// For any 404 when deleting, error code returned is nil.
	//So we can safely ignore this and catch any other error.

	return errors.Wrap(resource.Ignore(hostedzone.IsNotFound, err), errDelete)
}
