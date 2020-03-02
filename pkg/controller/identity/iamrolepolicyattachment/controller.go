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

package iamrolepolicyattachment

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMRolePolicyAttachment resource"
	errClient           = "cannot create a new RolePolicyAttachmentClient"
	errGet              = "failed to get IAMRolePolicyAttachments for role with name: %v"
	errAttach           = "failed to attach the policy with arn %v to role %v"
	errDetach           = "failed to detach the policy with arn %v to role %v"
)

// SetupIAMRolePolicyAttachment adds a controller that reconciles
// IAMRolePolicyAttachments.
func SetupIAMRolePolicyAttachment(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.IAMRolePolicyAttachmentGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.IAMRolePolicyAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.IAMRolePolicyAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: iam.NewRolePolicyAttachmentClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (iam.RolePolicyAttachmentClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}

	return &external{c}, nil
}

type external struct {
	client iam.RolePolicyAttachmentClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.ListAttachedRolePoliciesRequest(&awsiam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(cr.Spec.ForProvider.RoleName),
	})

	observed, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	var attachedPolicyObject *awsiam.AttachedPolicy
	for i := 0; i < len(observed.AttachedPolicies); i++ {
		if cr.Spec.ForProvider.PolicyARN == aws.StringValue(observed.AttachedPolicies[i].PolicyArn) {
			attachedPolicyObject = &(observed.AttachedPolicies[i])
			break
		}
	}

	if attachedPolicyObject == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	cr.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = iam.GenerateRolePolicyObservation(*attachedPolicyObject)

	return managed.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	req := e.client.AttachRolePolicyRequest(iam.GenerateAttachRolePolicyInput(&cr.Spec.ForProvider))
	req.SetContext(ctx)

	_, err := req.Send(ctx)

	return managed.ExternalCreation{}, errors.Wrapf(err, errAttach, cr.Spec.ForProvider.PolicyARN, cr.Spec.ForProvider.RoleName)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	// there is not a dedicated update method, so a basic update is implemented below
	// based on changes on PolicyArn:
	// if the previously attached policy is different than what is stated in the spec,
	// for update it needs to first attach the updated one, and then detach the previous one
	if cr.Status.AtProvider.AttachedPolicyARN == "" || cr.Spec.ForProvider.PolicyARN == cr.Status.AtProvider.AttachedPolicyARN {
		// update is only necessary if the PolicyArn in the Status is set and is different
		// from the one in Spec
		return managed.ExternalUpdate{}, nil
	}

	aReq := e.client.AttachRolePolicyRequest(iam.GenerateAttachRolePolicyInput(&cr.Spec.ForProvider))
	if _, err := aReq.Send(ctx); err != nil {
		return managed.ExternalUpdate{}, errors.Wrapf(err, errAttach, cr.Spec.ForProvider.PolicyARN, cr.Spec.ForProvider.RoleName)
	}

	dReq := e.client.DetachRolePolicyRequest(iam.GenerateUpdateRolePolicyInput(&cr.Status.AtProvider))

	if _, err := dReq.Send(ctx); err != nil {
		return managed.ExternalUpdate{}, errors.Wrapf(err, errDetach, cr.Status.AtProvider.AttachedPolicyARN, cr.Spec.ForProvider.RoleName)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DetachRolePolicyRequest(iam.GenerateDetachRolePolicyInput(&cr.Spec.ForProvider))
	req.SetContext(ctx)

	_, err := req.Send(ctx)

	if iam.IsErrorNotFound(err) {
		return nil
	}

	return errors.Wrapf(err, errDetach, cr.Spec.ForProvider.PolicyARN, cr.Spec.ForProvider.RoleName)
}
