/*
Copyright 2021 The Crossplane Authors.

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

package openidconnectprovider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	errUnexpectedObject = "managed resource is not an OpenIDConnectProvider resource"

	errList             = "cannot list OpenIDConnectProvider in AWS"
	errListTags         = "cannot list OpenIDConnectProvider tags in AWS"
	errGet              = "cannot get OpenIDConnectProvider in AWS"
	errCreate           = "cannot create OpenIDConnectProvider in AWS"
	errUpdateThumbprint = "cannot update OpenIDConnectProvider thumbprint list in AWS"
	errAddClientID      = "cannot add clientID to OpenIDConnectProvider in AWS"
	errRemoveClientID   = "cannot remove clientID to OpenIDConnectProvider in AWS"
	errDelete           = "failed to delete OpenIDConnectProvider"
	errSDK              = "empty OpenIDConnectProvider received from IAM API"
	errAddTags          = "cannot add tags to OpenIDConnectProvider in AWS"
	errRemoveTags       = "cannot remove tags to OpenIDConnectProvider in AWS"
	errKubeUpdateFailed = "cannot update OpenIDConnectProvider instance custom resource"
)

// SetupOpenIDConnectProvider adds a controller that reconciles OpenIDConnectProvider.
func SetupOpenIDConnectProvider(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.OpenIDConnectProviderGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewOpenIDConnectProviderClient}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithInitializers(),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.OpenIDConnectProviderGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.OpenIDConnectProvider{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.OpenIDConnectProviderClient
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{
		kube:   c.kube,
		client: c.newClientFn(*cfg),
	}, nil
}

type external struct {
	kube   client.Client
	client iam.OpenIDConnectProviderClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		arn, err := e.getOpenIDConnectProviderByTags(ctx, resource.GetExternalTags(mgd))
		if arn == nil || err != nil {
			return managed.ExternalObservation{}, resource.Ignore(iam.IsErrorNotFound, err)
		}

		meta.SetExternalName(cr, aws.ToString(arn))
		_ = e.kube.Update(ctx, cr)
	}

	observedProvider, err := e.client.GetOpenIDConnectProvider(ctx, &awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}
	if observedProvider == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	// No observation or late initialization is necessary for this resource
	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider = iam.GenerateOIDCProviderObservation(*observedProvider)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: iam.IsOIDCProviderUpToDate(cr.Spec.ForProvider, *observedProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	iamTags := make([]iamtypes.Tag, len(cr.Spec.ForProvider.Tags))
	for i := range cr.Spec.ForProvider.Tags {
		iamTags[i] = iamtypes.Tag{Key: aws.String(cr.Spec.ForProvider.Tags[i].Key), Value: aws.String(cr.Spec.ForProvider.Tags[i].Value)}
	}

	observed, err := e.client.CreateOpenIDConnectProvider(ctx, &awsiam.CreateOpenIDConnectProviderInput{
		ClientIDList:   cr.Spec.ForProvider.ClientIDList,
		ThumbprintList: cr.Spec.ForProvider.ThumbprintList,
		Url:            aws.String(cr.Spec.ForProvider.URL),
		Tags:           iamTags,
	})

	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(observed.OpenIDConnectProviderArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	// NOTE(cebernardi) gocyclo is disabled because the method needs to check the different components (Thumbprint,
	// ClientID and Tags and it's updating them with dedicated calls, hence increasing the complexity

	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	arn := aws.String(meta.GetExternalName(cr))
	observedProvider, err := e.client.GetOpenIDConnectProvider(ctx, &awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: arn,
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errGet)
	}
	if observedProvider == nil {
		return managed.ExternalUpdate{}, errors.New(errSDK)
	}

	if !cmp.Equal(cr.Spec.ForProvider.ThumbprintList, observedProvider.ThumbprintList, cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(x, y string) bool {
			return x < y
		})) {
		if _, err := e.client.UpdateOpenIDConnectProviderThumbprint(ctx, &awsiam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: arn,
			ThumbprintList:           cr.Spec.ForProvider.ThumbprintList,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdateThumbprint)
		}
	}

	addClientIDList, removeClientIDList := iam.SliceDifference(observedProvider.ClientIDList, cr.Spec.ForProvider.ClientIDList)
	for _, clientID := range addClientIDList {
		if _, err := e.client.AddClientIDToOpenIDConnectProvider(ctx, &awsiam.AddClientIDToOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
			ClientID:                 aws.String(clientID),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddClientID)
		}
	}

	for _, clientID := range removeClientIDList {
		if _, err := e.client.RemoveClientIDFromOpenIDConnectProvider(ctx, &awsiam.RemoveClientIDFromOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
			ClientID:                 aws.String(clientID),
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errRemoveClientID)
		}
	}

	addTags, removeTags, _ := iam.DiffIAMTagsWithUpdates(cr.Spec.ForProvider.Tags, observedProvider.Tags)

	if len(addTags) > 0 {
		if _, err := e.client.TagOpenIDConnectProvider(ctx, &awsiam.TagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
			Tags:                     addTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddTags)
		}
	}

	if len(removeTags) > 0 {
		if _, err := e.client.UntagOpenIDConnectProvider(ctx, &awsiam.UntagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
			TagKeys:                  removeTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errRemoveTags)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	_, err := e.client.DeleteOpenIDConnectProvider(ctx, &awsiam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}

func (e *external) getOpenIDConnectProviderByTags(ctx context.Context, tags map[string]string) (*string, error) {
	name, ok := tags[resource.ExternalResourceTagKeyName]
	if !ok {
		return nil, nil
	}

	oidcs, err := e.client.ListOpenIDConnectProviders(ctx, &awsiam.ListOpenIDConnectProvidersInput{})
	if err != nil || len(oidcs.OpenIDConnectProviderList) == 0 {
		return nil, errorutils.Wrap(err, errList)
	}

	for _, o := range oidcs.OpenIDConnectProviderList {
		tags, err := e.client.ListOpenIDConnectProviderTags(ctx, &awsiam.ListOpenIDConnectProviderTagsInput{
			OpenIDConnectProviderArn: o.Arn,
		})
		if err != nil {
			return nil, errorutils.Wrap(err, errListTags)
		}

		for _, t := range tags.Tags {
			if *t.Key == resource.ExternalResourceTagKeyName && *t.Value == name {
				return o.Arn, nil
			}
		}
	}
	return nil, nil
}
