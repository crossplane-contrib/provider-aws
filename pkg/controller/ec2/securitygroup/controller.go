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

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an SecurityGroup resource"
	errKubeUpdateFailed = "cannot update Security Group instance custom resource"

	errDescribe         = "failed to describe SecurityGroup"
	errMultipleItems    = "retrieved multiple SecurityGroups for the given securityGroupId"
	errCreate           = "failed to create the SecurityGroup resource"
	errAuthorizeIngress = "failed to authorize ingress rules"
	errAuthorizeEgress  = "failed to authorize egress rules"
	errDelete           = "failed to delete the SecurityGroup resource"
	errSpecUpdate       = "cannot update spec of the SecurityGroup custom resource"
	errRevokeEgress     = "cannot remove the default egress rule"
	errStatusUpdate     = "cannot update status of the SecurityGroup custom resource"
	errUpdate           = "failed to update the SecurityGroup resource"
	errCreateTags       = "failed to create tags for the Security Group resource"
)

// SetupSecurityGroup adds a controller that reconciles SecurityGroups.
func SetupSecurityGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.SecurityGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.SecurityGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.SecurityGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewClient, awsConfigFn: awscommon.GetConfig}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.SecurityGroupClient
	awsConfigFn func(client.Client, context.Context, resource.Managed, string) (*aws.Config, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := c.awsConfigFn(c.kube, ctx, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{sg: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	sg   ec2.SecurityGroupClient
	kube client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.SecurityGroup)
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

	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeSG(&cr.Spec.ForProvider, &observed)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = ec2.GenerateSGObservation(observed)

	upToDate, err := ec2.IsSGUpToDate(cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	// this is to make sure that the security group exists with the specified traffic rules.
	if upToDate {
		cr.SetConditions(runtimev1alpha1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	// Creating the SecurityGroup itself
	result, err := e.sg.CreateSecurityGroupRequest(&awsec2.CreateSecurityGroupInput{
		GroupName:   aws.String(cr.Spec.ForProvider.GroupName),
		VpcId:       cr.Spec.ForProvider.VPCID,
		Description: aws.String(cr.Spec.ForProvider.Description),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}
	if result.CreateSecurityGroupOutput.GroupId == nil {
		return managed.ExternalCreation{}, errors.New(errCreate)
	}
	meta.SetExternalName(cr, aws.StringValue(result.GroupId))

	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSpecUpdate)
	}
	// NOTE(muvaf): AWS creates an initial egress rule and there is no way to
	// disable it with the create call. So, we revoke it right after the creation.
	_, err = e.sg.RevokeSecurityGroupEgressRequest(&awsec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(meta.GetExternalName(cr)),
		IpPermissions: []awsec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				IpRanges:   []awsec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			},
		},
	}).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errRevokeEgress)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.sg.DescribeSecurityGroupsRequest(&awsec2.DescribeSecurityGroupsInput{
		GroupIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDescribe)
	}

	patch, err := ec2.CreateSGPatch(response.SecurityGroups[0], cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errUpdate)
	}

	if len(patch.Tags) != 0 {
		if _, err := e.sg.CreateTagsRequest(&awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
		}
	}

	if patch.Ingress != nil {
		if _, err := e.sg.AuthorizeSecurityGroupIngressRequest(&awsec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(meta.GetExternalName(cr)),
			IpPermissions: ec2.GenerateEC2Permissions(cr.Spec.ForProvider.Ingress),
		}).Send(ctx); err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAuthorizeIngress)
		}
	}

	if patch.Egress != nil {
		if _, err = e.sg.AuthorizeSecurityGroupEgressRequest(&awsec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(meta.GetExternalName(cr)),
			IpPermissions: ec2.GenerateEC2Permissions(cr.Spec.ForProvider.Egress),
		}).Send(ctx); err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
			return managed.ExternalUpdate{}, errors.Wrap(err, errAuthorizeEgress)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.sg.DeleteSecurityGroupRequest(&awsec2.DeleteSecurityGroupInput{
		GroupId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDelete)
}
