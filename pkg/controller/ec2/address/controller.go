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

package address

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an Address resource"
	errKubeUpdateFailed = "cannot update Address custom resource"

	errDescribe      = "failed to describe Address with id"
	errMultipleItems = "retrieved multiple Addresss for the given AddressId"
	errCreate        = "failed to create the Address resource"
	errCreateTags    = "failed to create tags for the Address resource"
	errDelete        = "failed to delete the Address resource"
	errStatusUpdate  = "cannot update status of Address custom resource"
)

// SetupAddress adds a controller that reconciles Address.
func SetupAddress(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.AddressGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithCreationGracePeriod(3 * time.Minute),
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

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.AddressGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Address{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Address)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: awsec2.NewFromConfig(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.AddressClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.Address)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	var output *awsec2.DescribeAddressesOutput
	var err error

	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		output, err = e.client.DescribeAddresses(ctx, &awsec2.DescribeAddressesInput{
			PublicIps: []string{meta.GetExternalName(cr)},
		})
		if err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDescribe)
		}
	} else {
		output, err = e.client.DescribeAddresses(ctx, &awsec2.DescribeAddressesInput{
			AllocationIds: []string{meta.GetExternalName(cr)},
		})
		if err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDescribe)
		}

	}

	// in a successful response, there should be one and only one object
	if len(output.Addresses) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := output.Addresses[0]

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeAddress(&cr.Spec.ForProvider, &observed)

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = ec2.GenerateAddressObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsAddressUpToDate(cr.Spec.ForProvider, observed),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Address)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	result, err := e.client.AllocateAddress(ctx, &awsec2.AllocateAddressInput{
		Address:               cr.Spec.ForProvider.Address,
		CustomerOwnedIpv4Pool: cr.Spec.ForProvider.CustomerOwnedIPv4Pool,
		Domain:                awsec2types.DomainType(aws.ToString(cr.Spec.ForProvider.Domain)),
		NetworkBorderGroup:    cr.Spec.ForProvider.NetworkBorderGroup,
		PublicIpv4Pool:        cr.Spec.ForProvider.PublicIPv4Pool,
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		meta.SetExternalName(cr, aws.ToString(result.PublicIp))
	} else {
		meta.SetExternalName(cr, aws.ToString(result.AllocationId))
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Address)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// NOTE: Addresss can only be tagged after the creation and this request
	// is idempotent.
	if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
		Resources: []string{meta.GetExternalName(cr)},
		Tags:      ec2.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	}); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreateTags)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Address)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	var err error
	if ec2.IsStandardDomain(cr.Spec.ForProvider) {
		_, err = e.client.ReleaseAddress(ctx, &awsec2.ReleaseAddressInput{
			PublicIp: aws.String(meta.GetExternalName(cr)),
		})
	} else {
		_, err = e.client.ReleaseAddress(ctx, &awsec2.ReleaseAddressInput{
			AllocationId:       aws.String(meta.GetExternalName(cr)),
			NetworkBorderGroup: cr.Spec.ForProvider.NetworkBorderGroup,
		})
	}

	return errorutils.Wrap(resource.Ignore(ec2.IsAddressNotFoundErr, err), errDelete)
}
