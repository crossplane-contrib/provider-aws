/*
Copyright 2023 The Crossplane Authors.

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

package rolepolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
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
	errUnexpectedObject      = "The managed resource is not a RolePolicy resource"
	errGet                   = "failed to get RolePolicy for role with name"
	errPutRolePolicy         = "cannot put role policy"
	errInvalidPolicyDocument = "invalid policy document"
)

// SetupRolePolicy adds a controller that reconciles RolePolicy.
func SetupRolePolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RoleGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewRolePolicyClient, newSTSClientFn: iam.NewSTSClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RolePolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.RolePolicy{}).
		Complete(r)
}

type connector struct {
	kube           client.Client
	newClientFn    func(config aws.Config) iam.RolePolicyClient
	newSTSClientFn func(config aws.Config) iam.STSClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), sts: c.newSTSClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.RolePolicyClient
	sts    iam.STSClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.RolePolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	observed, err := e.client.GetRolePolicy(ctx, &awsiam.GetRolePolicyInput{
		PolicyName: aws.String(meta.GetExternalName(cr)),
		RoleName:   aws.String(cr.Spec.ForProvider.RoleName),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	cr.SetConditions(xpv1.Available())

	upToDate, diff, err := IsInlinePolicyUpToDate(string(cr.Spec.ForProvider.Document.Raw), observed.PolicyDocument)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
		Diff:             diff,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.RolePolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	return managed.ExternalCreation{}, e.putRolePolicy(ctx, cr)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.RolePolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	return managed.ExternalUpdate{}, e.putRolePolicy(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.RolePolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteRolePolicy(ctx, &awsiam.DeleteRolePolicyInput{
		PolicyName: aws.String(meta.GetExternalName(cr)),
		RoleName:   aws.String(cr.Spec.ForProvider.RoleName),
	})

	return err
}

func (e *external) putRolePolicy(ctx context.Context, cr *v1beta1.RolePolicy) error {
	if err := iam.ValidatePolicyObject(string(cr.Spec.ForProvider.Document.Raw)); err != nil {
		return errors.Wrap(err, errInvalidPolicyDocument)
	}
	_, err := e.client.PutRolePolicy(ctx, &awsiam.PutRolePolicyInput{
		PolicyName:     aws.String(meta.GetExternalName(cr)),
		RoleName:       aws.String(cr.Spec.ForProvider.RoleName),
		PolicyDocument: aws.String(string(cr.Spec.ForProvider.Document.Raw)),
	})
	return errors.Wrap(err, errPutRolePolicy)
}

// IsRolePolicyNotFoundErr returns true if the aws exception indicates the role policy was not found
func IsRolePolicyNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == iam.ErrRolePolicyNotFound
}

// IsInlinePolicyUpToDate checks whether there is a change in any of the modifiable fields in policy.
func IsInlinePolicyUpToDate(cr string, external *string) (bool, string, error) {
	// The AWS API returns Policy Document as an escaped string.
	// Due to differences in the methods to escape a string, the comparison result between
	// the spec.Document and policy.Document can sometimes be false negative (due to spaces, line feeds).
	// Escaping with a common method and then comparing is a safe way.

	// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_grammar.html
	return iam.IsPolicyDocumentUpToDate(cr, external)
}
