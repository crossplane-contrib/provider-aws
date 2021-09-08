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

package iamgroupusermembership

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an GroupUserMembership resource"

	errGet    = "failed to get groups for user"
	errAdd    = "failed to add the user to group"
	errRemove = "failed to remove the user to group"
)

// SetupIAMGroupUserMembership adds a controller that reconciles
// IAMGroupUserMemberships.
func SetupIAMGroupUserMembership(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.IAMGroupUserMembershipGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.IAMGroupUserMembership{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.IAMGroupUserMembershipGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewGroupUserMembershipClient}),
			managed.WithConnectionPublishers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.GroupUserMembershipClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.GroupUserMembershipClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMGroupUserMembership)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// We need to use external name if it exists in order to support resource import.
	nn := strings.Split(meta.GetExternalName(cr), "/")
	if len(nn) != 2 {
		return managed.ExternalObservation{}, errors.New("external name has to be in the following format <group-name>/<user-name>")
	}
	groupName, userName := nn[0], nn[1]

	observed, err := e.client.ListGroupsForUserRequest(&awsiam.ListGroupsForUserInput{
		UserName: &userName,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errGet)
	}

	var attachedGroupObject *awsiam.Group
	for i, group := range observed.Groups {
		if groupName == aws.StringValue(group.GroupName) {
			attachedGroupObject = &observed.Groups[i]
			break
		}
	}

	if attachedGroupObject == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider = v1alpha1.IAMGroupUserMembershipObservation{
		AttachedGroupARN: aws.StringValue(attachedGroupObject.Arn),
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMGroupUserMembership)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.AddUserToGroupRequest(&awsiam.AddUserToGroupInput{
		GroupName: &cr.Spec.ForProvider.GroupName,
		UserName:  &cr.Spec.ForProvider.UserName,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errAdd)
	}

	// This resource is interesting in that it's a binding without its own
	// external identity. We therefore derive an external name from the
	// names of the group and user that are bound.
	meta.SetExternalName(cr, cr.Spec.ForProvider.GroupName+"/"+cr.Spec.ForProvider.UserName)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// Updating any field will create a new Group-User Membership in AWS, which will be
	// irrelevant/out-of-sync to the original defined attachment.
	// It is encouraged to instead create a new IAMGroupUserMembership resource.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMGroupUserMembership)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.RemoveUserFromGroupRequest(&awsiam.RemoveUserFromGroupInput{
		GroupName: &cr.Spec.ForProvider.GroupName,
		UserName:  &cr.Spec.ForProvider.UserName,
	}).Send(ctx)

	return awsclient.Wrap(err, errRemove)
}
