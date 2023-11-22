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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/hostedzone"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an Hosted Zone resource"

	errCreate = "failed to create the Hosted Zone resource"
	errDelete = "failed to delete the Hosted Zone resource"
	errUpdate = "failed to update the Hosted Zone resource"
	errGet    = "failed to get the Hosted Zone resource"

	errListTags   = "cannot list tags"
	errUpdateTags = "cannot update tags"
)

// SetupHostedZone adds a controller that reconciles Hosted Zones.
func SetupHostedZone(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(route53v1alpha1.HostedZoneGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: hostedzone.NewClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(
		mgr, resource.ManagedKind(route53v1alpha1.HostedZoneGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&route53v1alpha1.HostedZone{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) hostedzone.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client hostedzone.Client

	tagsToAdd    []route53types.Tag
	tagsToRemove []string
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

	hostedZoneID := aws.String(hostedzone.GetHostedZoneID(cr))
	res, err := e.client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: hostedZoneID,
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(hostedzone.IsNotFound, err), errGet)
	}

	resTags, err := e.client.ListTagsForResource(ctx, &route53.ListTagsForResourceInput{
		ResourceId:   aws.String(meta.GetExternalName(cr)), // id w/o prefix
		ResourceType: route53types.TagResourceTypeHostedzone,
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errListTags)
	}
	if resTags.ResourceTagSet == nil {
		resTags.ResourceTagSet = &route53types.ResourceTagSet{}
	}

	var areTagsUpToDate bool
	e.tagsToAdd, e.tagsToRemove, areTagsUpToDate = hostedzone.AreTagsUpToDate(cr.Spec.ForProvider.Tags, resTags.ResourceTagSet.Tags)

	current := cr.Spec.ForProvider.DeepCopy()
	hostedzone.LateInitialize(&cr.Spec.ForProvider, res)

	cr.Status.AtProvider = hostedzone.GenerateObservation(res)
	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        hostedzone.IsUpToDate(cr.Spec.ForProvider, *res.HostedZone) && areTagsUpToDate,
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
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
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

	hostedZoneID := hostedzone.GetHostedZoneID(cr)
	_, err := e.client.UpdateHostedZoneComment(ctx,
		hostedzone.GenerateUpdateHostedZoneCommentInput(cr.Spec.ForProvider, hostedZoneID),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	// Update tags if necessary
	if len(e.tagsToAdd) > 0 || len(e.tagsToRemove) > 0 {

		changeTagsInput := &route53.ChangeTagsForResourceInput{
			ResourceId:   aws.String(meta.GetExternalName(cr)), // id w/o prefix
			ResourceType: route53types.TagResourceTypeHostedzone,
		}

		// AWS throws error when provided AddTags or RemoveTagKeys are empty lists
		if len(e.tagsToAdd) > 0 {
			changeTagsInput.AddTags = e.tagsToAdd
		}
		if len(e.tagsToRemove) > 0 {
			changeTagsInput.RemoveTagKeys = e.tagsToRemove
		}

		_, err := e.client.ChangeTagsForResource(ctx, changeTagsInput)
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdateTags)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*route53v1alpha1.HostedZone)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: aws.String(hostedzone.GetHostedZoneID(cr)),
	})

	return errorutils.Wrap(resource.Ignore(hostedzone.IsNotFound, err), errDelete)
}
