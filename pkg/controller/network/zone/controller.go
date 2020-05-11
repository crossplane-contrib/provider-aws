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

package zone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/zone"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an Zone resource"
	errCreate           = "failed to create the Zone resource"
	errDelete           = "failed to delete the Zone resource"
	errUpdate           = "failed to update the Zone resource"
	errGet              = "failed to get the Zone resource"
	errKubeUpdate       = "failed to update the Zone custom resource"
)

// SetupZone adds a controller that reconciles Hosted Zones.
func SetupZone(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.ZoneGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.Zone{}).
		Complete(managed.NewReconciler(
			mgr, resource.ManagedKind(v1alpha3.ZoneGroupVersionKind),
			managed.WithExternalConnecter(
				&connector{client: mgr.GetClient(),
					newClientFn: zone.NewClient,
					awsConfigFn: utils.RetrieveAwsConfigFromProvider},
			),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(
				mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(
				mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) zone.Client
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.Zone)
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
	client zone.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.Zone)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	err := e.kube.Update(ctx, configSanitize(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdate)
	}

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists:    false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	}

	res, err := e.client.GetZoneRequest(&cr.Status.AtProvider.ID).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(zone.IsErrorNoSuchHostedZone, err), errGet)
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	if (cr.Spec.ForProvider.Comment != nil && res.HostedZone.Config.Comment == nil) ||
		(cr.Spec.ForProvider.Comment == nil && res.HostedZone.Config.Comment != nil) ||
		(*cr.Spec.ForProvider.Comment != *res.HostedZone.Config.Comment) {
		return managed.ExternalObservation{
			ResourceExists:    true,
			ResourceUpToDate:  false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha3.Zone)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	res, err := e.client.CreateZoneRequest(cr).Send(ctx)

	if res != nil {
		cr.Status.AtProvider.Update(res.CreateHostedZoneOutput)
	}

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha3.Zone)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateZoneRequest(&cr.Status.AtProvider.ID, cr.Spec.ForProvider.Comment).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.Zone)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteZoneRequest(&cr.Status.AtProvider.ID).Send(ctx)

	return errors.Wrap(resource.Ignore(zone.IsErrorNoSuchHostedZone, err), errDelete)
}

func configSanitize(cr *v1alpha3.Zone) *v1alpha3.Zone {
	if cr.Spec.ForProvider.PrivateZone == nil {
		cr.Spec.ForProvider.PrivateZone = new(bool)
	}

	if cr.Spec.ForProvider.Comment == nil {
		cr.Spec.ForProvider.Comment = new(string)
	}

	return cr
}
