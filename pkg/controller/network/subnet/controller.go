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

package subnet

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha3 "github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject    = "The managed resource is not an Subnet resource"
	errClient              = "cannot create a new SubnetClient"
	errDescribe            = "failed to describe Subnet with id"
	errMultipleItems       = "retrieved multiple Subnet for the given subnetId"
	errCreate              = "failed to create the Subnet resource"
	errPersistExternalName = "failed to persist InternetGateway ID"
	errDelete              = "failed to delete the Subnet resource"
)

// SetupSubnet adds a controller that reconciles Subnets.
func SetupSubnet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.SubnetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.Subnet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.SubnetGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewSubnetClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (ec2.SubnetClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.Subnet)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}

	return &external{kube: conn.client, client: c}, nil
}

type external struct {
	kube   client.Client
	client ec2.SubnetClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.Subnet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// AWS network resources are uniquely identified by an ID that is returned
	// on create time; we can't tell whether they exist unless we have recorded
	// their ID.
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	req := e.client.DescribeSubnetsRequest(&awsec2.DescribeSubnetsInput{
		SubnetIds: []string{meta.GetExternalName(cr)},
	})

	response, err := req.Send(ctx)
	if ec2.IsSubnetNotFoundErr(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Subnets) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Subnets[0]

	if observed.State == awsec2.SubnetStateAvailable {
		cr.SetConditions(runtimev1alpha1.Available())
	} else if observed.State == awsec2.SubnetStatePending {
		cr.SetConditions(runtimev1alpha1.Creating())
	}

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.Subnet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.CreateSubnetRequest(&awsec2.CreateSubnetInput{
		VpcId:            aws.String(cr.Spec.VPCID),
		AvailabilityZone: aws.String(cr.Spec.AvailabilityZone),
		CidrBlock:        aws.String(cr.Spec.CIDRBlock),
	})

	rsp, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(rsp.Subnet.SubnetId))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	cr.UpdateExternalStatus(*rsp.Subnet)
	return managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{}}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.Subnet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DeleteSubnetRequest(&awsec2.DeleteSubnetInput{
		SubnetId: aws.String(meta.GetExternalName(cr)),
	})

	_, err := req.Send(ctx)
	return errors.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDelete)
}
