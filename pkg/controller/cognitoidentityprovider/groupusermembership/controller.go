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
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscognitoidp "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awscognitoidptypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
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

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/manualv1alpha1"
	awsv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awscognitoidpclient "github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an GroupUserMembership resource"

	errGet       = "failed to get groups for user"
	errAdd       = "failed to add the user to group"
	errMarshal   = "failed to marshal external name of GroupUserMembership"
	errUnmarshal = "external name could not be unmarshalled"
	errRemove    = "failed to remove the user to group"
)

// SetupGroupUserMembership adds a controller that reconciles
// GroupUserMemberships.
func SetupGroupUserMembership(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.GroupUserMembershipGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), awsv1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: awscognitoidpclient.NewGroupUserMembershipClient}),
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
		resource.ManagedKind(svcapitypes.GroupUserMembershipGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.GroupUserMembership{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) awscognitoidpclient.GroupUserMembershipClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.GroupUserMembership)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client awscognitoidpclient.GroupUserMembershipClient
	kube   client.Client
}

// FindAttachedGroupObject uses an paginator in order to iterate over
// all groups of the user and returns a non-nil *awscognitoidptypes.GroupType
// in case a matching group was found. In case no matching group was found nil is returned.
// It returns an error, in case there is a AWS API error.
func FindAttachedGroupObject(ctx context.Context, e *external, userPoolID *string, username *string, groupname *string) (*awscognitoidptypes.GroupType, error) {
	req := &awscognitoidp.AdminListGroupsForUserInput{
		UserPoolId: userPoolID,
		Username:   username,
		Limit:      aws.Int32(60),
	}

	var observedGroups []awscognitoidptypes.GroupType
	paginator := awscognitoidp.NewAdminListGroupsForUserPaginator(e.client, req)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if resource.Ignore(awscognitoidpclient.IsErrorNotFound, err) != nil {
			// We will fail in case a generic API call error occurs
			return nil, err
		}
		observedGroups = append(observedGroups, page.Groups...)
	}

	var attachedGroupObject *awscognitoidptypes.GroupType
	for i, group := range observedGroups {
		if *groupname == aws.ToString(group.GroupName) {
			attachedGroupObject = &observedGroups[i]
			break
		}
	}
	return attachedGroupObject, nil
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*svcapitypes.GroupUserMembership)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// We need to use special external name annotation in order to to support existing resource resource imports.
	externalAnnotation := &svcapitypes.ExternalAnnotation{}
	err := json.Unmarshal([]byte(meta.GetExternalName(cr)), externalAnnotation)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUnmarshal)
	}

	attachedGroupObject, err := FindAttachedGroupObject(ctx, e, externalAnnotation.UserPoolID, externalAnnotation.Username, externalAnnotation.Groupname)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}

	if attachedGroupObject == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.Status.AtProvider = svcapitypes.GroupUserMembershipObservation{
		Groupname: attachedGroupObject.GroupName,
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*svcapitypes.GroupUserMembership)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.AdminAddUserToGroup(ctx, &awscognitoidp.AdminAddUserToGroupInput{
		GroupName:  &cr.Spec.ForProvider.Groupname,
		Username:   &cr.Spec.ForProvider.Username,
		UserPoolId: &cr.Spec.ForProvider.UserPoolID,
	})

	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errAdd)
	}

	// This resource is interesting in that it's a binding without its own
	// external identity. We therefore derive an external name from the
	// names of the groupname, username and userPoolId that are bound.
	externalAnnotation := &svcapitypes.ExternalAnnotation{
		UserPoolID: &cr.Spec.ForProvider.UserPoolID,
		Groupname:  &cr.Spec.ForProvider.Groupname,
		Username:   &cr.Spec.ForProvider.Username,
	}

	payload, err := json.Marshal(externalAnnotation)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errMarshal)
	}

	meta.SetExternalName(cr, string(payload))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// Updating any field will create a new Group-User Membership in AWS, which will be
	// irrelevant/out-of-sync to the original defined attachment.
	// It is encouraged to instead create a new GroupUserMembership resource.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.GroupUserMembership)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.AdminRemoveUserFromGroup(ctx, &awscognitoidp.AdminRemoveUserFromGroupInput{
		GroupName:  &cr.Spec.ForProvider.Groupname,
		Username:   &cr.Spec.ForProvider.Username,
		UserPoolId: &cr.Spec.ForProvider.UserPoolID,
	})

	return errorutils.Wrap(err, errRemove)
}
