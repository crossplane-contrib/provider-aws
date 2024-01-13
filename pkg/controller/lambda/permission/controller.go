/*
Copyright 2022 The Crossplane Authors.

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

package permission

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/lambda/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errSetExternalNameFailed = "cannot set external name of lambda permission"
	errNotLambdaPermission   = "managed resource is not an Lambda permission custom resource"
	errAddPermission         = "cannot add Lambda permission"
	errRemovePermission      = "cannot remove Lambda permission"
	errGetPolicyFailed       = "cannot get Lambda policy"
	errParsePolicy           = "cannot parse policy"
)

// SetupPermission adds a controller that reconciles Permissions.
func SetupPermission(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.PermissionKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newLambdaClientFn: svcsdk.NewFromConfig}),
		managed.WithInitializers(&externalNameGenerator{kube: mgr.GetClient()}),
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
		resource.ManagedKind(svcapitypes.PermissionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Permission{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newLambdaClientFn func(config aws.Config, optFns ...func(*svcsdk.Options)) *svcsdk.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return nil, errors.New(errNotLambdaPermission)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: *c.newLambdaClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client svcsdk.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLambdaPermission)
	}

	resp, err := e.client.GetPolicy(ctx, &svcsdk.GetPolicyInput{
		FunctionName: cr.Spec.ForProvider.FunctionName,
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errGetPolicyFailed)
	}

	policyDocument, err := parsePolicy(pointer.StringValue(resp.Policy))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errParsePolicy)
	}

	// Check if the policy contains the permission
	if policyDocument.StatementByID(meta.GetExternalName(cr)) == nil {
		return managed.ExternalObservation{}, nil
	}

	currentPermission := generatePermission(policyDocument, meta.GetExternalName(cr))

	current := cr.Spec.ForProvider.DeepCopy()
	e.lateInitialize(&cr.Spec.ForProvider, &currentPermission.Spec.ForProvider)

	cr.Status.AtProvider = currentPermission.Status.AtProvider
	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isUpToDate(cr, currentPermission),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) isUpToDate(spec, current *svcapitypes.Permission) bool {
	diff := cmp.Diff(
		spec.Spec.ForProvider,
		current.Spec.ForProvider,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.PermissionParameters{}, "Region", "FunctionName"))
	if diff != "" {
		fmt.Println(diff)
		return false
	}
	return true
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotLambdaPermission)
	}

	_, err := e.client.AddPermission(ctx, generateAddPermissionInput(cr))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errAddPermission)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotLambdaPermission)
	}

	if _, err := e.client.RemovePermission(ctx, generateRemovePermissionInput(cr)); resource.Ignore(isErrorNotFound, err) != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errRemovePermission)
	}

	_, err := e.client.AddPermission(ctx, generateAddPermissionInput(cr))
	return managed.ExternalUpdate{}, errors.Wrap(err, errAddPermission)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return errors.New(errNotLambdaPermission)
	}
	_, err := e.client.RemovePermission(ctx, generateRemovePermissionInput(cr))
	return errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errRemovePermission)
}

func (e *external) lateInitialize(spec, current *svcapitypes.PermissionParameters) {
	spec.EventSourceToken = pointer.LateInitialize(spec.EventSourceToken, current.EventSourceToken)
	spec.PrincipalOrgID = pointer.LateInitialize(spec.PrincipalOrgID, current.PrincipalOrgID)
	spec.SourceAccount = pointer.LateInitialize(spec.SourceAccount, current.SourceAccount)
	spec.SourceArn = pointer.LateInitialize(spec.SourceArn, current.SourceArn)
}

// IsErrorNotFound helper function to test for ResourceNotFoundException error.
func isErrorNotFound(err error) bool {
	var nfe *svcsdktypes.ResourceNotFoundException
	return errors.As(err, &nfe)
}

type externalNameGenerator struct {
	kube client.Client
}

// Initialize the given managed resource.
func (e *externalNameGenerator) Initialize(ctx context.Context, mg resource.Managed) error {
	if meta.GetExternalName(mg) != "" {
		return nil
	}
	cr, ok := mg.(*svcapitypes.Permission)
	if !ok {
		return errors.New(errNotLambdaPermission)
	}

	externalName := fmt.Sprintf("%s-%s", mg.GetName(), cr.Spec.Hash())
	meta.SetExternalName(mg, externalName)
	return errors.Wrap(e.kube.Update(ctx, mg), errSetExternalNameFailed)
}
