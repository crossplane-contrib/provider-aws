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

package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awselbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
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

	elasticloadbalancingv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/elasticloadbalancing/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticloadbalancing/elb"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an ELB resource"

	errDescribe      = "cannot describe ELB with given name"
	errDescribeTags  = "cannot describe tags for ELB with given name"
	errMultipleItems = "retrieved multiple ELBs for the given name"
	errCreate        = "cannot create the ELB resource"
	errUpdate        = "cannot update ELB resource"
	errDelete        = "cannot delete the ELB resource"
	errSpecUpdate    = "cannot update spec of ELB custom resource"
	errUpToDate      = "cannot check if the resource is up to date"
)

// SetupELB adds a controller that reconciles ELBs.
func SetupELB(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(elasticloadbalancingv1alpha1.ELBGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elb.NewClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(elasticloadbalancingv1alpha1.ELBGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&elasticloadbalancingv1alpha1.ELB{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elb.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*elasticloadbalancingv1alpha1.ELB)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client elb.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*elasticloadbalancingv1alpha1.ELB)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeLoadBalancers(ctx, &awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	tagsResponse, err := e.client.DescribeTags(ctx, &awselb.DescribeTagsInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribeTags)
	}

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	elb.LateInitializeELB(&cr.Spec.ForProvider, &observed, tagsResponse.TagDescriptions[0].Tags)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errSpecUpdate)
		}
	}

	cr.Status.SetConditions(xpv1.Available())

	cr.Status.AtProvider = elb.GenerateELBObservation(observed)

	upToDate, err := elb.IsUpToDate(cr.Spec.ForProvider, observed, tagsResponse.TagDescriptions[0].Tags)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*elasticloadbalancingv1alpha1.ELB)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateLoadBalancer(ctx, elb.GenerateCreateELBInput(meta.GetExternalName(cr),
		cr.Spec.ForProvider))

	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // //nolint:gocyclo
	cr, ok := mgd.(*elasticloadbalancingv1alpha1.ELB)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeLoadBalancers(ctx, &awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errUpdate)
	}

	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalUpdate{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	tagsResponse, err := e.client.DescribeTags(ctx, &awselb.DescribeTagsInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribeTags)
	}

	// AWS ELB API doesn't have a single PUT/PATCH API.
	// Hence, create a patch to figure which fields are to be updated.
	patch, err := elb.CreatePatch(observed, cr.Spec.ForProvider, tagsResponse.TagDescriptions[0].Tags)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errUpdate)
	}

	if len(patch.AvailabilityZones) != 0 {
		if err := e.updateAvailabilityZones(ctx, cr.Spec.ForProvider.AvailabilityZones, observed.AvailabilityZones, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if len(patch.SecurityGroupIDs) != 0 {
		if _, err := e.client.ApplySecurityGroupsToLoadBalancer(ctx, &awselb.ApplySecurityGroupsToLoadBalancerInput{
			SecurityGroups:   cr.Spec.ForProvider.SecurityGroupIDs,
			LoadBalancerName: aws.String(meta.GetExternalName(cr)),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if len(patch.SubnetIDs) != 0 {
		if err := e.updateSubnets(ctx, cr.Spec.ForProvider.SubnetIDs, observed.Subnets, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if patch.HealthCheck != nil {
		if _, err := e.client.ConfigureHealthCheck(ctx, &awselb.ConfigureHealthCheckInput{
			LoadBalancerName: aws.String(meta.GetExternalName(cr)),
			HealthCheck: &awselbtypes.HealthCheck{
				HealthyThreshold:   cr.Spec.ForProvider.HealthCheck.HealthyThreshold,
				Interval:           cr.Spec.ForProvider.HealthCheck.Interval,
				Target:             aws.String(cr.Spec.ForProvider.HealthCheck.Target),
				Timeout:            cr.Spec.ForProvider.HealthCheck.Timeout,
				UnhealthyThreshold: cr.Spec.ForProvider.HealthCheck.HealthyThreshold,
			},
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if len(patch.Listeners) != 0 {
		if err := e.updateListeners(ctx, cr.Spec.ForProvider.Listeners, observed.ListenerDescriptions, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if len(patch.Tags) != 0 {
		if err := e.updateTags(ctx, cr.Spec.ForProvider.Tags, tagsResponse.TagDescriptions[0].Tags, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*elasticloadbalancingv1alpha1.ELB)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteLoadBalancer(ctx, &awselb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDelete)
}

func (e *external) updateAvailabilityZones(ctx context.Context, zones, elbZones []string, name string) error {

	addZones := stringSliceDiff(zones, elbZones)
	if len(addZones) != 0 {
		if _, err := e.client.EnableAvailabilityZonesForLoadBalancer(ctx, &awselb.EnableAvailabilityZonesForLoadBalancerInput{
			AvailabilityZones: addZones,
			LoadBalancerName:  aws.String(name),
		}); err != nil {
			return err
		}
	}

	removeZones := stringSliceDiff(elbZones, zones)
	if len(removeZones) != 0 {
		if _, err := e.client.DisableAvailabilityZonesForLoadBalancer(ctx, &awselb.DisableAvailabilityZonesForLoadBalancerInput{
			AvailabilityZones: removeZones,
			LoadBalancerName:  aws.String(name),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateSubnets(ctx context.Context, subnets, elbSubnets []string, name string) error {

	addSubnets := stringSliceDiff(subnets, elbSubnets)
	if len(addSubnets) != 0 {
		if _, err := e.client.AttachLoadBalancerToSubnets(ctx, &awselb.AttachLoadBalancerToSubnetsInput{
			LoadBalancerName: aws.String(name),
			Subnets:          addSubnets,
		}); err != nil {
			return err
		}
	}

	removeSubnets := stringSliceDiff(elbSubnets, subnets)
	if len(elbSubnets) != 0 {
		if _, err := e.client.DetachLoadBalancerFromSubnets(ctx, &awselb.DetachLoadBalancerFromSubnetsInput{
			LoadBalancerName: aws.String(name),
			Subnets:          removeSubnets,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateListeners(ctx context.Context, listeners []elasticloadbalancingv1alpha1.Listener, elbListeners []awselbtypes.ListenerDescription, name string) error {

	if len(elbListeners) != 0 {
		ports := []int32{}
		for _, v := range elbListeners {
			ports = append(ports, v.Listener.LoadBalancerPort)
		}

		if _, err := e.client.DeleteLoadBalancerListeners(ctx, &awselb.DeleteLoadBalancerListenersInput{
			LoadBalancerName:  aws.String(name),
			LoadBalancerPorts: ports,
		}); err != nil {
			return err
		}
	}

	if len(listeners) != 0 {
		if _, err := e.client.CreateLoadBalancerListeners(ctx, &awselb.CreateLoadBalancerListenersInput{
			Listeners:        elb.BuildELBListeners(listeners),
			LoadBalancerName: aws.String(name),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateTags(ctx context.Context, tags []elasticloadbalancingv1alpha1.Tag, elbTags []awselbtypes.Tag, name string) error {

	if len(elbTags) > 0 {
		keysOnly := make([]awselbtypes.TagKeyOnly, len(elbTags))
		for i, v := range elbTags {
			keysOnly[i] = awselbtypes.TagKeyOnly{Key: v.Key}
		}
		if _, err := e.client.RemoveTags(ctx, &awselb.RemoveTagsInput{
			LoadBalancerNames: []string{name},
			Tags:              keysOnly,
		}); err != nil {
			return err
		}
	}

	if len(tags) > 0 {
		if _, err := e.client.AddTags(ctx, &awselb.AddTagsInput{
			LoadBalancerNames: []string{name},
			Tags:              elb.BuildELBTags(tags),
		}); err != nil {
			return err
		}
	}

	return nil
}

// stringSliceDiff generate a difference between given string slices a and b.
func stringSliceDiff(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
