/*
Copyright 2020 The Crossplane Authors.

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

package accesskey

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUnexpectedObject = "The managed resource is not an AccessKey resource"
	errList             = "failed to list AccessKeys"
	errCreate           = "failed to create the AccessKey resource"
	errDelete           = "failed to delete the AccessKey resource"
	errUpdate           = "failed to update the AccessKey resource"
)

// SetupAccessKey adds a controller that reconciles AccessKeys.
func SetupAccessKey(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.AccessKeyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.AccessKey{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.AccessKeyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewAccessClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.AccessClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.AccessClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.AccessKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	keys, err := e.client.ListAccessKeys(ctx, &awsiam.ListAccessKeysInput{UserName: aws.String(cr.Spec.ForProvider.Username)})
	if err != nil || len(keys.AccessKeyMetadata) == 0 {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errList)
	}
	found := false
	var accessKey awsiamtypes.AccessKeyMetadata
	for _, key := range keys.AccessKeyMetadata {
		if aws.ToString(key.AccessKeyId) == meta.GetExternalName(cr) {
			found = true
			accessKey = key
		}
	}
	if !found {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	switch accessKey.Status {
	case awsiamtypes.StatusTypeActive:
		cr.SetConditions(xpv1.Available())
	case awsiamtypes.StatusTypeInactive:
		cr.SetConditions(xpv1.Unavailable())
	}
	current := cr.Spec.ForProvider.Status
	cr.Spec.ForProvider.Status = awsclient.LateInitializeString(cr.Spec.ForProvider.Status, aws.String(string(accessKey.Status)))
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        string(accessKey.Status) == cr.Spec.ForProvider.Status,
		ResourceLateInitialized: current != cr.Spec.ForProvider.Status,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.AccessKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.CreateAccessKey(ctx, &awsiam.CreateAccessKeyInput{UserName: aws.String(cr.Spec.ForProvider.Username)})
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	var conn managed.ConnectionDetails
	if response != nil && response.AccessKey != nil {
		conn = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretUserKey:     []byte(aws.ToString(response.AccessKey.AccessKeyId)),
			xpv1.ResourceCredentialsSecretPasswordKey: []byte(aws.ToString(response.AccessKey.SecretAccessKey)),
		}
	}
	meta.SetExternalName(cr, aws.ToString(response.AccessKey.AccessKeyId))
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.AccessKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateAccessKey(ctx, &awsiam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(meta.GetExternalName(cr)),
		Status:      awsiamtypes.StatusType(cr.Spec.ForProvider.Status),
		UserName:    aws.String(cr.Spec.ForProvider.Username),
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.AccessKey)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteAccessKey(ctx, &awsiam.DeleteAccessKeyInput{
		UserName:    aws.String(cr.Spec.ForProvider.Username),
		AccessKeyId: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
