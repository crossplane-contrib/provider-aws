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

package instance

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an Instance resource"
	errKubeUpdateFailed = "cannot update Instance custom resource"

	errDescribe   = "failed to describe Instance with id"
	errCreate     = "failed to create the Instance resource"
	errUpdate     = "failed to update Instance resource"
	errCreateTags = "failed to create tags for the Instance resource"
	errDelete     = "failed to delete the Instance resource"
)

// SetupInstance adds a controller that reconciles Instances.
func SetupInstance(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.InstanceGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Instance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.InstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewInstanceClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.InstanceClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Instance)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.StringValue(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.InstanceClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeInstancesRequest(
		ec2.GenerateDescribeInstancesByExternalTags(resource.GetExternalTags(mgd)),
	).Send(ctx)

	// deleted instances that have not yet been cleaned up from the cluster return a
	// 200 OK with a nil response.Reservations slice
	if err == nil && len(response.Reservations) == 0 {
		return managed.ExternalObservation{}, nil
	}

	if err != nil {
		return managed.ExternalObservation{},
			awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDescribe)
	}

	observed := response.Reservations[0].Instances

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()

	// o := awsec2.DescribeInstanceAttributeOutput{}

	// ec2.LateInitializeInstance(&cr.Spec.ForProvider, &observed, &o)

	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	observation := ec2.GenerateInstanceObservation(observed)
	condition := ec2.GenerateInstanceCondition(observation)

	switch condition {
	case ec2.Creating:
		cr.SetConditions(xpv1.Creating())
	case ec2.Available:
		cr.SetConditions(xpv1.Available())
	case ec2.Deleting:
		cr.SetConditions(xpv1.Deleting())
	case ec2.Deleted:
		// Terminated instances remain visible on API calls for a time before
		// being automatically deleted. Rather than having the delete command
		// hang for that entire time, return an empty ExternalObservation in
		// this case.
		//
		// ref: (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html)
		return managed.ExternalObservation{}, nil
	}

	cr.Status.AtProvider = observation

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
		// ResourceUpToDate:        ec2.IsInstanceUpToDate(cr.Spec.ForProvider, observed, o),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.RunInstancesRequest(
		ec2.GenerateEC2RunInstancesInput(mgd.GetName(), &cr.Spec.ForProvider),
	).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	instanceIds := make([]string, len(result.Instances))
	for i, ins := range result.Instances {
		instanceIds[i] = *ins.InstanceId
	}

	// specify the tags post runInstances request so that we can maintain a way
	// to track the resources as a group
	if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
		Resources: instanceIds,
		Tags:      manualv1alpha1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	}).Send(ctx); err != nil {
		return managed.ExternalCreation{ExternalNameAssigned: false}, awsclient.Wrap(err, errCreateTags)
	}

	// instance count could be greater than 1.
	// use the name from metadata.Name to define the externalName
	meta.SetExternalName(cr, mgd.GetName())

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	return managed.ExternalUpdate{}, awsclient.Wrap(errors.New("fix this"), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := e.client.DescribeInstancesRequest(
		ec2.GenerateDescribeInstancesByExternalTags(resource.GetExternalTags(mgd)),
	).Send(ctx)

	if err != nil {
		return awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDelete)
	}

	instances := response.Reservations[0].Instances

	instanceIds := make([]string, len(instances))
	for i, ins := range instances {
		instanceIds[i] = *ins.InstanceId
	}

	_, err = e.client.TerminateInstancesRequest(&awsec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*manualv1alpha1.Instance)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]manualv1alpha1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = manualv1alpha1.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
