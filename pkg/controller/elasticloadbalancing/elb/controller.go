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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha1 "github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/elasticloadbalancing/elb"
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
func SetupELB(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ELBGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ELB{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ELBGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: elb.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) elb.Client
	awsConfigFn func(context.Context, client.Client, resource.Managed, string) (*aws.Config, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := c.awsConfigFn(ctx, c.kube, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client elb.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.ELB)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeLoadBalancersRequest(&awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	tagsResponse, err := e.client.DescribeTagsRequest(&awselb.DescribeTagsInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribeTags)
	}

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	elb.LateInitializeELB(&cr.Spec.ForProvider, &observed, tagsResponse.TagDescriptions[0].Tags)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errSpecUpdate)
		}
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

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
	cr, ok := mgd.(*v1alpha1.ELB)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	_, err := e.client.CreateLoadBalancerRequest(elb.GenerateCreateELBInput(meta.GetExternalName(cr),
		cr.Spec.ForProvider)).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.ELB)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeLoadBalancersRequest(&awselb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errUpdate)
	}

	if len(response.LoadBalancerDescriptions) != 1 {
		return managed.ExternalUpdate{}, errors.New(errMultipleItems)
	}

	observed := response.LoadBalancerDescriptions[0]

	tagsResponse, err := e.client.DescribeTagsRequest(&awselb.DescribeTagsInput{
		LoadBalancerNames: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDescribeTags)
	}

	// AWS ELB API doesn't have a single PUT/PATCH API.
	// Hence, create a patch to figure which fields are to be updated.
	patch, err := elb.CreatePatch(observed, cr.Spec.ForProvider, tagsResponse.TagDescriptions[0].Tags)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errUpdate)
	}

	if len(patch.AvailabilityZones) != 0 {
		if err := e.updateAvailabilityZones(ctx, cr.Spec.ForProvider.AvailabilityZones, observed.AvailabilityZones, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if len(patch.SecurityGroupIDs) != 0 {
		if _, err := e.client.ApplySecurityGroupsToLoadBalancerRequest(&awselb.ApplySecurityGroupsToLoadBalancerInput{
			SecurityGroups:   cr.Spec.ForProvider.SecurityGroupIDs,
			LoadBalancerName: aws.String(meta.GetExternalName(cr)),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if len(patch.SubnetIDs) != 0 {
		if err := e.updateSubnets(ctx, cr.Spec.ForProvider.SubnetIDs, observed.Subnets, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if patch.HealthCheck != nil {
		if _, err := e.client.ConfigureHealthCheckRequest(&awselb.ConfigureHealthCheckInput{
			LoadBalancerName: aws.String(meta.GetExternalName(cr)),
			HealthCheck: &awselb.HealthCheck{
				HealthyThreshold:   aws.Int64(cr.Spec.ForProvider.HealthCheck.HealthyThreshold),
				Interval:           aws.Int64(cr.Spec.ForProvider.HealthCheck.Interval),
				Target:             aws.String(cr.Spec.ForProvider.HealthCheck.Target),
				Timeout:            aws.Int64(cr.Spec.ForProvider.HealthCheck.Timeout),
				UnhealthyThreshold: aws.Int64(cr.Spec.ForProvider.HealthCheck.HealthyThreshold),
			},
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if len(patch.Listeners) != 0 {
		if err := e.updateListeners(ctx, cr.Spec.ForProvider.Listeners, observed.ListenerDescriptions, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if len(patch.Tags) != 0 {
		if err := e.updateTags(ctx, cr.Spec.ForProvider.Tags, tagsResponse.TagDescriptions[0].Tags, meta.GetExternalName(cr)); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ELB)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteLoadBalancerRequest(&awselb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(elb.IsELBNotFound, err), errDelete)
}

func (e *external) updateAvailabilityZones(ctx context.Context, zones, elbZones []string, name string) error {

	addZones := stringSliceDiff(zones, elbZones)
	if len(addZones) != 0 {
		if _, err := e.client.EnableAvailabilityZonesForLoadBalancerRequest(&awselb.EnableAvailabilityZonesForLoadBalancerInput{
			AvailabilityZones: addZones,
			LoadBalancerName:  aws.String(name),
		}).Send(ctx); err != nil {
			return err
		}
	}

	removeZones := stringSliceDiff(elbZones, zones)
	if len(removeZones) != 0 {
		if _, err := e.client.DisableAvailabilityZonesForLoadBalancerRequest(&awselb.DisableAvailabilityZonesForLoadBalancerInput{
			AvailabilityZones: removeZones,
			LoadBalancerName:  aws.String(name),
		}).Send(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateSubnets(ctx context.Context, subnets, elbSubnets []string, name string) error {

	addSubnets := stringSliceDiff(subnets, elbSubnets)
	if len(addSubnets) != 0 {
		if _, err := e.client.AttachLoadBalancerToSubnetsRequest(&awselb.AttachLoadBalancerToSubnetsInput{
			LoadBalancerName: aws.String(name),
			Subnets:          addSubnets,
		}).Send(ctx); err != nil {
			return err
		}
	}

	removeSubnets := stringSliceDiff(elbSubnets, subnets)
	if len(elbSubnets) != 0 {
		if _, err := e.client.DetachLoadBalancerFromSubnetsRequest(&awselb.DetachLoadBalancerFromSubnetsInput{
			LoadBalancerName: aws.String(name),
			Subnets:          removeSubnets,
		}).Send(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateListeners(ctx context.Context, listeners []v1alpha1.Listener, elbListeners []awselb.ListenerDescription, name string) error {

	if len(elbListeners) != 0 {
		ports := []int64{}
		for _, v := range elbListeners {
			ports = append(ports, aws.Int64Value(v.Listener.LoadBalancerPort))
		}

		if _, err := e.client.DeleteLoadBalancerListenersRequest(&awselb.DeleteLoadBalancerListenersInput{
			LoadBalancerName:  aws.String(name),
			LoadBalancerPorts: ports,
		}).Send(ctx); err != nil {
			return err
		}
	}

	if len(listeners) != 0 {
		if _, err := e.client.CreateLoadBalancerListenersRequest(&awselb.CreateLoadBalancerListenersInput{
			Listeners:        elb.BuildELBListeners(listeners),
			LoadBalancerName: aws.String(name),
		}).Send(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (e *external) updateTags(ctx context.Context, tags []v1alpha1.Tag, elbTags []awselb.Tag, name string) error {

	if len(elbTags) > 0 {
		keysOnly := make([]awselb.TagKeyOnly, len(elbTags))
		for i, v := range elbTags {
			keysOnly[i] = awselb.TagKeyOnly{Key: v.Key}
		}
		if _, err := e.client.RemoveTagsRequest(&awselb.RemoveTagsInput{
			LoadBalancerNames: []string{name},
			Tags:              keysOnly,
		}).Send(ctx); err != nil {
			return err
		}
	}

	if len(tags) > 0 {
		if _, err := e.client.AddTagsRequest(&awselb.AddTagsInput{
			LoadBalancerNames: []string{name},
			Tags:              elb.BuildELBTags(tags),
		}).Send(ctx); err != nil {
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
