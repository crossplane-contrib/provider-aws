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

package iampolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
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
	errNotPolicyInstance = "managed resource is not an IAMPolicy custom resource"

	errCreatePolicyClient = "cannot create IAM Policy client"
	errGetProvider        = "cannot get provider"
	errGetProviderSecret  = "cannot get provider secret"

	errUnexpectedObject = "The managed resource is not a IAMPolicy resource"
	errGet              = "failed to get IAM Policy"
	errCreate           = "failed to create the IAM Policy"
	errDelete           = "failed to delete the IAM Policy"
	errUpdate           = "failed to update the IAM Policy"
	errEmptyPolicy      = "empty IAM Policy received from IAM API"
	errPolicyVersion    = "No version for policy received from IAM API"
	errUpToDate         = "cannt check if policy is up to date"
)

// SetupIAMPolicy adds a controller that reconciles IAM Policy.
func SetupIAMPolicy(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.IAMPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.IAMPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.IAMPolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewPolicyClient}),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (iam.PolicyClient, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.IAMPolicy)
	if !ok {
		return nil, errors.New(errNotPolicyInstance)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		policyClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: policyClient, kube: c.kube}, errors.Wrap(err, errCreatePolicyClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	policyClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: policyClient, kube: c.kube}, errors.Wrap(err, errCreatePolicyClient)
}

type external struct {
	client iam.PolicyClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if !awsarn.IsARN(meta.GetExternalName(cr)) {
		return managed.ExternalObservation{}, nil
	}

	policyResp, err := e.client.GetPolicyRequest(&awsiam.GetPolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if policyResp.Policy == nil {
		return managed.ExternalObservation{}, errors.New(errEmptyPolicy)
	}
	policy := policyResp.Policy

	cr.SetConditions(runtimev1alpha1.Available())

	cr.Status.AtProvider = v1alpha1.IAMPolicyObservation{
		ARN:                           aws.StringValue(policy.Arn),
		AttachmentCount:               aws.Int64Value(policy.AttachmentCount),
		DefaultVersionID:              aws.StringValue(policy.DefaultVersionId),
		IsAttachable:                  aws.BoolValue(policy.IsAttachable),
		PermissionsBoundaryUsageCount: aws.Int64Value(policy.PermissionsBoundaryUsageCount),
		PolicyID:                      aws.StringValue(policy.PolicyId),
	}

	versionRsp, err := e.client.GetPolicyVersionRequest(&awsiam.GetPolicyVersionInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
		VersionId: aws.String(cr.Status.AtProvider.DefaultVersionID),
	}).Send(ctx)

	if err != nil || versionRsp.PolicyVersion == nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errPolicyVersion)
	}

	update, err := iam.IsPolicyUpToDate(cr.Spec.ForProvider, *versionRsp.PolicyVersion)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: update,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	createResp, err := e.client.CreatePolicyRequest(&awsiam.CreatePolicyInput{
		Description:    cr.Spec.ForProvider.Description,
		Path:           cr.Spec.ForProvider.Path,
		PolicyDocument: aws.String(cr.Spec.ForProvider.Document),
		PolicyName:     aws.String(cr.Spec.ForProvider.Name),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(createResp.CreatePolicyOutput.Policy.Arn))

	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// An update to AWS Policy is a new version of that policy.
	// A maximum of 5 versions are allowed. Below, the oldest version is deleted
	// for an update request when 5 versions already exist.
	// The new version is set as default.

	if err := e.deleteOldestVersion(ctx, meta.GetExternalName(cr)); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	_, err := e.client.CreatePolicyVersionRequest(&awsiam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(meta.GetExternalName(cr)),
		PolicyDocument: aws.String(cr.Spec.ForProvider.Document),
		SetAsDefault:   aws.Bool(true),
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	if err := e.deleteNonDefaultVersions(ctx, meta.GetExternalName(cr)); err != nil {
		return errors.Wrap(err, errDelete)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeletePolicyRequest(&awsiam.DeletePolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}

func (e *external) listPolicyVersions(ctx context.Context, policyArn string) ([]awsiam.PolicyVersion, error) {
	resp, err := e.client.ListPolicyVersionsRequest(&awsiam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	}).Send(ctx)

	if err != nil || resp.Versions == nil {
		return nil, err
	}

	return resp.Versions, nil
}

func (e *external) deleteOldestVersion(ctx context.Context, arn string) error {
	allVersions, err := e.listPolicyVersions(ctx, arn)
	if err != nil {
		return err
	}

	if len(allVersions) < 5 {
		return nil
	}

	var oldestVersion awsiam.PolicyVersion

	// loop through all version to find the oldest version.
	for _, version := range allVersions {
		if *version.IsDefaultVersion {
			continue
		}
		if oldestVersion.CreateDate == nil ||
			version.CreateDate.Before(*oldestVersion.CreateDate) {
			oldestVersion = version
		}
	}

	_, err = e.client.DeletePolicyVersionRequest(&awsiam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: oldestVersion.VersionId,
	}).Send(ctx)

	return err
}

func (e *external) deleteNonDefaultVersions(ctx context.Context, policyArn string) error {
	allVersions, err := e.listPolicyVersions(ctx, policyArn)
	if err != nil {
		return err
	}

	// loop through all the version and delete all non-default versions.
	for _, version := range allVersions {
		if *version.IsDefaultVersion {
			continue
		}
		if _, err := e.client.DeletePolicyVersionRequest(&awsiam.DeletePolicyVersionInput{
			PolicyArn: aws.String(policyArn),
			VersionId: version.VersionId,
		}).Send(ctx); err != nil {
			return err
		}
	}

	return nil
}
