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

package securitygroup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/event"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	v1alpha3 "github.com/crossplaneio/stack-aws/apis/network/v1alpha3"
	"github.com/crossplaneio/stack-aws/pkg/clients/ec2"
	"github.com/crossplaneio/stack-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an SecurityGroup resource"
	errClient           = "cannot create a new SecurityGroupClient"
	errDescribe         = "failed to describe SecurityGroup with id: %v"
	errMultipleItems    = "retrieved multiple SecurityGroups for the given securityGroupId: %v"
	errCreate           = "failed to create the SecurityGroup resource"
	errAuthorizeIngress = "failed to authorize ingress rules"
	errAuthorizeEgress  = "failed to authorize egress rules"
	errDeleteNotPresent = "cannot delete the SecurityGroup, since the SecurityGroupID is not present"
	errDelete           = "failed to delete the SecurityGroup resource"
)

// SetupSecurityGroup adds a controller that reconciles SecurityGroups.
func SetupSecurityGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.SecurityGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.SecurityGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.SecurityGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewSecurityGroupClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (ec2.SecurityGroupClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
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

	return &external{c}, nil
}

type external struct {
	client ec2.SecurityGroupClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// To find out whether a SecurityGroup exist:
	// - the object's ExternalState should have securityGroupId populated
	// - a SecurityGroup with the given securityGroupId should exist
	if cr.Status.SecurityGroupID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	req := e.client.DescribeSecurityGroupsRequest(&awsec2.DescribeSecurityGroupsInput{
		GroupIds: []string{cr.Status.SecurityGroupID},
	})
	req.SetContext(ctx)

	response, err := req.Send()

	if ec2.IsSecurityGroupNotFoundErr(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrapf(err, errDescribe, cr.Status.SecurityGroupID)
	}

	// in a successful response, there should be one and only one object
	if len(response.SecurityGroups) != 1 {
		return managed.ExternalObservation{}, errors.Errorf(errMultipleItems, cr.Status.SecurityGroupID)
	}

	observed := response.SecurityGroups[0]

	// the fact that the security group is successfully fetched,
	// indicates that it's available
	cr.SetConditions(runtimev1alpha1.Available())

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	// Creating the SecurityGroup itself
	req := e.client.CreateSecurityGroupRequest(&awsec2.CreateSecurityGroupInput{
		GroupName:   aws.String(cr.Spec.GroupName),
		VpcId:       aws.String(cr.Spec.VPCID),
		Description: aws.String(cr.Spec.Description),
	})
	req.SetContext(ctx)

	result, err := req.Send()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	cr.UpdateExternalStatus(awsec2.SecurityGroup{GroupId: result.GroupId})

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	// Authorizing Ingress permissions for the SecurityGroup
	ingressPerms := v1alpha3.BuildEC2Permissions(cr.Spec.IngressPermissions)
	if len(ingressPerms) > 0 {
		air := e.client.AuthorizeSecurityGroupIngressRequest(&awsec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(cr.Status.SecurityGroupID),
			IpPermissions: ingressPerms,
		})
		air.SetContext(ctx)

		_, err = air.Send()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errAuthorizeIngress)
		}
	}

	// Authorizing Egress permissions for the SecurityGroup
	egressPerms := v1alpha3.BuildEC2Permissions(cr.Spec.EgressPermissions)
	if len(egressPerms) > 0 {
		aer := e.client.AuthorizeSecurityGroupEgressRequest(&awsec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(cr.Status.SecurityGroupID),
			IpPermissions: egressPerms,
		})
		aer.SetContext(ctx)

		_, err = aer.Send()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errAuthorizeEgress)
		}
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	if cr.Status.SecurityGroupID == "" {
		return errors.New(errDeleteNotPresent)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DeleteSecurityGroupRequest(&awsec2.DeleteSecurityGroupInput{
		GroupId: aws.String(cr.Status.SecurityGroupID),
	})

	req.SetContext(ctx)

	_, err := req.Send()

	if ec2.IsSecurityGroupNotFoundErr(err) {
		return nil
	}

	return errors.Wrap(err, errDelete)
}
