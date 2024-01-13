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

package identityproviderconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	tagutils "github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errNotEKSIdentityProviderConfig = "managed resource is not an EKS Identity Provider Config custom resource"
	errKubeUpdateFailed             = "cannot update EKS identity provider config custom resource"

	errCreateFailed   = "cannot associate EKS identity provider config"
	errDeleteFailed   = "cannot disassociate EKS identity provider config"
	errDescribeFailed = "cannot describe EKS identity provider config"
	errAddTagsFailed  = "cannot add tags to EKS identity provider config"
)

// SetupIdentityProviderConfig adds a controller that reconciles IdentityProviderConfigs.
func SetupIdentityProviderConfig(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.IdentityProviderConfigKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newEKSClientFn: eks.NewEKSClient}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(manualv1alpha1.IdentityProviderConfigGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&manualv1alpha1.IdentityProviderConfig{}).
		Complete(r)
}

type connector struct {
	kube           client.Client
	newEKSClientFn func(config aws.Config) eks.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.IdentityProviderConfig)
	if !ok {
		return nil, errors.New(errNotEKSIdentityProviderConfig)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newEKSClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client eks.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*manualv1alpha1.IdentityProviderConfig)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSIdentityProviderConfig)
	}

	rsp, err := e.client.DescribeIdentityProviderConfig(ctx, &awseks.DescribeIdentityProviderConfigInput{
		IdentityProviderConfig: &types.IdentityProviderConfig{
			Name: aws.String(meta.GetExternalName(cr)),
			Type: aws.String(string(manualv1alpha1.OidcIdentityProviderConfigType)),
		},
		ClusterName: &cr.Spec.ForProvider.ClusterName})
	if err != nil {
		// Failed IdentityProviderConfigs will be garbage collected by AWS after
		// some time.
		// Since we are using cr.Status.AtProvider.Status in Create() to
		// determine if we need to associate this fields needs to be reset if
		// the config does not exist.
		// Otherwise the controller will never retry associating again.
		if eks.IsErrorNotFound(err) {
			cr.Status.AtProvider.Status = ""
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errorutils.Wrap(err, errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitializeIdentityProviderConfig(&cr.Spec.ForProvider, rsp.IdentityProviderConfig)

	// Generate observation
	cr.Status.AtProvider = eks.GenerateIdentityProviderConfigObservation(rsp.IdentityProviderConfig)

	// Any of the statuses we don't explicitly address should be considered as
	// the identity provider config being unavailable.
	switch cr.Status.AtProvider.Status { //nolint:exhaustive
	case manualv1alpha1.IdentityProviderConfigStatusActive:
		cr.Status.SetConditions(xpv1.Available())
	case manualv1alpha1.IdentityProviderConfigStatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case manualv1alpha1.IdentityProviderConfigStatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        eks.IsIdentityProviderConfigUpToDate(&cr.Spec.ForProvider, rsp.IdentityProviderConfig),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*manualv1alpha1.IdentityProviderConfig)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSIdentityProviderConfig)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.Status == manualv1alpha1.IdentityProviderConfigStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.AssociateIdentityProviderConfig(ctx, eks.GenerateAssociateIdentityProviderConfigInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*manualv1alpha1.IdentityProviderConfig)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSIdentityProviderConfig)
	}

	if cr.Status.AtProvider.Status == manualv1alpha1.IdentityProviderConfigStatusCreating {
		return managed.ExternalUpdate{}, nil
	}
	rsp, err := e.client.DescribeIdentityProviderConfig(ctx, &awseks.DescribeIdentityProviderConfigInput{
		IdentityProviderConfig: &types.IdentityProviderConfig{
			Name: aws.String(meta.GetExternalName(cr)),
			Type: aws.String(string(manualv1alpha1.OidcIdentityProviderConfigType)),
		},
		ClusterName: &cr.Spec.ForProvider.ClusterName})

	if err != nil || rsp.IdentityProviderConfig == nil || rsp.IdentityProviderConfig.Oidc == nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errDescribeFailed)
	}
	add, remove := tagutils.DiffTags(cr.Spec.ForProvider.Tags, rsp.IdentityProviderConfig.Oidc.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResource(ctx, &awseks.UntagResourceInput{ResourceArn: rsp.IdentityProviderConfig.Oidc.IdentityProviderConfigArn, TagKeys: remove}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResource(ctx, &awseks.TagResourceInput{ResourceArn: rsp.IdentityProviderConfig.Oidc.IdentityProviderConfigArn, Tags: add}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*manualv1alpha1.IdentityProviderConfig)
	if !ok {
		return errors.New(errNotEKSIdentityProviderConfig)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == manualv1alpha1.IdentityProviderConfigStatusDeleting {
		return nil
	}
	_, err := e.client.DisassociateIdentityProviderConfig(ctx, eks.GenerateDisassociateIdentityProviderConfigInput(meta.GetExternalName(cr), cr.Spec.ForProvider.ClusterName))
	return errorutils.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}
