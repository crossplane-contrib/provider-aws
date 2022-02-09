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
	"time"

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

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"
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
func SetupOpenIDConnectProvider(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.OpenIDConnectProviderGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1beta1.OpenIDConnectProvider{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.OpenIDConnectProviderGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewOpenIDConnectProviderClient}),
			managed.WithInitializers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
      managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
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
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observedProvider, err := e.client.GetOpenIDConnectProvider(ctx, &awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
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
	observed, err := e.client.CreateOpenIDConnectProvider(ctx, &awsiam.CreateOpenIDConnectProviderInput{
		ClientIDList:   cr.Spec.ForProvider.ClientIDList,
		ThumbprintList: cr.Spec.ForProvider.ThumbprintList,
		Url:            aws.String(cr.Spec.ForProvider.URL),
    Tags:           iam.BuildIAMTags(cr.Spec.ForProvider.Tags)
	})

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(observed.OpenIDConnectProviderArn))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	observedProvider, err := e.client.GetOpenIDConnectProvider(ctx, &awsiam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errGet)
	}
	if observedProvider == nil {
		return managed.ExternalUpdate{}, errors.New(errSDK)
	}

	if !cmp.Equal(cr.Spec.ForProvider.ThumbprintList, observedProvider.ThumbprintList, cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(x, y string) bool {
			return x < y
		})) {
		if _, err := e.client.UpdateOpenIDConnectProviderThumbprint(ctx, &awsiam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ThumbprintList:           cr.Spec.ForProvider.ThumbprintList,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateThumbprint)
		}
	}

	addClientIDList, removeClientIDList := iam.SliceDifference(observedProvider.ClientIDList, cr.Spec.ForProvider.ClientIDList)
	for _, clientID := range addClientIDList {
		if _, err := e.client.AddClientIDToOpenIDConnectProvider(ctx, &awsiam.AddClientIDToOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ClientID:                 aws.String(clientID),
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errAddClientID)
		}
	}

	for _, clientID := range removeClientIDList {
		if _, err := e.client.RemoveClientIDFromOpenIDConnectProvider(ctx, &awsiam.RemoveClientIDFromOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(meta.GetExternalName(cr)),
			ClientID:                 aws.String(clientID),
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errRemoveClientID)
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

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.OpenIDConnectProvider)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	added := false
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		if p, ok := tagMap[k]; !ok || v != p {
			cr.Spec.ForProvider.Tags = append(cr.Spec.ForProvider.Tags, v1beta1.Tag{Key: k, Value: v})
			added = true
		}
	}
	if !added {
		return nil
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
