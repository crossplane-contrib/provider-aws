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

package rolepolicyattachment

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
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
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an RolePolicyAttachment resource"
	errGet              = "failed to get RolePolicyAttachments for role with name"
	errAttach           = "failed to attach the policy to role"
	errDetach           = "failed to detach the policy to role"
)

// SetupRolePolicyAttachment adds a controller that reconciles
// RolePolicyAttachments.
func SetupRolePolicyAttachment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RolePolicyAttachmentGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewRolePolicyAttachmentClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RolePolicyAttachmentGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.RolePolicyAttachment{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.RolePolicyAttachmentClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.RolePolicyAttachmentClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.RolePolicyAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.ListAttachedRolePolicies(ctx, &awsiam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(cr.Spec.ForProvider.RoleName),
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	var attachedPolicyObject *awsiamtypes.AttachedPolicy
	for i, policy := range observed.AttachedPolicies {
		if cr.Spec.ForProvider.PolicyARN == aws.ToString(policy.PolicyArn) {
			attachedPolicyObject = &observed.AttachedPolicies[i]
			break
		}
	}
	if attachedPolicyObject == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider.AttachedPolicyARN = pointer.StringValue(attachedPolicyObject.PolicyArn)
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.RolePolicyAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	_, err := e.client.AttachRolePolicy(ctx, &awsiam.AttachRolePolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		RoleName:  aws.String(cr.Spec.ForProvider.RoleName),
	})
	return managed.ExternalCreation{}, errorutils.Wrap(err, errAttach)
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// PolicyARN is the only distinguishing field and on update to that, new policy is attached
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.RolePolicyAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	_, err := e.client.DetachRolePolicy(ctx, &awsiam.DetachRolePolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		RoleName:  aws.String(cr.Spec.ForProvider.RoleName),
	})
	return errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDetach)
}
