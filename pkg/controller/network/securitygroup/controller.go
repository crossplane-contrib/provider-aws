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

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject    = "The managed resource is not an SecurityGroup resource"
	errClient              = "cannot create a new SecurityGroupClient"
	errDescribe            = "failed to describe SecurityGroup"
	errMultipleItems       = "retrieved multiple SecurityGroups for the given securityGroupId"
	errCreate              = "failed to create the SecurityGroup resource"
	errPersistExternalName = "failed to persist InternetGateway ID"
	errAuthorizeIngress    = "failed to authorize ingress rules"
	errAuthorizeEgress     = "failed to authorize egress rules"
	errDelete              = "failed to delete the SecurityGroup resource"
)

// SetupSecurityGroup adds a controller that reconciles SecurityGroups.
func SetupSecurityGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.SecurityGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.SecurityGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.SecurityGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewSecurityGroupClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(*aws.Config) (ec2.SecurityGroupClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.kube, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}

	return &external{sg: c, kube: conn.kube}, nil
}

type external struct {
	sg   ec2.SecurityGroupClient
	kube client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	response, err := e.sg.DescribeSecurityGroupsRequest(&awsec2.DescribeSecurityGroupsInput{
		GroupIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.SecurityGroups) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.SecurityGroups[0]

	// the fact that the security group is successfully fetched,
	// indicates that it's available
	cr.SetConditions(runtimev1alpha1.Available())

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	// Creating the SecurityGroup itself
	req := e.sg.CreateSecurityGroupRequest(&awsec2.CreateSecurityGroupInput{
		GroupName:   aws.String(cr.Spec.GroupName),
		VpcId:       cr.Spec.VPCID,
		Description: aws.String(cr.Spec.Description),
	})

	rsp, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	// NOTE(muvaf): GroupID is used as external name instead of GroupName because
	// there are cases where only GroupID is accepted as identifier.
	meta.SetExternalName(cr, aws.StringValue(rsp.GroupId))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	cr.UpdateExternalStatus(awsec2.SecurityGroup{GroupId: rsp.GroupId})

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	if len(cr.Spec.Ingress) > 0 {
		_, err := e.sg.AuthorizeSecurityGroupIngressRequest(&awsec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(meta.GetExternalName(cr)),
			IpPermissions: ec2.GenerateEC2Permissions(cr.Spec.Ingress),
		}).Send(ctx)
		if err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAuthorizeIngress)
		}
	}
	if len(cr.Spec.Egress) > 0 {
		_, err := e.sg.AuthorizeSecurityGroupEgressRequest(&awsec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(meta.GetExternalName(cr)),
			IpPermissions: ec2.GenerateEC2Permissions(cr.Spec.Egress),
		}).Send(ctx)
		if err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAuthorizeEgress)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.SecurityGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.sg.DeleteSecurityGroupRequest(&awsec2.DeleteSecurityGroupInput{
		GroupId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDelete)
}
