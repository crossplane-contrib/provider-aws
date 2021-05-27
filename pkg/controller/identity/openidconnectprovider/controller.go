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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "managed resource is not an OpenIDConnectProvider resource"

	errGet              = "cannot get OpenIDConnectProvider in AWS"
	errCreate           = "cannot create OpenIDConnectProvider in AWS"
	errUpdateThumbprint = "cannot update OpenIDConnectProvider thumbprint list in AWS"
	errAddClientID      = "cannot add clientID to OpenIDConnectProvider in AWS"
	errRemoveClientID   = "cannot remove clientID to OpenIDConnectProvider in AWS"
	errDelete           = "failed to delete OpenIDConnectProvider"
	errSDK              = "empty OpenIDConnectProvider received from IAM API"
)

// SetupOpenIDConnectProvider adds a controller that reconciles OpenIDConnectProvider.
func SetupOpenIDConnectProvider(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.OpenIDConnectProviderGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.OpenIDConnectProvider{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.OpenIDConnectProviderGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewOpenIDConnectProviderClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.OpenIDConnectProviderClient
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
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
	cr, ok := mgd.(*svcapitypes.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observed, err := e.client.GetOpenIDConnectProviderRequest(&awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}
	if observed.GetOpenIDConnectProviderOutput == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	observedProvider := observed.GetOpenIDConnectProviderOutput

	// No observation or late initialization is necessary for this resource
	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider = iam.GenerateOIDCProviderObservation(*observedProvider)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: iam.IsOIDCProviderUpToDate(cr.Spec.ForProvider, *observedProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*svcapitypes.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	observed, err := e.client.CreateOpenIDConnectProviderRequest(&awsiam.CreateOpenIDConnectProviderInput{
		ClientIDList:   cr.Spec.ForProvider.ClientIDList,
		ThumbprintList: cr.Spec.ForProvider.ThumbprintList,
		Url:            aws.String(cr.Spec.ForProvider.URL),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(observed.OpenIDConnectProviderArn))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*svcapitypes.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	observed, err := e.client.GetOpenIDConnectProviderRequest(&awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}
	if observed.GetOpenIDConnectProviderOutput == nil {
		return managed.ExternalUpdate{}, errors.New(errSDK)
	}
	observedProvider := observed.GetOpenIDConnectProviderOutput

	if !cmp.Equal(cr.Spec.ForProvider.ThumbprintList, observedProvider.ThumbprintList, cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(x, y string) bool {
			return x < y
		})) {
		if _, err := e.client.UpdateOpenIDConnectProviderThumbprintRequest(&awsiam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ThumbprintList:           cr.Spec.ForProvider.ThumbprintList,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateThumbprint)
		}
	}

	addClientIDList, removeClientIDList := iam.SliceDifference(observedProvider.ClientIDList, cr.Spec.ForProvider.ClientIDList)
	for _, clientID := range addClientIDList {
		if _, err := e.client.AddClientIDToOpenIDConnectProviderRequest(&awsiam.AddClientIDToOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ClientID:                 aws.String(clientID),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errAddClientID)
		}
	}

	for _, clientID := range removeClientIDList {
		if _, err := e.client.RemoveClientIDFromOpenIDConnectProviderRequest(&awsiam.RemoveClientIDFromOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ClientID:                 aws.String(clientID),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errRemoveClientID)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.OpenIDConnectProvider)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	_, err := e.client.DeleteOpenIDConnectProviderRequest(&awsiam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
