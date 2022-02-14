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

package userpolicyattachment

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an UserPolicyAttachment resource"

	errGet    = "failed to get UserPolicyAttachments for user"
	errAttach = "failed to attach the policy to user"
	errDetach = "failed to detach the policy to user"
)

// SetupUserPolicyAttachment adds a controller that reconciles
// UserPolicyAttachments.
func SetupUserPolicyAttachment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.UserPolicyAttachmentGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.UserPolicyAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.UserPolicyAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewUserPolicyAttachmentClient}),
			managed.WithConnectionPublishers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.UserPolicyAttachmentClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.UserPolicyAttachmentClient
	kube   client.Client
}

// Return an array of policy ARNs from attached policies
func getPolicyARNs(p []awsiamtypes.AttachedPolicy) []string {
	parns := make([]string, 0, len(p))
	for _, tp := range p {
		parns = append(parns, aws.ToString(tp.PolicyArn))
	}
	return parns
}

func (e *external) isUpToDate(cr *v1beta1.UserPolicyAttachment, resp *awsiam.ListAttachedUserPoliciesOutput) bool {
	if len(cr.Spec.ForProvider.PolicyARNs) != len(resp.AttachedPolicies) {
		return false
	}
	attachedARNs := getPolicyARNs(resp.AttachedPolicies)
	sort.Strings(attachedARNs)
	for _, v := range cr.Spec.ForProvider.PolicyARNs {
		if sort.SearchStrings(attachedARNs, v) == len(attachedARNs) {
			return false
		}
	}
	return true
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.UserPolicyAttachment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.ListAttachedUserPolicies(ctx, &awsiam.ListAttachedUserPoliciesInput{
		UserName: aws.String(cr.Spec.ForProvider.UserName),
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if len(observed.AttachedPolicies) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider.AttachedPolicyARNs = getPolicyARNs(observed.AttachedPolicies)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: e.isUpToDate(cr, observed),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.UserPolicyAttachment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	for _, policy := range cr.Spec.ForProvider.PolicyARNs {
		_, err := e.client.AttachUserPolicy(ctx, &awsiam.AttachUserPolicyInput{
			PolicyArn: aws.String(policy),
			UserName:  aws.String(cr.Spec.ForProvider.UserName),
		})
		if err != nil {
			return managed.ExternalCreation{}, awsclient.Wrap(err, errAttach)
		}
	}
	return managed.ExternalCreation{}, nil
}

// Return ARNs from target that are not contained in match
func unmatchedPolicyARNs(target []string, match []string) []string {
	unmatchedPolicyARNs := make([]string, 0, len(target))
	sort.Strings(match)
	for _, t := range target {
		if sort.SearchStrings(match, t) == len(match) {
			unmatchedPolicyARNs = append(unmatchedPolicyARNs, t)
		}
	}
	return unmatchedPolicyARNs
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.UserPolicyAttachment)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.ListAttachedUserPolicies(ctx, &awsiam.ListAttachedUserPoliciesInput{
		UserName: aws.String(cr.Spec.ForProvider.UserName),
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errGet)
	}

	needsAttachPolicyARNs := unmatchedPolicyARNs(cr.Spec.ForProvider.PolicyARNs, getPolicyARNs(observed.AttachedPolicies))
	for _, policy := range needsAttachPolicyARNs {
		_, err := e.client.AttachUserPolicy(ctx, &awsiam.AttachUserPolicyInput{
			PolicyArn: aws.String(policy),
			UserName:  aws.String(cr.Spec.ForProvider.UserName),
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errAttach)
		}
	}

	needsDeletePolicyARNs := unmatchedPolicyARNs(getPolicyARNs(observed.AttachedPolicies), cr.Spec.ForProvider.PolicyARNs)
	for _, policy := range needsDeletePolicyARNs {
		_, err = e.client.DetachUserPolicy(ctx, &awsiam.DetachUserPolicyInput{
			PolicyArn: aws.String(policy),
			UserName:  aws.String(cr.Spec.ForProvider.UserName),
		})
		if resource.Ignore(iam.IsErrorNotFound, err) != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errDetach)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.UserPolicyAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	for _, policy := range cr.Spec.ForProvider.PolicyARNs {
		_, err := e.client.DetachUserPolicy(ctx, &awsiam.DetachUserPolicyInput{
			PolicyArn: aws.String(policy),
			UserName:  aws.String(cr.Spec.ForProvider.UserName),
		})
		if resource.Ignore(iam.IsErrorNotFound, err) != nil {
			return awsclient.Wrap(err, errDetach)
		}
	}
	return nil
}
