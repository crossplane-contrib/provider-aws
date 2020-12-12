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

package elasticip

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
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

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an ElasticIP resource"
	errKubeUpdateFailed = "cannot update ElasticIP custom resource"

	errDescribe      = "failed to describe ElasticIP with id"
	errMultipleItems = "retrieved multiple ElasticIPs for the given ElasticIPId"
	errCreate        = "failed to create the ElasticIP resource"
	errCreateTags    = "failed to create tags for the ElasticIP resource"
	errDelete        = "failed to delete the ElasticIP resource"
	errStatusUpdate  = "cannot update status of ElasticIP custom resource"
)

// SetupElasticIP adds a controller that reconciles ElasticIP.
func SetupElasticIP(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ElasticIPGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ElasticIP{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ElasticIPGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ElasticIP)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclients.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: awsec2.New(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.ElasticIPClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.ElasticIP)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	var response *awsec2.DescribeAddressesResponse
	var err error

	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		response, err = e.client.DescribeAddressesRequest(&awsec2.DescribeAddressesInput{
			PublicIps: []string{meta.GetExternalName(cr)},
		}).Send(ctx)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrapf(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDescribe)
		}
	} else {
		response, err = e.client.DescribeAddressesRequest(&awsec2.DescribeAddressesInput{
			AllocationIds: []string{meta.GetExternalName(cr)},
		}).Send(ctx)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrapf(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDescribe)
		}

	}

	// in a successful response, there should be one and only one object
	if len(response.Addresses) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Addresses[0]

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeElasticIP(&cr.Spec.ForProvider, &observed)

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = ec2.GenerateElasticIPObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsElasticIPUpToDate(cr.Spec.ForProvider, observed),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.ElasticIP)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	result, err := e.client.AllocateAddressRequest(&awsec2.AllocateAddressInput{
		Address:               cr.Spec.ForProvider.Address,
		CustomerOwnedIpv4Pool: cr.Spec.ForProvider.CustomerOwnedIPv4Pool,
		Domain:                awsec2.DomainType(aws.StringValue(cr.Spec.ForProvider.Domain)),
		NetworkBorderGroup:    cr.Spec.ForProvider.NetworkBorderGroup,
		PublicIpv4Pool:        cr.Spec.ForProvider.PublicIPv4Pool,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		meta.SetExternalName(cr, aws.StringValue(result.AllocateAddressOutput.PublicIp))
	} else {
		meta.SetExternalName(cr, aws.StringValue(result.AllocateAddressOutput.AllocationId))
	}
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.ElasticIP)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// NOTE: ElasticIPs can only be tagged after the creation and this request
	// is idempotent.
	if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
		Resources: []string{meta.GetExternalName(cr)},
		Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	}).Send(ctx); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ElasticIP)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	var err error
	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		_, err = e.client.ReleaseAddressRequest(&awsec2.ReleaseAddressInput{
			PublicIp: aws.String(meta.GetExternalName(cr)),
		}).Send(ctx)
	} else {
		_, err = e.client.ReleaseAddressRequest(&awsec2.ReleaseAddressInput{
			AllocationId:       aws.String(meta.GetExternalName(cr)),
			NetworkBorderGroup: cr.Spec.ForProvider.NetworkBorderGroup,
		}).Send(ctx)
	}

	return errors.Wrap(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ElasticIP)
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
	cr.Spec.ForProvider.Tags = make([]v1beta1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = v1beta1.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
