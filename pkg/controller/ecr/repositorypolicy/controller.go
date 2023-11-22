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

package repositorypolicy

import (
	"context"

	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	ecr "github.com/crossplane-contrib/provider-aws/pkg/clients/ecr"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	legacypolicy "github.com/crossplane-contrib/provider-aws/pkg/utils/policy/old"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "managed resource is not an repository resource"

	errCreate = "failed to create repository policy"
	errGet    = "failed to get repository policy"
	errUpdate = "failed to update repository policy"
	errDelete = "failed to delete the repository resource"
)

// SetupRepositoryPolicy adds a controller that reconciles ECR.
func SetupRepositoryPolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RepositoryPolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
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
		resource.ManagedKind(v1beta1.RepositoryPolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.RepositoryPolicy{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RepositoryPolicy)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: awsecr.NewFromConfig(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ecr.RepositoryPolicyClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.RepositoryPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.GetRepositoryPolicy(ctx, &awsecr.GetRepositoryPolicyInput{
		RegistryId:     cr.Spec.ForProvider.RegistryID,
		RepositoryName: cr.Spec.ForProvider.RepositoryName,
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.IgnoreAny(err, ecr.IsRepoNotFoundErr, ecr.IsPolicyNotFoundErr), errGet)
	}

	policyData, err := ecr.RawPolicyData(cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGet)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	ecr.LateInitializeRepositoryPolicy(&cr.Spec.ForProvider, response)

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        legacypolicy.IsPolicyUpToDate(&policyData, response.PolicyText),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.RepositoryPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	policyData, err := ecr.RawPolicyData(cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}
	_, err = e.client.SetRepositoryPolicy(ctx, ecr.GenerateSetRepositoryPolicyInput(&cr.Spec.ForProvider, &policyData))

	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.RepositoryPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	policyData, err := ecr.RawPolicyData(cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}
	_, err = e.client.SetRepositoryPolicy(ctx, ecr.GenerateSetRepositoryPolicyInput(&cr.Spec.ForProvider, &policyData))
	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.RepositoryPolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	_, err := e.client.DeleteRepositoryPolicy(ctx, &awsecr.DeleteRepositoryPolicyInput{
		RepositoryName: cr.Spec.ForProvider.RepositoryName,
		RegistryId:     cr.Spec.ForProvider.RegistryID,
	})

	return errorutils.Wrap(resource.Ignore(ecr.IsPolicyNotFoundErr, err), errDelete)
}
