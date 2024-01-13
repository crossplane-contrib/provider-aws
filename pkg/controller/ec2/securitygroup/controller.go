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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an SecurityGroup resource"

	errDescribe         = "failed to describe SecurityGroup"
	errGetSecurityGroup = "failed to get SecurityGroup based on groupName"
	errMultipleItems    = "retrieved multiple SecurityGroups for the given securityGroupId"
	errCreate           = "failed to create the SecurityGroup resource"
	errAuthorizeIngress = "failed to authorize ingress rules"
	errAuthorizeEgress  = "failed to authorize egress rules"
	errDelete           = "failed to delete the SecurityGroup resource"
	errSpecUpdate       = "cannot update spec of the SecurityGroup custom resource"
	errRevokeEgress     = "failed to revoke egress rules"
	errRevokeIngress    = "failed to revoke ingress rules"
	errStatusUpdate     = "cannot update status of the SecurityGroup custom resource"
	errCreateTags       = "failed to create tags for the Security Group resource"
	errDeleteTags       = "failed to delete tags for the Security Group resource"
)

// SetupSecurityGroup adds a controller that reconciles SecurityGroups.
func SetupSecurityGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.SecurityGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewSecurityGroupClient}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(),
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
		resource.ManagedKind(v1beta1.SecurityGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.SecurityGroup{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.SecurityGroupClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.SecurityGroup)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{sg: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	sg   ec2.SecurityGroupClient
	kube client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		securityGroupArn, err := e.getSecurityGroupByName(ctx, cr.Spec.ForProvider.GroupName)
		if securityGroupArn == nil || err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errGetSecurityGroup)
		}

		meta.SetExternalName(cr, aws.ToString(securityGroupArn))
		_ = e.kube.Update(ctx, cr)
	}

	response, err := e.sg.DescribeSecurityGroups(ctx, &awsec2.DescribeSecurityGroupsInput{
		GroupIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.SecurityGroups) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.SecurityGroups[0]

	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeSG(&cr.Spec.ForProvider, &observed)

	cr.Status.AtProvider = ec2.GenerateSGObservation(observed)

	upToDate := ec2.IsSGUpToDate(cr.Spec.ForProvider, observed)
	// this is to make sure that the security group exists with the specified traffic rules.
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	// Creating the SecurityGroup itself
	result, err := e.sg.CreateSecurityGroup(ctx, &awsec2.CreateSecurityGroupInput{
		GroupName:   aws.String(cr.Spec.ForProvider.GroupName),
		VpcId:       cr.Spec.ForProvider.VPCID,
		Description: aws.String(cr.Spec.ForProvider.Description),
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	en := aws.ToString(result.GroupId)
	// NOTE(muvaf): We have this code block in managed reconciler but this resource
	// has an exception where it needs to make another API call right after the
	// Create and we cannot afford losing the identifier in case RevokeSecurityGroupEgressRequest
	// fails.
	err = retry.OnError(retry.DefaultRetry, resource.IsAPIError, func() error {
		nn := types.NamespacedName{Name: cr.GetName()}
		if err := e.kube.Get(ctx, nn, cr); err != nil {
			return err
		}
		meta.SetExternalName(cr, en)
		return e.kube.Update(ctx, cr)
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSpecUpdate)
	}
	// NOTE(muvaf): AWS creates an initial egress rule and there is no way to
	// disable it with the create call. So, we revoke it right after the creation.
	_, err = e.sg.RevokeSecurityGroupEgress(ctx, &awsec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(meta.GetExternalName(cr)),
		IpPermissions: []awsec2types.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				IpRanges:   []awsec2types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			},
		},
	})
	return managed.ExternalCreation{}, errorutils.Wrap(err, errRevokeEgress)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.sg.DescribeSecurityGroups(ctx, &awsec2.DescribeSecurityGroupsInput{
		GroupIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDescribe)
	}

	add, remove := ec2.DiffEC2Tags(ec2.GenerateEC2TagsV1Beta1(cr.Spec.ForProvider.Tags), response.SecurityGroups[0].Tags)
	if len(remove) > 0 {
		if _, err := e.sg.DeleteTags(ctx, &awsec2.DeleteTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      remove,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errDeleteTags)
		}
	}

	if len(add) > 0 {
		if _, err := e.sg.CreateTags(ctx, &awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      add,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreateTags)
		}
	}

	if !pointer.BoolValue(cr.Spec.ForProvider.IgnoreIngress) {
		add, remove := ec2.DiffPermissions(ec2.GenerateEC2Permissions(cr.Spec.ForProvider.Ingress), response.SecurityGroups[0].IpPermissions)
		if len(remove) > 0 {
			if _, err := e.sg.RevokeSecurityGroupIngress(ctx, &awsec2.RevokeSecurityGroupIngressInput{
				GroupId:       aws.String(meta.GetExternalName(cr)),
				IpPermissions: remove,
			}); err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errRevokeIngress)
			}
		}
		if len(add) > 0 {
			if _, err := e.sg.AuthorizeSecurityGroupIngress(ctx, &awsec2.AuthorizeSecurityGroupIngressInput{
				GroupId:       aws.String(meta.GetExternalName(cr)),
				IpPermissions: add,
			}); err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errAuthorizeIngress)
			}
		}
	}

	if !pointer.BoolValue(cr.Spec.ForProvider.IgnoreEgress) {
		add, remove := ec2.DiffPermissions(ec2.GenerateEC2Permissions(cr.Spec.ForProvider.Egress), response.SecurityGroups[0].IpPermissionsEgress)
		if len(remove) > 0 {
			if _, err = e.sg.RevokeSecurityGroupEgress(ctx, &awsec2.RevokeSecurityGroupEgressInput{
				GroupId:       aws.String(meta.GetExternalName(cr)),
				IpPermissions: remove,
			}); err != nil {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errRevokeEgress)
			}
		}

		if len(add) > 0 {
			if _, err = e.sg.AuthorizeSecurityGroupEgress(ctx, &awsec2.AuthorizeSecurityGroupEgressInput{
				GroupId:       aws.String(meta.GetExternalName(cr)),
				IpPermissions: add,
			}); err != nil && !ec2.IsRuleAlreadyExistsErr(err) {
				return managed.ExternalUpdate{}, errorutils.Wrap(err, errAuthorizeEgress)
			}
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.SecurityGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.sg.DeleteSecurityGroup(ctx, &awsec2.DeleteSecurityGroupInput{
		GroupId: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(ec2.IsSecurityGroupNotFoundErr, err), errDelete)
}

func (e *external) getSecurityGroupByName(ctx context.Context, groupName string) (*string, error) {
	groups, err := e.sg.DescribeSecurityGroups(ctx, &awsec2.DescribeSecurityGroupsInput{
		Filters: []awsec2types.Filter{
			{Name: aws.String("group-name"), Values: []string{groupName}},
		},
	})

	if err != nil || len(groups.SecurityGroups) == 0 {
		return nil, err
	}

	return groups.SecurityGroups[0].GroupId, nil
}
