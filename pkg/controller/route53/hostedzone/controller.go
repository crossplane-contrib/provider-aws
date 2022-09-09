/*
Copyright 2020 The Crossplane Authors.

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

package hostedzone

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/hostedzone"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUnexpectedObject = "The managed resource is not an Hosted Zone resource"

	errCreate = "failed to create the Hosted Zone resource"
	errDelete = "failed to delete the Hosted Zone resource"
	errUpdate = "failed to update the Hosted Zone resource"
	errGet    = "failed to get the Hosted Zone resource"
)

// SetupHostedZone adds a controller that reconciles Hosted Zones.
func SetupHostedZone(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(route53v1alpha1.HostedZoneGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&route53v1alpha1.HostedZone{}).
		Complete(managed.NewReconciler(
			mgr, resource.ManagedKind(route53v1alpha1.HostedZoneGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: hostedzone.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) hostedzone.Client
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
	client hostedzone.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*route53v1alpha1.HostedZone)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	res, err := e.client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: aws.String(fmt.Sprintf("%s%s", hostedzone.IDPrefix, meta.GetExternalName(cr))),
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(hostedzone.IsNotFound, err), errGet)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	hostedzone.LateInitialize(&cr.Spec.ForProvider, res)

	cr.Status.AtProvider = hostedzone.GenerateObservation(res)
	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        hostedzone.IsUpToDate(cr.Spec.ForProvider, *res.HostedZone),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*route53v1alpha1.HostedZone)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	res, err := e.client.CreateHostedZone(ctx, hostedzone.GenerateCreateHostedZoneInput(cr))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}
	id := strings.SplitAfter(aws.ToString(res.HostedZone.Id), hostedzone.IDPrefix)
	if len(id) < 2 {
		return managed.ExternalCreation{}, errors.Wrap(errors.New("returned id does not contain /hostedzone/ prefix"), errCreate)
	}
	meta.SetExternalName(cr, id[1])
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*route53v1alpha1.HostedZone)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateHostedZoneComment(ctx,
		hostedzone.GenerateUpdateHostedZoneCommentInput(cr.Spec.ForProvider, fmt.Sprintf("%s%s", hostedzone.IDPrefix, meta.GetExternalName(cr))),
	)

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*route53v1alpha1.HostedZone)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: aws.String(fmt.Sprintf("%s%s", hostedzone.IDPrefix, meta.GetExternalName(cr))),
	})

	return awsclient.Wrap(resource.Ignore(hostedzone.IsNotFound, err), errDelete)
}
