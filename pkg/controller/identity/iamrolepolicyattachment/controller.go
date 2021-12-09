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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	iam "github.com/crossplane/provider-aws/pkg/clients/identity"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMRolePolicyAttachment resource"
	errGet              = "failed to get IAMRolePolicyAttachments for role with name"
	errAttach           = "failed to attach the policy to role"
	errDetach           = "failed to detach the policy to role"
)

// SetupIAMRolePolicyAttachment adds a controller that reconciles
// IAMRolePolicyAttachments.
func SetupIAMRolePolicyAttachment(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.IAMRolePolicyAttachmentGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1beta1.IAMRolePolicyAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.IAMRolePolicyAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewRolePolicyAttachmentClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.RolePolicyAttachmentClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
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
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.ListAttachedRolePolicies(ctx, &awsiam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(cr.Spec.ForProvider.RoleName),
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
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
	cr.Status.AtProvider.AttachedPolicyARN = awsclient.StringValue(attachedPolicyObject.PolicyArn)
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	_, err := e.client.AttachRolePolicy(ctx, &awsiam.AttachRolePolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		RoleName:  aws.String(cr.Spec.ForProvider.RoleName),
	})
	return managed.ExternalCreation{}, awsclient.Wrap(err, errAttach)
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// PolicyARN is the only distinguishing field and on update to that, new policy is attached
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.IAMRolePolicyAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	_, err := e.client.DetachRolePolicy(ctx, &awsiam.DetachRolePolicyInput{
		PolicyArn: aws.String(cr.Spec.ForProvider.PolicyARN),
		RoleName:  aws.String(cr.Spec.ForProvider.RoleName),
	})
	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDetach)
}
