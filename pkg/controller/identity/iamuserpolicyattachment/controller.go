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

package iamuserpolicyattachment

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errNotUserInstance = "managed resource is not an User instance custom resource"

	errCreateUserClient  = "cannot create IAM User client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errUnexpectedObject = "The managed resource is not an UserPolicyAttachment resource"
	errGet              = "failed to get UserPolicyAttachments for user"
	errAttach           = "failed to attach the policy to user"
	errDetach           = "failed to detach the policy to user"

	errKubeUpdateFailed = "cannot late initialize UserPolicyAttachment"
)

// SetupIAMUserPolicyAttachment adds a controller that reconciles
// IAMUserPolicyAttachments.
func SetupIAMUserPolicyAttachment(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.UserPolicyAttachmentGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.IAMUserPolicyAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.UserPolicyAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewUserPolicyAttachmentClient}),
			managed.WithConnectionPublishers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (iam.UserPolicyAttachmentClient, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IAMUserPolicyAttachment)
	if !ok {
		return nil, errors.New(errNotUserInstance)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		userPolicyClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: userPolicyClient, kube: c.kube}, errors.Wrap(err, errCreateUserClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	userClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: userClient, kube: c.kube}, errors.Wrap(err, errCreateUserClient)
}

type external struct {
	client iam.UserPolicyAttachmentClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMUserPolicyAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.ListAttachedUserPoliciesRequest(&awsiam.ListAttachedUserPoliciesInput{
		UserName: cr.Spec.ForProvider.UserName,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	var attachedPolicyObject *awsiam.AttachedPolicy
	for i, policy := range observed.AttachedPolicies {
		if cr.Spec.ForProvider.PolicyARN == aws.StringValue(policy.PolicyArn) {
			attachedPolicyObject = &observed.AttachedPolicies[i]
			break
		}
	}

	if attachedPolicyObject == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	iam.LateInitializeUserPolicy(&cr.Spec.ForProvider, attachedPolicyObject)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = v1alpha1.IAMUserPolicyAttachmentObservation{
		AttachedPolicyARN: aws.StringValue(attachedPolicyObject.PolicyArn),
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMUserPolicyAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	_, err := e.client.AttachUserPolicyRequest(&awsiam.AttachUserPolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		UserName:  cr.Spec.ForProvider.UserName,
	}).Send(ctx)

	return managed.ExternalCreation{}, errors.Wrap(err, errAttach)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// PolicyARN is the only distinguishing field and on update to that, new policy is attached
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMUserPolicyAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DetachUserPolicyRequest(&awsiam.DetachUserPolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		UserName:  cr.Spec.ForProvider.UserName,
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDetach)
}
