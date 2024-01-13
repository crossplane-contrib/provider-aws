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

package groupusermembership

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an GroupUserMembership resource"

	errGet    = "failed to get groups for user"
	errAdd    = "failed to add the user to group"
	errRemove = "failed to remove the user to group"
)

// SetupGroupUserMembership adds a controller that reconciles
// GroupUserMemberships.
func SetupGroupUserMembership(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.GroupUserMembershipGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewGroupUserMembershipClient}),
		managed.WithConnectionPublishers(),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
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
		resource.ManagedKind(v1beta1.GroupUserMembershipGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.GroupUserMembership{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.GroupUserMembershipClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
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
	cr, ok := mgd.(*v1beta1.GroupUserMembership)
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

	observed, err := e.client.ListGroupsForUser(ctx, &awsiam.ListGroupsForUserInput{
		UserName: &userName,
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	var attachedGroupObject *awsiamtypes.Group
	for i, group := range observed.Groups {
		if groupName == aws.ToString(group.GroupName) {
			attachedGroupObject = &observed.Groups[i]
			break
		}
	}

	if attachedGroupObject == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider = v1beta1.GroupUserMembershipObservation{
		AttachedGroupARN: aws.ToString(attachedGroupObject.Arn),
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.GroupUserMembership)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.AddUserToGroup(ctx, &awsiam.AddUserToGroupInput{
		GroupName: &cr.Spec.ForProvider.GroupName,
		UserName:  &cr.Spec.ForProvider.UserName,
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errAdd)
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
	// It is encouraged to instead create a new GroupUserMembership resource.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.GroupUserMembership)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.RemoveUserFromGroup(ctx, &awsiam.RemoveUserFromGroupInput{
		GroupName: &cr.Spec.ForProvider.GroupName,
		UserName:  &cr.Spec.ForProvider.UserName,
	})

	return errorutils.Wrap(err, errRemove)
}
